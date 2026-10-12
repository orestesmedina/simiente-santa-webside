package agenda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// --- Actividades (T420) ---

// validActividadInput devuelve un alta de actividad válida.
func validActividadInput() ActividadInput {
	return ActividadInput{
		TitleEs:   "Repartir comida",
		DetailEs:  "Repartimos comida en el barrio.",
		PlaceEs:   "Salón comunal",
		Date:      "2026-10-12",
		StartTime: "10:00",
		EndTime:   "13:00",
	}
}

// TestCreateActividadValidatesEveryField fija FR-015 sobre el modelo de fecha
// única: cada dato inválido responde 400 con `details` y sin guardar nada.
func TestCreateActividadValidatesEveryField(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ActividadInput)
		field  string
	}{
		{name: "título vacío", mutate: func(in *ActividadInput) { in.TitleEs = " " }, field: "titleEs"},
		{name: "detalle vacío", mutate: func(in *ActividadInput) { in.DetailEs = "" }, field: "detailEs"},
		{name: "detalle largo", mutate: func(in *ActividadInput) { in.DetailEs = strings.Repeat("a", 2001) }, field: "detailEs"},
		{name: "lugar vacío", mutate: func(in *ActividadInput) { in.PlaceEs = "  " }, field: "placeEs"},
		{name: "enlace inválido", mutate: func(in *ActividadInput) { in.MapURL = "ftp://x.example" }, field: "mapUrl"},
		{name: "sin fecha", mutate: func(in *ActividadInput) { in.Date = "" }, field: "date"},
		{name: "fecha inexistente", mutate: func(in *ActividadInput) { in.Date = "2026-02-30" }, field: "date"},
		{name: "fin sin inicio", mutate: func(in *ActividadInput) { in.StartTime = "" }, field: "endTime"},
		{name: "fin anterior", mutate: func(in *ActividadInput) { in.EndTime = "09:00" }, field: "endTime"},
		{name: "hora mal formada", mutate: func(in *ActividadInput) { in.StartTime = "25:00" }, field: "startTime"},
		{name: "imagen sin alt", mutate: func(in *ActividadInput) {
			in.ImageFile = "img_11111111-1111-1111-1111-111111111111.png"
		}, field: "imageAltEs"},
		{name: "categoría inexistente", mutate: func(in *ActividadInput) { in.CategoryID = uuid.New().String() }, field: "categoryId"},
		{name: "estado inválido", mutate: func(in *ActividadInput) { in.PublicationState = "archived" }, field: "publicationState"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			svc := newAdminService(t, repo)

			in := validActividadInput()
			tt.mutate(&in)

			_, err := svc.CreateActividad(context.Background(), uuid.New(), in)
			if errorKind(err) != apperr.KindInvalid {
				t.Fatalf("error = %v, se esperaba invalid", err)
			}
			if _, ok := detailsOf(t, err)[tt.field]; !ok {
				t.Errorf("sin details.%s: %v", tt.field, err)
			}
			if len(repo.activities) != 0 || len(repo.committed) != 0 {
				t.Errorf("un rechazo guardó algo: %+v / %+v", repo.activities, repo.committed)
			}
			if len(repo.recorded) != 1 || repo.recorded[0].Result != audit.ResultFailure {
				t.Errorf("registro del rechazo = %+v", repo.recorded)
			}
		})
	}
}

// TestCreateActividadStoresNormalizedContent fija el alta completa (FR-011):
// normalización analyze I6, `activity.create` aunque nazca publicada (R4-10) y
// el DTO del panel con su fecha y su marca de pasado.
func TestCreateActividadStoresNormalizedContent(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	actor := uuid.New()

	in := validActividadInput()
	in.TitleEs = "  Repartir   comida ¡ya! 🍲 "
	in.DetailEn = "   " // sin traducción → NULL
	in.PublicationState = "published"

	out, err := svc.CreateActividad(context.Background(), actor, in)
	if err != nil {
		t.Fatalf("CreateActividad: %v", err)
	}

	stored := repo.activities[0]
	if stored.TitleEs != "Repartir comida ¡ya! 🍲" {
		t.Errorf("título = %q", stored.TitleEs)
	}
	if stored.DetailEn != nil {
		t.Errorf("DetailEn = %v, se esperaba NULL", stored.DetailEn)
	}
	if stored.Date != "2026-10-12" || stored.StartTime == nil || *stored.StartTime != "10:00" {
		t.Errorf("fecha/horario = %+v", stored)
	}

	if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionActivityCreate {
		t.Fatalf("auditoría = %+v, se esperaba activity.create", repo.committed)
	}
	if repo.committed[0].TargetLabel != "Actividades · Repartir comida ¡ya! 🍲" {
		t.Errorf("targetLabel = %q", repo.committed[0].TargetLabel)
	}

	if out.PublicationState != "published" || out.IsPast {
		t.Errorf("DTO = %+v", out)
	}
	if out.Date != "2026-10-12" {
		t.Errorf("date = %q", out.Date)
	}
}

