package agenda

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Tests de los handlers del panel de eventos (T424): el handler real sobre el
// service real, con el repositorio falso en memoria (sin PostgreSQL).
// Comprueban los códigos y sobres de cada operación, la validación con details
// por campo, el 401 sin sesión y que el DTO del panel traiga `nextDate` siempre
// presente (nullable) e `isPast`.

// newAdminHandler construye el handler sobre el service real y el almacén falso.
func newAdminHandler(t *testing.T, repo *fakeRepository) *Handler {
	t.Helper()
	return newTestHandler(t, repo)
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

// TestCreateEventoHandlerOK fija el 201 con el DTO del panel y el guardado del
// DTO validado por el service.
func TestCreateEventoHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	h := newAdminHandler(t, repo)

	body := `{"titleEs":"  Noche de aposentos  ","detailEs":"Detalle","placeEs":"Templo",
		"dates":[{"date":"2026-10-12","startTime":"19:00","endTime":"21:00"}],"publicationState":"published"}`
	rec := httptest.NewRecorder()
	h.CreateEvento(rec, adminRequest(http.MethodPost, "/api/v1/admin/eventos", body, true))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba 201 (%s)", rec.Code, rec.Body)
	}
	var got EventoAdmin
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es EventoAdmin: %v (%s)", err, rec.Body)
	}
	if got.TitleEs != "Noche de aposentos" {
		t.Errorf("titleEs = %q, se esperaba normalizado", got.TitleEs)
	}
	if got.PublicationState != string(StatePublished) {
		t.Errorf("publicationState = %q", got.PublicationState)
	}
	if len(got.Dates) != 1 || got.Dates[0].IsPast {
		t.Errorf("dates = %+v", got.Dates)
	}
	if got.NextDate == nil || *got.NextDate != "2026-10-12" || got.IsPast {
		t.Errorf("nextDate/isPast = %v/%v", got.NextDate, got.IsPast)
	}
	if len(repo.events) != 1 {
		t.Fatalf("el service no guardó el evento: %+v", repo.events)
	}
}

// TestCreateEventoHandlerInvalid fija el 400 con `details` por campo y que un
// DTO inválido no guarde nada.
func TestCreateEventoHandlerInvalid(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		detail string
	}{
		{name: "título vacío", body: `{"titleEs":"","detailEs":"D","placeEs":"P","dates":[{"date":"2026-10-12"}]}`, detail: "titleEs"},
		{name: "sin fechas", body: `{"titleEs":"T","detailEs":"D","placeEs":"P","dates":[]}`, detail: "dates"},
		{name: "fecha inexistente", body: `{"titleEs":"T","detailEs":"D","placeEs":"P","dates":[{"date":"2026-13-32"}]}`, detail: "date"},
		{name: "fin sin inicio", body: `{"titleEs":"T","detailEs":"D","placeEs":"P","dates":[{"date":"2026-10-12","endTime":"21:00"}]}`, detail: "endTime"},
		{name: "campo desconocido", body: `{"titleEs":"T","otro":"x"}`, detail: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			h := newAdminHandler(t, repo)

			rec := httptest.NewRecorder()
			h.CreateEvento(rec, adminRequest(http.MethodPost, "/api/v1/admin/eventos", tt.body, true))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), `"code":"invalid"`) {
				t.Errorf("sobre inesperado: %s", rec.Body)
			}
			if tt.detail != "" && !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
			if len(repo.events) != 0 {
				t.Errorf("un DTO inválido no debe guardar nada: %+v", repo.events)
			}
		})
	}
}

