package agenda

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/storage"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de la gestión de eventos del panel (T419): validación FR-015 sin
// guardado parcial, normalización analyze I6, duplicado del módulo (R4-9),
// reemplazo completo de fechas (R4-11), publicación al guardar (FR-014) y la
// regla exacta de códigos de auditoría (R4-10).

// validEventoInput devuelve un alta de evento válida.
func validEventoInput() EventoInput {
	return EventoInput{
		TitleEs:  "Noche de aposentos",
		DetailEs: "Trae tu Biblia y algo para compartir.",
		PlaceEs:  "Templo Central",
		Dates:    []FechaInput{{Date: "2026-10-12", StartTime: "19:00", EndTime: "21:00"}},
	}
}

// newAdminService construye el service con el reloj fijo y un almacén en memoria.
func newAdminService(t *testing.T, repo *fakeRepository) *service {
	t.Helper()
	location, err := time.LoadLocation("America/Costa_Rica")
	if err != nil {
		t.Fatalf("cargar America/Costa_Rica: %v", err)
	}
	logger, _ := testutil.NewLogger()
	return NewService(ServiceDeps{
		Repository: repo,
		Store:      &fakeStore{files: map[string][]byte{}},
		Location:   location,
		Now:        func() time.Time { return time.Date(2026, time.October, 11, 18, 0, 0, 0, time.UTC) },
		Logger:     logger,
	})
}

// fakeStore es un almacén en memoria para las pruebas de subida.
type fakeStore struct {
	files     map[string][]byte
	saveErr   error
	deleteErr error
	saved     []string
	deleted   []string
}

func (s *fakeStore) Save(_ context.Context, data []byte) (storage.File, error) {
	if s.saveErr != nil {
		return storage.File{}, s.saveErr
	}
	name := "img_99999999-9999-9999-9999-999999999999.jpg"
	s.files[name] = data
	s.saved = append(s.saved, name)
	return storage.File{Name: name, ContentType: "image/jpeg", SizeBytes: int64(len(data))}, nil
}

func (s *fakeStore) Open(_ context.Context, name string) (io.ReadCloser, error) {
	data, ok := s.files[name]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return io.NopCloser(strings.NewReader(string(data))), nil
}

func (s *fakeStore) Delete(_ context.Context, name string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.files, name)
	s.deleted = append(s.deleted, name)
	return nil
}

// assertNothingSaved comprueba que un rechazo no dejó nada guardado ni auditado.
func assertNothingSaved(t *testing.T, repo *fakeRepository) {
	t.Helper()
	if len(repo.events) != 0 {
		t.Errorf("se guardó un evento con datos inválidos: %+v", repo.events)
	}
	if len(repo.committed) != 0 {
		t.Errorf("se registró auditoría de una mutación que no ocurrió: %+v", repo.committed)
	}
}

