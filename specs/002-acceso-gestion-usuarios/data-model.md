# Data Model — F2 Acceso y gestión de usuarios

**Fecha**: 2026-10-04 · **Rama**: `002-acceso-gestion-usuarios` · **Spec**: `spec.md` (aprobada)
· **Convenciones**: `specs/001-estructura-base/data-model.md` (migraciones), skill `postgres-db`,
`docs/tecnico/arquitectura.md` §8.1 (PK UUID, mapeo `pgtype`, orden de listados).

## Decisión: cómo se aterriza la referencia de F1 §"F2"

F1 dejó una asignación **orientativa**: `users`, `sessions`, `roles`, `user_roles`, permisos por
módulo (p. ej. `role_permissions`). Este plan la **simplifica y completa** así (actualizado el
2026-10-04 con las decisiones del humano: **la sesión vive en Redis**, no en PostgreSQL):

| Tabla de la referencia de F1 | Decisión F2 | Por qué |
|---|---|---|
| `users` | **Se mantiene** | Cuenta del panel (FR-009…FR-013) |
| `sessions` | **No es tabla: vive en Redis** | D-A7 **confirmada** el 2026-10-04 con sesión en Redis (decisión explícita del humano para adoptar/probar Redis); ver `research.md` R1 y "Almacenamiento en Redis" más abajo |
| `roles` | **Se mantiene** | Roles creados por el administrador (FR-014/FR-017) |
| `user_roles` | **Se elimina** | La decisión Q4 fija **un solo rol por cuenta**: la relación es `users.role_id` (FK). Una tabla de unión permitiría varios roles, exactamente lo que la spec prohíbe |
| permisos por módulo (ej. `role_permissions`) | **Se mantiene `role_permissions`** y se **añade `permissions`** como catálogo fijo sembrado por la migración | FK y unicidad reales + `GET /api/v1/admin/permisos` puede listar el catálogo con etiquetas sin duplicarlo en código |
| (no previsto) `login_attempts` | **No es tabla: contadores en Redis con TTL** | FR-006 (5 intentos / 15 min) sin enumerar cuentas (`research.md` R5). El estado es efímero por definición y Redis ya existe por la sesión: su TTL reemplaza a la limpieza manual |

Reglas que se respetan de las convenciones: nombres en inglés y plural; `id UUID PRIMARY KEY
DEFAULT gen_random_uuid()`; `created_at`/`updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`; `NOT NULL`
por defecto; `TEXT` con `CHECK` de longitud; claves foráneas explícitas con `ON DELETE` decidido;
índices en FK y en las columnas de `WHERE`/`ORDER BY`; `UNIQUE`/`CHECK` **en la base**, no solo en el
código; `up`/`down` completos; **nunca** se edita una migración aplicada. **F2 consta de dos
migraciones, `000002` y `000003`** (la numeración sigue desde el baseline `000001` de F1; al no
crear tablas de sesión ni de intentos no hay huecos que renumerar).

## Resumen de tablas

```text
PostgreSQL:  permissions ──< role_permissions >── roles ──< users

Redis:       sess:<sha256(token)>      → sesión (TTL 30 min, vida absoluta 1 h)
             user_sessions:<user_id>   → SET de sesiones abiertas (revocación por cuenta)
             login:fail:<identificador>  /  login:block:<identificador>  → intentos y bloqueo
```

