package agenda

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
)

// repository_eventos.go es el acceso a datos de los eventos y de su lista de
// fechas (FR-004/FR-010/FR-013/FR-018). Cada mutación escribe el evento, su
// lista COMPLETA de fechas y sus filas de auditoría en la MISMA transacción
// (R4-10/R4-11): o todo o nada.
//
// La única vía pública de un evento es `ListUpcomingEvents` (+
// `ListUpcomingEventDates`): exige `publication_state='published'` Y una fecha
// próxima según la regla R4-2, con `today`/`now_time` que el service calcula en
// `APP_TZ` y pasa parametrizados (nunca SQL concatenado, §IV/CWE-89).

// InsertEvent crea un evento con su lista completa de fechas (1..50, ya
// validada por el service) y sus filas de auditoría, todo en una transacción.
func (r *repository) InsertEvent(ctx context.Context, event Event, dates []EventDate, actions ...audit.Action) (Event, error) {
	var result Event
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.InsertEvent(ctx, gendb.InsertEventParams{
			TitleEs:          event.TitleEs,
			TitleEn:          pgTextPtr(event.TitleEn),
			FrequencyEs:      pgTextPtr(event.FrequencyEs),
			FrequencyEn:      pgTextPtr(event.FrequencyEn),
			DetailEs:         event.DetailEs,
			DetailEn:         pgTextPtr(event.DetailEn),
			PlaceEs:          event.PlaceEs,
			PlaceEn:          pgTextPtr(event.PlaceEn),
			MapUrl:           pgTextPtr(event.MapURL),
			OrganizerEs:      pgTextPtr(event.OrganizerEs),
			OrganizerEn:      pgTextPtr(event.OrganizerEn),
			CategoryID:       pgUUIDPtr(event.CategoryID),
			ImageFile:        pgTextPtr(event.ImageFile),
			ImageAltEs:       pgTextPtr(event.ImageAltEs),
			ImageAltEn:       pgTextPtr(event.ImageAltEn),
			PublicationState: string(event.PublicationState),
		})
		if err != nil {
			return wrap(err, "insert event", "", "El evento no se pudo guardar: revisa su categoría")
		}
		saved := mapEvent(row)
		if err := replaceEventDates(ctx, tx, saved.ID, dates); err != nil {
			return err
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = saved
		return nil
	})
	if err != nil {
		return Event{}, err
	}
	return result, nil
}

// UpdateEvent reemplaza los datos del evento y, si `dates` no es nil, la lista
// COMPLETA de fechas (R4-11: el PATCH con `dates` reemplaza, no diferencia); las
// filas de auditoría van en la misma transacción. `dates == nil` deja las fechas
// intactas (PATCH sin `dates`); una lista vacía las deja sin ninguna.
func (r *repository) UpdateEvent(ctx context.Context, event Event, dates []EventDate, actions ...audit.Action) (Event, error) {
	var result Event
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.UpdateEvent(ctx, gendb.UpdateEventParams{
			ID:               pgUUID(event.ID),
			TitleEs:          event.TitleEs,
			TitleEn:          pgTextPtr(event.TitleEn),
			FrequencyEs:      pgTextPtr(event.FrequencyEs),
			FrequencyEn:      pgTextPtr(event.FrequencyEn),
			DetailEs:         event.DetailEs,
			DetailEn:         pgTextPtr(event.DetailEn),
			PlaceEs:          event.PlaceEs,
			PlaceEn:          pgTextPtr(event.PlaceEn),
			MapUrl:           pgTextPtr(event.MapURL),
			OrganizerEs:      pgTextPtr(event.OrganizerEs),
			OrganizerEn:      pgTextPtr(event.OrganizerEn),
			CategoryID:       pgUUIDPtr(event.CategoryID),
			ImageFile:        pgTextPtr(event.ImageFile),
			ImageAltEs:       pgTextPtr(event.ImageAltEs),
			ImageAltEn:       pgTextPtr(event.ImageAltEn),
			PublicationState: string(event.PublicationState),
		})
		if err != nil {
			return wrap(err, "update event", "El evento no existe", "El evento no se pudo guardar: revisa su categoría")
		}
		saved := mapEvent(row)
		if dates != nil {
			if err := replaceEventDates(ctx, tx, saved.ID, dates); err != nil {
				return err
			}
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = saved
		return nil
	})
	if err != nil {
		return Event{}, err
	}
	return result, nil
}

