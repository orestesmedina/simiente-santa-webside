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

// --- Colecciones: horario, WhatsApp y redes (T325) ---

// adminPathRequest construye una petición con identidad y el id de ruta.
func adminPathRequest(method, target, body, id string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", id)
	return withAdminIdentity(req, adminIdentityForTest())
}

func TestCreateScheduleHandler(t *testing.T) {
	repo := newFakeRepository()
	h := newAdminHandler(t, repo)

	body := `{"dayOfWeek":0,"startTime":"10:00","nameEs":"Culto","placeEs":"Sede"}`
	rec := httptest.NewRecorder()
	h.CreateSchedule(rec, adminRequest(http.MethodPost, "/api/v1/admin/portada/horario", body, true))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba 201 (%s)", rec.Code, rec.Body)
	}
	var got ScheduleItemAdmin
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es ScheduleItemAdmin: %v (%s)", err, rec.Body)
	}
	if got.NameEs != "Culto" || got.PublicationState != string(StateDraft) {
		t.Fatalf("DTO inesperado: %+v", got)
	}
	if len(repo.services) != 1 {
		t.Fatalf("el service no guardó el elemento: %+v", repo.services)
	}
}

func TestUpdateScheduleHandler(t *testing.T) {
	repo := newFakeRepository()
	seeded := repo.seedService(Service{DayOfWeek: 0, StartTime: "10:00", NameEs: "Culto", PlaceEs: "Sede", PublicationState: StateDraft})
	h := newAdminHandler(t, repo)

	rec := httptest.NewRecorder()
	h.UpdateSchedule(rec, adminPathRequest(http.MethodPatch, "/api/v1/admin/portada/horario/"+seeded.ID.String(),
		`{"nameEs":"Culto dominical"}`, seeded.ID.String()))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if repo.services[0].NameEs != "Culto dominical" {
		t.Errorf("no se aplicó el PATCH: %+v", repo.services[0])
	}
}

func TestUpdateScheduleHandlerEmptyPatch(t *testing.T) {
	repo := newFakeRepository()
	seeded := repo.seedService(Service{DayOfWeek: 0, StartTime: "10:00", NameEs: "Culto", PlaceEs: "Sede", PublicationState: StateDraft})
	h := newAdminHandler(t, repo)

	rec := httptest.NewRecorder()
	h.UpdateSchedule(rec, adminPathRequest(http.MethodPatch, "/api/v1/admin/portada/horario/"+seeded.ID.String(), `{}`, seeded.ID.String()))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

func TestDeleteScheduleHandler(t *testing.T) {
	repo := newFakeRepository()
	seeded := repo.seedService(Service{DayOfWeek: 0, StartTime: "10:00", NameEs: "Culto", PlaceEs: "Sede", PublicationState: StateDraft})
	h := newAdminHandler(t, repo)

	rec := httptest.NewRecorder()
	h.DeleteSchedule(rec, adminPathRequest(http.MethodDelete, "/api/v1/admin/portada/horario/"+seeded.ID.String(), "", seeded.ID.String()))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, se esperaba 204 (%s)", rec.Code, rec.Body)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("204 debe ir sin cuerpo: %s", rec.Body)
	}
	if len(repo.services) != 0 {
		t.Errorf("el elemento no se borró: %+v", repo.services)
	}
}

func TestDeleteScheduleHandlerNotFound(t *testing.T) {
	h := newAdminHandler(t, newFakeRepository())
	missing := uuid.New().String()

	rec := httptest.NewRecorder()
	h.DeleteSchedule(rec, adminPathRequest(http.MethodDelete, "/api/v1/admin/portada/horario/"+missing, "", missing))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404 (%s)", rec.Code, rec.Body)
	}
}