// TestCreateEventoHandlerDuplicate fija el 409 del duplicado del módulo (R4-9).
func TestCreateEventoHandlerDuplicate(t *testing.T) {
	repo := newFakeRepository()
	repo.duplicate = true
	h := newAdminHandler(t, repo)

	body := `{"titleEs":"Noche de aposentos","detailEs":"D","placeEs":"P","dates":[{"date":"2026-10-12"}]}`
	rec := httptest.NewRecorder()
	h.CreateEvento(rec, adminRequest(http.MethodPost, "/api/v1/admin/eventos", body, true))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, se esperaba 409 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"conflict"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestEventoHandlersRequireSession fija el 401 de red de seguridad de todas las
// operaciones del panel (la cadena real la monta routes.go, T428).
func TestEventoHandlersRequireSession(t *testing.T) {
	h := newAdminHandler(t, newFakeRepository())
	id := uuid.New().String()

	tests := []struct {
		name   string
		method string
		path   string
		call   func(h *Handler, rec *httptest.ResponseRecorder, req *http.Request)
	}{
		{name: "listado", method: http.MethodGet, path: "/api/v1/admin/eventos", call: func(h *Handler, rec *httptest.ResponseRecorder, req *http.Request) {
			h.ListEventos(rec, req)
		}},
		{name: "alta", method: http.MethodPost, path: "/api/v1/admin/eventos", call: func(h *Handler, rec *httptest.ResponseRecorder, req *http.Request) {
			h.CreateEvento(rec, req)
		}},
		{name: "ficha", method: http.MethodGet, path: "/api/v1/admin/eventos/" + id, call: func(h *Handler, rec *httptest.ResponseRecorder, req *http.Request) {
			h.GetEvento(rec, req)
		}},
		{name: "edición", method: http.MethodPatch, path: "/api/v1/admin/eventos/" + id, call: func(h *Handler, rec *httptest.ResponseRecorder, req *http.Request) {
			h.UpdateEvento(rec, req)
		}},
		{name: "borrado", method: http.MethodDelete, path: "/api/v1/admin/eventos/" + id, call: func(h *Handler, rec *httptest.ResponseRecorder, req *http.Request) {
			h.DeleteEvento(rec, req)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := withPathValue(adminRequest(tt.method, tt.path, "{}", false), id)
			rec := httptest.NewRecorder()
			tt.call(h, rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, se esperaba 401 (%s)", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), `"code":"unauthenticated"`) {
				t.Errorf("sobre inesperado: %s", rec.Body)
			}
		})
	}
}

// TestListEventosHandler fija el listado del panel: paginación, filtro por
// estado y el DTO con `nextDate` siempre presente (null sin fecha próxima) e
// `isPast` (FR-006/FR-010).
func TestListEventosHandler(t *testing.T) {
	repo := newFakeRepository()
	h := newAdminHandler(t, repo)
	ctx := contextWithIdentity()

	// Un evento publicado con fecha próxima y otro en borrador solo con fechas
	// pasadas.
	upcoming := samplePublicEvent("Con próxima", "2026-10-12")
	past := samplePublicEvent("Solo pasada", "2026-10-01")
	past.PublicationState = StateDraft

	if _, err := h.service.CreateEvento(ctx, adminIdentityForTest().UserID, EventoInput{
		TitleEs:          upcoming.TitleEs,
		DetailEs:         "D",
		PlaceEs:          "P",
		PublicationState: "published",
		Dates:            []FechaInput{{Date: "2026-10-12"}},
	}); err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}
	if _, err := h.service.CreateEvento(ctx, adminIdentityForTest().UserID, EventoInput{
		TitleEs:  past.TitleEs,
		DetailEs: "D",
		PlaceEs:  "P",
		Dates:    []FechaInput{{Date: "2026-10-01"}},
	}); err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}

	t.Run("sin filtros", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ListEventos(rec, adminRequest(http.MethodGet, "/api/v1/admin/eventos", "", true))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
		var got EventoAdminList
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo no es EventoAdminList: %v (%s)", err, rec.Body)
		}
		if got.Total != 2 || len(got.Items) != 2 || got.Limit != 20 {
			t.Fatalf("listado = %+v", got)
		}
		// `nextDate` SIEMPRE presente: null en el evento sin fecha próxima.
		if !strings.Contains(rec.Body.String(), `"nextDate":null`) {
			t.Errorf("falta nextDate:null en el listado: %s", rec.Body)
		}
		if !strings.Contains(rec.Body.String(), `"isPast":true`) {
			t.Errorf("falta isPast en el listado: %s", rec.Body)
		}
	})

	t.Run("con paginación", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ListEventos(rec, adminRequest(http.MethodGet, "/api/v1/admin/eventos?limit=1&offset=1", "", true))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
		var got EventoAdminList
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo inesperado: %v", err)
		}
		if len(got.Items) != 1 || got.Total != 2 || got.Limit != 1 || got.Offset != 1 {
			t.Errorf("listado = %+v", got)
		}
	})

	t.Run("con filtro de estado", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ListEventos(rec, adminRequest(http.MethodGet, "/api/v1/admin/eventos?publicationState=draft", "", true))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
		var got EventoAdminList
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo inesperado: %v", err)
		}
		if got.Total != 1 || got.Items[0].TitleEs != "Solo pasada" {
			t.Errorf("listado = %+v", got)
		}
	})

	t.Run("filtro inválido", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ListEventos(rec, adminRequest(http.MethodGet, "/api/v1/admin/eventos?publicationState=archived", "", true))
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "publicationState") {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("paginación inválida", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ListEventos(rec, adminRequest(http.MethodGet, "/api/v1/admin/eventos?limit=cero", "", true))
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "limit") {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})
}

