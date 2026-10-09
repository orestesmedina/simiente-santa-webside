# Data Model — F2 Acceso y gestión de usuarios

**Fecha**: 2026-10-04 · **Rama**: `002-acceso-gestion-usuarios` · **Spec**: `spec.md` (aprobada;
**cambio de alcance: auditoría** re-aprobado el 2026-10-04 — US8, FR-021…FR-026)
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
| (no previsto) `login_attempts` | **No es tabla: contadores en Redis con TTL** | FR-006 (5 intentos / 15 min) sin enumerar cuentas (`research.md` R5). El estado es efímero por definición y Redis ya existe por la sesión: su TTL reemplaza a la limpieza manual. **No confundir con `login_events`** (abajo): esa tabla descartada era el *contador de bloqueo*; el **historial de auditoría** sí es tabla porque debe durar |
| (no previsto por F1) `login_events` | **Tabla nueva** *(cambio de alcance: auditoría, 2026-10-04)* | Historial duradero de intentos de acceso (FR-022); `research.md` R22 |
| (no previsto por F1) `admin_actions` | **Tabla nueva** *(cambio de alcance: auditoría, 2026-10-04)* | Historial duradero de acciones administrativas (FR-023); `research.md` R22 |

Reglas que se respetan de las convenciones: nombres en inglés y plural; `id UUID PRIMARY KEY
DEFAULT gen_random_uuid()`; `created_at`/`updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`; `NOT NULL`
por defecto; `TEXT` con `CHECK` de longitud; claves foráneas explícitas con `ON DELETE` decidido;
índices en FK y en las columnas de `WHERE`/`ORDER BY`; `UNIQUE`/`CHECK` **en la base**, no solo en el
código; `up`/`down` completos; **nunca** se edita una migración aplicada. **F2 consta de tres
migraciones, `000002`, `000003` y `000004`** (la numeración sigue desde el baseline `000001` de F1 y
no hay huecos; al no crear tablas de sesión ni de contadores no faltan números. `000004` es la del
**cambio de alcance: auditoría** y lleva además la proyección del último acceso en `users`).

## Resumen de tablas

```text
PostgreSQL:  permissions ──< role_permissions >── roles ──< users
                                                               ├─< login_events   (historial de accesos, FR-022)
                                                               ├─< admin_actions  (quién actuó, FR-023)
                                                               └── last_login_at / last_login_ip (proyección del último acceso, FR-021)
             roles >─ (admin_actions.target_role_id, ON DELETE SET NULL)

Redis:       sess:<sha256(token)>      → sesión (TTL 30 min, vida absoluta 1 h)
             user_sessions:<user_id>   → SET de sesiones abiertas (revocación por cuenta)
             login:fail:<identificador>  /  login:block:<identificador>  → contadores de intentos y bloqueo (FR-006)
```

| Tabla / almacén | Propietario (paquete) | Reglas que cubre |
|---|---|---|
| `permissions` | `internal/usuarios` | FR-015 (catálogo = módulos del producto) |
| `roles` | `internal/usuarios` | FR-014, FR-017 |
| `role_permissions` | `internal/usuarios` | FR-014 (≥1 permiso), FR-018 |
| `users` | `internal/usuarios` | FR-009…FR-013, FR-017 (un rol), FR-019, FR-021 (`last_login_at`/`last_login_ip`) |
| `login_events` | `internal/usuarios` (auditoría) | FR-022, FR-025, FR-026 |
| `admin_actions` | `internal/usuarios` (auditoría) | FR-023, FR-025, FR-026 |
| Redis: `sess:*` / `user_sessions:*` | `internal/platform/session` (implementación Redis de `session.Store`) | FR-001, FR-004, FR-005, FR-012 |
| Redis: `login:fail:*` / `login:block:*` | **mecanismo** en `internal/platform/session` (throttle sobre el `Store` de Redis) · **semántica y constantes** (`maxFailedAttempts = 5`, `lockoutDuration = 15 min`) en el dominio `internal/usuarios` | FR-006 (solo el **contador de bloqueo**, efímero) |

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
-- Los `code` son los identificadores estables del catálogo (los usan el contrato y `AuthzByModule`);
-- `admin_usuarios_roles` es "el permiso de administrar usuarios y roles" de la spec (FR-015/FR-016/FR-024).
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
  documento de identidad: el riesgo RG3 del plan queda **cerrado**.
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

### `000004_create_login_events_and_admin_actions.up.sql`

