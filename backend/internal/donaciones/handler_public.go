package donaciones

import (
	"net/http"
	"strings"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
)

// handler_public.go publica las operaciones públicas del contrato:
// GET /api/v1/ofrendas (T717) y GET /api/v1/portada/ofrendas (T718). Es solo
// HTTP: valida `lang`, delega la lectura localizada en el service y responde 200
// con `Cache-Control: no-store` (SC-003/FR-013). Sin autenticación (FR-002) y sin
// middlewares de sesión (solo los globales de la aplicación).

// GetOfrendas responde GET /api/v1/ofrendas?lang=es|en (FR-002…FR-006): devuelve
// el documento de la sección —texto introductorio y de uso si están publicados y
// los métodos publicados— en el orden de la sección y ya resuelto al idioma
// pedido (el service aplica el fallback en → es por campo y omite las claves sin
// dato). El DTO público nunca contiene `publicationState` ni un borrador.
//
// Sin nada publicado el cuerpo lleva solo `lang` (claves `items`/`intro`/`usage`
// omitidas, SC-010): la API responde 200 y no redirige —un 3xx/404 rompería el
// sobre y a sus consumidores—; la redirección de la ruta SPA `/ofrendas` es una
// decisión del frontend (P7-21).
//
// Códigos: 200 con OfrendasPublicas; `lang` fuera de es|en → 400 invalid con
// details.lang.
func (h *Handler) GetOfrendas(w http.ResponseWriter, r *http.Request) {
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

	ofrendas, err := h.service.GetOfrendas(ctx, lang)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	// Los cambios guardados se ven en la primera carga posterior (SC-003): la
	// sección siempre viaja sin caché.
	w.Header().Set("Cache-Control", "no-store")
	httpserver.WriteJSON(w, http.StatusOK, ofrendas)
}

// GetPortadaOfrendas responde GET /api/v1/portada/ofrendas?lang=es|en
// (US6/FR-022/SC-014): el extracto del bloque de la portada —hasta 5 métodos
// publicados, con el **tope firme del servidor** (no hay parámetro de límite ni
// paginación)— en el orden de la sección y ya resuelto al idioma pedido con
// fallback `en → es` por campo. El DTO solo lleva `lang` e `items` (mismo esquema
// de ítem que la sección, FR-022): nunca `publicationState` ni datos de borrador.
// Sin métodos publicados la respuesta viene **sin `items`** (clave omitida, nunca
// `null`): la portada no renderiza el bloque (sin mensaje, a diferencia del
// bloque de F4). Este bloque solo mira los métodos: los textos de la sección no
// forman parte de él.
//
// El recurso lo sirve el dominio `donaciones` aunque la ruta cuelgue del segmento
// `portada` (que es el agregado de F3): así el contrato no toca ninguna operación
// de F3 (Complexity 4 del plan).
//
// Códigos: 200 con PortadaOfrendas; `lang` fuera de es|en → 400 invalid con
// details.lang.
func (h *Handler) GetPortadaOfrendas(w http.ResponseWriter, r *http.Request) {
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
	// de límite (FR-022).
	portada, err := h.service.GetPortadaOfrendas(ctx, lang)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	// El bloque refleja los cambios guardados en la primera carga: siempre sin
	// caché (coherente con el resto del público).
	w.Header().Set("Cache-Control", "no-store")
	httpserver.WriteJSON(w, http.StatusOK, portada)
}
