package agenda

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/audit"
)

// repository_categorias.go es el acceso a datos del catálogo de categorías
// (Q6/R4-5), compartido por eventos y actividades: CRUD completo y el recuento
// de uso. Cada mutación va en una transacción con sus filas de auditoría
// (R4-10).
//
// El borrado lo protege la BD: `category_id` es ON DELETE RESTRICT, así que una
// categoría en uso falla con violación de clave foránea y se traduce a
// `conflict` (decisión firme de la puerta 2, P4-6); el nombre duplicado
// (`UNIQUE (name_es)`) también es `conflict`.

// ListCategories devuelve el catálogo completo en su orden de presentación
// (`sort_order, id`): lo usan el panel y la validación de `categoryId`.
func (r *repository) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := r.q.ListCategories(ctx)
	if err != nil {
		return nil, wrap(err, "list categories", "", "")
	}
	categories := make([]Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, mapCategory(row))
	}
	return categories, nil
}

// GetCategoryByID devuelve una categoría del catálogo. El segundo valor es
// false si el id no existe.
func (r *repository) GetCategoryByID(ctx context.Context, id uuid.UUID) (Category, bool, error) {
	row, err := r.q.GetCategoryByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Category{}, false, nil
		}
		return Category{}, false, wrap(err, "get category by id", "", "")
	}
	return mapCategory(row), true, nil
}

// InsertCategory da de alta una categoría y registra su auditoría en la misma
// transacción. Un nombre duplicado (`UNIQUE (name_es)`) es `conflict`.
func (r *repository) InsertCategory(ctx context.Context, category Category, actions ...audit.Action) (Category, error) {
	var result Category
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.InsertCategory(ctx, gendb.InsertCategoryParams{
			NameEs:    category.NameEs,
			NameEn:    pgTextPtr(category.NameEn),
			SortOrder: int32(category.SortOrder),
		})
		if err != nil {
			return wrap(err, "insert category", "", duplicateCategoryMessage)
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapCategory(row)
		return nil
	})
	if err != nil {
		return Category{}, err
	}
	return result, nil
}

// UpdateCategory renombra o reordena una categoría (renombrar está permitido:
// el cambio se ve de inmediato en lo publicado, R4-5) con su auditoría.
func (r *repository) UpdateCategory(ctx context.Context, category Category, actions ...audit.Action) (Category, error) {
	var result Category
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.UpdateCategory(ctx, gendb.UpdateCategoryParams{
			ID:        pgUUID(category.ID),
			NameEs:    category.NameEs,
			NameEn:    pgTextPtr(category.NameEn),
			SortOrder: int32(category.SortOrder),
		})
		if err != nil {
			return wrap(err, "update category", "La categoría no existe", duplicateCategoryMessage)
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapCategory(row)
		return nil
	})
	if err != nil {
		return Category{}, err
	}
	return result, nil
}

// DeleteCategory borra una categoría del catálogo y registra su auditoría en la
// misma transacción. Si algún evento o actividad la usa, la FK ON DELETE
// RESTRICT hace fallar el borrado y se traduce a `conflict` (P4-6: la categoría
// en uso NO se desliga automáticamente). Devuelve false si el id no existía.
func (r *repository) DeleteCategory(ctx context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	deleted := false
	err := r.withTx(ctx, func(tx *repository) error {
		rows, err := tx.q.DeleteCategory(ctx, pgUUID(id))
		if err != nil {
			return wrap(err, "delete category", "", categoryInUseMessage)
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

// CountCategories cuenta el catálogo sin traerlo (tope `MaxCategories`, Q6).
func (r *repository) CountCategories(ctx context.Context) (int64, error) {
	total, err := r.q.CountCategories(ctx)
	if err != nil {
		return 0, wrap(err, "count categories", "", "")
	}
	return total, nil
}

// CountCategoryUsage cuenta cuántos elementos (eventos y actividades) usan la
// categoría: es el número del mensaje "está en uso por N elementos" del 409
// (R4-5/P4-6).
func (r *repository) CountCategoryUsage(ctx context.Context, id uuid.UUID) (int64, error) {
	total, err := r.q.CountCategoryUsage(ctx, pgUUID(id))
	if err != nil {
		return 0, wrap(err, "count category usage", "", "")
	}
	return total, nil
}

// Mensajes del 409 del catálogo (R4-5): el service los usa tal cual y añade el
// detalle del recuento cuando borra una categoría en uso.
const (
	duplicateCategoryMessage = "Ya existe una categoría con ese nombre"
	categoryInUseMessage     = "La categoría está en uso por eventos o actividades"
)

// mapCategory traduce la fila de una categoría (sqlc) a la entidad del dominio.
func mapCategory(row gendb.AgendaCategory) Category {
	return Category{
		ID:        uuidValue(row.ID),
		NameEs:    row.NameEs,
		NameEn:    textPtr(row.NameEn),
		SortOrder: int(row.SortOrder),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
