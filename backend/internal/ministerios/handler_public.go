package ministerios

import (
	"net/http"
	"strings"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
)

// handler_public.go publica las operaciones públicas del contrato:
// GET /api/v1/ministerios (T617) y GET /api/v1/portada/ministerios (T618). Es
// solo HTTP: valida `lang`, delega la lectura localizada en el service y responde
// 200 con `Cache-Control: no-store` (SC-003/FR-013). Sin autenticación (FR-002) y
// sin middlewares de sesión (solo los globales de la aplicación).

// GetMinisterios responde GET /api/v1/ministerios?lang=es|en (FR-002…FR-008):
// devuelve los ministerios PUBLICADOS, en el orden manual del módulo y ya
// resueltos al idioma pedido (el service aplica el fallback en → es por campo y
// omite las claves sin dato). El DTO público nunca contiene `publicationState`
// ni —con la ficha no publicada— ningún dato del encargado (FR-026/SC-014). Sin
// publicados el cuerpo lleva solo `lang` (la clave `items` se omite, SC-010): la
// API responde 200 y no redirige (la redirección de la ruta SPA es del
// frontend, P6-21).
//
// Códigos: 200 con MinisteriosPublicos; `lang` fuera de es|en → 400 invalid con
// details.lang.
func (h *Handler) GetMinisterios(w http.ResponseWriter, r *http.Request) {
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

	ministerios, err := h.service.GetMinisterios(ctx, lang)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	// Los cambios guardados se ven en la primera carga posterior (SC-003): la
	// sección siempre viaja sin caché.
	w.Header().Set("Cache-Control", "no-store")
	httpserver.WriteJSON(w, http.StatusOK, ministerios)
}

// GetPortadaMinisterios responde GET /api/v1/portada/ministerios?lang=es|en
// (US6/FR-027…FR-029): el extracto del bloque de la portada —hasta 5
// ministerios publicados, con el **tope firme del servidor** (no hay parámetro
// de límite ni paginación)— en el orden manual del módulo y ya resuelto al
// idioma pedido con fallback `en → es` por campo. El DTO solo lleva `id`,
// `name`, `schedule`/`place` e `image`+`alt`: nunca `publicationState` ni datos
// del encargado (el extracto no publica encargado ni botón de unión, D-12).
// Sin publicados la respuesta viene **sin `items`** (clave omitida, nunca
// `null`): la portada no renderiza el bloque (FR-028).
//
// El recurso lo sirve el dominio `ministerios` aunque la ruta cuelgue del
// segmento `portada` (que es el agregado de F3): así el contrato no toca ninguna
// operación de F3 (Complexity 4 del plan).
//
// Códigos: 200 con PortadaMinisterios; `lang` fuera de es|en → 400 invalid con
// details.lang.
func (h *Handler) GetPortadaMinisterios(w http.ResponseWriter, r *http.Request) {
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

	// El tope 5 se aplica en el service: la petición no admite ningún parámetro
	// de límite (FR-027).
	ministerios, err := h.service.GetPortadaMinisterios(ctx, lang)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	// El bloque refleja los cambios guardados en la primera carga: siempre sin
	// caché (coherente con el resto del público).
	w.Header().Set("Cache-Control", "no-store")
	httpserver.WriteJSON(w, http.StatusOK, ministerios)
}
