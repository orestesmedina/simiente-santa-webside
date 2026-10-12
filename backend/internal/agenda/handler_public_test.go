package agenda

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests del handler público GET /api/v1/proximamente (T423) y de los helpers
// HTTP del dominio: el handler real sobre el service real, con el repositorio
// falso en memoria (sin PostgreSQL). Comprueban el sobre de éxito, la cabecera
// no-store, la validación de `lang`, la omisión de bloques vacíos, el fallo del
// service y la validación de la entrada de los helpers compartidos.

// newTestHandler construye el handler sobre el service real y el almacén falso.
func newTestHandler(t *testing.T, repo *fakeRepository) *Handler {
	t.Helper()
	logger, _ := testutil.NewLogger()
	location, err := time.LoadLocation("America/Costa_Rica")
	if err != nil {
		t.Fatalf("cargar America/Costa_Rica: %v", err)
	}
	svc := NewService(ServiceDeps{
		Repository: repo,
		Store:      &fakeStore{files: map[string][]byte{}},
		Location:   location,
		Now:        func() time.Time { return time.Date(2026, time.October, 11, 18, 0, 0, 0, time.UTC) },
		Logger:     logger,
	})
	return NewHandler(HandlerDeps{Service: svc, MaxUploadBytes: 1 << 20, Logger: logger})
}

// withIdentity deja la identidad en el contexto como haría authn.
func withIdentity(req *http.Request, identity session.Identity) *http.Request {
	return req.WithContext(session.ContextWithIdentity(req.Context(), identity))
}

// adminIdentityForTest es la identidad de las pruebas del panel.
func adminIdentityForTest() session.Identity {
	return session.Identity{UserID: uuid.New(), Email: "ana@ejemplo.com", Permissions: []string{"eventos"}}
}

// doGetProximamente lanza el handler público.
func doGetProximamente(t *testing.T, h *Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	h.GetProximamente(rec, req)
	return rec
}

// TestGetProximamenteHandlerOK fija el sobre de éxito, la cabecera no-store y el
// contenido público (solo publicado y próximo).
func TestGetProximamenteHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	event := samplePublicEvent("Noche de aposentos", "2026-10-12")
	repo.events = []Event{event}
	repo.eventDates[event.ID] = []EventDate{{Date: "2026-10-12"}}
	repo.activities = []Activity{samplePublicActivity("Repartir comida", "2026-10-12")}

	h := newTestHandler(t, repo)

	rec := doGetProximamente(t, h, "/api/v1/proximamente")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, se esperaba no-store", got)
	}

	var got PublicProximamente
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es ProximamentePublica: %v (%s)", err, rec.Body)
	}
	if got.Lang != LangES {
		t.Errorf("lang = %q, se esperaba es (por defecto)", got.Lang)
	}
	if len(got.Events) != 1 || got.Events[0].Title != "Noche de aposentos" {
		t.Fatalf("eventos = %+v", got.Events)
	}
	if len(got.Activities) != 1 || got.Activities[0].Title != "Repartir comida" {
		t.Fatalf("actividades = %+v", got.Activities)
	}
	if strings.Contains(rec.Body.String(), "publicationState") {
		t.Errorf("el DTO público no debe exponer el estado: %s", rec.Body)
	}
}

