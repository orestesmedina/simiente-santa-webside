-- Consultas del dominio `medios` (F9): el registro de cada «Buscar episodios
-- nuevos» (`media_sync_runs`, FR-018/R9-9). Convenciones: skill `postgres-db` y
-- specs/009-medios-predicas-y-podcasts/data-model.md (fuente de verdad).
--
-- Una fila por búsqueda **completada** (con conteos parciales si el presupuesto de
-- 90 s se agotó o un feed falló: los fallos se informan, no se esconden). El
-- resumen viaja también en la auditoría (`media.sync`, una fila de
-- `admin_actions` por búsqueda, P9-14) y este registro guarda el detalle numérico.
--
-- Columnas listadas explícitamente (sin `SELECT *`), `ORDER BY` determinista
-- (§8.1.7) y todo parametrizado (§IV).

-- name: InsertMediaSyncRun :one
-- Fin de cada búsqueda: quién la pidió (`actor_user_id`, NULL solo si la
-- ejecución no tiene actor), cuándo empezó y terminó, el resumen de FR-018
-- (`imported_count`, `joined_count`, `marked_count` —**solo** marcas «posible
-- duplicado» nuevas, criterio B2 de la fase 5—, `hidden_count`) y los contadores
-- de disponibilidad de FR-017 (`availability_checked_count`,
-- `availability_inconclusive_count`), más el estado de cada feed (`failed` nunca
-- oculta ni marca nada por esa causa, FR-016).
INSERT INTO media_sync_runs (
    actor_user_id, started_at, finished_at,
    imported_count, joined_count, marked_count, hidden_count,
    availability_checked_count, availability_inconclusive_count,
    youtube_status, podcast_status
) VALUES (
    sqlc.narg('actor_user_id'),
    sqlc.arg('started_at'),
    sqlc.arg('finished_at'),
    sqlc.arg('imported_count'),
    sqlc.arg('joined_count'),
    sqlc.arg('marked_count'),
    sqlc.arg('hidden_count'),
    sqlc.arg('availability_checked_count'),
    sqlc.arg('availability_inconclusive_count'),
    sqlc.arg('youtube_status'),
    sqlc.arg('podcast_status')
)
RETURNING id, actor_user_id, started_at, finished_at,
          imported_count, joined_count, marked_count, hidden_count,
          availability_checked_count, availability_inconclusive_count,
          youtube_status, podcast_status, created_at, updated_at;

-- name: GetLastMediaSyncRun :one
-- «Última búsqueda» del listado del panel (FR-018/P9-14): la más reciente por
-- `started_at` con desempate por `id` (determinista, §8.1.7). Sin ninguna
-- búsqueda todavía, `sqlc` devuelve `pgx.ErrNoRows` y el panel muestra el bloque
-- como «nunca se ha buscado» (no es un error).
SELECT id,
       actor_user_id,
       started_at,
       finished_at,
       imported_count,
       joined_count,
       marked_count,
       hidden_count,
       availability_checked_count,
       availability_inconclusive_count,
       youtube_status,
       podcast_status,
       created_at,
       updated_at
FROM media_sync_runs
ORDER BY started_at DESC, id
LIMIT 1;
