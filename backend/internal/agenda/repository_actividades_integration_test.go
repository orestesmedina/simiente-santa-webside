//go:build integration

package agenda

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/paginate"
)

// Pruebas de integración del repository de actividades contra PostgreSQL real
// (T413): el duplicado exacto del módulo, los CHECK del horario y la única vía
// pública (`…Upcoming`). Las fechas y el "hoy"/"ahora" son fijos (R4-2), como en
// las pruebas de eventos.

// sampleActivity devuelve una actividad válida y publicable.
func sampleActivity(title string) Activity {
	return Activity{
		TitleEs:          title,
		DetailEs:         "Detalle de la prueba",
		PlaceEs:          "Salón",
		Date:             testFuture,
		PublicationState: StatePublished,
	}
}

// insertActivity crea la actividad y falla la prueba si no se puede.
func insertActivity(t *testing.T, repo *repository, activity Activity) Activity {
	t.Helper()
	saved, err := repo.InsertActivity(context.Background(), activity)
	if err != nil {
		t.Fatalf("InsertActivity(%q): %v", activity.TitleEs, err)
	}
	return saved
}

// upcomingActivityTitles devuelve los títulos de la vista pública, en su orden.
func upcomingActivityTitles(t *testing.T, repo *repository, limit int) []string {
	t.Helper()
	activities, err := repo.ListUpcomingActivities(context.Background(), testToday, testNowTime, limit)
	if err != nil {
		t.Fatalf("ListUpcomingActivities: %v", err)
	}
	titles := make([]string, 0, len(activities))
	for _, activity := range activities {
		titles = append(titles, activity.TitleEs)
	}
	return titles
}

// TestIntegrationActivityDuplicateIsConflict fija R4-9 en el módulo de
// actividades: el duplicado exacto (título + fecha + lugar) es un 409, y una
// actividad idéntica a un EVENTO se guarda sin problema (Decisión 3).
func TestIntegrationActivityDuplicateIsConflict(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	title := "Repartir comida " + uuid.NewString()
	insertActivity(t, repo, sampleActivity(title))

	// Duplicado exacto: mismo título, fecha y lugar (UNIQUE de la BD).
	_, err := repo.InsertActivity(ctx, sampleActivity(title))
	if err == nil {
		t.Fatal("InsertActivity aceptó el duplicado exacto")
	}
	if kind := errorKind(err); kind != apperr.KindConflict {
		t.Fatalf("kind = %v, se esperaba conflict (UNIQUE del módulo)", kind)
	}

	// Mismo título y fecha pero otro lugar: NO es duplicado.
	otherPlace := sampleActivity(title)
	otherPlace.PlaceEs = "Otro salón"
	insertActivity(t, repo, otherPlace)

	// Un EVENTO con los mismos datos no cuenta como duplicado de la actividad
	// (Decisión 3): se guarda sin conflicto.
	event := sampleEvent(title)
	event.PlaceEs = "Salón"
	if _, err := repo.InsertEvent(ctx, event, []EventDate{date(testFuture, nil, nil)}); err != nil {
		t.Fatalf("InsertEvent(idéntico a una actividad): %v", err)
	}

	total, err := repo.CountActivities(ctx, nil)
	if err != nil || total != 2 {
		t.Fatalf("actividades = %d (%v), se esperaban 2", total, err)
	}
}

// TestIntegrationActivityScheduleChecks comprueba que los CHECK de horario y de
// imagen+alt de la BD se traducen a `invalid` (FR-015/FR-020).
func TestIntegrationActivityScheduleChecks(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	endWithoutStart := sampleActivity("Fin sin inicio " + uuid.NewString())
	endWithoutStart.StartTime = nil
	endWithoutStart.EndTime = startOf("21:00")
	if _, err := repo.InsertActivity(ctx, endWithoutStart); errorKind(err) != apperr.KindInvalid {
		t.Fatalf("fin sin inicio = %v, se esperaba invalid", err)
	}

	endBeforeStart := sampleActivity("Fin anterior " + uuid.NewString())
	endBeforeStart.StartTime = startOf("21:00")
	endBeforeStart.EndTime = startOf("20:00")
	if _, err := repo.InsertActivity(ctx, endBeforeStart); errorKind(err) != apperr.KindInvalid {
		t.Fatalf("fin anterior al inicio = %v, se esperaba invalid", err)
	}

	imageWithoutAlt := sampleActivity("Imagen sin alt " + uuid.NewString())
	imageWithoutAlt.ImageFile = textOf("img_44444444-4444-4444-4444-444444444444.jpg")
	if _, err := repo.InsertActivity(ctx, imageWithoutAlt); errorKind(err) != apperr.KindInvalid {
		t.Fatalf("imagen sin alt = %v, se esperaba invalid", err)
	}

	// Un horario coherente sí se guarda.
	valid := sampleActivity("Horario válido " + uuid.NewString())
	valid.StartTime = startOf("19:00")
	valid.EndTime = startOf("20:30")
	saved := insertActivity(t, repo, valid)
	if saved.StartTime == nil || *saved.StartTime != "19:00" || saved.EndTime == nil || *saved.EndTime != "20:30" {
		t.Fatalf("horario guardado = %v-%v", saved.StartTime, saved.EndTime)
	}
}

