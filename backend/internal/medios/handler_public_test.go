package medios

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/testutil"
)

// uuidDePrueba devuelve un identificador aleatorio para los casos que solo
// necesitan un UUID válido.
func uuidDePrueba() uuid.UUID { return uuid.New() }

// Tests de los handlers públicos del dominio (T923/T924) con httptest: el handler
// real sobre el service real, con el repositorio falso en memoria (sin
// PostgreSQL). Comprueban el sobre de éxito, la cabecera `Cache-Control:
// no-store`, la validación de `lang` y de la paginación, las claves opcionales
// omitidas, el tope 5 del bloque de la portada y —aserción sobre el JSON crudo—
// que el DTO público nunca expone `visibility`, marcas, `mergedIntoId` ni
// `sources`, ni deja salir una ficha no publicable.

// newTestHandler construye el handler sobre el service real y el repositorio
// falso en memoria.
func newTestHandler(t *testing.T, repo Repository) *Handler {
	t.Helper()
	return newTestHandlerCon(t, repo, ServiceDeps{})
}

// doGet lanza un handler GET del dominio y devuelve la respuesta.
func doGet(t *testing.T, handler func(http.ResponseWriter, *http.Request), target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

// episodioPublicable arma una ficha publicable con descripción y video.
func episodioPublicable(titulo string) Episodio {
	descripcion := "Descripción de " + titulo
	return Episodio{
		Titulo:        titulo,
		PublicadoEn:   time.Date(2026, time.July, 3, 12, 0, 0, 0, time.UTC),
		DescripcionEs: &descripcion,
		Origen:        OrigenVideo,
		Visibilidad:   VisibilidadVisible,
	}
}

// --- Sección pública (T923, FR-002…FR-006) ---

// TestGetMediosHandlerOK fija el 200 con el sobre del contrato, la cabecera
// `no-store` y la tanda por defecto (12) resuelta por el service.
func TestGetMediosHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(episodioPublicable("Potencial sin Resultados"))
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetMedios, "/api/v1/medios")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, se esperaba no-store", got)
	}

	var got MediosPublicos
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es MediosPublicos: %v (%s)", err, rec.Body)
	}
	if got.Lang != LangES || got.Total != 1 || len(got.Items) != 1 {
		t.Fatalf("sobre = %+v", got)
	}
	if got.Limit != pageDefault || got.Offset != 0 {
		t.Errorf("tanda = {%d, %d}, se esperaba {%d, 0}", got.Limit, got.Offset, pageDefault)
	}
	if item := got.Items[0]; item.Titulo != "Potencial sin Resultados" || item.Descripcion != "Descripción de Potencial sin Resultados" {
		t.Errorf("item = %+v", item)
	}
}