// TestCreateActividadDuplicateIsConflict fija R4-9 en el módulo de actividades:
// el duplicado exacto (título + fecha + lugar) es 409 y un EVENTO idéntico no
// cuenta como duplicado (Decisión 3).
func TestCreateActividadDuplicateIsConflict(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	if _, err := svc.CreateActividad(ctx, uuid.New(), validActividadInput()); err != nil {
		t.Fatalf("CreateActividad: %v", err)
	}
	repo.committed = nil
	repo.recorded = nil

	_, err := svc.CreateActividad(ctx, uuid.New(), validActividadInput())
	if errorKind(err) != apperr.KindConflict {
		t.Fatalf("duplicado = %v, se esperaba conflict", err)
	}
	if len(repo.activities) != 1 {
		t.Errorf("actividades = %d, se esperaba 1", len(repo.activities))
	}
	if len(repo.recorded) != 1 || repo.recorded[0].Result != audit.ResultFailure {
		t.Errorf("registro del rechazo = %+v", repo.recorded)
	}

	// Un evento con los mismos datos se guarda sin conflicto (Decisión 3).
	in := validEventoInput()
	in.TitleEs = validActividadInput().TitleEs
	in.PlaceEs = validActividadInput().PlaceEs
	in.Dates = []FechaInput{{Date: validActividadInput().Date}}
	if _, err := svc.CreateEvento(ctx, uuid.New(), in); err != nil {
		t.Fatalf("CreateEvento(idéntico a una actividad): %v", err)
	}
}

// TestActividadPastDateIsNotPublic fija el edge case: la fecha pasada se guarda
// (marcada `isPast`) y la actividad no aparece en público, sin cambiar su estado
// (FR-006/Q4).
func TestActividadPastDateIsNotPublic(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	in := validActividadInput()
	in.Date = "2026-10-01"
	in.PublicationState = "published"
	out, err := svc.CreateActividad(ctx, uuid.New(), in)
	if err != nil {
		t.Fatalf("CreateActividad: %v", err)
	}
	if !out.IsPast {
		t.Errorf("isPast = false, se esperaba true")
	}

	public, err := svc.GetProximamente(ctx, LangES)
	if err != nil {
		t.Fatalf("GetProximamente: %v", err)
	}
	if len(public.Activities) != 0 {
		t.Errorf("una actividad vencida no debe ser pública: %+v", public.Activities)
	}

	stored := repo.activities[0]
	if stored.PublicationState != StatePublished {
		t.Errorf("el estado de la vencida cambió: %q", stored.PublicationState)
	}
}

// TestUpdateActividadKeepsModulesApart fija la independencia de los módulos y la
// regla de códigos (R4-10): publicar o retirar una actividad deja UNA fila
// `activity.*` y no toca a los eventos.
func TestUpdateActividadKeepsModulesApart(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	eventIn := validEventoInput()
	eventIn.PublicationState = "published"
	event, err := svc.CreateEvento(ctx, uuid.New(), eventIn)
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}
	activity, err := svc.CreateActividad(ctx, uuid.New(), validActividadInput())
	if err != nil {
		t.Fatalf("CreateActividad: %v", err)
	}
	repo.committed = nil

	published := "published"
	updated, err := svc.UpdateActividad(ctx, uuid.New(), uuid.MustParse(activity.ID), ActividadPatch{PublicationState: &published})
	if err != nil {
		t.Fatalf("UpdateActividad: %v", err)
	}
	if updated.PublicationState != "published" {
		t.Errorf("estado = %q", updated.PublicationState)
	}
	if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionActivityPublish {
		t.Fatalf("auditoría = %+v, se esperaba solo activity.publish", repo.committed)
	}

	// El evento no se tocó.
	storedEvent, _, err := repo.GetEventByID(ctx, uuid.MustParse(event.ID))
	if err != nil {
		t.Fatalf("GetEventByID: %v", err)
	}
	if storedEvent.PublicationState != StatePublished {
		t.Errorf("el evento cambió de estado: %q", storedEvent.PublicationState)
	}

	// Datos + estado → dos filas.
	repo.committed = nil
	title := "Repartir comida (editada)"
	draft := "draft"
	if _, err := svc.UpdateActividad(ctx, uuid.New(), uuid.MustParse(activity.ID), ActividadPatch{
		TitleEs:          &title,
		PublicationState: &draft,
	}); err != nil {
		t.Fatalf("UpdateActividad(editar): %v", err)
	}
	if len(repo.committed) != 2 {
		t.Fatalf("auditoría = %+v, se esperaban 2 filas", repo.committed)
	}
	if repo.committed[0].Code != audit.ActionActivityUpdate || repo.committed[1].Code != audit.ActionActivityUnpublish {
		t.Errorf("códigos = %q/%q", repo.committed[0].Code, repo.committed[1].Code)
	}
}

