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

// GetPortadaProximamente responde GET /api/v1/portada/proximamente?lang=es|en
// (US7/FR-025…FR-027, enmienda de alcance del 2026-10-12/Q8): el extracto del
// bloque «Lo que viene» de la portada —hasta 5 ítems próximos, eventos y
// actividades mezclados por cercanía, preferentemente de los próximos 2 meses y
// con respaldo a los siguientes próximos—, ya resuelto al idioma pedido con
// fallback `en → es` por campo (FR-008). El DTO solo lleva `kind`, `title`,
// `date`/horario, `place` e `image`+`alt`: nunca `publicationState`, detalle ni
// lista de fechas. Sin ítems la respuesta viene **sin `items`** (clave omitida,
// nunca `null`): la portada muestra entonces el bloque con su mensaje (FR-026).
//
// El recurso lo sirve el dominio `agenda` aunque la ruta cuelgue del segmento
// `portada` (que es el agregado de F3): es un extracto de la agenda y así el
// contrato no toca ninguna operación de F3 (FR-027: la portada no se rediseña).
//
// Códigos: 200 con PortadaProximamente; `lang` fuera de es|en → 400 invalid con
// details.lang.
func (h *Handler) GetPortadaProximamente(w http.ResponseWriter, r *http.Request) {
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

	proximamente, err := h.service.GetPortadaProximamente(ctx, lang)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	// El bloque refleja los cambios guardados y el retiro automático en la
	// primera carga: siempre sin caché (coherente con el resto del público).
	w.Header().Set("Cache-Control", "no-store")
	httpserver.WriteJSON(w, http.StatusOK, proximamente)
}
