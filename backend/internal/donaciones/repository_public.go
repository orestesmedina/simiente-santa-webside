package donaciones

import (
	"context"
)

// repository_public.go es la ÚNICA vía pública de métodos y textos
// (FR-012/SC-002): las dos consultas exigen `publication_state = 'published'`,
// así que ningún borrador sale del servidor —ni de métodos ni de textos—.
//
// La regla "0 borradores expuestos" está garantizada **en la consulta**, no solo
// en el servicio: no existe ninguna función que devuelva al visitante un
// borrador. El orden de los métodos es el de la sección (`CASE kind` → iban
// primero, luego `sort_order`, `id`) y el tope lo pasa el service
// (`maxPublicItems = 200` en la sección, `maxPortadaItems = 5` en el bloque: el
// tope 5 es del servidor, FR-022).

// ListPublishedDonationMethods devuelve hasta `limit` métodos publicados en el
// orden de la sección (iban primero, luego sinpe; cada grupo por `sort_order`,
// `id`).
func (r *repository) ListPublishedDonationMethods(ctx context.Context, limit int) ([]MetodoDonacion, error) {
	rows, err := r.q.ListPublishedDonationMethods(ctx, int32(limit))
	if err != nil {
		return nil, wrap(err, "list published donation methods", "", "")
	}
	items := make([]MetodoDonacion, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPublishedMetodoDonacion(row))
	}
	return items, nil
}

// ListPublishedDonationTexts devuelve los textos publicados, en orden de clave
// (`intro` < `usage`); el service los empaqueta por clave (P7-12).
func (r *repository) ListPublishedDonationTexts(ctx context.Context) ([]TextoDonacion, error) {
	rows, err := r.q.ListPublishedDonationTexts(ctx)
	if err != nil {
		return nil, wrap(err, "list published donation texts", "", "")
	}
	items := make([]TextoDonacion, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapTextoDonacion(row))
	}
	return items, nil
}