// TestUpdateActividadValidatesAndFinds cubre los rechazos de la edición de
// actividades (incluido el horario resultante) y el 404.
func TestUpdateActividadValidatesAndFinds(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	created, err := svc.CreateActividad(ctx, uuid.New(), validActividadInput())
	if err != nil {
		t.Fatalf("CreateActividad: %v", err)
	}
	id := uuid.MustParse(created.ID)

	t.Run("sin cambios", func(t *testing.T) {
		if _, err := svc.UpdateActividad(ctx, uuid.New(), id, ActividadPatch{}); errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
	})

	t.Run("id inexistente", func(t *testing.T) {
		title := "X"
		if _, err := svc.UpdateActividad(ctx, uuid.New(), uuid.New(), ActividadPatch{TitleEs: &title}); errorKind(err) != apperr.KindNotFound {
			t.Fatalf("error = %v, se esperaba not_found", err)
		}
	})

	t.Run("título vacío", func(t *testing.T) {
		empty := "  "
		if _, err := svc.UpdateActividad(ctx, uuid.New(), id, ActividadPatch{TitleEs: &empty}); errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
	})

	t.Run("fin sin inicio al editar el horario", func(t *testing.T) {
		_, err := svc.UpdateActividad(ctx, uuid.New(), id, ActividadPatch{StartTime: nullOptional()})
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["endTime"]; !ok {
			t.Errorf("sin details.endTime: %v", err)
		}
	})

	t.Run("fecha inexistente", func(t *testing.T) {
		bad := "2026-13-32"
		if _, err := svc.UpdateActividad(ctx, uuid.New(), id, ActividadPatch{Date: &bad}); errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
	})

	t.Run("imagen sin alt", func(t *testing.T) {
		file := "img_11111111-1111-1111-1111-111111111111.png"
		_, err := svc.UpdateActividad(ctx, uuid.New(), id, ActividadPatch{ImageFile: OptionalOf(file)})
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["imageAltEs"]; !ok {
			t.Errorf("sin details.imageAltEs: %v", err)
		}
	})
}

// TestUpdateActividadClearsOptionalFieldsWithNull fija el contrato del PATCH.
func TestUpdateActividadClearsOptionalFieldsWithNull(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)

	in := validActividadInput()
	in.OrganizerEs = "Equipo"
	in.TitleEn = "Food distribution"
	created, err := svc.CreateActividad(context.Background(), uuid.New(), in)
	if err != nil {
		t.Fatalf("CreateActividad: %v", err)
	}

	updated, err := svc.UpdateActividad(context.Background(), uuid.New(), uuid.MustParse(created.ID), ActividadPatch{
		OrganizerEs: OptionalOf(""),
		TitleEn:     nullOptional(),
	})
	if err != nil {
		t.Fatalf("UpdateActividad: %v", err)
	}
	if updated.OrganizerEs != nil || updated.TitleEn != nil {
		t.Errorf("los opcionales no se limpiaron: %+v", updated)
	}
	if updated.StartTime == nil || *updated.StartTime != "10:00" {
		t.Errorf("el horario no debía tocarse: %+v", updated.StartTime)
	}
}

// TestDeleteGetAndListActividades cubre el borrado, la ficha y el listado del
// panel con su filtro y paginación.
func TestDeleteGetAndListActividades(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	published := validActividadInput()
	published.PublicationState = "published"
	first, err := svc.CreateActividad(ctx, uuid.New(), published)
	if err != nil {
		t.Fatalf("CreateActividad: %v", err)
	}
	draftIn := validActividadInput()
	draftIn.TitleEs = "Actividad en borrador"
	draftIn.PlaceEs = "Otro salón"
	if _, err := svc.CreateActividad(ctx, uuid.New(), draftIn); err != nil {
		t.Fatalf("CreateActividad(borrador): %v", err)
	}
	repo.committed = nil

	list, err := svc.ListActividades(ctx, nil, paginate.Params{Limit: 20})
	if err != nil {
		t.Fatalf("ListActividades: %v", err)
	}
	if list.Total != 2 || len(list.Items) != 2 {
		t.Errorf("listado = %+v", list)
	}
	state := StatePublished
	filtered, err := svc.ListActividades(ctx, &state, paginate.Params{Limit: 20})
	if err != nil {
		t.Fatalf("ListActividades(publicadas): %v", err)
	}
	if filtered.Total != 1 {
		t.Errorf("publicadas = %+v", filtered)
	}
	page, err := svc.ListActividades(ctx, nil, paginate.Params{Limit: 1, Offset: 1})
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("página = %+v (%v)", page, err)
	}

	got, err := svc.GetActividad(ctx, uuid.MustParse(first.ID))
	if err != nil || got.ID != first.ID {
		t.Fatalf("GetActividad = %+v (%v)", got, err)
	}
	if _, err := svc.GetActividad(ctx, uuid.New()); errorKind(err) != apperr.KindNotFound {
		t.Errorf("ficha inexistente = %v, se esperaba not_found", err)
	}

	if err := svc.DeleteActividad(ctx, uuid.New(), uuid.MustParse(first.ID)); err != nil {
		t.Fatalf("DeleteActividad: %v", err)
	}
	if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionActivityDelete {
		t.Errorf("auditoría = %+v, se esperaba activity.delete", repo.committed)
	}
	if err := svc.DeleteActividad(ctx, uuid.New(), uuid.New()); errorKind(err) != apperr.KindNotFound {
		t.Errorf("borrado inexistente = %v, se esperaba not_found", err)
	}
}

