-- Consultas de roles y su relación con el catálogo de permisos (FR-014, FR-017).
-- Convenciones: skill `postgres-db` y specs/002-acceso-gestion-usuarios/data-model.md.
-- Columnas listadas explícitamente (sin comodines), parámetros $1…$n y ORDER BY determinista.
-- `ListRoles`/`GetRoleByID` embeben los códigos de permiso y el recuento de cuentas
-- que usan el rol (FR-017) para servir `RoleItem` sin consultas N+1.

-- name: InsertRole :one
INSERT INTO roles (name)
VALUES (sqlc.arg('name'))
RETURNING id, name, created_at, updated_at;

-- name: GetRoleByID :one
SELECT r.id,
       r.name,
       r.created_at,
       r.updated_at,
       COALESCE((
           SELECT array_agg(p.code ORDER BY p.code)
           FROM role_permissions rp
           JOIN permissions p ON p.id = rp.permission_id
           WHERE rp.role_id = r.id
       ), '{}')::text[] AS permissions,
       (SELECT count(*) FROM users u WHERE u.role_id = r.id)::bigint AS user_count
FROM roles r
WHERE r.id = sqlc.arg('id');

-- name: GetRoleByNameLower :one
-- Unicidad normalizada (Q5): dos nombres que solo difieren en mayúsculas (o en
-- espacios sobrantes en los extremos) son el mismo rol. El service normaliza el
-- nombre antes de llamar; `btrim` aquí es la red de seguridad.
SELECT id, name, created_at, updated_at
FROM roles
WHERE lower(name) = lower(btrim(sqlc.arg('name')))
LIMIT 1;

-- name: ListRoles :many
SELECT r.id,
       r.name,
       r.created_at,
       r.updated_at,
       COALESCE((
           SELECT array_agg(p.code ORDER BY p.code)
           FROM role_permissions rp
           JOIN permissions p ON p.id = rp.permission_id
           WHERE rp.role_id = r.id
       ), '{}')::text[] AS permissions,
       (SELECT count(*) FROM users u WHERE u.role_id = r.id)::bigint AS user_count
FROM roles r
ORDER BY r.created_at DESC, r.id DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountRoles :one
SELECT count(*)::bigint AS total
FROM roles;

-- name: UpdateRoleName :one
UPDATE roles
SET name = sqlc.arg('name'),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, name, created_at, updated_at;

-- name: DeleteRolePermissions :exec
-- Borra los permisos actuales del rol; el service reinserta el conjunto nuevo en
-- la misma transacción (un rol conserva ≥ 1 permiso, FR-014).
DELETE FROM role_permissions
WHERE role_id = sqlc.arg('role_id');

-- name: InsertRolePermission :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES (sqlc.arg('role_id'), sqlc.arg('permission_id'));

-- name: DeleteRole :execrows
-- Solo llega aquí un rol sin cuentas asignadas (FR-017); el `ON DELETE RESTRICT`
-- de `users.role_id` es la red de seguridad. `role_permissions` cae en cascada.
DELETE FROM roles
WHERE id = sqlc.arg('id');

-- name: CountRoleUsers :one
-- Cuentas asignadas al rol: `0` = el rol se puede eliminar (FR-017).
SELECT count(*)::bigint AS total
FROM users
WHERE role_id = sqlc.arg('role_id');

-- name: LockAdminGuard :exec
-- Cerradura de asesoramiento transaccional (P7). Clave FIJA y compartida por el
-- guard anti-bloqueo (FR-008) y la inicialización única (FR-007): serializa las
-- mutaciones que podrían dejar el panel sin administración. Se libera al cerrar
-- la transacción. No cambiar el valor sin coordinarlo con esas rutas.
SELECT pg_advisory_xact_lock(726112001);
