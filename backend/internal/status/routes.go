package status

import (
	"net/http"

	"simiente-santa/backend/internal/platform/httpserver"
)

// RegisterPublic publica las rutas PÚBLICAS del dominio status (sin
// autenticación). En F1 la única operación es el estado del sistema; queda
// separada de cualquier futura RegisterAdmin para que la superficie pública sea
// auditable de un vistazo.
func RegisterPublic(r httpserver.Registrar, h *Handler) {
	r.Handle(http.MethodGet, "/healthz", h.GetSystemStatus)
}