// TestCreateEventoValidatesEveryField fija FR-015: cada dato inválido responde
// 400 con `details` por campo y SIN guardar nada.
func TestCreateEventoValidatesEveryField(t *testing.T) {
	unknownCategory := uuid.New()

	tests := []struct {
		name   string
		mutate func(*EventoInput)
		field  string
	}{
		{name: "título vacío", mutate: func(in *EventoInput) { in.TitleEs = "   " }, field: "titleEs"},
		{name: "título largo", mutate: func(in *EventoInput) { in.TitleEs = strings.Repeat("a", 161) }, field: "titleEs"},
		{name: "detalle vacío", mutate: func(in *EventoInput) { in.DetailEs = "" }, field: "detailEs"},
		{name: "detalle largo", mutate: func(in *EventoInput) { in.DetailEs = strings.Repeat("a", 2001) }, field: "detailEs"},
		{name: "lugar vacío", mutate: func(in *EventoInput) { in.PlaceEs = "  " }, field: "placeEs"},
		{name: "lugar largo", mutate: func(in *EventoInput) { in.PlaceEs = strings.Repeat("a", 201) }, field: "placeEs"},
		{name: "enlace de mapa inválido", mutate: func(in *EventoInput) { in.MapURL = "http://maps.example/x" }, field: "mapUrl"},
		{name: "enlace de mapa sin host", mutate: func(in *EventoInput) { in.MapURL = "https://" }, field: "mapUrl"},
		{name: "sin fechas", mutate: func(in *EventoInput) { in.Dates = nil }, field: "dates"},
		{name: "fecha inexistente", mutate: func(in *EventoInput) { in.Dates = []FechaInput{{Date: "2026-13-32"}} }, field: "date"},
		{name: "fecha repetida", mutate: func(in *EventoInput) {
			in.Dates = []FechaInput{{Date: "2026-10-12"}, {Date: "2026-10-12"}}
		}, field: "dates"},
		{name: "fin sin inicio", mutate: func(in *EventoInput) {
			in.Dates = []FechaInput{{Date: "2026-10-12", EndTime: "21:00"}}
		}, field: "endTime"},
		{name: "fin anterior", mutate: func(in *EventoInput) {
			in.Dates = []FechaInput{{Date: "2026-10-12", StartTime: "21:00", EndTime: "20:00"}}
		}, field: "endTime"},
		{name: "hora mal formada", mutate: func(in *EventoInput) {
			in.Dates = []FechaInput{{Date: "2026-10-12", StartTime: "7:30"}}
		}, field: "startTime"},
		{name: "imagen sin alt", mutate: func(in *EventoInput) {
			in.ImageFile = "img_11111111-1111-1111-1111-111111111111.jpg"
		}, field: "imageAltEs"},
		{name: "imagen con nombre inválido", mutate: func(in *EventoInput) {
			in.ImageFile = "../../etc/passwd"
			in.ImageAltEs = "Algo"
		}, field: "imageFile"},
		{name: "categoría inexistente", mutate: func(in *EventoInput) { in.CategoryID = unknownCategory.String() }, field: "categoryId"},
		{name: "categoría no identificable", mutate: func(in *EventoInput) { in.CategoryID = "no-es-un-uuid" }, field: "categoryId"},
		{name: "estado inválido", mutate: func(in *EventoInput) { in.PublicationState = "archived" }, field: "publicationState"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			svc := newAdminService(t, repo)

			in := validEventoInput()
			tt.mutate(&in)

			_, err := svc.CreateEvento(context.Background(), uuid.New(), in)
			if errorKind(err) != apperr.KindInvalid {
				t.Fatalf("error = %v, se esperaba invalid", err)
			}
			if _, ok := detailsOf(t, err)[tt.field]; !ok {
				t.Errorf("sin details.%s: %v", tt.field, err)
			}
			assertNothingSaved(t, repo)

			// El rechazo queda registrado best-effort con result='failure'.
			if len(repo.recorded) != 1 || repo.recorded[0].Result != audit.ResultFailure {
				t.Errorf("registro del rechazo = %+v", repo.recorded)
			}
		})
	}
}

