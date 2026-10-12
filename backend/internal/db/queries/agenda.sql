-- Consultas del dominio `agenda` (F4): eventos con lista de fechas, actividades
-- con fecha única y el catálogo de categorías compartido.
-- Convenciones: skill `postgres-db` y
-- specs/004-eventos-actividades/data-model.md (fuente de verdad).
--
-- Columnas listadas explícitamente (sin `SELECT *`), parámetros nombrados
-- (`sqlc.arg`/`sqlc.narg`) y `ORDER BY` determinista (§8.1.7).
--
-- Separación público/panel en consultas distintas (data-model):
--   * `…Upcoming` es la ÚNICA vía pública: exige `publication_state =
--     'published'` Y la condición "próximo" (R4-2) en la MISMA consulta, con
--     `today`/`now_time` que el service calcula en la zona `APP_TZ` y pasa
--     PARAMETRIZADOS (nunca concatenados, §IV/CWE-89).
--   * Las del panel (sin sufijo) son las únicas que devuelven borradores y
--     vencidos, y viven detrás de su permiso (FR-013/SC-002/SC-009).
--
-- Regla "próximo" (R4-2), literal en cada consulta:
--     (fecha > today) OR (fecha = today AND (end_time IS NULL OR end_time >= now_time))
--
-- La auditoría NO se define aquí: se reutiliza `InsertAdminAction` de
-- `queries/audit.sql` dentro de la transacción de cada mutación (R4-10).

-- ------------------------------------------------------------- categorías

-- name: ListCategories :many
-- Catálogo completo para el panel y para la validación de `categoryId`.
SELECT id, name_es, name_en, sort_order, created_at, updated_at
FROM agenda_categories
ORDER BY sort_order, id;

-- name: GetCategoryByID :one
SELECT id, name_es, name_en, sort_order, created_at, updated_at
FROM agenda_categories
WHERE id = sqlc.arg('id');

-- name: InsertCategory :one
INSERT INTO agenda_categories (name_es, name_en, sort_order)
VALUES (sqlc.arg('name_es'), sqlc.narg('name_en'), sqlc.arg('sort_order'))
RETURNING id, name_es, name_en, sort_order, created_at, updated_at;

-- name: UpdateCategory :one
-- Renombrar está permitido: el cambio se ve en lo publicado de inmediato
-- (R4-5). `UNIQUE (name_es)` → el service traduce el duplicado a `409`.
UPDATE agenda_categories
SET name_es    = sqlc.arg('name_es'),
    name_en    = sqlc.narg('name_en'),
    sort_order = sqlc.arg('sort_order'),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, name_es, name_en, sort_order, created_at, updated_at;

-- name: DeleteCategory :execrows
-- El borrado lo protege la BD: `category_id` es `ON DELETE RESTRICT` (R4-5),
-- así que una categoría en uso falla con violación de FK y el service la
-- traduce a `409 conflict` ("está en uso por N elementos", P4-6).
DELETE FROM agenda_categories
WHERE id = sqlc.arg('id');

-- name: CountCategoryUsage :one
-- Recuento de usos en AMBOS módulos para el mensaje del `409` (R4-5/P4-6).
SELECT (
    (SELECT count(*) FROM agenda_events ev WHERE ev.category_id = sqlc.arg('id')::uuid)
    + (SELECT count(*) FROM agenda_activities ac WHERE ac.category_id = sqlc.arg('id')::uuid)
)::bigint AS total;

-- name: CountCategories :one
-- Tope del catálogo (Q6): cuenta las categorías sin traerlas todas (el service
-- lo compara con `MaxCategories`). El tope es blando: dos altas concurrentes
-- podrían superarlo en una fila.
SELECT count(*)::bigint AS total
FROM agenda_categories;

-- ----------------------------------------------------------------- eventos

-- name: InsertEvent :one
INSERT INTO agenda_events (
    title_es, title_en, frequency_es, frequency_en, detail_es, detail_en,
    place_es, place_en, map_url, organizer_es, organizer_en, category_id,
    image_file, image_alt_es, image_alt_en, publication_state
) VALUES (
    sqlc.arg('title_es'), sqlc.narg('title_en'), sqlc.narg('frequency_es'),
    sqlc.narg('frequency_en'), sqlc.arg('detail_es'), sqlc.narg('detail_en'),
    sqlc.arg('place_es'), sqlc.narg('place_en'), sqlc.narg('map_url'),
    sqlc.narg('organizer_es'), sqlc.narg('organizer_en'), sqlc.narg('category_id'),
    sqlc.narg('image_file'), sqlc.narg('image_alt_es'), sqlc.narg('image_alt_en'),
    sqlc.arg('publication_state')
)
RETURNING id, title_es, title_en, frequency_es, frequency_en, detail_es, detail_en,
          place_es, place_en, map_url, organizer_es, organizer_en, category_id,
          image_file, image_alt_es, image_alt_en, publication_state,
          created_at, updated_at;

