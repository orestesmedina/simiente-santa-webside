package medios

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/testutil"
)

// Tests de los handlers del panel de medios (T925): el handler real sobre el
// service real, con el repositorio falso en memoria (sin PostgreSQL). Comprueban
// los códigos y sobres de cada operación (200/400/401/404/500), la validación con
// `details` por campo, el rechazo de claves desconocidas del `PATCH`
// (`additionalProperties: false`), la semántica del `PATCH` parcial (un solo
// campo no toca los demás) y la regla de auditoría de FR-022.

// adminIdentityForTest es la identidad de las pruebas del panel.
func adminIdentityForTest() session.Identity {
	return session.Identity{UserID: uuid.New(), Email: "ana@ejemplo.com", Permissions: []string{"medios"}}
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

// newTestHandlerCon construye el handler sobre el service real con las
// dependencias indicadas (el repositorio falso y el logger siempre van).
func newTestHandlerCon(t *testing.T, repo Repository, deps ServiceDeps) *Handler {
	t.Helper()
	logger, _ := testutil.NewLogger()
	deps.Repository = repo
	deps.Logger = logger
	return NewHandler(HandlerDeps{Service: NewService(deps), Logger: logger})
}

// --- Listado del panel (T925, FR-018/FR-019) ---

// TestListMediosHandlerOK fija el 200 con el sobre paginado (§8.1.2/§8.1.3): el
// panel ve **todo** (ocultas incluidas) menos las absorbidas, con la procedencia
// de cada fila y el `lastSync` cuando ya hubo alguna búsqueda.
func TestListMediosHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(episodioPublicable("Publicada"))
	oculta := episodioPublicable("Oculta por el equipo")
	oculta.Visibilidad = VisibilidadOculta
	repo.seed(oculta)
	repo.syncRuns = append(repo.syncRuns, SyncRun{
		IniciadoEn:    time.Date(2026, time.October, 17, 10, 0, 0, 0, time.UTC),
		TerminadoEn:   time.Date(2026, time.October, 17, 10, 0, 30, 0, time.UTC),
		Importados:    9,
		EstadoYouTube: FeedOK,
		EstadoPodcast: FeedFallido,
	})
	h := newTestHandler(t, repo)

	rec := httptest.NewRecorder()
	h.ListMedios(rec, adminRequest(http.MethodGet, "/api/v1/admin/medios", "", true))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got MedioAdminList
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es MedioAdminList: %v (%s)", err, rec.Body)
	}
	if got.Total != 2 || len(got.Items) != 2 {
		t.Fatalf("sobre = %+v, se esperaba el módulo completo (ocultas incluidas)", got)
	}
	if got.Limit != 20 || got.Offset != 0 {
		t.Errorf("tanda = {%d, %d}, se esperaba {20, 0}", got.Limit, got.Offset)
	}
	if got.LastSync == nil {
		t.Fatal("falta lastSync: el listado debe traer la última búsqueda")
	}
	if got.LastSync.Importados != 9 || got.LastSync.EstadoYouTube != string(FeedOK) ||
		got.LastSync.EstadoPodcast != string(FeedFallido) {
		t.Errorf("lastSync = %+v", got.LastSync)
	}
	// La procedencia viaja en cada fila (FR-019): es lo que el equipo revisa
	// antes de unir.
	if got.Items[0].Fuentes == nil {
		t.Errorf("la fila no trae sources: %+v", got.Items[0])
	}
	if got.Items[0].Visibilidad == "" {
		t.Errorf("la fila no trae visibility: %+v", got.Items[0])
	}
}

// TestListMediosHandlerSinBusquedas fija FR-018: sin ninguna búsqueda previa el
// sobre **no** trae `lastSync` (clave omitida, nunca `null`).
func TestListMediosHandlerSinBusquedas(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.ListMedios(rec, adminRequest(http.MethodGet, "/api/v1/admin/medios", "", true))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "lastSync") {
		t.Errorf("sin búsquedas no debe haber lastSync: %s", rec.Body)
	}
}