func TestDeleteScheduleHandlerInvalidID(t *testing.T) {
	h := newAdminHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.DeleteSchedule(rec, adminPathRequest(http.MethodDelete, "/api/v1/admin/portada/horario/no-uuid", "", "no-uuid"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"id"`) {
		t.Errorf("se esperaba details.id: %s", rec.Body)
	}
}

func TestCreateWhatsappHandlerDuplicate(t *testing.T) {
	repo := newFakeRepository()
	repo.seedChannel(WhatsappChannel{Kind: KindDirect, Destination: "+584121234567", NameEs: "General", PublicationState: StateDraft})
	h := newAdminHandler(t, repo)

	body := `{"nameEs":"General","kind":"direct","destination":"+58 412-1234567"}`
	rec := httptest.NewRecorder()
	h.CreateWhatsapp(rec, adminRequest(http.MethodPost, "/api/v1/admin/portada/whatsapp", body, true))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, se esperaba 409 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"conflict"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

func TestCreateSocialHandlerDuplicate(t *testing.T) {
	repo := newFakeRepository()
	repo.seedSocialLink(SocialLink{Network: "facebook", URL: "https://facebook.com/simiente", PublicationState: StateDraft})
	h := newAdminHandler(t, repo)

	body := `{"network":"facebook","url":"https://www.facebook.com/otra"}`
	rec := httptest.NewRecorder()
	h.CreateSocial(rec, adminRequest(http.MethodPost, "/api/v1/admin/portada/redes", body, true))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, se esperaba 409 (%s)", rec.Code, rec.Body)
	}
}

func TestCreateSocialHandlerInvalidNetwork(t *testing.T) {
	h := newAdminHandler(t, newFakeRepository())

	body := `{"network":"x","url":"https://x.com/simiente"}`
	rec := httptest.NewRecorder()
	h.CreateSocial(rec, adminRequest(http.MethodPost, "/api/v1/admin/portada/redes", body, true))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"network"`) {
		t.Errorf("se esperaba details.network: %s", rec.Body)
	}
}

func TestUpdateWhatsappHandler(t *testing.T) {
	repo := newFakeRepository()
	seeded := repo.seedChannel(WhatsappChannel{Kind: KindDirect, Destination: "+584121234567", NameEs: "General", PublicationState: StateDraft})
	h := newAdminHandler(t, repo)

	rec := httptest.NewRecorder()
	h.UpdateWhatsapp(rec, adminPathRequest(http.MethodPatch, "/api/v1/admin/portada/whatsapp/"+seeded.ID.String(),
		`{"publicationState":"published"}`, seeded.ID.String()))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if repo.channels[0].PublicationState != StatePublished {
		t.Errorf("no se publicó el canal: %+v", repo.channels[0])
	}
}

// --- Agregado del panel GET /api/v1/admin/portada (T340) ---

func TestGetPortadaAdminHandler(t *testing.T) {
	repo := newFakeRepository()
	repo.identity = &Identity{NameEs: "Iglesia", PublicationState: StatePublished}
	repo.seedService(Service{DayOfWeek: 0, StartTime: "10:00", NameEs: "Culto", PlaceEs: "Sede", PublicationState: StateDraft})
	h := newAdminHandler(t, repo)

	rec := httptest.NewRecorder()
	h.GetPortadaAdmin(rec, adminRequest(http.MethodGet, "/api/v1/admin/portada", "", true))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"identity":{"nameEs":"Iglesia"`) {
		t.Errorf("falta la identidad: %s", body)
	}
	if !strings.Contains(body, `"schedule":{"items":[`) {
		t.Errorf("falta el sobre del horario: %s", body)
	}
	if !strings.Contains(body, `"publicationState":"draft"`) {
		t.Errorf("el panel debe ver borradores: %s", body)
	}
}

func TestGetPortadaAdminHandlerEmpty(t *testing.T) {
	h := newAdminHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.GetPortadaAdmin(rec, adminRequest(http.MethodGet, "/api/v1/admin/portada", "", true))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"identity":null`) {
		t.Errorf("los singletons deben ser null: %s", body)
	}
	if !strings.Contains(body, `"schedule":{"items":[]}`) ||
		!strings.Contains(body, `"whatsapp":{"items":[]}`) ||
		!strings.Contains(body, `"socials":{"items":[]}`) {
		t.Errorf("las colecciones deben ser sobres vacíos, no null: %s", body)
	}
}

func TestGetPortadaAdminHandlerNoSession(t *testing.T) {
	h := newAdminHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.GetPortadaAdmin(rec, adminRequest(http.MethodGet, "/api/v1/admin/portada", "", false))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401 (%s)", rec.Code, rec.Body)
	}
}
