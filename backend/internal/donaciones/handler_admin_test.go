package donaciones

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

// Tests de los handlers del panel de donaciones (T719): el handler real sobre el
// service real, con el repositorio falso en memoria (sin PostgreSQL). Comprueban
// los códigos y sobres de cada operación (200/201/204/400/404/409), la validación
// con `details` por campo, el 401 sin sesión, la normalización del número tal como
// lo escribe el equipo y la semántica del `PATCH` (un solo campo → el resto
// intacto; `null` explícito → opcional borrado).

// adminIdentityForTest es la identidad de las pruebas del panel.
func adminIdentityForTest() session.Identity {
	return session.Identity{UserID: uuid.New(), Email: "ana@ejemplo.com", Permissions: []string{"donaciones"}}
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

// withTextKey deja la clave del texto en la ruta como el router de la stdlib.
func withTextKey(req *http.Request, key string) *http.Request {
	req.SetPathValue("textKey", key)
	return req
}

// TestListDonacionesHandlerOK fija el 200 con el sobre paginado (§8.1.2/§8.1.3)
// y que el listado incluye borradores.
func TestListDonacionesHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	repo.seedMethods(storedMethod())
	h := newTestHandler(t, repo)

	rec := httptest.NewRecorder()
	h.ListDonaciones(rec, adminRequest(http.MethodGet, "/api/v1/admin/donaciones", "", true))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got DonacionAdminList
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es DonacionAdminList: %v (%s)", err, rec.Body)
	}
	if got.Total != 1 || len(got.Items) != 1 || got.Limit != 20 || got.Offset != 0 {
		t.Fatalf("sobre = %+v", got)
	}
	if got.Items[0].PublicationState != string(StateDraft) {
		t.Errorf("publicationState = %q, el panel ve los borradores", got.Items[0].PublicationState)
	}
	if got.Items[0].Number != "CR05015202001234567890" {
		t.Errorf("number = %q, se esperaba el normalizado", got.Items[0].Number)
	}
}

// TestListDonacionesHandlerInvalidParams fija el 400 con details por campo de la
// paginación y del filtro de estado.
func TestListDonacionesHandlerInvalidParams(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	tests := []struct {
		target string
		detail string
	}{
		{target: "/api/v1/admin/donaciones?limit=0", detail: "limit"},
		{target: "/api/v1/admin/donaciones?limit=x", detail: "limit"},
		{target: "/api/v1/admin/donaciones?offset=-1", detail: "offset"},
		{target: "/api/v1/admin/donaciones?publicationState=archived", detail: "publicationState"},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ListDonaciones(rec, adminRequest(http.MethodGet, tt.target, "", true))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
		})
	}
}

// TestListDonacionesHandlerRequiresSession fija el 401 de la red de seguridad del
// handler (la cadena authn responde antes en producción).
func TestListDonacionesHandlerRequiresSession(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.ListDonaciones(rec, adminRequest(http.MethodGet, "/api/v1/admin/donaciones", "", false))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"unauthenticated"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestCreateDonacionHandlerOK fija el 201 con el DTO del panel y que el número
// escrito por el equipo (con espacios y guiones) se guarda normalizado
// (FR-015/FR-016).
func TestCreateDonacionHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	h := newTestHandler(t, repo)

	body := `{"kind":"iban","number":"CR05 0152-0200-1234-5678-90","institution":"  Banco Nacional  ",
		"holderEs":"Ana Pérez","currency":"CRC","publicationState":"published","sortOrder":2}`
	rec := httptest.NewRecorder()
	h.CreateDonacion(rec, adminRequest(http.MethodPost, "/api/v1/admin/donaciones", body, true))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba 201 (%s)", rec.Code, rec.Body)
	}
	var got DonacionAdmin
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es DonacionAdmin: %v (%s)", err, rec.Body)
	}
	if got.Number != "CR05015202001234567890" || got.Institution != "Banco Nacional" {
		t.Errorf("DTO = %+v", got)
	}
	if got.PublicationState != string(StatePublished) || got.SortOrder != 2 || got.Currency == nil || *got.Currency != "CRC" {
		t.Errorf("DTO = %+v", got)
	}
	if len(repo.methods) != 1 || repo.methods[0].Number != "CR05015202001234567890" {
		t.Fatalf("el service no guardó el método normalizado: %+v", repo.methods)
	}
}