// TestGetEventoHandler fija la ficha del panel, el 400 del id no UUID y el 404.
func TestGetEventoHandler(t *testing.T) {
	repo := newFakeRepository()
	h := newAdminHandler(t, repo)
	ctx := contextWithIdentity()

	created, err := h.service.CreateEvento(ctx, adminIdentityForTest().UserID, EventoInput{
		TitleEs:  "Noche de aposentos",
		DetailEs: "D",
		PlaceEs:  "P",
		Dates:    []FechaInput{{Date: "2026-10-12"}},
	})
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}

	t.Run("ficha existente", func(t *testing.T) {
		req := withPathValue(adminRequest(http.MethodGet, "/api/v1/admin/eventos/"+created.ID, "", true), created.ID)
		rec := httptest.NewRecorder()
		h.GetEvento(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
		var got EventoAdmin
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo inesperado: %v", err)
		}
		if got.ID != created.ID || got.TitleEs != "Noche de aposentos" {
			t.Errorf("ficha = %+v", got)
		}
	})

	t.Run("id no uuid", func(t *testing.T) {
		req := withPathValue(adminRequest(http.MethodGet, "/api/v1/admin/eventos/nope", "", true), "nope")
		rec := httptest.NewRecorder()
		h.GetEvento(rec, req)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"id"`) {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("inexistente", func(t *testing.T) {
		missing := uuid.New().String()
		req := withPathValue(adminRequest(http.MethodGet, "/api/v1/admin/eventos/"+missing, "", true), missing)
		rec := httptest.NewRecorder()
		h.GetEvento(rec, req)
		if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})
}

// TestUpdateEventoHandler fija el 200 de la edición, el 400 de un PATCH vacío o
// inválido y el 404 del id inexistente.
func TestUpdateEventoHandler(t *testing.T) {
	repo := newFakeRepository()
	h := newAdminHandler(t, repo)
	ctx := contextWithIdentity()

	created, err := h.service.CreateEvento(ctx, adminIdentityForTest().UserID, EventoInput{
		TitleEs:  "Noche de aposentos",
		DetailEs: "D",
		PlaceEs:  "P",
		Dates:    []FechaInput{{Date: "2026-10-12"}},
	})
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}

	t.Run("edición correcta", func(t *testing.T) {
		body := `{"titleEs":"Noche de vigilia","publicationState":"published"}`
		req := withPathValue(adminRequest(http.MethodPatch, "/api/v1/admin/eventos/"+created.ID, body, true), created.ID)
		rec := httptest.NewRecorder()
		h.UpdateEvento(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
		var got EventoAdmin
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo inesperado: %v", err)
		}
		if got.TitleEs != "Noche de vigilia" || got.PublicationState != string(StatePublished) {
			t.Errorf("evento = %+v", got)
		}
	})

	t.Run("PATCH vacío", func(t *testing.T) {
		req := withPathValue(adminRequest(http.MethodPatch, "/api/v1/admin/eventos/"+created.ID, `{}`, true), created.ID)
		rec := httptest.NewRecorder()
		h.UpdateEvento(rec, req)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"code":"invalid"`) {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("datos inválidos", func(t *testing.T) {
		req := withPathValue(adminRequest(http.MethodPatch, "/api/v1/admin/eventos/"+created.ID, `{"titleEs":"  "}`, true), created.ID)
		rec := httptest.NewRecorder()
		h.UpdateEvento(rec, req)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "titleEs") {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("inexistente", func(t *testing.T) {
		missing := uuid.New().String()
		req := withPathValue(adminRequest(http.MethodPatch, "/api/v1/admin/eventos/"+missing, `{"titleEs":"X"}`, true), missing)
		rec := httptest.NewRecorder()
		h.UpdateEvento(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})
}

// TestDeleteEventoHandler fija el 204 sin cuerpo y el 404 del id inexistente.
func TestDeleteEventoHandler(t *testing.T) {
	repo := newFakeRepository()
	h := newAdminHandler(t, repo)
	ctx := contextWithIdentity()

	created, err := h.service.CreateEvento(ctx, adminIdentityForTest().UserID, EventoInput{
		TitleEs:  "Noche de aposentos",
		DetailEs: "D",
		PlaceEs:  "P",
		Dates:    []FechaInput{{Date: "2026-10-12"}},
	})
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}

	req := withPathValue(adminRequest(http.MethodDelete, "/api/v1/admin/eventos/"+created.ID, "", true), created.ID)
	rec := httptest.NewRecorder()
	h.DeleteEvento(rec, req)
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
		t.Fatalf("status = %d, cuerpo = %q", rec.Code, rec.Body)
	}
	if len(repo.events) != 0 {
		t.Errorf("el evento no se borró: %+v", repo.events)
	}

	missing := uuid.New().String()
	req = withPathValue(adminRequest(http.MethodDelete, "/api/v1/admin/eventos/"+missing, "", true), missing)
	rec = httptest.NewRecorder()
	h.DeleteEvento(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
	}
}

// contextWithIdentity devuelve un contexto con la identidad del panel, para
// llamar al service directamente desde las pruebas.
func contextWithIdentity() context.Context {
	req := withIdentity(httptest.NewRequest(http.MethodGet, "/", nil), adminIdentityForTest())
	return req.Context()
}

// --- Actividades (T425) ---

// TestCreateActividadHandler fija el 201 con el DTO del panel, el 400 con
// `details` por campo (sin guardar nada) y el 409 del duplicado del módulo.
func TestCreateActividadHandler(t *testing.T) {
	t.Run("alta correcta", func(t *testing.T) {
		repo := newFakeRepository()
		h := newAdminHandler(t, repo)

		body := `{"titleEs":"  Repartir comida  ","detailEs":"Detalle","placeEs":"Salón",
			"date":"2026-10-12","startTime":"10:00","endTime":"13:00","publicationState":"published"}`
		rec := httptest.NewRecorder()
		h.CreateActividad(rec, adminRequest(http.MethodPost, "/api/v1/admin/actividades", body, true))

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, se esperaba 201 (%s)", rec.Code, rec.Body)
		}
		var got ActividadAdmin
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo no es ActividadAdmin: %v (%s)", err, rec.Body)
		}
		if got.TitleEs != "Repartir comida" || got.Date != "2026-10-12" || got.IsPast {
			t.Errorf("actividad = %+v", got)
		}
		if got.StartTime == nil || *got.StartTime != "10:00" {
			t.Errorf("startTime = %v", got.StartTime)
		}
		if len(repo.activities) != 1 {
			t.Fatalf("el service no guardó la actividad: %+v", repo.activities)
		}
	})

	tests := []struct {
		name   string
		body   string
		detail string
	}{
		{name: "título vacío", body: `{"titleEs":"","detailEs":"D","placeEs":"P","date":"2026-10-12"}`, detail: "titleEs"},
		{name: "sin fecha", body: `{"titleEs":"T","detailEs":"D","placeEs":"P","date":""}`, detail: "date"},
		{name: "fecha inexistente", body: `{"titleEs":"T","detailEs":"D","placeEs":"P","date":"2026-13-32"}`, detail: "date"},
		{name: "fin sin inicio", body: `{"titleEs":"T","detailEs":"D","placeEs":"P","date":"2026-10-12","endTime":"21:00"}`, detail: "endTime"},
		{name: "campo desconocido", body: `{"titleEs":"T","otro":1}`, detail: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			h := newAdminHandler(t, repo)

			rec := httptest.NewRecorder()
			h.CreateActividad(rec, adminRequest(http.MethodPost, "/api/v1/admin/actividades", tt.body, true))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
			}
			if tt.detail != "" && !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
			if len(repo.activities) != 0 {
				t.Errorf("un DTO inválido no debe guardar nada: %+v", repo.activities)
			}
		})
	}

	t.Run("duplicado", func(t *testing.T) {
		repo := newFakeRepository()
		h := newAdminHandler(t, repo)
		body := `{"titleEs":"Repartir comida","detailEs":"D","placeEs":"Salón","date":"2026-10-12"}`

		rec := httptest.NewRecorder()
		h.CreateActividad(rec, adminRequest(http.MethodPost, "/api/v1/admin/actividades", body, true))
		if rec.Code != http.StatusCreated {
			t.Fatalf("primera alta: status = %d (%s)", rec.Code, rec.Body)
		}
		rec = httptest.NewRecorder()
		h.CreateActividad(rec, adminRequest(http.MethodPost, "/api/v1/admin/actividades", body, true))
		if rec.Code != http.StatusConflict {
			t.Fatalf("duplicado: status = %d, se esperaba 409 (%s)", rec.Code, rec.Body)
		}
	})
}

