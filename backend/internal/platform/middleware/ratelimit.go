package middleware

import (
	"log/slog"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
)

// Parámetros por defecto del rate-limit de P17/R12: 20 peticiones por minuto
// por IP sobre los endpoints públicos escribibles.
const (
	// DefaultRateLimit es el umbral por IP dentro de la ventana.
	DefaultRateLimit = 20
	// DefaultRateWindow es la ventana deslizante.
	DefaultRateWindow = time.Minute
)

// messageRateLimited es el 429 del rate-limit.
const messageRateLimited = "Demasiadas solicitudes. Inténtalo de nuevo en un momento"

// cleanupEvery indica cada cuántas admisiones se barren las claves sin
// actividad: limpieza oportunista para acotar la memoria del mapa sin un
// proceso en segundo plano (R12).
const cleanupEvery = 128

// DefaultRateLimitPaths devuelve las rutas públicas escribibles de R12
// (`POST /api/v1/auth/login` y `POST /api/v1/setup/initialize`). Se devuelve
// una copia en cada llamada para no exponer estado de paquete mutable (R6).
func DefaultRateLimitPaths() []string {
	return []string{"/api/v1/auth/login", "/api/v1/setup/initialize"}
}

// RateLimitConfig configura la ventana deslizante por IP. Los valores cero usan
// los de R12.
type RateLimitConfig struct {
	// Limit es el número máximo de peticiones por IP y ruta dentro de la
	// ventana. <= 0 → DefaultRateLimit (20).
	Limit int
	// Window es la duración de la ventana deslizante. <= 0 →
	// DefaultRateWindow (1 minuto).
	Window time.Duration
	// Paths restringe el middleware a esas rutas exactas; vacío aplica a toda
	// petición que lo atraviese (cuando se monta ya acotado por grupo). Si se
	// indica, usa DefaultRateLimitPaths() para el alcance de R12.
	Paths []string
	// Now inyecta el reloj (pruebas de ventana). nil → time.Now.
	Now func() time.Time
	// ClientIP extrae la IP del cliente. nil → host de RemoteAddr, sin confiar
	// en cabeceras como X-Forwarded-For (falsificables sin un proxy de
	// confianza; ver nota).
	ClientIP func(*http.Request) string
}

// RateLimit aplica una ventana deslizante en memoria por IP (P17/R12): cuando
// una IP supera el umbral de peticiones en la ventana, responde 429
// rate_limited con Retry-After (segundos hasta que la petición más antigua
// salga de la ventana). Es un middleware propio, sin dependencias, pensado para
// los endpoints públicos escribibles: amortigua la prueba masiva de
// contraseñas y el bombardeo del endpoint de inicialización.
//
// Limitación declarada (RG5): el estado vive en memoria y no se comparte entre
// instancias; hoy hay una sola. En un despliegue tras un proxy inverso, la IP
// es la del proxy salvo que se configure un ClientIP de confianza.
func RateLimit(cfg RateLimitConfig, logger *slog.Logger) httpserver.Middleware {
	limit := cfg.Limit
	if limit <= 0 {
		limit = DefaultRateLimit
	}
	window := cfg.Window
	if window <= 0 {
		window = DefaultRateWindow
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	clientIP := cfg.ClientIP
	if clientIP == nil {
		clientIP = remoteIP
	}

	var paths map[string]struct{}
	if len(cfg.Paths) > 0 {
		paths = make(map[string]struct{}, len(cfg.Paths))
		for _, p := range cfg.Paths {
			paths[p] = struct{}{}
		}
	}

	windowState := newSlidingWindow(limit, window, now)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if paths != nil {
				if _, limited := paths[r.URL.Path]; !limited {
					next.ServeHTTP(w, r)
					return
				}
			}

			key := r.URL.Path + "\x00" + clientIP(r)
			retryAfter, allowed := windowState.allow(key)
			if !allowed {
				httpserver.WriteError(r.Context(), w, logger, apperr.RateLimited(
					messageRateLimited,
					apperr.WithRetryAfter(retryAfterSeconds(retryAfter)),
				))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// slidingWindow es el estado en memoria de la ventana deslizante. No es estado
// de paquete: cada RateLimit construye el suyo (R6).
type slidingWindow struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	now    func() time.Time
	hits   map[string][]time.Time
	calls  int
}

func newSlidingWindow(limit int, window time.Duration, now func() time.Time) *slidingWindow {
	return &slidingWindow{
		limit:  limit,
		window: window,
		now:    now,
		hits:   make(map[string][]time.Time),
	}
}

// allow poda las marcas fuera de la ventana y decide si la petición pasa. Si se
// deniega, devuelve el tiempo hasta que la marca más antigua caduque.
func (s *slidingWindow) allow(key string) (time.Duration, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	cutoff := now.Add(-s.window)
	hits := s.hits[key]

	kept := hits[:0]
	for _, at := range hits {
		if at.After(cutoff) {
			kept = append(kept, at)
		}
	}
	hits = kept

	if len(hits) >= s.limit {
		s.hits[key] = hits
		retryAfter := hits[0].Add(s.window).Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return retryAfter, false
	}

	s.hits[key] = append(hits, now)
	s.calls++
	if s.calls%cleanupEvery == 0 {
		s.sweep(cutoff)
	}
	return 0, true
}

// sweep elimina las claves sin marcas vivas. Se llama con el mutex tomado.
func (s *slidingWindow) sweep(cutoff time.Time) {
	for key, hits := range s.hits {
		alive := false
		for _, at := range hits {
			if at.After(cutoff) {
				alive = true
				break
			}
		}
		if !alive {
			delete(s.hits, key)
		}
	}
}

// remoteIP devuelve el host de RemoteAddr. No usa X-Forwarded-For: sin un proxy
// de confianza, esa cabecera es falsificable y permitiría saltarse el umbral.
func remoteIP(r *http.Request) string {
	addr := strings.TrimSpace(r.RemoteAddr)
	if addr == "" {
		return "unknown"
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

// retryAfterSeconds redondea hacia arriba los segundos de espera, con un mínimo
// de 1 para no anunciar un reintento inmediato cuando aún queda ventana.
func retryAfterSeconds(d time.Duration) int {
	if d <= 0 {
		return 1
	}
	seconds := int(math.Ceil(d.Seconds()))
	if seconds < 1 {
		return 1
	}
	return seconds
}