// TestIntegrationListUpcomingActivitiesFollowsTheNextRule fija la regla R4-2 y
// el orden de la vista pública de actividades (FR-003/FR-013).
func TestIntegrationListUpcomingActivitiesFollowsTheNextRule(t *testing.T) {
	repo, _ := newAgendaRepo(t)

	draft := sampleActivity("Borrador " + uuid.NewString())
	draft.PublicationState = StateDraft
	insertActivity(t, repo, draft)

	expired := sampleActivity("Vencida " + uuid.NewString())
	expired.Date = testPast
	expired = insertActivity(t, repo, expired)

	todayEnded := sampleActivity("Hoy terminada " + uuid.NewString())
	todayEnded.Date = testToday
	todayEnded.StartTime = startOf(testStart)
	todayEnded.EndTime = startOf(testEndedEnd)
	insertActivity(t, repo, todayEnded)

	todayOpen := sampleActivity("Hoy sin fin " + uuid.NewString())
	todayOpen.Date = testToday
	insertActivity(t, repo, todayOpen)

	todayRunning := sampleActivity("Hoy en curso " + uuid.NewString())
	todayRunning.Date = testToday
	todayRunning.StartTime = startOf(testRunningStart)
	todayRunning.EndTime = startOf(testRunningEnd)
	insertActivity(t, repo, todayRunning)

	later := sampleActivity("Más adelante " + uuid.NewString())
	later.Date = testLater
	insertActivity(t, repo, later)

	titles := upcomingActivityTitles(t, repo, 200)
	want := map[string]bool{
		todayOpen.TitleEs:    true,
		todayRunning.TitleEs: true,
		later.TitleEs:        true,
	}
	if len(titles) != len(want) {
		t.Fatalf("actividades públicas = %v, se esperaban %d", titles, len(want))
	}
	for _, title := range titles {
		if !want[title] {
			t.Errorf("actividad inesperada en la vista pública: %q", title)
		}
	}
	for _, forbidden := range []string{draft.TitleEs, expired.TitleEs, todayEnded.TitleEs} {
		for _, title := range titles {
			if title == forbidden {
				t.Errorf("actividad no pública en la vista pública: %q", forbidden)
			}
		}
	}

	// Orden: primero las de hoy (sin hora de inicio primero), luego la posterior.
	if titles[0] != todayOpen.TitleEs {
		t.Errorf("primera = %q, se esperaba la de hoy sin hora (%q)", titles[0], todayOpen.TitleEs)
	}
	if titles[2] != later.TitleEs {
		t.Errorf("última = %q, se esperaba %q", titles[2], later.TitleEs)
	}

	// El estado de la vencida NO cambia (FR-006/Q4).
	stored, found, err := repo.GetActivityByID(context.Background(), expired.ID)
	if err != nil || !found {
		t.Fatalf("GetActivityByID(vencida) = %v (%v)", found, err)
	}
	if stored.PublicationState != StatePublished {
		t.Errorf("el estado de la vencida cambió a %q", stored.PublicationState)
	}
}

// TestIntegrationActivityPanelListAndDelete cubre el listado del panel
// (paginado y con filtro), la edición y el borrado físico.
func TestIntegrationActivityPanelListAndDelete(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	draft := sampleActivity("Panel borrador " + uuid.NewString())
	draft.PublicationState = StateDraft
	draft = insertActivity(t, repo, draft)
	published := insertActivity(t, repo, sampleActivity("Panel publicada "+uuid.NewString()))

	total, err := repo.CountActivities(ctx, nil)
	if err != nil || total != 2 {
		t.Fatalf("CountActivities = %d (%v), se esperaban 2", total, err)
	}
	draftState := StateDraft
	if total, err = repo.CountActivities(ctx, &draftState); err != nil || total != 1 {
		t.Fatalf("CountActivities(borradores) = %d (%v), se esperaba 1", total, err)
	}

	page, err := repo.ListActivities(ctx, nil, paginate.Params{Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("ListActivities: %v", err)
	}
	if len(page) != 1 {
		t.Fatalf("página = %d, se esperaba 1", len(page))
	}

	// Edición: el cambio de detalle y de estado queda guardado (FR-014).
	published.DetailEs = "Detalle editado"
	published.PublicationState = StateDraft
	updated, err := repo.UpdateActivity(ctx, published)
	if err != nil {
		t.Fatalf("UpdateActivity: %v", err)
	}
	if updated.DetailEs != "Detalle editado" || updated.PublicationState != StateDraft {
		t.Errorf("actividad editada = %+v", updated)
	}

	// Borrado físico.
	deleted, err := repo.DeleteActivity(ctx, draft.ID)
	if err != nil || !deleted {
		t.Fatalf("DeleteActivity = %v (%v), se esperaba true", deleted, err)
	}
	if _, found, err := repo.GetActivityByID(ctx, draft.ID); err != nil || found {
		t.Fatalf("GetActivityByID tras borrar = %v (%v), se esperaba false", found, err)
	}
	if deleted, err := repo.DeleteActivity(ctx, uuid.New()); err != nil || deleted {
		t.Fatalf("DeleteActivity(id inexistente) = %v (%v), se esperaba false", deleted, err)
	}
	if _, err := repo.UpdateActivity(ctx, Activity{ID: uuid.New()}); errorKind(err) != apperr.KindNotFound {
		t.Fatalf("UpdateActivity(id inexistente) = %v, se esperaba not_found", err)
	}
}
