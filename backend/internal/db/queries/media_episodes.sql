-- Consultas del dominio `medios` (F9): las fichas por sermón (`media_items`), sus
-- fuentes y el repunte de la unión. Convenciones: skill `postgres-db` y
-- specs/009-medios-predicas-y-podcasts/data-model.md (fuente de verdad).
--
-- Columnas listadas explícitamente (sin `SELECT *`), parámetros nombrados
-- (`sqlc.arg`/`sqlc.narg`) y `ORDER BY` determinista
-- `published_at DESC, created_at ASC, id` (§8.1.7; FR-006/default revisable #11:
-- del más reciente al más antiguo y, ante la misma fecha, orden de importación
-- FIFO; el panel ve **el mismo** orden que el visitante).
--
-- Separación público/panel en consultas DISTINTAS (FR-009/SC-002, R9-7):
--   * `ListPublishedMediaItems` / `CountPublishedMediaItems` /
--     `ListPortadaMediaItems` son la ÚNICA vía pública: llevan el predicado
--     público **completo** —`visibility = 'visible'` + no absorbida
--     (`merged_into_id IS NULL`) + sin «posible duplicado» ABIERTO
--     (`duplicate_of_id IS NULL OR duplicate_mark_resolved_at IS NOT NULL`)—,
--     de modo que las fichas ocultas, las absorbidas y los pares pendientes
--     **no salen de la consulta**, no se filtran en el service. Una marca de
--     disponibilidad sin resolver («ya no aparece en el feed») **sí** sale
--     (FR-017: sigue visible).
--   * Las del panel (`ListMediaItems`, `GetMediaItemByID`) son las únicas que
--     ven ocultas, absorbidas y marcas, y viven detrás del permiso `medios`.
--
-- Los dos enlaces de la ficha salen como `video_url` / `audio_url`: un
-- `LEFT JOIN` por proveedor sobre `media_episode_sources` (`vy` = YouTube,
-- `vp` = podcast). El DTO público y el de panel del contrato los piden
-- (`MedioPublico`/`MedioAdmin`, FR-002/FR-019) y así la tanda de 12 fichas se
-- resuelve en UNA consulta (sin N+1); la invariante «una ficha = 1 video y/o
-- 1 audio» la mantiene el service (T920), de modo que el join no multiplica
-- filas. Se usa `LEFT JOIN` —y no una subconsulta escalar— porque sqlc tipa la
-- ausencia como `pgtype.Text` anulable (con la subconsulta la inferiría como
-- `string` no anulable y la lectura de una ficha de solo audio fallaría). La
-- procedencia completa (proveedor, identificador externo, título del feed) la
-- trae `ListMediaItemSourcesByItem`.
--
-- La auditoría NO se define aquí: se reutiliza `InsertAdminAction` de
-- `queries/audit.sql` dentro de la transacción de cada mutación (R7-10/P9-14);
-- no hay ninguna escritura sobre `admin_actions` en este archivo.

-- name: ListPublishedMediaItems :many
-- Público (sección `/medios`): solo fichas **publicables** (predicado de R9-7),
-- en el orden de la sección y con la tanda `limit`/`offset` (12 por defecto, tope
-- 50: los aplica el service, P9-12/R9-10). `visibility` y las marcas no se
-- seleccionan: el DTO público nunca las expone y la condición ya está en el
-- `WHERE` (misma defensa en profundidad que F6/F7).
SELECT m.id,
       m.title,
       m.published_at,
       m.description_es,
       m.description_en,
       m.cover_file,
       m.origin,
       vy.url AS video_url,
       vp.url AS audio_url,
       m.created_at,
       m.updated_at
FROM media_items m
LEFT JOIN media_episode_sources vy
       ON vy.item_id = m.id AND vy.provider = 'youtube'
LEFT JOIN media_episode_sources vp
       ON vp.item_id = m.id AND vp.provider = 'podcast'
WHERE m.visibility = 'visible'
  AND m.merged_into_id IS NULL
  AND (m.duplicate_of_id IS NULL OR m.duplicate_mark_resolved_at IS NOT NULL)
ORDER BY m.published_at DESC, m.created_at ASC, m.id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountPublishedMediaItems :one
-- Total del listado público (sobre de paginación «Cargar más» y testigo de «nada
-- visible» → redirección de `/medios`, FR-008/SC-016): **mismo predicado** que
-- `ListPublishedMediaItems` para que `total` corresponda a la sección.
SELECT count(*)::bigint AS total
FROM media_items m
WHERE m.visibility = 'visible'
  AND m.merged_into_id IS NULL
  AND (m.duplicate_of_id IS NULL OR m.duplicate_mark_resolved_at IS NOT NULL);

