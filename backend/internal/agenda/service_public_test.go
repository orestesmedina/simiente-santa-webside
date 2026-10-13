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

// --- Extracto «Lo que viene» de la portada (US7/FR-025, T445) ---

// TestPortadaProximamenteConstants fija que el tope 5 y la ventana de 2 meses son
// decisiones firmes (P4-18/R4-16): no son parámetros ni crecen con el contenido.
func TestPortadaProximamenteConstants(t *testing.T) {
	if maxPortadaItems != 5 {
		t.Errorf("maxPortadaItems = %d, se esperaban 5 (FR-025)", maxPortadaItems)
	}
	if portadaWindowMonths != 2 {
		t.Errorf("portadaWindowMonths = %d, se esperaban 2 (FR-025)", portadaWindowMonths)
	}
}

// TestPortadaWindowEndAddsCalendarMonths fija la aritmética de la ventana: meses
// de CALENDARIO (no 60 días).
func TestPortadaWindowEndAddsCalendarMonths(t *testing.T) {
	tests := []struct {
		today string
		want  string
	}{
		{today: "2026-10-12", want: "2026-12-12"},
		{today: "2026-01-31", want: "2026-03-31"},
		{today: "2026-11-30", want: "2027-01-30"},
		{today: "2026-10-11", want: "2026-12-11"},
	}
	for _, tt := range tests {
		got, err := portadaWindowEnd(tt.today, portadaWindowMonths)
		if err != nil {
			t.Fatalf("portadaWindowEnd(%q): %v", tt.today, err)
		}
		if got != tt.want {
			t.Errorf("portadaWindowEnd(%q) = %q, se esperaba %q", tt.today, got, tt.want)
		}
	}

	// Una fecha de corte imposible es un error interno (nunca la produce cutoff).
	if _, err := portadaWindowEnd("no-es-fecha", portadaWindowMonths); errorKind(err) != apperr.KindInternal {
		t.Errorf("portadaWindowEnd(inválida) = %v, se esperaba internal", err)
	}
}

// seedPortadaItems llena el repositorio falso con eventos y actividades
// publicados (o en el estado indicado) con sus fechas.
func seedPortadaItems(repo *fakeRepository, events map[string]string, activities map[string]string) {
	for title, day := range events {
		event := samplePublicEvent(title, day)
		repo.events = append(repo.events, event)
		repo.eventDates[event.ID] = []EventDate{{Date: day}}
	}
	for title, day := range activities {
		repo.activities = append(repo.activities, samplePublicActivity(title, day))
	}
}

// TestGetPortadaProximamenteFixesLimitAndWindow fija el tope y el corte: con 8
// ítems próximos salen exactamente los 5 más cercanos mezclados (eventos y
// actividades alternándose por fecha) y la consulta recibe el tope 5, el corte en
// APP_TZ y la ventana de 2 meses.
func TestGetPortadaProximamenteFixesLimitAndWindow(t *testing.T) {
	repo := newFakeRepository()
	// Fechas alternando concepto: evento el 12, actividad el 13, evento el 14…
	for i := 0; i < 4; i++ {
		day := fmt.Sprintf("2026-10-%02d", 12+i*2)
		event := samplePublicEvent(fmt.Sprintf("Evento %d", i), day)
		repo.events = append(repo.events, event)
		repo.eventDates[event.ID] = []EventDate{{Date: day}}

		activityDay := fmt.Sprintf("2026-10-%02d", 13+i*2)
		repo.activities = append(repo.activities, samplePublicActivity(fmt.Sprintf("Actividad %d", i), activityDay))
	}
	svc := newPublicService(t, repo)

	out, err := svc.GetPortadaProximamente(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetPortadaProximamente: %v", err)
	}
	if len(out.Items) != maxPortadaItems {
		t.Fatalf("items = %d, se esperaban %d", len(out.Items), maxPortadaItems)
	}
	wantKinds := []string{"event", "activity", "event", "activity", "event"}
	wantDates := []string{"2026-10-12", "2026-10-13", "2026-10-14", "2026-10-15", "2026-10-16"}
	for i, item := range out.Items {
		if item.Kind != wantKinds[i] || item.Date != wantDates[i] {
			t.Errorf("items[%d] = %s %s, se esperaba %s %s", i, item.Kind, item.Date, wantKinds[i], wantDates[i])
		}
	}

	if len(repo.nextItemsCalls) != 1 {
		t.Fatalf("llamadas al extracto = %d, se esperaba 1", len(repo.nextItemsCalls))
	}
	call := repo.nextItemsCalls[0]
	if call.limit != maxPortadaItems {
		t.Errorf("limit = %d, se esperaban %d", call.limit, maxPortadaItems)
	}
	if call.today != publicToday || call.nowTime != publicNowTime {
		t.Errorf("corte = %s %s, se esperaba %s %s", call.today, call.nowTime, publicToday, publicNowTime)
	}
	if call.windowEnd != "2026-12-11" {
		t.Errorf("windowEnd = %q, se esperaba 2026-12-11 (2 meses de calendario)", call.windowEnd)
	}
}