// TestCreateEventoStoresNormalizedContent fija el alta completa (FR-010/FR-018):
// normalización de los `*En` vacíos a NULL (analyze I6), tildes/ñ/¿¡/emojis
// intactos, `event.create` aunque nazca publicado y el DTO del panel con sus
// fechas, `nextDate` e `isPast`.
func TestCreateEventoStoresNormalizedContent(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	actor := uuid.New()

	in := validEventoInput()
	in.TitleEs = "  Noche   de ¿aposentos? 🎉  "
	in.TitleEn = "   " // sin traducción → NULL
	in.DetailEs = "Culto de ñandú"
	in.FrequencyEs = "Todos los viernes"
	in.MapURL = "https://maps.example/iglesia"
	in.OrganizerEs = "Equipo de jóvenes"
	in.ImageFile = "img_11111111-1111-1111-1111-111111111111.jpg"
	in.ImageAltEs = "Foto del evento"
	in.PublicationState = "published"

	out, err := svc.CreateEvento(context.Background(), actor, in)
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}

	if len(repo.events) != 1 {
		t.Fatalf("eventos guardados = %d, se esperaba 1", len(repo.events))
	}
	stored := repo.events[0]
	if stored.TitleEs != "Noche de ¿aposentos? 🎉" {
		t.Errorf("título guardado = %q", stored.TitleEs)
	}
	if stored.TitleEn != nil {
		t.Errorf("TitleEn = %v, se esperaba NULL (analyze I6)", stored.TitleEn)
	}
	if stored.PublicationState != StatePublished {
		t.Errorf("estado guardado = %q", stored.PublicationState)
	}

	if len(repo.committed) != 1 {
		t.Fatalf("filas de auditoría = %d, se esperaba 1", len(repo.committed))
	}
	action := repo.committed[0]
	if action.Code != audit.ActionEventCreate {
		t.Errorf("código = %q, se esperaba event.create (el alta usa create aunque nazca publicada)", action.Code)
	}
	if action.TargetKind != audit.TargetContent || action.TargetLabel != "Eventos · Noche de ¿aposentos? 🎉" {
		t.Errorf("objetivo = %+v", action)
	}
	if action.ActorUserID == nil || *action.ActorUserID != actor {
		t.Errorf("actor = %v", action.ActorUserID)
	}

	if out.ID == "" || out.PublicationState != "published" {
		t.Errorf("DTO = %+v", out)
	}
	if len(out.Dates) != 1 || out.Dates[0].Date != "2026-10-12" || out.Dates[0].IsPast {
		t.Errorf("fechas del DTO = %+v", out.Dates)
	}
	if out.NextDate == nil || *out.NextDate != "2026-10-12" || out.IsPast {
		t.Errorf("nextDate/isPast = %v/%v", out.NextDate, out.IsPast)
	}
	if out.TitleEn != nil {
		t.Errorf("TitleEn del DTO = %v, se esperaba null", out.TitleEn)
	}
}

// TestCreateEventoDuplicateIsConflict fija R4-9: el duplicado del módulo (mismo
// título, lugar y una fecha en común) es un 409 sin guardar nada.
func TestCreateEventoDuplicateIsConflict(t *testing.T) {
	repo := newFakeRepository()
	repo.duplicate = true
	svc := newAdminService(t, repo)

	_, err := svc.CreateEvento(context.Background(), uuid.New(), validEventoInput())
	if errorKind(err) != apperr.KindConflict {
		t.Fatalf("error = %v, se esperaba conflict", err)
	}
	assertNothingSaved(t, repo)
	if len(repo.recorded) != 1 || repo.recorded[0].Result != audit.ResultFailure {
		t.Errorf("registro del rechazo = %+v", repo.recorded)
	}
}

// TestCreateEventoAllowsPastDates fija el edge case: las fechas pasadas se
// guardan, pero el elemento no es público (ni `nextDate`) sin una próxima, y su
// estado NO cambia (FR-006/Q4).
func TestCreateEventoAllowsPastDates(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)

	in := validEventoInput()
	in.Dates = []FechaInput{{Date: "2026-10-01"}}
	out, err := svc.CreateEvento(context.Background(), uuid.New(), in)
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}
	if out.NextDate != nil || !out.IsPast {
		t.Errorf("nextDate/isPast = %v/%v, se esperaba nil/true", out.NextDate, out.IsPast)
	}
	if !out.Dates[0].IsPast {
		t.Errorf("la fecha pasada debería venir marcada: %+v", out.Dates[0])
	}
	if out.PublicationState != "draft" {
		t.Errorf("estado = %q, se esperaba draft (el default del contrato)", out.PublicationState)
	}

	// La vista pública no lo muestra (no tiene ninguna fecha próxima).
	public, err := svc.GetProximamente(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetProximamente: %v", err)
	}
	if len(public.Events) != 0 {
		t.Errorf("un evento sin fecha próxima no debe ser público: %+v", public.Events)
	}
}