| Tabla / almacén | Propietario (paquete) | Reglas que cubre |
|---|---|---|
| `permissions` | `internal/usuarios` | FR-015 (catálogo = módulos del producto) |
| `roles` | `internal/usuarios` | FR-014, FR-017 |
| `role_permissions` | `internal/usuarios` | FR-014 (≥1 permiso), FR-018 |
| `users` | `internal/usuarios` | FR-009…FR-013, FR-017 (un rol), FR-019 |
| Redis: `sess:*` / `user_sessions:*` | `internal/platform/session` (implementación Redis de `session.Store`) | FR-001, FR-004, FR-005, FR-012 |
| Redis: `login:fail:*` / `login:block:*` | `internal/usuarios` | FR-006 |

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
    first_name           TEXT NOT NULL CHECK (char_length(first_name) BETWEEN 1 AND 80),
    last_name            TEXT NOT NULL CHECK (char_length(last_name) BETWEEN 1 AND 120),
    phone                TEXT NOT NULL CHECK (char_length(phone) BETWEEN 7 AND 32),
    password_hash        TEXT NOT NULL,
    must_change_password BOOLEAN NOT NULL DEFAULT true,
    is_active            BOOLEAN NOT NULL DEFAULT true,
    role_id              UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- El correo se guarda normalizado (Q5): trim + minúsculas.
    CHECK (email = lower(btrim(email))),
    -- El teléfono se guarda sin espacios sobrantes; su formato telefónico (al menos
    -- 7 dígitos, separadores habituales, prefijo internacional opcional) lo valida
    -- `platform/validate` (etiqueta `phone`) en el service: la restricción de la
    -- base evita vacíos y basura, no sustituye a la validación de formato.
    CHECK (phone = btrim(phone))
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
- **Datos de la cuenta (confirmados por el humano el 2026-10-04, `research.md` R20)**: `first_name`
  (nombre) y `last_name` (apellidos) son **campos obligatorios separados**, `email` es la
  identificación de acceso y `phone` es **obligatorio** con formato telefónico razonable (FR-009:
  dígitos con espacios, guiones o paréntesis, prefijo internacional opcional y ≥ 7 dígitos). No hay
  documento de identidad: el riesgo R3 del plan queda **cerrado**.
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

## Almacenamiento en Redis (sesiones e intentos de acceso)

No hay más migraciones: `sessions` y `login_attempts` **no** se crean en PostgreSQL (decisión
confirmada del humano el 2026-10-04; `research.md` R1/R5). El contrato con Redis se documenta aquí
porque es parte del modelo de datos aunque viva fuera de la base relacional. Cliente:
`github.com/redis/go-redis/v9` (justificado en `research.md` R16); la interfaz `session.Store` y su
implementación viven en `internal/platform/session`, y los contadores de acceso se manejan desde el
dominio `usuarios`.

| Clave | Tipo | Valor | TTL | Para qué |
|---|---|---|---|---|
| `sess:<sha256(token)>` | string | JSON: `userId`, `createdAt`, `lastSeenAt`, `absoluteExpiresAt` | **30 min de inactividad**, refrescado en cada actividad y acotado a la vida absoluta (`min(30 min, absoluteExpiresAt - now)`) | Sesión de la cookie `ss_session` (FR-001/FR-005). El token en claro **nunca** se guarda |
| `user_sessions:<userId>` | set | hashes de token de las sesiones abiertas (la clave `sess:*` se reconstruye con `sess:<hash>`) | ≤ 1 h (vida absoluta) | Revocar **todas** las sesiones de una cuenta al desactivarla o al definir/restablecer su contraseña (FR-012, R17), sin `SCAN` |
| `login:fail:<identificador>` | string (contador) | número de fallos consecutivos | 15 min | FR-006: `INCR` por fallo, `DEL` al entrar bien; con 5 fallos se crea la bandera de bloqueo |
| `login:block:<identificador>` | string (bandera) | `"1"` | 15 min (la crea el 5.º fallo) | Bloqueo vigente → `429` con `Retry-After` = TTL restante |

Notas:

- `<identificador>` es el **correo normalizado** (trim + minúsculas), **exista o no la cuenta**: el
  comportamiento del bloqueo (mensaje, código y tiempos) es idéntico en ambos casos y por eso no
  revela existencia (FR-003/SC-008, `research.md` R5). Los valores **5 intentos** y **15 minutos**
  siguen siendo constantes de código (confirmados por el humano el 2026-10-04).
- **Vida absoluta de 1 h + inactividad de 30 min** (confirmadas el 2026-10-04, `research.md` R15):
  la vida absoluta no se implementa con TTL sino con `absoluteExpiresAt` **inmóvil** dentro del
  valor de la sesión; el TTL de la clave es siempre el de inactividad, acotado al tiempo que quede
  de vida absoluta. La cookie se emite con `Max-Age` de 1 h.
