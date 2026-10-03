// Package config carga la configuración de la aplicación desde variables de
// entorno, aplica valores por defecto de desarrollo y la valida al arrancar
// (falla rápido con un mensaje que identifica la variable problemática).
//
// No usa ninguna dependencia externa: os.Getenv + validación (D12/D-A6).
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Entornos reconocidos por APP_ENV. APP_ENV identifica el entorno y viaja en
// los logs; no se restringe a un conjunto cerrado para no bloquear entornos
// nuevos, pero estos son los de F1.
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvTest        = "test"
)

// Nombres canónicos de las variables que lee este paquete. Deben coincidir
// exactamente con los que documenta .env.example (T004). No existe HTTP_ADDR:
// el servidor escucha en todas las interfaces con :HTTP_PORT.
const (
	envAppEnv      = "APP_ENV"
	envHTTPPort    = "HTTP_PORT"
	envDatabaseURL = "DATABASE_URL"
	envLogLevel    = "LOG_LEVEL"
	envCORSOrigins = "CORS_ALLOWED_ORIGINS"
)

// Valores por defecto de desarrollo: permiten `make up` en un clon limpio sin
// ningún archivo .env (FR-001, research R19). La contraseña por defecto es la
// del servicio `db` del compose (valor de ejemplo, no un secreto real).
const (
	defaultAppEnv      = EnvDevelopment
	defaultHTTPPort    = 8080
	defaultDatabaseURL = "postgres://app:app_dev_password@localhost:5432/app?sslmode=disable"
	defaultLogLevel    = "info"
	defaultCORSOrigins = "http://localhost:5173"
)

// Config es la configuración validada de la aplicación.
type Config struct {
	// AppEnv identifica el entorno (APP_ENV); viaja en los logs.
	AppEnv string
	// HTTPPort es el puerto de escucha del servidor HTTP (HTTP_PORT).
	HTTPPort int
	// DatabaseURL es la cadena de conexión de PostgreSQL (DATABASE_URL).
	DatabaseURL string
	// LogLevel es el nivel mínimo de los logs estructurados (LOG_LEVEL).
	LogLevel slog.Level
	// CORSAllowedOrigins son los orígenes permitidos (CORS_ALLOWED_ORIGINS).
	CORSAllowedOrigins []string
}

// Addr devuelve la dirección de escucha del servidor HTTP (":8080").
func (c Config) Addr() string {
	return fmt.Sprintf(":%d", c.HTTPPort)
}

// Load lee las variables de entorno canónicas, aplica los valores por defecto
// de desarrollo y valida el resultado. Devuelve un error que identifica la
// variable si un valor es inválido.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:      getString(envAppEnv, defaultAppEnv),
		DatabaseURL: getString(envDatabaseURL, defaultDatabaseURL),
	}

	port, err := parsePort(getString(envHTTPPort, strconv.Itoa(defaultHTTPPort)))
	if err != nil {
		return Config{}, err
	}
	cfg.HTTPPort = port

	level, err := parseLogLevel(getString(envLogLevel, defaultLogLevel))
	if err != nil {
		return Config{}, err
	}
	cfg.LogLevel = level

	origins, err := parseCORSOrigins(getString(envCORSOrigins, defaultCORSOrigins))
	if err != nil {
		return Config{}, err
	}
	cfg.CORSAllowedOrigins = origins

	return cfg, nil
}

// getString devuelve el valor de key sin espacios o def si está ausente o en
// blanco (un valor vacío equivale a no definido).
func getString(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// parsePort valida que value sea un puerto TCP utilizable.
func parsePort(value string) (int, error) {
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("%s: %q no es un puerto válido (1-65535)", envHTTPPort, value)
	}
	return port, nil
}

// parseLogLevel traduce LOG_LEVEL a un nivel de slog.
func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("%s: %q no es válido (usa debug, info, warn o error)", envLogLevel, value)
	}
}

// parseCORSOrigins parte una lista separada por comas, descarta los elementos
// vacíos y exige al menos un origen.
func parseCORSOrigins(value string) ([]string, error) {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if origin := strings.TrimSpace(part); origin != "" {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return nil, fmt.Errorf("%s: no contiene ningún origen válido", envCORSOrigins)
	}
	return origins, nil
}
