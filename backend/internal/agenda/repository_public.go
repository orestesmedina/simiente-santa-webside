package agenda

import (
	"context"

	gendb "simiente-santa/backend/internal/db"
)

// repository_public.go es el acceso a datos del extracto público de la portada
// (US7/FR-025, enmienda de alcance del 2026-10-12/Q8): una sola lectura que
// cruza eventos y actividades.
//
// La regla del extracto —tope de 5 firme, ventana de 2 meses y respaldo— vive en
// la consulta `ListNextUpcomingItems` (T444, research R4-16): el service solo
// calcula el corte de "hoy"/"ahora" (R4-2) y el final de la ventana en `APP_TZ`
// y los pasa parametrizados.

// ListNextUpcomingItems devuelve hasta `limit` ítems próximos —eventos y
// actividades mezclados por cercanía— de lo publicado y próximo, aplicando la
// ventana `[today, windowEnd]` con respaldo (si la ventana está vacía, los
// siguientes próximos; nunca se mezclan). Orden: `date`, `startTime` (sin hora
// primero), `title`, `id` (FR-003).
func (r *repository) ListNextUpcomingItems(ctx context.Context, today, nowTime, windowEnd string, limit int) ([]NextItem, error) {
	rows, err := r.q.ListNextUpcomingItems(ctx, gendb.ListNextUpcomingItemsParams{
		Lim:       int32(limit),
		Today:     pgDate(today),
		NowTime:   pgTime(nowTime),
		WindowEnd: pgDate(windowEnd),
	})
	if err != nil {
		return nil, wrap(err, "list next upcoming items", "", "")
	}
	items := make([]NextItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapNextItem(row))
	}
	return items, nil
}

// mapNextItem traduce la fila del extracto (sqlc) a la entidad del dominio.
func mapNextItem(row gendb.ListNextUpcomingItemsRow) NextItem {
	return NextItem{
		Kind:       NextItemKind(row.Kind),
		ID:         uuidValue(row.ItemID),
		TitleEs:    row.TitleEs,
		TitleEn:    textPtr(row.TitleEn),
		Date:       dateString(row.ItemDate),
		StartTime:  textPtr(row.StartTime),
		EndTime:    textPtr(row.EndTime),
		PlaceEs:    row.PlaceEs,
		PlaceEn:    textPtr(row.PlaceEn),
		ImageFile:  textPtr(row.ImageFile),
		ImageAltEs: textPtr(row.ImageAltEs),
		ImageAltEn: textPtr(row.ImageAltEn),
	}
}