// TestListMediosHandlerFilters fija los tres filtros del contrato: `visibility`,
// `reviewMark` (la marca de esa clase **pendiente**) e `includeMerged` (por
// defecto las absorbidas no se listan).
func TestListMediosHandlerFilters(t *testing.T) {
	idPar := uuidDePrueba()
	par := episodioPublicable("Par dudoso")
	par.DuplicadoDeID = &idPar

	absorbida := episodioPublicable("Absorbida")
	resultante := uuidDePrueba()
	absorbida.UnidoAID = &resultante

	oculta := episodioPublicable("Oculta")
	oculta.Visibilidad = VisibilidadOculta

	fuera := episodioPublicable("Fuera del feed")
	marca := MarcaDisponibilidadFueraDelFeed
	fuera.MarcaDisponibilidad = &marca

	nuevoRepo := func() *fakeRepository {
		repo := newFakeRepository()
		repo.seed(episodioPublicable("Publicada"), par, absorbida, oculta, fuera)
		return repo
	}

	tests := []struct {
		name    string
		target  string
		want    []string
		ausente []string
	}{
		{
			name:    "por defecto sin absorbidas",
			target:  "/api/v1/admin/medios",
			want:    []string{"Publicada", "Par dudoso", "Oculta", "Fuera del feed"},
			ausente: []string{"Absorbida"},
		},
		{
			name:    "incluir absorbidas",
			target:  "/api/v1/admin/medios?includeMerged=true",
			want:    []string{"Absorbida"},
			ausente: nil,
		},
		{
			name:    "filtro por visibilidad",
			target:  "/api/v1/admin/medios?visibility=hidden",
			want:    []string{"Oculta"},
			ausente: []string{"Publicada", "Par dudoso"},
		},
		{
			name:    "filtro por par pendiente",
			target:  "/api/v1/admin/medios?reviewMark=possible_duplicate",
			want:    []string{"Par dudoso"},
			ausente: []string{"Publicada", "Oculta"},
		},
		{
			name:    "filtro por marca de disponibilidad pendiente",
			target:  "/api/v1/admin/medios?reviewMark=missing_from_feed",
			want:    []string{"Fuera del feed"},
			ausente: []string{"Publicada", "Par dudoso"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(t, nuevoRepo())

			rec := httptest.NewRecorder()
			h.ListMedios(rec, adminRequest(http.MethodGet, tt.target, "", true))
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
			}
			raw := rec.Body.String()
			for _, titulo := range tt.want {
				if !strings.Contains(raw, titulo) {
					t.Errorf("falta %q en el listado: %s", titulo, raw)
				}
			}
			for _, titulo := range tt.ausente {
				if strings.Contains(raw, titulo) {
					t.Errorf("%q no debía aparecer: %s", titulo, raw)
				}
			}
		})
	}
}

// TestListMediosHandlerInvalidParams fija el 400 con `details` por campo de la
// paginación y de los filtros: el catálogo de `visibility`/`reviewMark` es
// cerrado y `includeMerged` es booleano.
func TestListMediosHandlerInvalidParams(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	tests := []struct {
		target string
		detail string
	}{
		{target: "/api/v1/admin/medios?limit=0", detail: "limit"},
		{target: "/api/v1/admin/medios?limit=x", detail: "limit"},
		{target: "/api/v1/admin/medios?offset=-1", detail: "offset"},
		{target: "/api/v1/admin/medios?visibility=draft", detail: "visibility"},
		{target: "/api/v1/admin/medios?reviewMark=otro", detail: "reviewMark"},
		{target: "/api/v1/admin/medios?includeMerged=x", detail: "includeMerged"},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ListMedios(rec, adminRequest(http.MethodGet, tt.target, "", true))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
		})
	}
}

// TestListMediosHandlerRequiresSession fija el 401 de la red de seguridad del
// handler (la cadena authn responde antes en producción).
func TestListMediosHandlerRequiresSession(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.ListMedios(rec, adminRequest(http.MethodGet, "/api/v1/admin/medios", "", false))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"unauthenticated"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// --- Edición de la ficha (T925, FR-019/FR-020/FR-022) ---

// TestUpdateMedioHandlerDescripcion fija el 200 con la ficha del panel y la regla
// de auditoría de FR-022: editar la descripción deja **una** fila
// `media.description.update` y la descripción pasa a ser del equipo
// (`descriptionEdited`).
func TestUpdateMedioHandlerDescripcion(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(episodioPublicable("Potencial sin Resultados"))
	id := repo.episodios[0].ID
	h := newTestHandler(t, repo)

	body := `{"descriptionEs":"  Un sermón sobre la higuera.  "}`
	rec := httptest.NewRecorder()
	h.UpdateMedio(rec, withPathValue(
		adminRequest(http.MethodPatch, "/api/v1/admin/medios/"+id.String(), body, true), id.String()))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got MedioAdmin
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es MedioAdmin: %v (%s)", err, rec.Body)
	}
	if got.DescripcionEs == nil || *got.DescripcionEs != "Un sermón sobre la higuera." {
		t.Errorf("descriptionEs = %v, se esperaba el texto recortado", got.DescripcionEs)
	}
	if !got.DescripcionEditada {
		t.Error("la descripción guardada debe quedar marcada como del equipo")
	}
	if got.Titulo != "Potencial sin Resultados" || got.Visibilidad != string(VisibilidadVisible) {
		t.Errorf("el PATCH de un campo no debe tocar los demás: %+v", got)
	}
	if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaDescriptionUpdate {
		t.Fatalf("auditoría = %+v, se esperaba una fila media.description.update", repo.committed)
	}
	if repo.committed[0].TargetLabel != "Medios · Potencial sin Resultados" {
		t.Errorf("targetLabel = %q", repo.committed[0].TargetLabel)
	}
}