- Sesión válida solo si la clave existe, `now < absoluteExpiresAt` y **la cuenta sigue activa**:
  `authn` revalida la cuenta en cada petición (red de seguridad de FR-012). El `lastSeenAt` se
  escribe estrangulado a una vez por minuto.
- Redis guarda estado efímero: un reinicio sin persistencia solo obliga a volver a iniciar sesión
  (y reinicia los contadores de intentos — riesgo aceptado y registrado en el plan). Si en
  despliegue se exigiera dureza, se activa AOF/RDB.

## Consultas sqlc (`backend/internal/db/queries/`)

Todas parametrizadas, sin `SELECT *`, con `ORDER BY` determinista y `LIMIT`/`OFFSET` en listados
(§8.1.3 y §8.1.7). Inventario previsto (nombres de sqlc):

| Archivo | Consultas |
|---|---|
| `users.sql` | `InsertUser`, `GetUserByID`, `GetUserByEmail`, `GetUserAuthByEmail` (correo + hash + estado + rol + permisos, para login/`authn`), `ListUsers` (con `total` aparte), `CountUsers`, `UpdateUser` (nombre, apellidos, correo, teléfono, rol, activo), `UpdateUserPassword`, `SetUserMustChangePassword`, `CountActiveAdmins` (recuento **post-mutación** para FR-008), `CountUsersByRole` |
| `roles.sql` | `InsertRole`, `GetRoleByID`, `GetRoleByNameLower`, `ListRoles` (+ `CountRoles`), `UpdateRoleName`, `DeleteRolePermissions`, `InsertRolePermission`, `DeleteRole`, `CountRoleUsers` |
| `permissions.sql` | `ListPermissions` (catálogo), `GetPermissionIDsByCodes` (para crear/editar roles) |

**No hay `sessions.sql` ni `login_attempts.sql`**: la sesión y los contadores de acceso viven en
Redis (sección anterior; interfaz `session.Store` en `internal/platform/session` y contadores en el
dominio `usuarios`), fuera de sqlc.

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
| Teléfono con formato telefónico razonable (FR-009: ≥ 7 dígitos, separadores habituales, prefijo internacional opcional) | Service (`platform/validate`, etiqueta `phone`) | `CHECK (phone = btrim(phone))` + longitud 7–32 |
| Nombre de rol único normalizado (Q5) | Service (mensaje claro) | `UNIQUE (lower(name))` |
| Siempre ≥1 cuenta activa con `admin_usuarios_roles` (FR-008) | Service, transacción + advisory lock | Índice parcial `users_is_active_idx` |
| Inicialización única (FR-007) | Service, transacción + advisory lock + `users` vacío | — (la transacción serializada lo garantiza) |
| Bloqueo 5 intentos / 15 min (FR-006) | Service (contadores `login:fail:*` / `login:block:*` en Redis) | TTL de 15 min + clave por identificador normalizado (exista o no la cuenta) |
| Sesión válida solo con cuenta activa (FR-012) | `authn` en **cada petición** + revocación de claves al desactivar (`user_sessions:*`) | TTL de inactividad + `absoluteExpiresAt` inmóvil en Redis |

## Validación del modelo (qué se comprobará en implementación)

1. `make db-migrate` aplica `000002` y `000003` y `migrate down` las revierte por completo, en
   orden inverso.
2. Pruebas de integración (`//go:build integration`) contra PostgreSQL real:
   duplicados normalizados rechazados (`UNIQUE`), `ON DELETE RESTRICT`, `CHECK` de
   normalización del correo, **carrera anti-bloqueo** (dos transacciones concurrentes no dejan 0
   administradores) e **inicialización única** (dos peticiones simultáneas → una sola crea).
3. Pruebas de integración de Redis (real, levantado con `testcontainers-go` — ver `research.md`
   R19): crear/resolver sesión, TTL de inactividad y su refresco, corte por vida absoluta de 1 h,
   revocación por cuenta (`user_sessions:*`), contadores de intentos y bloqueo con `Retry-After`.
4. `make sqlc-verify` sin diferencias (el código generado acompaña al SQL).
5. `GET /healthz` sigue respondiendo igual tras las migraciones (el esquema nuevo no afecta a F1).