// TestUpdateEventoReplacesDatesAndPublishesOnSave fija R4-11 y FR-014: el PATCH
// con `dates` reemplaza la lista completa, el elemento publicado sigue publicado
// y el cambio de datos deja UNA fila `event.update`.
func TestUpdateEventoReplacesDatesAndPublishesOnSave(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)

	in := validEventoInput()
	in.PublicationState = "published"
	created, err := svc.CreateEvento(context.Background(), uuid.New(), in)
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}
	repo.committed = nil // se comprueba solo la edición

	detail := "Detalle editado"
	dates := []FechaInput{
		{Date: "2026-10-12"}, // la misma fecha que ya tenía
		{Date: "2026-10-19", StartTime: "19:00"},
	}
	updated, err := svc.UpdateEvento(context.Background(), uuid.New(), uuid.MustParse(created.ID), EventoPatch{
		DetailEs: &detail,
		Dates:    &dates,
	})
	if err != nil {
		t.Fatalf("UpdateEvento: %v", err)
	}

	if updated.DetailEs != "Detalle editado" {
		t.Errorf("detalle = %q", updated.DetailEs)
	}
	if updated.PublicationState != "published" {
		t.Errorf("estado = %q, se esperaba published (FR-014)", updated.PublicationState)
	}
	if len(updated.Dates) != 2 {
		t.Fatalf("fechas = %+v, se esperaban 2 (lista reemplazada)", updated.Dates)
	}
	if updated.NextDate == nil || *updated.NextDate != "2026-10-12" {
		t.Errorf("nextDate = %v", updated.NextDate)
	}

	if len(repo.committed) != 1 {
		t.Fatalf("filas de auditoría = %d, se esperaba 1 (solo datos)", len(repo.committed))
	}
	if code := repo.committed[0].Code; code != audit.ActionEventUpdate {
		t.Errorf("código = %q, se esperaba event.update", code)
	}
}

// TestUpdateEventoStateOnlyIsOneRow fija R4-10: publicar o retirar por elemento
// registra SOLO el código de estado.
func TestUpdateEventoStateOnlyIsOneRow(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)

	created, err := svc.CreateEvento(context.Background(), uuid.New(), validEventoInput())
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}
	repo.committed = nil

	published := "published"
	out, err := svc.UpdateEvento(context.Background(), uuid.New(), uuid.MustParse(created.ID), EventoPatch{PublicationState: &published})
	if err != nil {
		t.Fatalf("UpdateEvento: %v", err)
	}
	if out.PublicationState != "published" {
		t.Errorf("estado = %q", out.PublicationState)
	}
	if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionEventPublish {
		t.Fatalf("filas = %+v, se esperaba solo event.publish", repo.committed)
	}

	repo.committed = nil
	draft := "draft"
	if _, err := svc.UpdateEvento(context.Background(), uuid.New(), uuid.MustParse(created.ID), EventoPatch{PublicationState: &draft}); err != nil {
		t.Fatalf("UpdateEvento(retirar): %v", err)
	}
	if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionEventUnpublish {
		t.Fatalf("filas = %+v, se esperaba solo event.unpublish", repo.committed)
	}
}

// TestUpdateEventoDataAndStateLeavesTwoRows fija la regla de dos filas de R4-10.
func TestUpdateEventoDataAndStateLeavesTwoRows(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)

	created, err := svc.CreateEvento(context.Background(), uuid.New(), validEventoInput())
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}
	repo.committed = nil

	title := "Otro título"
	published := "published"
	if _, err := svc.UpdateEvento(context.Background(), uuid.New(), uuid.MustParse(created.ID), EventoPatch{
		TitleEs:          &title,
		PublicationState: &published,
	}); err != nil {
		t.Fatalf("UpdateEvento: %v", err)
	}

	if len(repo.committed) != 2 {
		t.Fatalf("filas = %+v, se esperaban 2 (datos + estado)", repo.committed)
	}
	if repo.committed[0].Code != audit.ActionEventUpdate || repo.committed[1].Code != audit.ActionEventPublish {
		t.Errorf("códigos = %q/%q", repo.committed[0].Code, repo.committed[1].Code)
	}
}

