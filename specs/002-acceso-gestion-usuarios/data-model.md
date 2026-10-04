# Data Model — F2 Acceso y gestión de usuarios

**Fecha**: 2026-10-04 · **Rama**: `002-acceso-gestion-usuarios` · **Spec**: `spec.md` (aprobada)
· **Convenciones**: `specs/001-estructura-base/data-model.md` (migraciones), skill `postgres-db`,
`docs/tecnico/arquitectura.md` §8.1 (PK UUID, mapeo `pgtype`, orden de listados).

## Decisión: cómo se aterriza la referencia de F1 §"F2"

F1 dejó una asignación **orientativa**: `users`, `sessions`, `roles`, `user_roles`, permisos por
módulo (p. ej. `role_permissions`). Este plan la **simplifica y completa** así:

| Tabla de la referencia de F1 | Decisión F2 | Por qué |
|---|---|---|
| `users` | **Se mantiene** | Cuenta del panel (FR-009…FR-013) |
| `sessions` | **Se mantiene** | Sesión en servidor (D-A7, R1 de `research.md`) |
| `roles` | **Se mantiene** | Roles creados por el administrador (FR-014/FR-017) |
| `user_roles` | **Se elimina** | La decisión Q4 fija **un solo rol por cuenta**: la relación es `users.role_id` (FK). Una tabla de unión permitiría varios roles, exactamente lo que la spec prohíbe |
| permisos por módulo (ej. `role_permissions`) | **Se mantiene `role_permissions`** y se **añade `permissions`** como catálogo fijo sembrado por la migración | FK y unicidad reales + `GET /api/v1/admin/permisos` puede listar el catálogo con etiquetas sin duplicarlo en código |
| (no previsto) | **Se añade `login_attempts`** | FR-006 (5 intentos / 15 min) sin enumerar cuentas (R5 de `research.md`) |

Reglas que se respetan de las convenciones: nombres en inglés y plural; `id UUID PRIMARY KEY
DEFAULT gen_random_uuid()`; `created_at`/`updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`; `NOT NULL`
por defecto; `TEXT` con `CHECK` de longitud; claves foráneas explícitas con `ON DELETE` decidido;
índices en FK y en las columnas de `WHERE`/`ORDER BY`; `UNIQUE`/`CHECK` **en la base**, no solo en el
código; `up`/`down` completos; **nunca** se edita una migración aplicada (F2 arranca en `000002`).

## Resumen de tablas

```text
permissions ──< role_permissions >── roles ──< users >── sessions
                                          └────< login_attempts (por identificador, sin FK)
```

| Tabla | Propietario (paquete) | Reglas que cubre |
|---|---|---|
| `permissions` | `internal/usuarios` | FR-015 (catálogo = módulos del producto) |
| `roles` | `internal/usuarios` | FR-014, FR-017 |
| `role_permissions` | `internal/usuarios` | FR-014 (≥1 permiso), FR-018 |
| `users` | `internal/usuarios` | FR-009…FR-013, FR-017 (un rol), FR-019 |
| `sessions` | `internal/usuarios` (plumbing de token/cookie en `platform/session`) | FR-001, FR-004, FR-005, FR-012 |
| `login_attempts` | `internal/usuarios` | FR-006 |

## Migraciones

### `000002_create_roles_and_permissions.up.sql`

```sql
CREATE TABLE permissions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code       TEXT NOT NULL UNIQUE CHECK (char_length(code) BETWEEN 1 AND 64),
    label      TEXT NOT NULL CHECK (char_length(label) BETWEEN 1 AND 120),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Catálogo fijo de FR-015: módulos del producto + administración de usuarios y roles.
-- Los de F3–F9 quedan reservados: existen pero no dan acceso a nada hasta que se construyan.
INSERT INTO permissions (code, label) VALUES
    ('portada',             'Portada e información general'),
    ('eventos',             'Eventos'),
    ('actividades',         'Actividades'),
    ('grupos',              'Grupos de conexión'),
    ('ministerios',         'Ministerios'),
    ('donaciones',          'Donaciones'),
    ('noticias',            'Noticias y galería'),
    ('medios',              'Medios (prédicas y podcasts)'),
    ('admin_usuarios_roles','Administración de usuarios y roles');

CREATE TABLE roles (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Unicidad normalizada (Q5): solo difiere en mayúsculas = el mismo rol.
CREATE UNIQUE INDEX roles_name_lower_idx ON roles (lower(name));

CREATE TABLE role_permissions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE RESTRICT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (role_id, permission_id)
);

CREATE INDEX role_permissions_permission_id_idx ON role_permissions (permission_id);
```

Notas:

- "Un rol tiene al menos un permiso" (FR-014) **no** se puede expresar como restricción simple de
  una tabla de unión (exigiría un trigger o `DEFERRABLE`); se garantiza en el **service** (toda
  creación/edición de roles es una transacción con `DELETE`+`INSERT` de `role_permissions`) y se
  verifica con pruebas. Es el único invariante de F2 fuera de la base; queda registrado aquí a
  propósito (skill `postgres-db`: `CHECK` en la base salvo lo no expresable).
