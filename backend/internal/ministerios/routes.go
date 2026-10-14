package ministerios

import (
	"net/http"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/middleware"
)

// routes.go publica la superficie HTTP del dominio. Se mantiene separada de los
// handlers para que el reparto público/panel sea auditable de un vistazo (patrón
// de F2): la superficie pública queda en RegisterPublic y la de panel en
// RegisterAdmin, esta última tras la cadena aprobada con el permiso `ministerios`.

// PermissionMinisterios es el permiso de módulo del catálogo de F2 que activa F6
// (FR-011/P6-6): "Ministerios". Es un `code` estable sembrado por la migración
// 000002; **F6 no añade ningún permiso** ni toca el catálogo de F2.
const PermissionMinisterios = "ministerios"

// RegisterPublic publica la superficie pública de ministerios, SIN middlewares
// de sesión (solo los globales de la aplicación): es de solo lectura y solo sirve
// contenido publicado (FR-002/FR-012/FR-026).
//
// La segunda ruta cuelga del segmento `portada` (el agregado de F3) pero la sirve
// este dominio (Complexity 4 del plan): así el contrato no toca ninguna operación
// de F3 y la portada no se rediseña.
func RegisterPublic(root httpserver.Registrar, h *Handler) {
	if h == nil {
		return
	}
	// Sección pública completa (FR-002…FR-008).
	root.Handle(http.MethodGet, "/api/v1/ministerios", h.GetMinisterios)
	// Bloque de la portada: hasta 5 ítems publicados, con el tope en el servidor
	// (US6/FR-027…FR-029).
	root.Handle(http.MethodGet, "/api/v1/portada/ministerios", h.GetPortadaMinisterios)
}

// AdminDeps reúne la configuración del grupo de panel de ministerios.
type AdminDeps struct {
	// Deps son las dependencias de la cadena del panel (middleware.AdminDeps);
	// el Recorder debe ser el service de F6 (T616) para registrar las denegaciones
	// del módulo.
	Deps middleware.AdminDeps
}

// RegisterAdmin publica el grupo /api/v1/admin/ministerios con la cadena
// aprobada de F2 —authn → guard de cambio de contraseña → authz(`ministerios`) →
// CSRF—, **sin helpers nuevos de middleware** (P6-6): el permiso ya existe en el
// catálogo y se verifica en el servidor, de forma independiente de los demás
// permisos de módulo (FR-011/SC-005). Las denegaciones las registra el Recorder
// de la cadena con la acción de la ruta (T616).
func RegisterAdmin(root httpserver.Registrar, h *Handler, deps AdminDeps) {
	if h == nil {
		return
	}
	group := root.Group("/api/v1/admin/ministerios", middleware.AdminChain(PermissionMinisterios, deps.Deps)...)

	// Gestión (FR-010/FR-014/FR-016).
	group.Handle(http.MethodGet, "", h.ListMinisterios)
	group.Handle(http.MethodPost, "", h.CreateMinisterio)
	// Imágenes (FR-020/R6-7).
	group.Handle(http.MethodPost, "/imagenes", h.UploadMinisterioImage)
	group.Handle(http.MethodGet, "/{id}", h.GetMinisterio)
	group.Handle(http.MethodPatch, "/{id}", h.UpdateMinisterio)
	group.Handle(http.MethodDelete, "/{id}", h.DeleteMinisterio)
}
