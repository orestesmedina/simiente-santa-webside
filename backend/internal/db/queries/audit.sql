-- Consultas de auditoría (FR-022…FR-025): historial de intentos de acceso y de
-- acciones administrativas. Convenciones: skill `postgres-db` y
-- specs/002-acceso-gestion-usuarios/data-model.md.
--
-- SOLO inserción y consulta: el registro es de solo inserción (FR-025). No hay
-- sentencias de modificación ni de borrado sobre `login_events`/`admin_actions`
-- en ningún PR.
--
-- Los campos derivados del contrato se resuelven con `LEFT JOIN users` (F-01):
-- no se guardan duplicados en el registro. El `LEFT JOIN` trae `user_email` /
-- `actor_email` y las partes del nombre (`*_first_name` / `*_last_name`), todas
-- anulables; el `repository.go` compone `userName`/`actorName` como
-- `firstName lastName` (mapRow, §8.1.6). Se devuelven las partes —y no una
-- columna concatenada— para que sqlc infiera `pgtype.Text` anulable: sqlc
-- tipa mal las expresiones `||`/`CASE`/`NULLIF` sobre `LEFT JOIN` (`interface{}`,
-- `string` no anulable o `bool`).
-- Cuando el intento no tiene cuenta asociada, el `LEFT JOIN` deja esas columnas
-- en NULL (nunca se guarda ni se muestra el correo probado, FR-026).
--
-- Filtro por rango de fechas: semirango `[from, to)` — `created_at >= from` y
-- `created_at < to` (FR-024). `from`/`to` son opcionales (`narg` NULL = sin cota).

-- name: InsertLoginEvent :one
-- Una fila por intento (éxito o fallo). `user_id` NULL cuando el correo no
-- corresponde a ninguna cuenta: se registra el intento sin asociarlo ni crear
-- nada (FR-003/FR-022).
INSERT INTO login_events (user_id, result, ip)
VALUES (sqlc.narg('user_id'), sqlc.arg('result'), sqlc.arg('ip'))
RETURNING id, user_id, result, ip, created_at;

-- name: ListLoginEvents :many
SELECT le.id,
       le.user_id,
       u.email AS user_email,
       u.first_name AS user_first_name,
       u.last_name AS user_last_name,
       le.result,
       le.ip,
       le.created_at
FROM login_events le
LEFT JOIN users u ON u.id = le.user_id
WHERE (sqlc.narg('user_id')::uuid IS NULL OR le.user_id = sqlc.narg('user_id'))
  AND (sqlc.narg('from')::timestamptz IS NULL OR le.created_at >= sqlc.narg('from'))
  AND (sqlc.narg('to')::timestamptz IS NULL OR le.created_at < sqlc.narg('to'))
ORDER BY le.created_at DESC, le.id DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountLoginEvents :one
-- Mismo filtro que `ListLoginEvents` para que `total` corresponda a la página.
SELECT count(*)::bigint AS total
FROM login_events le
WHERE (sqlc.narg('user_id')::uuid IS NULL OR le.user_id = sqlc.narg('user_id'))
  AND (sqlc.narg('from')::timestamptz IS NULL OR le.created_at >= sqlc.narg('from'))
  AND (sqlc.narg('to')::timestamptz IS NULL OR le.created_at < sqlc.narg('to'));

-- name: InsertAdminAction :one
-- Una fila por acción administrativa sensible, incluidos los intentos que fallan
-- o se deniegan (FR-023). La inicialización (FR-007) deja `actor_user_id` NULL
-- con `action = 'user.create'`, el único caso sin actor (CHECK de la tabla).
INSERT INTO admin_actions (
    actor_user_id,
    action,
    target_kind,
    target_user_id,
    target_role_id,
    target_label,
    result
)
VALUES (
    sqlc.narg('actor_user_id'),
    sqlc.arg('action'),
    sqlc.arg('target_kind'),
    sqlc.narg('target_user_id'),
    sqlc.narg('target_role_id'),
    sqlc.narg('target_label'),
    sqlc.arg('result')
)
RETURNING id, actor_user_id, action, target_kind, target_user_id, target_role_id,
          target_label, result, created_at;

-- name: ListAdminActions :many
-- Filtro `userId`: cuenta **involucrada** — la que hizo la acción (`actor_user_id`)
-- o la cuenta objetivo (`target_user_id`). La inicialización del sistema encaja
-- por su `target_user_id`.
SELECT aa.id,
       aa.actor_user_id,
       u.email AS actor_email,
       u.first_name AS actor_first_name,
       u.last_name AS actor_last_name,
       aa.action,
       aa.target_kind,
       COALESCE(aa.target_user_id, aa.target_role_id) AS target_id,
       aa.target_label,
       aa.result,
       aa.created_at
FROM admin_actions aa
LEFT JOIN users u ON u.id = aa.actor_user_id
WHERE (
        sqlc.narg('user_id')::uuid IS NULL
        OR aa.actor_user_id = sqlc.narg('user_id')
        OR aa.target_user_id = sqlc.narg('user_id')
      )
  AND (sqlc.narg('from')::timestamptz IS NULL OR aa.created_at >= sqlc.narg('from'))
  AND (sqlc.narg('to')::timestamptz IS NULL OR aa.created_at < sqlc.narg('to'))
ORDER BY aa.created_at DESC, aa.id DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountAdminActions :one
-- Mismo filtro que `ListAdminActions` para que `total` corresponda a la página.
SELECT count(*)::bigint AS total
FROM admin_actions aa
WHERE (
        sqlc.narg('user_id')::uuid IS NULL
        OR aa.actor_user_id = sqlc.narg('user_id')
        OR aa.target_user_id = sqlc.narg('user_id')
      )
  AND (sqlc.narg('from')::timestamptz IS NULL OR aa.created_at >= sqlc.narg('from'))
  AND (sqlc.narg('to')::timestamptz IS NULL OR aa.created_at < sqlc.narg('to'));