// TestUpdateEventoValidatesAndFinds cubre los rechazos de la edición: sin
// cambios, id inexistente y datos inválidos (sin guardar nada).
func TestUpdateEventoValidatesAndFinds(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)

	created, err := svc.CreateEvento(context.Background(), uuid.New(), validEventoInput())
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}
	id := uuid.MustParse(created.ID)
	repo.committed = nil
	repo.recorded = nil

	t.Run("sin cambios", func(t *testing.T) {
		if _, err := svc.UpdateEvento(context.Background(), uuid.New(), id, EventoPatch{}); errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
	})

	t.Run("id inexistente", func(t *testing.T) {
		title := "X"
		if _, err := svc.UpdateEvento(context.Background(), uuid.New(), uuid.New(), EventoPatch{TitleEs: &title}); errorKind(err) != apperr.KindNotFound {
			t.Fatalf("error = %v, se esperaba not_found", err)
		}
	})

	t.Run("título vacío", func(t *testing.T) {
		empty := "   "
		_, err := svc.UpdateEvento(context.Background(), uuid.New(), id, EventoPatch{TitleEs: &empty})
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["titleEs"]; !ok {
			t.Errorf("sin details.titleEs: %v", err)
		}
	})

	t.Run("fin sin inicio en las fechas", func(t *testing.T) {
		dates := []FechaInput{{Date: "2026-10-12", EndTime: "21:00"}}
		_, err := svc.UpdateEvento(context.Background(), uuid.New(), id, EventoPatch{Dates: &dates})
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["endTime"]; !ok {
			t.Errorf("sin details.endTime: %v", err)
		}
	})

	t.Run("imagen sin alt al quitarla", func(t *testing.T) {
		file := "img_11111111-1111-1111-1111-111111111111.jpg"
		_, err := svc.UpdateEvento(context.Background(), uuid.New(), id, EventoPatch{ImageFile: OptionalOf(file)})
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["imageAltEs"]; !ok {
			t.Errorf("sin details.imageAltEs: %v", err)
		}
	})

	t.Run("duplicado del módulo", func(t *testing.T) {
		repo.duplicate = true
		title := "Otro"
		if _, err := svc.UpdateEvento(context.Background(), uuid.New(), id, EventoPatch{TitleEs: &title}); errorKind(err) != apperr.KindConflict {
			t.Fatalf("error = %v, se esperaba conflict", err)
		}
	})

	// Ningún rechazo guardó nada.
	if len(repo.committed) != 0 {
		t.Errorf("se auditó una edición rechazada: %+v", repo.committed)
	}
	if len(repo.recorded) != 6 {
		t.Errorf("rechazos registrados = %d, se esperaban 6", len(repo.recorded))
	}
}

// TestUpdateEventoDoesNotSaveWhenAuditFails fija el fail-closed (FR-017): si el
// registro de auditoría falla, la mutación no se aplica.
func TestUpdateEventoDoesNotSaveWhenAuditFails(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)

	created, err := svc.CreateEvento(context.Background(), uuid.New(), validEventoInput())
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}
	repo.committed = nil
	repo.mutationErr = errors.New("la auditoría falló")

	title := "No debería guardarse"
	if _, err := svc.UpdateEvento(context.Background(), uuid.New(), uuid.MustParse(created.ID), EventoPatch{TitleEs: &title}); err == nil {
		t.Fatal("se esperaba error")
	}
	stored, _, err := repo.GetEventByID(context.Background(), uuid.MustParse(created.ID))
	if err != nil {
		t.Fatalf("GetEventByID: %v", err)
	}
	if stored.TitleEs == title {
		t.Error("la mutación se aplicó sin registro de auditoría")
	}
}