// TestGetMediosHandlerLang fija la resolución de `lang` en la capa HTTP: el
// fallback por campo lo aplica el servidor (FR-025) y el cliente no lo
// reimplementa. `title` y `publishedAt` son idénticos en los dos idiomas
// (FR-026).
func TestGetMediosHandlerLang(t *testing.T) {
	en := "The Lesson of the Fig Tree"
	ep := episodioPublicable("Potencial sin Resultados")
	ep.DescripcionEn = &en

	repo := newFakeRepository()
	repo.seed(ep)
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetMedios, "/api/v1/medios?lang=EN")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"lang":"en"`) {
		t.Errorf("cuerpo = %s, se esperaba lang=en", rec.Body)
	}
	if !strings.Contains(rec.Body.String(), en) {
		t.Errorf("cuerpo = %s, se esperaba la traducción", rec.Body)
	}
	if !strings.Contains(rec.Body.String(), "Potencial sin Resultados") {
		t.Errorf("el nombre no cambia con el idioma (FR-026): %s", rec.Body)
	}
}

// TestGetMediosHandlerInvalidLang fija el 400 invalid con details.lang.
func TestGetMediosHandlerInvalidLang(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := doGet(t, h.GetMedios, "/api/v1/medios?lang=fr")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid"`) || !strings.Contains(rec.Body.String(), `"lang"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestGetMediosHandlerPaginacion fija la normalización de la tanda (R9-10): el
// defecto (12) y el tope (50) son del service —el tope **nunca** es un error— y
// un parámetro imposible es `400 invalid` con su campo en `details`.
func TestGetMediosHandlerPaginacion(t *testing.T) {
	repo := newFakeRepository()
	for i := 0; i < 13; i++ {
		ep := episodioPublicable("Episodio")
		ep.PublicadoEn = ep.PublicadoEn.Add(time.Duration(i) * time.Hour)
		repo.seed(ep)
	}
	h := newTestHandler(t, repo)

	t.Run("el tope de la sección se aplica sin error", func(t *testing.T) {
		rec := doGet(t, h.GetMedios, "/api/v1/medios?limit=51")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
		}
		var got MediosPublicos
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo no es MediosPublicos: %v", err)
		}
		if got.Limit != maxPage || len(got.Items) != 13 || got.Total != 13 {
			t.Errorf("sobre = {%d items, limit %d, total %d}, se esperaba {13, %d, 13}",
				len(got.Items), got.Limit, got.Total, maxPage)
		}
	})

	t.Run("la segunda tanda pide offset", func(t *testing.T) {
		rec := doGet(t, h.GetMedios, "/api/v1/medios?offset=12")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
		}
		var got MediosPublicos
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo no es MediosPublicos: %v", err)
		}
		if len(got.Items) != 1 || got.Offset != pageDefault {
			t.Errorf("sobre = {%d items, offset %d}, se esperaba {1, %d}", len(got.Items), got.Offset, pageDefault)
		}
	})

	t.Run("parámetros imposibles", func(t *testing.T) {
		for _, tt := range []struct {
			target string
			detail string
		}{
			{target: "/api/v1/medios?limit=-1", detail: "limit"},
			{target: "/api/v1/medios?offset=-1", detail: "offset"},
			{target: "/api/v1/medios?limit=x", detail: "limit"},
			{target: "/api/v1/medios?offset=x", detail: "offset"},
		} {
			rec := doGet(t, h.GetMedios, tt.target)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("GET %s: status = %d, se esperaba 400 (%s)", tt.target, rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("GET %s: sin details.%s: %s", tt.target, tt.detail, rec.Body)
			}
		}
	})
}

// TestGetMediosHandlerOmitsKeys fija SC-010: una ficha sin descripción, sin
// carátula y sin audio viaja con las claves **omitidas** (nunca `null`).
func TestGetMediosHandlerOmitsKeys(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(Episodio{
		Titulo:      "Solo video",
		PublicadoEn: time.Date(2026, time.July, 3, 12, 0, 0, 0, time.UTC),
		Origen:      OrigenVideo,
		Visibilidad: VisibilidadVisible,
	})
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetMedios, "/api/v1/medios")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	raw := rec.Body.String()
	for _, clave := range []string{"description", "coverUrl", "audioUrl", "videoUrl", "null"} {
		if strings.Contains(raw, clave) {
			t.Errorf("el JSON contiene %q: %s", clave, raw)
		}
	}
}

// TestGetMediosHandlerNeverLeaksInternal fija SC-002/SC-011 en la capa HTTP: las
// fichas no publicables (oculta, absorbida y par abierto) no salen, y el JSON
// crudo no lleva ningún campo interno del modelo.
func TestGetMediosHandlerNeverLeaksInternal(t *testing.T) {
	pareja := episodioPublicable("Par dudoso")
	parID := uuidDePrueba()
	pareja.DuplicadoDeID = &parID

	absorbida := episodioPublicable("Absorbida")
	resultante := uuidDePrueba()
	absorbida.UnidoAID = &resultante

	oculta := episodioPublicable("Oculta por el equipo")
	oculta.Visibilidad = VisibilidadOculta

	repo := newFakeRepository()
	repo.seed(episodioPublicable("Publicable"), pareja, absorbida, oculta)
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetMedios, "/api/v1/medios")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}

	var got MediosPublicos
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es MediosPublicos: %v", err)
	}
	if got.Total != 1 || len(got.Items) != 1 || got.Items[0].Titulo != "Publicable" {
		t.Fatalf("sobre = %+v, se esperaba solo la publicable", got)
	}

	raw := rec.Body.String()
	for _, secreto := range []string{
		"Oculta por el equipo", "Absorbida", "Par dudoso",
		"visibility", "reviewMark", "duplicateOfId", "duplicateMarkResolvedAt",
		"availabilityMark", "availabilityResult", "mergedIntoId", "sources",
	} {
		if strings.Contains(raw, secreto) {
			t.Errorf("el JSON crudo contiene %q: %s", secreto, raw)
		}
	}
}

// TestGetMediosHandlerServiceError fija el sobre genérico: un fallo interno no
// filtra la causa (FR-018).
func TestGetMediosHandlerServiceError(t *testing.T) {
	repo := newFakeRepository()
	repo.listErr = errors.New("boom")
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetMedios, "/api/v1/medios")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("la respuesta filtra la causa interna: %s", rec.Body)
	}
}

// TestGetMediosHandlerWithoutService fija el 500 de composición: un handler sin
// service no revienta la petición.
func TestGetMediosHandlerWithoutService(t *testing.T) {
	logger, _ := testutil.NewLogger()
	h := NewHandler(HandlerDeps{Logger: logger})

	rec := doGet(t, h.GetMedios, "/api/v1/medios")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// --- Bloque de la portada (T924, US6/FR-027/SC-015) ---

// TestGetPortadaMediosHandlerOK fija el 200 con el sobre del contrato, la
// cabecera `no-store` y el **tope 5 del servidor**: la ruta no admite ningún
// parámetro de límite, así que los 8 publicables salen acotados a 5.
func TestGetPortadaMediosHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	for i := 0; i < 8; i++ {
		ep := episodioPublicable(fmt.Sprintf("Episodio %d", i))
		ep.PublicadoEn = ep.PublicadoEn.Add(time.Duration(i) * time.Hour)
		repo.seed(ep)
	}
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetPortadaMedios, "/api/v1/portada/medios?limit=100")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, se esperaba no-store", got)
	}

	var got PortadaMedios
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es PortadaMedios: %v (%s)", err, rec.Body)
	}
	if got.Lang != LangES {
		t.Errorf("lang = %q, se esperaba es", got.Lang)
	}
	if len(got.Items) != maxPortada {
		t.Fatalf("items = %d, se esperaban %d (tope del servidor)", len(got.Items), maxPortada)
	}
	// El mismo orden de la sección: los 5 primeros (los más recientes), no otros.
	if got.Items[0].Titulo != "Episodio 7" {
		t.Errorf("el bloque no empieza por el más reciente: %+v", got.Items[0])
	}
	if raw := rec.Body.String(); strings.Contains(raw, "Episodio 2") {
		t.Errorf("el bloque publica más allá del tope: %s", raw)
	}
}

// TestGetPortadaMediosHandlerOmitsItems fija FR-027/SC-015: sin ninguna ficha
// publicable la respuesta lleva solo `lang` (nunca `items: null`) y la portada no
// dibuja el bloque, sin mensaje.
func TestGetPortadaMediosHandlerOmitsItems(t *testing.T) {
	repo := newFakeRepository()
	oculta := episodioPublicable("Oculta")
	oculta.Visibilidad = VisibilidadOculta
	repo.seed(oculta)
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetPortadaMedios, "/api/v1/portada/medios")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"lang":"es"}` {
		t.Fatalf("cuerpo = %s, se esperaba solo lang", body)
	}
}

// TestGetPortadaMediosHandlerInvalidLang fija el 400 invalid con details.lang.
func TestGetPortadaMediosHandlerInvalidLang(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := doGet(t, h.GetPortadaMedios, "/api/v1/portada/medios?lang=fr")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid"`) || !strings.Contains(rec.Body.String(), `"lang"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestGetPortadaMediosHandlerServiceError fija el sobre genérico: un fallo
// interno no filtra la causa.
func TestGetPortadaMediosHandlerServiceError(t *testing.T) {
	repo := newFakeRepository()
	repo.listErr = errors.New("boom")
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetPortadaMedios, "/api/v1/portada/medios")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("la respuesta filtra la causa interna: %s", rec.Body)
	}
}

// TestGetPortadaMediosHandlerWithoutService fija el 500 de composición.
func TestGetPortadaMediosHandlerWithoutService(t *testing.T) {
	logger, _ := testutil.NewLogger()
	h := NewHandler(HandlerDeps{Logger: logger})

	rec := doGet(t, h.GetPortadaMedios, "/api/v1/portada/medios")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
}
