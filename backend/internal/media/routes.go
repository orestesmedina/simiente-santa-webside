package media

import (
	"net/http"

	"simiente-santa/backend/internal/platform/httpserver"
)

// routes.go publica la superficie HTTP del dominio: la descarga pública de
// archivos, SIN middlewares de sesión (solo los globales de la aplicación): es
// de solo lectura y solo sirve contenido publicado y próximo (FR-013/R4-6).

// RegisterPublic publica `GET /api/v1/media/{fileName}`.
func RegisterPublic(root httpserver.Registrar, h *Handler) {
	if h == nil {
		return
	}
	root.Handle(http.MethodGet, "/api/v1/media/{fileName}", h.GetMedia)
}
