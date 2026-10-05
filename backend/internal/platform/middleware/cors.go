package middleware

import (
	"net/http"
	"strings"

	"simiente-santa/backend/internal/platform/httpserver"
)

// Cabeceras CORS (D16 ampliado en P11/R13): la sesión viaja en cookie, así que
// las respuestas van con credenciales y el Origin se refleja exacto (nunca `*`
// con credenciales).
const (
	headerOrigin           = "Origin"
	headerAllowOrigin      = "Access-Control-Allow-Origin"
	headerAllowMethods     = "Access-Control-Allow-Methods"
	headerAllowHeaders     = "Access-Control-Allow-Headers"
	headerAllowCredentials = "Access-Control-Allow-Credentials"
	headerExposeHeaders    = "Access-Control-Expose-Headers"
	headerMaxAge           = "Access-Control-Max-Age"
	headerRequestMethod    = "Access-Control-Request-Method"
	headerVary             = "Vary"
	allowMethodsValue      = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	allowHeadersValue      = "Content-Type, X-CSRF-Token, X-Request-ID"
	exposeHeadersValue     = "X-Request-ID"
	allowCredentialsValue  = "true"
	allowMaxAgeValue       = "600"
)

// CORS responde los preflight OPTIONS y fija las cabeceras CORS solo para los
// orígenes permitidos (CORS_ALLOWED_ORIGINS). Escrito a mano, sin dependencias.
//
// A diferencia de F1, la respuesta admite credenciales (P11): echo exacto del
// Origin permitido —nunca `*`, incompatible con cookies—, Allow-Credentials,
// las cabeceras del contrato (Content-Type, X-CSRF-Token, X-Request-ID) y
// Expose-Headers para que la SPA pueda leer X-Request-ID. Vary: Origin evita
// que una caché sirva la respuesta de un origen a otro.
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
				w.Header().Set(headerAllowCredentials, allowCredentialsValue)
				w.Header().Set(headerExposeHeaders, exposeHeadersValue)
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