Migración del **cambio de alcance: auditoría** (US8, FR-021…FR-026). Añade las dos tablas duraderas
del registro y la proyección del último acceso en `users`. Se añade como migración propia (y no
editando el diseño de `000003`) para que el cambio de alcance quede aislado y reversible; la
numeración queda sin huecos: `000001` (F1) → `000002` → `000003` → `000004`.

```sql
-- Proyección del último acceso exitoso (FR-021): dos columnas en users para que la ficha y el
-- listado la muestren sin consultar el historial (ver "Último acceso por cuenta" más abajo).
-- Solo las escribe el login exitoso, en el mismo paso que inserta su fila en login_events.
ALTER TABLE users
    ADD COLUMN last_login_at TIMESTAMPTZ NULL,
    ADD COLUMN last_login_ip TEXT NULL CHECK (char_length(last_login_ip) BETWEEN 3 AND 45);

-- Historial de intentos de inicio de sesión (FR-022): una fila por intento, exitoso o fallido.
-- user_id queda NULL cuando el correo no corresponde a ninguna cuenta: el intento se registra
-- igual pero no se asocia a nada y no se crea ninguna cuenta (FR-003, sin "cuentas fantasma").
CREATE TABLE login_events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NULL REFERENCES users(id) ON DELETE RESTRICT,
    result     TEXT NOT NULL CHECK (result IN ('success', 'failure')),
    ip         TEXT NOT NULL CHECK (char_length(ip) BETWEEN 3 AND 45),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Orden por defecto del historial (§8.1.7) y filtro por cuenta + rango de fechas (FR-024).
CREATE INDEX login_events_created_at_id_idx ON login_events (created_at DESC, id DESC);
CREATE INDEX login_events_user_created_at_id_idx ON login_events (user_id, created_at DESC, id DESC);

-- Historial de acciones administrativas sensibles (FR-023): quién, qué, sobre qué, cuándo y con
-- qué resultado —incluidos los intentos que no se completan y los denegados por falta de permiso.
CREATE TABLE admin_actions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id  UUID NULL REFERENCES users(id) ON DELETE RESTRICT,
    action         TEXT NOT NULL CHECK (action IN (
                       'user.create', 'user.update', 'user.activate', 'user.deactivate',
                       'user.password_reset', 'role.create', 'role.update', 'role.delete')),
    target_kind    TEXT NOT NULL CHECK (target_kind IN ('user', 'role')),
    target_user_id UUID NULL REFERENCES users(id) ON DELETE RESTRICT,
    target_role_id UUID NULL REFERENCES roles(id) ON DELETE SET NULL,
    target_label   TEXT NULL CHECK (char_length(target_label) BETWEEN 1 AND 254),
    result         TEXT NOT NULL CHECK (result IN ('success', 'failure', 'denied')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Sin actor solo puede quedar la inicialización del sistema (FR-007), que cuenta como
    -- creación de cuenta y por eso se registra como 'user.create'.
    CHECK (actor_user_id IS NOT NULL OR action = 'user.create'),
    -- Cada objetivo usa la FK de su tipo. Ambas son anulables: una creación rechazada no llega a
    -- tener id, y un rol eliminado (FR-017) deja target_role_id en NULL conservando target_label.
    CHECK ((target_kind = 'user' AND target_role_id IS NULL) OR
           (target_kind = 'role' AND target_user_id IS NULL))
);

CREATE INDEX admin_actions_created_at_id_idx ON admin_actions (created_at DESC, id DESC);
CREATE INDEX admin_actions_actor_created_at_id_idx ON admin_actions (actor_user_id, created_at DESC, id DESC);
CREATE INDEX admin_actions_target_user_created_at_id_idx ON admin_actions (target_user_id, created_at DESC, id DESC);
CREATE INDEX admin_actions_target_role_id_idx ON admin_actions (target_role_id);
```

### `000004_create_login_events_and_admin_actions.down.sql`

```sql
DROP TABLE IF EXISTS admin_actions;
DROP TABLE IF EXISTS login_events;
ALTER TABLE users
    DROP COLUMN IF EXISTS last_login_ip,
    DROP COLUMN IF EXISTS last_login_at;
```

Notas de las tablas de auditoría:

