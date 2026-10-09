-- Consultas de cuentas del panel (FR-009…FR-013, FR-017, FR-019, FR-021).
-- Convenciones: skill `postgres-db` y specs/002-acceso-gestion-usuarios/data-model.md.
-- Columnas listadas explícitamente (sin comodines), parámetros $1…$n, ORDER BY determinista
-- (`created_at DESC, id DESC`, §8.1.7) y LIMIT/OFFSET en los listados.
-- NO existe `DELETE` sobre `users`: retirar acceso es `is_active = false` (FR-013).

-- name: InsertUser :one
INSERT INTO users (
    email,
    first_name,
    last_name,
    phone,
    password_hash,
    must_change_password,
    is_active,
    role_id
)
VALUES (
    sqlc.arg('email'),
    sqlc.arg('first_name'),
    sqlc.arg('last_name'),
    sqlc.arg('phone'),
    sqlc.arg('password_hash'),
    sqlc.arg('must_change_password'),
    sqlc.arg('is_active'),
    sqlc.arg('role_id')
)
RETURNING id, email, first_name, last_name, phone, must_change_password, is_active,
          role_id, last_login_at, last_login_ip, created_at, updated_at;

-- name: GetUserByID :one
-- `JOIN roles` (no `LEFT`): `users.role_id` es NOT NULL con FK, siempre hay rol.
SELECT u.id,
       u.email,
       u.first_name,
       u.last_name,
       u.phone,
       u.must_change_password,
       u.is_active,
       u.role_id,
       r.name AS role_name,
       u.last_login_at,
       u.last_login_ip,
       u.created_at,
       u.updated_at
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.id = sqlc.arg('id');

-- name: GetUserByEmail :one
-- El correo se guarda normalizado (Q5) y el service normaliza la entrada; la
-- igualdad exacta usa el índice `UNIQUE` de `email`.
SELECT u.id,
       u.email,
       u.first_name,
       u.last_name,
       u.phone,
       u.must_change_password,
       u.is_active,
       u.role_id,
       r.name AS role_name,
       u.last_login_at,
       u.last_login_ip,
       u.created_at,
       u.updated_at
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.email = sqlc.arg('email');

-- name: GetUserAuthByEmail :one
-- Para login y `authn`: credenciales de la cuenta (hash y estado) + su rol y los
-- códigos de permiso efectivos (FR-018). El hash nunca sale del repository como
-- DTO (§IV, FR-003); `permissions` vacío es `{}`, nunca NULL.
SELECT u.id,
       u.email,
       u.first_name,
       u.last_name,
       u.password_hash,
       u.must_change_password,
       u.is_active,
       u.role_id,
       r.name AS role_name,
       COALESCE((
           SELECT array_agg(p.code ORDER BY p.code)
           FROM role_permissions rp
           JOIN permissions p ON p.id = rp.permission_id
           WHERE rp.role_id = u.role_id
       ), '{}')::text[] AS permissions
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.email = sqlc.arg('email');

-- name: ListUsers :many
SELECT u.id,
       u.email,
       u.first_name,
       u.last_name,
       u.phone,
       u.must_change_password,
       u.is_active,
       u.role_id,
       r.name AS role_name,
       u.last_login_at,
       u.last_login_ip,
       u.created_at,
       u.updated_at
FROM users u
JOIN roles r ON r.id = u.role_id
ORDER BY u.created_at DESC, u.id DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountUsers :one
SELECT count(*)::bigint AS total
FROM users;

-- name: UpdateUser :one
-- Edita datos, rol y estado (FR-011); `email` normalizado por el service (Q5).
-- `RETURNING` no puede traer `role_name` de `roles`: el service relee con
-- `GetUserByID` cuando necesita el `UserItem` completo.
UPDATE users
SET email = sqlc.arg('email'),
    first_name = sqlc.arg('first_name'),
    last_name = sqlc.arg('last_name'),
    phone = sqlc.arg('phone'),
    role_id = sqlc.arg('role_id'),
    is_active = sqlc.arg('is_active'),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, email, first_name, last_name, phone, must_change_password, is_active,
          role_id, last_login_at, last_login_ip, created_at, updated_at;

-- name: UpdateUserPassword :exec
-- Solo el hash (bcrypt, §IV). El flag de cambio obligatorio se maneja aparte con
-- `SetUserMustChangePassword`, para distinguir el restablecimiento por un
-- administrador del cambio propio (FR-020).
UPDATE users
SET password_hash = sqlc.arg('password_hash'),
    updated_at = now()
WHERE id = sqlc.arg('id');

-- name: SetUserMustChangePassword :exec
UPDATE users
SET must_change_password = sqlc.arg('must_change_password'),
    updated_at = now()
WHERE id = sqlc.arg('id');

-- name: UpdateUserLastLogin :exec
-- Proyección del último acceso exitoso (FR-021); SOLO la escribe el login exitoso,
-- junto con su fila de `login_events`. Ninguna vía de API la acepta como entrada.
UPDATE users
SET last_login_at = sqlc.arg('last_login_at'),
    last_login_ip = sqlc.arg('last_login_ip'),
    updated_at = now()
WHERE id = sqlc.arg('id');

-- name: CountActiveAdmins :one
-- Recuento post-mutación del guard anti-bloqueo (FR-008/P7): cuentas activas cuyo
-- rol concede `admin_usuarios_roles`. `admin_usuarios_roles` es un `code` estable
-- y sembrado por la migración 000002.
SELECT count(*)::bigint AS total
FROM users u
WHERE u.is_active
  AND EXISTS (
      SELECT 1
      FROM role_permissions rp
      JOIN permissions p ON p.id = rp.permission_id
      WHERE rp.role_id = u.role_id
        AND p.code = 'admin_usuarios_roles'
  );

-- name: CountUsersByRole :one
SELECT count(*)::bigint AS total
FROM users
WHERE role_id = sqlc.arg('role_id');