// TestCreateDonacionHandlerInvalid fija el 400 con `details` por campo y que un
// DTO inválido no guarde nada (sin guardado parcial).
func TestCreateDonacionHandlerInvalid(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		detail string
	}{
		{name: "sin tipo", body: `{"number":"88888888","institution":"Banco"}`, detail: "kind"},
		{name: "tipo fuera del registro", body: `{"kind":"otro","number":"88888888","institution":"Banco"}`, detail: "kind"},
		{name: "sin entidad", body: `{"kind":"sinpe","number":"88888888","institution":""}`, detail: "institution"},
		{name: "número mal formado", body: `{"kind":"iban","number":"hola","institution":"Banco","currency":"CRC"}`, detail: "number"},
		{name: "sinpe que no es celular", body: `{"kind":"sinpe","number":"22222222","institution":"Banco"}`, detail: "number"},
		{name: "moneda faltante en iban", body: `{"kind":"iban","number":"CR05015202001234567890","institution":"Banco"}`, detail: "currency"},
		{name: "moneda en sinpe", body: `{"kind":"sinpe","number":"88888888","institution":"Banco","currency":"CRC"}`, detail: "currency"},
		{name: "titular en sin español", body: `{"kind":"sinpe","number":"88888888","institution":"Banco","holderEn":"Ana"}`, detail: "holderEs"},
		{name: "entidad demasiado larga", body: `{"kind":"sinpe","number":"88888888","institution":"` + strings.Repeat("a", 161) + `"}`, detail: "institution"},
		{name: "estado fuera del registro", body: `{"kind":"sinpe","number":"88888888","institution":"Banco","publicationState":"archived"}`, detail: "publicationState"},
		{name: "campo desconocido", body: `{"kind":"sinpe","number":"88888888","institution":"Banco","otro":"x"}`, detail: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			h := newTestHandler(t, repo)

			rec := httptest.NewRecorder()
			h.CreateDonacion(rec, adminRequest(http.MethodPost, "/api/v1/admin/donaciones", tt.body, true))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), `"code":"invalid"`) {
				t.Errorf("sobre inesperado: %s", rec.Body)
			}
			if tt.detail != "" && !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
			if len(repo.methods) != 0 {
				t.Errorf("un DTO inválido no debe guardar nada: %+v", repo.methods)
			}
		})
	}
}

// TestCreateDonacionHandlerDuplicate fija el 409 del duplicado exacto (P7-10).
func TestCreateDonacionHandlerDuplicate(t *testing.T) {
	repo := newFakeRepository()
	repo.duplicate = true
	h := newTestHandler(t, repo)

	rec := httptest.NewRecorder()
	h.CreateDonacion(rec, adminRequest(http.MethodPost, "/api/v1/admin/donaciones",
		`{"kind":"sinpe","number":"88888888","institution":"Banco"}`, true))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, se esperaba 409 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"conflict"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestGetDonacionHandler fija el 200 del detalle y sus dos rechazos: id que no es
// UUID → 400 con details.id; inexistente → 404.
func TestGetDonacionHandler(t *testing.T) {
	repo := newFakeRepository()
	current := storedMethod()
	repo.seedMethods(current)
	h := newTestHandler(t, repo)

	t.Run("detalle", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.GetDonacion(rec, withPathValue(
			adminRequest(http.MethodGet, "/api/v1/admin/donaciones/"+current.ID.String(), "", true),
			current.ID.String(),
		))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
		}
		var got DonacionAdmin
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo no es DonacionAdmin: %v", err)
		}
		if got.ID != current.ID.String() || got.Institution != "Banco Nacional" {
			t.Errorf("DTO = %+v", got)
		}
	})

	t.Run("id inválido", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.GetDonacion(rec, withPathValue(
			adminRequest(http.MethodGet, "/api/v1/admin/donaciones/no-es-uuid", "", true), "no-es-uuid",
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
		h.GetDonacion(rec, withPathValue(
			adminRequest(http.MethodGet, "/api/v1/admin/donaciones/"+id, "", true), id,
		))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, se esperaba 404 (%s)", rec.Code, rec.Body)
		}
		if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
			t.Errorf("sobre inesperado: %s", rec.Body)
		}
	})
}

