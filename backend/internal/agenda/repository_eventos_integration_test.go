//go:build integration

package agenda

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/paginate"
)

// Pruebas de integración del repository de eventos contra PostgreSQL real
// (T412): invariantes de las fechas, reemplazo completo de la lista, detección
// del duplicado del módulo y la única vía pública (`…Upcoming`).
//
// Las fechas y el "hoy"/"ahora" de la regla R4-2 son FIJOS: las pruebas no
// dependen del reloj real (el service los calcula con APP_TZ en T417).

const (
	testToday   = "2026-10-11"
	testNowTime = "12:00"
	testPast    = "2026-10-10"
	testFuture  = "2026-10-12"
	testLater   = "2026-10-19"

	// Horarios fijos de la regla R4-2: un rango ya terminado antes de las
	// 12:00 (`testEnded`) y otro que aún está en curso (`testRunning`).
	testStart        = "09:00"
	testEndedEnd     = "10:00"
	testRunningStart = "11:00"
	testRunningEnd   = "22:00"
)

// sampleEvent devuelve un evento válido y publicable (título único por prueba
// para que las búsquedas no se crucen entre casos).
func sampleEvent(title string) Event {
	return Event{
		TitleEs:          title,
		DetailEs:         "Detalle de la prueba",
		PlaceEs:          "Templo",
		PublicationState: StatePublished,
	}
}

// date arma una fecha del evento con su horario opcional.
func date(day string, start, end *string) EventDate {
	return EventDate{Date: day, StartTime: start, EndTime: end}
}

// startOf devuelve un puntero a una hora de inicio para las fechas con horario.
func startOf(clock string) *string { return &clock }

// insertEvent crea el evento con sus fechas y falla la prueba si no se puede.
func insertEvent(t *testing.T, repo *repository, event Event, dates []EventDate) Event {
	t.Helper()
	saved, err := repo.InsertEvent(context.Background(), event, dates)
	if err != nil {
		t.Fatalf("InsertEvent(%q): %v", event.TitleEs, err)
	}
	return saved
}

// upcomingTitles devuelve los títulos de la vista pública, en su orden.
func upcomingTitles(t *testing.T, repo *repository, limit int) []string {
	t.Helper()
	events, err := repo.ListUpcomingEvents(context.Background(), testToday, testNowTime, limit)
	if err != nil {
		t.Fatalf("ListUpcomingEvents: %v", err)
	}
	titles := make([]string, 0, len(events))
	for _, event := range events {
		titles = append(titles, event.TitleEs)
	}
	return titles
}

// TestIntegrationEventDatesUniqueRejectsRepeatedDate comprueba que
// `UNIQUE (event_id, event_date)` rechaza la fecha repetida dentro del mismo
// evento (FR-015) con un error de conflicto del dominio.
func TestIntegrationEventDatesUniqueRejectsRepeatedDate(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	event := sampleEvent("Evento con fecha repetida " + uuid.NewString())
	_, err := repo.InsertEvent(ctx, event, []EventDate{
		date(testFuture, nil, nil),
		date(testFuture, nil, nil),
	})
	if err == nil {
		t.Fatal("InsertEvent aceptó dos veces la misma fecha")
	}
	if kind := errorKind(err); kind != apperr.KindConflict {
		t.Fatalf("kind = %v, se esperaba conflict (UNIQUE de la base)", kind)
	}
}

