package agenda

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
)

// fakeRepository es el repositorio falso del service: un almacén en memoria que
// imita el comportamiento del repository real (filtros, límites, restricciones
// de unicidad y uso del catálogo) y recuerda lo que se le pidió, para poder
// comprobar los parámetros (corte de R4-2, tope de bloque) y las filas de
// auditoría.
//
// Para la cara pública aplica la MISMA función de dominio de la regla R4-2
// (isUpcoming): el SQL real (que es quien filtra en producción) se verifica en
// las pruebas de integración (T412/T413).
type fakeRepository struct {
	events     []Event
	eventDates map[uuid.UUID][]EventDate
	activities []Activity
	categories []Category

	// Auditoría: `recorded` son las acciones best-effort (denegaciones y
	// rechazos) y `committed` las que acompañan a una mutación en su transacción.
	recorded  []audit.Action
	committed []audit.Action

	// Guionización de errores y de la detección de duplicados.
	recordErr   error
	mutationErr error
	listErr     error
	duplicate   bool

	// Observaciones de las consultas.
	cutoffs     []cutoffCall
	blockLimits []int
}

// cutoffCall es una llamada al repositorio con el corte "hoy"/"ahora".
type cutoffCall struct {
	today   string
	nowTime string
}

// compile-time check: el falso satisface el puerto del service.
var _ Repository = (*fakeRepository)(nil)

// newFakeRepository construye el almacén en memoria vacío.
func newFakeRepository() *fakeRepository {
	return &fakeRepository{eventDates: map[uuid.UUID][]EventDate{}}
}

// upcomingDates devuelve las fechas próximas de un evento.
func (f *fakeRepository) upcomingDates(eventID uuid.UUID, today, nowTime string) []EventDate {
	out := make([]EventDate, 0, len(f.eventDates[eventID]))
	for _, date := range f.eventDates[eventID] {
		if isUpcoming(date.Date, today, nowTime, date.EndTime) {
			out = append(out, date)
		}
	}
	return out
}

// --- Auditoría ---

func (f *fakeRepository) RecordAction(_ context.Context, action audit.Action) error {
	if f.recordErr != nil {
		return f.recordErr
	}
	f.recorded = append(f.recorded, action)
	return nil
}

// --- Catálogo de categorías ---

func (f *fakeRepository) ListCategories(context.Context) ([]Category, error) {
	return f.categories, nil
}

func (f *fakeRepository) GetCategoryByID(_ context.Context, id uuid.UUID) (Category, bool, error) {
	for _, category := range f.categories {
		if category.ID == id {
			return category, true, nil
		}
	}
	return Category{}, false, nil
}

func (f *fakeRepository) InsertCategory(_ context.Context, category Category, actions ...audit.Action) (Category, error) {
	if f.mutationErr != nil {
		return Category{}, f.mutationErr
	}
	for _, stored := range f.categories {
		if stored.NameEs == category.NameEs {
			return Category{}, apperr.Conflict("Ya existe una categoría con ese nombre")
		}
	}
	category.ID = uuid.New()
	category.CreatedAt, category.UpdatedAt = time.Now(), time.Now()
	f.categories = append(f.categories, category)
	f.committed = append(f.committed, actions...)
	return category, nil
}

func (f *fakeRepository) UpdateCategory(_ context.Context, category Category, actions ...audit.Action) (Category, error) {
	if f.mutationErr != nil {
		return Category{}, f.mutationErr
	}
	for _, stored := range f.categories {
		if stored.ID != category.ID && stored.NameEs == category.NameEs {
			return Category{}, apperr.Conflict("Ya existe una categoría con ese nombre")
		}
	}
	for i, stored := range f.categories {
		if stored.ID == category.ID {
			f.categories[i] = category
			f.committed = append(f.committed, actions...)
			return category, nil
		}
	}
	return Category{}, apperr.NotFound("La categoría no existe")
}

