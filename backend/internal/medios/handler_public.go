package medios

import (
	"net/http"
	"strconv"
	"strings"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
)

// handler_public.go publica las operaciones públicas del contrato:
// GET /api/v1/medios (T923) y GET /api/v1/portada/medios (T924). Es solo HTTP:
// valida `lang` y la paginación, delega la lectura localizada en el service y
// responde 200 con `Cache-Control: no-store` (SC-003/FR-015: los cambios
// guardados en el panel se ven en la primera carga posterior). Sin autenticación
// (FR-002) y sin middlewares de sesión (solo los globales de la aplicación): la
// superficie pública es de solo lectura y no hay rate-limit porque no hay
// superficie pública escribible.
//
// El DTO público (`MedioPublico`) lo compone el service en un único lugar: el
// handler no filtra, no reordena y no reimplementa el fallback de idioma (el
// cliente tampoco). El cuerpo nunca contiene `visibility`, marcas de ningún eje,
// `mergedIntoId` ni `sources`, y las claves opcionales viajan **omitidas**
// (nunca `null`).

// GetMedios responde GET /api/v1/medios?lang=es|en&limit=&offset= (FR-002…FR-006):
// la tanda de fichas **publicables** en el orden de la sección, con el sobre
// `{lang, items, total, limit, offset}` que sostiene «Cargar más».
//
// `lang` vacío es el idioma base (`es`) y cualquier otro valor es `400 invalid`
// con `details.lang`. La tanda la normaliza el service (default 12, tope 50): sin
// parámetro viaja `0` para que el defecto sea del dominio y **no** el del panel;
// un `limit`/`offset` no numérico es `400 invalid` con su campo en `details`, y
// uno negativo lo rechaza el service con la misma redacción (una sola regla para
// el rango).
//
// Sin nada publicable la respuesta sigue siendo `200` con `items` vacío y
// `total: 0` (la redirección de la ruta SPA `/medios` es del frontend: un
// `3xx`/`404` rompería el sobre).
func (h *Handler) GetMedios(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !h.serviceConfigured(ctx, w) {
		return
	}

	lang, err := normalizeLang(r.URL.Query().Get("lang"))
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	limit, offset, err := paginacionPublica(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	medios, err := h.service.GetMedios(ctx, lang, limit, offset)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	// La sección siempre viaja sin caché: el panel publica al guardar (FR-015).
	w.Header().Set("Cache-Control", "no-store")
	httpserver.WriteJSON(w, http.StatusOK, medios)
}

// GetPortadaMedios responde GET /api/v1/portada/medios?lang=es|en
// (US6/FR-027/SC-015): el extracto del bloque de la portada —hasta `maxPortada`
// (5) fichas publicables— en el **mismo orden** que la sección y con el **mismo
// esquema de ítem** («el bloque es un extracto de los mismos episodios, no un
// modelo nuevo»).
//
// El tope **5** lo aplica el service: la ruta **no** admite ningún parámetro de
// límite (el bloque no crece con el contenido). Sin ninguna ficha publicable la
// respuesta lleva solo `lang` (`items` omitida, nunca `null`) y la portada no
// renderiza el bloque, sin mensaje.
//
// El recurso lo sirve el dominio `medios` aunque la ruta cuelgue del segmento
// `portada` (que es el agregado de F3): así el contrato no toca ninguna operación
// de F3 (Complexity 5 del plan).
//
// Códigos: 200 con PortadaMedios; `lang` fuera de es|en → 400 invalid con
// details.lang.
func (h *Handler) GetPortadaMedios(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !h.serviceConfigured(ctx, w) {
		return
	}

	lang, err := normalizeLang(r.URL.Query().Get("lang"))
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	portada, err := h.service.GetPortadaMedios(ctx, lang)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	// El bloque refleja los cambios guardados en la primera carga: sin caché,
	// como el resto del público.
	w.Header().Set("Cache-Control", "no-store")
	httpserver.WriteJSON(w, http.StatusOK, portada)
}

// paginacionPublica lee `limit`/`offset` de la query de la sección sin aplicar
// el defecto del panel (`platform/paginate` usa 20): el ausente viaja como `0` y
// es el service quien aplica `pageDefault` (12) y el tope `maxPage` (50) —el tope
// nunca es un error—. Un valor no numérico es `400 invalid` con el campo en
// `details`, con la misma redacción que usa `platform/paginate`.
func paginacionPublica(r *http.Request) (int, int, error) {
	limit, err := enteroDeQuery(r, "limit")
	if err != nil {
		return 0, 0, err
	}
	offset, err := enteroDeQuery(r, "offset")
	if err != nil {
		return 0, 0, err
	}
	return limit, offset, nil
}

// enteroDeQuery lee un entero de la query: vacío es `0` (sin parámetro) y un
// valor no numérico es `400 invalid` con el campo en `details` (CWE-20).
func enteroDeQuery(r *http.Request, campo string) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(campo))
	if raw == "" {
		return 0, nil
	}
	valor, err := strconv.Atoi(raw)
	if err != nil {
		return 0, apperr.Invalid(
			"Parámetros de paginación inválidos",
			apperr.WithDetails(map[string]any{campo: "debe ser un número entero"}),
		)
	}
	return valor, nil
}
