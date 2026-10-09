package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/session"
)

// messageForbidden es el 403 por falta de permiso del módulo. Es genérico a
// propósito: no revela qué permiso concreto falta.
const messageForbidden = "No tienes permiso para acceder a este módulo"

// AuthzByModule exige que el rol de la identidad resuelta incluya el permiso
// del módulo code (P16/FR-016, «authz» de la cadena de arq. §6). Sin permiso
// responde 403 y registra el intento como denegación en `admin_actions` con
// result='denied' a través de audit.Recorder (P20): el middleware no conoce los
// códigos de acción, solo entrega al dominio el actor, el método y la ruta para
// que este resuelva la acción y el objetivo.
//
// Se monta SIEMPRE después de authn: sin identidad en el contexto responde 401
// como red de seguridad (nunca deja pasar una petición de panel sin autenticar)
// y nunca se da por bueno un rol ausente. Va antes de CSRF (el orden aprobado:
// authn → guard de contraseña → authz → CSRF).
//
// recorder puede ser nil en pruebas unitarias: en ese caso la denegación se
// aplica igual (fail-closed) pero no se persiste. Un fallo al registrar tampoco
// cambia la respuesta: la petición ya está denegada.
func AuthzByModule(code string, recorder audit.Recorder, logger *slog.Logger) httpserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			identity, ok := session.IdentityFromContext(ctx)
			if !ok {
				writeUnauthenticated(ctx, w, logger)
				return
			}
			if !identity.HasPermission(code) {
				recordDenied(ctx, recorder, logger, identity.UserID, r)
				httpserver.WriteError(ctx, w, logger, apperr.Forbidden(messageForbidden))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// recordDenied registra la denegación de permiso de forma best-effort: si el
// registro falla, se anota en el log pero la denegación ya está decidida. El
// actor es la cuenta autenticada; el dominio resuelve acción y objetivo desde
// method+path (P20/R23).
func recordDenied(ctx context.Context, recorder audit.Recorder, logger *slog.Logger, actor uuid.UUID, r *http.Request) {
	if recorder == nil {
		return
	}
	denial := audit.Denial{
		ActorUserID: &actor,
		Method:      r.Method,
		Path:        r.URL.Path,
	}
	if err := recorder.RecordDenied(ctx, denial); err != nil {
		loggerFor(ctx, logger).Warn("no se pudo registrar la denegación de permiso", slog.String("error", err.Error()))
	}
}

// loggerFor devuelve el logger por petición (con request_id, método y ruta) si
// está en el contexto; si no, el logger recibido. Evita duplicar los campos
// cuando logging ya publicó el suyo.
func loggerFor(ctx context.Context, logger *slog.Logger) *slog.Logger {
	if reqLogger := httpserver.RequestLoggerFromContext(ctx); reqLogger != nil {
		return reqLogger
	}
	if logger == nil {
		return slog.Default()
	}
	return logger
}
