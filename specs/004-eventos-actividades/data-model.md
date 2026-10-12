# Data Model — F4 Eventos y actividades

**Fecha**: 2026-10-11 · **Spec**: [spec.md](./spec.md) · **Decisiones**: [research.md](./research.md)
(R4-1, R4-2, R4-4, R4-5, R4-9, R4-10, R4-12) · **Plantilla de estilo**: el `data-model.md` de F3.

> **Actualizado el 2026-10-11 (puerta 2)**: el modelo de tablas **no cambia**; la única novedad de
> datos es el **permiso nuevo `categorias`** — 10.º del catálogo de F2 — que vive en la tabla
> `permissions` de F2 y se siembra en la migración `000009` (ver abajo).

Fuente de verdad del esquema: **este documento** → migraciones `backend/migrations/` → consultas
`backend/internal/db/queries/agenda.sql` y `media.sql` (sqlc). Convenciones: skill `postgres-db`,
§VI de la constitución y §8.1 de `docs/tecnico/arquitectura.md` (UUID, `created_at`/`updated_at`,
mapeo `pgtype` → dominio en el repository, consultas parametrizadas).

## Patrones transversales heredados de F3 (no se reabren)

- **Contenido bilingüe (R3-1)**: cada campo traducible vive en dos columnas, `*_es`
  (**obligatorio**, idioma base) y `*_en` (**opcional**). `""`/espacios se guardan como `NULL`
  (normalización I6: trim en el service antes de persistir). No traducibles: `map_url`,
  `image_file`, `event_date`/`activity_date`, `start_time`/`end_time`.
- **Estado de publicación (R3-3)**: `publication_state TEXT NOT NULL DEFAULT 'draft' CHECK
  (publication_state IN ('draft','published'))` **por elemento** (evento y actividad; las
  categorías no lo tienen: son catálogo, R4-5). Las consultas del visitante filtran
  `publication_state = 'published'` **y** la condición "próximo" (R4-2) en la **misma** consulta
  (`…Upcoming`): es la única vía pública y no existe ninguna función que devuelva al visitante un
  borrador o un elemento vencido (FR-013/SC-002, SC-009).
- **Regla de "próximo" (R4-2)**: toda consulta pública recibe `$today DATE` y `$now_time TIME`
  calculados por el service en la zona horaria `APP_TZ` y aplicados como parámetros:

  ```sql
  (x.event_date > sqlc.narg('today')) OR (x.event_date = sqlc.narg('today')
      AND (x.end_time IS NULL OR x.end_time >= sqlc.narg('now_time')))
  ```

- **Auditoría (R3-11)**: las mutaciones insertan `admin_actions` con `InsertAdminAction` (consulta
  compartida de F2) **dentro de su transacción**; objetivo `target_kind='content'`, ambas FK en
  `NULL` y `target_label` descriptivo. Registro de solo lectura (sin `UPDATE`/`DELETE`).

## Resumen de tablas

| Tabla | Cardinalidad | Qué guarda |
|---|---|---|
| `agenda_categories` | 0..N (catálogo, ≤100 acotado en el service) | Categoría opcional (acción social, predicación…): nombre es/en y orden (Q6, R4-5) |
| `agenda_events` | 0..N | Evento (encuentro periódico): título, texto de frecuencia, detalle, lugar + enlace de mapa, organizador, categoría, imagen + alt, estado (FR-018) |
| `agenda_event_dates` | 1..N por evento (≤50, acotado en el service) | Fecha concreta del evento con hora de inicio/fin opcionales (FR-004/FR-005, Q1/Q2) |
| `agenda_activities` | 0..N | Actividad (acción puntual): todos los campos + **una** fecha con horas opcionales (FR-019, Decisión 3) |
| `admin_actions` (**existente, F2**) | — | Se **amplía** el registro cerrado: 15 códigos `event.*`/`activity.*`/`category.*` (FR-017, R4-10) |
| `permissions` (**existente, F2**) | — | Se **amplía** el catálogo fijo de F2 con la fila **`categorias`** (10.º permiso; R4-4, cambio de alcance de la puerta 2). Solo un `INSERT` en la migración `000009`: el resto del modelo de permisos de F2 (`roles`, `role_permissions`) **no cambia** |

