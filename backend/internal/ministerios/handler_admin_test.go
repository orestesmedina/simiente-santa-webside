package ministerios

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

// Tests de los handlers del panel de ministerios (T619): el handler real sobre
// el service real, con el repositorio falso en memoria (sin PostgreSQL).
// Comprueban los códigos y sobres de cada operación (200/201/204/400/404/409),
// la validación con `details` por campo, el 401 sin sesión, la semántica del
// `PATCH` (A6-3: un solo campo → el resto intacto; `null` explícito → opcional
// borrado) y que el DTO del panel lleva los datos guardados del encargado con el
// interruptor apagado.

// adminIdentityForTest es la identidad de las pruebas del panel.
func adminIdentityForTest() session.Identity {
	return session.Identity{UserID: uuid.New(), Email: "ana@ejemplo.com", Permissions: []string{"ministerios"}}
}

// withIdentity deja la identidad en el contexto como haría authn.
func withIdentity(req *http.Request, identity session.Identity) *http.Request {
	return req.WithContext(session.ContextWithIdentity(req.Context(), identity))
}

// adminRequest arma una petición del panel con (o sin) identidad.
func adminRequest(method, target, body string, withSession bool) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if withSession {
		req = withIdentity(req, adminIdentityForTest())
	}
	return req
}

// withPathValue deja el id en la ruta como lo haría el router de la stdlib.
func withPathValue(req *http.Request, id string) *http.Request {
	req.SetPathValue("id", id)
	return req
}

// TestListMinisteriosHandlerOK fija el 200 con el sobre paginado (§8.1.2/§8.1.3)
// y que el listado incluye borradores.
func TestListMinisteriosHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(storedMinistry("Alabanza"))
	h := newTestHandler(t, repo)

	rec := httptest.NewRecorder()
	h.ListMinisterios(rec, adminRequest(http.MethodGet, "/api/v1/admin/ministerios", "", true))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got MinisterioAdminList
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es MinisterioAdminList: %v (%s)", err, rec.Body)
	}
	if got.Total != 1 || len(got.Items) != 1 || got.Limit != 20 || got.Offset != 0 {
		t.Fatalf("sobre = %+v", got)
	}
	if got.Items[0].PublicationState != string(StateDraft) {
		t.Errorf("publicationState = %q, el panel ve los borradores", got.Items[0].PublicationState)
	}
	if raw := rec.Body.String(); !strings.Contains(raw, `"whatsappUrlShared":false`) {
		t.Errorf("falta whatsappUrlShared en el DTO del panel: %s", raw)
	}
}

// TestListMinisteriosHandlerInvalidParams fija el 400 con details por campo de
// la paginación y del filtro de estado.
func TestListMinisteriosHandlerInvalidParams(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	tests := []struct {
		target string
		detail string
	}{
		{target: "/api/v1/admin/ministerios?limit=0", detail: "limit"},
		{target: "/api/v1/admin/ministerios?limit=x", detail: "limit"},
		{target: "/api/v1/admin/ministerios?offset=-1", detail: "offset"},
		{target: "/api/v1/admin/ministerios?publicationState=archived", detail: "publicationState"},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ListMinisterios(rec, adminRequest(http.MethodGet, tt.target, "", true))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
		})
	}
}

// TestListMinisteriosHandlerRequiresSession fija el 401 de la red de seguridad
// del handler (la cadena authn responde antes en producción).
func TestListMinisteriosHandlerRequiresSession(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.ListMinisterios(rec, adminRequest(http.MethodGet, "/api/v1/admin/ministerios", "", false))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"unauthenticated"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestCreateMinisterioHandlerOK fija el 201 con el DTO del panel y el guardado
// del DTO validado por el service.
func TestCreateMinisterioHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	h := newTestHandler(t, repo)

	body := `{"nameEs":"  Alabanza  ","descriptionEs":"Grupo de alabanza",
		"leaderNameEs":"Ana Pérez","leaderPhone":"+58 412 000 0000",
		"publicationState":"published","sortOrder":2}`
	rec := httptest.NewRecorder()
	h.CreateMinisterio(rec, adminRequest(http.MethodPost, "/api/v1/admin/ministerios", body, true))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba 201 (%s)", rec.Code, rec.Body)
	}
	var got MinisterioAdmin
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es MinisterioAdmin: %v (%s)", err, rec.Body)
	}
	if got.NameEs != "Alabanza" || got.PublicationState != string(StatePublished) || got.SortOrder != 2 {
		t.Errorf("DTO = %+v", got)
	}
	if got.LeaderVisible {
		t.Error("leaderVisible = true, un alta sin la clave nace oculta (P6-3)")
	}
	if got.LeaderNameEs == nil || *got.LeaderNameEs != "Ana Pérez" {
		t.Error("el DTO del panel debe llevar los datos guardados del encargado")
	}
	if len(repo.ministries) != 1 {
		t.Fatalf("el service no guardó el ministerio: %+v", repo.ministries)
	}
}

