package middleware

import (
	"log/slog"
	"net/http"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/session"
)

// messageCSRF es el 403 cuando falta o no cuadra el double-submit firmado.
const messageCSRF = "El token de seguridad no es válido o ha caducado"

// safeMethods son los métodos considerados seguros (no mutan estado) y que, por
// tanto, no exigen comprobación CSRF (RFC 9110). OPTIONS entra aquí porque los
// preflight los corta CORS antes de llegar a este middleware.
var safeMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodOptions: {},
	http.MethodTrace:   {},
}

// CSRF aplica el double-submit firmado de P10: en todo método NO seguro exige
// que la cabecera X-CSRF-Token coincida con la cookie csrf_token y que esta
// lleve una firma HMAC válida (session.ValidateCSRF). Sin par válido → 403
// forbidden. Los métodos seguros pasan sin comprobación.
//
// No comprueba la sesión: se monta dentro de un grupo con sesión, después de
// authn, así que la petición ya está autenticada. Vive en la cadena de panel y
// en la de las rutas de sesión (/api/v1/auth), nunca en las públicas sin sesión
// (login e inicialización), que no tienen token que proteger (R3 del research).
func CSRF(secret string, logger *slog.Logger) httpserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isSafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			cookieToken, _ := session.CSRFFromRequest(r)
			headerToken := r.Header.Get(session.HeaderCSRF)
			if !session.ValidateCSRF(secret, cookieToken, headerToken) {
				httpserver.WriteError(r.Context(), w, logger, apperr.Forbidden(messageCSRF))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isSafeMethod indica si el método no muta estado y queda exento de CSRF.
func isSafeMethod(method string) bool {
	_, safe := safeMethods[method]
	return safe
}