Ninguna tabla de F4 tiene FK hacia `users`: el "quién" vive en `admin_actions.actor_user_id`
(FR-017). Únicas FK nuevas: `agenda_event_dates.event_id → agenda_events` y
`agenda_events.category_id` / `agenda_activities.category_id → agenda_categories` (R4-5).

---

## Migraciones

Tres migraciones nuevas (F3 terminó en `000006`): **`000007`** crea las tablas y siembra el
catálogo de categorías; **`000008`** amplía el registro de auditoría; **`000009`** siembra el
permiso **`categorias`** en la tabla `permissions` de F2. Ninguna migración existente se toca (§VI):
`000002_create_roles_and_permissions` está **aplicada** y su catálogo se amplía solo por `INSERT`
nuevo, nunca editando la migración vieja.

### `000007_create_agenda_content.up.sql`

```sql
-- 000007_create_agenda_content (up)
-- F4 Eventos y actividades: modelos separados (Decisión 3) con contenido
-- bilingüe es (obligatorio) / en (opcional) y estado de publicación por
-- elemento (research R4-1/R4-3/R4-12). Convenciones: skill `postgres-db` y
-- specs/004-eventos-actividades/data-model.md (fuente de verdad).

-- Catálogo de categorías compartido por eventos y actividades (Q6, R4-5):
-- opcional en el elemento, administrable desde el panel, borrado restringido
-- mientras esté en uso (FK RESTRICT abajo). Catálogo inicial sembrado al final.
CREATE TABLE agenda_categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name_es    TEXT NOT NULL CHECK (char_length(name_es) BETWEEN 1 AND 120),
    name_en    TEXT NULL CHECK (char_length(name_en) BETWEEN 1 AND 120),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (name_es),
    CHECK (name_es = btrim(name_es))
);

CREATE INDEX agenda_categories_sort_order_id_idx ON agenda_categories (sort_order, id);

-- Evento = encuentro periódico (FR-018). La periodicidad NO se calcula: es la
-- lista explícita de fechas de agenda_event_dates (Q1) + el texto opcional de
-- frecuencia (frequency_es/frequency_en). Lugar (texto) obligatorio + enlace de
-- mapa opcional (default revisable de la spec); imagen opcional con texto
-- alternativo obligatorio en español si hay imagen (FR-020).
CREATE TABLE agenda_events (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title_es          TEXT NOT NULL CHECK (char_length(title_es) BETWEEN 1 AND 160),
    title_en          TEXT NULL CHECK (char_length(title_en) BETWEEN 1 AND 160),
    frequency_es      TEXT NULL CHECK (char_length(frequency_es) BETWEEN 1 AND 160),
    frequency_en      TEXT NULL CHECK (char_length(frequency_en) BETWEEN 1 AND 160),
    detail_es         TEXT NOT NULL CHECK (char_length(detail_es) BETWEEN 1 AND 2000),
    detail_en         TEXT NULL CHECK (char_length(detail_en) BETWEEN 1 AND 2000),
    place_es          TEXT NOT NULL CHECK (char_length(place_es) BETWEEN 1 AND 200),
    place_en          TEXT NULL CHECK (char_length(place_en) BETWEEN 1 AND 200),
    map_url           TEXT NULL CHECK (char_length(map_url) BETWEEN 1 AND 500),
    organizer_es      TEXT NULL CHECK (char_length(organizer_es) BETWEEN 1 AND 160),
    organizer_en      TEXT NULL CHECK (char_length(organizer_en) BETWEEN 1 AND 160),
    category_id       UUID NULL REFERENCES agenda_categories (id) ON DELETE RESTRICT,
    image_file        TEXT NULL CHECK (char_length(image_file) BETWEEN 1 AND 120),
    image_alt_es      TEXT NULL CHECK (char_length(image_alt_es) BETWEEN 1 AND 300),
    image_alt_en      TEXT NULL CHECK (char_length(image_alt_en) BETWEEN 1 AND 300),
    publication_state TEXT NOT NULL DEFAULT 'draft'
                      CHECK (publication_state IN ('draft', 'published')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- FR-020: toda imagen lleva texto alternativo en español (idioma base).
    CHECK (image_file IS NULL OR image_alt_es IS NOT NULL),
    CHECK (title_es = btrim(title_es)),
    CHECK (place_es = btrim(place_es)),
    CHECK (map_url IS NULL OR map_url = btrim(map_url))
);

CREATE INDEX agenda_events_created_at_id_idx ON agenda_events (created_at DESC, id DESC);

-- Fechas concretas del evento (Q1/Q2, R4-1): cada fila es un día concreto con
-- hora de inicio y hora de fin opcionales (rango dentro del mismo día). Una
-- fecha por día por evento (UNIQUE); fin exige inicio y nunca es anterior
-- (FR-015). Se permite una fecha pasada (edge case de la spec): el filtro
-- "próximo" lo resuelve la consulta pública (R4-2), no la tabla.
CREATE TABLE agenda_event_dates (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id   UUID NOT NULL REFERENCES agenda_events (id) ON DELETE CASCADE,
    event_date DATE NOT NULL,
    start_time TEXT NULL CHECK (start_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
    end_time   TEXT NULL CHECK (end_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (event_id, event_date),
    CHECK (end_time IS NULL OR start_time IS NOT NULL),
    CHECK (end_time IS NULL OR end_time > start_time)
);

CREATE INDEX agenda_event_dates_date_idx ON agenda_event_dates (event_date);

-- Actividad = acción puntual (FR-019, Decisión 3): UNA fecha concreta con su
-- horario opcional, en la misma fila (sin tabla de fechas ni periodicidad).
-- El duplicado exacto del módulo (título+fecha+lugar normalizados) se rechaza
-- (edge case de la spec, R4-9); el service lo avisa antes con 409.
CREATE TABLE agenda_activities (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title_es          TEXT NOT NULL CHECK (char_length(title_es) BETWEEN 1 AND 160),
    title_en          TEXT NULL CHECK (char_length(title_en) BETWEEN 1 AND 160),
    detail_es         TEXT NOT NULL CHECK (char_length(detail_es) BETWEEN 1 AND 2000),
    detail_en         TEXT NULL CHECK (char_length(detail_en) BETWEEN 1 AND 2000),
    place_es          TEXT NOT NULL CHECK (char_length(place_es) BETWEEN 1 AND 200),
    place_en          TEXT NULL CHECK (char_length(place_en) BETWEEN 1 AND 200),
    map_url           TEXT NULL CHECK (char_length(map_url) BETWEEN 1 AND 500),
    organizer_es      TEXT NULL CHECK (char_length(organizer_es) BETWEEN 1 AND 160),
    organizer_en      TEXT NULL CHECK (char_length(organizer_en) BETWEEN 1 AND 160),
    category_id       UUID NULL REFERENCES agenda_categories (id) ON DELETE RESTRICT,
    image_file        TEXT NULL CHECK (char_length(image_file) BETWEEN 1 AND 120),
    image_alt_es      TEXT NULL CHECK (char_length(image_alt_es) BETWEEN 1 AND 300),
    image_alt_en      TEXT NULL CHECK (char_length(image_alt_en) BETWEEN 1 AND 300),
    activity_date     DATE NOT NULL,
    start_time        TEXT NULL CHECK (start_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
    end_time          TEXT NULL CHECK (end_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
    publication_state TEXT NOT NULL DEFAULT 'draft'
                      CHECK (publication_state IN ('draft', 'published')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (title_es, activity_date, place_es),
    CHECK (image_file IS NULL OR image_alt_es IS NOT NULL),
    CHECK (end_time IS NULL OR start_time IS NOT NULL),
    CHECK (end_time IS NULL OR end_time > start_time),
    CHECK (title_es = btrim(title_es)),
    CHECK (place_es = btrim(place_es)),
    CHECK (map_url IS NULL OR map_url = btrim(map_url))
);

CREATE INDEX agenda_activities_created_at_id_idx ON agenda_activities (created_at DESC, id DESC);
CREATE INDEX agenda_activities_date_idx ON agenda_activities (activity_date);

-- Catálogo inicial (Q6). name_en son traducciones de partida revisables desde
-- el panel (Decisión 8: el inglés es opcional; si se vacía, el público muestra
-- el español). sort_order deja "otro" al final.
INSERT INTO agenda_categories (name_es, name_en, sort_order) VALUES
    ('Acción social',   'Social action',   10),
    ('Predicación',     'Preaching',       20),
    ('Adoración',       'Worship',         30),
    ('Entrevista',      'Interview',       40),
    ('Noche de oración','Prayer night',    50),
    ('Otro',            'Other',           999);
```

