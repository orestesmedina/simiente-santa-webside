package ministerios

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

// repository_ministerios.go es el acceso a datos de la gestión del panel
// (FR-010/FR-014/FR-016/FR-019): el alta, la edición, el borrado físico, el
// detalle, el listado paginado con su total y la detección del duplicado
// exacto. Cada mutación escribe su fila (o filas) de auditoría en la MISMA
// transacción que el cambio (FR-017/R6-12): o todo o nada.
//
// El orden del panel es `sort_order, id` —el mismo que ve el visitante (P6-19)—
// y el filtro por estado de publicación es opcional (nil = todo, borradores
// incluidos). `whatsapp_url_shared` se **recalcula en cada lectura** y nunca se
// persiste (R6-11): es un aviso, no una regla.

// messageDuplicate es el mensaje del conflicto del módulo (R6-11): lo emite el
// service al detectar el duplicado exacto con `ExistsMinistryDuplicate` y el
// repository al traducir una violación de unicidad de la base.
const messageDuplicate = "Ya existe un ministerio con los mismos datos"

// InsertMinistry crea un ministerio y registra su auditoría en una única
// transacción (FR-010/FR-017). El service decide el estado de publicación
// (por defecto `draft`, «se publica al guardar») y el interruptor del encargado
// (por defecto apagado, «Oculto por defecto»): el repository solo persiste.
func (r *repository) InsertMinistry(ctx context.Context, ministry Ministry, actions ...audit.Action) (Ministry, error) {
	var result Ministry
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.InsertMinistry(ctx, insertParams(ministry))
		if err != nil {
			return wrap(err, "insert ministry", "", messageDuplicate)
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapMinistry(row)
		return nil
	})
	if err != nil {
		return Ministry{}, err
	}
	return result, nil
}

// UpdateMinistry escribe la fila COMPLETA ya fusionada que le pasa el service
// (A6-3/P6-20: la fusión de los campos presentes del `PATCH` sobre la fila
// actual vive en el service) y registra su auditoría en la misma transacción.
// Un id inexistente es `apperr.NotFound`.
func (r *repository) UpdateMinistry(ctx context.Context, ministry Ministry, actions ...audit.Action) (Ministry, error) {
	var result Ministry
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.UpdateMinistry(ctx, updateParams(ministry))
		if err != nil {
			return wrap(err, "update ministry", "El ministerio no existe", messageDuplicate)
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapMinistry(row)
		return nil
	})
	if err != nil {
		return Ministry{}, err
	}
	return result, nil
}