// TestGetPortadaProximamenteWindowWithBackup fija la regla de la ventana: si
// tiene ítems, solo salen los de dentro (los lejanos esperan); si está vacía,
// entran los siguientes próximos; nunca se mezclan.
func TestGetPortadaProximamenteWindowWithBackup(t *testing.T) {
	t.Run("la ventana manda y los lejanos no se mezclan", func(t *testing.T) {
		repo := newFakeRepository()
		seedPortadaItems(repo,
			map[string]string{"Evento cerca": "2026-11-01", "Evento lejos": "2027-06-01"},
			map[string]string{"Actividad lejos": "2027-03-01"})
		svc := newPublicService(t, repo)

		out, err := svc.GetPortadaProximamente(context.Background(), LangES)
		if err != nil {
			t.Fatalf("GetPortadaProximamente: %v", err)
		}
		if len(out.Items) != 1 || out.Items[0].Title != "Evento cerca" {
			t.Fatalf("items = %+v, se esperaba solo el de la ventana", out.Items)
		}
	})

	t.Run("sin nada en la ventana entra el respaldo", func(t *testing.T) {
		repo := newFakeRepository()
		seedPortadaItems(repo,
			map[string]string{"Evento lejano": "2027-06-01"},
			map[string]string{"Actividad lejana": "2027-03-01"})
		svc := newPublicService(t, repo)

		out, err := svc.GetPortadaProximamente(context.Background(), LangES)
		if err != nil {
			t.Fatalf("GetPortadaProximamente: %v", err)
		}
		if len(out.Items) != 2 {
			t.Fatalf("items = %+v, se esperaban los 2 del respaldo", out.Items)
		}
		// Orden por cercanía: la actividad de 2027-03-01 va antes que el evento de 2027-06-01.
		if out.Items[0].Title != "Actividad lejana" || out.Items[1].Title != "Evento lejano" {
			t.Errorf("items = %+v, se esperaba el orden por fecha", out.Items)
		}
	})

	t.Run("el límite de la ventana es la fecha de cierre", func(t *testing.T) {
		repo := newFakeRepository()
		// El corte es 2026-10-11 y la ventana cierra el 2026-12-11: el mismo día
		// entra (date ≤ window_end) y el siguiente no.
		seedPortadaItems(repo,
			map[string]string{"Último día de la ventana": "2026-12-11", "Un día después": "2026-12-12"},
			nil)
		svc := newPublicService(t, repo)

		out, err := svc.GetPortadaProximamente(context.Background(), LangES)
		if err != nil {
			t.Fatalf("GetPortadaProximamente: %v", err)
		}
		if len(out.Items) != 1 || out.Items[0].Title != "Último día de la ventana" {
			t.Fatalf("items = %+v, se esperaba solo el del último día de la ventana", out.Items)
		}
	})
}

// TestGetPortadaProximamenteWithoutItems fija FR-026: sin nada publicado y
// próximo la respuesta lleva `lang` y la clave `items` se omite (nunca null).
func TestGetPortadaProximamenteWithoutItems(t *testing.T) {
	repo := newFakeRepository()
	draft := samplePublicEvent("Borrador", "2026-10-12")
	draft.PublicationState = StateDraft
	repo.events = append(repo.events, draft)
	repo.eventDates[draft.ID] = []EventDate{{Date: "2026-10-12"}}
	svc := newPublicService(t, repo)

	out, err := svc.GetPortadaProximamente(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetPortadaProximamente: %v", err)
	}
	if len(out.Items) != 0 {
		t.Fatalf("items = %+v, se esperaba vacío", out.Items)
	}
	body, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := string(body); got != `{"lang":"es"}` {
		t.Errorf("cuerpo = %s, se esperaba solo lang (clave items omitida)", got)
	}
}

