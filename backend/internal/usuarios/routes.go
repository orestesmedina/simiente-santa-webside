package usuarios

import (
	"net/http"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/middleware"
)

// routes.go publica la superficie HTTP del dominio. Se mantiene separada de los
// handlers para que el reparto público/panel sea auditable de un vistazo: la
// superficie pública queda en RegisterPublic y la de panel en RegisterAdmin.

// PermissionAdminUsersRoles es el permiso de módulo que exige el grupo de panel
// en F2 (FR-015/FR-016): administrar usuarios y roles, y consultar la auditoría
// (FR-024). Es un `code` estable del catálogo sembrado por la migración 000002.
const PermissionAdminUsersRoles = "admin_usuarios_roles"

// PublicDeps reúne los middlewares de las rutas públicas de acceso (T228).
type PublicDeps struct {
	// Login se monta solo en POST /api/v1/auth/login: rate-limit por IP sobre
	// la superficie pública escribible (P17). No incluye authn: es la ruta de
	// entrada.
	Login []httpserver.Middleware
	// Session se monta en las rutas de sesión (/auth/session, /auth/logout y
	// /auth/password): authn → CSRF y NUNCA el guard de cambio obligatorio, de
	// modo que una cuenta con mustChangePassword pueda ver su sesión, salir y
	// cambiar la contraseña (rutas blanqueadas; F-15).
	Session []httpserver.Middleware
}

// AdminDeps reúne la configuración del grupo de panel (T228).
type AdminDeps struct {
	// Module es el permiso que exige authz (en F2, admin_usuarios_roles).
	Module string
	// Deps son las dependencias de la cadena del panel, que arma
	// middleware.AdminChain en el orden aprobado.
	Deps middleware.AdminDeps
	// Routes publica las rutas de administración dentro del grupo ya protegido.
	// T234/T237/T238 las añaden; puede ser nil mientras no existan.
	Routes func(httpserver.Registrar)
}

// RegisterPublic publica la superficie de acceso del dominio: login (público,
// con rate-limit) y las rutas de sesión (authn → CSRF, sin guard). La cadena
// completa por grupo es la del plan §"Cadena de middleware".
func RegisterPublic(root httpserver.Registrar, h *Handler, deps PublicDeps) {
	login := root.Group("/api/v1/auth", deps.Login...)
	login.Handle(http.MethodPost, "/login", h.Login)

	session := root.Group("/api/v1/auth", deps.Session...)
	session.Handle(http.MethodGet, "/session", h.GetSession)
	session.Handle(http.MethodPost, "/logout", h.Logout)
	session.Handle(http.MethodPost, "/password", h.ChangePassword)
}

// RegisterAdmin publica el grupo /api/v1/admin con la cadena aprobada (arq. §6):
// authn → guard de cambio de contraseña → authz(módulo) → CSRF. El guard se
// monta AQUÍ y solo aquí: las rutas de /auth/* quedan blanqueadas para que una
// cuenta con mustChangePassword pueda resolverlo.
func RegisterAdmin(root httpserver.Registrar, deps AdminDeps) {
	group := root.Group("/api/v1/admin", middleware.AdminChain(deps.Module, deps.Deps)...)
	if deps.Routes != nil {
		deps.Routes(group)
	}
}
