// Package config carga la configuración de la aplicación desde variables de
// entorno, aplica valores por defecto de desarrollo y la valida al arrancar
// (falla rápido con un mensaje que identifica la variable problemática).
//
// No usa ninguna dependencia externa: os.Getenv + validación (D12/D-A6).
package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
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
// exactamente con los que documenta .env.example (T004, T206) y con los que
// inyecta docker-compose.yml (T205). No existe HTTP_ADDR: el servidor escucha
// en todas las interfaces con :HTTP_PORT.
const (
	envAppEnv                    = "APP_ENV"
	envHTTPPort                  = "HTTP_PORT"
	envDatabaseURL               = "DATABASE_URL"
	envLogLevel                  = "LOG_LEVEL"
	envCORSOrigins               = "CORS_ALLOWED_ORIGINS"
	envRedisURL                  = "REDIS_URL"
	envSessionSecret             = "SESSION_SECRET"
	envSessionCookieSecure       = "SESSION_COOKIE_SECURE"
	envSessionIdleTTLMinutes     = "SESSION_IDLE_TTL_MINUTES"
	envSessionAbsoluteTTLMinutes = "SESSION_ABSOLUTE_TTL_MINUTES"
	envBootstrapToken            = "BOOTSTRAP_TOKEN"
	envUploadDir                 = "UPLOAD_DIR"
	envUploadMaxBytes            = "UPLOAD_MAX_BYTES"
)

// Valores por defecto de desarrollo: permiten `make up` en un clon limpio sin
// ningún archivo .env (FR-001, research R19). La contraseña por defecto es la
// del servicio `db` del compose (valor de ejemplo, no un secreto real).
//
// REDIS_URL, SESSION_SECRET y BOOTSTRAP_TOKEN no tienen defecto: son variables
// obligatorias o secretos que nunca viven en el código (§IV). SESSION_SECRET y
// BOOTSTRAP_TOKEN los documenta .env.example con valores de ejemplo; REDIS_URL
// la inyecta siempre compose (T205), así que su ausencia es un error de
// arranque (R1, §VII).
const (
	defaultAppEnv                    = EnvDevelopment
	defaultHTTPPort                  = 8080
	defaultDatabaseURL               = "postgres://app:app_dev_password@localhost:5432/app?sslmode=disable"
	defaultLogLevel                  = "info"
	defaultCORSOrigins               = "http://localhost:5173"
	defaultSessionCookieSecure       = false
	defaultSessionIdleTTLMinutes     = 30
	defaultSessionAbsoluteTTLMinutes = 60
	// Imágenes de la portada (F3, R3-8): fuera de Docker `./uploads`; 8 MB de
	// tope por subida. El compose inyecta el directorio del volumen.
	defaultUploadDir      = "./uploads"
	defaultUploadMaxBytes = 8388608
)

