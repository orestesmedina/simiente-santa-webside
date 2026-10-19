package medios

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

// repository_episodios.go es el acceso a datos de las fichas (`media_items`):
// la sección pública, el bloque de la portada, el listado y el detalle del panel
// y las mutaciones (alta al importar, escritura completa ya fusionada y la unión
// por puntero).
//
// La separación público/panel vive en CONSULTAS distintas (FR-009/SC-002, R9-7):
// `ListPublishedMediaItems`/`CountPublishedMediaItems`/`ListPortadaMediaItems`
// son la ÚNICA vía pública y llevan el predicado público **completo**
// —`visibility='visible'` + no absorbida + sin «posible duplicado» ABIERTO—, de
// modo que una ficha oculta, absorbida o con par pendiente **no sale de la
// consulta** (no se filtra en el service). Una marca de disponibilidad sin
// resolver («ya no aparece en el feed») **sí** sale: sigue visible (FR-017).
//
// El orden es el mismo en las tres: `published_at DESC, created_at ASC, id`
// (FR-006/default revisable #11). Las mutaciones escriben sus filas de auditoría
// en la MISMA transacción (FR-022).

// FiltroPanel reúne los filtros opcionales del listado del panel (FR-019). El
// cero del tipo es «todo lo que se ve por defecto»: sin filtro de visibilidad,
// sin filtro de marca y **sin** absorbidas (la revisión avanzada pide
// `IncluirAbsorbidas`).
type FiltroPanel struct {
	// Visibilidad filtra por estado (nil = sin filtro).
	Visibilidad *Visibilidad
	// MarcaRevision filtra por la marca **pendiente** de esa clase: el par
	// abierto de `possible_duplicate` o la marca de disponibilidad sin resolver
	// (corrección de la fase 5: los dos ejes son independientes y «pendiente» se
	// pregunta por eje).
	MarcaRevision *MarcaRevision
	// IncluirAbsorbidas lista también las fichas absorbidas por una unión
	// (false por defecto: la absorbida desaparece de todas las consultas).
	IncluirAbsorbidas bool
}

// ListPublishedMediaItems devuelve la tanda de la sección pública (FR-002/FR-006)
// con el predicado público completo y el orden de la sección. `params` ya viene
// normalizado por el service (12 por defecto, tope 50: P9-12/R9-10).
func (r *repository) ListPublishedMediaItems(ctx context.Context, params paginate.Params) ([]Episodio, error) {
	rows, err := r.q.ListPublishedMediaItems(ctx, gendb.ListPublishedMediaItemsParams{
		Offset: int32(params.Offset),
		Limit:  int32(params.Limit),
	})
	if err != nil {
		return nil, wrap(err, "list published media items", "", "")
	}
	items := make([]Episodio, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapEpisodioPublico(row))
	}
	return items, nil
}

// CountPublishedMediaItems cuenta las fichas publicables con el mismo predicado
// que `ListPublishedMediaItems` (sobre de paginación «Cargar más» y testigo de
// «nada visible» → redirección de `/medios`, FR-008/SC-016).
func (r *repository) CountPublishedMediaItems(ctx context.Context) (int64, error) {
	total, err := r.q.CountPublishedMediaItems(ctx)
	if err != nil {
		return 0, wrap(err, "count published media items", "", "")
	}
	return total, nil
}

// ListPortadaMediaItems devuelve los primeros `limit` publicables del bloque de
// la portada (FR-027/SC-015): mismo predicado y mismo orden que la sección; el
// tope 5 lo pasa el service (constante del dominio, no parámetro de la API).
func (r *repository) ListPortadaMediaItems(ctx context.Context, limit int) ([]Episodio, error) {
	rows, err := r.q.ListPortadaMediaItems(ctx, int32(limit))
	if err != nil {
		return nil, wrap(err, "list portada media items", "", "")
	}
	items := make([]Episodio, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapEpisodioPortada(row))
	}
	return items, nil
}

// ListMediaItems devuelve la página del panel (FR-019): **todo** —ocultas y
// marcas incluidas— menos las absorbidas, que solo aparecen con
// `IncluirAbsorbidas`, y con el MISMO orden que ve el visitante.
func (r *repository) ListMediaItems(ctx context.Context, filtro FiltroPanel, params paginate.Params) ([]Episodio, error) {
	rows, err := r.q.ListMediaItems(ctx, gendb.ListMediaItemsParams{
		Visibility:    pgEnumPtr(filtro.Visibilidad),
		IncludeMerged: pgBool(filtro.IncluirAbsorbidas),
		ReviewMark:    pgEnumPtr(filtro.MarcaRevision),
		Offset:        int32(params.Offset),
		Limit:         int32(params.Limit),
	})
	if err != nil {
		return nil, wrap(err, "list media items", "", "")
	}
	items := make([]Episodio, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapEpisodioListado(row))
	}
	return items, nil
}

