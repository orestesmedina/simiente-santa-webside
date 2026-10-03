package status

import (
	"context"
	"log/slog"
	"net/http"

	"simiente-santa/backend/internal/platform/httpserver"
)

// Service es la interfaz que este handler necesita del servicio. La define
// quien la consume (arq. R3) y la cumple service.go.
type Service interface {
	// Status devuelve el estado real del sistema o un apperr cuando la base de
	// datos no está conectada.
	Status(ctx context.Context) (SystemStatus, error)
}

// Handler expone GET /healthz. Es solo HTTP: delega en el service y responde
// con el sobre de éxito (el DTO directo) o deriva cualquier error a
// httpserver.WriteError, único punto de traducción (arq. R7).
type Handler struct {
	svc    Service
	logger *slog.Logger
}

// NewHandler construye el handler del dominio status.
func NewHandler(svc Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// GetSystemStatus responde 200 con el DTO SystemStatus o 503 con el sobre de
// error database_unavailable (FR-002, FR-004, D7). Fija Cache-Control: no-store
// en toda respuesta: el estado no debe ser cacheado por intermediarios.
// Ningún código de error se escribe a mano; todos salen por WriteError.
func (h *Handler) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Cache-Control", "no-store")

	st, err := h.svc.Status(ctx)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, st)
}