// TestIntegrationReplaceEventDatesReplacesTheWholeList fija R4-11: la lista que
// se guarda es exactamente la pedida (las anteriores desaparecen, incluso
// reinsertando una fecha que ya existía) y `dates == nil` no toca las fechas.
func TestIntegrationReplaceEventDatesReplacesTheWholeList(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	start := "20:00"
	end := "21:30"
	saved := insertEvent(t, repo, sampleEvent("Evento con lista de fechas "+uuid.NewString()),
		[]EventDate{
			date(testPast, nil, nil),
			date(testFuture, &start, &end),
		})

	// Reemplazo con una lista que reinserta la MISMA fecha (caso que rompería un
	// DELETE + INSERT en una sola sentencia) y añade otra.
	saved.PlaceEs = "Templo"
	if _, err := repo.UpdateEvent(ctx, saved, []EventDate{
		date(testFuture, &start, nil),
		date(testLater, nil, nil),
	}); err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}

	dates, err := repo.ListEventDates(ctx, saved.ID)
	if err != nil {
		t.Fatalf("ListEventDates: %v", err)
	}
	if len(dates) != 2 {
		t.Fatalf("fechas = %d, se esperaban 2: %+v", len(dates), dates)
	}
	if dates[0].Date != testFuture || dates[1].Date != testLater {
		t.Errorf("fechas = %s/%s, se esperaba %s/%s", dates[0].Date, dates[1].Date, testFuture, testLater)
	}
	if dates[0].StartTime == nil || *dates[0].StartTime != start {
		t.Errorf("hora de inicio = %v, se esperaba %s", dates[0].StartTime, start)
	}
	if dates[0].EndTime != nil {
		t.Errorf("hora de fin = %v, se esperaba nil (se quitó)", *dates[0].EndTime)
	}
	if dates[1].StartTime != nil || dates[1].EndTime != nil {
		t.Errorf("la fecha sin horario debe venir sin horas: %+v", dates[1])
	}

	// `dates == nil` (PATCH sin fechas) no toca la lista.
	if _, err := repo.UpdateEvent(ctx, saved, nil); err != nil {
		t.Fatalf("UpdateEvent(sin fechas): %v", err)
	}
	dates, err = repo.ListEventDates(ctx, saved.ID)
	if err != nil {
		t.Fatalf("ListEventDates tras el PATCH: %v", err)
	}
	if len(dates) != 2 {
		t.Fatalf("el PATCH sin fechas cambió la lista: %+v", dates)
	}
}

// TestIntegrationExistsEventDuplicateDetectsSameModuleDuplicate fija R4-9: el
// duplicado se detecta con título y lugar normalizados (mayúsculas y espacios
// sobrantes no cuentan) y con al menos una fecha en común; el propio evento se
// excluye al editarlo.
func TestIntegrationExistsEventDuplicateDetectsSameModuleDuplicate(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	title := "Noche de aposentos " + uuid.NewString()
	event := sampleEvent(title)
	event.PlaceEs = "Templo Central"
	saved := insertEvent(t, repo, event, []EventDate{date(testFuture, nil, nil)})

	t.Run("normalización de título y lugar", func(t *testing.T) {
		duplicate, err := repo.ExistsEventDuplicate(ctx, upperTitle(title), "  Templo Central ", []string{testFuture}, nil)
		if err != nil {
			t.Fatalf("ExistsEventDuplicate: %v", err)
		}
		if !duplicate {
			t.Error("no detectó el duplicado con el mismo título, lugar y fecha")
		}
	})

	t.Run("sin fecha en común", func(t *testing.T) {
		duplicate, err := repo.ExistsEventDuplicate(ctx, upperTitle(title), "Templo Central", []string{testLater}, nil)
		if err != nil {
			t.Fatalf("ExistsEventDuplicate: %v", err)
		}
		if duplicate {
			t.Error("detectó un duplicado sin ninguna fecha en común")
		}
	})

	t.Run("otro lugar", func(t *testing.T) {
		duplicate, err := repo.ExistsEventDuplicate(ctx, upperTitle(title), "Otro lugar", []string{testFuture}, nil)
		if err != nil {
			t.Fatalf("ExistsEventDuplicate: %v", err)
		}
		if duplicate {
			t.Error("detectó un duplicado en otro lugar")
		}
	})

	t.Run("el propio evento se excluye al editarlo", func(t *testing.T) {
		duplicate, err := repo.ExistsEventDuplicate(ctx, title, "Templo Central", []string{testFuture}, &saved.ID)
		if err != nil {
			t.Fatalf("ExistsEventDuplicate: %v", err)
		}
		if duplicate {
			t.Error("el evento se detectó a sí mismo al editar")
		}
	})
}

// upperTitle devuelve el título en mayúsculas para probar la normalización
// (la consulta compara en minúsculas y sin espacios sobrantes).
func upperTitle(title string) string {
	return "  " + strings.ToUpper(title) + "  "
}