### `000007_create_agenda_content.down.sql`

```sql
-- 000007_create_agenda_content (down) — revierte 000007 por completo (§VI).
-- El borrado del catálogo sembrado es implícito al tirar la tabla.
DROP TABLE IF EXISTS agenda_activities;
DROP TABLE IF EXISTS agenda_event_dates;
DROP TABLE IF EXISTS agenda_events;
DROP TABLE IF EXISTS agenda_categories;
```

### `000008_extend_admin_actions_for_agenda.up.sql`

Amplía el registro de auditoría de F2 (FR-017/R4-10). El `CHECK` que se sustituye es
`admin_actions_action_check`, **con nombre propio** desde la migración `000006` de F3 (F3 ya hizo
este mismo ejercicio y su prueba de integración lo verifica con up → down → up). El
`target_kind` **no cambia**: los elementos de F4 son `content` como los de F3. Los **15 códigos**
son 6 `event.*`, 6 `activity.*` y **3 `category.*`** (`category.create`, `category.update`,
`category.delete`: los tres del catálogo de categorías, incluidos en la lista de abajo).

```sql
-- 000008_extend_admin_actions_for_agenda (up)
-- F4: la auditoría de F2 se amplía a eventos, actividades y categorías (FR-017).
-- El registro de códigos de acción sigue siendo CERRADO y lo garantiza la BD.

ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_action_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_action_check CHECK (action IN (
    'user.create', 'user.update', 'user.activate', 'user.deactivate', 'user.password_reset',
    'role.create', 'role.update', 'role.delete',
    'home.identity.update', 'home.about.update', 'home.contact.update',
    'home.schedule.create', 'home.schedule.update', 'home.schedule.delete',
    'home.whatsapp.create', 'home.whatsapp.update', 'home.whatsapp.delete',
    'home.social.create', 'home.social.update', 'home.social.delete',
    'home.image.upload',
    'home.publish', 'home.unpublish',
    'event.create', 'event.update', 'event.delete', 'event.publish', 'event.unpublish',
    'event.image.upload',
    'activity.create', 'activity.update', 'activity.delete', 'activity.publish',
    'activity.unpublish', 'activity.image.upload',
    'category.create', 'category.update', 'category.delete'
));
```