// TestUploadActividadImageAudits fija la subida del módulo de actividades
// (R4-10: `activity.image.upload`).
func TestUploadActividadImageAudits(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)

	out, err := svc.UploadActividadImage(context.Background(), uuid.New(), []byte("PNG"))
	if err != nil {
		t.Fatalf("UploadActividadImage: %v", err)
	}
	if len(repo.recorded) != 1 || repo.recorded[0].Code != audit.ActionActivityImageUpload {
		t.Fatalf("auditoría = %+v, se esperaba activity.image.upload", repo.recorded)
	}
	if repo.recorded[0].TargetLabel != "Actividades · Imagen · "+out.FileName {
		t.Errorf("targetLabel = %q", repo.recorded[0].TargetLabel)
	}
}

// --- Catálogo de categorías (T421) ---

// TestCreateCategoriaValidatesAndAudits fija Q6/Q7: nombre en español
// obligatorio y acotado, inglés opcional y registro `category.create` (R4-10).
func TestCreateCategoriaValidatesAndAudits(t *testing.T) {
	t.Run("nombre vacío", func(t *testing.T) {
		repo := newFakeRepository()
		svc := newAdminService(t, repo)

		_, err := svc.CreateCategoria(context.Background(), uuid.New(), CategoriaInput{NameEs: "   "})
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["nameEs"]; !ok {
			t.Errorf("sin details.nameEs: %v", err)
		}
		if len(repo.categories) != 0 || len(repo.committed) != 0 {
			t.Errorf("un rechazo guardó algo: %+v", repo.categories)
		}
	})

	t.Run("nombre largo", func(t *testing.T) {
		repo := newFakeRepository()
		svc := newAdminService(t, repo)

		_, err := svc.CreateCategoria(context.Background(), uuid.New(), CategoriaInput{NameEs: strings.Repeat("a", 121)})
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["nameEs"]; !ok {
			t.Errorf("sin details.nameEs: %v", err)
		}
	})

	t.Run("alta correcta", func(t *testing.T) {
		repo := newFakeRepository()
		svc := newAdminService(t, repo)

		out, err := svc.CreateCategoria(context.Background(), uuid.New(), CategoriaInput{
			NameEs: "  Acción   social ", NameEn: "   ", SortOrder: 10,
		})
		if err != nil {
			t.Fatalf("CreateCategoria: %v", err)
		}
		if out.NameEs != "Acción social" || out.NameEn != nil || out.SortOrder != 10 {
			t.Errorf("categoría = %+v", out)
		}
		if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionCategoryCreate {
			t.Fatalf("auditoría = %+v, se esperaba category.create", repo.committed)
		}
		if repo.committed[0].TargetLabel != "Categorías · Acción social" {
			t.Errorf("targetLabel = %q", repo.committed[0].TargetLabel)
		}
	})

	t.Run("límite del catálogo", func(t *testing.T) {
		repo := newFakeRepository()
		svc := newAdminService(t, repo)
		for i := 0; i < MaxCategories; i++ {
			repo.categories = append(repo.categories, Category{ID: uuid.New(), NameEs: fmt.Sprintf("Categoría %d", i)})
		}

		_, err := svc.CreateCategoria(context.Background(), uuid.New(), CategoriaInput{NameEs: "Una más"})
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("error = %v, se esperaba invalid", err)
		}
		if len(repo.categories) != MaxCategories {
			t.Errorf("el catálogo creció: %d", len(repo.categories))
		}
	})

	t.Run("nombre duplicado", func(t *testing.T) {
		repo := newFakeRepository()
		svc := newAdminService(t, repo)
		ctx := context.Background()

		if _, err := svc.CreateCategoria(ctx, uuid.New(), CategoriaInput{NameEs: "Predicación"}); err != nil {
			t.Fatalf("CreateCategoria: %v", err)
		}
		repo.recorded = nil

		if _, err := svc.CreateCategoria(ctx, uuid.New(), CategoriaInput{NameEs: "Predicación"}); errorKind(err) != apperr.KindConflict {
			t.Fatalf("duplicado = %v, se esperaba conflict", err)
		}
		if len(repo.recorded) != 1 || repo.recorded[0].Result != audit.ResultFailure {
			t.Errorf("registro del rechazo = %+v", repo.recorded)
		}
	})
}