-- name: ListPortadaMediaItems :many
-- Bloque de la portada (FR-027/SC-015): **mismo predicado público** y mismo orden,
-- acotado por `limit` (el service pasa `maxPortada = 5`: el tope 5 es del
-- SERVIDOR y no hay parámetro de límite en la API pública del bloque).
SELECT m.id,
       m.title,
       m.published_at,
       m.description_es,
       m.description_en,
       m.cover_file,
       m.origin,
       vy.url AS video_url,
       vp.url AS audio_url,
       m.created_at,
       m.updated_at
FROM media_items m
LEFT JOIN media_episode_sources vy
       ON vy.item_id = m.id AND vy.provider = 'youtube'
LEFT JOIN media_episode_sources vp
       ON vp.item_id = m.id AND vp.provider = 'podcast'
WHERE m.visibility = 'visible'
  AND m.merged_into_id IS NULL
  AND (m.duplicate_of_id IS NULL OR m.duplicate_mark_resolved_at IS NOT NULL)
ORDER BY m.published_at DESC, m.created_at ASC, m.id
LIMIT sqlc.arg('limit');

-- name: ListMediaItems :many
-- Panel paginado (FR-019): **todo** (ocultas, absorbidas y marcas incluidas) y con
-- el **mismo** orden que ve el visitante. Filtros opcionales:
--   * `visibility` (`NULL` = sin filtro);
--   * `include_merged` (`NULL`/`false` = sin absorbidas — por defecto; `true` solo
--     para la revisión avanzada del panel, semántica de T925);
--   * `review_mark` = marca **PENDIENTE** (sin resolver) de esa clase:
--     `possible_duplicate` → par abierto (`duplicate_of_id` con resolución NULL);
--     `unavailable`/`missing_from_feed` → `availability_mark` de esa clase con
--     `availability_mark_resolved_at` NULL (corrección de la fase 5: los dos ejes
--     son independientes y «pendiente» se pregunta por eje).
SELECT m.id,
       m.title,
       m.published_at,
       m.description_es,
       m.description_en,
       m.description_edited,
       m.cover_file,
       m.origin,
       m.visibility,
       m.duplicate_of_id,
       m.duplicate_mark_resolved_at,
       m.availability_checked_at,
       m.availability_result,
       m.availability_mark,
       m.availability_mark_resolved_at,
       m.merged_into_id,
       vy.url AS video_url,
       vp.url AS audio_url,
       m.created_at,
       m.updated_at
FROM media_items m
LEFT JOIN media_episode_sources vy
       ON vy.item_id = m.id AND vy.provider = 'youtube'
LEFT JOIN media_episode_sources vp
       ON vp.item_id = m.id AND vp.provider = 'podcast'
WHERE (sqlc.narg('visibility')::text IS NULL
       OR m.visibility = sqlc.narg('visibility')::text)
  AND (sqlc.narg('include_merged')::boolean IS TRUE
       OR m.merged_into_id IS NULL)
  AND (sqlc.narg('review_mark')::text IS NULL
       OR (sqlc.narg('review_mark')::text = 'possible_duplicate'
           AND m.duplicate_of_id IS NOT NULL
           AND m.duplicate_mark_resolved_at IS NULL)
       OR (sqlc.narg('review_mark')::text IN ('unavailable', 'missing_from_feed')
           AND m.availability_mark = sqlc.narg('review_mark')::text
           AND m.availability_mark_resolved_at IS NULL))
ORDER BY m.published_at DESC, m.created_at ASC, m.id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountMediaItems :one
-- Mismo filtro que `ListMediaItems` para que `total` corresponda a la página.
SELECT count(*)::bigint AS total
FROM media_items m
WHERE (sqlc.narg('visibility')::text IS NULL
       OR m.visibility = sqlc.narg('visibility')::text)
  AND (sqlc.narg('include_merged')::boolean IS TRUE
       OR m.merged_into_id IS NULL)
  AND (sqlc.narg('review_mark')::text IS NULL
       OR (sqlc.narg('review_mark')::text = 'possible_duplicate'
           AND m.duplicate_of_id IS NOT NULL
           AND m.duplicate_mark_resolved_at IS NULL)
       OR (sqlc.narg('review_mark')::text IN ('unavailable', 'missing_from_feed')
           AND m.availability_mark = sqlc.narg('review_mark')::text
           AND m.availability_mark_resolved_at IS NULL));

-- name: GetMediaItemByID :one
-- Panel (detalle y respuestas de escritura): la fila completa —incluidas
-- visibilidad y marcas de los dos ejes— más los dos enlaces de sus fuentes.
SELECT m.id,
       m.title,
       m.published_at,
       m.description_es,
       m.description_en,
       m.description_edited,
       m.cover_file,
       m.origin,
       m.visibility,
       m.duplicate_of_id,
       m.duplicate_mark_resolved_at,
       m.availability_checked_at,
       m.availability_result,
       m.availability_mark,
       m.availability_mark_resolved_at,
       m.merged_into_id,
       vy.url AS video_url,
       vp.url AS audio_url,
       m.created_at,
       m.updated_at
