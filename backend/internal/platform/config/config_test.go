package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

// testRedisURL es un valor de desarrollo, no un secreto: REDIS_URL es
// obligatoria (T216/R1), así que las pruebas que no la ejercitan la definen.
const testRedisURL = "redis://localhost:6379/0"

// testSessionSecret cumple la longitud mínima de SESSION_SECRET en producción.
const testSessionSecret = "secreto-de-prueba-con-longitud-suficiente-123456"

// canonicalKeys son todas las variables que lee platform/config. Deben coincidir
// con .env.example (T206) y con lo que inyecta docker-compose.yml (T205).
func canonicalKeys() []string {
	return []string{
		envAppEnv, envHTTPPort, envDatabaseURL, envLogLevel, envCORSOrigins,
		envRedisURL, envSessionSecret, envSessionCookieSecure,
		envSessionIdleTTLMinutes, envSessionAbsoluteTTLMinutes, envBootstrapToken,
	}
}

// clearEnv deja las variables canónicas sin valor para simular un clon limpio
// (sin .env). t.Setenv restaura los valores originales al terminar.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range canonicalKeys() {
		t.Setenv(key, "")
	}
}

// setRedis aporta REDIS_URL (obligatoria) a las pruebas que no la ejercitan.
func setRedis(t *testing.T) {
	t.Helper()
	t.Setenv(envRedisURL, testRedisURL)
}