// CountMediaItems cuenta las fichas con el mismo filtro que `ListMediaItems`,
// para que `total` corresponda a la página.
func (r *repository) CountMediaItems(ctx context.Context, filtro FiltroPanel) (int64, error) {
	total, err := r.q.CountMediaItems(ctx, gendb.CountMediaItemsParams{
		Visibility:    pgEnumPtr(filtro.Visibilidad),
		IncludeMerged: pgBool(filtro.IncluirAbsorbidas),
		ReviewMark:    pgEnumPtr(filtro.MarcaRevision),
	})
	if err != nil {
		return 0, wrap(err, "count media items", "", "")
	}
	return total, nil
}

// GetMediaItemByID devuelve la ficha completa del panel —marcas de los dos ejes,
// unión y los dos enlaces de sus fuentes—. El segundo valor es false si el id no
// existe.
func (r *repository) GetMediaItemByID(ctx context.Context, id uuid.UUID) (Episodio, bool, error) {
	row, err := r.q.GetMediaItemByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Episodio{}, false, nil
		}
		return Episodio{}, false, wrap(err, "get media item by id", "", "")
	}
	return mapEpisodioDetalle(row), true, nil
}

// InsertMediaItem crea la ficha que trae el feed y registra su auditoría en una
// única transacción (FR-011/FR-012/FR-022). Los ejes de marcas y `merged_into_id`
// nacen en NULL y `description_edited` en false (DEFAULT de la tabla: la
// descripción es del feed hasta que el equipo la ajuste, FR-015); `origin` y
// `visibility` los decide el service (FR-014: la entrada adicional de un par
// dudoso nace oculta).
func (r *repository) InsertMediaItem(ctx context.Context, episodio Episodio, actions ...audit.Action) (Episodio, error) {
	var result Episodio
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.InsertMediaItem(ctx, insertMediaItemParams(episodio))
		if err != nil {
			return wrap(err, "insert media item", "", "")
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapEpisodio(row)
		return nil
	})
	if err != nil {
		return Episodio{}, err
	}
	return result, nil
}

