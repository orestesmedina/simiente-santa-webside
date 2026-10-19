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

	"simiente-santa/backend/internal/agenda"
	"simiente-santa/backend/internal/donaciones"
	"simiente-santa/backend/internal/media"
	"simiente-santa/backend/internal/medios"
	"simiente-santa/backend/internal/ministerios"
	"simiente-santa/backend/internal/platform/config"
	"simiente-santa/backend/internal/platform/database"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/logger"
	"simiente-santa/backend/internal/platform/middleware"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/storage"
	"simiente-santa/backend/internal/portada"
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

	// Avisos de configuración no fatales (p. ej. SESSION_SECRET débil en
	// desarrollo): se registran sin impedir el arranque.
	for _, warning := range cfg.Warnings {
		appLog.Warn(warning)
	}

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

	// Dominio portada (F3): repositorio sobre el mismo pool, almacén local de
	// imágenes (UPLOAD_DIR/UPLOAD_MAX_BYTES) y servicio que también es el
	// audit.Recorder del grupo de panel (T321).
	portadaRepo := portada.NewRepository(pool)
	portadaStore := storage.NewLocalStore(cfg.UploadDir, cfg.UploadMaxBytes)
	portadaSvc := portada.NewService(portada.ServiceDeps{
		Repository: portadaRepo,
		Store:      portadaStore,
		Logger:     appLog,
	})
	portadaHandler := portada.NewHandler(portada.HandlerDeps{
		Service:        portadaSvc,
		MaxUploadBytes: cfg.UploadMaxBytes,
		Logger:         appLog,
	})

	// Dominio media (F4/R4-6): la descarga pública de archivos, promovida desde
	// la portada. Comparte el almacén local y usa la zona del lugar (APP_TZ,
	// T409) para el corte "publicado y próximo" de las imágenes de agenda.
	mediaSvc := media.NewService(media.ServiceDeps{
		Repository: media.NewRepository(pool),
		Store:      portadaStore,
		Location:   cfg.TimeZone,
		Logger:     appLog,
	})
	mediaHandler := media.NewHandler(media.HandlerDeps{Service: mediaSvc, Logger: appLog})

	// Dominio agenda (F4): eventos, actividades y el catálogo de categorías.
	// Comparte el almacén de imágenes de F3 (misma política de subida) y usa la
	// zona del lugar (APP_TZ) para la regla "próximo"; su service también es el
	// audit.Recorder de los tres grupos de panel (T422).
	agendaSvc := agenda.NewService(agenda.ServiceDeps{
		Repository: agenda.NewRepository(pool),
		Store:      portadaStore,
		Location:   cfg.TimeZone,
		Logger:     appLog,
	})
	agendaHandler := agenda.NewHandler(agenda.HandlerDeps{
		Service:        agendaSvc,
		MaxUploadBytes: cfg.UploadMaxBytes,
		Logger:         appLog,
	})

	// Dominio ministerios (F6): la sección pública, su bloque en la portada y la
	// gestión desde el panel. Comparte el almacén de imágenes de F3 (misma
	// política de subida) y su service también es el audit.Recorder del grupo de
	// panel (T616). El permiso `ministerios` ya está sembrado por 000002: F6 no
	// añade permisos.
	ministeriosSvc := ministerios.NewService(ministerios.ServiceDeps{
		Repository: ministerios.NewRepository(pool),
		Store:      portadaStore,
		Logger:     appLog,
	})
	ministeriosHandler := ministerios.NewHandler(ministerios.HandlerDeps{
		Service:        ministeriosSvc,
		MaxUploadBytes: cfg.UploadMaxBytes,
		Logger:         appLog,
	})

	// Dominio donaciones (F7): la sección pública de ofrendas, su bloque en la
	// portada y la gestión desde el panel. Sin imágenes (Q6/FR-020) y con su
	// service como audit.Recorder del grupo de panel (T716). El permiso
	// `donaciones` ya está sembrado por 000002: F7 no añade permisos.
	donacionesSvc := donaciones.NewService(donaciones.ServiceDeps{
		Repository: donaciones.NewRepository(pool),
		Logger:     appLog,
	})
	donacionesHandler := donaciones.NewHandler(donaciones.HandlerDeps{
		Service: donacionesSvc,
		Logger:  appLog,
	})

	// Dominio medios (F9): la sección pública de prédicas y podcasts, su bloque en
	// la portada y la gestión desde el panel. Comparte el almacén de archivos de
	// F3/F4/F6 (las carátulas se copian al sitio y se sirven como
	// `/api/v1/media/<file>`), sale a internet **solo** por la fábrica única de
	// cliente HTTP del dominio (P9-20) y usa la zona del lugar (APP_TZ) para el
	// «mismo día» de la unión automática (FR-013) y el corte de la ventana de los
	// feeds (FR-017). El permiso `medios` ya está sembrado por 000002: F9 no añade
	// permisos y el catálogo sigue en 10 (P9-15).
	mediosRed, err := medios.NewRedSaliente(portadaStore, cfg.UploadMaxBytes)
	if err != nil {
		appLog.Error("crear la red saliente de medios", "error", err)
		os.Exit(1)
	}
	mediosSvc := medios.NewService(medios.ServiceDeps{
		Repository:         medios.NewRepository(pool),
		FeedSource:         mediosRed.Feeds,
		AvailabilityProber: mediosRed.Probador,
		CoverFetcher:       mediosRed.Caratulas,
		Store:              portadaStore,
		Location:           cfg.TimeZone,
		Logger:             appLog,
	})
	mediosHandler := medios.NewHandler(medios.HandlerDeps{
		Service: mediosSvc,
		Logger:  appLog,
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
		portadaHandler: portadaHandler,
		portadaAdmin: portada.AdminDeps{
			Deps: middleware.AdminDeps{
				Sessions: sessions,
				Resolver: authSvc,
				// El Recorder del grupo de la portada es el service de F3
				// (T321): registra las denegaciones del módulo.
				Recorder:   portadaSvc,
				CSRFSecret: cfg.SessionSecret,
				Logger:     appLog,
			},
		},
		mediaHandler:  mediaHandler,
		agendaHandler: agendaHandler,
		agendaAdmin: agenda.AdminDeps{
			Deps: middleware.AdminDeps{
				Sessions: sessions,
				Resolver: authSvc,
				// El Recorder de los tres grupos es el service de F4 (T422):
				// registra las denegaciones de `event.*`/`activity.*`/`category.*`.
				Recorder:   agendaSvc,
				CSRFSecret: cfg.SessionSecret,
				Logger:     appLog,
			},
		},
		ministeriosHandler: ministeriosHandler,
		ministeriosAdmin: ministerios.AdminDeps{
			Deps: middleware.AdminDeps{
				Sessions: sessions,
				Resolver: authSvc,
				// El Recorder del grupo es el service de F6 (T616): registra las
				// denegaciones de `ministry.*`.
				Recorder:   ministeriosSvc,
				CSRFSecret: cfg.SessionSecret,
				Logger:     appLog,
			},
		},
		donacionesHandler: donacionesHandler,
		donacionesAdmin: donaciones.AdminDeps{
			Deps: middleware.AdminDeps{
				Sessions: sessions,
				Resolver: authSvc,
				// El Recorder del grupo es el service de F7 (T716): registra las
				// denegaciones de `donation.*`.
				Recorder:   donacionesSvc,
				CSRFSecret: cfg.SessionSecret,
				Logger:     appLog,
			},
		},
		mediosHandler: mediosHandler,
		mediosAdmin: medios.AdminDeps{
			Deps: middleware.AdminDeps{
				Sessions: sessions,
				Resolver: authSvc,
				// El Recorder del grupo es el service de F9 (T922): registra las
				// denegaciones de `media.*` (y los rechazos del módulo).
				Recorder:   mediosSvc,
				CSRFSecret: cfg.SessionSecret,
				Logger:     appLog,
			},
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
	statusRepo         status.Repository
	authHandler        *usuarios.Handler
	public             usuarios.PublicDeps
	admin              usuarios.AdminDeps
	portadaHandler     *portada.Handler
	portadaAdmin       portada.AdminDeps
	mediaHandler       *media.Handler
	agendaHandler      *agenda.Handler
	agendaAdmin        agenda.AdminDeps
	ministeriosHandler *ministerios.Handler
	ministeriosAdmin   ministerios.AdminDeps
	donacionesHandler  *donaciones.Handler
	donacionesAdmin    donaciones.AdminDeps
	mediosHandler      *medios.Handler
	mediosAdmin        medios.AdminDeps
}

// newMux compone el enrutador con las rutas de status y las de los dominios
// usuarios (acceso y grupo de panel), portada (superficie pública y panel), media
// (descarga pública de archivos), agenda («lo que viene» y los tres grupos de
// panel de F4), ministerios (sección pública, bloque de la portada y grupo de
// panel de F6), donaciones (sección pública, bloque de la portada y grupo de panel
// de F7) y medios (sección pública de F9, su bloque de la portada y el grupo de
// panel con el permiso `medios`). Recibe las dependencias ya construidas.
func newMux(deps apiDeps, appLog *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	root := httpserver.NewMuxRegistrar(mux)
	status.RegisterPublic(root, status.NewHandler(status.NewService(deps.statusRepo), appLog))
	usuarios.RegisterPublic(root, deps.authHandler, deps.public)
	usuarios.RegisterAdmin(root, deps.admin)
	portada.RegisterPublic(root, deps.portadaHandler)
	portada.RegisterAdmin(root, deps.portadaHandler, deps.portadaAdmin)
	media.RegisterPublic(root, deps.mediaHandler)
	agenda.RegisterPublic(root, deps.agendaHandler)
	agenda.RegisterAdmin(root, deps.agendaHandler, deps.agendaAdmin)
	ministerios.RegisterPublic(root, deps.ministeriosHandler)
	ministerios.RegisterAdmin(root, deps.ministeriosHandler, deps.ministeriosAdmin)
	donaciones.RegisterPublic(root, deps.donacionesHandler)
	donaciones.RegisterAdmin(root, deps.donacionesHandler, deps.donacionesAdmin)
	medios.RegisterPublic(root, deps.mediosHandler)
	medios.RegisterAdmin(root, deps.mediosHandler, deps.mediosAdmin)
	return mux
}