// TestUpdateCategoriaAllowsRenaming fija Q6/R4-5: renombrar está permitido (el
// cambio se ve de inmediato en lo publicado) y el duplicado es 409.
func TestUpdateCategoriaAllowsRenaming(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	first, err := svc.CreateCategoria(ctx, uuid.New(), CategoriaInput{NameEs: "Adoración", SortOrder: 30})
	if err != nil {
		t.Fatalf("CreateCategoria: %v", err)
	}
	second, err := svc.CreateCategoria(ctx, uuid.New(), CategoriaInput{NameEs: "Entrevista", SortOrder: 40})
	if err != nil {
		t.Fatalf("CreateCategoria: %v", err)
	}
	repo.committed = nil

	name := "Adoración y alabanza"
	order := 35
	updated, err := svc.UpdateCategoria(ctx, uuid.New(), uuid.MustParse(first.ID), CategoriaPatch{NameEs: &name, SortOrder: &order})
	if err != nil {
		t.Fatalf("UpdateCategoria: %v", err)
	}
	if updated.NameEs != "Adoración y alabanza" || updated.SortOrder != 35 {
		t.Errorf("categoría = %+v", updated)
	}
	if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionCategoryUpdate {
		t.Fatalf("auditoría = %+v, se esperaba category.update", repo.committed)
	}

	// Renombrar a un nombre ya usado por OTRA categoría → 409.
	duplicated := "Adoración y alabanza"
	if _, err := svc.UpdateCategoria(ctx, uuid.New(), uuid.MustParse(second.ID), CategoriaPatch{NameEs: &duplicated}); errorKind(err) != apperr.KindConflict {
		t.Fatalf("duplicado = %v, se esperaba conflict", err)
	}

	// Sin cambios / id inexistente / nombre vacío.
	if _, err := svc.UpdateCategoria(ctx, uuid.New(), uuid.MustParse(first.ID), CategoriaPatch{}); errorKind(err) != apperr.KindInvalid {
		t.Errorf("sin cambios = %v, se esperaba invalid", err)
	}
	if _, err := svc.UpdateCategoria(ctx, uuid.New(), uuid.New(), CategoriaPatch{NameEs: &name}); errorKind(err) != apperr.KindNotFound {
		t.Errorf("id inexistente = %v, se esperaba not_found", err)
	}
	empty := "  "
	if _, err := svc.UpdateCategoria(ctx, uuid.New(), uuid.MustParse(first.ID), CategoriaPatch{NameEs: &empty}); errorKind(err) != apperr.KindInvalid {
		t.Errorf("nombre vacío = %v, se esperaba invalid", err)
	}

	// `null` explícito vacía el nombre en inglés.
	nullEn, err := svc.UpdateCategoria(ctx, uuid.New(), uuid.MustParse(first.ID), CategoriaPatch{NameEn: nullOptional()})
	if err != nil {
		t.Fatalf("UpdateCategoria(nameEn null): %v", err)
	}
	if nullEn.NameEn != nil {
		t.Errorf("nameEn = %v, se esperaba null", nullEn.NameEn)
	}
}

// TestDeleteCategoriaInUseIsConflict fija la decisión firme de la puerta 2
// (P4-6): una categoría en uso NO se borra (409 con el recuento) y NO se desliga
// de los elementos; sin uso, el borrado procede con su registro.
func TestDeleteCategoriaInUseIsConflict(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	category, err := svc.CreateCategoria(ctx, uuid.New(), CategoriaInput{NameEs: "Noche de oración"})
	if err != nil {
		t.Fatalf("CreateCategoria: %v", err)
	}
	categoryID := uuid.MustParse(category.ID)
	free, err := svc.CreateCategoria(ctx, uuid.New(), CategoriaInput{NameEs: "Otro"})
	if err != nil {
		t.Fatalf("CreateCategoria: %v", err)
	}

	// Un evento y una actividad usan la categoría.
	eventIn := validEventoInput()
	eventIn.CategoryID = category.ID
	if _, err := svc.CreateEvento(ctx, uuid.New(), eventIn); err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}
	activityIn := validActividadInput()
	activityIn.CategoryID = category.ID
	if _, err := svc.CreateActividad(ctx, uuid.New(), activityIn); err != nil {
		t.Fatalf("CreateActividad: %v", err)
	}
	repo.committed = nil
	repo.recorded = nil

	err = svc.DeleteCategoria(ctx, uuid.New(), categoryID)
	if errorKind(err) != apperr.KindConflict {
		t.Fatalf("borrado en uso = %v, se esperaba conflict", err)
	}
	if !strings.Contains(err.Error(), "está en uso por 2 elementos") {
		t.Errorf("mensaje = %q, se esperaba el recuento", err.Error())
	}
	if _, ok := detailsOf(t, err)["categoryId"]; !ok {
		t.Errorf("sin details.categoryId: %v", err)
	}
	if len(repo.categories) != 2 {
		t.Errorf("la categoría en uso desapareció: %+v", repo.categories)
	}
	if len(repo.recorded) != 1 || repo.recorded[0].Result != audit.ResultFailure {
		t.Errorf("registro del rechazo = %+v", repo.recorded)
	}

	// Al quitar la categoría de sus usos, el borrado procede.
	stored, _, err := repo.GetEventByID(ctx, repo.events[0].ID)
	if err != nil {
		t.Fatalf("GetEventByID: %v", err)
	}
	stored.CategoryID = nil
	if _, err := repo.UpdateEvent(ctx, stored, nil); err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}
	storedActivity := repo.activities[0]
	storedActivity.CategoryID = nil
	if _, err := repo.UpdateActivity(ctx, storedActivity); err != nil {
		t.Fatalf("UpdateActivity: %v", err)
	}

	if err := svc.DeleteCategoria(ctx, uuid.New(), categoryID); err != nil {
		t.Fatalf("DeleteCategoria(sin uso): %v", err)
	}
	if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionCategoryDelete {
		t.Errorf("auditoría = %+v, se esperaba category.delete", repo.committed)
	}
	if repo.committed[0].TargetLabel != "Categorías · Noche de oración" {
		t.Errorf("targetLabel = %q", repo.committed[0].TargetLabel)
	}

	// Sin uso y sin existir.
	if err := svc.DeleteCategoria(ctx, uuid.New(), uuid.MustParse(free.ID)); err != nil {
		t.Fatalf("DeleteCategoria(libre): %v", err)
	}
	if err := svc.DeleteCategoria(ctx, uuid.New(), uuid.New()); errorKind(err) != apperr.KindNotFound {
		t.Errorf("borrado inexistente = %v, se esperaba not_found", err)
	}
}