// TestActividadHandlersRequireSession fija el 401 de red de seguridad de las
// cinco operaciones del panel de actividades.
func TestActividadHandlersRequireSession(t *testing.T) {
	h := newAdminHandler(t, newFakeRepository())
	id := uuid.New().String()

	tests := []struct {
		name   string
		method string
		path   string
		call   func(h *Handler, rec *httptest.ResponseRecorder, req *http.Request)
	}{
		{name: "listado", method: http.MethodGet, path: "/api/v1/admin/actividades", call: func(h *Handler, rec *httptest.ResponseRecorder, req *http.Request) {
			h.ListActividades(rec, req)
		}},
		{name: "alta", method: http.MethodPost, path: "/api/v1/admin/actividades", call: func(h *Handler, rec *httptest.ResponseRecorder, req *http.Request) {
			h.CreateActividad(rec, req)
		}},
		{name: "ficha", method: http.MethodGet, path: "/api/v1/admin/actividades/" + id, call: func(h *Handler, rec *httptest.ResponseRecorder, req *http.Request) {
			h.GetActividad(rec, req)
		}},
		{name: "edición", method: http.MethodPatch, path: "/api/v1/admin/actividades/" + id, call: func(h *Handler, rec *httptest.ResponseRecorder, req *http.Request) {
			h.UpdateActividad(rec, req)
		}},
		{name: "borrado", method: http.MethodDelete, path: "/api/v1/admin/actividades/" + id, call: func(h *Handler, rec *httptest.ResponseRecorder, req *http.Request) {
			h.DeleteActividad(rec, req)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := withPathValue(adminRequest(tt.method, tt.path, "{}", false), id)
			rec := httptest.NewRecorder()
			tt.call(h, rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, se esperaba 401 (%s)", rec.Code, rec.Body)
			}
		})
	}
}