- `roles.name` conserva las mayúsculas de presentación (US6 esc. 3: "aparece… con el nuevo nombre");
  la unicidad va en `lower(name)`.
- `role_permissions.role_id` con `ON DELETE CASCADE`: al eliminar un rol (solo si no está en uso,
  FR-017) se van sus filas de permisos. `permission_id` con `ON DELETE RESTRICT`: el catálogo de
  FR-015 no se borra nunca.

### `000002_create_roles_and_permissions.down.sql`

```sql
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS permissions;
```

### `000003_create_users.up.sql`

```sql
CREATE TABLE users (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                TEXT NOT NULL UNIQUE CHECK (char_length(email) BETWEEN 3 AND 254),
    full_name            TEXT NOT NULL CHECK (char_length(full_name) BETWEEN 1 AND 120),
    password_hash        TEXT NOT NULL,
    must_change_password BOOLEAN NOT NULL DEFAULT true,
    is_active            BOOLEAN NOT NULL DEFAULT true,
    role_id              UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- El correo se guarda normalizado (Q5): trim + minúsculas.
    CHECK (email = lower(btrim(email)))
);

CREATE INDEX users_role_id_idx ON users (role_id);
CREATE INDEX users_created_at_id_idx ON users (created_at DESC, id DESC);
-- Recuento anti-bloqueo (FR-008): cuentas activas con un permiso dado.
CREATE INDEX users_is_active_idx ON users (is_active) WHERE is_active;
```

Notas:

- **No hay borrado de cuentas** (FR-013): no existe `DELETE` sobre `users` en ningún código ni en el
  contrato; retirar acceso = `is_active = false` (FR-012) y los datos se conservan.
- `role_id NOT NULL` con `ON DELETE RESTRICT`: un rol en uso no se puede eliminar (FR-017) — la
  restricción es la red de seguridad del service, y el mensaje explicativo lo pone el service
  (`409 conflict`).
- `full_name` es el campo de "identificación" de la spec (ver riesgo R3 del plan: pendiente de
  confirmar si además hace falta un documento de identidad).
- `password_hash` guarda el hash **bcrypt** (nunca la contraseña, §IV); ningún DTO de salida incluye
  este campo (FR-003).
- `must_change_password = true` cuando un administrador define/restablece la contraseña (US3 esc. 6,
  US7 esc. 4–5). `is_active = true` al crear (US3 esc. 1).
- `users_created_at_id_idx` refleja el orden por defecto del listado (§8.1.7: `created_at DESC,
  id DESC`). El parágrafo `WHERE is_active` del índice parcial apoya el recuento del guard
  anti-bloqueo, que solo mira cuentas activas.

### `000003_create_users.down.sql`

```sql
DROP TABLE IF EXISTS users;
```

### `000004_create_sessions.up.sql`

```sql
CREATE TABLE sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash   BYTEA NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);
```

Notas:

- `token_hash` es el **SHA-256** del token de la cookie (32 bytes); el token en claro **no** se
  guarda (una lectura indebida de la BD no permite secuestrar sesiones). `UNIQUE` hace la
  resolución por índice.
- `last_seen_at` mide la **inactividad** (30 min; FR-005) y `expires_at` la **vida absoluta**
  propuesta de 12 h (R15 de `research.md`, pendiente de confirmación junto con D-A7; si no se
  aprueba, `expires_at` se fija igual y la inactividad basta para invalidar).
- `ON DELETE CASCADE` sobre `users`: una sesión sin cuenta no significa nada. Los usuarios **no se
  borran** en el MVP (FR-013), así que el `CASCADE` es la red de seguridad para un borrado futuro.
- Limpieza: se borran las sesiones expiradas al crear una nueva (oportunista) y todas las de una
  cuenta al desactivarla o al restablecer su contraseña (FR-012, R17). No hay cron: no hay
  infraestructura de trabajos en F2 y el volumen es mínimo.

### `000004_create_sessions.down.sql`

```sql
DROP TABLE IF EXISTS sessions;
```

### `000005_create_login_attempts.up.sql`

```sql
CREATE TABLE login_attempts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier    TEXT NOT NULL UNIQUE CHECK (char_length(identifier) BETWEEN 3 AND 254),
    failed_count  INTEGER NOT NULL DEFAULT 0 CHECK (failed_count >= 0),
    last_failed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    blocked_until TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (identifier = lower(btrim(identifier)))
);
```

Notas:

- Una fila **por identificador normalizado (correo), exista o no la cuenta** (R5 de
  `research.md`): el bloqueo de FR-006 se comporta igual para cuentas reales e inexistentes y por
  tanto no revela existencia (FR-003/SC-008).
- Sin FK a `users` a propósito: también debe registrar intentos sobre correos inexistentes.
- Semántica (en el service): fallo → `failed_count + 1`, `last_failed_at = now()`; con
  `failed_count >= 5` → `blocked_until = now() + 15 min`; acierto → `DELETE` de la fila; bloqueo
  vencido → la fila se resetea en el siguiente intento. Los valores **5** y **15 min** son
  constantes de código (confirmados por el humano el 2026-10-04).