// TestUpdateMedioHandlerVaciaLaDescripcion fija la decisión del humano del
// 2026-10-18: `{"descriptionEs":""}` **vacía** la descripción desde el panel
// (200, ya no 400), la ficha queda sin descripción —la clave se omite en el DTO
// del panel (SC-010)—, pasa a ser del equipo (`descriptionEdited`) y la
// auditoría registra el cambio igual que cualquier otro ajuste (FR-022).
func TestUpdateMedioHandlerVaciaLaDescripcion(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "vacío", body: `{"descriptionEs":""}`},
		{name: "solo espacios", body: `{"descriptionEs":"   "}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.seed(episodioPublicable("Potencial sin Resultados"))
			id := repo.episodios[0].ID
			h := newTestHandler(t, repo)

			rec := httptest.NewRecorder()
			h.UpdateMedio(rec, withPathValue(
				adminRequest(http.MethodPatch, "/api/v1/admin/medios/"+id.String(), tt.body, true), id.String()))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
			}
			var got MedioAdmin
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("cuerpo no es MedioAdmin: %v (%s)", err, rec.Body)
			}
			if got.DescripcionEs != nil {
				t.Errorf("descriptionEs = %v, se esperaba la ficha sin descripción", *got.DescripcionEs)
			}
			if strings.Contains(rec.Body.String(), `"descriptionEs"`) {
				t.Errorf("la descripción vacía no debe viajar en la respuesta: %s", rec.Body)
			}
			if !got.DescripcionEditada {
				t.Error("la descripción vaciada debe quedar marcada como del equipo")
			}
			if got.Titulo != "Potencial sin Resultados" || got.Visibilidad != string(VisibilidadVisible) {
				t.Errorf("el PATCH de un campo no debe tocar los demás: %+v", got)
			}
			if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaDescriptionUpdate {
				t.Fatalf("auditoría = %+v, se esperaba una fila media.description.update", repo.committed)
			}
			if repo.committed[0].TargetLabel != "Medios · Potencial sin Resultados" {
				t.Errorf("targetLabel = %q", repo.committed[0].TargetLabel)
			}
		})
	}
}

// TestUpdateMedioHandlerVisibilidad fija el cambio de visibilidad: `media.hide`/
// `media.show` **solo** por el cambio de estado, y las dos cosas a la vez dejan
// **dos** filas (FR-022).
func TestUpdateMedioHandlerVisibilidad(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantCodes []string
	}{
		{name: "ocultar", body: `{"visibility":"hidden"}`, wantCodes: []string{audit.ActionMediaHide}},
		{name: "mostrar", body: `{"visibility":"visible"}`, wantCodes: []string{audit.ActionMediaShow}},
		{
			name:      "datos y visibilidad",
			body:      `{"descriptionEs":"Otro texto.","visibility":"hidden"}`,
			wantCodes: []string{audit.ActionMediaDescriptionUpdate, audit.ActionMediaHide},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			ep := episodioPublicable("Potencial sin Resultados")
			if tt.name == "mostrar" {
				ep.Visibilidad = VisibilidadOculta
			}
			repo.seed(ep)
			id := repo.episodios[0].ID
			h := newTestHandler(t, repo)

			rec := httptest.NewRecorder()
			h.UpdateMedio(rec, withPathValue(
				adminRequest(http.MethodPatch, "/api/v1/admin/medios/"+id.String(), tt.body, true), id.String()))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
			}
			var got MedioAdmin
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("cuerpo no es MedioAdmin: %v", err)
			}
			want := string(VisibilidadVisible)
			if tt.name == "ocultar" || tt.name == "datos y visibilidad" {
				want = string(VisibilidadOculta)
			}
			if got.Visibilidad != want {
				t.Errorf("visibility = %q, se esperaba %q", got.Visibilidad, want)
			}
			if len(repo.committed) != len(tt.wantCodes) {
				t.Fatalf("auditoría = %+v, se esperaban %d filas", repo.committed, len(tt.wantCodes))
			}
			for i, code := range tt.wantCodes {
				if repo.committed[i].Code != code {
					t.Errorf("fila %d = %q, se esperaba %q", i, repo.committed[i].Code, code)
				}
			}
		})
	}
}

// TestUpdateMedioHandlerInvalid fija el 400 con `details` por campo y que un DTO
// inválido **no guarda nada** (FR-023, sin guardado a medias).
func TestUpdateMedioHandlerInvalid(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		detail string
	}{
		{
			name:   "descripción demasiado larga",
			body:   `{"descriptionEs":"` + strings.Repeat("a", 301) + `"}`,
			detail: "descriptionEs",
		},
		{
			name:   "inglés sin español",
			body:   `{"descriptionEn":"Only English"}`,
			detail: "descriptionEs",
		},
		{name: "visibilidad fuera del catálogo", body: `{"visibility":"draft"}`, detail: "visibility"},
		{name: "clave desconocida", body: `{"title":"Otro nombre"}`, detail: "title"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			ep := episodioPublicable("Potencial sin Resultados")
			ep.DescripcionEs = nil // sin descripción previa: el inglés queda huérfano
			repo.seed(ep)
			id := repo.episodios[0].ID
			h := newTestHandler(t, repo)

			rec := httptest.NewRecorder()
			h.UpdateMedio(rec, withPathValue(
				adminRequest(http.MethodPatch, "/api/v1/admin/medios/"+id.String(), tt.body, true), id.String()))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), `"code":"invalid"`) {
				t.Errorf("sobre inesperado: %s", rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
			if len(repo.committed) != 0 {
				t.Errorf("un DTO inválido no debe guardar nada: %+v", repo.committed)
			}
			if repo.episodios[0].Titulo != "Potencial sin Resultados" {
				t.Errorf("la ficha cambió con un DTO inválido: %+v", repo.episodios[0])
			}
		})
	}
}

// TestUpdateMedioHandlerNotFound fija el 404 de una ficha inexistente.
func TestUpdateMedioHandlerNotFound(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())
	id := uuidDePrueba()

	rec := httptest.NewRecorder()
	h.UpdateMedio(rec, withPathValue(
		adminRequest(http.MethodPatch, "/api/v1/admin/medios/"+id.String(), `{"visibility":"hidden"}`, true), id.String()))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestUpdateMedioHandlerInvalidID fija el 400 con details.id de un id que no es
// UUID (el router solo deja pasar el segmento).
func TestUpdateMedioHandlerInvalidID(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.UpdateMedio(rec, withPathValue(
		adminRequest(http.MethodPatch, "/api/v1/admin/medios/no-es-uuid", `{"visibility":"hidden"}`, true), "no-es-uuid"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"id"`) {
		t.Errorf("sin details.id: %s", rec.Body)
	}
}

