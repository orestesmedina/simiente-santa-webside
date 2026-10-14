package ministerios

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/httpserver"
)

// handler_admin.go publica la gestión del módulo desde el panel bajo
// /api/v1/admin/ministerios (T619). Es solo HTTP: decodifica y valida los DTOs
// del contrato, delega en el service y responde el DTO del panel —que sí lleva
// los datos guardados del encargado, aunque su ficha no esté publicada (dato
// guardado ≠ dato publicado) y el aviso `whatsappUrlShared`—. Nada de SQL ni de
// credenciales en las respuestas. La cadena (authn → guard → authz(`ministerios`)
// → CSRF) la monta routes.go (T621).

// ListMinisterios responde GET /api/v1/admin/ministerios (FR-010): listado
// paginado completo —borradores incluidos— en el orden del módulo, con filtro
// opcional `publicationState` y el aviso de enlace compartido por fila.
// Códigos: 200 con MinisterioAdminList; parámetros inválidos → 400 con details.
func (h *Handler) ListMinisterios(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, ok := h.adminIdentity(ctx, w); !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}

	params, state, err := h.listParams(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	list, err := h.service.ListMinisterios(ctx, state, params)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, list)
}

// CreateMinisterio responde POST /api/v1/admin/ministerios (FR-010/FR-014):
// alta con validación completa en el servidor. `leaderVisible` nace en `false`
// si la clave falta («Oculto por defecto», P6-3).
// Códigos: 201 con MinisterioAdmin; DTO inválido → 400 con details por campo;
// duplicado exacto → 409.
func (h *Handler) CreateMinisterio(w http.ResponseWriter, r *http.Request) {
	h.createItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var in MinisterioInput
		if !h.decodeAndValidate(w, r, &in) {
			return nil, errAlreadyHandled
		}
		return h.service.CreateMinisterio(ctx, actorID, in)
	})
}

// GetMinisterio responde GET /api/v1/admin/ministerios/{id} (FR-010): ficha
// completa para el formulario de edición (ambos idiomas y la ficha del encargado
// aunque no esté publicada).
// Códigos: 200 con MinisterioAdmin; id no UUID → 400 con details.id;
// inexistente → 404.
func (h *Handler) GetMinisterio(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := ministerioIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	if _, ok := h.adminIdentity(ctx, w); !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}

	item, err := h.service.GetMinisterio(ctx, id)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, item)
}

// UpdateMinisterio responde PATCH /api/v1/admin/ministerios/{id} (FR-010): el
// handler transmite el DTO **parcial** tal cual y el service fusiona los campos
// presentes sobre la fila actual (A6-3/P6-20): los ausentes no cambian y un
// `null` explícito borra el opcional. `publicationState` cambia el estado por
// elemento (FR-012) y `leaderVisible` la publicación de la ficha del encargado
// (FR-025/FR-026), sin tocar los datos guardados.
// Códigos: 200 con MinisterioAdmin; PATCH vacío o DTO inválido → 400 con
// details; id no UUID → 400; inexistente → 404; duplicado exacto → 409.
func (h *Handler) UpdateMinisterio(w http.ResponseWriter, r *http.Request) {
	id, err := ministerioIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.updateItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var patch MinisterioPatch
		if !h.decodeAndValidate(w, r, &patch) {
			return nil, errAlreadyHandled
		}
		return h.service.UpdateMinisterio(ctx, actorID, id, patch)
	})
}

// DeleteMinisterio responde DELETE /api/v1/admin/ministerios/{id} (FR-010):
// borrado definitivo (la interfaz pide confirmación antes de llamar). El archivo
// de la imagen **no** se borra (A6-11): deja de servirse.
// Códigos: 204 sin cuerpo; id no UUID → 400; inexistente → 404.
func (h *Handler) DeleteMinisterio(w http.ResponseWriter, r *http.Request) {
	id, err := ministerioIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.deleteItem(w, r, func(ctx context.Context, actorID uuid.UUID) error {
		return h.service.DeleteMinisterio(ctx, actorID, id)
	})
}
