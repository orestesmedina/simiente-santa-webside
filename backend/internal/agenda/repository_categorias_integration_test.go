//go:build integration

package agenda

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/paginate"
)

// Pruebas de integración del repository del catálogo de categorías contra
// PostgreSQL real (T414): nombre duplicado, borrado restringido mientras esté en
// uso y el recuento de uso de los dos módulos (Q6/R4-5/P4-6).

// sampleCategory devuelve una categoría válida con nombre único por prueba.
func sampleCategory(name string) Category {
	return Category{NameEs: name, SortOrder: 10}
}

// paginateAll devuelve los parámetros de una página con el tope máximo (§8.1.3).
func paginateAll() paginate.Params {
	return paginate.Params{Limit: paginate.MaxLimit, Offset: 0}
}

// insertCategory crea la categoría y falla la prueba si no se puede.
func insertCategory(t *testing.T, repo *repository, category Category) Category {
	t.Helper()
	saved, err := repo.InsertCategory(context.Background(), category)
	if err != nil {
		t.Fatalf("InsertCategory(%q): %v", category.NameEs, err)
	}
	return saved
}

// TestIntegrationCategoryDuplicateNameIsConflict fija `UNIQUE (name_es)`: el
// nombre repetido es un 409 del dominio, también al renombrar.
func TestIntegrationCategoryDuplicateNameIsConflict(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	name := "Acción social " + uuid.NewString()
	first := insertCategory(t, repo, sampleCategory(name))

	if _, err := repo.InsertCategory(ctx, sampleCategory(name)); errorKind(err) != apperr.KindConflict {
		t.Fatalf("InsertCategory(duplicado) = %v, se esperaba conflict", err)
	}

	other := insertCategory(t, repo, sampleCategory("Otra "+uuid.NewString()))
	other.NameEs = name
	if _, err := repo.UpdateCategory(ctx, other); errorKind(err) != apperr.KindConflict {
		t.Fatalf("UpdateCategory(duplicado) = %v, se esperaba conflict", err)
	}

	// Renombrar a un nombre libre sí procede (R4-5).
	other.NameEs = "Renombrada " + uuid.NewString()
	renamed, err := repo.UpdateCategory(ctx, other)
	if err != nil {
		t.Fatalf("UpdateCategory(renombrar): %v", err)
	}
	if renamed.NameEs != other.NameEs {
		t.Errorf("NameEs = %q, se esperaba %q", renamed.NameEs, other.NameEs)
	}

	// El catálogo conserva las dos categorías y su orden de presentación.
	categories, err := repo.ListCategories(ctx)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(categories) != 2 {
		t.Fatalf("categorías = %d, se esperaban 2", len(categories))
	}
	if _, found, err := repo.GetCategoryByID(ctx, first.ID); err != nil || !found {
		t.Fatalf("GetCategoryByID = %v (%v), se esperaba true", found, err)
	}
	if _, found, err := repo.GetCategoryByID(ctx, uuid.New()); err != nil || found {
		t.Fatalf("GetCategoryByID(inexistente) = %v (%v), se esperaba false", found, err)
	}
}

// TestIntegrationCategoryDeleteRestrictedInUse fija la decisión firme de la
// puerta 2 (P4-6): una categoría usada por un evento o una actividad NO se
// borra (409) y sigue en el catálogo; sin uso, el borrado procede.
func TestIntegrationCategoryDeleteRestrictedInUse(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	usedByEvent := insertCategory(t, repo, sampleCategory("Usada por evento "+uuid.NewString()))
	usedByActivity := insertCategory(t, repo, sampleCategory("Usada por actividad "+uuid.NewString()))
	free := insertCategory(t, repo, sampleCategory("Libre "+uuid.NewString()))

	event := sampleEvent("Evento con categoría " + uuid.NewString())
	event.CategoryID = &usedByEvent.ID
	insertEvent(t, repo, event, []EventDate{date(testFuture, nil, nil)})

	activity := sampleActivity("Actividad con categoría " + uuid.NewString())
	activity.CategoryID = &usedByActivity.ID
	insertActivity(t, repo, activity)

	t.Run("en uso por un evento", func(t *testing.T) {
		deleted, err := repo.DeleteCategory(ctx, usedByEvent.ID)
		if deleted {
			t.Fatal("DeleteCategory borró una categoría en uso")
		}
		if kind := errorKind(err); kind != apperr.KindConflict {
			t.Fatalf("kind = %v, se esperaba conflict (FK RESTRICT)", kind)
		}
		if _, found, err := repo.GetCategoryByID(ctx, usedByEvent.ID); err != nil || !found {
			t.Fatalf("la categoría desapareció: %v (%v)", found, err)
		}
	})

	t.Run("en uso por una actividad", func(t *testing.T) {
		deleted, err := repo.DeleteCategory(ctx, usedByActivity.ID)
		if deleted || errorKind(err) != apperr.KindConflict {
			t.Fatalf("DeleteCategory = %v (%v), se esperaba conflict sin borrar", deleted, err)
		}
	})

	t.Run("sin uso", func(t *testing.T) {
		deleted, err := repo.DeleteCategory(ctx, free.ID)
		if err != nil || !deleted {
			t.Fatalf("DeleteCategory(libre) = %v (%v), se esperaba true", deleted, err)
		}
		if _, found, err := repo.GetCategoryByID(ctx, free.ID); err != nil || found {
			t.Fatalf("la categoría libre sigue ahí: %v (%v)", found, err)
		}
		if deleted, err := repo.DeleteCategory(ctx, uuid.New()); err != nil || deleted {
			t.Fatalf("DeleteCategory(id inexistente) = %v (%v), se esperaba false", deleted, err)
		}
	})
}