// TestUpdateMedioHandlerRequiresSession fija el 401 de la red de seguridad.
func TestUpdateMedioHandlerRequiresSession(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())
	id := uuidDePrueba()

	rec := httptest.NewRecorder()
	h.UpdateMedio(rec, withPathValue(
		adminRequest(http.MethodPatch, "/api/v1/admin/medios/"+id.String(), `{"visibility":"hidden"}`, false), id.String()))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401 (%s)", rec.Code, rec.Body)
	}
}

// TestUpdateMedioHandlerServiceError fija el sobre genérico: un fallo interno no
// filtra la causa.
func TestUpdateMedioHandlerServiceError(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(episodioPublicable("Potencial sin Resultados"))
	id := repo.episodios[0].ID
	repo.mutationErr = errors.New("boom")
	h := newTestHandler(t, repo)

	rec := httptest.NewRecorder()
	h.UpdateMedio(rec, withPathValue(
		adminRequest(http.MethodPatch, "/api/v1/admin/medios/"+id.String(), `{"visibility":"hidden"}`, true), id.String()))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("la respuesta filtra la causa interna: %s", rec.Body)
	}
}

// TestUpdateMedioHandlerWithoutService fija el 500 de composición.
func TestUpdateMedioHandlerWithoutService(t *testing.T) {
	h := NewHandler(HandlerDeps{})
	id := uuidDePrueba()

	rec := httptest.NewRecorder()
	h.UpdateMedio(rec, withPathValue(
		adminRequest(http.MethodPatch, "/api/v1/admin/medios/"+id.String(), `{"visibility":"hidden"}`, true), id.String()))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
}

// --- Buscar episodios nuevos (T926, FR-011/FR-016/FR-018) ---

// entradaDePrueba arma un elemento de feed válido para el doble de `FeedSource`.
func entradaFeedDePrueba(idExterno, titulo string) EntradaFeed {
	return EntradaFeed{
		IDExterno:   idExterno,
		Titulo:      titulo,
		Enlace:      "https://www.youtube.com/watch?v=" + idExterno,
		Publicado:   time.Date(2026, time.July, 3, 12, 0, 0, 0, time.UTC),
		Descripcion: "Un sermón sobre la higuera.",
	}
}