// TestCreateMinisterioHandlerInvalid fija el 400 con `details` por campo y que
// un DTO inválido no guarde nada.
func TestCreateMinisterioHandlerInvalid(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		detail string
	}{
		{name: "nombre vacío", body: `{"nameEs":"","descriptionEs":"D"}`, detail: "nameEs"},
		{name: "sin descripción", body: `{"nameEs":"Alabanza"}`, detail: "descriptionEs"},
		{
			name:   "los dos contactos",
			body:   `{"nameEs":"A","descriptionEs":"D","leaderNameEs":"Ana","leaderPhone":"+58 412 000 0000","leaderEmail":"ana@ejemplo.com"}`,
			detail: "leaderEmail",
		},
		{
			name:   "enlace no https",
			body:   `{"nameEs":"A","descriptionEs":"D","whatsappUrl":"javascript:alert(1)"}`,
			detail: "whatsappUrl",
		},
		{
			name:   "imagen sin alt",
			body:   `{"nameEs":"A","descriptionEs":"D","imageFile":"img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.png"}`,
			detail: "imageAltEs",
		},
		{name: "campo desconocido", body: `{"nameEs":"A","descriptionEs":"D","otro":"x"}`, detail: ""},
		{name: "estado fuera del registro", body: `{"nameEs":"A","descriptionEs":"D","publicationState":"archived"}`, detail: "publicationState"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			h := newTestHandler(t, repo)

			rec := httptest.NewRecorder()
			h.CreateMinisterio(rec, adminRequest(http.MethodPost, "/api/v1/admin/ministerios", tt.body, true))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), `"code":"invalid"`) {
				t.Errorf("sobre inesperado: %s", rec.Body)
			}
			if tt.detail != "" && !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
			if len(repo.ministries) != 0 {
				t.Errorf("un DTO inválido no debe guardar nada: %+v", repo.ministries)
			}
		})
	}
}

// TestCreateMinisterioHandlerDuplicate fija el 409 del duplicado exacto (P6-12).
func TestCreateMinisterioHandlerDuplicate(t *testing.T) {
	repo := newFakeRepository()
	repo.duplicate = true
	h := newTestHandler(t, repo)

	rec := httptest.NewRecorder()
	h.CreateMinisterio(rec, adminRequest(http.MethodPost, "/api/v1/admin/ministerios",
		`{"nameEs":"Alabanza","descriptionEs":"D"}`, true))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, se esperaba 409 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"conflict"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestGetMinisterioHandler fija el 200 del detalle y sus dos rechazos: id que no
// es UUID → 400 con details.id; inexistente → 404.
func TestGetMinisterioHandler(t *testing.T) {
	repo := newFakeRepository()
	current := storedMinistry("Alabanza")
	repo.seed(current)
	h := newTestHandler(t, repo)

	t.Run("detalle", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.GetMinisterio(rec, withPathValue(
			adminRequest(http.MethodGet, "/api/v1/admin/ministerios/"+current.ID.String(), "", true),
			current.ID.String(),
		))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
		}
		var got MinisterioAdmin
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo no es MinisterioAdmin: %v", err)
		}
		if got.ID != current.ID.String() || got.NameEs != "Alabanza" {
			t.Errorf("DTO = %+v", got)
		}
	})

	t.Run("id inválido", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.GetMinisterio(rec, withPathValue(
			adminRequest(http.MethodGet, "/api/v1/admin/ministerios/no-es-uuid", "", true), "no-es-uuid",
		))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
		}
		if !strings.Contains(rec.Body.String(), `"id"`) {
			t.Errorf("sin details.id: %s", rec.Body)
		}
	})

	t.Run("inexistente", func(t *testing.T) {
		id := uuid.New().String()
		rec := httptest.NewRecorder()
		h.GetMinisterio(rec, withPathValue(
			adminRequest(http.MethodGet, "/api/v1/admin/ministerios/"+id, "", true), id,
		))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, se esperaba 404 (%s)", rec.Code, rec.Body)
		}
		if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
			t.Errorf("sobre inesperado: %s", rec.Body)
		}
	})
}