// TestUpdateEventoClearsOptionalFieldsWithNull fija el contrato: `null` explícito
// vacía los campos opcionales.
func TestUpdateEventoClearsOptionalFieldsWithNull(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)

	in := validEventoInput()
	in.MapURL = "https://maps.example/iglesia"
	in.OrganizerEs = "Equipo"
	in.TitleEn = "Dormitory night"
	created, err := svc.CreateEvento(context.Background(), uuid.New(), in)
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}

	updated, err := svc.UpdateEvento(context.Background(), uuid.New(), uuid.MustParse(created.ID), EventoPatch{
		MapURL:      Optional[string]{}, // ausente: no se toca
		OrganizerEs: OptionalOf(""),     // vacío: se limpia
		TitleEn:     nullOptional(),     // null explícito: se limpia
	})
	if err != nil {
		t.Fatalf("UpdateEvento: %v", err)
	}
	if updated.MapURL == nil || *updated.MapURL != "https://maps.example/iglesia" {
		t.Errorf("mapUrl = %v, no debía tocarse", updated.MapURL)
	}
	if updated.OrganizerEs != nil {
		t.Errorf("organizerEs = %v, se esperaba null", updated.OrganizerEs)
	}
	if updated.TitleEn != nil {
		t.Errorf("titleEn = %v, se esperaba null", updated.TitleEn)
	}
}

// nullOptional construye un Optional presente con valor `null`.
func nullOptional() Optional[string] {
	var o Optional[string]
	if err := o.UnmarshalJSON([]byte("null")); err != nil {
		panic(err)
	}
	return o
}

// TestDeleteEventoAuditsAndFinds cubre el borrado con su registro y el 404.
func TestDeleteEventoAuditsAndFinds(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)

	created, err := svc.CreateEvento(context.Background(), uuid.New(), validEventoInput())
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}
	id := uuid.MustParse(created.ID)
	repo.committed = nil

	if err := svc.DeleteEvento(context.Background(), uuid.New(), id); err != nil {
		t.Fatalf("DeleteEvento: %v", err)
	}
	if len(repo.events) != 0 {
		t.Errorf("el evento no se borró: %+v", repo.events)
	}
	if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionEventDelete {
		t.Errorf("auditoría = %+v, se esperaba event.delete", repo.committed)
	}
	if repo.committed[0].TargetLabel != "Eventos · Noche de aposentos" {
		t.Errorf("targetLabel = %q", repo.committed[0].TargetLabel)
	}

	if err := svc.DeleteEvento(context.Background(), uuid.New(), uuid.New()); errorKind(err) != apperr.KindNotFound {
		t.Errorf("borrar un id inexistente = %v, se esperaba not_found", err)
	}
}

// TestGetAndListEventos cubre la ficha y el listado paginado del panel.
func TestGetAndListEventos(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	published := validEventoInput()
	published.PublicationState = "published"
	first, err := svc.CreateEvento(ctx, uuid.New(), published)
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}
	draft := validEventoInput()
	draft.TitleEs = "Evento en borrador"
	if _, err := svc.CreateEvento(ctx, uuid.New(), draft); err != nil {
		t.Fatalf("CreateEvento(borrador): %v", err)
	}

	if _, err := svc.GetEvento(ctx, uuid.New()); errorKind(err) != apperr.KindNotFound {
		t.Errorf("ficha de un id inexistente = %v, se esperaba not_found", err)
	}
	got, err := svc.GetEvento(ctx, uuid.MustParse(first.ID))
	if err != nil || got.ID != first.ID {
		t.Fatalf("GetEvento = %+v (%v)", got, err)
	}

	list, err := svc.ListEventos(ctx, nil, paginate.Params{Limit: 20})
	if err != nil {
		t.Fatalf("ListEventos: %v", err)
	}
	if list.Total != 2 || len(list.Items) != 2 || list.Limit != 20 {
		t.Errorf("listado = %+v", list)
	}

	state := StateDraft
	drafts, err := svc.ListEventos(ctx, &state, paginate.Params{Limit: 20})
	if err != nil {
		t.Fatalf("ListEventos(borradores): %v", err)
	}
	if drafts.Total != 1 || drafts.Items[0].TitleEs != "Evento en borrador" {
		t.Errorf("borradores = %+v", drafts)
	}

	page, err := svc.ListEventos(ctx, nil, paginate.Params{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("ListEventos(página 2): %v", err)
	}
	if len(page.Items) != 1 {
		t.Errorf("página = %+v", page.Items)
	}
}

