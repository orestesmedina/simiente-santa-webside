package donaciones

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/audit"
)

// repository_textos.go es el acceso a datos de los dos textos libres de la
// sección (FR-021): el listado del panel, el detalle por clave y el guardado
// completo del singleton por clave (`UpsertDonationText`). El panel ve los
// borradores (dato guardado ≠ dato publicado); la lectura pública vive en
// `repository_public.go`.
//
// No hay operación de borrado (P7-17): un texto se retira con
// `publication_state = 'draft'`. Cada guardado escribe su fila de auditoría en
// la MISMA transacción que el cambio (FR-017/R7-10).

// ListDonationTexts devuelve las filas de texto del panel (existentes o no: una
// clave sin guardar simplemente no aparece) en orden de clave.
func (r *repository) ListDonationTexts(ctx context.Context) ([]TextoDonacion, error) {
	rows, err := r.q.ListDonationTexts(ctx)
	if err != nil {
		return nil, wrap(err, "list donation texts", "", "")
	}
	items := make([]TextoDonacion, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapTextoDonacion(row))
	}
	return items, nil
}

// GetDonationTextByKey devuelve el texto de la clave (borradores incluidos). El
// segundo valor es false si la clave no se ha guardado nunca.
func (r *repository) GetDonationTextByKey(ctx context.Context, key TextKey) (TextoDonacion, bool, error) {
	row, err := r.q.GetDonationTextByKey(ctx, string(key))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TextoDonacion{}, false, nil
		}
		return TextoDonacion{}, false, wrap(err, "get donation text by key", "", "")
	}
	return mapTextoDonacion(row), true, nil
}

// UpsertDonationText crea o reemplaza el singleton de la clave y registra su
// auditoría en la misma transacción (FR-017/FR-021). La consulta absorbe el
// `UNIQUE (text_key)`: guardar dos veces la misma clave actualiza en vez de
// duplicar (patrón del singleton de F3).
func (r *repository) UpsertDonationText(ctx context.Context, text TextoDonacion, actions ...audit.Action) (TextoDonacion, error) {
	var result TextoDonacion
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.UpsertDonationText(ctx, upsertTextParams(text))
		if err != nil {
			return wrap(err, "upsert donation text", "", "")
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapTextoDonacion(row)
		return nil
	})
	if err != nil {
		return TextoDonacion{}, err
	}
	return result, nil
}

// upsertTextParams traduce el texto del dominio al upsert de sqlc: `body_en`
// opcional va como NULL cuando no existe y el estado es propio del texto (Q10).
func upsertTextParams(text TextoDonacion) gendb.UpsertDonationTextParams {
	return gendb.UpsertDonationTextParams{
		TextKey:          string(text.TextKey),
		BodyEs:           text.BodyEs,
		BodyEn:           pgTextPtr(text.BodyEn),
		PublicationState: string(text.PublicationState),
	}
}