// UpdateMediaItem escribe la fila COMPLETA ya fusionada que le pasa el service
// (descripción, visibilidad, marcas de los dos ejes, disponibilidad y unión) y
// registra su auditoría en la misma transacción. `title`, `published_at` y
// `cover_file` no son editables (FR-019) y se reescriben tal cual para que la
// instantánea de la importación no cambie (R9-6/P9-9). Un id inexistente es
// `apperr.NotFound`.
func (r *repository) UpdateMediaItem(ctx context.Context, episodio Episodio, actions ...audit.Action) (Episodio, error) {
	var result Episodio
	err := r.withTx(ctx, func(tx *repository) error {
		row, err := tx.q.UpdateMediaItem(ctx, updateMediaItemParams(episodio))
		if err != nil {
			return wrap(err, "update media item", messageEpisodioNoExiste, "")
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		result = mapEpisodio(row)
		return nil
	})
	if err != nil {
		return Episodio{}, err
	}
	return result, nil
}

// UpdateMediaItemCoverIfNull escribe la carátula copiada al sitio **solo si la
// ficha todavía no tiene** (P9-9/default #10: la carátula es una instantánea y la
// búsqueda no la refresca) y devuelve el segundo valor `false` cuando otra
// escritura se le adelantó. La condición vive en el `UPDATE`
// (`cover_file IS NULL`), sin lectura previa, así que dos búsquedas simultáneas
// no se pisan: exactamente una escribe y la otra retira su copia. La consulta
// toca únicamente `cover_file`/`updated_at`: nada del equipo se reescribe.
func (r *repository) UpdateMediaItemCoverIfNull(ctx context.Context, id uuid.UUID, archivo string) (Episodio, bool, error) {
	row, err := r.q.UpdateMediaItemCoverIfNull(ctx, gendb.UpdateMediaItemCoverIfNullParams{
		CoverFile: pgText(archivo),
		ID:        pgUUID(id),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// La ficha ya tenía carátula (o desapareció): no se escribe nada.
			return Episodio{}, false, nil
		}
		return Episodio{}, false, wrap(err, "update media item cover if null", messageEpisodioNoExiste, "")
	}
	return mapEpisodio(row), true, nil
}

// InsertMediaItemWithSource da de alta **una ficha con su fuente** en UNA sola
// transacción (H3/SC-003): es el alta de un elemento importado por la búsqueda
// (`crearFicha`). Si la fuente conflictúa —otra búsqueda simultánea importó el
// mismo `(provider, external_id)`— la transacción se **revierte** y no queda
// ninguna ficha huérfana sin enlaces; el segundo valor es `false` y el service lo
// cuenta como no importado (nada se escribió). Un fallo de la base aborta igual:
// el alta de un elemento nunca deja datos a medias.
func (r *repository) InsertMediaItemWithSource(ctx context.Context, episodio Episodio, fuente FuenteEpisodio) (Episodio, FuenteEpisodio, bool, error) {
	var (
		creada    Episodio
		insertada FuenteEpisodio
	)
	err := r.withTx(ctx, func(tx *repository) error {
		ficha, origen, err := altaFichaYFuente(ctx, tx, episodio, fuente)
		if err != nil {
			return err
		}
		creada, insertada = ficha, origen
		return nil
	})
	return resultadoAlta(creada, insertada, err)
}

// MarkDuplicatePair crea la ficha **adicional** de un par dudoso con su fuente y
// marca la **pareja recíproca** en UNA sola transacción (FR-014/H3/M1): la nueva
// apunta al candidato (lo trae el service) y el candidato apunta a la nueva, de
// modo que la invariante «cada miembro apunta al otro» no tiene ventana. Si la
// fuente conflictúa, la transacción se revierte entera: ni ficha oculta sin
// enlaces ni pareja a medias. La marca del candidato se escribe por columnas
// (`MarkMediaItemDuplicate`): su visibilidad y la descripción del equipo quedan
// fuera del alcance de la búsqueda (SC-007).
func (r *repository) MarkDuplicatePair(ctx context.Context, nueva Episodio, fuente FuenteEpisodio, candidatoID uuid.UUID) (Episodio, FuenteEpisodio, bool, error) {
	var (
		creada    Episodio
		insertada FuenteEpisodio
	)
	err := r.withTx(ctx, func(tx *repository) error {
		ficha, origen, err := altaFichaYFuente(ctx, tx, nueva, fuente)
		if err != nil {
			return err
		}
		if _, err := tx.q.MarkMediaItemDuplicate(ctx, gendb.MarkMediaItemDuplicateParams{
			DuplicateOfID: pgUUID(ficha.ID),
			ID:            pgUUID(candidatoID),
		}); err != nil {
			return wrap(err, "mark media item duplicate", messageEpisodioNoExiste, "")
		}
		creada, insertada = ficha, origen
		return nil
	})
	return resultadoAlta(creada, insertada, err)
}

// altaFichaYFuente ejecuta el alta (ficha + fuente) dentro de la transacción
// vigente: la fuente que conflictúa devuelve `errFuenteYaImportada`, que revierte
// la transacción —la ficha recién creada desaparece con ella—.
func altaFichaYFuente(ctx context.Context, tx *repository, episodio Episodio, fuente FuenteEpisodio) (Episodio, FuenteEpisodio, error) {
	row, err := tx.q.InsertMediaItem(ctx, insertMediaItemParams(episodio))
	if err != nil {
		return Episodio{}, FuenteEpisodio{}, wrap(err, "insert media item", "", "")
	}
	creada := mapEpisodio(row)

	fuente.EpisodioID = creada.ID
	srow, err := tx.q.InsertMediaEpisodeSource(ctx, insertFuenteParams(fuente))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Episodio{}, FuenteEpisodio{}, errFuenteYaImportada
		}
		return Episodio{}, FuenteEpisodio{}, wrap(err, "insert media episode source", "", messageFuenteDuplicada)
	}
	return creada, mapFuenteEpisodio(srow), nil
}

// resultadoAlta traduce el resultado de una transacción de alta: el conflicto de
// la fuente es `ok=false` (nada escrito, no es un error) y cualquier otro fallo
// sube como error.
func resultadoAlta(creada Episodio, insertada FuenteEpisodio, err error) (Episodio, FuenteEpisodio, bool, error) {
	switch {
	case errors.Is(err, errFuenteYaImportada):
		return Episodio{}, FuenteEpisodio{}, false, nil
	case err != nil:
		return Episodio{}, FuenteEpisodio{}, false, err
	default:
		return creada, insertada, true, nil
	}
}

