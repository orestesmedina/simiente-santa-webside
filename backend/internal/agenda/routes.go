package agenda

import (
	"net/http"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/middleware"
)

// routes.go publica la superficie HTTP del dominio. Se mantiene separada de los
// handlers para que el reparto público/panel sea auditable de un vistazo (patrón
// de F2): la superficie pública queda en RegisterPublic y los tres grupos de
// panel en RegisterAdmin, cada uno tras la cadena aprobada con SU permiso.

// Permisos del catálogo de F2 que activa F4 (FR-012/P4-5). Son `code` estables
// sembrados por la migración 000002 (`eventos`, `actividades`) y por la 000009
// (`categorias`, el 10.º del catálogo: cambio de alcance aprobado el
// 2026-10-11). Los tres son **independientes entre sí**: tener uno no da acceso a
// los otros.
const (
	// PermissionEventos autoriza el grupo /api/v1/admin/eventos.
	PermissionEventos = "eventos"
	// PermissionActividades autoriza el grupo /api/v1/admin/actividades.
	PermissionActividades = "actividades"
	// PermissionCategorias autoriza el grupo /api/v1/admin/categorias.
	PermissionCategorias = "categorias"
)

// RegisterPublic publica la superficie pública de la agenda, SIN middlewares de
// sesión (solo los globales de la aplicación): es de solo lectura y solo sirve
// contenido publicado y próximo (FR-001/FR-013).
func RegisterPublic(root httpserver.Registrar, h *Handler) {
	if h == nil {
		return
	}
	// «Lo que viene»: los próximos eventos y las próximas actividades (FR-001…FR-009).
	root.Handle(http.MethodGet, "/api/v1/proximamente", h.GetProximamente)
	// Extracto del bloque «Lo que viene» de la portada (US7/FR-025…FR-027, Q8).
	// El segmento `portada` es el agregado de F3, pero el recurso es de la agenda
	// (Complexity 8 del plan): así el contrato no toca ninguna operación de F3 y
	// la portada no se rediseña (FR-027).
	root.Handle(http.MethodGet, "/api/v1/portada/proximamente", h.GetPortadaProximamente)
}

// AdminDeps reúne la configuración de los grupos de panel de la agenda.
type AdminDeps struct {
	// Deps son las dependencias de la cadena del panel (middleware.AdminDeps);
	// el Recorder debe ser el service de F4 (T422) para registrar las
	// denegaciones de los tres módulos (`event.*`, `activity.*`, `category.*`).
	Deps middleware.AdminDeps
}

// RegisterAdmin publica los tres grupos de panel con la cadena aprobada de F2
// —authn → guard de cambio de contraseña → authz(permiso del módulo) → CSRF—,
// **sin helpers nuevos de middleware** (P4-5): cada grupo exige SU permiso y las
// denegaciones las registra el Recorder de la cadena (result='denied').
//
//	/api/v1/admin/eventos      → AdminChain("eventos")
//	/api/v1/admin/actividades  → AdminChain("actividades")
//	/api/v1/admin/categorias   → AdminChain("categorias")
//
// Los tres permisos son independientes: la autorización se verifica en el
// servidor (FR-012/SC-005/SC-011).
func RegisterAdmin(root httpserver.Registrar, h *Handler, deps AdminDeps) {
	if h == nil {
		return
	}

	// Eventos (FR-010/FR-018).
	eventos := root.Group("/api/v1/admin/eventos", middleware.AdminChain(PermissionEventos, deps.Deps)...)
	eventos.Handle(http.MethodGet, "", h.ListEventos)
	eventos.Handle(http.MethodPost, "", h.CreateEvento)
	eventos.Handle(http.MethodPost, "/imagenes", h.UploadEventoImage)
	eventos.Handle(http.MethodGet, "/{id}", h.GetEvento)
	eventos.Handle(http.MethodPatch, "/{id}", h.UpdateEvento)
	eventos.Handle(http.MethodDelete, "/{id}", h.DeleteEvento)

	// Actividades (FR-011/FR-019).
	actividades := root.Group("/api/v1/admin/actividades", middleware.AdminChain(PermissionActividades, deps.Deps)...)
	actividades.Handle(http.MethodGet, "", h.ListActividades)
	actividades.Handle(http.MethodPost, "", h.CreateActividad)
	actividades.Handle(http.MethodPost, "/imagenes", h.UploadActividadImage)
	actividades.Handle(http.MethodGet, "/{id}", h.GetActividad)
	actividades.Handle(http.MethodPatch, "/{id}", h.UpdateActividad)
	actividades.Handle(http.MethodDelete, "/{id}", h.DeleteActividad)

	// Catálogo de categorías (Q6/Q7): permiso propio `categorias`.
	categorias := root.Group("/api/v1/admin/categorias", middleware.AdminChain(PermissionCategorias, deps.Deps)...)
	categorias.Handle(http.MethodGet, "", h.ListCategorias)
	categorias.Handle(http.MethodPost, "", h.CreateCategoria)
	categorias.Handle(http.MethodPatch, "/{id}", h.UpdateCategoria)
	categorias.Handle(http.MethodDelete, "/{id}", h.DeleteCategoria)
}