FROM media_items m
LEFT JOIN media_episode_sources vy
       ON vy.item_id = m.id AND vy.provider = 'youtube'
LEFT JOIN media_episode_sources vp
       ON vp.item_id = m.id AND vp.provider = 'podcast'
WHERE m.id = sqlc.arg('id');

-- name: InsertMediaItem :one
-- Alta al importar (FR-011/FR-012): solo los datos que trae el feed —título,
-- fecha, descripción corta ya recortada y carátula ya copiada— más `origin` y
-- `visibility`, que calcula el service en la misma transacción que las fuentes
-- (FR-014: la entrada adicional de un par dudoso nace `hidden`). Los ejes de
-- marcas y `merged_into_id` nacen en NULL y `description_edited` en `false`
-- (DEFAULT de la tabla: la descripción es del feed hasta que el equipo la ajuste,
-- FR-015).
INSERT INTO media_items (
    title, published_at, description_es, description_en, cover_file, origin, visibility
) VALUES (
    sqlc.arg('title'),
    sqlc.arg('published_at'),
    sqlc.narg('description_es'),
    sqlc.narg('description_en'),
    sqlc.narg('cover_file'),
    sqlc.arg('origin'),
    sqlc.arg('visibility')
)
RETURNING id, title, published_at, description_es, description_en, description_edited,
          cover_file, origin, visibility, duplicate_of_id, duplicate_mark_resolved_at,
          availability_checked_at, availability_result, availability_mark,
          availability_mark_resolved_at, merged_into_id, created_at, updated_at;

-- name: UpdateMediaItem :one
-- El service **fusiona** en memoria (descripción, visibilidad, marcas por eje,
-- disponibilidad, unión) y esta consulta escribe **la fila completa ya fusionada**
-- (patrón `UpdateDonationMethod` de F7): lo que no llega en el `narg` queda NULL,
-- que es exactamente lo que el modelo quiere —una marca resuelta se guarda en su
-- columna `*_resolved_at`, y despejar el par de duplicados escribe `NULL`—.
-- `title`, `published_at` y `cover_file` no son editables (FR-019): se reescriben
-- tal cual para que la instantánea de la importación no cambie (R9-6/P9-9).
UPDATE media_items
SET title                      = sqlc.arg('title'),
    published_at               = sqlc.arg('published_at'),
    description_es             = sqlc.narg('description_es'),
    description_en             = sqlc.narg('description_en'),
    description_edited         = sqlc.arg('description_edited'),
    cover_file                 = sqlc.narg('cover_file'),
    origin                     = sqlc.arg('origin'),
    visibility                 = sqlc.arg('visibility'),
    duplicate_of_id            = sqlc.narg('duplicate_of_id'),
    duplicate_mark_resolved_at = sqlc.narg('duplicate_mark_resolved_at'),
    availability_checked_at    = sqlc.narg('availability_checked_at'),
    availability_result        = sqlc.narg('availability_result'),
    availability_mark          = sqlc.narg('availability_mark'),
    availability_mark_resolved_at = sqlc.narg('availability_mark_resolved_at'),
    merged_into_id             = sqlc.narg('merged_into_id'),
    updated_at                 = now()
WHERE id = sqlc.arg('id')
RETURNING id, title, published_at, description_es, description_en, description_edited,
          cover_file, origin, visibility, duplicate_of_id, duplicate_mark_resolved_at,
          availability_checked_at, availability_result, availability_mark,
          availability_mark_resolved_at, merged_into_id, created_at, updated_at;

-- name: RepointMediaItemSources :many
-- Unión por puntero (R9-12/P9-10): mueve las fuentes de la ficha absorbida a la
-- resultante (la procedencia no se pierde y **nunca** se borra una fuente: FR-019
-- «no hay borrado»). Devuelve las filas movidas —el service comprueba así que la
-- resultante queda con los dos enlaces y recalcula `origin`—; con `:execrows` no
-- habría forma de ver el resultado sin otra consulta.
UPDATE media_episode_sources
SET item_id    = sqlc.arg('to_item_id'),
    updated_at = now()
WHERE item_id = sqlc.arg('from_item_id')
RETURNING id, item_id, provider, external_id, feed_title, url, created_at, updated_at;

-- name: ListMediaItemSourcesByItem :many
-- Fuentes de una ficha: procedencia en el panel (`FuentesList`) y comprobación del
-- resultado de la unión. Orden de importación (FIFO, el mismo desempate del
-- listado) para que la procedencia se lea igual en cada carga.
SELECT id,
       item_id,
       provider,
       external_id,
       feed_title,
       url,
       created_at,
       updated_at
FROM media_episode_sources
WHERE item_id = sqlc.arg('item_id')
ORDER BY created_at, id;