- **Solo inserción (FR-025)**: no existe `UPDATE` ni `DELETE` sobre `login_events` ni
  `admin_actions` en ninguna consulta de sqlc, servicio ni contrato: el registro es de solo lectura
  desde el panel y se conserva aunque la cuenta se desactive o se le editen los datos (FR-012,
  FR-013, FR-025). Por eso `updated_at` es siempre igual a `created_at`; se mantiene para respetar
  la convención de esquema (`id`, `created_at`, `updated_at`), no porque haya nada que actualizar.
- **`created_at` es la fecha y hora del intento o de la acción**: la fila se inserta en el mismo
  proceso que lo registra, de modo que es el "cuándo" que piden FR-022/FR-023 y el campo por el que
  se filtra el rango de fechas (FR-024). El orden por defecto de ambos historiales es
  `created_at DESC, id DESC` (§8.1.7) y sus índices lo reflejan.
- **Sin credenciales en el registro (FR-026)**: de un inicio de sesión solo queda su resultado
  (`success`/`failure`); de un restablecimiento de contraseña solo quién lo hizo, sobre qué cuenta y
  cuándo (la acción `user.password_reset` no lleva ningún valor de contraseña). Tampoco se guarda
  el correo de un intento **no identificado**: solo `user_id` cuando la cuenta existe (mínimo dato
  necesario; los correos inventados no dejan "cuentas fantasma" ni datos de terceros en el registro).
- **Nombres y correos del contrato: derivados, nunca guardados (F-01)**: `userName`/`userEmail`
  (`AccessEventItem`) y `actorName`/`actorEmail` (`AdminActionItem`) se resuelven con un
  `LEFT JOIN users` al consultar (`ListLoginEvents`/`ListAdminActions`): el nombre es
  `firstName lastName` y el correo es el `users.email` actual. **No hay columnas de nombre ni de
  correo en las tablas del registro**: no se duplica el dato (solo `target_label` captura la
  etiqueta del momento). Cuando el intento no se asoció a ninguna cuenta (correo inexistente), esos
  campos son `NULL` y **no se guarda ni se muestra el correo probado** (FR-026): la interfaz
  muestra **"Intento sin cuenta asociada"**.
- **IP de origen**: texto de la dirección del par (`net.SplitHostPort(r.RemoteAddr)`), IPv4 o IPv6
  (longitud 3–45). No se lee `X-Forwarded-For` porque no hay proxy documentado en el MVP; si algún
  día lo hay, será una decisión nueva (ahora mismo se registraría la IP del proxy).
- **`admin_actions.action`**: los ocho códigos que pide FR-023. La **inicialización** (FR-007) se
  registra como `user.create` con `actor_user_id = NULL` —único caso permitido sin actor, por el
  `CHECK`— y `target_user_id` = la cuenta inicial: cuenta como creación de cuenta y debe registrarse.
  El rol "Administrador" que nace con ella forma parte de esa misma acción (no hubo un administrador
  que "creara un rol", que es lo que registra `role.create`).
- **Objetivos que desaparecen**: `users` nunca se borra (FR-013) → sus FK van con `ON DELETE
  RESTRICT` y no hay problema; los **roles sí se pueden eliminar** (FR-017) → `target_role_id` va
  con `ON DELETE SET NULL` y el registro sobrevive con `target_label` (el nombre del rol en el
  momento de la acción). Sin ese `target_label`, el "sobre qué" del registro dejaría de responderse
  al eliminar un rol, y FR-025 exige que el registro se conserve íntegro.
- **`result`**: `login_events` usa solo `success`/`failure` (FR-022: "exitoso/fallido"; los intentos
  durante un bloqueo temporal de FR-006 se registran como `failure`). `admin_actions` añade
  `denied` para separar la denegación por falta de permiso (FR-016) del intento que falló por datos
  inválidos, duplicado o por la regla anti-bloqueo; ambas clases "no se completan" y las dos deben
  quedar registradas (Edge Cases de la spec).
- **Qué pasa si el registro no se puede escribir**: el éxito de una operación sensible y su registro
  van en la **misma transacción** (o ambos, o ninguno); un login exitoso también deja su fila de
  `login_events` y su `last_login_*` antes de emitir la sesión (**sin registro, sin acceso**). Si lo
  que falla es el registro de un **intento que ya va a fallar** (un login con mala contraseña, una
  operación denegada), se loguea el error con `request_id` y se devuelve el error original: el
  registro nunca cambia la respuesta que ve la persona, pero su fallo queda en el log.

## Último acceso por cuenta (FR-021) — columnas en `users`, no derivado

