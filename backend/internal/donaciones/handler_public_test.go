package donaciones

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simiente-santa/backend/internal/platform/testutil"
)

// Tests del handler público de la sección GET /api/v1/ofrendas (T717) con
// httptest: el handler real sobre el service real, con el repositorio falso en
// memoria (sin PostgreSQL). Comprueban el sobre de éxito, la cabecera
// `Cache-Control: no-store`, la validación de `lang`, las claves de contenido
// omitidas cuando no hay nada publicado (lo que dispara la redirección del
// frontend) y —aserción sobre el JSON crudo— que el DTO público nunca expone
// `publicationState` ni un borrador.

// newTestHandler construye el handler sobre el service real y el almacén falso.
func newTestHandler(t *testing.T, repo Repository) *Handler {
	t.Helper()
	logger, _ := testutil.NewLogger()
	svc := NewService(ServiceDeps{Repository: repo, Logger: logger})
	return NewHandler(HandlerDeps{Service: svc, Logger: logger})
}

// doGet lanza un handler GET del dominio y devuelve la respuesta.
func doGet(t *testing.T, handler func(http.ResponseWriter, *http.Request), target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

// TestGetOfrendasHandlerOK fija el 200 con el sobre del contrato y la cabecera
// `Cache-Control: no-store` (SC-003/FR-013).
func TestGetOfrendasHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	method := publishedMethod(KindIban, "CR05015202001234567890")
	method.HolderEs = textOf("Ana Pérez")
	repo.seedMethods(method)
	repo.seedTexts(TextoDonacion{
		TextKey:          TextKeyIntro,
		BodyEs:           "Cómo donar",
		PublicationState: StatePublished,
	})
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetOfrendas, "/api/v1/ofrendas")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, se esperaba no-store", got)
	}

	var got OfrendasPublicas
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es OfrendasPublicas: %v (%s)", err, rec.Body)
	}
	if got.Lang != LangES {
		t.Errorf("lang = %q, se esperaba es (por defecto)", got.Lang)
	}
	if len(got.Items) != 1 {
		t.Fatalf("items = %+v", got.Items)
	}
	if item := got.Items[0]; item.Number != "CR05015202001234567890" || item.Holder != "Ana Pérez" || item.Currency != "CRC" {
		t.Errorf("item = %+v", item)
	}
	if got.Intro == nil || got.Intro.Text != "Cómo donar" {
		t.Errorf("intro = %+v", got.Intro)
	}
	if strings.Contains(rec.Body.String(), "publicationState") {
		t.Errorf("el DTO público expone el estado: %s", rec.Body)
	}
}

// TestGetOfrendasHandlerLang fija la resolución de `lang` en la capa HTTP: el
// fallback por campo lo aplica el servidor (FR-008) y el cliente no lo
// reimplementa.
func TestGetOfrendasHandlerLang(t *testing.T) {
	repo := newFakeRepository()
	method := publishedMethod(KindIban, "CR05015202001234567890")
	method.HolderEs = textOf("Ana Pérez")
	method.HolderEn = textOf("Ana Perez")
	repo.seedMethods(method)
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetOfrendas, "/api/v1/ofrendas?lang=EN")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"holder":"Ana Perez"`) {
		t.Errorf("cuerpo = %s, se esperaba la traducción", rec.Body)
	}
}

// TestGetOfrendasHandlerInvalidLang fija el 400 invalid con details.lang.
func TestGetOfrendasHandlerInvalidLang(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := doGet(t, h.GetOfrendas, "/api/v1/ofrendas?lang=fr")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid"`) || !strings.Contains(rec.Body.String(), `"lang"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestGetOfrendasHandlerOmitsKeysWhenEmpty fija SC-010/P7-13: sin nada publicado
// el cuerpo lleva solo `lang` (todas las claves de contenido omitidas): es lo que
// dispara la redirección de la ruta SPA en el frontend (P7-21) y la API sigue
// respondiendo 200.
func TestGetOfrendasHandlerOmitsKeysWhenEmpty(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := doGet(t, h.GetOfrendas, "/api/v1/ofrendas")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"lang":"es"}` {
		t.Fatalf("cuerpo = %s, se esperaba solo lang", body)
	}
}

// TestGetOfrendasHandlerNeverLeaksDrafts fija FR-012/SC-002 en la capa HTTP: con
// un método y un texto en borrador el JSON crudo no los contiene.
func TestGetOfrendasHandlerNeverLeaksDrafts(t *testing.T) {
	const (
		institution = "Banco Borrador"
		number      = "CR09999999999999999999"
		text        = "Texto en borrador"
	)
	repo := newFakeRepository()
	repo.seedMethods(MetodoDonacion{
		Kind:             KindIban,
		Number:           number,
		Institution:      institution,
		PublicationState: StateDraft,
	})
	repo.seedTexts(TextoDonacion{
		TextKey:          TextKeyIntro,
		BodyEs:           text,
		PublicationState: StateDraft,
	})
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetOfrendas, "/api/v1/ofrendas")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	raw := rec.Body.String()
	for _, secret := range []string{institution, number, text, "publicationState"} {
		if strings.Contains(raw, secret) {
			t.Errorf("el JSON crudo contiene contenido en borrador %q: %s", secret, raw)
		}
	}
}

// TestGetOfrendasHandlerServiceError fija el sobre genérico: un fallo interno no
// filtra la causa.
func TestGetOfrendasHandlerServiceError(t *testing.T) {
	repo := newFakeRepository()
	repo.readErr = errors.New("boom")
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetOfrendas, "/api/v1/ofrendas")
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

// TestGetOfrendasHandlerWithoutService fija el 500 de composición: un handler sin
// service no revienta la petición.
func TestGetOfrendasHandlerWithoutService(t *testing.T) {
	logger, _ := testutil.NewLogger()
	h := NewHandler(HandlerDeps{Logger: logger})

	rec := doGet(t, h.GetOfrendas, "/api/v1/ofrendas")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}