### `000008_extend_admin_actions_for_agenda.down.sql`

```sql
-- 000008_extend_admin_actions_for_agenda (down): restaura el CHECK de F3
-- (8 códigos de F2 + 15 de F3). Como en el down de 000006, es destructivo a
-- propósito: las filas con códigos F4 no caben en el CHECK anterior y se
-- eliminan antes de restaurar. Solo tiene sentido en desarrollo/rollback.
DELETE FROM admin_actions WHERE action IN (
    'event.create', 'event.update', 'event.delete', 'event.publish', 'event.unpublish',
    'event.image.upload',
    'activity.create', 'activity.update', 'activity.delete', 'activity.publish',
    'activity.unpublish', 'activity.image.upload',
    'category.create', 'category.update', 'category.delete'
);

ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_action_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_action_check CHECK (action IN (
    'user.create', 'user.update', 'user.activate', 'user.deactivate', 'user.password_reset',
    'role.create', 'role.update', 'role.delete',
    'home.identity.update', 'home.about.update', 'home.contact.update',
    'home.schedule.create', 'home.schedule.update', 'home.schedule.delete',
    'home.whatsapp.create', 'home.whatsapp.update', 'home.whatsapp.delete',
    'home.social.create', 'home.social.update', 'home.social.delete',
    'home.image.upload',
    'home.publish', 'home.unpublish'
));
```

### `000009_seed_permission_categorias.up.sql`

Amplía el **catálogo de permisos de F2** (R4-4, cambio de alcance autorizado por el humano en la
puerta 2 del 2026-10-11): la gestión del catálogo de categorías de agenda exige su propio permiso,
el **10.º** del catálogo. La fila vive en la tabla `permissions` creada por
`000002_create_roles_and_permissions` (F2), que está aplicada y **no se edita**: F4 solo hace un
`INSERT` nuevo. Mismo gesto de sembrado que F2/F3 (catálogo como datos, no como `CHECK`).

