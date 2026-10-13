//go:build integration

package agenda

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Pruebas de integración del extracto público de la portada contra PostgreSQL
// real (T445, US7/FR-025): la consulta `ListNextUpcomingItems` (T444) mezcla
// eventos y actividades publicados y próximos, aplica la ventana de 2 meses con
// respaldo y el tope 5, con el corte de R4-2 fijo (nunca el reloj real).
//
// El corte de las pruebas es `testToday`/`testNowTime` (2026-10-11 12:00) y la
// ventana cierra 2 meses de calendario después: `2026-12-11`.

const (
	// portadaWindowEndForTest es `testToday` + 2 meses de calendario (R4-16).
	portadaWindowEndForTest = "2026-12-11"
	// portadaFarDate está fuera de la ventana (respaldo).
	portadaFarDate = "2027-06-01"
	// portadaNearDate está dentro de la ventana.
	portadaNearDate = "2026-11-01"
)

// nextItems llama al extracto con el corte y la ventana de las pruebas.
func nextItems(t *testing.T, repo *repository, limit int) []NextItem {
	t.Helper()
	items, err := repo.ListNextUpcomingItems(context.Background(), testToday, testNowTime, portadaWindowEndForTest, limit)
	if err != nil {
		t.Fatalf("ListNextUpcomingItems: %v", err)
	}
	return items
}

// titles devuelve los títulos de los ítems, en su orden.
func titles(items []NextItem) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.TitleEs)
	}
	return out
}

// TestIntegrationListNextUpcomingItemsWindowAndBackup fija la regla de la
// ventana: con ítems dentro, solo salen los de dentro (los lejanos esperan); con
// la ventana vacía, entran los siguientes próximos; nunca se mezclan.
func TestIntegrationListNextUpcomingItemsWindowAndBackup(t *testing.T) {
	repo, _ := newAgendaRepo(t)

	insideEvent := insertEvent(t, repo, sampleEvent("Evento cerca "+uuid.NewString()),
		[]EventDate{date(portadaNearDate, startOf("19:00"), nil)})
	insideActivity := insertActivity(t, repo, sampleActivityWithDate("Actividad cerca", "2026-11-05"))
	farEvent := insertEvent(t, repo, sampleEvent("Evento lejos "+uuid.NewString()),
		[]EventDate{date(portadaFarDate, nil, nil)})
	farActivity := insertActivity(t, repo, sampleActivityWithDate("Actividad lejana", "2027-03-01"))

	// La ventana manda: solo los dos de dentro, ordenados por fecha.
	items := nextItems(t, repo, maxPortadaItems)
	if len(items) != 2 {
		t.Fatalf("items = %v, se esperaban los 2 de la ventana", titles(items))
	}
	if items[0].TitleEs != insideEvent.TitleEs || items[0].Kind != NextItemEvent {
		t.Errorf("items[0] = %+v, se esperaba el evento de la ventana", items[0])
	}
	if items[1].TitleEs != insideActivity.TitleEs || items[1].Kind != NextItemActivity {
		t.Errorf("items[1] = %+v, se esperaba la actividad de la ventana", items[1])
	}
	if items[0].Date != portadaNearDate || items[0].StartTime == nil || *items[0].StartTime != "19:00" {
		t.Errorf("el ítem del evento debería traer su próxima fecha y horario: %+v", items[0])
	}
	for _, item := range items {
		if item.TitleEs == farEvent.TitleEs || item.TitleEs == farActivity.TitleEs {
			t.Errorf("un ítem fuera de la ventana se coló: %+v", item)
		}
	}

	// Ventana vacía: entra el respaldo (los siguientes próximos, por cercanía).
	if err := deleteEventAndActivity(t, repo, insideEvent.ID, insideActivity.ID); err != nil {
		t.Fatalf("vaciar la ventana: %v", err)
	}
	items = nextItems(t, repo, maxPortadaItems)
	if len(items) != 2 {
		t.Fatalf("items = %v, se esperaban los 2 del respaldo", titles(items))
	}
	if items[0].TitleEs != farActivity.TitleEs || items[1].TitleEs != farEvent.TitleEs {
		t.Errorf("respaldo = %v, se esperaba el orden por cercanía (actividad 2027-03-01, evento 2027-06-01)", titles(items))
	}
}