// TestListActividadesHandler fija el listado del panel de actividades con su
// paginación, su filtro y el `isPast` de cada fila.
func TestListActividadesHandler(t *testing.T) {
	repo := newFakeRepository()
	h := newAdminHandler(t, repo)
	ctx := contextWithIdentity()
	actor := adminIdentityForTest().UserID

	if _, err := h.service.CreateActividad(ctx, actor, ActividadInput{
		TitleEs: "Futura", DetailEs: "D", PlaceEs: "P", Date: "2026-10-12", PublicationState: "published",
	}); err != nil {
		t.Fatalf("CreateActividad: %v", err)
	}
	if _, err := h.service.CreateActividad(ctx, actor, ActividadInput{
		TitleEs: "Vencida", DetailEs: "D", PlaceEs: "P", Date: "2026-10-01",
	}); err != nil {
		t.Fatalf("CreateActividad: %v", err)
	}

	t.Run("sin filtros", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ListActividades(rec, adminRequest(http.MethodGet, "/api/v1/admin/actividades", "", true))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
		var got ActividadAdminList
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo no es ActividadAdminList: %v (%s)", err, rec.Body)
		}
		if got.Total != 2 || len(got.Items) != 2 {
			t.Fatalf("listado = %+v", got)
		}
		if !strings.Contains(rec.Body.String(), `"isPast":true`) {
			t.Errorf("falta isPast en el listado: %s", rec.Body)
		}
	})

	t.Run("con filtro y paginación", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ListActividades(rec, adminRequest(http.MethodGet, "/api/v1/admin/actividades?publicationState=published&limit=1", "", true))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
		var got ActividadAdminList
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo inesperado: %v", err)
		}
		if got.Total != 1 || len(got.Items) != 1 || got.Items[0].TitleEs != "Futura" || got.Limit != 1 {
			t.Errorf("listado = %+v", got)
		}
	})

	t.Run("filtro inválido", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ListActividades(rec, adminRequest(http.MethodGet, "/api/v1/admin/actividades?publicationState=archived", "", true))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})
}

