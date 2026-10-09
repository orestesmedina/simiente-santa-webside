package middleware

import (
	"log/slog"

	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/session"
)

// Chain arma la cadena transversal de F1 en el orden aprobado (el primero, el
// más externo): request-id → recover → logging → CORS. Es el único constructor
// de la composición: lo usan cmd/api en producción y testutil en las pruebas,
// para que ambas no puedan divergir (FR-011). platform no conoce dominios (R1).
func Chain(logger *slog.Logger, allowedOrigins []string) []httpserver.Middleware {
	return []httpserver.Middleware{
		RequestID,
		Recover(logger),
		Logging(logger),
		CORS(allowedOrigins),
	}
}

// AdminDeps reúne las dependencias de la cadena del grupo de panel. Las
// interfaces viven en platform/session y platform/audit; las implementa el
// dominio usuarios (R3).
type AdminDeps struct {
	// Sessions resuelve la sesión de la cookie (session.Store, Redis).
	Sessions session.Store
	// Resolver resuelve la identidad con los permisos vigentes (session.Resolver).
	Resolver session.Resolver
	// Recorder registra la denegación de permiso (audit.Recorder). Puede ser
	// nil: la denegación se aplica igual, solo no se persiste.
	Recorder audit.Recorder
	// CSRFSecret firma la cookie CSRF (SESSION_SECRET).
	CSRFSecret string
	// Logger para los errores de la cadena.
	Logger *slog.Logger
}

// AdminChain arma la cadena del grupo de panel en el orden aprobado
// (plan §"Cadena de middleware", arq. §6):
//
//	authn → guard de cambio de contraseña → authz(módulo) → CSRF
//
// El guard se monta aquí, en el grupo de panel, y nunca en las rutas de sesión
// `/api/v1/auth` (que quedan blanqueadas para una cuenta con mustChangePassword;
// ver PasswordGuard). module es el permiso de módulo que exige authz (en F2,
// "admin_usuarios_roles"). Es el punto único de composición para que cmd/api y
// las pruebas no diverjan en el orden.
func AdminChain(module string, deps AdminDeps) []httpserver.Middleware {
	return []httpserver.Middleware{
		Authn(deps.Sessions, deps.Resolver, deps.Logger),
		PasswordGuard(deps.Logger),
		AuthzByModule(module, deps.Recorder, deps.Logger),
		CSRF(deps.CSRFSecret, deps.Logger),
	}
}
