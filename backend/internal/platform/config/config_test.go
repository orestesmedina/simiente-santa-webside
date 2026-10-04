package config

import (
	"log/slog"
	"strings"
	"testing"
)

// clearEnv deja las cinco variables canónicas sin valor para simular un clon
// limpio (sin .env). t.Setenv restaura los valores originales al terminar.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{envAppEnv, envHTTPPort, envDatabaseURL, envLogLevel, envCORSOrigins} {
		t.Setenv(key, "")
	}
}

func TestLoadDefaultsWithoutEnv(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() devolvió error en un clon limpio: %v", err)
	}

	if cfg.AppEnv != EnvDevelopment {
		t.Errorf("AppEnv = %q, se esperaba %q", cfg.AppEnv, EnvDevelopment)
	}
	if cfg.HTTPPort != defaultHTTPPort {
		t.Errorf("HTTPPort = %d, se esperaba %d", cfg.HTTPPort, defaultHTTPPort)
	}
	if cfg.DatabaseURL != defaultDatabaseURL {
		t.Errorf("DatabaseURL = %q, se esperaba %q", cfg.DatabaseURL, defaultDatabaseURL)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, se esperaba %v", cfg.LogLevel, slog.LevelInfo)
	}
	if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != defaultCORSOrigins {
		t.Errorf("CORSAllowedOrigins = %v, se esperaba [%q]", cfg.CORSAllowedOrigins, defaultCORSOrigins)
	}
	if cfg.Addr() != ":8080" {
		t.Errorf("Addr() = %q, se esperaba %q", cfg.Addr(), ":8080")
	}
}

func TestLoadEachCanonicalVariable(t *testing.T) {
	tests := []struct {
		name   string
		envKey string
		value  string
		check  func(t *testing.T, cfg Config)
	}{
		{
			name:   "APP_ENV",
			envKey: envAppEnv,
			value:  "test",
			check: func(t *testing.T, cfg Config) {
				if cfg.AppEnv != "test" {
					t.Errorf("AppEnv = %q, se esperaba %q", cfg.AppEnv, "test")
				}
			},
		},
		{
			name:   "HTTP_PORT",
			envKey: envHTTPPort,
			value:  "9090",
			check: func(t *testing.T, cfg Config) {
				if cfg.HTTPPort != 9090 {
					t.Errorf("HTTPPort = %d, se esperaba 9090", cfg.HTTPPort)
				}
				if cfg.Addr() != ":9090" {
					t.Errorf("Addr() = %q, se esperaba %q", cfg.Addr(), ":9090")
				}
			},
		},
		{
			name:   "DATABASE_URL",
			envKey: envDatabaseURL,
			value:  "postgres://u:p@db:5432/otra?sslmode=disable",
			check: func(t *testing.T, cfg Config) {
				const want = "postgres://u:p@db:5432/otra?sslmode=disable"
				if cfg.DatabaseURL != want {
					t.Errorf("DatabaseURL = %q, se esperaba %q", cfg.DatabaseURL, want)
				}
			},
		},
		{
			name:   "LOG_LEVEL",
			envKey: envLogLevel,
			value:  "warn",
			check: func(t *testing.T, cfg Config) {
				if cfg.LogLevel != slog.LevelWarn {
					t.Errorf("LogLevel = %v, se esperaba %v", cfg.LogLevel, slog.LevelWarn)
				}
			},
		},
		{
			name:   "CORS_ALLOWED_ORIGINS",
			envKey: envCORSOrigins,
			value:  "http://a.example, http://b.example",
			check: func(t *testing.T, cfg Config) {
				want := []string{"http://a.example", "http://b.example"}
				if len(cfg.CORSAllowedOrigins) != len(want) {
					t.Fatalf("CORSAllowedOrigins = %v, se esperaba %v", cfg.CORSAllowedOrigins, want)
				}
				for i := range want {
					if cfg.CORSAllowedOrigins[i] != want[i] {
						t.Errorf("CORSAllowedOrigins[%d] = %q, se esperaba %q", i, cfg.CORSAllowedOrigins[i], want[i])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv(tt.envKey, tt.value)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() devolvió error: %v", err)
			}
			tt.check(t, cfg)
		})
	}
}

func TestLoadInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		envKey string
		value  string
	}{
		{name: "HTTP_PORT no numérico", envKey: envHTTPPort, value: "ochenta"},
		{name: "HTTP_PORT cero", envKey: envHTTPPort, value: "0"},
		{name: "HTTP_PORT fuera de rango", envKey: envHTTPPort, value: "70000"},
		{name: "LOG_LEVEL desconocido", envKey: envLogLevel, value: "verbose"},
		{name: "CORS_ALLOWED_ORIGINS sin orígenes", envKey: envCORSOrigins, value: " , "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv(tt.envKey, tt.value)

			_, err := Load()
			if err == nil {
				t.Fatalf("Load() no devolvió error con %s=%q", tt.envKey, tt.value)
			}
			if !strings.Contains(err.Error(), tt.envKey) {
				t.Errorf("el error %q no identifica la variable %q", err.Error(), tt.envKey)
			}
		})
	}
}

// TestLoadIgnoresUnknownKeys fija que HTTP_ADDR no existe: definirla no cambia
// la configuración y el servidor sigue escuchando en :HTTP_PORT.
func TestLoadIgnoresUnknownKeys(t *testing.T) {
	clearEnv(t)
	t.Setenv("HTTP_ADDR", ":9999")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() devolvió error: %v", err)
	}
	if cfg.HTTPPort != defaultHTTPPort {
		t.Errorf("HTTPPort = %d, se esperaba el valor por defecto %d", cfg.HTTPPort, defaultHTTPPort)
	}
	if cfg.Addr() != ":8080" {
		t.Errorf("Addr() = %q, se esperaba %q", cfg.Addr(), ":8080")
	}
}
