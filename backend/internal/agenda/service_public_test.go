package agenda

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas del service público (T418): «lo que viene» solo con contenido
// publicado y próximo, resuelto a `lang` con fallback por campo, ordenado
// cronológicamente, con las claves de bloque omitidas si están vacías y acotado
// a `MaxBlockItems`.

// testToday/testNowTime son el corte fijo de las pruebas (el reloj real nunca se
// usa): coinciden con el reloj inyectado por newTestService (2026-10-11 12:00 en
// Costa Rica).
const (
	publicToday   = "2026-10-11"
	publicNowTime = "12:00"
)

// samplePublicEvent arma un evento publicado con una fecha.
func samplePublicEvent(title, date string) Event {
	return Event{
		ID:               uuid.New(),
		TitleEs:          title,
		DetailEs:         "Detalle " + title,
		PlaceEs:          "Templo",
		PublicationState: StatePublished,
	}
}

// samplePublicActivity arma una actividad publicada con su fecha.
func samplePublicActivity(title, date string) Activity {
	return Activity{
		ID:               uuid.New(),
		TitleEs:          title,
		DetailEs:         "Detalle " + title,
		PlaceEs:          "Salón",
		Date:             date,
		PublicationState: StatePublished,
	}
}

// newPublicService construye el service con el reloj fijo de las pruebas.
func newPublicService(t *testing.T, repo *fakeRepository) *service {
	t.Helper()
	location, err := time.LoadLocation("America/Costa_Rica")
	if err != nil {
		t.Fatalf("cargar America/Costa_Rica: %v", err)
	}
	logger, _ := testutil.NewLogger()
	return NewService(ServiceDeps{
		Repository: repo,
		Location:   location,
		// 2026-10-11 18:00 UTC = 12:00 en Costa Rica (UTC−06:00).
		Now:    func() time.Time { return time.Date(2026, time.October, 11, 18, 0, 0, 0, time.UTC) },
		Logger: logger,
	})
}

// TestGetProximamenteOnlyReturnsPublishedAndUpcoming fija FR-002/FR-013: con
// todo en borrador o vencido, la respuesta solo lleva `lang` (sin claves de
// bloque, SC-012).
func TestGetProximamenteOnlyReturnsPublishedAndUpcoming(t *testing.T) {
	earlyEnd := "10:00"

	draft := samplePublicEvent("Borrador", "2026-10-12")
	draft.PublicationState = StateDraft
	expired := samplePublicEvent("Vencido", "2026-10-10")
	todayEnded := samplePublicEvent("Hoy terminado", publicToday)
	draftActivity := samplePublicActivity("Actividad en borrador", "2026-10-12")
	draftActivity.PublicationState = StateDraft

	repo := &fakeRepository{
		events: []Event{draft, expired, todayEnded},
		eventDates: map[uuid.UUID][]EventDate{
			expired.ID:    {{Date: "2026-10-10"}},
			todayEnded.ID: {{Date: publicToday, EndTime: &earlyEnd}},
		},
		activities: []Activity{draftActivity},
	}
	svc := newPublicService(t, repo)

	out, err := svc.GetProximamente(context.Background(), "es")
	if err != nil {
		t.Fatalf("GetProximamente: %v", err)
	}
	if out.Lang != LangES {
		t.Errorf("lang = %q", out.Lang)
	}
	if len(out.Events) != 0 || len(out.Activities) != 0 {
		t.Fatalf("la respuesta no debería traer bloques: %+v", out)
	}

	body, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(body), "events") || strings.Contains(string(body), "activities") {
		t.Errorf("las claves vacías deben omitirse (SC-012): %s", body)
	}
}

// TestGetProximamenteAppliesTheNextRule comprueba los tres casos de R4-2 tal y
// como los ve el visitante (fechas fijas).
func TestGetProximamenteAppliesTheNextRule(t *testing.T) {
	earlyEnd := "10:00"
	lateEnd := "22:00"

	future := samplePublicEvent("Futuro", "2026-10-12")
	todayOpen := samplePublicEvent("Hoy sin fin", publicToday)
	todayEnded := samplePublicEvent("Hoy terminado", publicToday)
	todayRunning := samplePublicEvent("Hoy en curso", publicToday)
	past := samplePublicEvent("Pasado", "2026-10-10")

	repo := &fakeRepository{
		events: []Event{future, todayOpen, todayEnded, todayRunning, past},
		eventDates: map[uuid.UUID][]EventDate{
			future.ID:       {{Date: "2026-10-12"}},
			todayOpen.ID:    {{Date: publicToday}},
			todayEnded.ID:   {{Date: publicToday, EndTime: &earlyEnd}},
			todayRunning.ID: {{Date: publicToday, EndTime: &lateEnd}},
			past.ID:         {{Date: "2026-10-10"}},
		},
	}
	svc := newPublicService(t, repo)

	out, err := svc.GetProximamente(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetProximamente: %v", err)
	}
	titles := make([]string, 0, len(out.Events))
	for _, event := range out.Events {
		titles = append(titles, event.Title)
	}
	want := []string{"Hoy en curso", "Hoy sin fin", "Futuro"}
	if len(titles) != len(want) {
		t.Fatalf("eventos = %v, se esperaban %v", titles, want)
	}
	for i, title := range want {
		if titles[i] != title {
			t.Errorf("eventos[%d] = %q, se esperaba %q", i, titles[i], title)
		}
	}

	// El corte viaja parametrizado a la consulta pública.
	if len(repo.cutoffs) == 0 {
		t.Fatal("no se consultó el listado público")
	}
	for _, call := range repo.cutoffs {
		if call.today != publicToday || call.nowTime != publicNowTime {
			t.Errorf("corte = %s %s, se esperaba %s %s", call.today, call.nowTime, publicToday, publicNowTime)
		}
	}
}