func TestLoadDefaultsWithoutEnv(t *testing.T) {
	clearEnv(t)
	setRedis(t)

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

	// Variables de F2 con su valor por defecto de desarrollo.
	if cfg.RedisURL != testRedisURL {
		t.Errorf("RedisURL = %q, se esperaba %q", cfg.RedisURL, testRedisURL)
	}
	if cfg.SessionSecret != "" {
		t.Errorf("SessionSecret = %q, se esperaba vacío (nunca un secreto por defecto)", cfg.SessionSecret)
	}
	if cfg.SessionCookieSecure {
		t.Errorf("SessionCookieSecure = true, se esperaba false en desarrollo")
	}
	if cfg.SessionIdleTTL != 30*time.Minute {
		t.Errorf("SessionIdleTTL = %v, se esperaba %v", cfg.SessionIdleTTL, 30*time.Minute)
	}
	if cfg.SessionAbsoluteTTL != 60*time.Minute {
		t.Errorf("SessionAbsoluteTTL = %v, se esperaba %v", cfg.SessionAbsoluteTTL, 60*time.Minute)
	}
	if cfg.BootstrapToken != "" {
		t.Errorf("BootstrapToken = %q, se esperaba vacío (nunca un secreto por defecto)", cfg.BootstrapToken)
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
		{
			name:   "REDIS_URL",
			envKey: envRedisURL,
			value:  "redis://redis:6379/1",
			check: func(t *testing.T, cfg Config) {
				if cfg.RedisURL != "redis://redis:6379/1" {
					t.Errorf("RedisURL = %q, se esperaba %q", cfg.RedisURL, "redis://redis:6379/1")
				}
			},
		},
		{
			name:   "SESSION_SECRET",
			envKey: envSessionSecret,
			value:  "valor-de-prueba-no-secreto",
			check: func(t *testing.T, cfg Config) {
				if cfg.SessionSecret != "valor-de-prueba-no-secreto" {
					t.Errorf("SessionSecret = %q, se esperaba el valor configurado", cfg.SessionSecret)
				}
			},
		},
		{
			name:   "SESSION_COOKIE_SECURE",
			envKey: envSessionCookieSecure,
			value:  "true",
			check: func(t *testing.T, cfg Config) {
				if !cfg.SessionCookieSecure {
					t.Errorf("SessionCookieSecure = false, se esperaba true")
				}
			},
		},
		{
			name:   "SESSION_IDLE_TTL_MINUTES",
			envKey: envSessionIdleTTLMinutes,
			value:  "45",
			check: func(t *testing.T, cfg Config) {
				if cfg.SessionIdleTTL != 45*time.Minute {
					t.Errorf("SessionIdleTTL = %v, se esperaba %v", cfg.SessionIdleTTL, 45*time.Minute)
				}
			},
		},
		{
			name:   "SESSION_ABSOLUTE_TTL_MINUTES",
			envKey: envSessionAbsoluteTTLMinutes,
			value:  "120",
			check: func(t *testing.T, cfg Config) {
				if cfg.SessionAbsoluteTTL != 120*time.Minute {
					t.Errorf("SessionAbsoluteTTL = %v, se esperaba %v", cfg.SessionAbsoluteTTL, 120*time.Minute)
				}
			},
		},
		{
			name:   "BOOTSTRAP_TOKEN",
			envKey: envBootstrapToken,
			value:  "token-de-prueba-no-secreto",
			check: func(t *testing.T, cfg Config) {
				if cfg.BootstrapToken != "token-de-prueba-no-secreto" {
					t.Errorf("BootstrapToken = %q, se esperaba el valor configurado", cfg.BootstrapToken)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			if tt.envKey != envRedisURL {
				setRedis(t)
			}
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
		{name: "REDIS_URL ausente", envKey: envRedisURL, value: ""},
		{name: "REDIS_URL no parseable", envKey: envRedisURL, value: "no-es-una-url"},
		{name: "REDIS_URL esquema equivocado", envKey: envRedisURL, value: "http://localhost:6379"},
		{name: "SESSION_COOKIE_SECURE no booleano", envKey: envSessionCookieSecure, value: "quizá"},
		{name: "SESSION_IDLE_TTL_MINUTES no numérico", envKey: envSessionIdleTTLMinutes, value: "treinta"},
		{name: "SESSION_IDLE_TTL_MINUTES cero", envKey: envSessionIdleTTLMinutes, value: "0"},
		{name: "SESSION_IDLE_TTL_MINUTES negativo", envKey: envSessionIdleTTLMinutes, value: "-5"},
		{name: "SESSION_ABSOLUTE_TTL_MINUTES cero", envKey: envSessionAbsoluteTTLMinutes, value: "0"},
		{name: "SESSION_ABSOLUTE_TTL_MINUTES negativo", envKey: envSessionAbsoluteTTLMinutes, value: "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			if tt.envKey != envRedisURL {
				setRedis(t)
			}
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

// TestLoadProductionRequiresSecureCookie fija R11: en producción una cookie de
// sesión sin `Secure` impide el arranque; en desarrollo es válida.
func TestLoadProductionRequiresSecureCookie(t *testing.T) {
	t.Run("producción con cookie insegura falla", func(t *testing.T) {
		clearEnv(t)
		setRedis(t)
		t.Setenv(envAppEnv, EnvProduction)
		t.Setenv(envSessionCookieSecure, "false")
		t.Setenv(envSessionSecret, testSessionSecret)

		_, err := Load()
		if err == nil {
			t.Fatalf("Load() no devolvió error en producción con %s=false", envSessionCookieSecure)
		}
		if !strings.Contains(err.Error(), envSessionCookieSecure) {
			t.Errorf("el error %q no identifica la variable %q", err.Error(), envSessionCookieSecure)
		}
	})

	t.Run("producción con cookie segura arranca", func(t *testing.T) {
		clearEnv(t)
		setRedis(t)
		t.Setenv(envAppEnv, EnvProduction)
		t.Setenv(envSessionCookieSecure, "true")
		t.Setenv(envSessionSecret, testSessionSecret)

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() devolvió error en producción con cookie segura: %v", err)
		}
		if !cfg.SessionCookieSecure {
			t.Errorf("SessionCookieSecure = false, se esperaba true")
		}
	})
}

// TestLoadSessionSecretValidation fija M-1: en producción SESSION_SECRET vacío
// o corto impide el arranque; en desarrollo avisa pero permite trabajar.
func TestLoadSessionSecretValidation(t *testing.T) {
	t.Run("producción sin secreto falla", func(t *testing.T) {
		clearEnv(t)
		setRedis(t)
		t.Setenv(envAppEnv, EnvProduction)
		t.Setenv(envSessionCookieSecure, "true")
		t.Setenv(envSessionSecret, "")

		_, err := Load()
		if err == nil {
			t.Fatalf("Load() no devolvió error en producción con %s vacío", envSessionSecret)
		}
		if !strings.Contains(err.Error(), envSessionSecret) {
			t.Errorf("el error %q no identifica la variable %q", err.Error(), envSessionSecret)
		}
	})

	t.Run("producción con secreto corto falla", func(t *testing.T) {
		clearEnv(t)
		setRedis(t)
		t.Setenv(envAppEnv, EnvProduction)
		t.Setenv(envSessionCookieSecure, "true")
		t.Setenv(envSessionSecret, strings.Repeat("s", minSessionSecretLength-1))

		_, err := Load()
		if err == nil {
			t.Fatalf("Load() no devolvió error en producción con %s corto", envSessionSecret)
		}
		if !strings.Contains(err.Error(), envSessionSecret) {
			t.Errorf("el error %q no identifica la variable %q", err.Error(), envSessionSecret)
		}
	})

	t.Run("producción con secreto fuerte arranca sin aviso", func(t *testing.T) {
		clearEnv(t)
		setRedis(t)
		t.Setenv(envAppEnv, EnvProduction)
		t.Setenv(envSessionCookieSecure, "true")
		t.Setenv(envSessionSecret, testSessionSecret)

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() devolvió error con un %s válido: %v", envSessionSecret, err)
		}
		if warnsAbout(cfg.Warnings, envSessionSecret) {
			t.Errorf("no debería avisar de %s con un valor válido: %v", envSessionSecret, cfg.Warnings)
		}
	})

	t.Run("desarrollo sin secreto avisa y arranca", func(t *testing.T) {
		clearEnv(t)
		setRedis(t)

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() devolvió error en desarrollo: %v", err)
		}
		if cfg.SessionSecret != "" {
			t.Errorf("SessionSecret = %q, se esperaba vacío", cfg.SessionSecret)
		}
		if !warnsAbout(cfg.Warnings, envSessionSecret) {
			t.Errorf("no avisó de %s débil en desarrollo: %v", envSessionSecret, cfg.Warnings)
		}
	})
}

// warnsAbout indica si algún aviso menciona la variable indicada.
func warnsAbout(warnings []string, key string) bool {
	return strings.Contains(strings.Join(warnings, "\n"), key)
}

// TestLoadIgnoresUnknownKeys fija que HTTP_ADDR no existe: definirla no cambia
// la configuración y el servidor sigue escuchando en :HTTP_PORT.
func TestLoadIgnoresUnknownKeys(t *testing.T) {
	clearEnv(t)
	setRedis(t)
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
