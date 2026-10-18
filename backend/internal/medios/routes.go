package medios

import (
	"net/http"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/middleware"
)

// routes.go publica la superficie HTTP del dominio. Se mantiene separada de los
// handlers para que el reparto público/panel sea auditable de un vistazo (patrón
// de F2): la superficie pública queda en RegisterPublic y la de panel en
// RegisterAdmin, esta última tras la cadena aprobada con el permiso `medios`.

// PermissionMedios es el permiso de módulo del catálogo de F2 que activa F9
// (FR-021/P9-15): «Medios (prédicas y podcasts)». Es un `code` estable sembrado
// por la migración 000002; **F9 no añade ningún permiso** ni toca el catálogo de
// F2 (sigue en 10) y tampoco lo renombra.
const PermissionMedios = "medios"

// RegisterPublic publica la superficie pública de medios, SIN middlewares de
// sesión (solo los globales de la aplicación): es de solo lectura y solo sirve
// contenido publicable (FR-002/FR-009). No hay rate-limit: no hay superficie
// pública escribible.
//
// La segunda ruta cuelga del segmento `portada` (el agregado de F3) pero la sirve
// este dominio (Complexity 5 del plan): así el contrato no toca ninguna operación
// de F3 y la portada no se rediseña.
func RegisterPublic(root httpserver.Registrar, h *Handler) {
	if h == nil {
		return
	}
	// Sección pública completa (FR-002…FR-006). La redirección de la ruta SPA
	// `/medios` es del frontend: la API responde 200 con `items` vacío (P9-16).
	root.Handle(http.MethodGet, "/api/v1/medios", h.GetMedios)
	// Bloque de la portada: hasta 5 fichas publicables, con el tope firme en el
	// servidor (US6/FR-027).
	root.Handle(http.MethodGet, "/api/v1/portada/medios", h.GetPortadaMedios)
}

// AdminDeps reúne la configuración del grupo de panel de medios.
type AdminDeps struct {
	// Deps son las dependencias de la cadena del panel (middleware.AdminDeps); el
	// Recorder debe ser el service de F9 (T922) para registrar las denegaciones
	// del módulo.
	Deps middleware.AdminDeps
}

// RegisterAdmin publica el grupo /api/v1/admin/medios con la cadena aprobada de
// F2 —authn → guard de cambio de contraseña → authz(`medios`) → CSRF—, **sin
// helpers nuevos de middleware** (P9-15): el permiso ya existe en el catálogo y se
// verifica en el servidor, de forma independiente de los demás permisos de módulo
// (FR-021/SC-009). Las denegaciones las registra el Recorder de la cadena con la
// acción de la ruta (T922).
//
// F9 no tiene alta manual ni borrado (FR-019): el grupo publica el listado, la
// única automatización (`buscar-nuevos`), la edición del `PATCH` y las dos
// operaciones de revisión (unir y resolver marca).
func RegisterAdmin(root httpserver.Registrar, h *Handler, deps AdminDeps) {
	if h == nil {
		return
	}
	group := root.Group("/api/v1/admin/medios", middleware.AdminChain(PermissionMedios, deps.Deps)...)

	// Listado del panel (FR-019/FR-018).
	group.Handle(http.MethodGet, "", h.ListMedios)
	// «Buscar episodios nuevos»: la única automatización (FR-011).
	group.Handle(http.MethodPost, "/buscar-nuevos", h.BuscarNuevos)
	// Edición de la descripción corta y la visibilidad (FR-019/FR-020).
	group.Handle(http.MethodPatch, "/{id}", h.UpdateMedio)
	// Unión por puntero y resolución de marcas (FR-014/FR-017).
	group.Handle(http.MethodPost, "/{id}/unir", h.UnirMedios)
	group.Handle(http.MethodPost, "/{id}/marcas/resolver", h.ResolverMarca)
}