// TestListCategoriasKeepsOrder fija que el catálogo se devuelve en su orden de
// presentación (sort_order, id) y sin estado de publicación.
func TestListCategoriasKeepsOrder(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	for _, in := range []CategoriaInput{{NameEs: "Otro", SortOrder: 999}, {NameEs: "Acción social", SortOrder: 10}} {
		if _, err := svc.CreateCategoria(ctx, uuid.New(), in); err != nil {
			t.Fatalf("CreateCategoria: %v", err)
		}
	}
	// El repositorio real ordena por `sort_order, id`; el falso conserva el orden
	// de inserción, así que se comprueba el contrato del DTO.
	list, err := svc.ListCategorias(ctx)
	if err != nil {
		t.Fatalf("ListCategorias: %v", err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("categorías = %+v", list.Items)
	}
	body, err := json.Marshal(list)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(body), "publicationState") {
		t.Errorf("el catálogo no tiene estado de publicación: %s", body)
	}
	if list.Items[0].SortOrder != 999 || list.Items[1].SortOrder != 10 {
		t.Errorf("orden devuelto = %+v", list.Items)
	}
}

// TestUpdateEventoCoversEveryPatchField recorre todos los campos del PATCH de
// eventos (valores, cambios y `null` explícito) para fijar la mezcla completa.
func TestUpdateEventoCoversEveryPatchField(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	category, err := svc.CreateCategoria(ctx, uuid.New(), CategoriaInput{NameEs: "Acción social"})
	if err != nil {
		t.Fatalf("CreateCategoria: %v", err)
	}
	created, err := svc.CreateEvento(ctx, uuid.New(), validEventoInput())
	if err != nil {
		t.Fatalf("CreateEvento: %v", err)
	}
	id := uuid.MustParse(created.ID)

	// PATCH con TODOS los campos con valor.
	title, detail, place := "Noche de vigilia", "Detalle nuevo", "Templo Nuevo"
	frequency, mapURL, organizer := "Cada primer viernes", "https://maps.example/vigilia", "Equipo"
	titleEn, detailEn, frequencyEn, placeEn, organizerEn := "Vigil night", "New detail", "Every Friday", "New temple", "Team"
	file, altEs, altEn, published := "img_11111111-1111-1111-1111-111111111111.jpg", "Foto de la vigilia", "Vigil photo", "published"
	dates := []FechaInput{{Date: "2026-11-06", StartTime: "19:00", EndTime: "22:00"}}

	out, err := svc.UpdateEvento(ctx, uuid.New(), id, EventoPatch{
		TitleEs:          &title,
		TitleEn:          OptionalOf(titleEn),
		FrequencyEs:      OptionalOf(frequency),
		FrequencyEn:      OptionalOf(frequencyEn),
		DetailEs:         &detail,
		DetailEn:         OptionalOf(detailEn),
		PlaceEs:          &place,
		PlaceEn:          OptionalOf(placeEn),
		MapURL:           OptionalOf(mapURL),
		OrganizerEs:      OptionalOf(organizer),
		OrganizerEn:      OptionalOf(organizerEn),
		CategoryID:       OptionalOf(category.ID),
		ImageFile:        OptionalOf(file),
		ImageAltEs:       OptionalOf(altEs),
		ImageAltEn:       OptionalOf(altEn),
		PublicationState: &published,
		Dates:            &dates,
	})
	if err != nil {
		t.Fatalf("UpdateEvento(completo): %v", err)
	}
	if out.TitleEs != title || out.TitleEn == nil || *out.TitleEn != titleEn {
		t.Errorf("títulos = %+v", out)
	}
	if out.FrequencyEs == nil || *out.FrequencyEs != frequency || out.FrequencyEn == nil || *out.FrequencyEn != frequencyEn {
		t.Errorf("frecuencia = %+v", out)
	}
	if out.DetailEs != detail || out.DetailEn == nil || *out.DetailEn != detailEn {
		t.Errorf("detalle = %+v", out)
	}
	if out.PlaceEs != place || out.PlaceEn == nil || *out.PlaceEn != placeEn {
		t.Errorf("lugar = %+v", out)
	}
	if out.MapURL == nil || *out.MapURL != mapURL || out.OrganizerEn == nil || *out.OrganizerEn != organizerEn {
		t.Errorf("enlaces = %+v", out)
	}
	if out.CategoryID == nil || *out.CategoryID != category.ID {
		t.Errorf("categoría = %v", out.CategoryID)
	}
	if out.ImageFile == nil || *out.ImageFile != file || out.ImageAltEn == nil || *out.ImageAltEn != altEn {
		t.Errorf("imagen = %+v", out)
	}
	if len(out.Dates) != 1 || out.Dates[0].StartTime == nil || *out.Dates[0].StartTime != "19:00" {
		t.Errorf("fechas = %+v", out.Dates)
	}

	// PATCH que limpia los opcionales (valores vacíos y `null`).
	cleared, err := svc.UpdateEvento(ctx, uuid.New(), id, EventoPatch{
		FrequencyEs: OptionalOf("   "),
		FrequencyEn: nullOptional(),
		MapURL:      nullOptional(),
		OrganizerEs: OptionalOf(""),
		OrganizerEn: nullOptional(),
		CategoryID:  nullOptional(),
		ImageFile:   nullOptional(),
		ImageAltEs:  nullOptional(),
		ImageAltEn:  nullOptional(),
	})
	if err != nil {
		t.Fatalf("UpdateEvento(limpiar): %v", err)
	}
	if cleared.FrequencyEs != nil || cleared.FrequencyEn != nil || cleared.MapURL != nil {
		t.Errorf("los opcionales no se limpiaron: %+v", cleared)
	}
	if cleared.OrganizerEs != nil || cleared.OrganizerEn != nil || cleared.CategoryID != nil {
		t.Errorf("los opcionales no se limpiaron: %+v", cleared)
	}
	if cleared.ImageFile != nil || cleared.ImageAltEs != nil || cleared.ImageAltEn != nil {
		t.Errorf("la imagen no se quitó: %+v", cleared)
	}
	if cleared.TitleEs != title {
		t.Errorf("el título cambió sin pedirlo: %q", cleared.TitleEs)
	}
}