// dayAfter suma días de CALENDARIO a una fecha "YYYY-MM-DD" (pruebas del tope).
func dayAfter(day string, days int) string {
	parsed, err := time.Parse(DateFormat, day)
	if err != nil {
		panic(err)
	}
	return parsed.AddDate(0, 0, days).Format(DateFormat)
}

// sampleActivityWithDate arma una actividad publicada en el día indicado.
func sampleActivityWithDate(title, day string) Activity {
	activity := sampleActivity(title)
	activity.Date = day
	return activity
}

// deleteEventAndActivity borra el evento y la actividad indicados (para vaciar la
// ventana en la prueba del respaldo).
func deleteEventAndActivity(t *testing.T, repo *repository, eventID, activityID uuid.UUID) error {
	t.Helper()
	ctx := context.Background()
	if _, err := repo.DeleteEvent(ctx, eventID); err != nil {
		return err
	}
	_, err := repo.DeleteActivity(ctx, activityID)
	return err
}

// TestIntegrationListNextUpcomingItemsUsesTheFirstUpcomingDate comprueba que de
// cada evento se usa su PRIMERA fecha próxima (FR-004) y que el evento aparece
// una sola vez aunque tenga varias fechas próximas.
func TestIntegrationListNextUpcomingItemsUsesTheFirstUpcomingDate(t *testing.T) {
	repo, _ := newAgendaRepo(t)

	// Un evento con tres fechas próximas: hoy sin fin, una cercana y una lejana.
	event := insertEvent(t, repo, sampleEvent("Evento con varias fechas "+uuid.NewString()),
		[]EventDate{
			date(testToday, nil, nil),
			date(portadaNearDate, startOf("10:00"), nil),
			date(portadaFarDate, nil, nil),
		})

	items := nextItems(t, repo, maxPortadaItems)
	if len(items) != 1 {
		t.Fatalf("items = %v, se esperaba el evento una sola vez", titles(items))
	}
	if items[0].ID != event.ID {
		t.Fatalf("item = %+v, se esperaba el evento sembrado", items[0])
	}
	if items[0].Date != testToday {
		t.Errorf("date = %q, se esperaba la primera fecha próxima (%s)", items[0].Date, testToday)
	}
}

// TestIntegrationListNextUpcomingItemsOnlyPublishedAndUpcoming fija que el
// extracto nunca expone borradores ni vencidos (FR-013/SC-014) y que aplica los
// tres casos de la regla R4-2.
func TestIntegrationListNextUpcomingItemsOnlyPublishedAndUpcoming(t *testing.T) {
	repo, _ := newAgendaRepo(t)

	draft := sampleEvent("Borrador " + uuid.NewString())
	draft.PublicationState = StateDraft
	insertEvent(t, repo, draft, []EventDate{date(portadaNearDate, nil, nil)})

	expired := insertEvent(t, repo, sampleEvent("Vencido "+uuid.NewString()),
		[]EventDate{date(testPast, nil, nil)})

	// Hoy con fin ya terminado: fuera. Hoy sin fin: dentro.
	todayEnded := insertEvent(t, repo, sampleEvent("Hoy terminado "+uuid.NewString()),
		[]EventDate{date(testToday, startOf(testStart), ptrOf(testEndedEnd))})
	todayOpen := insertEvent(t, repo, sampleEvent("Hoy sin fin "+uuid.NewString()),
		[]EventDate{date(testToday, nil, nil)})

	draftActivity := sampleActivity("Actividad borrador " + uuid.NewString())
	draftActivity.PublicationState = StateDraft
	draftActivity.Date = portadaNearDate
	insertActivity(t, repo, draftActivity)

	expiredActivity := sampleActivityWithDate("Actividad vencida "+uuid.NewString(), testPast)
	insertActivity(t, repo, expiredActivity)

	items := nextItems(t, repo, maxPortadaItems)
	want := []string{todayOpen.TitleEs}
	if len(items) != len(want) {
		t.Fatalf("items = %v, se esperaba solo %v", titles(items), want)
	}
	for i, title := range want {
		if items[i].TitleEs != title {
			t.Errorf("items[%d] = %q, se esperaba %q", i, items[i].TitleEs, title)
		}
	}
	for _, item := range items {
		if item.ID == draft.ID || item.ID == expired.ID || item.ID == todayEnded.ID ||
			item.ID == draftActivity.ID || item.ID == expiredActivity.ID {
			t.Errorf("el extracto expuso un elemento que no debía: %+v", item)
		}
	}
}

