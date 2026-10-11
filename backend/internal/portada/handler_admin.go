package portada

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
)

// handler_admin.go publica la gestión del módulo desde el panel bajo
// /api/v1/admin/portada (T324/T325/T340). Es solo HTTP: decodifica y valida los
// DTOs del contrato, delega en el service y responde el DTO del panel. Nada de
// SQL (R4) ni de credenciales en las respuestas (FR-026 de F2). La cadena
// (authn → guard → authz(`portada`) → CSRF) la monta routes.go (T326).

// --- Singletons: identidad, «quiénes somos» y contacto (T324) ---

// SaveIdentity responde PUT /api/v1/admin/portada/identidad (FR-002/FR-011):
// reemplazo completo de la identidad, con sus dos idiomas y su estado de
// publicación (FR-008/FR-013). La auditoría la registra el service en la misma
// transacción (FR-017).
// Códigos: 200 con IdentityAdmin; DTO inválido → 400 con details por campo.
func (h *Handler) SaveIdentity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}

	var in IdentityInput
	if !h.decodeAndValidate(w, r, &in) {
		return
	}

	saved, err := h.service.SaveIdentity(ctx, identity.UserID, in)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, saved)
}

// SaveAbout responde PUT /api/v1/admin/portada/quienes-somos (FR-003/FR-011):
// texto plano con límite de 1.000 caracteres por idioma.
// Códigos: 200 con AboutAdmin; DTO inválido → 400 con details por campo.
func (h *Handler) SaveAbout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}

	var in AboutInput
	if !h.decodeAndValidate(w, r, &in) {
		return
	}

	saved, err := h.service.SaveAbout(ctx, identity.UserID, in)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, saved)
}

// SaveContact responde PUT /api/v1/admin/portada/contacto (FR-007/FR-011):
// dirección, correo y teléfono obligatorios (FR-015).
// Códigos: 200 con ContactAdmin; DTO inválido → 400 con details por campo.
func (h *Handler) SaveContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}

	var in ContactInput
	if !h.decodeAndValidate(w, r, &in) {
		return
	}

	saved, err := h.service.SaveContact(ctx, identity.UserID, in)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, saved)
}

// GetPortadaAdmin responde GET /api/v1/admin/portada (FR-011, analyze C1): el
// estado completo del módulo para precargar el panel —identidad, «quiénes
// somos», contacto, horario, WhatsApp y redes— en ambos idiomas y con
// `publicationState` por elemento (borradores incluidos). Es la única vista que
// los expone; exige el permiso `portada`.
// Códigos: 200 con PortadaAdmin; sin sesión → 401; sin permiso → 403 (cadena).
func (h *Handler) GetPortadaAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, ok := h.adminIdentity(ctx, w); !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}

	out, err := h.service.GetPortadaAdmin(ctx)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, out)
}

// --- Colecciones: horario, WhatsApp y redes (T325) ---

// CreateSchedule responde POST /api/v1/admin/portada/horario (FR-004): alta de
// un servicio del horario (día 0–6, inicio "HH:MM" y fin opcional).
// Códigos: 201 con ScheduleItemAdmin; DTO inválido → 400 con details por campo.
func (h *Handler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	h.createItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var in ScheduleItemInput
		if !h.decodeAndValidate(w, r, &in) {
			return nil, errAlreadyHandled
		}
		return h.service.CreateService(ctx, actorID, in)
	})
}

// UpdateSchedule responde PATCH /api/v1/admin/portada/horario/{id} (FR-004):
// edición parcial (al menos un campo) y/o cambio de estado por elemento.
// Códigos: 200 con ScheduleItemAdmin; id no UUID o PATCH vacío → 400;
// inexistente → 404.
func (h *Handler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := portadaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.updateItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var patch ScheduleItemPatch
		if !h.decodeAndValidate(w, r, &patch) {
			return nil, errAlreadyHandled
		}
		return h.service.UpdateService(ctx, actorID, id, patch)
	})
}

// DeleteSchedule responde DELETE /api/v1/admin/portada/horario/{id} (FR-004):
// borrado físico del elemento (la auditoría conserva su etiqueta).
// Códigos: 204 sin cuerpo; id no UUID → 400; inexistente → 404.
func (h *Handler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := portadaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.deleteItem(w, r, func(ctx context.Context, actorID uuid.UUID) error {
		return h.service.DeleteService(ctx, actorID, id)
	})
}

