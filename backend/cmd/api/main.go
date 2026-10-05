// Comando api: arranque del backend. Compone manualmente las dependencias
// (config → logger → pool → Redis → dominio status y usuarios → servidor),
// publica GET /healthz y la superficie de acceso y de panel de F2, y se apaga
// ordenadamente ante SIGINT/SIGTERM. Sin estado global (arq. R6).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"simiente-santa/backend/internal/platform/config"
	"simiente-santa/backend/internal/platform/database"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/logger"
	"simiente-santa/backend/internal/platform/middleware"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/status"
	"simiente-santa/backend/internal/usuarios"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	appLog := logger.New(cfg.LogLevel)

	// El pool es perezoso: una BD ausente o lenta no impide arrancar (R7).
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		appLog.Error("crear pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Redis (sesión y contadores de FR-006) también es perezoso: no abre
	// conexión al construirse (R7). Es estado efímero, sin persistencia (P23).
	redisClient, err := session.NewRedisClient(cfg.RedisURL)
	if err != nil {
		appLog.Error("crear cliente Redis", "error", err)
		os.Exit(1)
	}
	defer func() { _ = redisClient.Close() }()

	sessions, err := session.NewRedisStore(redisClient, session.StoreConfig{
		IdleTTL:     cfg.SessionIdleTTL,
		AbsoluteTTL: cfg.SessionAbsoluteTTL,
	})
	if err != nil {
		appLog.Error("crear store de sesiones", "error", err)
		os.Exit(1)
	}

	throttle, err := session.NewThrottle(redisClient, usuarios.MaxFailedAttempts, usuarios.LockoutDuration)
	if err != nil {
		appLog.Error("crear contadores de intentos", "error", err)
		os.Exit(1)
	}

	// DI manual del dominio usuarios: el repositorio sobre el pool, el servicio
	// de auditoría (que también es el audit.Recorder de la cadena), el servicio
	// de acceso, que resuelve la identidad de cada petición (session.Resolver), y
	// el de inicialización única (FR-007).
	repo := usuarios.NewRepository(pool)
	auditSvc := usuarios.NewAuditService(repo, appLog)
	authSvc := usuarios.NewAuthService(usuarios.AuthServiceDeps{
		Repository:  repo,
		Sessions:    sessions,
		Throttle:    throttle,
		LoginEvents: auditSvc,
		Cookies:     session.NewCookieConfig(cfg.SessionCookieSecure, cfg.SessionAbsoluteTTL),
		CSRFSecret:  cfg.SessionSecret,
		Logger:      appLog,
	})
	initSvc := usuarios.NewInitService(usuarios.InitServiceDeps{
		Repository: repo,
		Logger:     appLog,
	})
	// Gestión de cuentas (T231–T234): reutiliza el mismo repositorio, el Store
	// de sesiones (revocación al desactivar/restablecer) y el servicio de
	// auditoría (registro best-effort de los fallos).
	userSvc := usuarios.NewUserService(usuarios.UserServiceDeps{
		Repository: repo,
		Sessions:   sessions,
		Audit:      auditSvc,
		Logger:     appLog,
	})
	// Gestión de roles y catálogo de permisos (T235–T237): mismo repositorio y
	// servicio de auditoría (registro best-effort de los fallos).
	roleSvc := usuarios.NewRoleService(usuarios.RoleServiceDeps{
		Repository: repo,
		Audit:      auditSvc,
	})

	handler := usuarios.NewHandler(usuarios.HandlerDeps{
		Access:     authSvc,
		Setup:      initSvc,
		Users:      userSvc,
		Roles:      roleSvc,
		Audit:      auditSvc,
		SetupToken: cfg.BootstrapToken,
		Logger:     appLog,
	})

	appLog.Info("api arrancando", "env", cfg.AppEnv, "port", cfg.HTTPPort)

	deps := apiDeps{
		statusRepo:  status.NewRepository(pool),
		authHandler: handler,
		public: usuarios.PublicDeps{
			// Rate-limit por IP de la superficie pública escribible (P17):
			// POST /auth/login y POST /setup/initialize.
			Login: []httpserver.Middleware{
				middleware.RateLimit(middleware.RateLimitConfig{Paths: middleware.DefaultRateLimitPaths()}, appLog),
			},
			// Rutas de sesión: authn → CSRF, sin el guard de cambio obligatorio.
			Session: []httpserver.Middleware{
				middleware.Authn(sessions, authSvc, appLog),
				middleware.CSRF(cfg.SessionSecret, appLog),
			},
			Setup: []httpserver.Middleware{
				middleware.RateLimit(middleware.RateLimitConfig{Paths: middleware.DefaultRateLimitPaths()}, appLog),
			},
		},
		admin: usuarios.AdminDeps{
			Module: usuarios.PermissionAdminUsersRoles,
			Deps: middleware.AdminDeps{
				Sessions:   sessions,
				Resolver:   authSvc,
				Recorder:   auditSvc,
				CSRFSecret: cfg.SessionSecret,
				Logger:     appLog,
			},
			// Las rutas de cuentas (T234), de roles (T237) y de auditoría
			// (T240) se publican desde este handler.
			Handler: handler,
		},
	}

	srv := httpserver.New(newMux(deps, appLog),
		httpserver.Options{Addr: cfg.Addr()},
		appLog,
		middleware.Chain(appLog, cfg.CORSAllowedOrigins)...,
	)

	if err := srv.Run(ctx); err != nil {
		appLog.Error("servidor", "error", err)
		os.Exit(1)
	}
}

// apiDeps agrupa las dependencias del enrutador ya construidas por main. Es un
// struct para que newMux sea puro y las pruebas puedan inyectar dobles sin base
// de datos ni Redis (FR-011).
type apiDeps struct {
	statusRepo  status.Repository
	authHandler *usuarios.Handler
	public      usuarios.PublicDeps
	admin       usuarios.AdminDeps
}

// newMux compone el enrutador con las rutas de status y las del dominio usuarios
// (acceso y grupo de panel). Recibe las dependencias ya construidas.
func newMux(deps apiDeps, appLog *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	root := httpserver.NewMuxRegistrar(mux)
	status.RegisterPublic(root, status.NewHandler(status.NewService(deps.statusRepo), appLog))
	usuarios.RegisterPublic(root, deps.authHandler, deps.public)
	usuarios.RegisterAdmin(root, deps.admin)
	return mux
}
