package ministerios

import (
	"context"

	gendb "simiente-santa/backend/internal/db"
)

// repository_public.go es la ÚNICA vía pública de los ministerios (FR-002/FR-012)
// y el lugar donde se garantiza la privacidad del encargado ANTES de llegar a
// cualquier capa superior (FR-026/SC-014, R6-2/RG6-1).
//
// La consulta exige `publication_state = 'published'` —así ningún borrador sale
// del servidor— y devuelve los cuatro campos del encargado con `CASE WHEN
// leader_visible`: con el interruptor apagado la fila **no transporta** el dato
// (llega NULL) y aquí se traduce a "sin dato". Con el interruptor apagado, ni el
// nombre ni el celular/correo existen en la entidad que sale de este paquete, de
// modo que ninguna capa posterior puede publicarlos por descuido.
//
// El orden es el manual del módulo (`sort_order`, `id`) y el tope lo pasa el
// service (`maxPublicItems = 200` en la sección, `maxPortadaItems = 5` en el
// bloque de la portada: el tope 5 es del servidor, FR-027).

// ListPublishedMinistries devuelve hasta `limit` ministerios publicados, en
// orden `sort_order, id`, con los campos del encargado enmascarados si su ficha
// no está publicada.
func (r *repository) ListPublishedMinistries(ctx context.Context, limit int) ([]Ministry, error) {
	rows, err := r.q.ListPublishedMinistries(ctx, int32(limit))
	if err != nil {
		return nil, wrap(err, "list published ministries", "", "")
	}
	items := make([]Ministry, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapPublishedMinistry(row))
	}
	return items, nil
}

// mapPublishedMinistry traduce la fila pública (sqlc) a la entidad del dominio.
//
// Las cuatro columnas del encargado llegan como `interface{}` —sqlc v1.31.1 no
// infiere el tipo de un `CASE` (nota de queries/ministries.sql)— y se traducen
// con `maskedText`: NULL (la ficha no publicada) es "sin dato" y ningún
// `interface{}` sale de este paquete (§8.1.6).
//
// El estado de publicación se fija en `published` porque la consulta solo
// devuelve filas publicadas: el DTO público no expone el estado y la entidad
// pública no tiene otro valor posible.
func mapPublishedMinistry(row gendb.ListPublishedMinistriesRow) Ministry {
	return Ministry{
		ID:               uuidValue(row.ID),
		NameEs:           row.NameEs,
		NameEn:           textPtr(row.NameEn),
		DescriptionEs:    row.DescriptionEs,
		DescriptionEn:    textPtr(row.DescriptionEn),
		LeaderNameEs:     maskedText(row.LeaderNameEs),
		LeaderNameEn:     maskedText(row.LeaderNameEn),
		LeaderPhone:      maskedText(row.LeaderPhone),
		LeaderEmail:      maskedText(row.LeaderEmail),
		LeaderVisible:    row.LeaderVisible,
		WhatsappURL:      textPtr(row.WhatsappUrl),
		ScheduleEs:       textPtr(row.ScheduleEs),
		ScheduleEn:       textPtr(row.ScheduleEn),
		PlaceEs:          textPtr(row.PlaceEs),
		PlaceEn:          textPtr(row.PlaceEn),
		ImageFile:        textPtr(row.ImageFile),
		ImageAltEs:       textPtr(row.ImageAltEs),
		ImageAltEn:       textPtr(row.ImageAltEn),
		PublicationState: StatePublished,
		SortOrder:        int(row.SortOrder),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}