// DeleteEvent borra físicamente el evento (sus fechas caen por ON DELETE
// CASCADE) y registra su auditoría en la misma transacción. Devuelve false si el
// id no existía.
func (r *repository) DeleteEvent(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	deleted := false
	err := r.withTx(ctx, func(tx *repository) error {
		rows, err := tx.q.DeleteEvent(ctx, pgUUID(id))
		if err != nil {
			return wrap(err, "delete event", "", "")
		}
		if rows == 0 {
			return nil
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		deleted = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return deleted, nil
}

// GetEventByID devuelve el evento del panel (borradores incluidos). El segundo
// valor es false si el id no existe.
func (r *repository) GetEventByID(ctx context.Context, id uuid.UUID) (Event, bool, error) {
	row, err := r.q.GetEventByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Event{}, false, nil
		}
		return Event{}, false, wrap(err, "get event by id", "", "")
	}
	return mapEvent(row), true, nil
}

// replaceEventDates reemplaza la lista COMPLETA de fechas del evento dentro del
// tx vigente (R4-11). Implementa el `ReplaceEventDates` de data-model.md con las
// dos consultas de T404 (`DeleteEventDates` + `InsertEventDates`) en ese orden:
// van separadas a propósito, porque un DELETE + INSERT en una sola sentencia
// choca con `UNIQUE (event_id, event_date)` al reinsertar una fecha que ya
// existía. Las horas van como texto y "" significa "sin hora" (la consulta lo
// convierte en NULL).
func replaceEventDates(ctx context.Context, tx *repository, eventID uuid.UUID, dates []EventDate) error {
	if err := tx.q.DeleteEventDates(ctx, pgUUID(eventID)); err != nil {
		return wrap(err, "delete event dates", "", "")
	}
	if len(dates) == 0 {
		return nil
	}

	eventDates := make([]pgtype.Date, 0, len(dates))
	startTimes := make([]string, 0, len(dates))
	endTimes := make([]string, 0, len(dates))
	for _, date := range dates {
		eventDates = append(eventDates, pgDate(date.Date))
		startTimes = append(startTimes, clockValue(date.StartTime))
		endTimes = append(endTimes, clockValue(date.EndTime))
	}

	if err := tx.q.InsertEventDates(ctx, gendb.InsertEventDatesParams{
		EventID:    pgUUID(eventID),
		EventDates: eventDates,
		StartTimes: startTimes,
		EndTimes:   endTimes,
	}); err != nil {
		return wrap(err, "insert event dates", "", "La fecha ya está cargada en este evento")
	}
	return nil
}

// ListEventDates devuelve todas las fechas del evento (panel), en orden
// ascendente: el service calcula `isPast` y `nextDate` con la regla R4-2.
func (r *repository) ListEventDates(ctx context.Context, eventID uuid.UUID) ([]EventDate, error) {
	rows, err := r.q.ListEventDates(ctx, pgUUID(eventID))
	if err != nil {
		return nil, wrap(err, "list event dates", "", "")
	}
	dates := make([]EventDate, 0, len(rows))
	for _, row := range rows {
		dates = append(dates, mapEventDate(row))
	}
	return dates, nil
}

// ListUpcomingEventDates devuelve solo las fechas PRÓXIMAS del evento según la
// regla R4-2 (la primera es "la próxima fecha" de FR-004). Es la única vía
// pública de las fechas: las pasadas no salen de aquí.
func (r *repository) ListUpcomingEventDates(ctx context.Context, eventID uuid.UUID, today, nowTime string) ([]EventDate, error) {
	rows, err := r.q.ListUpcomingEventDates(ctx, gendb.ListUpcomingEventDatesParams{
		EventID: pgUUID(eventID),
		Today:   pgDate(today),
		NowTime: pgTime(nowTime),
	})
	if err != nil {
		return nil, wrap(err, "list upcoming event dates", "", "")
	}
	dates := make([]EventDate, 0, len(rows))
	for _, row := range rows {
		dates = append(dates, mapEventDate(row))
	}
	return dates, nil
}

// ExistsEventDuplicate detecta el duplicado DENTRO del módulo (R4-9): otro
// evento con el mismo título y lugar normalizados (minúsculas + `btrim`) y al
// menos una fecha en común. `excludeID` permite editar un evento sin que se
// detecte a sí mismo (nil al crear). Un evento y una actividad con los mismos
// datos NUNCA son duplicados (Decisión 3): la consulta solo mira agenda_events.
func (r *repository) ExistsEventDuplicate(ctx context.Context, title, place string, dates []string, excludeID *uuid.UUID) (bool, error) {
	duplicate, err := r.q.ExistsEventDuplicate(ctx, gendb.ExistsEventDuplicateParams{
		TitleEs:        title,
		PlaceEs:        place,
		EventDates:     pgDates(dates),
		ExcludeEventID: pgUUIDPtr(excludeID),
	})
	if err != nil {
		return false, wrap(err, "exists event duplicate", "", "")
	}
	return duplicate, nil
}

// ListEvents devuelve la página del panel: todos los eventos (borradores y
// pasados incluidos), ordenados por `created_at DESC, id DESC` (§8.1.7). El
// filtro por estado es opcional (nil = todos).
func (r *repository) ListEvents(ctx context.Context, state *PublicationState, params paginate.Params) ([]Event, error) {
	rows, err := r.q.ListEvents(ctx, gendb.ListEventsParams{
		PublicationState: pgStatePtr(state),
		Lim:              int32(params.Limit),
		Off:              int32(params.Offset),
	})
	if err != nil {
		return nil, wrap(err, "list events", "", "")
	}
	events := make([]Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, mapEvent(row))
	}
	return events, nil
}