// TestUploadEventoImageAuditsAndIsFailClosed fija FR-020/R4-10: la subida de
// imagen del módulo se audita con `event.image.upload` y es fail-closed (si el
// registro falla, se borra el archivo).
func TestUploadEventoImageAuditsAndIsFailClosed(t *testing.T) {
	t.Run("subida correcta", func(t *testing.T) {
		repo := newFakeRepository()
		svc := newAdminService(t, repo)
		store := svc.store.(*fakeStore)

		out, err := svc.UploadEventoImage(context.Background(), uuid.New(), []byte("\x89PNG\r\n\x1a\n"))
		if err != nil {
			t.Fatalf("UploadEventoImage: %v", err)
		}
		if out.FileName == "" || out.URL != MediaPathPrefix+out.FileName || out.MimeType != "image/jpeg" {
			t.Errorf("resultado = %+v", out)
		}
		if len(store.saved) != 1 {
			t.Errorf("archivos guardados = %v", store.saved)
		}
		if len(repo.recorded) != 1 || repo.recorded[0].Code != audit.ActionEventImageUpload {
			t.Fatalf("auditoría = %+v, se esperaba event.image.upload", repo.recorded)
		}
		if repo.recorded[0].TargetLabel != "Eventos · Imagen · "+out.FileName {
			t.Errorf("targetLabel = %q", repo.recorded[0].TargetLabel)
		}
	})

	t.Run("sin datos", func(t *testing.T) {
		repo := newFakeRepository()
		svc := newAdminService(t, repo)

		_, err := svc.UploadEventoImage(context.Background(), uuid.New(), nil)
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["file"]; !ok {
			t.Errorf("sin details.file: %v", err)
		}
	})

	t.Run("formato inválido", func(t *testing.T) {
		repo := newFakeRepository()
		svc := newAdminService(t, repo)
		svc.store.(*fakeStore).saveErr = storage.ErrInvalidFormat

		_, err := svc.UploadEventoImage(context.Background(), uuid.New(), []byte("no es una imagen"))
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["file"]; !ok {
			t.Errorf("sin details.file: %v", err)
		}
	})

	t.Run("demasiado grande", func(t *testing.T) {
		repo := newFakeRepository()
		svc := newAdminService(t, repo)
		svc.store.(*fakeStore).saveErr = storage.ErrTooLarge

		_, err := svc.UploadEventoImage(context.Background(), uuid.New(), []byte("grande"))
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["fileSize"]; !ok {
			t.Errorf("sin details.fileSize: %v", err)
		}
	})

	t.Run("fail-closed si falla el registro", func(t *testing.T) {
		repo := newFakeRepository()
		repo.recordErr = errors.New("auditoría caída")
		svc := newAdminService(t, repo)
		store := svc.store.(*fakeStore)

		if _, err := svc.UploadEventoImage(context.Background(), uuid.New(), []byte("PNG")); err == nil {
			t.Fatal("se esperaba error")
		}
		if len(store.deleted) != 1 {
			t.Errorf("el archivo no se borró tras el fallo del registro: %+v", store.deleted)
		}
	})
}