// TestBuscarNuevosHandlerOK fija el 200 con el sobre `ResumenBusqueda` completo
// (FR-018): los conteos, el estado de cada feed y la importación de lo que faltaba
// (FR-012).
func TestBuscarNuevosHandlerOK(t *testing.T) {
	repo := newFakeRepository()
	feeds := &fakeFeedSource{
		youtube: ResultadoFeed{Entradas: []EntradaFeed{entradaFeedDePrueba("abc12345678", "Potencial sin Resultados")}},
	}
	h := newTestHandlerCon(t, repo, ServiceDeps{FeedSource: feeds})

	rec := httptest.NewRecorder()
	h.BuscarNuevos(rec, adminRequest(http.MethodPost, "/api/v1/admin/medios/buscar-nuevos", "", true))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got ResumenBusqueda
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es ResumenBusqueda: %v (%s)", err, rec.Body)
	}
	if got.Importados != 1 {
		t.Errorf("importedCount = %d, se esperaba 1", got.Importados)
	}
	if got.EstadoYouTube != string(FeedOK) {
		t.Errorf("youtube = %q, se esperaba ok", got.EstadoYouTube)
	}
	if got.IniciadoEn.IsZero() || got.TerminadoEn.IsZero() {
		t.Errorf("el resumen debe traer sus fechas: %+v", got)
	}
	if len(repo.episodios) != 1 {
		t.Fatalf("el episodio no se importó: %+v", repo.episodios)
	}
	// Una búsqueda completa deja **una** fila de auditoría `media.sync` (P9-14).
	if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaSync {
		t.Errorf("auditoría = %+v, se esperaba una fila media.sync", repo.committed)
	}
}

// TestBuscarNuevosHandlerFeedCaido fija FR-016: un feed que no responde **no** es
// un error —el resumen lo informa y se importa lo del otro—, así que la respuesta
// sigue siendo 200 con el estado `failed`.
func TestBuscarNuevosHandlerFeedCaido(t *testing.T) {
	repo := newFakeRepository()
	feeds := &fakeFeedSource{
		errYoutube: errFeedNoDisponible,
		podcast:    ResultadoFeed{},
	}
	h := newTestHandlerCon(t, repo, ServiceDeps{FeedSource: feeds})

	rec := httptest.NewRecorder()
	h.BuscarNuevos(rec, adminRequest(http.MethodPost, "/api/v1/admin/medios/buscar-nuevos", "", true))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got ResumenBusqueda
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es ResumenBusqueda: %v (%s)", err, rec.Body)
	}
	if got.EstadoYouTube != string(FeedFallido) {
		t.Errorf("youtube = %q, se esperaba failed", got.EstadoYouTube)
	}
	if got.EstadoPodcast != string(FeedOK) {
		t.Errorf("podcast = %q, se esperaba ok", got.EstadoPodcast)
	}
	if got.Ocultos != 0 || got.Marcados != 0 {
		t.Errorf("un feed caído no oculta ni marca nada: %+v", got)
	}
}

// TestBuscarNuevosHandlerNeverLeaksNetworkText fija FR-018: el resumen nunca
// expone texto técnico de red (nombres de host internos, errores de Go, URLs),
// aunque los dos feeds fallen.
func TestBuscarNuevosHandlerNeverLeaksNetworkText(t *testing.T) {
	repo := newFakeRepository()
	feeds := &fakeFeedSource{
		errYoutube: errors.New("dial tcp 169.254.169.254:443: i/o timeout"),
		errPodcast: errFeedNoDisponible,
	}
	h := newTestHandlerCon(t, repo, ServiceDeps{FeedSource: feeds})

	rec := httptest.NewRecorder()
	h.BuscarNuevos(rec, adminRequest(http.MethodPost, "/api/v1/admin/medios/buscar-nuevos", "", true))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	raw := rec.Body.String()
	for _, secreto := range []string{
		"http", "169.254", "dial", "timeout", "youtube.com", "anchor.fm", "Medios", "error",
	} {
		if strings.Contains(raw, secreto) {
			t.Errorf("el resumen filtra %q: %s", secreto, raw)
		}
	}
}

// TestBuscarNuevosHandlerRequiresSession fija el 401 de la red de seguridad.
func TestBuscarNuevosHandlerRequiresSession(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())

	rec := httptest.NewRecorder()
	h.BuscarNuevos(rec, adminRequest(http.MethodPost, "/api/v1/admin/medios/buscar-nuevos", "", false))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401 (%s)", rec.Code, rec.Body)
	}
}

