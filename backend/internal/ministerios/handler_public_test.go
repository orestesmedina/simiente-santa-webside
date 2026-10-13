package ministerios

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/testutil"
)

// Tests del handler público de la sección GET /api/v1/ministerios (T617) con
// httptest: el handler real sobre el service real, con el repositorio falso en
// memoria (sin PostgreSQL). Comprueban el sobre de éxito, la cabecera no-store,
// la validación de `lang`, la clave `items` omitida y —aserción sobre el JSON
// crudo— que el encargado no publicado no aparece por ninguna vía (FR-026).

// newTestHandler construye el handler sobre el service real y el almacén falso.
func newTestHandler(t *testing.T, repo Repository) *Handler {
	t.Helper()
	logger, _ := testutil.NewLogger()
	svc := NewService(ServiceDeps{Repository: repo, Store: newFakeStore(), Logger: logger})
	return NewHandler(HandlerDeps{Service: svc, MaxUploadBytes: 1 << 20, Logger: logger})
}

// doGet lanza un handler GET del dominio y devuelve la respuesta.
func doGet(t *testing.T, handler func(http.ResponseWriter, *http.Request), target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

// TestGetMinisteriosHandlerOK fija el 200 con el sobre del contrato y la
// cabecera `Cache-Control: no-store` (SC-003/FR-013).
func TestGetMinisteriosHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	ministry := publishedMinistry("Alabanza")
	ministry.ScheduleEs = textOf("Todos los viernes")
	ministry.LeaderNameEs = textOf("Ana Pérez")
	ministry.LeaderPhone = textOf("+58 412 000 0000")
	ministry.LeaderVisible = true
	ministry.WhatsappURL = textOf("https://chat.whatsapp.com/abc")
	repo.seed(ministry)
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetMinisterios, "/api/v1/ministerios")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, se esperaba no-store", got)
	}

	var got MinisteriosPublicos
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es MinisteriosPublicos: %v (%s)", err, rec.Body)
	}
	if got.Lang != LangES {
		t.Errorf("lang = %q, se esperaba es (por defecto)", got.Lang)
	}
	if len(got.Items) != 1 {
		t.Fatalf("items = %+v", got.Items)
	}
	item := got.Items[0]
	if item.Name != "Alabanza" || item.Leader == nil || item.Leader.Name != "Ana Pérez" {
		t.Errorf("item = %+v", item)
	}
	if strings.Contains(rec.Body.String(), "publicationState") {
		t.Errorf("el DTO público expone el estado: %s", rec.Body)
	}
}

// TestGetMinisteriosHandlerLang fija la resolución de `lang` en la capa HTTP.
func TestGetMinisteriosHandlerLang(t *testing.T) {
	repo := newFakeRepository()
	ministry := publishedMinistry("Alabanza")
	ministry.NameEn = textOf("Worship")
	repo.seed(ministry)
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetMinisterios, "/api/v1/ministerios?lang=EN")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"name":"Worship"`) {
		t.Errorf("cuerpo = %s, se esperaba la traducción", rec.Body)
	}
}

// TestGetMinisteriosHandlerInvalidLang fija el 400 invalid con details.lang.
func TestGetMinisteriosHandlerInvalidLang(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := doGet(t, h.GetMinisterios, "/api/v1/ministerios?lang=fr")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid"`) || !strings.Contains(rec.Body.String(), `"lang"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestGetMinisteriosHandlerOmitsItemsWhenEmpty fija SC-010: sin publicados el
// cuerpo lleva solo `lang` (la clave `items` se omite, la API responde 200 y no
// redirige, P6-21).
func TestGetMinisteriosHandlerOmitsItemsWhenEmpty(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := doGet(t, h.GetMinisterios, "/api/v1/ministerios")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"lang":"es"}` {
		t.Fatalf("cuerpo = %s, se esperaba solo lang", body)
	}
}

// TestGetMinisteriosHandlerNeverLeaksHiddenLeader fija FR-026/SC-014/RG6-1 en la
// capa HTTP: con la ficha del encargado no publicada, el JSON crudo de la
// respuesta no contiene el nombre ni el contacto **en ninguna forma**.
func TestGetMinisteriosHandlerNeverLeaksHiddenLeader(t *testing.T) {
	const (
		name  = "Ana Pérez"
		phone = "+58 412 000 0000"
	)
	repo := newFakeRepository()
	hidden := publishedMinistry("Alabanza")
	hidden.ID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	hidden.LeaderNameEs = textOf(name)
	hidden.LeaderNameEn = textOf("Ana Perez")
	hidden.LeaderEmail = textOf("ana@ejemplo.com")
	repo.seed(hidden)
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetMinisterios, "/api/v1/ministerios")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	raw := rec.Body.String()
	for _, secret := range []string{name, "Ana Perez", "ana@ejemplo.com", "leader", "publicationState"} {
		if strings.Contains(raw, secret) {
			t.Errorf("el JSON crudo contiene %q: %s", secret, raw)
		}
	}
}

// TestGetMinisteriosHandlerServiceError fija el sobre genérico: un fallo interno
// no filtra la causa.
func TestGetMinisteriosHandlerServiceError(t *testing.T) {
	repo := newFakeRepository()
	repo.readErr = errors.New("boom")
	h := newTestHandler(t, repo)

	rec := doGet(t, h.GetMinisterios, "/api/v1/ministerios")
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

// TestGetMinisteriosHandlerWithoutService fija el 500 de composición: un handler
// sin service no revienta la petición.
func TestGetMinisteriosHandlerWithoutService(t *testing.T) {
	logger, _ := testutil.NewLogger()
	h := NewHandler(HandlerDeps{Logger: logger})

	rec := doGet(t, h.GetMinisterios, "/api/v1/ministerios")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}