// RepointMediaItemSources aplica la **unión por puntero** (R9-12/P9-10/FR-019)
// en UNA sola transacción: mueve las fuentes de la ficha absorbida a la
// resultante y escribe la fila completa ya fusionada de las dos —la absorbida
// queda con `merged_into_id` y oculta, y el eje de duplicados de ambas lo
// despeja el service—, más las filas de auditoría (`media.merge`). El nombre
// viene de la consulta `RepointMediaItemSources` que ejecuta el movimiento; el
// resto de la escritura viaja en el mismo tx para que la unión —la única
// operación irreversible del módulo— sea atómica y no quede a medias.
//
// Devuelve las fuentes movidas, de modo que el service comprueba el resultado
// (la resultante queda con los dos enlaces) sin una consulta extra. **Cero
// DELETE**: ninguna fuente se borra (conserva la procedencia).
func (r *repository) RepointMediaItemSources(ctx context.Context, resultante, absorbida Episodio, actions ...audit.Action) ([]FuenteEpisodio, error) {
	var movidas []FuenteEpisodio
	err := r.withTx(ctx, func(tx *repository) error {
		rows, err := tx.q.RepointMediaItemSources(ctx, gendb.RepointMediaItemSourcesParams{
			ToItemID:   pgUUID(resultante.ID),
			FromItemID: pgUUID(absorbida.ID),
		})
		if err != nil {
			return wrap(err, "repoint media item sources", "", "")
		}
		if _, err := tx.q.UpdateMediaItem(ctx, updateMediaItemParams(resultante)); err != nil {
			return wrap(err, "update media item", messageEpisodioNoExiste, "")
		}
		if _, err := tx.q.UpdateMediaItem(ctx, updateMediaItemParams(absorbida)); err != nil {
			return wrap(err, "update media item", messageEpisodioNoExiste, "")
		}
		if err := insertActions(ctx, tx, actions); err != nil {
			return err
		}
		movidas = make([]FuenteEpisodio, 0, len(rows))
		for _, row := range rows {
			movidas = append(movidas, mapFuenteEpisodio(row))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return movidas, nil
}

// --- Mapeo de la entidad a los parámetros de sqlc ---

// pgBool traduce un booleano del dominio al parámetro pgtype.Bool (el filtro
// `include_merged` de la consulta del panel: `false` = sin absorbidas).
func pgBool(v bool) pgtype.Bool {
	return pgtype.Bool{Bool: v, Valid: true}
}

// insertMediaItemParams traduce la ficha del dominio al alta de sqlc: los datos
// que trae el feed (título, fecha, descripción corta ya recortada y carátula ya
// copiada) más `origin` y `visibility` —que decide el service— y el puntero de
// par de la entrada adicional de un par dudoso (FR-014: la pareja recíproca la
// cierra la misma transacción, `MarkMediaItemDuplicate`). Los dos ejes de marcas,
// `merged_into_id` y `description_edited` nacen en su DEFAULT.
func insertMediaItemParams(episodio Episodio) gendb.InsertMediaItemParams {
	return gendb.InsertMediaItemParams{
		Title:         episodio.Titulo,
		PublishedAt:   pgTime(episodio.PublicadoEn),
		DescriptionEs: pgTextPtr(episodio.DescripcionEs),
		DescriptionEn: pgTextPtr(episodio.DescripcionEn),
		CoverFile:     pgTextPtr(episodio.Caratula),
		Origin:        string(episodio.Origen),
		Visibility:    string(episodio.Visibilidad),
		DuplicateOfID: pgUUIDPtr(episodio.DuplicadoDeID),
	}
}

// updateMediaItemParams traduce la ficha ya fusionada al update de sqlc: escribe
// todas las columnas (la fila completa que recibe del service), con las marcas
// ausentes como NULL —una marca resuelta se guarda en su columna
// `*_resolved_at` y despejar el par de duplicados escribe NULL—.
func updateMediaItemParams(episodio Episodio) gendb.UpdateMediaItemParams {
	return gendb.UpdateMediaItemParams{
		Title:                      episodio.Titulo,
		PublishedAt:                pgTime(episodio.PublicadoEn),
		DescriptionEs:              pgTextPtr(episodio.DescripcionEs),
		DescriptionEn:              pgTextPtr(episodio.DescripcionEn),
		DescriptionEdited:          episodio.DescripcionEditada,
		CoverFile:                  pgTextPtr(episodio.Caratula),
		Origin:                     string(episodio.Origen),
		Visibility:                 string(episodio.Visibilidad),
		DuplicateOfID:              pgUUIDPtr(episodio.DuplicadoDeID),
		DuplicateMarkResolvedAt:    pgTimePtr(episodio.MarcaDuplicadoResueltaEn),
		AvailabilityCheckedAt:      pgTimePtr(episodio.DisponibilidadComprobadaEn),
		AvailabilityResult:         pgEnumPtr(episodio.ResultadoDisponibilidad),
		AvailabilityMark:           pgEnumPtr(episodio.MarcaDisponibilidad),
		AvailabilityMarkResolvedAt: pgTimePtr(episodio.MarcaDisponibilidadResueltaEn),
		MergedIntoID:               pgUUIDPtr(episodio.UnidoAID),
		ID:                         pgUUID(episodio.ID),
	}
}
