package donaciones

import (
	"net/http"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/middleware"
)

// routes.go publica la superficie HTTP del dominio. Se mantiene separada de los
// handlers para que el reparto público/panel sea auditable de un vistazo (patrón
// de F2): la superficie pública queda en RegisterPublic y la de panel en
// RegisterAdmin, esta última tras la cadena aprobada con el permiso `donaciones`.

// PermissionDonaciones es el permiso de módulo del catálogo de F2 que activa F7
// (FR-011/P7-6): «Donaciones». Es un `code` estable sembrado por la migración
// 000002; **F7 no añade ningún permiso** ni toca el catálogo de F2 (sigue en 10).
const PermissionDonaciones = "donaciones"

// RegisterPublic publica la superficie pública de ofrendas, SIN middlewares de
// sesión (solo los globales de la aplicación): es de solo lectura y solo sirve
// contenido publicado (FR-002/FR-012). No hay rate-limit: no hay superficie
// pública escribible.
//
// La segunda ruta cuelga del segmento `portada` (el agregado de F3) pero la sirve
// este dominio (Complexity 4 del plan): así el contrato no toca ninguna operación
// de F3 y la portada no se rediseña.
func RegisterPublic(root httpserver.Registrar, h *Handler) {
	if h == nil {
		return
	}
	// Sección pública completa (FR-002…FR-006). La redirección de la ruta SPA
	// `/ofrendas` es del frontend: la API responde 200 (P7-21).
	root.Handle(http.MethodGet, "/api/v1/ofrendas", h.GetOfrendas)
	// Bloque de la portada: hasta 5 métodos publicados, con el tope en el servidor
	// (US6/FR-022).
	root.Handle(http.MethodGet, "/api/v1/portada/ofrendas", h.GetPortadaOfrendas)
}

// AdminDeps reúne la configuración del grupo de panel de donaciones.
type AdminDeps struct {
	// Deps son las dependencias de la cadena del panel (middleware.AdminDeps); el
	// Recorder debe ser el service de F7 (T716) para registrar las denegaciones del
	// módulo.
	Deps middleware.AdminDeps
}

// RegisterAdmin publica el grupo /api/v1/admin/donaciones con la cadena aprobada
// de F2 —authn → guard de cambio de contraseña → authz(`donaciones`) → CSRF—,
// **sin helpers nuevos de middleware** (P7-6): el permiso ya existe en el catálogo
// y se verifica en el servidor, de forma independiente de los demás permisos de
// módulo (FR-011/SC-005). Las denegaciones las registra el Recorder de la cadena
// con la acción de la ruta (T716).
func RegisterAdmin(root httpserver.Registrar, h *Handler, deps AdminDeps) {
	if h == nil {
		return
	}
	group := root.Group("/api/v1/admin/donaciones", middleware.AdminChain(PermissionDonaciones, deps.Deps)...)

	// Gestión (FR-010/FR-014/FR-019).
	group.Handle(http.MethodGet, "", h.ListDonaciones)
	group.Handle(http.MethodPost, "", h.CreateDonacion)
	// Textos de la sección (FR-021/Q5/Q11; sin borrado — P7-17).
	group.Handle(http.MethodGet, "/textos", h.GetTextos)
	group.Handle(http.MethodPut, "/textos/{textKey}", h.PutTexto)
	// Detalle, edición y borrado del método.
	group.Handle(http.MethodGet, "/{id}", h.GetDonacion)
	group.Handle(http.MethodPatch, "/{id}", h.UpdateDonacion)
	group.Handle(http.MethodDelete, "/{id}", h.DeleteDonacion)
}
