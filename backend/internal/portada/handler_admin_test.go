package portada

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests de los handlers del panel (T324): identidad, «quiénes somos» y
// contacto. El handler real sobre el service real, con el repositorio falso
// (sin PostgreSQL). Comprueban el sobre de éxito, la validación con details por
// campo, el 401 sin sesión y que el service recibe el DTO validado.

func adminIdentityForTest() session.Identity {
	return session.Identity{
		UserID:      uuid.New(),
		Email:       "ana@ejemplo.com",
		Permissions: []string{"portada"},
	}
}

func newAdminHandler(t *testing.T, repo *fakeRepository) *Handler {
	t.Helper()
	logger, _ := testutil.NewLogger()
	return NewHandler(HandlerDeps{
		Service: NewService(ServiceDeps{Repository: repo, Logger: logger}),
		Logger:  logger,
	})
}

func adminRequest(method, target, body string, withIdentity bool) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if withIdentity {
		req = withAdminIdentity(req, adminIdentityForTest())
	}
	return req
}

func TestSaveIdentityHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	h := newAdminHandler(t, repo)

	body := `{"nameEs":"  Iglesia Simiente Santa  ","taglineEs":"Sembrando","publicationState":"published"}`
	rec := httptest.NewRecorder()
	h.SaveIdentity(rec, adminRequest(http.MethodPut, "/api/v1/admin/portada/identidad", body, true))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got IdentityAdmin
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es IdentityAdmin: %v (%s)", err, rec.Body)
	}
	if got.NameEs != "Iglesia Simiente Santa" {
		t.Errorf("nameEs = %q, se esperaba normalizado", got.NameEs)
	}
	if got.PublicationState != string(StatePublished) {
		t.Errorf("publicationState = %q", got.PublicationState)
	}
	if repo.identity == nil || repo.identity.NameEs != "Iglesia Simiente Santa" {
		t.Fatalf("el service no guardó el DTO validado: %+v", repo.identity)
	}
}

func TestSaveIdentityHandlerInvalid(t *testing.T) {
	repo := newFakeRepository()
	h := newAdminHandler(t, repo)

	body := `{"nameEs":"","publicationState":"draft"}`
	rec := httptest.NewRecorder()
	h.SaveIdentity(rec, adminRequest(http.MethodPut, "/api/v1/admin/portada/identidad", body, true))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"nameEs"`) {
		t.Errorf("se esperaba details.nameEs: %s", rec.Body)
	}
	if repo.identity != nil {
		t.Errorf("un DTO inválido no debe guardar nada: %+v", repo.identity)
	}
}

func TestSaveIdentityHandlerMalformedJSON(t *testing.T) {
	h := newAdminHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.SaveIdentity(rec, adminRequest(http.MethodPut, "/api/v1/admin/portada/identidad", `{`, true))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

func TestSaveIdentityHandlerNoSession(t *testing.T) {
	h := newAdminHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.SaveIdentity(rec, adminRequest(http.MethodPut, "/api/v1/admin/portada/identidad", `{}`, false))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"unauthenticated"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

func TestSaveAboutHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	h := newAdminHandler(t, repo)

	body := `{"textEs":"  Somos una iglesia  ","publicationState":"draft"}`
	rec := httptest.NewRecorder()
	h.SaveAbout(rec, adminRequest(http.MethodPut, "/api/v1/admin/portada/quienes-somos", body, true))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if repo.about == nil || repo.about.TextEs != "Somos una iglesia" {
		t.Fatalf("about no guardado: %+v", repo.about)
	}
	if repo.about.PublicationState != StateDraft {
		t.Errorf("publicationState = %q", repo.about.PublicationState)
	}
}

func TestSaveAboutHandlerTooLong(t *testing.T) {
	h := newAdminHandler(t, newFakeRepository())

	body := `{"textEs":"` + strings.Repeat("a", 1001) + `","publicationState":"draft"}`
	rec := httptest.NewRecorder()
	h.SaveAbout(rec, adminRequest(http.MethodPut, "/api/v1/admin/portada/quienes-somos", body, true))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"textEs"`) {
		t.Errorf("se esperaba details.textEs: %s", rec.Body)
	}
}

func TestSaveContactHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	h := newAdminHandler(t, repo)

	body := `{"addressEs":"Calle 1","email":"hola@ejemplo.com","phone":"+58 412-1234567","publicationState":"published"}`
	rec := httptest.NewRecorder()
	h.SaveContact(rec, adminRequest(http.MethodPut, "/api/v1/admin/portada/contacto", body, true))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if repo.contact == nil || repo.contact.Email != "hola@ejemplo.com" {
		t.Fatalf("contacto no guardado: %+v", repo.contact)
	}
	if repo.contact.Phone != "+584121234567" {
		t.Errorf("phone = %q, se esperaba normalizado", repo.contact.Phone)
	}
}

func TestSaveContactHandlerInvalidEmail(t *testing.T) {
	h := newAdminHandler(t, newFakeRepository())

	body := `{"addressEs":"Calle 1","email":"no-es-correo","phone":"+58 412-1234567","publicationState":"published"}`
	rec := httptest.NewRecorder()
	h.SaveContact(rec, adminRequest(http.MethodPut, "/api/v1/admin/portada/contacto", body, true))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"email"`) {
		t.Errorf("se esperaba details.email: %s", rec.Body)
	}
}