// TestUpdateDonacionHandlerPatchSemantics fija R7-15 en la capa HTTP: el handler
// transmite el DTO parcial tal cual (el service fusiona), un solo campo deja el
// resto intacto y un `null` explícito borra el opcional.
func TestUpdateDonacionHandlerPatchSemantics(t *testing.T) {
	repo := newFakeRepository()
	current := storedMethod()
	repo.seedMethods(current)
	h := newTestHandler(t, repo)

	body := `{"institution":"Banco Nuevo","holderEs":null}`
	rec := httptest.NewRecorder()
	h.UpdateDonacion(rec, withPathValue(
		adminRequest(http.MethodPatch, "/api/v1/admin/donaciones/"+current.ID.String(), body, true),
		current.ID.String(),
	))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got DonacionAdmin
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es DonacionAdmin: %v (%s)", err, rec.Body)
	}
	if got.Institution != "Banco Nuevo" {
		t.Errorf("institution = %q", got.Institution)
	}
	if got.HolderEs != nil {
		t.Errorf("holderEs = %v, se esperaba borrado por el null explícito", got.HolderEs)
	}
	if got.Number != "CR05015202001234567890" {
		t.Errorf("number = %q, se esperaba intacto (campo ausente)", got.Number)
	}
	if got.Currency == nil || *got.Currency != "CRC" {
		t.Errorf("currency = %v, se esperaba intacta", got.Currency)
	}
}