```sql
-- 000009_seed_permission_categorias (up)
-- F4 (cambio de alcance aprobado el 2026-10-11): el catálogo de permisos de
-- F2 (000002) crece con un décimo permiso, `categorias`, que autoriza la
-- gestión del catálogo de categorías de agenda (/api/v1/admin/categorias).
-- Los permisos `eventos` y `actividades` NO autorizan este módulo (son
-- independientes; research R4-4).
INSERT INTO permissions (code, label) VALUES
    ('categorias', 'Categorías de agenda');
```

### `000009_seed_permission_categorias.down.sql`

```sql
-- 000009_seed_permission_categorias (down): retira la fila del permiso.
-- role_permissions.permission_id es ON DELETE RESTRICT (000002): primero se
-- retiran las asignaciones de ESTE permiso (solo las suyas; el resto de los
-- permisos de cada rol no se toca) y luego la fila. Destructivo a propósito,
-- como el down de 000006/000008: solo tiene sentido en desarrollo/rollback.
DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE code = 'categorias');
DELETE FROM permissions WHERE code = 'categorias';
```

**Notas de diseño del permiso.**

- `code = 'categorias'` es el identificador estable (lo usan `AuthzByModule`, el contrato y
  `lib/permissions.ts`); `label = 'Categorías de agenda'` es la etiqueta que ve el equipo en el
  formulario de roles de F2.
- **El administrador inicial lo recibe automáticamente**: el init de F2 (`Initialize`) lee el
  catálogo dinámicamente y crea el rol «Administrador» con **todos** los permisos (ahora 10); sin
  cambios de código en F2.
- **Las pruebas de F2 que fijaban el catálogo en 9 se actualizan a 10** como parte de F4
  (excepción aprobada, RG4-10 de [plan.md](./plan.md)): `repository_roles_integration_test.go:52`,
  `service_init_integration_test.go:57`, `service_init_test.go:199` (señaladas por el humano) y,
  detectadas en revisión de este plan, `service_roles_integration_test.go:69` y
  `service_roles_test.go:464`, más la fixture `sampleCatalog()` y los comentarios "los 9 permisos".
- Un `DELETE`/`UPDATE` sobre `permissions` por fuera de esta migración **no existe** en el código:
  el catálogo solo crece por migración (igual que en F2/F3).

---

## Consultas sqlc

### `backend/internal/db/queries/agenda.sql` (dominio `agenda`)

Listas con `ORDER BY` explícito y determinista (§8.1.7). Sin `SELECT *`. Todo parametrizado
(§IV, CWE-89).

| Consulta | Uso |
|---|---|
| `ListCategories` / `GetCategoryByID` | Catálogo para el panel (`ORDER BY sort_order, id`) y validación de `categoryId` |
| `InsertCategory` / `UpdateCategory` / `DeleteCategory` | CRUD del catálogo (el `DELETE` falla con `FK RESTRICT` si está en uso → `409`) |
| `CountCategoryUsage` | Mensaje del `409` de borrado ("está en uso por N elementos") |
| `CountCategories` | Tope del catálogo en el alta de categoría (1 conteo, sin listar el catálogo; `MaxCategories`, tope blando — M2) |
| `InsertEvent`, `UpdateEvent`, `DeleteEvent`, `GetEventByID` | Gestión del evento |
| `ReplaceEventDates` (`DELETE` + `INSERT` de la lista, en la transacción del guardado) | Reemplazo completo de fechas (R4-11) |
| `ListEventDates` (por evento) / `ListEventDatesForEvents` y `ListUpcomingEventDatesForEvents` (por lote, `event_id = ANY(...)`) | Fechas de los eventos: caminos de un solo elemento (`ListEventDates`), panel por página (`ListEventDatesForEvents`, todas, con `isPast` calculado en el service) / público (`ListUpcomingEventDatesForEvents`, solo próximas, R4-2) |
| `ExistsEventDuplicate` | Duplicado del módulo (título+lugar normalizados + fecha en común, R4-9) |
| `ListEvents` / `CountEvents` | Panel paginado (`ORDER BY created_at DESC, id DESC`, filtro opcional `publication_state`) |
| `ListUpcomingEvents` | Público: `publication_state='published'` **y** `EXISTS` una fecha próxima (R4-2); `ORDER BY` por la primera fecha próxima, `title_es, id` de desempate (orden estable, edge case) |
| `InsertActivity`, `UpdateActivity`, `DeleteActivity`, `GetActivityByID` | Gestión de la actividad |
| `ListActivities` / `CountActivities` | Panel paginado (`ORDER BY created_at DESC, id DESC`) |
| `ListUpcomingActivities` | Público: `published` **y** fecha próxima; `ORDER BY activity_date, start_time NULLS FIRST, title_es, id` |
| `ListNextUpcomingItems` | Extracto de la portada (`GET /api/v1/portada/proximamente`, enmienda Q8 del 2026-10-12): `UNION ALL` de eventos (con su primera fecha próxima) y actividades, solo `published` **y** próximas (misma regla R4-2 con `today`/`now_time` parametrizados), `ORDER BY date, start_time NULLS FIRST, title_es, id` y `LIMIT 3` (`maxPortadaItems`, tope firme) |