// TestUpdateActividadCoversEveryPatchField recorre todos los campos del PATCH de
// actividades (valores, fecha y horario, y `null` explícito).
func TestUpdateActividadCoversEveryPatchField(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	category, err := svc.CreateCategoria(ctx, uuid.New(), CategoriaInput{NameEs: "Predicación"})
	if err != nil {
		t.Fatalf("CreateCategoria: %v", err)
	}
	created, err := svc.CreateActividad(ctx, uuid.New(), validActividadInput())
	if err != nil {
		t.Fatalf("CreateActividad: %v", err)
	}
	id := uuid.MustParse(created.ID)

	title, detail, place, date := "Predicar en el parque", "Detalle nuevo", "Parque central", "2026-11-15"
	titleEn, detailEn, placeEn, mapURL, organizer := "Preach", "New detail", "Central park", "https://maps.example/parque", "Equipo"
	file, altEs, published := "img_22222222-2222-2222-2222-222222222222.png", "Foto", "published"

	out, err := svc.UpdateActividad(ctx, uuid.New(), id, ActividadPatch{
		TitleEs:          &title,
		TitleEn:          OptionalOf(titleEn),
		DetailEs:         &detail,
		DetailEn:         OptionalOf(detailEn),
		PlaceEs:          &place,
		PlaceEn:          OptionalOf(placeEn),
		MapURL:           OptionalOf(mapURL),
		OrganizerEs:      OptionalOf(organizer),
		CategoryID:       OptionalOf(category.ID),
		ImageFile:        OptionalOf(file),
		ImageAltEs:       OptionalOf(altEs),
		Date:             &date,
		StartTime:        OptionalOf("15:00"),
		EndTime:          OptionalOf("17:30"),
		PublicationState: &published,
	})
	if err != nil {
		t.Fatalf("UpdateActividad(completo): %v", err)
	}
	if out.Date != date || out.StartTime == nil || *out.StartTime != "15:00" || out.EndTime == nil || *out.EndTime != "17:30" {
		t.Errorf("fecha/horario = %+v", out)
	}
	if out.TitleEn == nil || out.DetailEn == nil || out.PlaceEn == nil {
		t.Errorf("traducciones = %+v", out)
	}
	if out.MapURL == nil || out.CategoryID == nil || out.ImageFile == nil || out.ImageAltEs == nil {
		t.Errorf("campos opcionales = %+v", out)
	}
	if out.PublicationState != "published" {
		t.Errorf("estado = %q", out.PublicationState)
	}

	// Quitar la hora de fin y el resto de los opcionales.
	cleared, err := svc.UpdateActividad(ctx, uuid.New(), id, ActividadPatch{
		EndTime:     nullOptional(),
		MapURL:      nullOptional(),
		OrganizerEs: OptionalOf("  "),
		CategoryID:  nullOptional(),
		ImageFile:   nullOptional(),
		ImageAltEs:  nullOptional(),
		TitleEn:     nullOptional(),
		DetailEn:    nullOptional(),
		PlaceEn:     nullOptional(),
	})
	if err != nil {
		t.Fatalf("UpdateActividad(limpiar): %v", err)
	}
	if cleared.EndTime != nil || cleared.MapURL != nil || cleared.OrganizerEs != nil {
		t.Errorf("los opcionales no se limpiaron: %+v", cleared)
	}
	if cleared.CategoryID != nil || cleared.ImageFile != nil || cleared.ImageAltEs != nil {
		t.Errorf("los opcionales no se limpiaron: %+v", cleared)
	}
	if cleared.TitleEn != nil || cleared.DetailEn != nil || cleared.PlaceEn != nil {
		t.Errorf("las traducciones no se limpiaron: %+v", cleared)
	}
	if cleared.StartTime == nil || *cleared.StartTime != "15:00" {
		t.Errorf("el inicio no debía tocarse: %v", cleared.StartTime)
	}
}

