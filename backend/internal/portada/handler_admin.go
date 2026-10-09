package portada

import (
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