### `backend/internal/db/queries/media.sql` (dominio `media`, nuevo — R4-6)

| Consulta | Uso |
|---|---|
| `IsFilePublished` | `TRUE` solo si `:file_name` está referenciado por contenido **publicado y visible hoy**: `home_identity` `published` (logo/imagen de portada) **o** `agenda_events`/`agenda_activities` `published` con fecha próxima (misma regla R4-2 con `:today`/`:now_time`). Sustituye a `IsHomeFilePublished` (F3), que se retira de `home.sql` al mover la descarga a `internal/media/` |

La auditoría **reutiliza** `InsertAdminAction` (F2): el repository de `agenda` la llama con
`gendb.New(tx)` dentro de la transacción de cada mutación. No se escribe ningún `UPDATE`/`DELETE`
sobre `admin_actions` (solo lectura, FR-025 de F2).

**Separación público/panel en consultas distintas** (`…Upcoming`/`…Published` frente a las del
panel): el filtro público no puede olvidarse en una llamada futura — la única función que ve
borradores o vencidos es la del panel, protegida con su permiso (FR-013/SC-002/SC-009).

## Escritura de la auditoría desde F4 (detalle)

- Mutación y fila de `admin_actions` comparten transacción (o ambos o ninguno). Si el INSERT del
  registro falla, **no se aplica** la operación.
- Objetivo: `target_kind='content'`, `target_user_id`/`target_role_id` = NULL, `target_label` =
  `Eventos · <título>` / `Actividades · <título>` / `Categorías · <nombre>` /
  `Eventos · Imagen · <fileName>` / `Actividades · Imagen · <fileName>` (máx. 254, `CHECK` de F2).
- `action` ∈ los 15 códigos nuevos (regla exacta en research R4-10: `create` en el alta aunque nazca
  publicada; `*.publish`/`*.unpublish` **solo** para cambios de estado; `*.image.upload` en la
  subida); `result`: `success` en la transacción, `failure` en rechazos y `denied` en denegaciones
  (best-effort).
- `actor_user_id` siempre la cuenta del panel que ejecutó (el `CHECK` de F2 sigue vigente).

## Invariantes de negocio y dónde se verifican