// TestUpdateMinisterioHandlerPatchSemantics fija A6-3 en la capa HTTP: el
// handler transmite el DTO parcial tal cual (el service fusiona), un solo campo
// deja el resto intacto y un `null` explícito borra el opcional.
func TestUpdateMinisterioHandlerPatchSemantics(t *testing.T) {
	repo := newFakeRepository()
	current := storedMinistry("Alabanza")
	current.PlaceEs = textOf("Templo")
	current.WhatsappURL = textOf("https://chat.whatsapp.com/abc")
	current.LeaderNameEs = textOf("Ana Pérez")
	current.LeaderPhone = textOf("+58 412 000 0000")
	repo.seed(current)
	h := newTestHandler(t, repo)

	body := `{"nameEs":"Alabanza y adoración","whatsappUrl":null}`
	rec := httptest.NewRecorder()
	h.UpdateMinisterio(rec, withPathValue(
		adminRequest(http.MethodPatch, "/api/v1/admin/ministerios/"+current.ID.String(), body, true),
		current.ID.String(),
	))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got MinisterioAdmin
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es MinisterioAdmin: %v (%s)", err, rec.Body)
	}
	if got.NameEs != "Alabanza y adoración" {
		t.Errorf("nameEs = %q", got.NameEs)
	}
	if got.WhatsappURL != nil {
		t.Errorf("whatsappUrl = %v, se esperaba borrado por el null explícito", got.WhatsappURL)
	}
	if got.PlaceEs == nil || *got.PlaceEs != "Templo" {
		t.Errorf("placeEs = %v, se esperaba intacto (campo ausente)", got.PlaceEs)
	}
	if got.LeaderPhone == nil || *got.LeaderPhone != "+58 412 000 0000" {
		t.Errorf("leaderPhone = %v, se esperaba intacto", got.LeaderPhone)
	}
}

// TestUpdateMinisterioHandlerRejections fija el 400 del PATCH vacío, el 400 con
// details por campo, el 400 del id inválido y el 404 del inexistente.
func TestUpdateMinisterioHandlerRejections(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		wantStatus int
		detail     string
	}{
		{name: "sin campos", id: "11111111-1111-1111-1111-111111111111", body: `{}`, wantStatus: http.StatusBadRequest},
		{
			name:       "estado fuera del registro",
			id:         "11111111-1111-1111-1111-111111111111",
			body:       `{"publicationState":"archived"}`,
			wantStatus: http.StatusBadRequest, detail: "publicationState",
		},
		{
			name:       "enlace no https",
			id:         "11111111-1111-1111-1111-111111111111",
			body:       `{"whatsappUrl":"data:text/html,x"}`,
			wantStatus: http.StatusBadRequest, detail: "whatsappUrl",
		},
		{name: "id inválido", id: "no-es-uuid", body: `{"nameEs":"A"}`, wantStatus: http.StatusBadRequest, detail: "id"},
		{name: "inexistente", id: uuid.New().String(), body: `{"nameEs":"A"}`, wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			current := storedMinistry("Alabanza")
			repo.seed(current)
			h := newTestHandler(t, repo)

			rec := httptest.NewRecorder()
			h.UpdateMinisterio(rec, withPathValue(
				adminRequest(http.MethodPatch, "/api/v1/admin/ministerios/"+tt.id, tt.body, true), tt.id,
			))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, se esperaba %d (%s)", rec.Code, tt.wantStatus, rec.Body)
			}
			if tt.detail != "" && !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
		})
	}
}

// TestDeleteMinisterioHandler fija el 204 sin cuerpo y el 404 del inexistente.
func TestDeleteMinisterioHandler(t *testing.T) {
	repo := newFakeRepository()
	current := storedMinistry("Alabanza")
	repo.seed(current)
	h := newTestHandler(t, repo)

	rec := httptest.NewRecorder()
	h.DeleteMinisterio(rec, withPathValue(
		adminRequest(http.MethodDelete, "/api/v1/admin/ministerios/"+current.ID.String(), "", true),
		current.ID.String(),
	))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, se esperaba 204 (%s)", rec.Code, rec.Body)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("el 204 lleva cuerpo: %s", rec.Body)
	}

	id := uuid.New().String()
	rec = httptest.NewRecorder()
	h.DeleteMinisterio(rec, withPathValue(
		adminRequest(http.MethodDelete, "/api/v1/admin/ministerios/"+id, "", true), id,
	))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404 (%s)", rec.Code, rec.Body)
	}
}

// TestAdminHandlersWithoutService fija el 500 de composición en todas las
// operaciones del panel.
func TestAdminHandlersWithoutService(t *testing.T) {
	logger, _ := testutil.NewLogger()
	h := NewHandler(HandlerDeps{Logger: logger})

	rec := httptest.NewRecorder()
	h.ListMinisterios(rec, adminRequest(http.MethodGet, "/api/v1/admin/ministerios", "", true))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("listado sin service: status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.CreateMinisterio(rec, adminRequest(http.MethodPost, "/api/v1/admin/ministerios",
		`{"nameEs":"A","descriptionEs":"D"}`, true))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("alta sin service: status = %d", rec.Code)
	}

	// Las operaciones con {id} exigen sesión y servicio: el 500 de composición
	// sale igualmente (sin tocar el repositorio).
	id := uuid.New().String()
	rec = httptest.NewRecorder()
	h.GetMinisterio(rec, withPathValue(
		adminRequest(http.MethodGet, "/api/v1/admin/ministerios/"+id, "", true), id,
	))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("detalle sin service: status = %d", rec.Code)
	}
}