-- name: UpdateEvent :one
-- Reemplazo completo de los campos mutables (el service fusiona el PATCH y
-- valida antes de llamar, R4-11).
UPDATE agenda_events
SET title_es          = sqlc.arg('title_es'),
    title_en          = sqlc.narg('title_en'),
    frequency_es      = sqlc.narg('frequency_es'),
    frequency_en      = sqlc.narg('frequency_en'),
    detail_es         = sqlc.arg('detail_es'),
    detail_en         = sqlc.narg('detail_en'),
    place_es          = sqlc.arg('place_es'),
    place_en          = sqlc.narg('place_en'),
    map_url           = sqlc.narg('map_url'),
    organizer_es      = sqlc.narg('organizer_es'),
    organizer_en      = sqlc.narg('organizer_en'),
    category_id       = sqlc.narg('category_id'),
    image_file        = sqlc.narg('image_file'),
    image_alt_es      = sqlc.narg('image_alt_es'),
    image_alt_en      = sqlc.narg('image_alt_en'),
    publication_state = sqlc.arg('publication_state'),
    updated_at        = now()
WHERE id = sqlc.arg('id')
RETURNING id, title_es, title_en, frequency_es, frequency_en, detail_es, detail_en,
          place_es, place_en, map_url, organizer_es, organizer_en, category_id,
          image_file, image_alt_es, image_alt_en, publication_state,
          created_at, updated_at;

-- name: DeleteEvent :execrows
-- Borrado físico; las fechas caen por `ON DELETE CASCADE`.
DELETE FROM agenda_events
WHERE id = sqlc.arg('id');

-- name: GetEventByID :one
SELECT id, title_es, title_en, frequency_es, frequency_en, detail_es, detail_en,
       place_es, place_en, map_url, organizer_es, organizer_en, category_id,
       image_file, image_alt_es, image_alt_en, publication_state,
       created_at, updated_at
FROM agenda_events
WHERE id = sqlc.arg('id');

-- name: ListEvents :many
-- Panel paginado (P4-9): `publication_state` opcional (NULL = sin filtro).
SELECT id, title_es, title_en, frequency_es, frequency_en, detail_es, detail_en,
       place_es, place_en, map_url, organizer_es, organizer_en, category_id,
       image_file, image_alt_es, image_alt_en, publication_state,
       created_at, updated_at
FROM agenda_events
WHERE (sqlc.narg('publication_state')::text IS NULL
       OR publication_state = sqlc.narg('publication_state')::text)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountEvents :one
-- Mismo filtro que `ListEvents` para que `total` corresponda a la página.
SELECT count(*)::bigint AS total
FROM agenda_events
WHERE (sqlc.narg('publication_state')::text IS NULL
       OR publication_state = sqlc.narg('publication_state')::text);

-- name: ListUpcomingEvents :many
-- Público: `published` Y con ≥ 1 fecha próxima (R4-2). Orden por la primera
-- fecha próxima con desempate estable `title_es, id` (FR-003; edge case de dos
-- elementos en la misma fecha). El tope `lim` es `maxBlockItems` (RG4-8).
WITH upcoming AS (
    SELECT e.id AS event_id, min(d.event_date)::date AS next_date
    FROM agenda_events e
    JOIN agenda_event_dates d ON d.event_id = e.id
    WHERE e.publication_state = 'published'
      AND (
          d.event_date > sqlc.arg('today')::date
          OR (d.event_date = sqlc.arg('today')::date
              AND (d.end_time IS NULL OR d.end_time::time >= sqlc.arg('now_time')::time))
      )
    GROUP BY e.id
)
SELECT e.id, e.title_es, e.title_en, e.frequency_es, e.frequency_en,
       e.detail_es, e.detail_en, e.place_es, e.place_en, e.map_url,
       e.organizer_es, e.organizer_en, e.category_id, e.image_file,
       e.image_alt_es, e.image_alt_en, e.publication_state,
       e.created_at, e.updated_at, u.next_date
FROM agenda_events e
JOIN upcoming u ON u.event_id = e.id
ORDER BY u.next_date, e.title_es, e.id
LIMIT sqlc.arg('lim');

-- ------------------------------------------------------- fechas del evento