// TestIntegrationCountCategoryUsageCountsBothModules comprueba el recuento del
// mensaje "está en uso por N elementos": suma eventos y actividades, y no cuenta
// los elementos sin categoría.
func TestIntegrationCountCategoryUsageCountsBothModules(t *testing.T) {
	repo, _ := newAgendaRepo(t)
	ctx := context.Background()

	category := insertCategory(t, repo, sampleCategory("Con usos "+uuid.NewString()))

	if total, err := repo.CountCategoryUsage(ctx, category.ID); err != nil || total != 0 {
		t.Fatalf("CountCategoryUsage(sin usos) = %d (%v), se esperaba 0", total, err)
	}

	for i := 0; i < 2; i++ {
		event := sampleEvent("Evento contado " + uuid.NewString())
		event.CategoryID = &category.ID
		insertEvent(t, repo, event, []EventDate{date(testFuture, nil, nil)})
	}
	activity := sampleActivity("Actividad contada " + uuid.NewString())
	activity.CategoryID = &category.ID
	insertActivity(t, repo, activity)

	// Un elemento sin categoría no suma.
	insertEvent(t, repo, sampleEvent("Evento sin categoría "+uuid.NewString()), []EventDate{date(testFuture, nil, nil)})

	total, err := repo.CountCategoryUsage(ctx, category.ID)
	if err != nil {
		t.Fatalf("CountCategoryUsage: %v", err)
	}
	if total != 3 {
		t.Errorf("usos = %d, se esperaban 3 (2 eventos + 1 actividad)", total)
	}

	// Al quitar la categoría de todos sus usos, el borrado procede.
	for _, event := range listEventsWithCategory(t, repo, category.ID) {
		event.CategoryID = nil
		if _, err := repo.UpdateEvent(ctx, event, nil); err != nil {
			t.Fatalf("UpdateEvent(quitar categoría): %v", err)
		}
	}
	for _, item := range listActivitiesWithCategory(t, repo, category.ID) {
		item.CategoryID = nil
		if _, err := repo.UpdateActivity(ctx, item); err != nil {
			t.Fatalf("UpdateActivity(quitar categoría): %v", err)
		}
	}
	if total, err := repo.CountCategoryUsage(ctx, category.ID); err != nil || total != 0 {
		t.Fatalf("CountCategoryUsage(tras desligar) = %d (%v), se esperaba 0", total, err)
	}
	if deleted, err := repo.DeleteCategory(ctx, category.ID); err != nil || !deleted {
		t.Fatalf("DeleteCategory(sin usos) = %v (%v), se esperaba true", deleted, err)
	}
}

// listEventsWithCategory devuelve los eventos que usan la categoría (lectura del
// panel sin paginar, para el escenario de desligar y borrar).
func listEventsWithCategory(t *testing.T, repo *repository, categoryID uuid.UUID) []Event {
	t.Helper()
	events, err := repo.ListEvents(context.Background(), nil, paginateAll())
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	filtered := make([]Event, 0, len(events))
	for _, event := range events {
		if event.CategoryID != nil && *event.CategoryID == categoryID {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// listActivitiesWithCategory devuelve las actividades que usan la categoría.
func listActivitiesWithCategory(t *testing.T, repo *repository, categoryID uuid.UUID) []Activity {
	t.Helper()
	activities, err := repo.ListActivities(context.Background(), nil, paginateAll())
	if err != nil {
		t.Fatalf("ListActivities: %v", err)
	}
	filtered := make([]Activity, 0, len(activities))
	for _, activity := range activities {
		if activity.CategoryID != nil && *activity.CategoryID == categoryID {
			filtered = append(filtered, activity)
		}
	}
	return filtered
}