// CountEvents cuenta los eventos con el mismo filtro que `ListEvents`, para que
// `total` corresponda a la página.
func (r *repository) CountEvents(ctx context.Context, state *PublicationState) (int64, error) {
	total, err := r.q.CountEvents(ctx, pgStatePtr(state))
	if err != nil {
		return 0, wrap(err, "count events", "", "")
	}
	return total, nil
}

// ListUpcomingEvents es la ÚNICA vía pública de los eventos (FR-013): devuelve
// solo los publicados con al menos una fecha próxima (R4-2), ordenados por su
// primera fecha próxima con desempate estable `title_es, id` (FR-003). `limit`
// es el tope de seguridad del bloque (`maxBlockItems`, RG4-8).
func (r *repository) ListUpcomingEvents(ctx context.Context, today, nowTime string, limit int) ([]Event, error) {
	rows, err := r.q.ListUpcomingEvents(ctx, gendb.ListUpcomingEventsParams{
		Lim:     int32(limit),
		Today:   pgDate(today),
		NowTime: pgTime(nowTime),
	})
	if err != nil {
		return nil, wrap(err, "list upcoming events", "", "")
	}
	events := make([]Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, mapUpcomingEventRow(row))
	}
	return events, nil
}

// --- Mapeo de filas → dominio ---

// mapEvent traduce la fila de un evento (sqlc) a la entidad del dominio.
func mapEvent(row gendb.AgendaEvent) Event {
	return Event{
		ID:               uuidValue(row.ID),
		TitleEs:          row.TitleEs,
		TitleEn:          textPtr(row.TitleEn),
		FrequencyEs:      textPtr(row.FrequencyEs),
		FrequencyEn:      textPtr(row.FrequencyEn),
		DetailEs:         row.DetailEs,
		DetailEn:         textPtr(row.DetailEn),
		PlaceEs:          row.PlaceEs,
		PlaceEn:          textPtr(row.PlaceEn),
		MapURL:           textPtr(row.MapUrl),
		OrganizerEs:      textPtr(row.OrganizerEs),
		OrganizerEn:      textPtr(row.OrganizerEn),
		CategoryID:       uuidPtr(row.CategoryID),
		ImageFile:        textPtr(row.ImageFile),
		ImageAltEs:       textPtr(row.ImageAltEs),
		ImageAltEn:       textPtr(row.ImageAltEn),
		PublicationState: PublicationState(row.PublicationState),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

// mapUpcomingEventRow traduce la fila del listado público: espeja `mapEvent`
// (sqlc genera un tipo de fila propio porque la consulta añade `next_date`, que
// el service no necesita: el orden ya viene resuelto del SQL).
func mapUpcomingEventRow(row gendb.ListUpcomingEventsRow) Event {
	return Event{
		ID:               uuidValue(row.ID),
		TitleEs:          row.TitleEs,
		TitleEn:          textPtr(row.TitleEn),
		FrequencyEs:      textPtr(row.FrequencyEs),
		FrequencyEn:      textPtr(row.FrequencyEn),
		DetailEs:         row.DetailEs,
		DetailEn:         textPtr(row.DetailEn),
		PlaceEs:          row.PlaceEs,
		PlaceEn:          textPtr(row.PlaceEn),
		MapURL:           textPtr(row.MapUrl),
		OrganizerEs:      textPtr(row.OrganizerEs),
		OrganizerEn:      textPtr(row.OrganizerEn),
		CategoryID:       uuidPtr(row.CategoryID),
		ImageFile:        textPtr(row.ImageFile),
		ImageAltEs:       textPtr(row.ImageAltEs),
		ImageAltEn:       textPtr(row.ImageAltEn),
		PublicationState: PublicationState(row.PublicationState),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

// mapEventDate traduce la fila de una fecha del evento.
func mapEventDate(row gendb.AgendaEventDate) EventDate {
	return EventDate{
		ID:        uuidValue(row.ID),
		EventID:   uuidValue(row.EventID),
		Date:      dateString(row.EventDate),
		StartTime: textPtr(row.StartTime),
		EndTime:   textPtr(row.EndTime),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

// --- Helpers de parámetros ---

// pgDates traduce una lista de fechas de calendario del dominio al parámetro
// `date[]` de sqlc.
func pgDates(dates []string) []pgtype.Date {
	out := make([]pgtype.Date, 0, len(dates))
	for _, date := range dates {
		out = append(out, pgDate(date))
	}
	return out
}

// pgStatePtr traduce el filtro opcional de estado de publicación (nil = sin
// filtro, que es un NULL en la consulta).
func pgStatePtr(state *PublicationState) pgtype.Text {
	if state == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: string(*state), Valid: true}
}

// clockValue devuelve la hora de una fecha del dominio como texto para el
// parámetro `text[]` de `InsertEventDates`: nil (sin hora) viaja como "" y la
// consulta lo convierte en NULL.
func clockValue(clock *string) string {
	if clock == nil {
		return ""
	}
	return *clock
}