// TestGetProximamenteHandlerLang fija la resolución de idioma en el handler
// (incluida la normalización a minúsculas) y el 400 de un idioma desconocido.
func TestGetProximamenteHandlerLang(t *testing.T) {
	repo := newFakeRepository()
	event := samplePublicEvent("Noche de aposentos", "2026-10-12")
	event.TitleEn = ptrOf("Dormitory night")
	repo.events = []Event{event}
	repo.eventDates[event.ID] = []EventDate{{Date: "2026-10-12"}}
	h := newTestHandler(t, repo)

	rec := doGetProximamente(t, h, "/api/v1/proximamente?lang=EN")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got PublicProximamente
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo inesperado: %v", err)
	}
	if got.Lang != LangEN || got.Events[0].Title != "Dormitory night" {
		t.Errorf("respuesta = %+v", got)
	}

	rec = doGetProximamente(t, h, "/api/v1/proximamente?lang=fr")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("lang=fr: status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid"`) || !strings.Contains(rec.Body.String(), `"lang"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestGetProximamenteHandlerOmitsEmptyBlocks fija SC-012: sin nada publicado y
// próximo, el documento solo lleva `lang`.
func TestGetProximamenteHandlerOmitsEmptyBlocks(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := doGetProximamente(t, h, "/api/v1/proximamente")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"lang":"es"}` {
		t.Fatalf("cuerpo = %s, se esperaba solo lang", body)
	}
}

// TestGetProximamenteHandlerServiceError fija el sobre de error genérico sin
// filtrar la causa interna.
func TestGetProximamenteHandlerServiceError(t *testing.T) {
	repo := newFakeRepository()
	repo.listErr = errors.New("boom")
	h := newTestHandler(t, repo)

	rec := doGetProximamente(t, h, "/api/v1/proximamente")
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

// TestHandlerWithoutService responde 500 (error de composición) cuando el
// servicio no está cableado, también en la ruta pública.
func TestHandlerWithoutService(t *testing.T) {
	h := NewHandler(HandlerDeps{})

	rec := doGetProximamente(t, h, "/api/v1/proximamente")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestHandlerHelpersRequireSession fija el 401 de red de seguridad de los
// helpers del panel cuando falta la identidad de authn.
func TestHandlerHelpersRequireSession(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/eventos", strings.NewReader("{}"))

	t.Run("createItem", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.createItem(rec, req, func(context.Context, uuid.UUID) (any, error) { return nil, nil })
		if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), `"code":"unauthenticated"`) {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("updateItem", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.updateItem(rec, req, func(context.Context, uuid.UUID) (any, error) { return nil, nil })
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("deleteItem", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.deleteItem(rec, req, func(context.Context, uuid.UUID) error { return nil })
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})
}

// TestHandlerCollectionWrappers fija los sobres 201/200/204 de los envoltorios
// de las colecciones y que un error del service viaja con su sobre.
func TestHandlerCollectionWrappers(t *testing.T) {
	repo := newFakeRepository()
	h := newTestHandler(t, repo)
	identity := adminIdentityForTest()

	t.Run("201 en el alta", func(t *testing.T) {
		req := withIdentity(httptest.NewRequest(http.MethodPost, "/api/v1/admin/eventos", nil), identity)
		rec := httptest.NewRecorder()
		h.createItem(rec, req, func(_ context.Context, actorID uuid.UUID) (any, error) {
			if actorID != identity.UserID {
				t.Errorf("actorID = %v, se esperaba el de la sesión", actorID)
			}
			return map[string]string{"id": "x"}, nil
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("200 en la edición", func(t *testing.T) {
		req := withIdentity(httptest.NewRequest(http.MethodPatch, "/api/v1/admin/eventos/x", nil), identity)
		rec := httptest.NewRecorder()
		h.updateItem(rec, req, func(context.Context, uuid.UUID) (any, error) {
			return map[string]string{"id": "x"}, nil
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("204 en el borrado", func(t *testing.T) {
		req := withIdentity(httptest.NewRequest(http.MethodDelete, "/api/v1/admin/eventos/x", nil), identity)
		rec := httptest.NewRecorder()
		h.deleteItem(rec, req, func(context.Context, uuid.UUID) error { return nil })
		if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
			t.Fatalf("status = %d, cuerpo = %q", rec.Code, rec.Body)
		}
	})

	t.Run("error del service con su sobre", func(t *testing.T) {
		req := withIdentity(httptest.NewRequest(http.MethodPost, "/api/v1/admin/eventos", nil), identity)
		rec := httptest.NewRecorder()
		h.createItem(rec, req, func(context.Context, uuid.UUID) (any, error) {
			return nil, apperrConflictForTest()
		})
		if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), `"code":"conflict"`) {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("errAlreadyHandled no responde dos veces", func(t *testing.T) {
		req := withIdentity(httptest.NewRequest(http.MethodPost, "/api/v1/admin/eventos", nil), identity)
		rec := httptest.NewRecorder()
		h.createItem(rec, req, func(context.Context, uuid.UUID) (any, error) {
			return nil, errAlreadyHandled
		})
		if rec.Code != http.StatusOK || rec.Body.Len() != 0 {
			t.Fatalf("status = %d, cuerpo = %q (el envoltorio no debe responder)", rec.Code, rec.Body)
		}
	})
}

// TestHandlerDecodeAndValidate fija la validación de forma del cuerpo (CWE-20).
func TestHandlerDecodeAndValidate(t *testing.T) {
	repo := newFakeRepository()
	h := newTestHandler(t, repo)
	identity := adminIdentityForTest()

	tests := []struct {
		name   string
		body   string
		wantOK bool
		detail string
	}{
		{name: "DTO válido", body: `{"titleEs":"Evento","detailEs":"Detalle","placeEs":"Templo","dates":[{"date":"2026-10-12"}]}`, wantOK: true},
		{name: "JSON inválido", body: `{`, wantOK: false},
		{name: "campo desconocido", body: `{"titleEs":"Evento","otro":1}`, wantOK: false},
		{name: "dos objetos", body: `{"titleEs":"A"}{"titleEs":"B"}`, wantOK: false},
		{name: "DTO inválido", body: `{"titleEs":"","detailEs":"","placeEs":""}`, wantOK: false, detail: "titleEs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := withIdentity(httptest.NewRequest(http.MethodPost, "/api/v1/admin/eventos", strings.NewReader(tt.body)), identity)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			var in EventoInput
			ok := h.decodeAndValidate(rec, req, &in)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, se esperaba %v (%s)", ok, tt.wantOK, rec.Body)
			}
			if tt.wantOK {
				if rec.Body.Len() != 0 {
					t.Errorf("un DTO válido no debe escribir nada: %s", rec.Body)
				}
				return
			}
			if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"code":"invalid"`) {
				t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
			}
			if tt.detail != "" && !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
		})
	}

	// Los rechazos quedan registrados best-effort con result='failure' (R4-10).
	if len(repo.recorded) == 0 {
		t.Error("los rechazos deben quedar registrados")
	}
}

