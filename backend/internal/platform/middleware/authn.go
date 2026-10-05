package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/session"
)

// messageUnauthenticated es el 401 genérico de la cadena de sesión: el mismo
// para «no hay cookie», «cookie inválida», «sesión expirada» y «cuenta
// desactivada», de modo que no se revela el motivo (FR-003/FR-012). Lo
// comparten authn y el guard de cambio de contraseña.
const messageUnauthenticated = "Necesitas iniciar sesión para continuar"

// Authn resuelve la sesión de la cookie ss_session y deja la identidad en el
// contexto de la petición (P1/P9, arq. §6):
//
//  1. Sin cookie → 401 unauthenticated.
//  2. `session.Store.Resolve` valida el token en Redis y refresca su TTL de
//     inactividad; un token desconocido o expirado por inactividad devuelve
//     ErrSessionNotFound → 401. La vida absoluta se comprueba además aquí de
//     forma defensiva (ExpiredAt): una sesión vencida por hora máxima no pasa
//     aunque el store la tuviera viva.
//  3. `session.Resolver` (lo implementa el dominio usuarios) resuelve los
//     permisos vigentes del rol y exige que la cuenta siga activa (FR-012); un
//     error de resolución (cuenta desactivada o ya inexistente) → 401.
//
// Se monta como primer middleware de todo grupo con sesión; va antes que el
// guard de cambio de contraseña, authz y CSRF. No toca la base de datos: eso es
// responsabilidad del Resolver (R1/R4).
func Authn(store session.Store, resolver session.Resolver, logger *slog.Logger) httpserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			token, ok := session.SessionTokenFromRequest(r)
			if !ok {
				writeUnauthenticated(ctx, w, logger)
				return
			}

			sess, err := store.Resolve(ctx, token)
			if err != nil || sess.ExpiredAt(time.Now()) {
				writeUnauthenticated(ctx, w, logger)
				return
			}

			identity, err := resolver.Resolve(ctx, sess.UserID)
			if err != nil {
				writeUnauthenticated(ctx, w, logger)
				return
			}

			next.ServeHTTP(w, r.WithContext(session.ContextWithIdentity(ctx, identity)))
		})
	}
}

// writeUnauthenticated responde el 401 genérico de sesión.
func writeUnauthenticated(ctx context.Context, w http.ResponseWriter, logger *slog.Logger) {
	httpserver.WriteError(ctx, w, logger, apperr.Unauthenticated(messageUnauthenticated))
}
