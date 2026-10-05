-- Consultas del catálogo de permisos (FR-015).
-- Convenciones: skill `postgres-db` y specs/002-acceso-gestion-usuarios/data-model.md.
-- Columnas listadas explícitamente (sin comodines), parámetros $1…$n y ORDER BY determinista.

-- name: ListPermissions :many
SELECT id, code, label, created_at, updated_at
FROM permissions
ORDER BY code ASC;

-- name: GetPermissionIDsByCodes :many
-- Devuelve el par (id, code) de cada código presente en el catálogo. Devolver
-- también `code` permite al service distinguir qué códigos no existen (400 con
-- `details`) al crear/editar roles; el `id` es lo que se inserta en
-- `role_permissions`. Códigos repetidos no duplican filas.
SELECT id, code
FROM permissions
WHERE code = ANY(sqlc.arg('codes')::text[])
ORDER BY code ASC;
