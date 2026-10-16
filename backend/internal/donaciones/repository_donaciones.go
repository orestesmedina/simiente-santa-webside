package donaciones

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
)

// repository_donaciones.go es el acceso a datos de la gestión del panel de los
// métodos de donación (FR-010/FR-014/FR-015/FR-016/FR-019): el alta, la edición,
// el borrado físico, el detalle, el listado paginado con su total y la detección
// del duplicado exacto. Cada mutación escribe su fila (o filas) de auditoría en
// la MISMA transacción que el cambio (FR-017/R7-10): o todo o nada.
//
// El orden del panel es `CASE kind … END, sort_order, id` —el mismo que ve el
// visitante (P7-20/US2 esc. 10)— y el filtro por estado de publicación es
// opcional (nil = todo, borradores incluidos).

// messageDuplicate es el mensaje del conflicto del módulo (R7-9/P7-10): lo
// emite el service al detectar el duplicado exacto con
// `ExistsDonationMethodDuplicate` y el repository al traducir una violación de
// unicidad de la base (hoy la tabla no tiene `UNIQUE (kind, number)`, pero la
// base sigue siendo la última línea de defensa).
const messageDuplicate = "Ya existe un método de donación con el mismo tipo y número"

// InsertDonationMethod crea un método y registra su auditoría en una única
// transacción (FR-010/FR-017). El service decide el estado de publicación (por
// defecto `draft`, «se publica al guardar») y normaliza el número antes de
// llamar (FR-015/FR-016): el repository solo persiste.
func (r *repository) InsertDonationMethod(ctx context.Context, method MetodoDonacion, actions ...audit.Action) (MetodoDonacion, error) {
	var result MetodoDonacion
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.InsertDonationMethod(ctx, insertMethodParams(method))
		if err != nil {
			return wrap(err, "insert donation method", "", messageDuplicate)
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapMetodoDonacion(row)
		return nil
	})
	if err != nil {
		return MetodoDonacion{}, err
	}
	return result, nil
}

// UpdateDonationMethod escribe la fila COMPLETA ya fusionada que le pasa el
// service (R7-15: la fusión de los campos presentes del `PATCH` sobre la fila
// actual vive en el service) y registra su auditoría en la misma transacción.
// Un id inexistente es `apperr.NotFound`.
func (r *repository) UpdateDonationMethod(ctx context.Context, method MetodoDonacion, actions ...audit.Action) (MetodoDonacion, error) {
	var result MetodoDonacion
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.UpdateDonationMethod(ctx, updateMethodParams(method))
		if err != nil {
			return wrap(err, "update donation method", "El método de donación no existe", messageDuplicate)
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapMetodoDonacion(row)
		return nil
	})
	if err != nil {
		return MetodoDonacion{}, err
	}
	return result, nil
}

// DeleteDonationMethod borra físicamente el método (sin `deleted_at`) y registra
// su auditoría en la misma transacción. Devuelve false si el id no existía.
func (r *repository) DeleteDonationMethod(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	deleted := false
	err := r.withTx(ctx, func(tx *repository) error {
		rows, err := tx.q.DeleteDonationMethod(ctx, pgUUID(id))
		if err != nil {
			return wrap(err, "delete donation method", "", "")
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

// GetDonationMethodByID devuelve el método del panel (borradores incluidos:
// dato guardado ≠ dato publicado). El segundo valor es false si el id no existe.
func (r *repository) GetDonationMethodByID(ctx context.Context, id uuid.UUID) (MetodoDonacion, bool, error) {
	row, err := r.q.GetDonationMethodByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MetodoDonacion{}, false, nil
		}
		return MetodoDonacion{}, false, wrap(err, "get donation method by id", "", "")
	}
	return mapMetodoDonacion(row), true, nil
}

// ListDonationMethods devuelve la página del panel: todos los métodos
// (borradores incluidos) en el orden del módulo (`CASE kind`, `sort_order`,
// `id`). El filtro por estado es opcional (nil = todos).
func (r *repository) ListDonationMethods(ctx context.Context, state *PublicationState, params paginate.Params) ([]MetodoDonacion, error) {
	rows, err := r.q.ListDonationMethods(ctx, gendb.ListDonationMethodsParams{
		PublicationState: pgStatePtr(state),
		Off:              int32(params.Offset),
		Lim:              int32(params.Limit),
	})
	if err != nil {
		return nil, wrap(err, "list donation methods", "", "")
	}
	items := make([]MetodoDonacion, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapMetodoDonacion(row))
	}
	return items, nil
}

// CountDonationMethods cuenta los métodos con el mismo filtro que
// `ListDonationMethods`, para que `total` corresponda a la página.
func (r *repository) CountDonationMethods(ctx context.Context, state *PublicationState) (int64, error) {
	total, err := r.q.CountDonationMethods(ctx, pgStatePtr(state))
	if err != nil {
		return 0, wrap(err, "count donation methods", "", "")
	}
	return total, nil
}

// ExistsDonationMethodDuplicate detecta el duplicado exacto del módulo (R7-9):
// otro método con el MISMO `kind` y el MISMO `number` **normalizado** (la
// igualdad es simple porque el número se guarda normalizado; el service lo
// normaliza antes de preguntar). `excludeID` permite editar un método sin que se
// detecte a sí mismo (nil al crear; su `id` en la edición): editar sin cambiar
// el número **no** devuelve `409`, y sí lo devuelve cuando el número coincide con
// el de **otro** método.
func (r *repository) ExistsDonationMethodDuplicate(ctx context.Context, kind MetodoKind, number string, excludeID *uuid.UUID) (bool, error) {
	duplicate, err := r.q.ExistsDonationMethodDuplicate(ctx, gendb.ExistsDonationMethodDuplicateParams{
		Kind:      string(kind),
		Number:    number,
		ExcludeID: pgUUIDPtr(excludeID),
	})
	if err != nil {
		return false, wrap(err, "exists donation method duplicate", "", "")
	}
	return duplicate, nil
}

// --- Mapeo de la entidad a los parámetros de sqlc ---

// insertMethodParams traduce el método del dominio al alta de sqlc: los
// opcionales van como NULL cuando no existen y `currency` solo se rellena en
// IBAN (el CHECK doble de la tabla lo garantiza, Q2/FR-014).
func insertMethodParams(method MetodoDonacion) gendb.InsertDonationMethodParams {
	return gendb.InsertDonationMethodParams{
		Kind:             string(method.Kind),
		Number:           method.Number,
		Institution:      method.Institution,
		HolderEs:         pgTextPtr(method.HolderEs),
		HolderEn:         pgTextPtr(method.HolderEn),
		Currency:         pgCurrencyPtr(method.Currency),
		PublicationState: string(method.PublicationState),
		SortOrder:        int32(method.SortOrder),
	}
}

// updateMethodParams traduce el método ya fusionado al update de sqlc: escribe
// todas las columnas (la fila completa que recibe del service).
func updateMethodParams(method MetodoDonacion) gendb.UpdateDonationMethodParams {
	return gendb.UpdateDonationMethodParams{
		Kind:             string(method.Kind),
		Number:           method.Number,
		Institution:      method.Institution,
		HolderEs:         pgTextPtr(method.HolderEs),
		HolderEn:         pgTextPtr(method.HolderEn),
		Currency:         pgCurrencyPtr(method.Currency),
		PublicationState: string(method.PublicationState),
		SortOrder:        int32(method.SortOrder),
		ID:               pgUUID(method.ID),
	}
}