// TestUpdateDonacionHandlerRejections fija el 400 con details por campo, el 400
// del id inválido y el 404 del inexistente.
func TestUpdateDonacionHandlerRejections(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		wantStatus int
		detail     string
	}{
		{
			name:       "número mal formado",
			id:         "11111111-1111-1111-1111-111111111111",
			body:       `{"number":"hola"}`,
			wantStatus: http.StatusBadRequest, detail: "number",
		},
		{
			name:       "estado fuera del registro",
			id:         "11111111-1111-1111-1111-111111111111",
			body:       `{"publicationState":"archived"}`,
			wantStatus: http.StatusBadRequest, detail: "publicationState",
		},
		{name: "id inválido", id: "no-es-uuid", body: `{"institution":"A"}`, wantStatus: http.StatusBadRequest, detail: "id"},
		{name: "inexistente", id: uuid.New().String(), body: `{"institution":"A"}`, wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.seedMethods(storedMethod())
			h := newTestHandler(t, repo)

			rec := httptest.NewRecorder()
			h.UpdateDonacion(rec, withPathValue(
				adminRequest(http.MethodPatch, "/api/v1/admin/donaciones/"+tt.id, tt.body, true), tt.id,
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

// TestDeleteDonacionHandler fija el 204 sin cuerpo y el 404 del inexistente.
func TestDeleteDonacionHandler(t *testing.T) {
	repo := newFakeRepository()
	current := storedMethod()
	repo.seedMethods(current)
	h := newTestHandler(t, repo)

	rec := httptest.NewRecorder()
	h.DeleteDonacion(rec, withPathValue(
		adminRequest(http.MethodDelete, "/api/v1/admin/donaciones/"+current.ID.String(), "", true),
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
	h.DeleteDonacion(rec, withPathValue(
		adminRequest(http.MethodDelete, "/api/v1/admin/donaciones/"+id, "", true), id,
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
	h.ListDonaciones(rec, adminRequest(http.MethodGet, "/api/v1/admin/donaciones", "", true))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("listado sin service: status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.CreateDonacion(rec, adminRequest(http.MethodPost, "/api/v1/admin/donaciones",
		`{"kind":"sinpe","number":"88888888","institution":"Banco"}`, true))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("alta sin service: status = %d", rec.Code)
	}

	id := uuid.New().String()
	rec = httptest.NewRecorder()
	h.GetDonacion(rec, withPathValue(
		adminRequest(http.MethodGet, "/api/v1/admin/donaciones/"+id, "", true), id,
	))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("detalle sin service: status = %d", rec.Code)
	}
}

// --- Textos de la sección (T720, FR-021) ---

// TestGetTextosHandler fija el 200 con los dos singletons (en `null` hasta el
// primer guardado, patrón F3) y su contenido una vez guardados.
func TestGetTextosHandler(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.GetTextos(rec, adminRequest(http.MethodGet, "/api/v1/admin/donaciones/textos", "", true))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"intro":null,"usage":null}` {
		t.Fatalf("cuerpo = %s, se esperaba los dos singletons en null", body)
	}

	repo := newFakeRepository()
	repo.seedTexts(TextoDonacion{
		TextKey:          TextKeyIntro,
		BodyEs:           "Cómo donar",
		PublicationState: StatePublished,
	})
	h = newTestHandler(t, repo)
	rec = httptest.NewRecorder()
	h.GetTextos(rec, adminRequest(http.MethodGet, "/api/v1/admin/donaciones/textos", "", true))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"textEs":"Cómo donar"`) {
		t.Errorf("cuerpo = %s, se esperaba el texto introductorio", rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"intro":null`) && !strings.Contains(rec.Body.String(), `"usage":null`) {
		t.Errorf("cuerpo = %s, se esperaba la clave no guardada en null", rec.Body)
	}
}

// TestPutTextoHandlerCreatesAndReplaces fija el 200 del guardado (crea el
// singleton y luego lo reemplaza), con `textEs` obligatorio y `textEn` opcional,
// y que el texto se guarda íntegro (tildes, ñ, signos y saltos de línea).
func TestPutTextoHandlerCreatesAndReplaces(t *testing.T) {
	repo := newFakeRepository()
	h := newTestHandler(t, repo)

	body := `{"textEs":"¿Cómo donar?\nPrimera línea","textEn":"How to give","publicationState":"published"}`
	rec := httptest.NewRecorder()
	h.PutTexto(rec, withTextKey(
		adminRequest(http.MethodPut, "/api/v1/admin/donaciones/textos/intro", body, true), "intro",
	))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got DonacionTextoAdmin
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es DonacionTextoAdmin: %v (%s)", err, rec.Body)
	}
	if got.TextKey != "intro" || got.TextEs != "¿Cómo donar?\nPrimera línea" || got.PublicationState != string(StatePublished) {
		t.Errorf("DTO = %+v", got)
	}
	if got.TextEn == nil || *got.TextEn != "How to give" {
		t.Errorf("textEn = %v", got.TextEn)
	}
	if len(repo.texts) != 1 {
		t.Fatalf("el singleton no se guardó: %+v", repo.texts)
	}

	// El segundo guardado reemplaza el singleton en vez de duplicarlo.
	rec = httptest.NewRecorder()
	h.PutTexto(rec, withTextKey(
		adminRequest(http.MethodPut, "/api/v1/admin/donaciones/textos/intro", `{"textEs":"Otro texto"}`, true), "intro",
	))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if len(repo.texts) != 1 || repo.texts[0].BodyEs != "Otro texto" {
		t.Fatalf("el singleton no se reemplazó: %+v", repo.texts)
	}
}

// TestPutTextoHandlerInvalid fija el 400 con details por campo y la clave fuera
// del catálogo.
func TestPutTextoHandlerInvalid(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		body   string
		detail string
	}{
		{name: "sin texto español", key: "intro", body: `{"textEn":"How to give"}`, detail: "textEs"},
		{name: "texto en inglés sin español", key: "intro", body: `{"textEs":"","textEn":"How to give"}`, detail: "textEs"},
		{name: "texto demasiado largo", key: "intro", body: `{"textEs":"` + strings.Repeat("a", 2001) + `"}`, detail: "textEs"},
		{name: "clave fuera del catálogo", key: "otro", body: `{"textEs":"Texto"}`, detail: "textKey"},
		{name: "campo desconocido", key: "intro", body: `{"textEs":"Texto","otro":"x"}`, detail: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			h := newTestHandler(t, repo)

			rec := httptest.NewRecorder()
			h.PutTexto(rec, withTextKey(
				adminRequest(http.MethodPut, "/api/v1/admin/donaciones/textos/"+tt.key, tt.body, true), tt.key,
			))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), `"code":"invalid"`) {
				t.Errorf("sobre inesperado: %s", rec.Body)
			}
			if tt.detail != "" && !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
			if len(repo.texts) != 0 {
				t.Errorf("un DTO inválido no debe guardar nada: %+v", repo.texts)
			}
		})
	}
}

// TestPutTextoHandlerRequiresSession fija el 401 de la red de seguridad del
// handler (la cadena authn responde antes en producción).
func TestPutTextoHandlerRequiresSession(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.PutTexto(rec, withTextKey(
		adminRequest(http.MethodPut, "/api/v1/admin/donaciones/textos/intro", `{"textEs":"Texto"}`, false), "intro",
	))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"unauthenticated"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}
