package portada

import (
	"net/http"
	"strings"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
)

// handler_public.go publica la operación pública del contrato:
// GET /api/v1/portada (T322). Es solo HTTP: valida `lang`, delega la lectura
// localizada en el service y responde 200 con `Cache-Control: no-store`
// (SC-003). Sin autenticación (FR-001).

// GetPortada responde GET /api/v1/portada?lang=es|en (FR-001/FR-008/FR-009):
// devuelve la información general PUBLICADA, ya resuelta al idioma pedido (el
// service aplica el fallback en → es y omite las secciones vacías; SC-006/
// SC-012). El DTO público nunca contiene `publicationState` ni campos sin
// resolver (FR-013).
// Códigos: 200; `lang` fuera de es|en → 400 invalid con details.lang.
func (h *Handler) GetPortada(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !h.serviceConfigured(ctx, w) {
		return
	}

	lang := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("lang")))
	if lang == "" {
		lang = LangES
	}
	if !validLang(lang) {
		httpserver.WriteError(ctx, w, h.logger, apperr.Invalid(
			"El idioma solicitado no es válido",
			apperr.WithDetails(map[string]any{"lang": "solo se admite es o en"}),
		))
		return
	}

	portada, err := h.service.GetPortada(ctx, lang)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	// Los cambios guardados se ven en la primera carga posterior (SC-003): la
	// portada siempre viaja sin caché.
	w.Header().Set("Cache-Control", "no-store")
	httpserver.WriteJSON(w, http.StatusOK, portada)
}