// TestIntegrationListUpcomingEventsFollowsTheNextRule fija los tres casos de la
// regla R4-2 en la única vía pública de los eventos (FR-002/FR-006/FR-013).
func TestIntegrationListUpcomingEventsFollowsTheNextRule(t *testing.T) {
	repo, _ := newAgendaRepo(t)

	todayEarlyEnd := testEndedEnd  // hoy, con fin YA terminado (12:00)
	todayLateEnd := testRunningEnd // hoy, con fin por delante

	draft := sampleEvent("Borrador " + uuid.NewString())
	draft.PublicationState = StateDraft
	insertEvent(t, repo, draft, []EventDate{date(testFuture, nil, nil)})

	expired := sampleEvent("Solo con fecha pasada " + uuid.NewString())
	expired = insertEvent(t, repo, expired, []EventDate{date(testPast, nil, nil)})

	todayEnded := sampleEvent("Hoy con fin terminado " + uuid.NewString())
	insertEvent(t, repo, todayEnded, []EventDate{date(testToday, startOf(testStart), &todayEarlyEnd)})

	todayOpen := sampleEvent("Hoy sin hora de fin " + uuid.NewString())
	insertEvent(t, repo, todayOpen, []EventDate{date(testToday, nil, nil)})

	todayRunning := sampleEvent("Hoy con fin por delante " + uuid.NewString())
	insertEvent(t, repo, todayRunning, []EventDate{date(testToday, startOf(testRunningStart), &todayLateEnd)})

	future := sampleEvent("Próximo " + uuid.NewString())
	insertEvent(t, repo, future, []EventDate{date(testFuture, nil, nil)})

	titles := upcomingTitles(t, repo, 200)
	want := map[string]bool{
		todayOpen.TitleEs:    true,
		todayRunning.TitleEs: true,
		future.TitleEs:       true,
	}
	if len(titles) != len(want) {
		t.Fatalf("eventos públicos = %v, se esperaban %d", titles, len(want))
	}
	for _, title := range titles {
		if !want[title] {
			t.Errorf("evento inesperado en la vista pública: %q", title)
		}
	}
	for _, forbidden := range []string{draft.TitleEs, expired.TitleEs, todayEnded.TitleEs} {
		for _, title := range titles {
			if title == forbidden {
				t.Errorf("evento no público en la vista pública: %q", forbidden)
			}
		}
	}

	// El estado del elemento vencido NO cambia (FR-006/Q4: el retiro es solo de
	// lectura).
	stored, found, err := repo.GetEventByID(context.Background(), expired.ID)
	if err != nil || !found {
		t.Fatalf("GetEventByID(viejo) = %v (%v)", found, err)
	}
	if stored.PublicationState != StatePublished {
		t.Errorf("el estado del vencido cambió a %q", stored.PublicationState)
	}
}

// TestIntegrationListUpcomingEventsOrdersByNextDate comprueba el orden
// cronológico por la primera fecha próxima con desempate estable (FR-003).
func TestIntegrationListUpcomingEventsOrdersByNextDate(t *testing.T) {
	repo, _ := newAgendaRepo(t)

	later := insertEvent(t, repo, sampleEvent("Segundo "+uuid.NewString()), []EventDate{date(testLater, nil, nil)})
	// El mismo día: dos eventos empatan en la próxima fecha.
	tieA := insertEvent(t, repo, sampleEvent("Empate A "+uuid.NewString()), []EventDate{date(testFuture, nil, nil)})
	tieB := insertEvent(t, repo, sampleEvent("Empate B "+uuid.NewString()), []EventDate{date(testFuture, nil, nil)})
	first := insertEvent(t, repo, sampleEvent("Primero "+uuid.NewString()), []EventDate{date(testToday, nil, nil)})

	titles := upcomingTitles(t, repo, 200)
	if len(titles) != 4 {
		t.Fatalf("eventos públicos = %v, se esperaban 4", titles)
	}
	if titles[0] != first.TitleEs {
		t.Errorf("primero = %q, se esperaba %q (hoy)", titles[0], first.TitleEs)
	}
	if titles[3] != later.TitleEs {
		t.Errorf("último = %q, se esperaba %q (fecha posterior)", titles[3], later.TitleEs)
	}
	// Empate: desempate estable por title_es.
	wantTie := []string{tieA.TitleEs, tieB.TitleEs}
	sort.Strings(wantTie)
	if titles[1] != wantTie[0] || titles[2] != wantTie[1] {
		t.Errorf("empate = %q/%q, se esperaba %q/%q", titles[1], titles[2], wantTie[0], wantTie[1])
	}
}

