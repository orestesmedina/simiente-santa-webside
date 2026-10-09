package portada

import (
	"net/http"

	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/middleware"
)

// routes.go publica la superficie HTTP del dominio. Se mantiene separada de los
// handlers para que el reparto público/panel sea auditable de un vistazo
// (patrón de F2): la superficie pública queda en RegisterPublic y la de panel en
// RegisterAdmin, esta última tras la cadena aprobada con el permiso `portada`.

// PermissionPortada es el permiso de módulo del catálogo de F2 que activa F3
// (FR-012/P3-11): "Portada e información general". Es un `code` estable sembrado
// por la migración 000002; F3 no lo modifica.
const PermissionPortada = "portada"

// RegisterPublic publica la superficie pública de la portada, SIN middlewares
// de sesión (solo los globales de la aplicación): es de solo lectura y solo
// sirve contenido publicado (FR-001/FR-013).
func RegisterPublic(root httpserver.Registrar, h *Handler) {
	if h == nil {
		return
	}
	// Portada pública (FR-001…FR-009).
	root.Handle(http.MethodGet, "/api/v1/portada", h.GetPortada)
	// Descarga pública de imágenes, limitada a lo publicado (analyze C4/M6).
	root.Handle(http.MethodGet, "/api/v1/media/{fileName}", h.GetMedia)
}

// AdminDeps reúne la configuración del grupo de panel de la portada.
type AdminDeps struct {
	// Deps son las dependencias de la cadena del panel (middleware.AdminDeps);
	// el Recorder debe ser el service de F3 (T321) para registrar las
	// denegaciones del módulo.
	Deps middleware.AdminDeps
}

// RegisterAdmin publica el grupo /api/v1/admin/portada con la cadena aprobada
// (authn → guard de cambio de contraseña → authz(`portada`) → CSRF) y **activa
// el permiso `portada`** (FR-012/P3-11). Todas las operaciones exigen sesión y
// permiso; las denegaciones las registra el Recorder de la cadena.
func RegisterAdmin(root httpserver.Registrar, h *Handler, deps AdminDeps) {
	if h == nil {
		return
	}
	group := root.Group("/api/v1/admin/portada", middleware.AdminChain(PermissionPortada, deps.Deps)...)

	// Agregado del panel (T340, analyze C1).
	group.Handle(http.MethodGet, "", h.GetPortadaAdmin)

	// Singletons (T324).
	group.Handle(http.MethodPut, "/identidad", h.SaveIdentity)
	group.Handle(http.MethodPut, "/quienes-somos", h.SaveAbout)
	group.Handle(http.MethodPut, "/contacto", h.SaveContact)

	// Imágenes (T323).
	group.Handle(http.MethodPost, "/imagenes", h.UploadImage)

	// Horario, WhatsApp y redes (T325).
	group.Handle(http.MethodPost, "/horario", h.CreateSchedule)
	group.Handle(http.MethodPatch, "/horario/{id}", h.UpdateSchedule)
	group.Handle(http.MethodDelete, "/horario/{id}", h.DeleteSchedule)
	group.Handle(http.MethodPost, "/whatsapp", h.CreateWhatsapp)
	group.Handle(http.MethodPatch, "/whatsapp/{id}", h.UpdateWhatsapp)
	group.Handle(http.MethodDelete, "/whatsapp/{id}", h.DeleteWhatsapp)
	group.Handle(http.MethodPost, "/redes", h.CreateSocial)
	group.Handle(http.MethodPatch, "/redes/{id}", h.UpdateSocial)
	group.Handle(http.MethodDelete, "/redes/{id}", h.DeleteSocial)
}