// TestGetPortadaProximamenteFiltersDraftsExpiredAndWithoutNextDate fija que solo
// entra lo publicado y próximo: los tres casos de R4-2 (futuro y «hoy en curso»
// dentro; «hoy terminado» fuera), los borradores fuera y un evento sin ninguna
// fecha próxima fuera.
func TestGetPortadaProximamenteFiltersDraftsExpiredAndWithoutNextDate(t *testing.T) {
	// Horarios del corte de las pruebas (2026-10-11 a las 12:00): un rango ya
	// terminado y otro en curso.
	start := "09:00"
	earlyEnd := "10:00"
	runningStart := "11:00"
	lateEnd := "22:00"

	repo := newFakeRepository()
	draft := samplePublicEvent("Borrador", "2026-10-12")
	draft.PublicationState = StateDraft
	expired := samplePublicEvent("Vencido", "2026-10-01")
	noNext := samplePublicEvent("Sin fecha próxima", "2026-10-01")
	todayEnded := samplePublicEvent("Hoy terminado", publicToday)
	todayRunning := samplePublicEvent("Hoy en curso", publicToday)
	future := samplePublicEvent("Futuro", "2026-10-12")

	repo.events = append(repo.events, draft, expired, noNext, todayEnded, todayRunning, future)
	repo.eventDates[draft.ID] = []EventDate{{Date: "2026-10-12"}}
	repo.eventDates[expired.ID] = []EventDate{{Date: "2026-10-01"}}
	repo.eventDates[noNext.ID] = []EventDate{{Date: "2026-10-01"}}
	repo.eventDates[todayEnded.ID] = []EventDate{{Date: publicToday, StartTime: &start, EndTime: &earlyEnd}}
	repo.eventDates[todayRunning.ID] = []EventDate{{Date: publicToday, StartTime: &runningStart, EndTime: &lateEnd}}
	repo.eventDates[future.ID] = []EventDate{{Date: "2026-10-12"}}

	expiredActivity := samplePublicActivity("Actividad vencida", "2026-10-01")
	todayActivity := samplePublicActivity("Actividad de hoy", publicToday)
	repo.activities = append(repo.activities, expiredActivity, todayActivity)

	svc := newPublicService(t, repo)

	out, err := svc.GetPortadaProximamente(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetPortadaProximamente: %v", err)
	}
	titles := make([]string, 0, len(out.Items))
	for _, item := range out.Items {
		titles = append(titles, item.Title)
	}
	want := []string{"Actividad de hoy", "Hoy en curso", "Futuro"}
	if len(titles) != len(want) {
		t.Fatalf("items = %v, se esperaban %v", titles, want)
	}
	for i, title := range want {
		if titles[i] != title {
			t.Errorf("items[%d] = %q, se esperaba %q", i, titles[i], title)
		}
	}
}