func (f *fakeRepository) DeleteCategory(_ context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	if f.mutationErr != nil {
		return false, f.mutationErr
	}
	if usage, err := f.CountCategoryUsage(context.Background(), id); err != nil {
		return false, err
	} else if usage > 0 {
		return false, apperr.Conflict("La categoría está en uso por eventos o actividades")
	}
	for i, stored := range f.categories {
		if stored.ID == id {
			f.categories = append(f.categories[:i], f.categories[i+1:]...)
			f.committed = append(f.committed, actions...)
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeRepository) CountCategoryUsage(_ context.Context, id uuid.UUID) (int64, error) {
	var total int64
	for _, event := range f.events {
		if event.CategoryID != nil && *event.CategoryID == id {
			total++
		}
	}
	for _, activity := range f.activities {
		if activity.CategoryID != nil && *activity.CategoryID == id {
			total++
		}
	}
	return total, nil
}

// --- Eventos ---

func (f *fakeRepository) InsertEvent(_ context.Context, event Event, dates []EventDate, actions ...audit.Action) (Event, error) {
	if f.mutationErr != nil {
		return Event{}, f.mutationErr
	}
	event.ID = uuid.New()
	event.CreatedAt, event.UpdatedAt = time.Now(), time.Now()
	f.events = append(f.events, event)
	f.storeDates(event.ID, dates)
	f.committed = append(f.committed, actions...)
	return event, nil
}

func (f *fakeRepository) UpdateEvent(_ context.Context, event Event, dates []EventDate, actions ...audit.Action) (Event, error) {
	if f.mutationErr != nil {
		return Event{}, f.mutationErr
	}
	for i, stored := range f.events {
		if stored.ID == event.ID {
			event.CreatedAt = stored.CreatedAt
			event.UpdatedAt = time.Now()
			f.events[i] = event
			if dates != nil {
				f.storeDates(event.ID, dates)
			}
			f.committed = append(f.committed, actions...)
			return event, nil
		}
	}
	return Event{}, apperr.NotFound("El evento no existe")
}

func (f *fakeRepository) DeleteEvent(_ context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	if f.mutationErr != nil {
		return false, f.mutationErr
	}
	for i, stored := range f.events {
		if stored.ID == id {
			f.events = append(f.events[:i], f.events[i+1:]...)
			delete(f.eventDates, id)
			f.committed = append(f.committed, actions...)
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeRepository) GetEventByID(_ context.Context, id uuid.UUID) (Event, bool, error) {
	for _, event := range f.events {
		if event.ID == id {
			return event, true, nil
		}
	}
	return Event{}, false, nil
}

func (f *fakeRepository) ListEventDates(_ context.Context, eventID uuid.UUID) ([]EventDate, error) {
	dates := f.eventDates[eventID]
	out := make([]EventDate, 0, len(dates))
	for _, date := range dates {
		date.EventID = eventID
		out = append(out, date)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out, nil
}

func (f *fakeRepository) ListUpcomingEventDates(_ context.Context, eventID uuid.UUID, today, nowTime string) ([]EventDate, error) {
	dates := f.upcomingDates(eventID, today, nowTime)
	sort.SliceStable(dates, func(i, j int) bool { return dates[i].Date < dates[j].Date })
	return dates, nil
}

func (f *fakeRepository) ExistsEventDuplicate(context.Context, string, string, []string, *uuid.UUID) (bool, error) {
	return f.duplicate, nil
}

func (f *fakeRepository) ListEvents(_ context.Context, state *PublicationState, params paginate.Params) ([]Event, error) {
	filtered := make([]Event, 0, len(f.events))
	for _, event := range f.events {
		if state != nil && event.PublicationState != *state {
			continue
		}
		filtered = append(filtered, event)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].CreatedAt.Equal(filtered[j].CreatedAt) {
			return filtered[i].ID.String() > filtered[j].ID.String()
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})
	return paginateSlice(filtered, params), nil
}

func (f *fakeRepository) CountEvents(_ context.Context, state *PublicationState) (int64, error) {
	var total int64
	for _, event := range f.events {
		if state == nil || event.PublicationState == *state {
			total++
		}
	}
	return total, nil
}

func (f *fakeRepository) ListUpcomingEvents(_ context.Context, today, nowTime string, limit int) ([]Event, error) {
	f.cutoffs = append(f.cutoffs, cutoffCall{today: today, nowTime: nowTime})
	f.blockLimits = append(f.blockLimits, limit)
	if f.listErr != nil {
		return nil, f.listErr
	}

	out := make([]Event, 0, len(f.events))
	for _, event := range f.events {
		if event.PublicationState != StatePublished {
			continue
		}
		if len(f.upcomingDates(event.ID, today, nowTime)) == 0 {
			continue
		}
		if len(out) >= limit {
			break
		}
		out = append(out, event)
	}
	return out, nil
}

// storeDates guarda las fechas de un evento como las guardaría la base: con id
// propio y la lista COMPLETA reemplazada (R4-11).
func (f *fakeRepository) storeDates(eventID uuid.UUID, dates []EventDate) {
	stored := make([]EventDate, 0, len(dates))
	for _, date := range dates {
		date.ID = uuid.New()
		date.EventID = eventID
		stored = append(stored, date)
	}
	f.eventDates[eventID] = stored
}

// --- Actividades ---

func (f *fakeRepository) InsertActivity(_ context.Context, activity Activity, actions ...audit.Action) (Activity, error) {
	if f.mutationErr != nil {
		return Activity{}, f.mutationErr
	}
	for _, stored := range f.activities {
		if stored.TitleEs == activity.TitleEs && stored.Date == activity.Date && stored.PlaceEs == activity.PlaceEs {
			return Activity{}, apperr.Conflict(duplicateActivityMessage)
		}
	}
	activity.ID = uuid.New()
	activity.CreatedAt, activity.UpdatedAt = time.Now(), time.Now()
	f.activities = append(f.activities, activity)
	f.committed = append(f.committed, actions...)
	return activity, nil
}

func (f *fakeRepository) UpdateActivity(_ context.Context, activity Activity, actions ...audit.Action) (Activity, error) {
	if f.mutationErr != nil {
		return Activity{}, f.mutationErr
	}
	for _, stored := range f.activities {
		if stored.ID != activity.ID && stored.TitleEs == activity.TitleEs &&
			stored.Date == activity.Date && stored.PlaceEs == activity.PlaceEs {
			return Activity{}, apperr.Conflict(duplicateActivityMessage)
		}
	}
	for i, stored := range f.activities {
		if stored.ID == activity.ID {
			activity.CreatedAt = stored.CreatedAt
			activity.UpdatedAt = time.Now()
			f.activities[i] = activity
			f.committed = append(f.committed, actions...)
			return activity, nil
		}
	}
	return Activity{}, apperr.NotFound("La actividad no existe")
}

func (f *fakeRepository) DeleteActivity(_ context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	if f.mutationErr != nil {
		return false, f.mutationErr
	}
	for i, stored := range f.activities {
		if stored.ID == id {
			f.activities = append(f.activities[:i], f.activities[i+1:]...)
			f.committed = append(f.committed, actions...)
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeRepository) GetActivityByID(_ context.Context, id uuid.UUID) (Activity, bool, error) {
	for _, activity := range f.activities {
		if activity.ID == id {
			return activity, true, nil
		}
	}
	return Activity{}, false, nil
}

func (f *fakeRepository) ListActivities(_ context.Context, state *PublicationState, params paginate.Params) ([]Activity, error) {
	filtered := make([]Activity, 0, len(f.activities))
	for _, activity := range f.activities {
		if state != nil && activity.PublicationState != *state {
			continue
		}
		filtered = append(filtered, activity)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].CreatedAt.Equal(filtered[j].CreatedAt) {
			return filtered[i].ID.String() > filtered[j].ID.String()
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})
	return paginateSlice(filtered, params), nil
}

func (f *fakeRepository) CountActivities(_ context.Context, state *PublicationState) (int64, error) {
	var total int64
	for _, activity := range f.activities {
		if state == nil || activity.PublicationState == *state {
			total++
		}
	}
	return total, nil
}

func (f *fakeRepository) ListUpcomingActivities(_ context.Context, today, nowTime string, limit int) ([]Activity, error) {
	f.cutoffs = append(f.cutoffs, cutoffCall{today: today, nowTime: nowTime})
	f.blockLimits = append(f.blockLimits, limit)

	out := make([]Activity, 0, len(f.activities))
	for _, activity := range f.activities {
		if activity.PublicationState != StatePublished {
			continue
		}
		if !isUpcoming(activity.Date, today, nowTime, activity.EndTime) {
			continue
		}
		if len(out) >= limit {
			break
		}
		out = append(out, activity)
	}
	return out, nil
}

// paginateSlice aplica limit/offset a una lista ya ordenada (§8.1.3).
func paginateSlice[T any](items []T, params paginate.Params) []T {
	start := params.Offset
	if start > len(items) {
		return []T{}
	}
	end := start + params.Limit
	if params.Limit <= 0 || end > len(items) {
		end = len(items)
	}
	return items[start:end]
}