- **Decisión**: `users.last_login_at` (fecha y hora) y `users.last_login_ip` (origen) son dos
  columnas **anulables** actualizadas **solo por el login exitoso**, en el mismo paso que inserta la
  fila en `login_events`. Una cuenta que nunca ha iniciado sesión las tiene en `NULL` y la ficha lo
  indica sin inventar ningún acceso (US8 esc. 6). La **ficha** de la cuenta es su
  **detalle/edición** (`GET /admin/usuarios/{id}` → `UserItem`): la UI muestra
  `lastLoginAt`/`lastLoginIp` ahí **y** en la fila del listado (`GET /admin/usuarios`).
- **Por qué columnas y no derivado del historial**: la ficha (`GET /admin/usuarios/{id}`) y el
  listado muestran el último acceso sin una subconsulta sobre una tabla que crece sin límite; y si
  algún día hay una política de retención/purga del registro (hoy fuera de alcance, riesgo RG17 del
  plan), el último acceso de la ficha **debe seguir mostrándose**. El coste son dos columnas y un
  `UPDATE` puntual por login exitoso.
- **Alternativa descartada**: derivarlo de `login_events` con
  `WHERE user_id = $1 AND result = 'success' ORDER BY created_at DESC LIMIT 1` (índice parcial).
  Funciona, pero añade una consulta por cuenta en cada listado, hace que la ficha dependa del
  crecimiento y la retención del historial y complica el DTO. La fuente de la verdad del "qué pasó"
  sigue siendo `login_events`; las columnas son su **proyección** y hay una prueba de coherencia:
  tras cualquier login, el `last_login_at` de la ficha coincide con la fila `success` más reciente
  del historial de esa cuenta.
- **Quién las escribe**: solo el flujo de login exitoso. Ningún DTO de entrada las acepta
  (`UpdateUserInput` no las incluye): no se pueden editar desde la API ni desde el panel.

## Almacenamiento en Redis (sesiones y contadores de acceso)

Quedan **fuera de PostgreSQL** la sesión y los contadores de FR-006: `sessions` y `login_attempts`
(el contador de bloqueo) **no** se crean como tablas (decisión confirmada del humano el 2026-10-04;
`research.md` R1/R5). El contrato con Redis se documenta aquí porque es parte del modelo de datos
aunque viva fuera de la base relacional. Cliente: `github.com/redis/go-redis/v9` (justificado en
`research.md` R16); la interfaz `session.Store` y su implementación viven en
`internal/platform/session`, y los contadores de acceso se manejan con ese mismo `Store`/throttle desde
`internal/platform/session`, mientras que su **semántica y sus constantes** (5 fallos / 15 min) las
decide el dominio `usuarios` (reparto F-13: el mecanismo efímero vive en el plumbing, la regla de
negocio en el dominio).

| Clave | Tipo | Valor | TTL | Para qué |
|---|---|---|---|---|
| `sess:<sha256(token)>` | string | JSON: `userId`, `createdAt`, `lastSeenAt`, `absoluteExpiresAt` | **30 min de inactividad**, refrescado en cada actividad y acotado a la vida absoluta (`min(30 min, absoluteExpiresAt - now)`) | Sesión de la cookie `ss_session` (FR-001/FR-005). El token en claro **nunca** se guarda |
| `user_sessions:<userId>` | set | hashes de token de las sesiones abiertas (la clave `sess:*` se reconstruye con `sess:<hash>`) | ≤ 1 h (vida absoluta) | Revocar **todas** las sesiones de una cuenta al desactivarla o al definir/restablecer su contraseña (FR-012, R17), sin `SCAN` |
| `login:fail:<identificador>` | string (contador) | número de fallos consecutivos | 15 min | FR-006: `INCR` por **cada** fallo, `DEL` al entrar bien; el **5.º fallo** crea la bandera de bloqueo (y aún responde el `401` genérico) |
| `login:block:<identificador>` | string (bandera) | `"1"` | 15 min (la crea el 5.º fallo) | Con la bandera vigente, **desde el 6.º intento** cada intento responde `429` con `Retry-After` = TTL restante |

Notas:

- `<identificador>` es el **correo normalizado** (trim + minúsculas), **exista o no la cuenta**: el
  comportamiento del bloqueo (mensaje, código y tiempos) es idéntico en ambos casos y por eso no
  revela existencia (FR-003/SC-008, `research.md` R5). Los valores **5 intentos** y **15 minutos**
  siguen siendo constantes de código (confirmados por el humano el 2026-10-04).