// TestGetUpdateDeleteActividadHandler cubre la ficha, la edición y el borrado
// con sus códigos (200/400/404/204).
func TestGetUpdateDeleteActividadHandler(t *testing.T) {
	repo := newFakeRepository()
	h := newAdminHandler(t, repo)
	ctx := contextWithIdentity()

	created, err := h.service.CreateActividad(ctx, adminIdentityForTest().UserID, ActividadInput{
		TitleEs: "Repartir comida", DetailEs: "D", PlaceEs: "P", Date: "2026-10-12",
	})
	if err != nil {
		t.Fatalf("CreateActividad: %v", err)
	}

	t.Run("ficha", func(t *testing.T) {
		req := withPathValue(adminRequest(http.MethodGet, "/api/v1/admin/actividades/"+created.ID, "", true), created.ID)
		rec := httptest.NewRecorder()
		h.GetActividad(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
		var got ActividadAdmin
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo inesperado: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("ficha = %+v", got)
		}
	})

	t.Run("ficha con id no uuid", func(t *testing.T) {
		req := withPathValue(adminRequest(http.MethodGet, "/api/v1/admin/actividades/nope", "", true), "nope")
		rec := httptest.NewRecorder()
		h.GetActividad(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("ficha inexistente", func(t *testing.T) {
		missing := uuid.New().String()
		req := withPathValue(adminRequest(http.MethodGet, "/api/v1/admin/actividades/"+missing, "", true), missing)
		rec := httptest.NewRecorder()
		h.GetActividad(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("edición", func(t *testing.T) {
		req := withPathValue(adminRequest(http.MethodPatch, "/api/v1/admin/actividades/"+created.ID,
			`{"titleEs":"Repartir comida (editada)","publicationState":"published"}`, true), created.ID)
		rec := httptest.NewRecorder()
		h.UpdateActividad(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
		var got ActividadAdmin
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("cuerpo inesperado: %v", err)
		}
		if got.TitleEs != "Repartir comida (editada)" || got.PublicationState != string(StatePublished) {
			t.Errorf("actividad = %+v", got)
		}
	})

	t.Run("edición con PATCH vacío", func(t *testing.T) {
		req := withPathValue(adminRequest(http.MethodPatch, "/api/v1/admin/actividades/"+created.ID, `{}`, true), created.ID)
		rec := httptest.NewRecorder()
		h.UpdateActividad(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("edición inexistente", func(t *testing.T) {
		missing := uuid.New().String()
		req := withPathValue(adminRequest(http.MethodPatch, "/api/v1/admin/actividades/"+missing, `{"titleEs":"X"}`, true), missing)
		rec := httptest.NewRecorder()
		h.UpdateActividad(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})

	t.Run("borrado", func(t *testing.T) {
		req := withPathValue(adminRequest(http.MethodDelete, "/api/v1/admin/actividades/"+created.ID, "", true), created.ID)
		rec := httptest.NewRecorder()
		h.DeleteActividad(rec, req)
		if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
			t.Fatalf("status = %d, cuerpo = %q", rec.Code, rec.Body)
		}
		if len(repo.activities) != 0 {
			t.Errorf("la actividad no se borró: %+v", repo.activities)
		}
	})

	t.Run("borrado inexistente", func(t *testing.T) {
		missing := uuid.New().String()
		req := withPathValue(adminRequest(http.MethodDelete, "/api/v1/admin/actividades/"+missing, "", true), missing)
		rec := httptest.NewRecorder()
		h.DeleteActividad(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
		}
	})
}
