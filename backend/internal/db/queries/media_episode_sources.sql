-- Consultas del dominio `medios` (F9): las fuentes de importación
-- (`media_episode_sources`), una fila por elemento de feed. Convenciones: skill
-- `postgres-db` y specs/009-medios-predicas-y-podcasts/data-model.md (fuente de
-- verdad).
--
-- `UNIQUE (provider, external_id)` (R9-11/P9-11) es la pieza que hace la
-- búsqueda **idempotente por construcción**: el `INSERT … ON CONFLICT DO NOTHING`
-- de `InsertMediaEpisodeSource` descarta lo ya visto sin leerlo antes, así que dos
-- búsquedas simultáneas no duplican nada. Además, la importación es una
-- **instantánea** (P9-9/R9-6): la búsqueda nunca refresca título, fecha, carátula
-- ni enlace de lo ya importado, de modo que no hay ningún `UPDATE` sobre esta
-- tabla salvo el repunte de la unión (`RepointMediaItemSources`, en
-- `media_episodes.sql`).
--
-- Columnas listadas explícitamente (sin `SELECT *`) y `ORDER BY` determinista
-- (§8.1.7). Todo parametrizado (§IV).

-- name: InsertMediaEpisodeSource :one
-- Importación idempotente (FR-012/R9-11): con `ON CONFLICT (provider,
-- external_id) DO NOTHING` la fila **no se duplica ni se actualiza** —la
-- instantánea se conserva— y, si el elemento ya estaba importado, `RETURNING` no
-- devuelve nada: el repository traduce `pgx.ErrNoRows` a «ya existía» y la
-- búsqueda lo cuenta como no importado (no es un error).
INSERT INTO media_episode_sources (
    item_id, provider, external_id, feed_title, url
) VALUES (
    sqlc.arg('item_id'),
    sqlc.arg('provider'),
    sqlc.arg('external_id'),
    sqlc.arg('feed_title'),
    sqlc.arg('url')
)
ON CONFLICT (provider, external_id) DO NOTHING
RETURNING id, item_id, provider, external_id, feed_title, url, created_at, updated_at;

-- name: ListMediaEpisodeSourcesByProvider :many
-- Qué identificadores hay importados de cada proveedor: es la lista contra la que
-- la búsqueda detecta las fuentes que **ya no están en la ventana del feed**
-- (FR-017: los feeds solo publican sus últimos elementos). Solo se comprueban las
-- de fichas `visibility = 'visible'` no absorbidas (tope 30, lo aplica el
-- service); esta consulta devuelve la procedencia completa para poder
-- comprobarlas y para repuntarlas en una unión. Orden de importación (FIFO).
SELECT id,
       item_id,
       provider,
       external_id,
       feed_title,
       url,
       created_at,
       updated_at
FROM media_episode_sources
WHERE provider = sqlc.arg('provider')
ORDER BY created_at, id;

-- name: GetMediaEpisodeSource :one
-- ¿Ya está importado este elemento? (salto de la importación, FR-012) y detalle
-- de la fuente para las comprobaciones de disponibilidad y para el repunte de la
-- unión. `sqlc` devuelve `pgx.ErrNoRows` cuando el elemento no existe.
SELECT id,
       item_id,
       provider,
       external_id,
       feed_title,
       url,
       created_at,
       updated_at
FROM media_episode_sources
WHERE provider = sqlc.arg('provider')
  AND external_id = sqlc.arg('external_id');