- Las filas se eliminan al resolverse (acierto o vencimiento): la tabla no crece sin control.

### `000005_create_login_attempts.down.sql`

```sql
DROP TABLE IF EXISTS login_attempts;
```

## Consultas sqlc (`backend/internal/db/queries/`)

Todas parametrizadas, sin `SELECT *`, con `ORDER BY` determinista y `LIMIT`/`OFFSET` en listados
(§8.1.3 y §8.1.7). Inventario previsto (nombres de sqlc):

| Archivo | Consultas |
|---|---|
| `users.sql` | `InsertUser`, `GetUserByID`, `GetUserByEmail`, `GetUserAuthByEmail` (correo + hash + estado + rol + permisos, para login/`authn`), `ListUsers` (con `total` aparte), `CountUsers`, `UpdateUser` (nombre, correo, rol, activo), `UpdateUserPassword`, `SetUserMustChangePassword`, `CountActiveAdmins` (recuento **post-mutación** para FR-008), `CountUsersByRole` |
| `roles.sql` | `InsertRole`, `GetRoleByID`, `GetRoleByNameLower`, `ListRoles` (+ `CountRoles`), `UpdateRoleName`, `DeleteRolePermissions`, `InsertRolePermission`, `DeleteRole`, `CountRoleUsers` |
| `permissions.sql` | `ListPermissions` (catálogo), `GetPermissionIDsByCodes` (para crear/editar roles) |
| `sessions.sql` | `InsertSession`, `GetSessionByTokenHash` (+ expiración), `TouchSession` (actualiza `last_seen_at`, estrangulado a 1/min en el service), `DeleteSession`, `DeleteSessionsByUser`, `DeleteExpiredSessions` |
| `login_attempts.sql` | `UpsertLoginAttempt` (incrementa o crea), `GetLoginAttempt`, `ResetLoginAttempt`, `DeleteLoginAttempt` |

Los tipos que emite sqlc (`pgtype.*`) se traducen **solo** en `repository.go`/`mapRow` (§8.1.6);
`sqlc.yaml` no lleva `overrides`. El guard anti-bloqueo usa `database.WithTx` (transacción) y una
consulta de recuento; la serialización entre transacciones concurrentes se hace con
`pg_advisory_xact_lock` llamado desde una consulta nombrada (`LockAdminGuard`), lo que queda
dentro de lo que D-A3 permite (SQL normal dentro del repository; **no** es un stored procedure).

## Invariantes de negocio y dónde se verifican

| Invariante | Dónde | Red de seguridad en la BD |
|---|---|---|
| Una cuenta tiene un solo rol (Q4) | Modelo (`users.role_id`) | FK `NOT NULL` |
| Un rol tiene al menos un permiso (FR-014) | Service (transacción al crear/editar) | `UNIQUE (role_id, permission_id)` evita duplicados; el mínimo **no** es expresable sin trigger |
| Un rol solo se elimina sin cuentas asignadas (FR-017) | Service (`CountRoleUsers`) | `ON DELETE RESTRICT` de `users.role_id` |
| Nunca se eliminan cuentas (FR-013) | No existe operación | — (no hay `DELETE` en las consultas de `users`) |
| Correo único normalizado (Q5) | Service (mensaje claro) | `UNIQUE` + `CHECK (email = lower(btrim(email)))` |
| Nombre de rol único normalizado (Q5) | Service (mensaje claro) | `UNIQUE (lower(name))` |
| Siempre ≥1 cuenta activa con `admin_usuarios_roles` (FR-008) | Service, transacción + advisory lock | Índice parcial `users_is_active_idx` |
| Inicialización única (FR-007) | Service, transacción + advisory lock + `users` vacío | — (la transacción serializada lo garantiza) |
| Bloqueo 5 intentos / 15 min (FR-006) | Service (`login_attempts`) | `UNIQUE (identifier)` |
| Sesión válida solo con cuenta activa (FR-012) | `authn` en **cada petición** + borrado de sesiones al desactivar | FK + limpieza |

## Validación del modelo (qué se comprobará en implementación)

1. `make db-migrate` aplica `000002`…`000005` y `migrate down` las revierte por completo, en orden
   inverso.
2. Pruebas de integración (`//go:build integration`) contra PostgreSQL real:
   duplicados normalizados rechazados (`UNIQUE`), `ON DELETE RESTRICT`/`CASCADE`, `CHECK` de
   normalización del correo, upsert de intentos, expiración de sesiones, **carrera anti-bloqueo**
   (dos transacciones concurrentes no dejan 0 administradores) e **inicialización única** (dos
   peticiones simultáneas → una sola crea).
3. `make sqlc-verify` sin diferencias (el código generado acompaña al SQL).
4. `GET /healthz` sigue respondiendo igual tras las migraciones (el esquema nuevo no afecta a F1).
