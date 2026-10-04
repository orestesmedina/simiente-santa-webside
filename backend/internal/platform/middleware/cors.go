package middleware

import (
	"net/http"
	"strings"

	"simiente-santa/backend/internal/platform/httpserver"
)

// Cabeceras CORS que maneja el middleware mínimo de F1 (D16): una respuesta
// simple sin credenciales y los preflight.
const (
	headerOrigin        = "Origin"
	headerAllowOrigin   = "Access-Control-Allow-Origin"
	headerAllowMethods  = "Access-Control-Allow-Methods"
	headerAllowHeaders  = "Access-Control-Allow-Headers"
	headerMaxAge        = "Access-Control-Max-Age"
	headerRequestMethod = "Access-Control-Request-Method"
	headerVary          = "Vary"
	allowMethodsValue   = "GET, OPTIONS"
	allowHeadersValue   = "Content-Type, X-Request-ID"
	allowMaxAgeValue    = "600"
)

// CORS responde los preflight OPTIONS y fija Access-Control-Allow-Origin solo
// para los orígenes permitidos (CORS_ALLOWED_ORIGINS). Escrito a mano, sin
// dependencias. No usa credenciales en F1 (D16).
func CORS(allowedOrigins []string) httpserver.Middleware {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if origin = strings.TrimSpace(origin); origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get(headerOrigin)
			_, permitted := allowed[origin]
			permitted = permitted && origin != ""

			if origin != "" {
				w.Header().Add(headerVary, headerOrigin)
			}
			if permitted {
				w.Header().Set(headerAllowOrigin, origin)
			}

			if isPreflight(r) {
				if permitted {
					w.Header().Set(headerAllowMethods, allowMethodsValue)
					w.Header().Set(headerAllowHeaders, allowHeadersValue)
					w.Header().Set(headerMaxAge, allowMaxAgeValue)
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isPreflight distingue un preflight real (OPTIONS con
// Access-Control-Request-Method) de un OPTIONS normal.
func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && r.Header.Get(headerRequestMethod) != ""
}