// TestIntegrationListNextUpcomingItemsRespectsTheLimit fija el tope firme 5: con
// más ítems próximos solo salen los 5 más cercanos, mezclados por fecha.
func TestIntegrationListNextUpcomingItemsRespectsTheLimit(t *testing.T) {
	repo, _ := newAgendaRepo(t)

	// 4 eventos y 4 actividades alternando días dentro de la ventana.
	for i := 0; i < 4; i++ {
		insertEvent(t, repo, sampleEvent("Evento "+string(rune('A'+i))+" "+uuid.NewString()),
			[]EventDate{date(dayAfter(portadaNearDate, i*2), nil, nil)})
		insertActivity(t, repo, sampleActivityWithDate("Actividad "+string(rune('A'+i)), dayAfter(portadaNearDate, i*2+1)))
	}

	items := nextItems(t, repo, maxPortadaItems)
	if len(items) != maxPortadaItems {
		t.Fatalf("items = %d, se esperaban %d (tope firme)", len(items), maxPortadaItems)
	}
	for i := 1; i < len(items); i++ {
		if items[i-1].Date > items[i].Date {
			t.Errorf("los ítems no vienen ordenados por cercanía: %v", titles(items))
			break
		}
	}
	// Mezcla: los 5 primeros son evento/actividad alternándose por fecha.
	for i, item := range items {
		wantKind := NextItemEvent
		if i%2 == 1 {
			wantKind = NextItemActivity
		}
		if item.Kind != wantKind {
			t.Errorf("items[%d].Kind = %q, se esperaba %q", i, item.Kind, wantKind)
		}
	}
}

// TestIntegrationListNextUpcomingItemsEmpty fija FR-026 a nivel de datos: sin
// nada publicado y próximo, la consulta no devuelve filas (el service responde
// sin `items`).
func TestIntegrationListNextUpcomingItemsEmpty(t *testing.T) {
	repo, _ := newAgendaRepo(t)

	// Solo borradores y vencidos.
	draft := sampleEvent("Borrador " + uuid.NewString())
	draft.PublicationState = StateDraft
	insertEvent(t, repo, draft, []EventDate{date(portadaNearDate, nil, nil)})
	insertEvent(t, repo, sampleEvent("Vencido "+uuid.NewString()), []EventDate{date(testPast, nil, nil)})

	if items := nextItems(t, repo, maxPortadaItems); len(items) != 0 {
		t.Fatalf("items = %v, se esperaba vacío", titles(items))
	}
}

// TestIntegrationListNextUpcomingItemsCarriesTheImage comprueba que el extracto
// trae la imagen del elemento con su alt (FR-025) y sin detalle.
func TestIntegrationListNextUpcomingItemsCarriesTheImage(t *testing.T) {
	repo, _ := newAgendaRepo(t)

	event := sampleEvent("Evento con imagen " + uuid.NewString())
	event.ImageFile = ptrOf("img_11111111-1111-1111-1111-111111111111.jpg")
	event.ImageAltEs = ptrOf("Foto del evento")
	event.ImageAltEn = ptrOf("Event photo")
	event.DetailEs = "Un detalle que el extracto no debe traer"
	insertEvent(t, repo, event, []EventDate{date(portadaNearDate, nil, nil)})

	items := nextItems(t, repo, maxPortadaItems)
	if len(items) != 1 {
		t.Fatalf("items = %v", titles(items))
	}
	got := items[0]
	if got.ImageFile == nil || *got.ImageFile != "img_11111111-1111-1111-1111-111111111111.jpg" {
		t.Errorf("imageFile = %v", got.ImageFile)
	}
	if got.ImageAltEs == nil || *got.ImageAltEs != "Foto del evento" || got.ImageAltEn == nil {
		t.Errorf("alt = %v / %v", got.ImageAltEs, got.ImageAltEn)
	}
}