// DeleteMinistry borra físicamente el ministerio (sin `deleted_at`) y registra
// su auditoría en la misma transacción. Devuelve false si el id no existía.
//
// **No borra el archivo de la imagen** (A6-11): la fila deja de referenciarlo y
// la consulta compartida `IsFilePublished` deja de servirlo; la limpieza de un
// archivo reemplazado o quitado es best-effort y vive en el service.
func (r *repository) DeleteMinistry(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	deleted := false
	err := r.withTx(ctx, func(tx *repository) error {
		rows, err := tx.q.DeleteMinistry(ctx, pgUUID(id))
		if err != nil {
			return wrap(err, "delete ministry", "", "")
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

// GetMinistryByID devuelve el ministerio del panel (borradores y fichas del
// encargado no publicadas incluidas: dato guardado ≠ dato publicado) con el
// aviso de enlace compartido. El segundo valor es false si el id no existe.
func (r *repository) GetMinistryByID(ctx context.Context, id uuid.UUID) (MinistryListItem, bool, error) {
	row, err := r.q.GetMinistryByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MinistryListItem{}, false, nil
		}
		return MinistryListItem{}, false, wrap(err, "get ministry by id", "", "")
	}
	return MinistryListItem{
		Ministry:          mapMinistry(detailRow(row)),
		WhatsappURLShared: boolValue(row.WhatsappUrlShared),
	}, true, nil
}

// ListMinistries devuelve la página del panel: todos los ministerios (borradores
// incluidos) en el orden del módulo (`sort_order`, `id`). El filtro por estado
// es opcional (nil = todos).
func (r *repository) ListMinistries(ctx context.Context, state *PublicationState, params paginate.Params) ([]MinistryListItem, error) {
	rows, err := r.q.ListMinistries(ctx, gendb.ListMinistriesParams{
		PublicationState: pgStatePtr(state),
		Lim:              int32(params.Limit),
		Off:              int32(params.Offset),
	})
	if err != nil {
		return nil, wrap(err, "list ministries", "", "")
	}
	items := make([]MinistryListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, MinistryListItem{
			Ministry:          mapMinistry(listRow(row)),
			WhatsappURLShared: boolValue(row.WhatsappUrlShared),
		})
	}
	return items, nil
}

// CountMinistries cuenta los ministerios con el mismo filtro que
// `ListMinistries`, para que `total` corresponda a la página.
func (r *repository) CountMinistries(ctx context.Context, state *PublicationState) (int64, error) {
	total, err := r.q.CountMinistries(ctx, pgStatePtr(state))
	if err != nil {
		return 0, wrap(err, "count ministries", "", "")
	}
	return total, nil
}

// ExistsMinistryDuplicate detecta el duplicado exacto del módulo (R6-11): otro
// ministerio con el mismo nombre, descripción, encargado y enlace
// **normalizados** (`btrim` + minúsculas; los opcionales con
// `IS NOT DISTINCT FROM`). `excludeID` permite editar un ministerio sin que se
// detecte a sí mismo (nil al crear). Compartir el enlace con otros datos **no**
// es duplicado: es el aviso no bloqueante de `whatsapp_url_shared`.
func (r *repository) ExistsMinistryDuplicate(ctx context.Context, nameEs, descriptionEs string, leaderNameEs, whatsappURL *string, excludeID *uuid.UUID) (bool, error) {
	duplicate, err := r.q.ExistsMinistryDuplicate(ctx, gendb.ExistsMinistryDuplicateParams{
		NameEs:        nameEs,
		DescriptionEs: descriptionEs,
		LeaderNameEs:  pgTextPtr(leaderNameEs),
		WhatsappUrl:   pgTextPtr(whatsappURL),
		ExcludeID:     pgUUIDPtr(excludeID),
	})
	if err != nil {
		return false, wrap(err, "exists ministry duplicate", "", "")
	}
	return duplicate, nil
}

// --- Mapeo de la entidad a los parámetros de sqlc ---

// insertParams traduce el ministerio del dominio al alta de sqlc: los opcionales
// van como NULL cuando no existen y `leader_visible` viaja siempre explícito
// (el service aplica `false` si el DTO no trae la clave, «Oculto por defecto»).
func insertParams(ministry Ministry) gendb.InsertMinistryParams {
	return gendb.InsertMinistryParams{
		NameEs:           ministry.NameEs,
		NameEn:           pgTextPtr(ministry.NameEn),
		DescriptionEs:    ministry.DescriptionEs,
		DescriptionEn:    pgTextPtr(ministry.DescriptionEn),
		LeaderNameEs:     pgTextPtr(ministry.LeaderNameEs),
		LeaderNameEn:     pgTextPtr(ministry.LeaderNameEn),
		LeaderPhone:      pgTextPtr(ministry.LeaderPhone),
		LeaderEmail:      pgTextPtr(ministry.LeaderEmail),
		LeaderVisible:    ministry.LeaderVisible,
		WhatsappUrl:      pgTextPtr(ministry.WhatsappURL),
		ScheduleEs:       pgTextPtr(ministry.ScheduleEs),
		ScheduleEn:       pgTextPtr(ministry.ScheduleEn),
		PlaceEs:          pgTextPtr(ministry.PlaceEs),
		PlaceEn:          pgTextPtr(ministry.PlaceEn),
		ImageFile:        pgTextPtr(ministry.ImageFile),
		ImageAltEs:       pgTextPtr(ministry.ImageAltEs),
		ImageAltEn:       pgTextPtr(ministry.ImageAltEn),
		PublicationState: string(ministry.PublicationState),
		SortOrder:        int32(ministry.SortOrder),
	}
}

// updateParams traduce el ministerio ya fusionado al update de sqlc: escribe
// todas las columnas (la fila completa que recibe del service).
func updateParams(ministry Ministry) gendb.UpdateMinistryParams {
	return gendb.UpdateMinistryParams{
		NameEs:           ministry.NameEs,
		NameEn:           pgTextPtr(ministry.NameEn),
		DescriptionEs:    ministry.DescriptionEs,
		DescriptionEn:    pgTextPtr(ministry.DescriptionEn),
		LeaderNameEs:     pgTextPtr(ministry.LeaderNameEs),
		LeaderNameEn:     pgTextPtr(ministry.LeaderNameEn),
		LeaderPhone:      pgTextPtr(ministry.LeaderPhone),
		LeaderEmail:      pgTextPtr(ministry.LeaderEmail),
		LeaderVisible:    ministry.LeaderVisible,
		WhatsappUrl:      pgTextPtr(ministry.WhatsappURL),
		ScheduleEs:       pgTextPtr(ministry.ScheduleEs),
		ScheduleEn:       pgTextPtr(ministry.ScheduleEn),
		PlaceEs:          pgTextPtr(ministry.PlaceEs),
		PlaceEn:          pgTextPtr(ministry.PlaceEn),
		ImageFile:        pgTextPtr(ministry.ImageFile),
		ImageAltEs:       pgTextPtr(ministry.ImageAltEs),
		ImageAltEn:       pgTextPtr(ministry.ImageAltEn),
		PublicationState: string(ministry.PublicationState),
		SortOrder:        int32(ministry.SortOrder),
		ID:               pgUUID(ministry.ID),
	}
}

// boolValue traduce una columna BOOLEAN calculada (sqlc la tipa anulable porque
// la expresión lleva `IS NOT NULL`): NULL es "no".
func boolValue(v pgtype.Bool) bool {
	return v.Valid && v.Bool
}

// listRow traduce la fila del listado del panel al modelo generado común: sqlc
// emite un tipo por consulta aunque las columnas sean las mismas.
func listRow(row gendb.ListMinistriesRow) gendb.Ministry {
	return gendb.Ministry{
		ID:               row.ID,
		NameEs:           row.NameEs,
		NameEn:           row.NameEn,
		DescriptionEs:    row.DescriptionEs,
		DescriptionEn:    row.DescriptionEn,
		LeaderNameEs:     row.LeaderNameEs,
		LeaderNameEn:     row.LeaderNameEn,
		LeaderPhone:      row.LeaderPhone,
		LeaderEmail:      row.LeaderEmail,
		LeaderVisible:    row.LeaderVisible,
		WhatsappUrl:      row.WhatsappUrl,
		ScheduleEs:       row.ScheduleEs,
		ScheduleEn:       row.ScheduleEn,
		PlaceEs:          row.PlaceEs,
		PlaceEn:          row.PlaceEn,
		ImageFile:        row.ImageFile,
		ImageAltEs:       row.ImageAltEs,
		ImageAltEn:       row.ImageAltEn,
		PublicationState: row.PublicationState,
		SortOrder:        row.SortOrder,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

// detailRow traduce la fila del detalle del panel al modelo generado común.
func detailRow(row gendb.GetMinistryByIDRow) gendb.Ministry {
	return gendb.Ministry{
		ID:               row.ID,
		NameEs:           row.NameEs,
		NameEn:           row.NameEn,
		DescriptionEs:    row.DescriptionEs,
		DescriptionEn:    row.DescriptionEn,
		LeaderNameEs:     row.LeaderNameEs,
		LeaderNameEn:     row.LeaderNameEn,
		LeaderPhone:      row.LeaderPhone,
		LeaderEmail:      row.LeaderEmail,
		LeaderVisible:    row.LeaderVisible,
		WhatsappUrl:      row.WhatsappUrl,
		ScheduleEs:       row.ScheduleEs,
		ScheduleEn:       row.ScheduleEn,
		PlaceEs:          row.PlaceEs,
		PlaceEn:          row.PlaceEn,
		ImageFile:        row.ImageFile,
		ImageAltEs:       row.ImageAltEs,
		ImageAltEn:       row.ImageAltEn,
		PublicationState: row.PublicationState,
		SortOrder:        row.SortOrder,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}