// TestGetPortadaProximamenteLocalizesAndResolvesImage fija el fallback por campo
// (FR-008) y la imagen con su alt (FR-020): sin imagen el ítem viene completo.
func TestGetPortadaProximamenteLocalizesAndResolvesImage(t *testing.T) {
	repo := newFakeRepository()
	event := samplePublicEvent("Noche de aposentos", "2026-10-12")
	event.TitleEn = ptrOf("Dormitory night")
	event.PlaceEn = ptrOf("Temple")
	event.DetailEn = ptrOf("Detalle en inglés") // el detalle no viaja al extracto
	event.ImageFile = ptrOf("img_11111111-1111-1111-1111-111111111111.jpg")
	event.ImageAltEs = ptrOf("Foto del evento")
	repo.events = append(repo.events, event)
	repo.eventDates[event.ID] = []EventDate{{Date: "2026-10-12", StartTime: ptrOf("19:00"), EndTime: ptrOf("21:00")}}

	activity := samplePublicActivity("Repartir comida", "2026-10-12")
	repo.activities = append(repo.activities, activity)

	svc := newPublicService(t, repo)

	out, err := svc.GetPortadaProximamente(context.Background(), LangEN)
	if err != nil {
		t.Fatalf("GetPortadaProximamente: %v", err)
	}
	if len(out.Items) != 2 {
		t.Fatalf("items = %+v", out.Items)
	}
	// Misma fecha: la actividad (sin hora de inicio) va primero
	// (`start_time NULLS FIRST`) y el evento después, con su horario.
	var got, other *ItemProximoPortada
	for i := range out.Items {
		switch out.Items[i].Kind {
		case "event":
			got = &out.Items[i]
		case "activity":
			other = &out.Items[i]
		}
	}
	if got == nil || other == nil {
		t.Fatalf("items = %+v, se esperaban un evento y una actividad", out.Items)
	}
	if out.Items[0].Kind != "activity" {
		t.Errorf("el primer ítem = %q, se esperaba la actividad (sin hora, NULLS FIRST)", out.Items[0].Kind)
	}
	if got.Title != "Dormitory night" || got.Place != "Temple" {
		t.Errorf("ítem localizado = %+v", *got)
	}
	if got.StartTime == nil || *got.StartTime != "19:00" || got.EndTime == nil || *got.EndTime != "21:00" {
		t.Errorf("horario = %+v", *got)
	}
	if got.Image == nil || got.Image.URL != MediaPathPrefix+"img_11111111-1111-1111-1111-111111111111.jpg" {
		t.Fatalf("imagen = %+v, se esperaba la del evento", got.Image)
	}
	if got.Image.Alt != "Foto del evento" {
		t.Errorf("alt = %q, se esperaba el español por fallback", got.Image.Alt)
	}
	// La actividad, sin imagen ni traducciones, sale en español y sin imagen.
	if other.Title != "Repartir comida" || other.Image != nil {
		t.Errorf("actividad = %+v", *other)
	}

	// El extracto NO lleva detalle, categoría ni lista de fechas (FR-025).
	body, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, forbidden := range []string{"detail", "category", "dates", "publicationState"} {
		if strings.Contains(string(body), forbidden) {
			t.Errorf("el extracto no debe llevar %q: %s", forbidden, body)
		}
	}
}

// TestGetPortadaProximamenteRejectsUnknownLang fija el 400 del idioma y el
// idioma base por defecto.
func TestGetPortadaProximamenteRejectsUnknownLang(t *testing.T) {
	svc := newPublicService(t, newFakeRepository())

	_, err := svc.GetPortadaProximamente(context.Background(), "fr")
	if errorKind(err) != apperr.KindInvalid {
		t.Fatalf("lang=fr = %v, se esperaba invalid", err)
	}
	if _, ok := detailsOf(t, err)["lang"]; !ok {
		t.Errorf("sin details.lang: %v", err)
	}

	out, err := svc.GetPortadaProximamente(context.Background(), "  ")
	if err != nil || out.Lang != LangES {
		t.Fatalf("lang vacío = %q (%v), se esperaba es", out.Lang, err)
	}
}

// TestGetPortadaProximamenteStableTieBreak fija el desempate estable del
// extracto: misma fecha → `title`, `id` (FR-003).
func TestGetPortadaProximamenteStableTieBreak(t *testing.T) {
	repo := newFakeRepository()
	seedPortadaItems(repo,
		map[string]string{"Beta": "2026-10-12", "Alfa": "2026-10-12"},
		map[string]string{"Gamma": "2026-10-12"})
	svc := newPublicService(t, repo)

	out, err := svc.GetPortadaProximamente(context.Background(), LangES)
	if err != nil {
		t.Fatalf("GetPortadaProximamente: %v", err)
	}
	titles := make([]string, 0, len(out.Items))
	for _, item := range out.Items {
		titles = append(titles, item.Title)
	}
	want := []string{"Alfa", "Beta", "Gamma"}
	if strings.Join(titles, ",") != strings.Join(want, ",") {
		t.Errorf("items = %v, se esperaban %v", titles, want)
	}
}
