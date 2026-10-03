// Comando api: arranque del backend. Compone manualmente las dependencias
// (config → logger → pool → dominio status → servidor), publica GET /healthz y
// se apaga ordenadamente ante SIGINT/SIGTERM. Sin estado global (arq. R6).
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
	"simiente-santa/backend/internal/status"
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

	appLog.Info("api arrancando", "env", cfg.AppEnv, "port", cfg.HTTPPort)

	srv := httpserver.New(newMux(status.NewRepository(pool), appLog),
		httpserver.Options{Addr: cfg.Addr()},
		appLog,
		middleware.Chain(appLog, cfg.CORSAllowedOrigins)...,
	)

	if err := srv.Run(ctx); err != nil {
		appLog.Error("servidor", "error", err)
		os.Exit(1)
	}
}

// newMux compone el enrutador con las rutas públicas del dominio status. Recibe
// el repositorio ya construido para que las pruebas puedan inyectar un doble
// sin base de datos.
func newMux(repo status.Repository, appLog *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	root := httpserver.NewMuxRegistrar(mux)
	status.RegisterPublic(root, status.NewHandler(status.NewService(repo), appLog))
	return mux
}
