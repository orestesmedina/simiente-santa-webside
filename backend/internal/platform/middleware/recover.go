package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
)

// Recover convierte cualquier panic de la cadena interna en un 500 con sobre de
// error: el proceso nunca cae por una petición. El stack completo va al log
// estructurado con el request_id; nunca a la respuesta.
func Recover(logger *slog.Logger) httpserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					cause := fmt.Errorf("panic recuperado: %v\n%s", rec, debug.Stack())
					httpserver.WriteError(r.Context(), w, logger, apperr.Internal(cause))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