| Invariante | Dónde |
|---|---|
| Español obligatorio, inglés opcional | Columnas `*_es NOT NULL` / `*_en NULL` + `required`/`omitempty` en los DTOs (`platform/validate`) |
| `""`/espacios en `*_en` → `NULL` (I6 de F3) | Normalización de trim en el service antes de persistir |
| Detalle ≤ 2.000 caracteres por idioma (FR-018/FR-019) | `CHECK` de `agenda_events`/`agenda_activities` + `max=2000` en el DTO |
| Cada fecha = un día concreto + horas opcionales coherentes (Q2) | `CHECK` de patrón `HH:MM` + `CHECK (end_time IS NULL OR start_time IS NOT NULL)` + `CHECK (end_time IS NULL OR end_time > start_time)`; formato de fecha validado en el service (fecha de calendario real) |
| ≥ 1 fecha por evento y ≤ 50 (FR-015) | Service (constante revisable); `UNIQUE (event_id, event_date)` rechaza repetidas |
| Imagen opcional con alt obligatorio si hay imagen (FR-020) | `CHECK (image_file IS NULL OR image_alt_es IS NOT NULL)` + validación condicional en el DTO |
| Categoría opcional válida y borrado protegido (Q6/R4-5) | FK `ON DELETE RESTRICT` + `409` en el service + validación de `categoryId` (desconocido → `400 invalid`) |
| Solo lo publicado **y próximo** es visible al visitante (FR-002/FR-006/FR-013) | Consultas `…Upcoming` (única vía pública) + `IsFilePublished` para imágenes + pruebas "0 borradores/vencidos expuestos" |
| Retiro automático sin cambiar estado (Q4) | La regla es de lectura (R4-2): `publication_state` intacto; el panel marca `isPast`/`nextDate: null` |
| Duplicado exacto dentro del módulo rechazado (R4-9) | `UNIQUE (title_es, activity_date, place_es)` en actividades + `ExistsEventDuplicate` en eventos + `UNIQUE (name_es)` en categorías → `409 conflict` |
| Auditoría atómica con la mutación | Transacción única (`InsertAdminAction` sobre `tx`) + prueba de integración |
| Registro de auditoría de solo lectura | Sin consultas `UPDATE`/`DELETE` sobre `admin_actions` en `queries/` |
| Catálogo de permisos de F2: `categorias` presente (10.º) y solo crece por migración | `INSERT` en `000009` (nunca se edita `000002`); las pruebas de F2 verifican 10 permisos (excepción aprobada, RG4-10) y el init asigna todo el catálogo al administrador inicial |
| "Hoy" y "ahora" del filtro público del lugar de la iglesia (R4-2) | Service calcula `today`/`now_time` con `APP_TZ=America/Costa_Rica` y los pasa **parametrizados**; pruebas de integración con fechas fijas |

## Validación del modelo (qué comprobará la implementación)

1. `migrate up` limpio desde F3 (`000006`) y `down` completo de `000007`/`000008`/`000009` sin
   errores (incluida la restauración del `CHECK` de `admin_actions` y del permiso `categorias`:
   up → down → up, evidencia de que los nombres de restricciones y filas referenciadas existen).
2. Pruebas de integración del repository (PostgreSQL real, `//go:build integration`):
   - el catálogo de `permissions` contiene **10** filas y `categorias` entre ellas (y el `down` de
     `000009` la retira junto a sus asignaciones de `role_permissions`);
   - `UNIQUE (event_id, event_date)`: segunda fecha igual en el mismo evento → restricción;
   - `CHECK` de `end_time` (exige `start_time` y es posterior), patrón `HH:MM`, longitudes y
     `publication_state`;
   - `UNIQUE (title_es, activity_date, place_es)` → actividad duplicada rechazada; y el evento
     duplicado lo resuelve `ExistsEventDuplicate` (sin índice único);
   - `FK RESTRICT` de `category_id`: borrar una categoría en uso falla y el service responde `409`;
     con `category_id IS NULL` el borrado procede;
   - `ReplaceEventDates` deja exactamente la lista pedida y borra las anteriores (reemplazo
     completo, R4-11);
   - lecturas `…Upcoming`: **nunca** devuelven `draft` ni elementos cuya última fecha ya pasó (con
     `today`/`now_time` de prueba se cubren los tres casos de la regla R4-2: futuro, hoy sin fin,
     hoy con fin ya terminado); un evento solo aparece si tiene ≥ 1 fecha próxima y su `dates`
     público no contiene pasadas;
   - `IsFilePublished`: `TRUE` solo con el elemento publicado **y próximo** que referencia el
     archivo; al pasar la fecha (o retirar a borrador) pasa a `FALSE` sin borrar el archivo;
   - cada mutación deja su fila en `admin_actions` en la misma transacción (y al forzar un fallo en
     el registro, la mutación no se aplica), con códigos y `target_label` correctos;
   - inserción de `admin_actions` con los 15 códigos nuevos dentro del `CHECK` ampliado (y rechazo
     de un código fuera del registro).
3. Tipos sqlc: `pgtype.*` traducido en `repository.go` (§8.1.6); `internal/db` regenerado y
   commiteado (`make sqlc-verify`).