// TestIntegrationListUpcomingEventDatesOnlyUpcoming comprueba que las fechas
// públicas son solo las próximas (las pasadas y las ya terminadas no salen de la
// consulta). Una fecha por día por evento (UNIQUE): los dos casos de "hoy" van
// en eventos distintos.
func TestIntegrationListUpcomingEventDatesOnlyUpcoming(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	earlyEnd := testEndedEnd
	mixed := insertEvent(t, repo, sampleEvent("Fechas mezcladas "+uuid.NewString()), []EventDate{
		date(testPast, nil, nil),
		date(testToday, startOf(testStart), &earlyEnd), // hoy con fin ya terminado
		date(testFuture, nil, nil),
	})

	upcoming, err := repo.ListUpcomingEventDates(ctx, mixed.ID, testToday, testNowTime)
	if err != nil {
		t.Fatalf("ListUpcomingEventDates: %v", err)
	}
	if len(upcoming) != 1 || upcoming[0].Date != testFuture {
		t.Fatalf("fechas próximas = %+v, se esperaba solo %s", upcoming, testFuture)
	}

	// El panel ve TODAS las fechas (borradores y pasadas incluidas).
	all, err := repo.ListEventDates(ctx, mixed.ID)
	if err != nil {
		t.Fatalf("ListEventDates: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("fechas del panel = %d, se esperaban 3", len(all))
	}

	// Hoy sin hora de fin: se muestra todo el día (segundo caso de R4-2).
	todayOpen := insertEvent(t, repo, sampleEvent("Hoy sin fin "+uuid.NewString()),
		[]EventDate{date(testToday, nil, nil)})
	upcoming, err = repo.ListUpcomingEventDates(ctx, todayOpen.ID, testToday, testNowTime)
	if err != nil {
		t.Fatalf("ListUpcomingEventDates(hoy sin fin): %v", err)
	}
	if len(upcoming) != 1 || upcoming[0].Date != testToday {
		t.Fatalf("fechas próximas de hoy sin fin = %+v, se esperaba %s", upcoming, testToday)
	}
}

// TestIntegrationDeleteEventRemovesItsDates comprueba el borrado físico con las
// fechas (ON DELETE CASCADE) y el 404 del id inexistente.
func TestIntegrationDeleteEventRemovesItsDates(t *testing.T) {
	repo, pool := newAgendaRepo(t)
	ctx := context.Background()

	saved := insertEvent(t, repo, sampleEvent("Evento a borrar "+uuid.NewString()), []EventDate{
		date(testFuture, nil, nil),
		date(testLater, nil, nil),
	})

	deleted, err := repo.DeleteEvent(ctx, saved.ID)
	if err != nil || !deleted {
		t.Fatalf("DeleteEvent = %v (%v), se esperaba true", deleted, err)
	}

	if _, found, err := repo.GetEventByID(ctx, saved.ID); err != nil || found {
		t.Fatalf("GetEventByID tras borrar = %v (%v), se esperaba false", found, err)
	}
	var orphans int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM agenda_event_dates WHERE event_id = $1`, saved.ID).Scan(&orphans); err != nil {
		t.Fatalf("contar fechas huérfanas: %v", err)
	}
	if orphans != 0 {
		t.Errorf("quedaron %d fechas huérfanas", orphans)
	}

	if deleted, err := repo.DeleteEvent(ctx, uuid.New()); err != nil || deleted {
		t.Fatalf("DeleteEvent(id inexistente) = %v (%v), se esperaba false sin error", deleted, err)
	}
	if _, found, err := repo.GetEventByID(ctx, uuid.New()); err != nil || found {
		t.Fatalf("GetEventByID(id inexistente) = %v (%v), se esperaba false sin error", found, err)
	}
}

// TestIntegrationListEventsPaginationAndFilter fija el listado del panel:
// paginación (§8.1.3), filtro opcional por estado y orden
// `created_at DESC, id DESC` (§8.1.7).
func TestIntegrationListEventsPaginationAndFilter(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	draftA := sampleEvent("Borrador A " + uuid.NewString())
	draftA.PublicationState = StateDraft
	draftB := sampleEvent("Borrador B " + uuid.NewString())
	draftB.PublicationState = StateDraft
	published := sampleEvent("Publicado " + uuid.NewString())

	created := []Event{
		insertEvent(t, repo, draftA, []EventDate{date(testFuture, nil, nil)}),
		insertEvent(t, repo, draftB, []EventDate{date(testFuture, nil, nil)}),
		insertEvent(t, repo, published, []EventDate{date(testFuture, nil, nil)}),
	}

	total, err := repo.CountEvents(ctx, nil)
	if err != nil || total != 3 {
		t.Fatalf("CountEvents(todos) = %d (%v), se esperaba 3", total, err)
	}
	publishedState := StatePublished
	total, err = repo.CountEvents(ctx, &publishedState)
	if err != nil || total != 1 {
		t.Fatalf("CountEvents(publicados) = %d (%v), se esperaba 1", total, err)
	}

	page, err := repo.ListEvents(ctx, nil, paginate.Params{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if len(page) != 2 {
		t.Fatalf("página = %d, se esperaban 2", len(page))
	}

	// Orden esperado: created_at DESC, id DESC (desempate estable).
	sort.Slice(created, func(i, j int) bool {
		if created[i].CreatedAt.Equal(created[j].CreatedAt) {
			return created[i].ID.String() > created[j].ID.String()
		}
		return created[i].CreatedAt.After(created[j].CreatedAt)
	})
	if page[0].ID != created[0].ID || page[1].ID != created[1].ID {
		t.Errorf("orden de la página = %v/%v, se esperaba %v/%v", page[0].TitleEs, page[1].TitleEs, created[0].TitleEs, created[1].TitleEs)
	}

	onlyDrafts, err := repo.ListEvents(ctx, &draftA.PublicationState, paginate.Params{Limit: 100, Offset: 0})
	if err != nil {
		t.Fatalf("ListEvents(borradores): %v", err)
	}
	if len(onlyDrafts) != 2 {
		t.Errorf("borradores = %d, se esperaban 2", len(onlyDrafts))
	}

	second, err := repo.ListEvents(ctx, nil, paginate.Params{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("ListEvents(segunda página): %v", err)
	}
	if len(second) != 1 || second[0].ID != created[2].ID {
		t.Errorf("segunda página = %+v, se esperaba el tercer evento", second)
	}
}

// TestIntegrationUpdateEventKeepsPublicationStateWhenDatesAreUntouched cubre el
// "se publica al guardar" (FR-014) desde el repository: el PATCH que solo toca
// datos conserva el estado y la lista de fechas.
func TestIntegrationUpdateEventKeepsPublicationStateWhenDatesAreUntouched(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	saved := insertEvent(t, repo, sampleEvent("Evento editable "+uuid.NewString()), []EventDate{date(testFuture, nil, nil)})

	saved.DetailEs = "Detalle editado"
	updated, err := repo.UpdateEvent(ctx, saved, nil)
	if err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}
	if updated.DetailEs != "Detalle editado" {
		t.Errorf("DetailEs = %q, se esperaba el detalle editado", updated.DetailEs)
	}
	if updated.PublicationState != StatePublished {
		t.Errorf("PublicationState = %q, se esperaba published (FR-014)", updated.PublicationState)
	}
	dates, err := repo.ListEventDates(ctx, saved.ID)
	if err != nil {
		t.Fatalf("ListEventDates: %v", err)
	}
	if len(dates) != 1 {
		t.Errorf("fechas = %d, se esperaba 1", len(dates))
	}
}

// TestIntegrationUpdateEventNotFound comprueba que editar un evento inexistente
// es un 404 del dominio (nunca un error interno).
func TestIntegrationUpdateEventNotFound(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	ghost := sampleEvent("Fantasma")
	ghost.ID = uuid.New()
	_, err := repo.UpdateEvent(ctx, ghost, nil)
	if kind := errorKind(err); kind != apperr.KindNotFound {
		t.Fatalf("UpdateEvent(id inexistente) = %v (%v), se esperaba not_found", err, kind)
	}
}
