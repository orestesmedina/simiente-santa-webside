package agenda

import (
	"net/http"
	"strings"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
)

// handler_public.go publica la operación pública del contrato:
// GET /api/v1/proximamente (T423). Es solo HTTP: valida `lang`, delega la
// lectura localizada en el service y responde 200 con `Cache-Control: no-store`
// (SC-003). Sin autenticación (FR-001) y sin middlewares de sesión (solo los
// globales de la aplicación).

// GetProximamente responde GET /api/v1/proximamente?lang=es|en (FR-001…FR-008):
// devuelve los próximos eventos y las próximas actividades PUBLICADOS y con
// fecha desde hoy en adelante, ya resueltos al idioma pedido (el service aplica
// el fallback en → es por campo y omite las claves de bloque vacías; SC-006/
// SC-012). El DTO público nunca contiene `publicationState` ni fechas pasadas
// (FR-013/SC-002).
// Códigos: 200 con ProximamentePublica; `lang` fuera de es|en → 400 invalid con
// details.lang.
func (h *Handler) GetProximamente(w http.ResponseWriter, r *http.Request) {
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

	proximamente, err := h.service.GetProximamente(ctx, lang)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	// Los cambios guardados se ven en la primera carga posterior (SC-003) y el
	// retiro automático depende de la fecha actual (R4-2): «lo que viene» viaja
	// siempre sin caché.
	w.Header().Set("Cache-Control", "no-store")
	httpserver.WriteJSON(w, http.StatusOK, proximamente)
}
