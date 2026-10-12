package agenda

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
)

// repository_actividades.go es el acceso a datos de las actividades (FR-005,
// FR-011, FR-013, FR-019): una acción puntual con UNA fecha concreta y su
// horario opcional (Decisión 3), sin periodicidad ni tabla de fechas. Cada
// mutación va en una transacción con sus filas de auditoría (R4-10).
//
// El duplicado exacto del módulo (mismo título, fecha y lugar) lo rechaza la BD
// con `UNIQUE (title_es, activity_date, place_es)` y se traduce a `conflict`
// (R4-9). Un evento con los mismos datos NUNCA es duplicado de una actividad
// (Decisión 3): son tablas distintas.

// InsertActivity crea una actividad y registra su auditoría en la misma
// transacción.
func (r *repository) InsertActivity(ctx context.Context, activity Activity, actions ...audit.Action) (Activity, error) {
	var result Activity
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.InsertActivity(ctx, gendb.InsertActivityParams{
			TitleEs:          activity.TitleEs,
			TitleEn:          pgTextPtr(activity.TitleEn),
			DetailEs:         activity.DetailEs,
			DetailEn:         pgTextPtr(activity.DetailEn),
			PlaceEs:          activity.PlaceEs,
			PlaceEn:          pgTextPtr(activity.PlaceEn),
			MapUrl:           pgTextPtr(activity.MapURL),
			OrganizerEs:      pgTextPtr(activity.OrganizerEs),
			OrganizerEn:      pgTextPtr(activity.OrganizerEn),
			CategoryID:       pgUUIDPtr(activity.CategoryID),
			ImageFile:        pgTextPtr(activity.ImageFile),
			ImageAltEs:       pgTextPtr(activity.ImageAltEs),
			ImageAltEn:       pgTextPtr(activity.ImageAltEn),
			ActivityDate:     pgDate(activity.Date),
			StartTime:        pgTextPtr(activity.StartTime),
			EndTime:          pgTextPtr(activity.EndTime),
			PublicationState: string(activity.PublicationState),
		})
		if err != nil {
			return wrap(err, "insert activity", "", duplicateActivityMessage)
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapActivity(row)
		return nil
	})
	if err != nil {
		return Activity{}, err
	}
	return result, nil
}

// UpdateActivity reemplaza los datos de la actividad (y su fecha única) con la
// auditoría, en la misma transacción.
func (r *repository) UpdateActivity(ctx context.Context, activity Activity, actions ...audit.Action) (Activity, error) {
	var result Activity
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.UpdateActivity(ctx, gendb.UpdateActivityParams{
			ID:               pgUUID(activity.ID),
			TitleEs:          activity.TitleEs,
			TitleEn:          pgTextPtr(activity.TitleEn),
			DetailEs:         activity.DetailEs,
			DetailEn:         pgTextPtr(activity.DetailEn),
			PlaceEs:          activity.PlaceEs,
			PlaceEn:          pgTextPtr(activity.PlaceEn),
			MapUrl:           pgTextPtr(activity.MapURL),
			OrganizerEs:      pgTextPtr(activity.OrganizerEs),
			OrganizerEn:      pgTextPtr(activity.OrganizerEn),
			CategoryID:       pgUUIDPtr(activity.CategoryID),
			ImageFile:        pgTextPtr(activity.ImageFile),
			ImageAltEs:       pgTextPtr(activity.ImageAltEs),
			ImageAltEn:       pgTextPtr(activity.ImageAltEn),
			ActivityDate:     pgDate(activity.Date),
			StartTime:        pgTextPtr(activity.StartTime),
			EndTime:          pgTextPtr(activity.EndTime),
			PublicationState: string(activity.PublicationState),
		})
		if err != nil {
			return wrap(err, "update activity", "La actividad no existe", duplicateActivityMessage)
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapActivity(row)
		return nil
	})
	if err != nil {
		return Activity{}, err
	}
	return result, nil
}

// DeleteActivity borra físicamente la actividad y registra su auditoría en la
// misma transacción. Devuelve false si el id no existía.
func (r *repository) DeleteActivity(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	deleted := false
	err := r.withTx(ctx, func(tx *repository) error {
		rows, err := tx.q.DeleteActivity(ctx, pgUUID(id))
		if err != nil {
			return wrap(err, "delete activity", "", "")
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

// GetActivityByID devuelve la actividad del panel (borradoras incluidas). El
// segundo valor es false si el id no existe.
func (r *repository) GetActivityByID(ctx context.Context, id uuid.UUID) (Activity, bool, error) {
	row, err := r.q.GetActivityByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Activity{}, false, nil
		}
		return Activity{}, false, wrap(err, "get activity by id", "", "")
	}
	return mapActivity(row), true, nil
}

// ListActivities devuelve la página del panel: todas las actividades
// (borradoras y pasadas incluidas), ordenadas por `created_at DESC, id DESC`
// (§8.1.7). El filtro por estado es opcional (nil = todas).
func (r *repository) ListActivities(ctx context.Context, state *PublicationState, params paginate.Params) ([]Activity, error) {
	rows, err := r.q.ListActivities(ctx, gendb.ListActivitiesParams{
		PublicationState: pgStatePtr(state),
		Lim:              int32(params.Limit),
		Off:              int32(params.Offset),
	})
	if err != nil {
		return nil, wrap(err, "list activities", "", "")
	}
	activities := make([]Activity, 0, len(rows))
	for _, row := range rows {
		activities = append(activities, mapActivity(row))
	}
	return activities, nil
}

// CountActivities cuenta las actividades con el mismo filtro que
// `ListActivities`, para que `total` corresponda a la página.
func (r *repository) CountActivities(ctx context.Context, state *PublicationState) (int64, error) {
	total, err := r.q.CountActivities(ctx, pgStatePtr(state))
	if err != nil {
		return 0, wrap(err, "count activities", "", "")
	}
	return total, nil
}

// ListUpcomingActivities es la ÚNICA vía pública de las actividades (FR-013):
// solo las publicadas con fecha próxima (R4-2), en orden cronológico con
// desempate estable `start_time NULLS FIRST, title_es, id`. `limit` es el tope
// de seguridad del bloque (`maxBlockItems`, RG4-8).
func (r *repository) ListUpcomingActivities(ctx context.Context, today, nowTime string, limit int) ([]Activity, error) {
	rows, err := r.q.ListUpcomingActivities(ctx, gendb.ListUpcomingActivitiesParams{
		Lim:     int32(limit),
		Today:   pgDate(today),
		NowTime: pgTime(nowTime),
	})
	if err != nil {
		return nil, wrap(err, "list upcoming activities", "", "")
	}
	activities := make([]Activity, 0, len(rows))
	for _, row := range rows {
		activities = append(activities, mapActivity(row))
	}
	return activities, nil
}

// duplicateActivityMessage es el mensaje del 409 del módulo (R4-9): el service
// lo conserva y añade el detalle del campo cuando responde.
const duplicateActivityMessage = "Ya existe una actividad con el mismo título, fecha y lugar"

// mapActivity traduce la fila de una actividad (sqlc) a la entidad del dominio.
func mapActivity(row gendb.AgendaActivity) Activity {
	return Activity{
		ID:               uuidValue(row.ID),
		TitleEs:          row.TitleEs,
		TitleEn:          textPtr(row.TitleEn),
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
		Date:             dateString(row.ActivityDate),
		StartTime:        textPtr(row.StartTime),
		EndTime:          textPtr(row.EndTime),
		PublicationState: PublicationState(row.PublicationState),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}