// TestHandlerAgendaItemID fija la lectura del id de la ruta.
func TestHandlerAgendaItemID(t *testing.T) {
	id := uuid.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/eventos/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	got, err := agendaItemIDFromRequest(req)
	if err != nil || got != id {
		t.Fatalf("id = %v (%v)", got, err)
	}

	req = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/eventos/no-es-uuid", nil)
	req.SetPathValue("id", "no-es-uuid")
	_, err = agendaItemIDFromRequest(req)
	if errorKind(err) != apperr.KindInvalid {
		t.Fatalf("id inválido = %v, se esperaba invalid", err)
	}
	if _, ok := detailsOf(t, err)["id"]; !ok {
		t.Errorf("sin details.id: %v", err)
	}
}

// TestHandlerListParams fija la lectura de la paginación y del filtro de estado.
func TestHandlerListParams(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	tests := []struct {
		name      string
		query     string
		wantLimit int
		wantState *PublicationState
		wantErr   bool
		detail    string
	}{
		{name: "sin parámetros", query: "", wantLimit: 20},
		{name: "con límites", query: "?limit=5&offset=10", wantLimit: 5},
		{name: "tope duro", query: "?limit=1000", wantLimit: 100},
		{name: "filtro published", query: "?publicationState=published", wantLimit: 20, wantState: statePtr(StatePublished)},
		{name: "filtro draft", query: "?publicationState=draft", wantLimit: 20, wantState: statePtr(StateDraft)},
		{name: "límite inválido", query: "?limit=cero", wantErr: true, detail: "limit"},
		{name: "offset negativo", query: "?offset=-1", wantErr: true, detail: "offset"},
		{name: "filtro inválido", query: "?publicationState=archived", wantErr: true, detail: "publicationState"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/eventos"+tt.query, nil)
			params, state, err := h.listParams(req)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("se esperaba error")
				}
				if !strings.Contains(err.Error(), tt.detail) && !strings.Contains(err.Error(), "válido") {
					t.Errorf("el error no identifica el campo %q: %v", tt.detail, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("listParams: %v", err)
			}
			if params.Limit != tt.wantLimit {
				t.Errorf("limit = %d, se esperaba %d", params.Limit, tt.wantLimit)
			}
			switch {
			case tt.wantState == nil && state != nil:
				t.Errorf("state = %q, se esperaba sin filtro", *state)
			case tt.wantState != nil && state == nil:
				t.Errorf("state = nil, se esperaba el filtro %q", *tt.wantState)
			case tt.wantState != nil && state != nil && *state != *tt.wantState:
				t.Errorf("state = %q, se esperaba %q", *state, *tt.wantState)
			}
		})
	}

	// El filtro válido llega con el valor pedido.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/eventos?publicationState=published", nil)
	_, state, err := h.listParams(req)
	if err != nil || state == nil || *state != StatePublished {
		t.Fatalf("filtro = %v (%v)", state, err)
	}
}

// statePtr devuelve un puntero al estado de publicación indicado.
func statePtr(state PublicationState) *PublicationState { return &state }

// apperrConflictForTest devuelve un conflicto de dominio para las pruebas.
func apperrConflictForTest() error {
	return apperr.Conflict("Ya existe un evento con esos datos")
}

// --- Extracto de la portada (US7/FR-025, T446) ---

// doGetPortadaProximamente lanza el handler del extracto.
func doGetPortadaProximamente(t *testing.T, h *Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	h.GetPortadaProximamente(rec, req)
	return rec
}