// TestGetProximamenteDatesAreOnlyUpcoming comprueba que las fechas del evento
// público son solo las próximas y que la primera es "la próxima fecha" (FR-004).
func TestGetProximamenteDatesAreOnlyUpcoming(t *testing.T) {
	event := samplePublicEvent("Con varias fechas", publicToday)
	repo := &fakeRepository{
		events: []Event{event},
		eventDates: map[uuid.UUID][]EventDate{
			event.ID: {
				{Date: "2026-10-10"}, // pasada: fuera
				{Date: publicToday},  // hoy sin fin: dentro
				{Date: "2026-10-12", StartTime: ptrOf("19:00"), EndTime: ptrOf("21:00")},
			},
		},
	}
	svc := newPublicService(t, repo)

	out, err := svc.GetProximamente(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetProximamente: %v", err)
	}
	if len(out.Events) != 1 {
		t.Fatalf("eventos = %d, se esperaba 1", len(out.Events))
	}
	dates := out.Events[0].Dates
	if len(dates) != 2 {
		t.Fatalf("fechas públicas = %+v, se esperaban 2 (sin la pasada)", dates)
	}
	if dates[0].Date != publicToday || dates[1].Date != "2026-10-12" {
		t.Errorf("fechas = %s/%s", dates[0].Date, dates[1].Date)
	}
	if dates[1].StartTime == nil || *dates[1].StartTime != "19:00" {
		t.Errorf("horario = %+v", dates[1])
	}
}

// TestGetProximamenteFallsBackPerField fija FR-008/SC-006: cada campo traducible
// se resuelve con fallback `en → es` y nunca sale vacío por falta de traducción.
func TestGetProximamenteFallsBackPerField(t *testing.T) {
	event := samplePublicEvent("Noche de aposentos", "2026-10-12")
	event.TitleEn = ptrOf("Dormitory night")
	event.DetailEn = nil // sin traducción: cae al español
	event.PlaceEn = ptrOf("Temple")
	event.OrganizerEs = ptrOf("Equipo de jóvenes")
	event.FrequencyEs = ptrOf("Todos los viernes")

	activity := samplePublicActivity("Repartir comida", "2026-10-12")
	activity.TitleEn = ptrOf("Food distribution")

	repo := &fakeRepository{
		events:     []Event{event},
		eventDates: map[uuid.UUID][]EventDate{event.ID: {{Date: "2026-10-12"}}},
		activities: []Activity{activity},
	}
	svc := newPublicService(t, repo)

	out, err := svc.GetProximamente(context.Background(), LangEN)
	if err != nil {
		t.Fatalf("GetProximamente: %v", err)
	}
	got := out.Events[0]
	if got.Title != "Dormitory night" {
		t.Errorf("title = %q, se esperaba la traducción", got.Title)
	}
	if got.Detail != "Detalle Noche de aposentos" {
		t.Errorf("detail = %q, se esperaba el español por fallback", got.Detail)
	}
	if got.Place != "Temple" {
		t.Errorf("place = %q", got.Place)
	}
	if got.Organizer == nil || *got.Organizer != "Equipo de jóvenes" {
		t.Errorf("organizer = %v, se esperaba el español", got.Organizer)
	}
	if got.Frequency == nil || *got.Frequency != "Todos los viernes" {
		t.Errorf("frequency = %v, se esperaba el español", got.Frequency)
	}
	if out.Activities[0].Title != "Food distribution" {
		t.Errorf("título de la actividad = %q", out.Activities[0].Title)
	}

	// En español, todo sale en español.
	out, err = svc.GetProximamente(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetProximamente(es): %v", err)
	}
	if out.Events[0].Title != "Noche de aposentos" {
		t.Errorf("title(es) = %q", out.Events[0].Title)
	}
}