- **Semántica del 5.º intento (FR-006)**: el contador se incrementa con **cada** fallo; el **5.º
  fallo** responde el error genérico `401` **y crea el bloqueo** (bandera `login:block:*`); desde el
  **6.º intento** —y durante los 15 minutos— cada intento responde `429` con el mensaje de bloqueo y
  `Retry-After`. Pasados los 15 min se puede volver a intentar.
- **Vida absoluta de 1 h + inactividad de 30 min** (confirmadas el 2026-10-04, `research.md` R15):
  la vida absoluta no se implementa con TTL sino con `absoluteExpiresAt` **inmóvil** dentro del
  valor de la sesión; el TTL de la clave es siempre el de inactividad, acotado al tiempo que quede
  de vida absoluta. La cookie se emite con `Max-Age` de 1 h.
- Sesión válida solo si la clave existe, `now < absoluteExpiresAt` y **la cuenta sigue activa**:
  `authn` revalida la cuenta en cada petición (red de seguridad de FR-012). El `lastSeenAt` se
  escribe estrangulado a una vez por minuto.
- Redis guarda **estado efímero y sin persistencia** (confirmado por el humano el 2026-10-04): el
  servicio `redis` de `docker-compose.yml` corre **sin `appendonly`, sin `save` y sin volumen**. Un
  reinicio solo obliga a volver a iniciar sesión y reinicia los contadores de intentos (riesgo
  aceptado, RG16 del plan; `research.md` R21); **no toca la auditoría**, que vive en PostgreSQL.

### Relación Redis ↔ PostgreSQL (contadores vs. auditoría)

Son dos responsabilidades distintas y **no se solapan** (decisión del cambio de alcance,
`research.md` R21/R22):

| | **Redis** (estado efímero) | **PostgreSQL** (registro duradero) |
|---|---|---|
| Qué guarda | Sesiones (`sess:*`, `user_sessions:*`) y **contadores de bloqueo** de FR-006 (`login:fail:*`, `login:block:*`) | **Auditoría**: `login_events` (cada intento de acceso) y `admin_actions` (cada acción administrativa) |
| Cuánto dura | TTL 15 min (contadores) / 30 min + vida absoluta 1 h (sesiones) | Indefinido en el MVP: sin purga ni retención (Out of Scope; riesgo RG17 del plan) |
| Para qué | Aplicar **ahora** el bloqueo de 5 intentos / 15 min y mantener la sesión | Responder **después** qué pasó, quién, desde dónde y cuándo (US8, FR-021…FR-025) |
| Si se pierde | Reinicio de Redis: se reabre la ventana de intentos y hay que volver a entrar | No se pierde: es la fuente duradera |

El **mismo intento de login** deja huella en los dos lados con distinto objetivo: incrementa el
contador efímero (`login:fail:<correo>`, que decide si se bloquea) y añade una fila duradera en
`login_events` (que cuenta para el historial y para el último acceso). Un reinicio de Redis no borra
ninguna fila del historial; una purga futura del historial (si la hubiera) no afectaría a los
contadores, porque son de minutos.

## Consultas sqlc (`backend/internal/db/queries/`)

Todas parametrizadas, sin `SELECT *`, con `ORDER BY` determinista y `LIMIT`/`OFFSET` en listados
(§8.1.3 y §8.1.7). Inventario previsto (nombres de sqlc):

| Archivo | Consultas |
|---|---|
| `users.sql` | `InsertUser`, `GetUserByID`, `GetUserByEmail`, `GetUserAuthByEmail` (correo + hash + estado + rol + permisos, para login/`authn`), `ListUsers` (con `total` aparte), `CountUsers`, `UpdateUser` (nombre, apellidos, correo, teléfono, rol, activo), `UpdateUserPassword`, `SetUserMustChangePassword`, `UpdateUserLastLogin` (`last_login_at`/`last_login_ip`, solo por login exitoso — FR-021), `CountActiveAdmins` (recuento **post-mutación** para FR-008), `CountUsersByRole` |
| `roles.sql` | `InsertRole`, `GetRoleByID`, `GetRoleByNameLower`, `ListRoles` (+ `CountRoles`), `UpdateRoleName`, `DeleteRolePermissions`, `InsertRolePermission`, `DeleteRole`, `CountRoleUsers` |
| `permissions.sql` | `ListPermissions` (catálogo), `GetPermissionIDsByCodes` (para crear/editar roles) |
| `audit.sql` | `InsertLoginEvent`, `ListLoginEvents` (+ `CountLoginEvents`, filtros `userId`/`from`/`to`, con `LEFT JOIN users` para derivar `userEmail`/`userName`), `InsertAdminAction`, `ListAdminActions` (+ `CountAdminActions`, filtros por cuenta involucrada `actor OR target` y rango de fechas, con `LEFT JOIN users` para derivar `actorEmail`/`actorName`). **Sin `UPDATE`/`DELETE`**: el registro es de solo inserción (FR-025) |