// minSessionSecretLength es la longitud mínima razonable de SESSION_SECRET: la
// clave HMAC firma la cookie CSRF, y una clave corta es atacable por fuerza
// bruta (CWE-326). En producción es obligatoria (error de arranque).
const minSessionSecretLength = 32

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
	// RedisURL es la cadena de conexión de Redis (REDIS_URL), usada por la
	// sesión y los contadores de intentos. Obligatoria y validada al arrancar.
	RedisURL string
	// SessionSecret firma la cookie CSRF de sesión (SESSION_SECRET). Secreto:
	// sin valor por defecto en el código (§IV).
	SessionSecret string
	// SessionCookieSecure marca la cookie de sesión como `Secure`
	// (SESSION_COOKIE_SECURE); debe ser true en producción (R11).
	SessionCookieSecure bool
	// SessionIdleTTL es la vida de sesión por inactividad
	// (SESSION_IDLE_TTL_MINUTES, 30 por defecto).
	SessionIdleTTL time.Duration
	// SessionAbsoluteTTL es la vida absoluta de sesión desde el login
	// (SESSION_ABSOLUTE_TTL_MINUTES, 60 por defecto).
	SessionAbsoluteTTL time.Duration
	// BootstrapToken autoriza la inicialización única del administrador
	// (BOOTSTRAP_TOKEN, cabecera X-Setup-Token). Secreto: sin valor por
	// defecto en el código (§IV).
	BootstrapToken string
	// UploadDir es el directorio donde el backend guarda/lee las imágenes de
	// la portada (UPLOAD_DIR; por defecto ./uploads fuera de Docker).
	UploadDir string
	// UploadMaxBytes es el tope de tamaño por subida (UPLOAD_MAX_BYTES;
	// por defecto 8388608 = 8 MB).
	UploadMaxBytes int64
	// Warnings son avisos de configuración no fatales que el arranque debe
	// registrar sin impedirlo (p. ej. SESSION_SECRET débil en desarrollo o
	// BOOTSTRAP_TOKEN vacío). En producción esas mismas debilidades son errores.
	Warnings []string
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

	// REDIS_URL es obligatoria (R1): sin ella la sesión y los contadores no
	// pueden funcionar, así que el arranque falla con un mensaje claro.
	redisURL, err := parseRedisURL(os.Getenv(envRedisURL))
	if err != nil {
		return Config{}, err
	}
	cfg.RedisURL = redisURL

	// Secretos: sin defecto en el código (§IV). Vacío equivale a no definido.
	cfg.SessionSecret = getString(envSessionSecret, "")
	cfg.BootstrapToken = getString(envBootstrapToken, "")

	secure, err := parseBoolDefault(envSessionCookieSecure, defaultSessionCookieSecure)
	if err != nil {
		return Config{}, err
	}
	cfg.SessionCookieSecure = secure

	idleMinutes, err := parsePositiveMinutes(envSessionIdleTTLMinutes, defaultSessionIdleTTLMinutes)
	if err != nil {
		return Config{}, err
	}
	cfg.SessionIdleTTL = time.Duration(idleMinutes) * time.Minute

	absoluteMinutes, err := parsePositiveMinutes(envSessionAbsoluteTTLMinutes, defaultSessionAbsoluteTTLMinutes)
	if err != nil {
		return Config{}, err
	}
	cfg.SessionAbsoluteTTL = time.Duration(absoluteMinutes) * time.Minute

	// Imágenes de la portada (F3): directorio y tope de subida. Valores por
	// defecto usables en local; un valor inválido impide el arranque.
	cfg.UploadDir = getString(envUploadDir, defaultUploadDir)
	uploadMaxBytes, err := parsePositiveInt64(envUploadMaxBytes, defaultUploadMaxBytes)
	if err != nil {
		return Config{}, err
	}
	cfg.UploadMaxBytes = uploadMaxBytes

	// R11: en producción la cookie de sesión debe viajar por HTTPS; una cookie
	// insegura es un fallo de configuración que debe impedir el arranque.
	if cfg.AppEnv == EnvProduction && !cfg.SessionCookieSecure {
		return Config{}, fmt.Errorf(
			"%s: debe ser true cuando %s=%s (la cookie de sesión viaja por HTTPS en producción)",
			envSessionCookieSecure, envAppEnv, EnvProduction,
		)
	}

	// M-1: SESSION_SECRET firma la cookie CSRF. Vacío o corto degrada el
	// double-submit firmado a uno simple, así que en producción impide el
	// arranque y en desarrollo avisa.
	if err := validateSessionSecret(cfg.AppEnv, cfg.SessionSecret, &cfg.Warnings); err != nil {
		return Config{}, err
	}
	if cfg.BootstrapToken == "" {
		cfg.Warnings = append(cfg.Warnings, fmt.Sprintf(
			"%s está vacío: la inicialización única del administrador quedará deshabilitada",
			envBootstrapToken,
		))
	}

	return cfg, nil
}

// validateSessionSecret exige SESSION_SECRET no vacío y con longitud mínima
// (minSessionSecretLength) cuando APP_ENV=production, con un error que nombra la
// variable. En el resto de entornos permite arrancar pero acumula un aviso.
func validateSessionSecret(appEnv, secret string, warnings *[]string) error {
	if secret != "" && len(secret) >= minSessionSecretLength {
		return nil
	}
	if appEnv == EnvProduction {
		return fmt.Errorf(
			"%s: es obligatoria y debe tener al menos %d caracteres cuando %s=%s (firma la cookie CSRF)",
			envSessionSecret, minSessionSecretLength, envAppEnv, EnvProduction,
		)
	}
	*warnings = append(*warnings, fmt.Sprintf(
		"%s vacío o con menos de %d caracteres: la firma del CSRF queda degradada; define un valor aleatorio largo",
		envSessionSecret, minSessionSecretLength,
	))
	return nil
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

// parseRedisURL exige que REDIS_URL esté definida y sea una URL de Redis
// parseable (esquema redis:// o rediss:// con host). Es obligatoria (R1, §VII).
func parseRedisURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s: es obligatoria (p. ej. redis://redis:6379/0)", envRedisURL)
	}
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "redis" && u.Scheme != "rediss") || u.Host == "" {
		return "", fmt.Errorf("%s: %q no es una URL de Redis válida (usa redis:// o rediss://)", envRedisURL, value)
	}
	return value, nil
}

// parseBoolDefault lee un booleano de key y devuelve def si está ausente o en
// blanco. Un valor no booleano produce un error que nombra la variable.
func parseBoolDefault(key string, def bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def, nil
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s: %q no es un booleano válido (usa true o false)", key, raw)
	}
	return b, nil
}

// parsePositiveMinutes lee un número entero de minutos de key y devuelve def si
// está ausente o en blanco. Debe ser estrictamente positivo.
func parsePositiveMinutes(key string, def int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s: %q no es un número de minutos válido (entero positivo)", key, raw)
	}
	return n, nil
}

// parsePositiveInt64 lee un entero de 64 bits de key y devuelve def si está
// ausente o en blanco. Debe ser estrictamente positivo (UPLOAD_MAX_BYTES).
func parsePositiveInt64(key string, def int64) (int64, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def, nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s: %q no es un número válido (entero positivo)", key, raw)
	}
	return n, nil
}