// TestGetProximamenteOrdersChronologicallyWithStableTieBreak fija FR-003 y el
// edge case de dos elementos en la misma fecha: orden por la próxima fecha y
// desempate estable (`title`, `id`).
func TestGetProximamenteOrdersChronologicallyWithStableTieBreak(t *testing.T) {
	// El repositorio devuelve el orden de la consulta; el service lo reafirma.
	later := samplePublicEvent("Zeta", "2026-10-19")
	tieB := samplePublicEvent("Beta", "2026-10-12")
	tieA := samplePublicEvent("Alfa", "2026-10-12")
	todayEvent := samplePublicEvent("Omega", publicToday)

	repo := &fakeRepository{
		events: []Event{later, tieB, tieA, todayEvent},
		eventDates: map[uuid.UUID][]EventDate{
			later.ID:      {{Date: "2026-10-19"}},
			tieA.ID:       {{Date: "2026-10-12"}},
			tieB.ID:       {{Date: "2026-10-12"}},
			todayEvent.ID: {{Date: publicToday}},
		},
	}
	// Actividades: mismo día, una sin hora (primero) y otra con hora.
	activityLate := samplePublicActivity("Con hora", "2026-10-12")
	activityLate.StartTime = ptrOf("19:00")
	activityOpen := samplePublicActivity("Sin hora", "2026-10-12")
	repo.activities = []Activity{activityLate, activityOpen}

	svc := newPublicService(t, repo)

	out, err := svc.GetProximamente(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetProximamente: %v", err)
	}

	eventTitles := []string{}
	for _, event := range out.Events {
		eventTitles = append(eventTitles, event.Title)
	}
	want := []string{"Omega", "Alfa", "Beta", "Zeta"}
	if strings.Join(eventTitles, ",") != strings.Join(want, ",") {
		t.Errorf("eventos = %v, se esperaban %v", eventTitles, want)
	}

	activityTitles := []string{}
	for _, activity := range out.Activities {
		activityTitles = append(activityTitles, activity.Title)
	}
	if strings.Join(activityTitles, ",") != "Sin hora,Con hora" {
		t.Errorf("actividades = %v, se esperaba la de hoy sin hora primero", activityTitles)
	}
}

// TestGetProximamenteResolvesCategoryAndImage comprueba que la categoría y la
// imagen (con su alt) se resuelven en el servidor (Q6/FR-020).
func TestGetProximamenteResolvesCategoryAndImage(t *testing.T) {
	category := Category{ID: uuid.New(), NameEs: "Acción social", NameEn: ptrOf("Social action")}

	event := samplePublicEvent("Con categoría", "2026-10-12")
	event.CategoryID = &category.ID
	event.ImageFile = ptrOf("img_11111111-1111-1111-1111-111111111111.jpg")
	event.ImageAltEs = ptrOf("Foto del evento")

	activity := samplePublicActivity("Sin categoría", "2026-10-12")

	repo := &fakeRepository{
		events:     []Event{event},
		eventDates: map[uuid.UUID][]EventDate{event.ID: {{Date: "2026-10-12"}}},
		activities: []Activity{activity},
		categories: []Category{category},
	}
	svc := newPublicService(t, repo)

	out, err := svc.GetProximamente(context.Background(), LangEN)
	if err != nil {
		t.Fatalf("GetProximamente: %v", err)
	}
	got := out.Events[0]
	if got.Category == nil || got.Category.ID != category.ID.String() || got.Category.Name != "Social action" {
		t.Errorf("categoría = %+v, se esperaba la resuelta y localizada", got.Category)
	}
	if got.Image == nil || got.Image.URL != MediaPathPrefix+"img_11111111-1111-1111-1111-111111111111.jpg" {
		t.Errorf("imagen = %+v", got.Image)
	}
	if got.Image.Alt != "Foto del evento" {
		t.Errorf("alt = %q (fallback al español)", got.Image.Alt)
	}
	if out.Activities[0].Category != nil {
		t.Errorf("la actividad sin categoría no debería traerla: %+v", out.Activities[0].Category)
	}
}