// TestBuscarNuevosHandlerServiceError fija el 500 genérico: un fallo interno (el
// registro de la búsqueda, la base de datos) no filtra la causa y lo ya importado
// queda intacto.
func TestBuscarNuevosHandlerServiceError(t *testing.T) {
	repo := newFakeRepository()
	repo.syncRunErr = errors.New("boom")
	feeds := &fakeFeedSource{}
	h := newTestHandlerCon(t, repo, ServiceDeps{FeedSource: feeds})

	rec := httptest.NewRecorder()
	h.BuscarNuevos(rec, adminRequest(http.MethodPost, "/api/v1/admin/medios/buscar-nuevos", "", true))

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

// TestBuscarNuevosHandlerWithoutService fija el 500 de composición.
func TestBuscarNuevosHandlerWithoutService(t *testing.T) {
	h := NewHandler(HandlerDeps{})

	rec := httptest.NewRecorder()
	h.BuscarNuevos(rec, adminRequest(http.MethodPost, "/api/v1/admin/medios/buscar-nuevos", "", true))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
	}
}

// --- Unión de duplicados (T927, FR-014/FR-019) ---

// parUnible siembra dos entradas del mismo sermón (video y podcast) con su
// procedencia y devuelve el repositorio con sus dos ids.
func parUnible() (*fakeRepository, uuid.UUID, uuid.UUID) {
	video := episodioPublicable("Potencial sin Resultados")
	audio := episodioPublicable("Potencial sin Resultados: La Lección de la Higuera")

	repo := newFakeRepository()
	repo.seed(video, audio)
	repo.seedFuentes(
		FuenteEpisodio{
			EpisodioID: repo.episodios[0].ID,
			Proveedor:  ProveedorYouTube,
			IDExterno:  "abc12345678",
			TituloFeed: "Potencial sin Resultados",
			Enlace:     "https://www.youtube.com/watch?v=abc12345678",
		},
		FuenteEpisodio{
			EpisodioID: repo.episodios[1].ID,
			Proveedor:  ProveedorPodcast,
			IDExterno:  "episodio-1",
			TituloFeed: "Potencial sin Resultados: La Lección de la Higuera",
			Enlace:     "https://podcasters.spotify.com/pod/show/iglesia/episodes/1",
		},
	)
	return repo, repo.episodios[0].ID, repo.episodios[1].ID
}

// TestUnirMediosHandlerOK fija la unión por puntero (FR-014/FR-019): una sola
// ficha con los dos enlaces y las dos procedencias, la absorbida oculta y
// apuntando a la resultante, cero borrados y **una** fila `media.merge`.
func TestUnirMediosHandlerOK(t *testing.T) {
	repo, id, otherID := parUnible()
	h := newTestHandler(t, repo)

	body := `{"otherId":"` + otherID.String() + `"}`
	rec := httptest.NewRecorder()
	h.UnirMedios(rec, withPathValue(
		adminRequest(http.MethodPost, "/api/v1/admin/medios/"+id.String()+"/unir", body, true), id.String()))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got MedioAdmin
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es MedioAdmin: %v (%s)", err, rec.Body)
	}
	if got.Origen != string(OrigenAmbos) || got.VideoURL == "" || got.AudioURL == "" {
		t.Errorf("la ficha resultante debe traer los dos enlaces y origin=both: %+v", got)
	}
	if len(got.Fuentes) != 2 {
		t.Errorf("sources = %+v, se esperaban las dos procedencias", got.Fuentes)
	}
	if got.Visibilidad != string(VisibilidadVisible) {
		t.Errorf("la resultante conserva la visibilidad de la base: %q", got.Visibilidad)
	}
	if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaMerge {
		t.Fatalf("auditoría = %+v, se esperaba una fila media.merge", repo.committed)
	}
	// La absorbida queda con el puntero y oculta: fuera de todas las consultas.
	if _, ok := repo.buscar(otherID); !ok {
		t.Fatal("la absorbida desapareció (cero DELETE, FR-019)")
	}
}

// TestUnirMediosHandlerInvalid fija los 400 de la unión: `otherId` ausente o mal
// formado (details.otherId), unir una ficha consigo misma (la regla del contrato,
// que valida el service) y una clave desconocida del cuerpo.
func TestUnirMediosHandlerInvalid(t *testing.T) {
	repo, id, otherID := parUnible()
	h := newTestHandler(t, repo)

	tests := []struct {
		name   string
		body   string
		detail string
	}{
		{name: "sin otherId", body: `{}`, detail: "otherId"},
		{name: "otherId no es uuid", body: `{"otherId":"no-es-uuid"}`, detail: "otherId"},
		{name: "consigo misma", body: `{"otherId":"` + id.String() + `"}`, detail: "otherId"},
		{name: "clave desconocida", body: `{"otherId":"` + otherID.String() + `","force":true}`, detail: "force"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.UnirMedios(rec, withPathValue(
				adminRequest(http.MethodPost, "/api/v1/admin/medios/"+id.String()+"/unir", tt.body, true), id.String()))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
			if len(repo.repoints) != 0 {
				t.Errorf("una unión inválida no debe aplicarse: %+v", repo.repoints)
			}
		})
	}
}

