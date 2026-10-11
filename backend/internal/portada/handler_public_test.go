package portada

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simiente-santa/backend/internal/platform/testutil"
)

// strPtr devuelve un puntero a la cadena indicada (campos *En opcionales).
func strPtr(s string) *string { return &s }

// Tests del handler público GET /api/v1/portada (T322) con httptest: el
// handler real sobre el service real, con el repositorio falso en memoria (sin
// PostgreSQL). Comprueban el sobre de éxito, la cabecera no-store, la validación
// de `lang` y el fallo del service.

func newPublicHandler(t *testing.T, repo Repository) *Handler {
	t.Helper()
	logger, _ := testutil.NewLogger()
	return NewHandler(HandlerDeps{Service: NewService(ServiceDeps{Repository: repo, Logger: logger}), Logger: logger})
}

func doGetPortada(t *testing.T, h *Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	h.GetPortada(rec, req)
	return rec
}

func TestGetPortadaHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	repo.identity = &Identity{
		NameEs:           "Iglesia Simiente Santa",
		NameEn:           strPtr("Simiente Santa Church"),
		PublicationState: StatePublished,
	}
	repo.about = &About{TextEs: "Somos una iglesia.", PublicationState: StatePublished}
	repo.seedService(Service{DayOfWeek: 0, StartTime: "10:00", NameEs: "Culto", PlaceEs: "Sede", PublicationState: StatePublished})

	h := newPublicHandler(t, repo)

	rec := doGetPortada(t, h, "/api/v1/portada")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, se esperaba no-store", got)
	}

	var got PublicPortada
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es PortadaPublica: %v (%s)", err, rec.Body)
	}
	if got.Lang != LangES {
		t.Errorf("lang = %q, se esperaba es (por defecto)", got.Lang)
	}
	if got.Identity == nil || got.Identity.Name != "Iglesia Simiente Santa" {
		t.Fatalf("identity inesperada: %+v", got.Identity)
	}
	if got.About == nil || got.About.Text != "Somos una iglesia." {
		t.Fatalf("about inesperado: %+v", got.About)
	}
	if len(got.Schedule) != 1 {
		t.Fatalf("schedule = %+v", got.Schedule)
	}
}

func TestGetPortadaHandlerLangEN(t *testing.T) {
	repo := newFakeRepository()
	repo.identity = &Identity{
		NameEs:           "Iglesia Simiente Santa",
		NameEn:           strPtr("Simiente Santa Church"),
		PublicationState: StatePublished,
	}
	h := newPublicHandler(t, repo)

	rec := doGetPortada(t, h, "/api/v1/portada?lang=EN")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got PublicPortada
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es PortadaPublica: %v", err)
	}
	if got.Lang != LangEN {
		t.Errorf("lang = %q, se esperaba en", got.Lang)
	}
	if got.Identity == nil || got.Identity.Name != "Simiente Santa Church" {
		t.Fatalf("identity inesperada: %+v", got.Identity)
	}
}

func TestGetPortadaHandlerInvalidLang(t *testing.T) {
	h := newPublicHandler(t, newFakeRepository())

	rec := doGetPortada(t, h, "/api/v1/portada?lang=fr")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid"`) ||
		!strings.Contains(rec.Body.String(), `"lang"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

func TestGetPortadaHandlerOmitsEmptySections(t *testing.T) {
	h := newPublicHandler(t, newFakeRepository())

	rec := doGetPortada(t, h, "/api/v1/portada")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	// Sin nada publicado, el documento solo lleva `lang` (SC-012): ninguna
	// sección viaja como `null` ni vacía.
	if body := strings.TrimSpace(rec.Body.String()); body != `{"lang":"es"}` {
		t.Fatalf("cuerpo = %s, se esperaba solo lang", body)
	}
}

func TestGetPortadaHandlerServiceError(t *testing.T) {
	repo := newFakeRepository()
	repo.getErr = errors.New("boom")
	h := newPublicHandler(t, repo)

	rec := doGetPortada(t, h, "/api/v1/portada")
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