**No hay `sessions.sql` ni contadores de intentos en sqlc**: la sesión y el bloqueo de FR-006 viven
en Redis (sección anterior; interfaz `session.Store` en `internal/platform/session` y contadores en
el dominio `usuarios`), fuera de sqlc. El **historial** de accesos sí es SQL (`audit.sql`): no se
confunde con el contador descartado de `login_attempts`.

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
| Bloqueo 5 intentos / 15 min (FR-006) | Service (contadores `login:fail:*` / `login:block:*` en Redis; mecanismo en `platform/session`, constantes en el dominio) | TTL de 15 min + clave por identificador normalizado (exista o no la cuenta). El **5.º fallo** crea el bloqueo y aún responde `401`; **desde el 6.º** intento → `429` |
| Sesión válida solo con cuenta activa (FR-012) | `authn` en **cada petición** + revocación de claves al desactivar (`user_sessions:*`) | TTL de inactividad + `absoluteExpiresAt` inmóvil en Redis |
| Todo intento de acceso queda en `login_events` (FR-022), asociado a la cuenta solo si existe | Service (`service_auth.go`, los cuatro desenlaces: éxito, fallo, cuenta inactiva, bloqueo) | `user_id` anulable con FK; sin columna de correo para intentos no identificados |
| Toda acción administrativa sensible queda en `admin_actions`, también si falla o se deniega (FR-023) | Service (éxito y fallo de negocio) + `handler` (JSON inválido/validación) + `authz` (denegación) — P20 del plan | `result` con `success`/`failure`/`denied`; `CHECK` de acción/actor |
| Registro de solo lectura (FR-025) | No existe operación | Sin `UPDATE`/`DELETE` en `audit.sql`; contrato solo con `GET` sobre `/admin/auditoria/*` |
| Sin credenciales en el registro (FR-026) | Service + revisiones de `seguridad` | Ninguna columna admite contraseñas: solo resultado, actor, objetivo y fechas |
| Último acceso = último login exitoso (FR-021) | Service (`UpdateUserLastLogin` solo en el éxito) | Columnas anulables en `users`; ninguna vía de API las escribe |

## Validación del modelo (qué se comprobará en implementación)

1. `make db-migrate` aplica `000002`, `000003` y `000004` y `migrate down` las revierte por completo,
   en orden inverso (incluidas las columnas `last_login_*` de `users`).
2. Pruebas de integración (`//go:build integration`) contra PostgreSQL real:
   duplicados normalizados rechazados (`UNIQUE`), `ON DELETE RESTRICT`, `CHECK` de
   normalización del correo, **carrera anti-bloqueo** (dos transacciones concurrentes no dejan 0
   administradores) e **inicialización única** (dos peticiones simultáneas → una sola crea).
3. Pruebas de integración de las tablas de auditoría: cada desenlace de login inserta su fila
   (incluido el correo inexistente, con `user_id` NULL y sin crear nada), el login exitoso
   actualiza `last_login_*` y coincide con la última fila `success` del historial, los filtros por
   cuenta y rango de fechas y la paginación devuelven lo esperado, eliminar un rol deja el registro
   con `target_role_id` NULL y `target_label` intacto, y **no existe** consulta de `UPDATE`/`DELETE`
   sobre `login_events`/`admin_actions`.
4. Pruebas de integración de Redis (real, levantado con `testcontainers-go` — ver `research.md`
   R19): crear/resolver sesión, TTL de inactividad y su refresco, corte por vida absoluta de 1 h,
   revocación por cuenta (`user_sessions:*`), contadores de intentos y bloqueo con `Retry-After`.
5. `make sqlc-verify` sin diferencias (el código generado acompaña al SQL).
6. `GET /healthz` sigue respondiendo igual tras las migraciones (el esquema nuevo no afecta a F1).