// TestGetProximamenteCapsBlocksAtMaxBlockItems fija RG4-8: cada bloque se acota
// a 200 elementos (los primeros en orden) y el service pasa ese tope a la
// consulta.
func TestGetProximamenteCapsBlocksAtMaxBlockItems(t *testing.T) {
	repo := &fakeRepository{eventDates: map[uuid.UUID][]EventDate{}}
	for i := 0; i < MaxBlockItems+25; i++ {
		event := samplePublicEvent(fmt.Sprintf("Evento %03d", i), "2026-10-12")
		repo.events = append(repo.events, event)
		repo.eventDates[event.ID] = []EventDate{{Date: "2026-10-12"}}
	}
	svc := newPublicService(t, repo)

	out, err := svc.GetProximamente(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetProximamente: %v", err)
	}
	if len(out.Events) != MaxBlockItems {
		t.Fatalf("eventos = %d, se esperaban %d", len(out.Events), MaxBlockItems)
	}
	if len(repo.blockLimits) == 0 || repo.blockLimits[0] != MaxBlockItems {
		t.Errorf("el tope pasado a la consulta = %v, se esperaba %d", repo.blockLimits, MaxBlockItems)
	}
}

// TestGetProximamenteRejectsUnknownLang fija que un idioma fuera de `es|en` es
// un 400 invalid con details.lang.
func TestGetProximamenteRejectsUnknownLang(t *testing.T) {
	svc := newPublicService(t, &fakeRepository{})

	_, err := svc.GetProximamente(context.Background(), "fr")
	if errorKind(err) != apperr.KindInvalid {
		t.Fatalf("lang=fr = %v, se esperaba invalid", err)
	}
	if _, ok := detailsOf(t, err)["lang"]; !ok {
		t.Errorf("sin details.lang: %v", err)
	}

	// Sin idioma (o vacío) se usa el base.
	out, err := svc.GetProximamente(context.Background(), "  ")
	if err != nil || out.Lang != LangES {
		t.Fatalf("lang vacío = %q (%v), se esperaba es", out.Lang, err)
	}
}

// TestGetProximamenteNeverExposesPublicationState fija FR-013: el DTO público no
// lleva el estado de publicación (ninguna vía pública a borradores).
func TestGetProximamenteNeverExposesPublicationState(t *testing.T) {
	event := samplePublicEvent("Evento", "2026-10-12")
	activity := samplePublicActivity("Actividad", "2026-10-12")
	repo := &fakeRepository{
		events:     []Event{event},
		eventDates: map[uuid.UUID][]EventDate{event.ID: {{Date: "2026-10-12"}}},
		activities: []Activity{activity},
	}
	svc := newPublicService(t, repo)

	out, err := svc.GetProximamente(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetProximamente: %v", err)
	}
	body, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(body), "publicationState") || strings.Contains(string(body), "draft") {
		t.Errorf("el DTO público expone el estado de publicación: %s", body)
	}
}

// ptrOf devuelve un puntero al valor indicado.
func ptrOf(value string) *string { return &value }

// TestGetProximamenteReadsDatesInOneBatchQuery fija el hallazgo I1: la vista
// pública lee las fechas de TODOS los eventos del bloque en UNA sola consulta
// (por lote), sin crecer con el número de eventos. El repositorio falso cuenta
// las llamadas al lote.
func TestGetProximamenteReadsDatesInOneBatchQuery(t *testing.T) {
	repo := &fakeRepository{eventDates: map[uuid.UUID][]EventDate{}}
	const events = 25
	for i := 0; i < events; i++ {
		event := samplePublicEvent(fmt.Sprintf("Evento %02d", i), "2026-10-12")
		repo.events = append(repo.events, event)
		repo.eventDates[event.ID] = []EventDate{
			{Date: "2026-10-10"},                            // pasada: fuera del bloque público
			{Date: "2026-10-12"},                            // próxima
			{Date: "2026-10-19", StartTime: ptrOf("19:00")}, // segunda próxima
		}
	}
	svc := newPublicService(t, repo)

	out, err := svc.GetProximamente(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetProximamente: %v", err)
	}
	if len(out.Events) != events {
		t.Fatalf("eventos públicos = %d, se esperaban %d", len(out.Events), events)
	}
	if repo.upcomingBatchCalls != 1 {
		t.Errorf("consultas de fechas por lote = %d, se esperaba 1 (sin N+1)", repo.upcomingBatchCalls)
	}
	// Cada evento trae solo sus fechas próximas, en orden.
	for _, event := range out.Events {
		if len(event.Dates) != 2 || event.Dates[0].Date != "2026-10-12" || event.Dates[1].Date != "2026-10-19" {
			t.Fatalf("fechas de %q = %+v, se esperaban las 2 próximas", event.Title, event.Dates)
		}
	}

	// Sin eventos no se consulta la base (ni se inventa una llamada).
	emptyRepo := newFakeRepository()
	emptySvc := newPublicService(t, emptyRepo)
	if _, err := emptySvc.GetProximamente(context.Background(), LangES); err != nil {
		t.Fatalf("GetProximamente(sin contenido): %v", err)
	}
	if emptyRepo.upcomingBatchCalls != 0 {
		t.Errorf("sin eventos no debe consultarse el lote: %d llamadas", emptyRepo.upcomingBatchCalls)
	}
}