// TestUnirMediosHandlerNotFound fija el 404 de un `otherId` inexistente o ya
// absorbido (el service lo distingue del 400 de un id mal formado).
func TestUnirMediosHandlerNotFound(t *testing.T) {
	repo, id, _ := parUnible()
	h := newTestHandler(t, repo)

	rec := httptest.NewRecorder()
	h.UnirMedios(rec, withPathValue(
		adminRequest(http.MethodPost, "/api/v1/admin/medios/"+id.String()+"/unir",
			`{"otherId":"`+uuidDePrueba().String()+`"}`, true), id.String()))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Errorf("sobre inesperado: %s", rec.Body)
	}
}

// TestUnirMediosHandlerRequiresSession fija el 401 de la red de seguridad.
func TestUnirMediosHandlerRequiresSession(t *testing.T) {
	repo, id, otherID := parUnible()
	h := newTestHandler(t, repo)

	rec := httptest.NewRecorder()
	h.UnirMedios(rec, withPathValue(
		adminRequest(http.MethodPost, "/api/v1/admin/medios/"+id.String()+"/unir",
			`{"otherId":"`+otherID.String()+`"}`, false), id.String()))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401 (%s)", rec.Code, rec.Body)
	}
}

// --- Resolución de marcas (T927, FR-017) ---

// TestResolverMarcaHandlerOK fija la resolución **por eje** (P9-19): resolver el
// par de duplicados deja las dos fichas visibles, con la marca cerrada y su
// registro `media.mark.resolve`.
func TestResolverMarcaHandlerOK(t *testing.T) {
	idA, idB := uuidDePrueba(), uuidDePrueba()
	primera := episodioPublicable("Par dudoso")
	primera.ID = idA
	primera.DuplicadoDeID = &idB
	segunda := episodioPublicable("Par dudoso: el otro")
	segunda.ID = idB
	segunda.DuplicadoDeID = &idA

	repo := newFakeRepository()
	repo.seed(primera, segunda)
	id := idA
	h := newTestHandler(t, repo)

	body := `{"mark":"possible_duplicate","visibility":"visible"}`
	rec := httptest.NewRecorder()
	h.ResolverMarca(rec, withPathValue(
		adminRequest(http.MethodPost, "/api/v1/admin/medios/"+id.String()+"/marcas/resolver", body, true), id.String()))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got MedioAdmin
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es MedioAdmin: %v (%s)", err, rec.Body)
	}
	if got.Visibilidad != string(VisibilidadVisible) {
		t.Errorf("visibility = %q, se esperaba visible", got.Visibilidad)
	}
	if got.MarcaDuplicadoResueltaEn == nil {
		t.Error("la marca debe quedar resuelta con su fecha")
	}
	if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaMarkResolve {
		t.Fatalf("auditoría = %+v, se esperaba una fila media.mark.resolve", repo.committed)
	}
}

// TestResolverMarcaHandlerPorEje fija que los dos ejes son independientes
// (P9-19): resolver la marca de disponibilidad de una ficha con las dos marcas
// activas **no** toca el par de duplicados.
func TestResolverMarcaHandlerPorEje(t *testing.T) {
	marca := MarcaDisponibilidadFueraDelFeed
	ep := episodioPublicable("Con dos marcas")
	idPareja := uuidDePrueba()
	ep.DuplicadoDeID = &idPareja
	ep.MarcaDisponibilidad = &marca

	repo := newFakeRepository()
	repo.seed(ep)
	id := repo.episodios[0].ID
	h := newTestHandler(t, repo)

	body := `{"mark":"missing_from_feed","visibility":"visible"}`
	rec := httptest.NewRecorder()
	h.ResolverMarca(rec, withPathValue(
		adminRequest(http.MethodPost, "/api/v1/admin/medios/"+id.String()+"/marcas/resolver", body, true), id.String()))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200 (%s)", rec.Code, rec.Body)
	}
	var got MedioAdmin
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("cuerpo no es MedioAdmin: %v", err)
	}
	if got.MarcaDisponibilidadResueltaEn == nil {
		t.Error("la marca de disponibilidad debe quedar resuelta")
	}
	if got.MarcaDuplicadoResueltaEn != nil {
		t.Error("resolver un eje no debe resolver el otro (P9-19)")
	}
	if got.DuplicadoDeID == "" {
		t.Error("el par de duplicados sigue activo")
	}
}