-- name: DeleteEventDates :exec
-- Primer paso del reemplazo COMPLETO de la lista de fechas (R4-11): el
-- repository (`ReplaceEventDates`, T412) ejecuta esta consulta y la siguiente
-- DENTRO de la transacción del guardado. Van separadas a propósito: un
-- `DELETE` + `INSERT` en una sola sentencia (o en un CTE) choca con el índice
-- `UNIQUE (event_id, event_date)` al reinsertar una fecha que ya existía,
-- porque las restricciones se comprueban contra la instantánea previa a la
-- sentencia. En dos sentencias de la misma transacción el borrado sí es
-- visible para la comprobación del `INSERT`.
DELETE FROM agenda_event_dates
WHERE event_id = sqlc.arg('event_id');

-- name: InsertEventDates :exec
-- Segundo paso del reemplazo: inserta la lista completa de fechas del evento.
-- Las horas se pasan como texto y un valor vacío significa "sin hora"
-- (`NULLIF` → NULL, coherente con el CHECK `end_time IS NULL OR …`). Una lista
-- vacía no inserta nada (el evento queda sin fechas; el service lo rechaza).
INSERT INTO agenda_event_dates (event_id, event_date, start_time, end_time)
SELECT sqlc.arg('event_id'), d.event_date,
       NULLIF(d.start_time, ''), NULLIF(d.end_time, '')
FROM ROWS FROM (
    unnest(sqlc.arg('event_dates')::date[]),
    unnest(sqlc.arg('start_times')::text[]),
    unnest(sqlc.arg('end_times')::text[])
) AS d(event_date, start_time, end_time);

-- name: ListEventDates :many
-- Panel: todas las fechas del evento (el service calcula `isPast`).
SELECT id, event_id, event_date, start_time, end_time, created_at, updated_at
FROM agenda_event_dates
WHERE event_id = sqlc.arg('event_id')
ORDER BY event_date, start_time NULLS FIRST, id;

-- name: ListEventDatesForEvents :many
-- Panel (lote): todas las fechas de VARIOS eventos en UNA sola consulta, para
-- que el listado paginado no haga una consulta por fila (N+1). El orden es por
-- `event_id` y luego ascendente por fecha, así que el service agrupa por evento
-- conservando el orden de `ListEventDates`.
SELECT id, event_id, event_date, start_time, end_time, created_at, updated_at
FROM agenda_event_dates
WHERE event_id = ANY(sqlc.arg('event_ids')::uuid[])
ORDER BY event_id, event_date, start_time NULLS FIRST, id;

-- name: ListUpcomingEventDatesForEvents :many
-- Público (lote): solo las fechas PRÓXIMAS de VARIOS eventos en UNA sola
-- consulta (regla R4-2, única implementación del filtro para las fechas
-- públicas; sustituye a la variante por evento). La primera fecha de cada evento
-- es "la próxima fecha" (FR-004).
SELECT id, event_id, event_date, start_time, end_time, created_at, updated_at
FROM agenda_event_dates
WHERE event_id = ANY(sqlc.arg('event_ids')::uuid[])
  AND (
      event_date > sqlc.arg('today')::date
      OR (event_date = sqlc.arg('today')::date
          AND (end_time IS NULL OR end_time::time >= sqlc.arg('now_time')::time))
  )
ORDER BY event_id, event_date, start_time NULLS FIRST, id;

-- name: ExistsEventDuplicate :one
-- Duplicado DENTRO del módulo (R4-9): mismo título y lugar normalizados
-- (minúsculas + `btrim`) y al menos una fecha en común. `exclude_event_id`
-- permite editarlo sin detectarse a sí mismo. Un evento y una actividad con
-- datos idénticos NUNCA son duplicados (Decisión 3).
SELECT EXISTS (
    SELECT 1
    FROM agenda_events e
    JOIN agenda_event_dates d ON d.event_id = e.id
    WHERE lower(btrim(e.title_es)) = lower(btrim(sqlc.arg('title_es')::text))
      AND lower(btrim(e.place_es)) = lower(btrim(sqlc.arg('place_es')::text))
      AND d.event_date = ANY(sqlc.arg('event_dates')::date[])
      AND (sqlc.narg('exclude_event_id')::uuid IS NULL
           OR e.id <> sqlc.narg('exclude_event_id')::uuid)
) AS duplicate;

-- ------------------------------------------------------------- actividades