// CreateWhatsapp responde POST /api/v1/admin/portada/whatsapp (FR-005).
// Códigos: 201 con WhatsappChannelAdmin; DTO inválido → 400; duplicado → 409.
func (h *Handler) CreateWhatsapp(w http.ResponseWriter, r *http.Request) {
	h.createItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var in WhatsappChannelInput
		if !h.decodeAndValidate(w, r, &in) {
			return nil, errAlreadyHandled
		}
		return h.service.CreateWhatsappChannel(ctx, actorID, in)
	})
}

// UpdateWhatsapp responde PATCH /api/v1/admin/portada/whatsapp/{id} (FR-005).
// Códigos: 200; id no UUID o PATCH vacío → 400; inexistente → 404; duplicado →
// 409.
func (h *Handler) UpdateWhatsapp(w http.ResponseWriter, r *http.Request) {
	id, err := portadaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.updateItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var patch WhatsappChannelPatch
		if !h.decodeAndValidate(w, r, &patch) {
			return nil, errAlreadyHandled
		}
		return h.service.UpdateWhatsappChannel(ctx, actorID, id, patch)
	})
}

// DeleteWhatsapp responde DELETE /api/v1/admin/portada/whatsapp/{id} (FR-005).
// Códigos: 204; id no UUID → 400; inexistente → 404.
func (h *Handler) DeleteWhatsapp(w http.ResponseWriter, r *http.Request) {
	id, err := portadaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.deleteItem(w, r, func(ctx context.Context, actorID uuid.UUID) error {
		return h.service.DeleteWhatsappChannel(ctx, actorID, id)
	})
}

// CreateSocial responde POST /api/v1/admin/portada/redes (FR-006).
// Códigos: 201 con SocialLinkAdmin; red/URL inválida → 400; duplicado → 409.
func (h *Handler) CreateSocial(w http.ResponseWriter, r *http.Request) {
	h.createItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var in SocialLinkInput
		if !h.decodeAndValidate(w, r, &in) {
			return nil, errAlreadyHandled
		}
		return h.service.CreateSocialLink(ctx, actorID, in)
	})
}

// UpdateSocial responde PATCH /api/v1/admin/portada/redes/{id} (FR-006).
// Códigos: 200; id no UUID o PATCH vacío → 400; inexistente → 404; duplicado →
// 409.
func (h *Handler) UpdateSocial(w http.ResponseWriter, r *http.Request) {
	id, err := portadaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.updateItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var patch SocialLinkPatch
		if !h.decodeAndValidate(w, r, &patch) {
			return nil, errAlreadyHandled
		}
		return h.service.UpdateSocialLink(ctx, actorID, id, patch)
	})
}

// DeleteSocial responde DELETE /api/v1/admin/portada/redes/{id} (FR-006).
// Códigos: 204; id no UUID → 400; inexistente → 404.
func (h *Handler) DeleteSocial(w http.ResponseWriter, r *http.Request) {
	id, err := portadaItemIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.deleteItem(w, r, func(ctx context.Context, actorID uuid.UUID) error {
		return h.service.DeleteSocialLink(ctx, actorID, id)
	})
}

// errAlreadyHandled señala que el helper ya escribió la respuesta (p. ej. el
// 400 de un DTO inválido): el envoltorio de las colecciones no debe volver a
// responder.
var errAlreadyHandled = errors.New("respuesta ya escrita")

// createItem unifica el preludio de los POST de colección: exige sesión y
// servicio, delega en fn y responde 201 con el DTO del panel.
func (h *Handler) createItem(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, actorID uuid.UUID) (any, error)) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}
	item, err := fn(ctx, identity.UserID)
	if err != nil {
		if errors.Is(err, errAlreadyHandled) {
			return
		}
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, item)
}

// updateItem unifica el preludio de los PATCH de colección: exige sesión y
// servicio, delega en fn y responde 200 con el DTO del panel.
func (h *Handler) updateItem(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, actorID uuid.UUID) (any, error)) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}
	item, err := fn(ctx, identity.UserID)
	if err != nil {
		if errors.Is(err, errAlreadyHandled) {
			return
		}
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, item)
}

// deleteItem unifica el preludio de los DELETE de colección: exige sesión y
// servicio, delega en fn y responde 204 sin cuerpo.
func (h *Handler) deleteItem(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, actorID uuid.UUID) error) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}
	if err := fn(ctx, identity.UserID); err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// portadaItemIDFromRequest lee el id de la ruta ({id}) y lo convierte a UUID.
// Un valor que no es UUID → 400 invalid con details.id.
func portadaItemIDFromRequest(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return uuid.Nil, apperr.Invalid(
			"El identificador del elemento no es válido",
			apperr.WithDetails(map[string]any{"id": "Debe ser un identificador UUID válido"}),
		)
	}
	return id, nil
}
