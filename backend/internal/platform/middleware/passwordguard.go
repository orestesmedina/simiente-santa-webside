package middleware

import (
	"log/slog"
	"net/http"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/session"
)

// passwordChangeRequiredReason es el `details.reason` con el que el guard
// señala a la UI que debe redirigir al formulario de cambio (FR-010/US7 esc.
// 4–5).
const passwordChangeRequiredReason = "password_change_required"

// PasswordGuard impide usar el panel a una cuenta que tiene pendiente el cambio
// obligatorio de contraseña (FR-010/US7 esc. 4–5): con `mustChangePassword`,
// toda petición del grupo donde se monte responde 403 con
// `details.reason = "password_change_required"`. Se monta SOLO en el grupo
// `/api/v1/admin` (entre authn y authz; lo cablea routes.go), de modo que las
// rutas `/api/v1/auth/session`, `/auth/logout` y `/auth/password` —que viven en
// el grupo `/api/v1/auth` y no lo montan— quedan disponibles para que la
// persona vea su sesión, salga y cambie la contraseña. No es una whitelist
// genérica sobre `/auth/*`.
//
// Lee la identidad resuelta del contexto (la deja authn, que va antes). Sin
// identidad responde 401 como red de seguridad: el guard nunca deja pasar una
// petición de panel sin autenticar.
func PasswordGuard(logger *slog.Logger) httpserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			identity, ok := session.IdentityFromContext(ctx)
			if !ok {
				httpserver.WriteError(ctx, w, logger, apperr.Unauthenticated(messageUnauthenticated))
				return
			}
			if identity.MustChangePassword {
				httpserver.WriteError(ctx, w, logger, apperr.Forbidden(
					"Debes cambiar tu contraseña antes de usar el panel",
					apperr.WithDetails(map[string]any{"reason": passwordChangeRequiredReason}),
				))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