-- name: InsertActivity :one
INSERT INTO agenda_activities (
    title_es, title_en, detail_es, detail_en, place_es, place_en, map_url,
    organizer_es, organizer_en, category_id, image_file, image_alt_es,
    image_alt_en, activity_date, start_time, end_time, publication_state
) VALUES (
    sqlc.arg('title_es'), sqlc.narg('title_en'), sqlc.arg('detail_es'),
    sqlc.narg('detail_en'), sqlc.arg('place_es'), sqlc.narg('place_en'),
    sqlc.narg('map_url'), sqlc.narg('organizer_es'), sqlc.narg('organizer_en'),
    sqlc.narg('category_id'), sqlc.narg('image_file'), sqlc.narg('image_alt_es'),
    sqlc.narg('image_alt_en'), sqlc.arg('activity_date'), sqlc.narg('start_time'),
    sqlc.narg('end_time'), sqlc.arg('publication_state')
)
RETURNING id, title_es, title_en, detail_es, detail_en, place_es, place_en,
          map_url, organizer_es, organizer_en, category_id, image_file,
          image_alt_es, image_alt_en, activity_date, start_time, end_time,
          publication_state, created_at, updated_at;

-- name: UpdateActivity :one
UPDATE agenda_activities
SET title_es          = sqlc.arg('title_es'),
    title_en          = sqlc.narg('title_en'),
    detail_es         = sqlc.arg('detail_es'),
    detail_en         = sqlc.narg('detail_en'),
    place_es          = sqlc.arg('place_es'),
    place_en          = sqlc.narg('place_en'),
    map_url           = sqlc.narg('map_url'),
    organizer_es      = sqlc.narg('organizer_es'),
    organizer_en      = sqlc.narg('organizer_en'),
    category_id       = sqlc.narg('category_id'),
    image_file        = sqlc.narg('image_file'),
    image_alt_es      = sqlc.narg('image_alt_es'),
    image_alt_en      = sqlc.narg('image_alt_en'),
    activity_date     = sqlc.arg('activity_date'),
    start_time        = sqlc.narg('start_time'),
    end_time          = sqlc.narg('end_time'),
    publication_state = sqlc.arg('publication_state'),
    updated_at        = now()
WHERE id = sqlc.arg('id')
RETURNING id, title_es, title_en, detail_es, detail_en, place_es, place_en,
          map_url, organizer_es, organizer_en, category_id, image_file,
          image_alt_es, image_alt_en, activity_date, start_time, end_time,
          publication_state, created_at, updated_at;

-- name: DeleteActivity :execrows
DELETE FROM agenda_activities
WHERE id = sqlc.arg('id');

-- name: GetActivityByID :one
SELECT id, title_es, title_en, detail_es, detail_en, place_es, place_en,
       map_url, organizer_es, organizer_en, category_id, image_file,
       image_alt_es, image_alt_en, activity_date, start_time, end_time,
       publication_state, created_at, updated_at
FROM agenda_activities
WHERE id = sqlc.arg('id');

-- name: ListActivities :many
-- Panel paginado (P4-9): `publication_state` opcional (NULL = sin filtro).
SELECT id, title_es, title_en, detail_es, detail_en, place_es, place_en,
       map_url, organizer_es, organizer_en, category_id, image_file,
       image_alt_es, image_alt_en, activity_date, start_time, end_time,
       publication_state, created_at, updated_at
FROM agenda_activities
WHERE (sqlc.narg('publication_state')::text IS NULL
       OR publication_state = sqlc.narg('publication_state')::text)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountActivities :one
-- Mismo filtro que `ListActivities` para que `total` corresponda a la página.
SELECT count(*)::bigint AS total
FROM agenda_activities
WHERE (sqlc.narg('publication_state')::text IS NULL
       OR publication_state = sqlc.narg('publication_state')::text);

-- name: ListUpcomingActivities :many
-- Público: `published` Y con fecha próxima (R4-2). La actividad es de fecha
-- única (Decisión 3); el orden es cronológico con desempate estable. El tope
-- `lim` es `maxBlockItems` (RG4-8).
SELECT id, title_es, title_en, detail_es, detail_en, place_es, place_en,
       map_url, organizer_es, organizer_en, category_id, image_file,
       image_alt_es, image_alt_en, activity_date, start_time, end_time,
       publication_state, created_at, updated_at
FROM agenda_activities
WHERE publication_state = 'published'
  AND (
      activity_date > sqlc.arg('today')::date
      OR (activity_date = sqlc.arg('today')::date
          AND (end_time IS NULL OR end_time::time >= sqlc.arg('now_time')::time))
  )
ORDER BY activity_date, start_time NULLS FIRST, title_es, id
LIMIT sqlc.arg('lim');