// TestResolverMarcaHandlerInvalid fija los 400 del catálogo cerrado: `mark` y
// `visibility` solo admiten sus valores.
func TestResolverMarcaHandlerInvalid(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(episodioPublicable("Potencial sin Resultados"))
	id := repo.episodios[0].ID
	h := newTestHandler(t, repo)

	tests := []struct {
		name   string
		body   string
		detail string
	}{
		{name: "mark fuera del catálogo", body: `{"mark":"otra","visibility":"visible"}`, detail: "mark"},
		{name: "visibility fuera del catálogo", body: `{"mark":"unavailable","visibility":"draft"}`, detail: "visibility"},
		{name: "sin mark", body: `{"visibility":"visible"}`, detail: "mark"},
		{name: "sin visibility", body: `{"mark":"unavailable"}`, detail: "visibility"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ResolverMarca(rec, withPathValue(
				adminRequest(http.MethodPost, "/api/v1/admin/medios/"+id.String()+"/marcas/resolver", tt.body, true), id.String()))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba 400 (%s)", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tt.detail) {
				t.Errorf("sin details.%s: %s", tt.detail, rec.Body)
			}
		})
	}
}

// TestResolverMarcaHandlerConflict fija el 409 de una ficha **sin** la marca
// indicada: no hay nada que resolver en ese eje.
func TestResolverMarcaHandlerConflict(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(episodioPublicable("Sin marcas"))
	id := repo.episodios[0].ID
	h := newTestHandler(t, repo)

	for _, mark := range []string{"possible_duplicate", "unavailable", "missing_from_feed"} {
		t.Run(mark, func(t *testing.T) {
			body := `{"mark":"` + mark + `","visibility":"visible"}`
			rec := httptest.NewRecorder()
			h.ResolverMarca(rec, withPathValue(
				adminRequest(http.MethodPost, "/api/v1/admin/medios/"+id.String()+"/marcas/resolver", body, true), id.String()))

			if rec.Code != http.StatusConflict {
				t.Fatalf("status = %d, se esperaba 409 (%s)", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), `"code":"conflict"`) {
				t.Errorf("sobre inesperado: %s", rec.Body)
			}
		})
	}
}

// TestResolverMarcaHandlerNotFound fija el 404 de una ficha inexistente.
func TestResolverMarcaHandlerNotFound(t *testing.T) {
	h := newTestHandler(t, newFakeRepository())
	id := uuidDePrueba()

	rec := httptest.NewRecorder()
	h.ResolverMarca(rec, withPathValue(
		adminRequest(http.MethodPost, "/api/v1/admin/medios/"+id.String()+"/marcas/resolver",
			`{"mark":"unavailable","visibility":"visible"}`, true), id.String()))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404 (%s)", rec.Code, rec.Body)
	}
}

// TestResolverMarcaHandlerRequiresSession fija el 401 de la red de seguridad.
func TestResolverMarcaHandlerRequiresSession(t *testing.T) {
	repo := newFakeRepository()
	repo.seed(episodioPublicable("Potencial sin Resultados"))
	id := repo.episodios[0].ID
	h := newTestHandler(t, repo)

	rec := httptest.NewRecorder()
	h.ResolverMarca(rec, withPathValue(
		adminRequest(http.MethodPost, "/api/v1/admin/medios/"+id.String()+"/marcas/resolver",
			`{"mark":"unavailable","visibility":"visible"}`, false), id.String()))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401 (%s)", rec.Code, rec.Body)
	}
}

// TestUnirResolverHandlerServiceError fija el sobre genérico de las dos
// operaciones: un fallo interno no filtra la causa.
func TestUnirResolverHandlerServiceError(t *testing.T) {
	t.Run("unir", func(t *testing.T) {
		repo, id, otherID := parUnible()
		repo.repointErr = errors.New("boom")
		h := newTestHandler(t, repo)

		rec := httptest.NewRecorder()
		h.UnirMedios(rec, withPathValue(
			adminRequest(http.MethodPost, "/api/v1/admin/medios/"+id.String()+"/unir",
				`{"otherId":"`+otherID.String()+`"}`, true), id.String()))

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
		}
		if strings.Contains(rec.Body.String(), "boom") {
			t.Errorf("la respuesta filtra la causa interna: %s", rec.Body)
		}
	})

	t.Run("resolver marca", func(t *testing.T) {
		marca := MarcaDisponibilidadFueraDelFeed
		repo := newFakeRepository()
		ep := episodioPublicable("Con marca")
		ep.MarcaDisponibilidad = &marca
		repo.seed(ep)
		repo.mutationErr = errors.New("boom")
		id := repo.episodios[0].ID
		h := newTestHandler(t, repo)

		rec := httptest.NewRecorder()
		h.ResolverMarca(rec, withPathValue(
			adminRequest(http.MethodPost, "/api/v1/admin/medios/"+id.String()+"/marcas/resolver",
				`{"mark":"missing_from_feed","visibility":"visible"}`, true), id.String()))

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, se esperaba 500 (%s)", rec.Code, rec.Body)
		}
		if strings.Contains(rec.Body.String(), "boom") {
			t.Errorf("la respuesta filtra la causa interna: %s", rec.Body)
		}
	})
}