// TestUpdateCategoriaCoversEveryPatchField recorre los campos del PATCH del
// catálogo (nombre en inglés con valor y con `null`).
func TestUpdateCategoriaCoversEveryPatchField(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	created, err := svc.CreateCategoria(ctx, uuid.New(), CategoriaInput{NameEs: "Otro", SortOrder: 999})
	if err != nil {
		t.Fatalf("CreateCategoria: %v", err)
	}
	id := uuid.MustParse(created.ID)

	withEnglish, err := svc.UpdateCategoria(ctx, uuid.New(), id, CategoriaPatch{NameEn: OptionalOf("Other")})
	if err != nil {
		t.Fatalf("UpdateCategoria(nameEn): %v", err)
	}
	if withEnglish.NameEn == nil || *withEnglish.NameEn != "Other" {
		t.Errorf("nameEn = %v", withEnglish.NameEn)
	}

	cleared, err := svc.UpdateCategoria(ctx, uuid.New(), id, CategoriaPatch{NameEn: nullOptional()})
	if err != nil {
		t.Fatalf("UpdateCategoria(nameEn null): %v", err)
	}
	if cleared.NameEn != nil {
		t.Errorf("nameEn = %v, se esperaba null", cleared.NameEn)
	}
}

// TestListEventosReadsDatesInOneBatchQuery fija el hallazgo I1 en el panel: el
// listado de eventos lee las fechas de TODA la página en UNA sola consulta (por
// lote), sin crecer con el número de filas.
func TestListEventosReadsDatesInOneBatchQuery(t *testing.T) {
	repo := newFakeRepository()
	svc := newAdminService(t, repo)
	ctx := context.Background()

	const events = 12
	for i := 0; i < events; i++ {
		in := validEventoInput()
		in.TitleEs = fmt.Sprintf("Evento %02d", i)
		in.Dates = []FechaInput{
			{Date: "2026-10-01"}, // pasada
			{Date: "2026-10-12"}, // próxima
		}
		if _, err := svc.CreateEvento(ctx, uuid.New(), in); err != nil {
			t.Fatalf("CreateEvento: %v", err)
		}
	}
	repo.datesBatchCalls = 0

	list, err := svc.ListEventos(ctx, nil, paginate.Params{Limit: 100})
	if err != nil {
		t.Fatalf("ListEventos: %v", err)
	}
	if list.Total != events || len(list.Items) != events {
		t.Fatalf("listado = %+v", list)
	}
	if repo.datesBatchCalls != 1 {
		t.Errorf("consultas de fechas por lote = %d, se esperaba 1 (sin N+1)", repo.datesBatchCalls)
	}
	// Cada fila conserva sus fechas (ascendentes) y su `nextDate`.
	for _, item := range list.Items {
		if len(item.Dates) != 2 || item.Dates[0].Date != "2026-10-01" || item.Dates[1].Date != "2026-10-12" {
			t.Fatalf("fechas de %q = %+v", item.TitleEs, item.Dates)
		}
		if item.NextDate == nil || *item.NextDate != "2026-10-12" || item.IsPast {
			t.Fatalf("nextDate/isPast de %q = %v/%v", item.TitleEs, item.NextDate, item.IsPast)
		}
	}

	// Una página vacía no consulta el lote.
	repo.datesBatchCalls = 0
	if _, err := svc.ListEventos(ctx, nil, paginate.Params{Limit: 10, Offset: events}); err != nil {
		t.Fatalf("ListEventos(página vacía): %v", err)
	}
	if repo.datesBatchCalls != 0 {
		t.Errorf("una página sin filas no debe consultar el lote: %d llamadas", repo.datesBatchCalls)
	}
}