// TestGetPortadaProximamenteHandlerOK fija el sobre de éxito, la cabecera
// no-store y el contenido del extracto (solo publicado y próximo, con su tope).
func TestGetPortadaProximamenteHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	event := samplePublicEvent("Noche de aposentos", "2026-10-12")
	event.ImageFile = ptrOf("img_11111111-1111-1111-1111-111111111111.jpg")
	event.ImageAltEs = ptrOf("Foto del evento")
	repo.events = []Event{event}
	repo.eventDates[event.ID] = []EventDate{{Date: "2026-10-12", StartTime: ptrOf("19:00")}}
	repo.activities = []Activity{samplePublicActivity("Repartir comida", "2026-10-13")}

	h := newTestHandler(t, repo)

	rec := doGetPortadaProximamente(t, h, "/api/v1/portada/proximamente")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, se esperaba no-store", got)
	}

	var got PortadaProximamente
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es PortadaProximamente: %v (%s)", err, rec.Body)
	}
	if got.Lang != LangES {
		t.Errorf("lang = %q, se esperaba es (por defecto)", got.Lang)
	}
	if len(got.Items) != 2 {
		t.Fatalf("items = %+v, se esperaban 2", got.Items)
	}
	if got.Items[0].Kind != "event" || got.Items[0].Title != "Noche de aposentos" {
		t.Errorf("items[0] = %+v", got.Items[0])
	}
	if got.Items[0].StartTime == nil || *got.Items[0].StartTime != "19:00" {
		t.Errorf("startTime = %v", got.Items[0].StartTime)
	}
	if got.Items[0].Image == nil || got.Items[0].Image.Alt != "Foto del evento" {
		t.Errorf("imagen = %+v", got.Items[0].Image)
	}
	if got.Items[1].Kind != "activity" || got.Items[1].Title != "Repartir comida" {
		t.Errorf("items[1] = %+v", got.Items[1])
	}
	// El extracto no lleva detalle, categoría ni lista de fechas.
	for _, forbidden := range []string{"detail", "category", "dates", "publicationState"} {
		if strings.Contains(rec.Body.String(), forbidden) {
			t.Errorf("el extracto no debe llevar %q: %s", forbidden, rec.Body)
		}
	}
}

// TestGetPortadaProximamenteHandlerOmitsEmptyItems fija FR-026 en el sobre: sin
// ítems la respuesta solo lleva `lang` (clave `items` omitida, nunca null).
func TestGetPortadaProximamenteHandlerOmitsEmptyItems(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := doGetPortadaProximamente(t, h, "/api/v1/portada/proximamente")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"lang":"es"}` {
		t.Fatalf("cuerpo = %s, se esperaba solo lang", body)
	}
	if strings.Contains(rec.Body.String(), "null") {
		t.Errorf("items nunca debe ser null: %s", rec.Body)
	}
}

// TestGetPortadaProximamenteHandlerLang fija la resolución de idioma del handler
// y el 400 de un idioma desconocido.
func TestGetPortadaProximamenteHandlerLang(t *testing.T) {
	repo := newFakeRepository()
	event := samplePublicEvent("Noche de aposentos", "2026-10-12")
	event.TitleEn = ptrOf("Dormitory night")
	repo.events = []Event{event}
	repo.eventDates[event.ID] = []EventDate{{Date: "2026-10-12"}}
	h := newTestHandler(t, repo)

	rec := doGetPortadaProximamente(t, h, "/api/v1/portada/proximamente?lang=EN")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got PortadaProximamente
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo inesperado: %v", err)
	}
	if got.Lang != LangEN || got.Items[0].Title != "Dormitory night" {
		t.Errorf("respuesta = %+v", got)
	}

	rec = doGetPortadaProximamente(t, h, "/api/v1/portada/proximamente?lang=fr")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("lang=fr: status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid"`) || !strings.Contains(rec.Body.String(), `"lang"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestGetPortadaProximamenteHandlerServiceError fija el sobre de error genérico
// sin filtrar la causa interna, y el 500 si el servicio no está cableado.
func TestGetPortadaProximamenteHandlerServiceError(t *testing.T) {
	repo := newFakeRepository()
	repo.listErr = errors.New("boom")
	h := newTestHandler(t, repo)

	rec := doGetPortadaProximamente(t, h, "/api/v1/portada/proximamente")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("la respuesta filtra la causa interna: %s", rec.Body)
	}

	rec = doGetPortadaProximamente(t, NewHandler(HandlerDeps{}), "/api/v1/portada/proximamente")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("sin servicio: status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
}
