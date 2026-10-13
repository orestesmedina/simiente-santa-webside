# Data Model — F6 Ministerios

**Fecha**: 2026-10-13 · **Spec**: [spec.md](./spec.md) · **Decisiones**: [research.md](./research.md)
(R6-1…R6-12) · **Plantilla de estilo**: el `data-model.md` de F4.

Fuente de verdad del esquema: **este documento** → migraciones `backend/migrations/` → consultas
`backend/internal/db/queries/ministries.sql` y `media.sql` (sqlc). Convenciones: skill `postgres-db`,
§VI de la constitución y §8.1 de `docs/tecnico/arquitectura.md` (UUID, `created_at`/`updated_at`,
mapeo `pgtype` → dominio en el repository, consultas parametrizadas).

## Patrones transversales heredados de F3/F4 (no se reabren)

- **Contenido bilingüe (R3-1)**: cada campo traducible vive en dos columnas, `*_es` (idioma base)
  y `*_en` (**opcional**). `""`/espacios se guardan como `NULL` (normalización I6: trim en el
  service antes de persistir). **La regla es por campo** (FR-007/Q6): si un campo **opcional** se
  rellena en algún idioma, su versión en español debe existir (`*_en IS NULL OR *_es IS NOT NULL`,
  con `CHECK` en cada par). No traducibles: `whatsapp_url`, `image_file`, `leader_phone`,
  `leader_email`, `leader_visible`.
- **Estado de publicación (R3-3)**: `publication_state TEXT NOT NULL DEFAULT 'draft' CHECK
  (publication_state IN ('draft','published'))` **por ministerio**. Las consultas del visitante
  filtran `publication_state = 'published'` en la **misma** consulta (`…Published`): es la única
  vía pública y no existe ninguna función que devuelva al visitante un borrador (FR-012/SC-002).
- **Orden manual (Q5/FR-004)**: `sort_order INTEGER NOT NULL DEFAULT 0` con índice
  `(sort_order, id)` y `ORDER BY sort_order, id` — literal del precedente de
  `000005_create_home_content` (`home_services`, `home_whatsapp_channels`) y
  `000007_create_agenda_content` (`agenda_categories`). El desempate es `id`, **no** el nombre.
- **Auditoría (R3-11/R4-10)**: las mutaciones insertan `admin_actions` con `InsertAdminAction`
  (consulta compartida de F2) **dentro de su transacción**; objetivo `target_kind='content'`, ambas
  FK en `NULL` y `target_label` descriptivo. Registro de solo lectura (sin `UPDATE`/`DELETE`).
- **Privacidad del encargado (FR-026, R6-2)**: las consultas **públicas** devuelven los campos del
  encargado con `CASE WHEN leader_visible THEN … END` (con el interruptor apagado la fila pública
  sale con `NULL`: ni nombre ni contacto). Las consultas del panel sí seleccionan todo.

## Resumen de tablas

| Tabla | Cardinalidad | Qué guarda |
|---|---|---|
| `ministries` (**nueva**) | 0..N | Un ministerio (área de servicio): nombre y descripción (obligatorios en es), encargado **opcional** (nombre + un medio de contacto opcional + `leader_visible`), enlace opcional al grupo de WhatsApp, horario y lugar opcionales como **textos libres**, imagen opcional + alt, `sort_order` y `publication_state` (FR-001, FR-019, Q1/Q2/Q4/Q5/Q6/Q8/Q9) |
| `admin_actions` (**existente, F2**) | — | Se **amplía** el registro cerrado: 8 códigos `ministry.*` (FR-017, R6-12) |
| `permissions` (**existente, F2**) | — | **SIN CAMBIOS**: `('ministerios', 'Ministerios')` ya está sembrado por `000002`; F6 **no** crea permisos nuevos (FR-011; a diferencia de F4, que sí necesitó `000009`) |
| `roles` / `role_permissions` (**existentes, F2**) | — | **SIN CAMBIOS** (Decisión 9: el modelo de cuentas y roles es F2, intacto) |

Ninguna tabla de F6 tiene FK hacia `users`: el "quién" vive en `admin_actions.actor_user_id`
(FR-017). **Ninguna FK nueva**: el encargado es un campo del ministerio (R6-1: "sin ficha
compartida") y F6 no tiene catálogo. Índices: `(sort_order, id)` para el orden público y de panel,
y `publication_state` para el filtro/lectura pública.

---

## Migraciones

Dos migraciones nuevas (F4 terminó en `000009`): **`000010`** crea la tabla `ministries`;
**`000011`** amplía el registro de auditoría con los 8 códigos `ministry.*`. **No hay migración de
permisos** (el permiso `ministerios` ya existe) y ninguna migración existente se toca (§VI).

### `000010_create_ministries.up.sql`

```sql
-- 000010_create_ministries (up)
-- F6 Ministerios: una sola tabla con el encargado como campos del ministerio
-- (research R6-1), contenido bilingüe es/en por campo (R6-9), estado de
-- publicación por elemento (R3-3) y orden manual sort_order (Q5, patrón de
-- 000005/000007). Convenciones: skill `postgres-db` y
-- specs/006-ministerios/data-model.md (fuente de verdad).

CREATE TABLE ministries (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name_es           TEXT NOT NULL CHECK (char_length(name_es) BETWEEN 1 AND 160),
    name_en           TEXT NULL CHECK (char_length(name_en) BETWEEN 1 AND 160),
    description_es    TEXT NOT NULL CHECK (char_length(description_es) BETWEEN 1 AND 2000),
    description_en    TEXT NULL CHECK (char_length(description_en) BETWEEN 1 AND 2000),

    -- Encargado OPCIONAL (Q1/Q8/FR-025): nombre + un solo medio de contacto
    -- opcional (celular O correo). leader_visible gobierna solo la PUBLICACIÓN
    -- de la ficha (nombre y contacto juntos, Q9/FR-026): el dato guardado se
    -- conserva en la fila mientras el encargado exista.
    -- DEFAULT FALSE = "Oculto por defecto" (decisión del humano 2026-10-13,
    -- puerta 2, research R6-3): la ficha del encargado —nombre y contacto— NO
    -- se publica hasta que el equipo encienda el interruptor (privacidad por
    -- defecto: publicar un celular/correo personal exige un acto afirmativo).
    -- El interruptor es propio de F6 (F3/F4 no lo tienen).
    leader_name_es    TEXT NULL CHECK (char_length(leader_name_es) BETWEEN 1 AND 160),
    leader_name_en    TEXT NULL CHECK (char_length(leader_name_en) BETWEEN 1 AND 160),
    leader_phone      TEXT NULL CHECK (char_length(leader_phone) BETWEEN 7 AND 32),
    leader_email      TEXT NULL CHECK (char_length(leader_email) BETWEEN 3 AND 254),
    leader_visible    BOOLEAN NOT NULL DEFAULT FALSE,

    -- Enlace OPCIONAL al grupo de WhatsApp (Q2): cualquier https válido y
    -- seguro, sin restricción de dominios (Q3/FR-015). No traducible (FR-008).
    whatsapp_url      TEXT NULL CHECK (char_length(whatsapp_url) BETWEEN 1 AND 500),

    -- Horario y lugar: textos libres OPCIONALES (Q6/FR-024). Sin fechas, sin
    -- recurrencia calculable y sin calendario: el modelo de F4 no aplica aquí.
    schedule_es       TEXT NULL CHECK (char_length(schedule_es) BETWEEN 1 AND 200),
    schedule_en       TEXT NULL CHECK (char_length(schedule_en) BETWEEN 1 AND 200),
    place_es          TEXT NULL CHECK (char_length(place_es) BETWEEN 1 AND 200),
    place_en          TEXT NULL CHECK (char_length(place_en) BETWEEN 1 AND 200),

    -- Imagen OPCIONAL (Q4/FR-020) con texto alternativo obligatorio en español
    -- si hay imagen (misma regla que F3/F4).
    image_file        TEXT NULL CHECK (char_length(image_file) BETWEEN 1 AND 120),
    image_alt_es      TEXT NULL CHECK (char_length(image_alt_es) BETWEEN 1 AND 300),
    image_alt_en      TEXT NULL CHECK (char_length(image_alt_en) BETWEEN 1 AND 300),

    publication_state TEXT NOT NULL DEFAULT 'draft'
                      CHECK (publication_state IN ('draft', 'published')),
    sort_order        INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Celular o correo, nunca ambos (Q8).
    CHECK (leader_phone IS NULL OR leader_email IS NULL),
    -- El contacto exige encargado con nombre (el encargado se compone de nombre;
    -- Q8 solo permite "solo nombre", nunca "solo contacto").
    CHECK (leader_name_es IS NOT NULL OR (leader_phone IS NULL AND leader_email IS NULL)),
    -- Español base POR CAMPO: si existe la versión en inglés, existe la española
    -- (FR-007/Q6; "un campo vacío no obliga a nada").
    CHECK (name_en IS NULL OR name_es IS NOT NULL),
    CHECK (description_en IS NULL OR description_es IS NOT NULL),
    CHECK (leader_name_en IS NULL OR leader_name_es IS NOT NULL),
    CHECK (schedule_en IS NULL OR schedule_es IS NOT NULL),
    CHECK (place_en IS NULL OR place_es IS NOT NULL),
    CHECK (image_alt_en IS NULL OR image_alt_es IS NOT NULL),
    -- Toda imagen lleva texto alternativo en español (FR-020).
    CHECK (image_file IS NULL OR image_alt_es IS NOT NULL),
    -- Enlace: solo https sin espacios (FR-015; la validación completa —host,
    -- caracteres de control, esquemas no https— vive en el service, R6-5).
    CHECK (whatsapp_url IS NULL OR whatsapp_url ~ '^https://[^[:space:]]+$'),
    CHECK (name_es = btrim(name_es)),
    CHECK (leader_name_es = btrim(leader_name_es)),
    CHECK (schedule_es = btrim(schedule_es)),
    CHECK (place_es = btrim(place_es)),
    CHECK (whatsapp_url IS NULL OR whatsapp_url = btrim(whatsapp_url))
);

-- Orden de presentación (Q5/FR-004): sort_order, id, como home_services y
-- agenda_categories.
CREATE INDEX ministries_sort_order_id_idx ON ministries (sort_order, id);

-- WHERE publication_state en la lectura pública y en el filtro del panel.
CREATE INDEX ministries_publication_state_idx ON ministries (publication_state);
```

### `000010_create_ministries.down.sql`

```sql
-- 000010_create_ministries (down) — revierte 000010 por completo (§VI).
DROP TABLE IF EXISTS ministries;
```

### `000011_extend_admin_actions_for_ministries.up.sql`

Amplía el registro de auditoría de F2 (FR-017/R6-12). El `CHECK` que se sustituye es
`admin_actions_action_check`, **con nombre propio** desde `000006` de F3 (F3 y F4 ya hicieron este
mismo ejercicio y su prueba de integración lo verifica con up → down → up). El `target_kind` **no
cambia**: los ministerios son `content` como los elementos de F3/F4. Los **8 códigos** nuevos son
los de R6-12 (incluidos los dos de la visibilidad de la ficha del encargado).

```sql
-- 000011_extend_admin_actions_for_ministries (up)
-- F6: la auditoría de F2 se amplía a ministerios (FR-017). El registro de
-- códigos de acción sigue siendo CERRADO y lo garantiza la BD.

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
    'category.create', 'category.update', 'category.delete',
    'ministry.create', 'ministry.update', 'ministry.delete', 'ministry.publish',
    'ministry.unpublish', 'ministry.leader.show', 'ministry.leader.hide', 'ministry.image.upload'
));
```

### `000011_extend_admin_actions_for_ministries.down.sql`

```sql
-- 000011_extend_admin_actions_for_ministries (down): restaura el CHECK de F4
-- (8 códigos de F2 + 15 de F3 + 15 de F4). Como en el down de 000006/000008, es
-- destructivo a propósito: las filas con códigos F6 no caben en el CHECK
-- anterior y se eliminan antes de restaurar. Solo tiene sentido en
-- desarrollo/rollback.
DELETE FROM admin_actions WHERE action IN (
    'ministry.create', 'ministry.update', 'ministry.delete', 'ministry.publish',
    'ministry.unpublish', 'ministry.leader.show', 'ministry.leader.hide', 'ministry.image.upload'
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
    'home.publish', 'home.unpublish',
    'event.create', 'event.update', 'event.delete', 'event.publish', 'event.unpublish',
    'event.image.upload',
    'activity.create', 'activity.update', 'activity.delete', 'activity.publish',
    'activity.unpublish', 'activity.image.upload',
    'category.create', 'category.update', 'category.delete'
));
```

**Notas de diseño.**

- **`leader_visible` y la privacidad**: la columna guarda la decisión del equipo; la **no
  publicación** se garantiza en la lectura (consultas públicas con `CASE WHEN leader_visible`,
  R6-2), no escondiendo la columna. Con el interruptor apagado, el panel sigue viendo
  `leader_name_*` y `leader_phone`/`leader_email` (dato guardado ≠ dato publicado). **Nace
  apagado** (`DEFAULT FALSE`, «Oculto por defecto», decisión del humano 2026-10-13): un ministerio
  nuevo **no publica** la ficha del encargado hasta que el equipo la enciende; el `DEFAULT` cubre
  también los `INSERT` que omitan la columna. El interruptor es **propio de F6** (F3/F4 no lo
  tienen).
- **`leader_phone`/`leader_email`**: dos columnas anulables mutuamente excluyentes (CHECK) en vez
  de un par `kind`/`value`, para validar con los tags `phone`/`email` ya existentes (R6-4).
- **`whatsapp_url`**: sin `UNIQUE` — dos ministerios **pueden** compartir grupo (Q: "se señala
  como aviso pero se permite", R6-11); el duplicado exacto lo rechaza el service (`409`) con
  `ExistsMinistryDuplicate`.
- **`sort_order`** en la misma tabla (no hay lista de orden aparte): el panel lo edita como un
  campo más del `PATCH` y el visitante ve exactamente ese orden (US2 esc. 9).
- **Sin `deleted_at`**: la eliminación es definitiva, con confirmación y registrada (default
  revisable de la spec, igual que F3/F4).

---

## Consultas sqlc

### `backend/internal/db/queries/ministries.sql` (dominio `ministerios`, nuevo)

Listas con `ORDER BY` explícito y determinista (§8.1.7). Sin `SELECT *`. Todo parametrizado
(§IV, CWE-89). **Separación público/panel en consultas distintas**: la única función que ve
borradores y datos ocultos del encargado es la del panel (FR-012/FR-026).

| Consulta | Uso |
|---|---|
| `ListPublishedMinistries` (`:limit`) | **Público** (sección y bloque de la portada): solo `publication_state = 'published'`, `ORDER BY sort_order, id`, `LIMIT :limit`. Los campos del encargado salen con `CASE WHEN leader_visible THEN leader_name_es END AS leader_name_es` (ídem `leader_phone`, `leader_email`, `leader_name_en`): con el interruptor apagado la fila **no transporta** el dato (R6-2). El límite lo pasa el service (`maxPublicItems = 200` en la sección, `maxPortadaItems = 5` en el bloque): el tope 5 del bloque es de servidor (FR-027) |
| `ListMinistries` (`:limit`, `:offset`, `sqlc.narg('publication_state')`) | Panel paginado: **todo** (incluidos borradores y fichas ocultas), `ORDER BY sort_order, id` (el mismo orden que ve el visitante, US2 esc. 9), filtro opcional por estado |
| `CountMinistries` (`sqlc.narg('publication_state')`) | Total del listado del panel (sobre de paginación) |
| `GetMinistryByID` | Panel (detalle y respuestas de escritura): todas las columnas + `whatsapp_url_shared BOOLEAN` calculado con `EXISTS (SELECT 1 FROM ministries m2 WHERE m2.whatsapp_url = ministries.whatsapp_url AND m2.id <> ministries.id AND ministries.whatsapp_url IS NOT NULL)` — el "aviso" de enlace compartido (R6-11) |
| `InsertMinistry` / `UpdateMinistry` / `DeleteMinistry` | Mutaciones del panel (el `Update` escribe el DTO completo; los opcionales van como `NULL` cuando se borran) |
| `ExistsMinistryDuplicate` | Duplicado exacto (nombre + descripción + encargado + enlace, **normalizados**: `btrim` + minúsculas, con `IS NOT DISTINCT FROM` en los opcionales) → `409 conflict` (R6-11) |

### `backend/internal/db/queries/media.sql` (dominio `media`, **editado** — R6-7)

| Consulta | Uso |
|---|---|
| `IsFilePublished` | **Crece** con una tercera rama: `TRUE` también si `:file_name` está referenciado por `ministries.image_file` con `publication_state = 'published'`. Misma regla que F3/F4: un borrador **no** sirve su imagen y la URL vuelve a `200` al publicar (FR-012 "por ninguna vía"). El resto de la consulta (portada F3, agenda F4) no cambia |

La auditoría **reutiliza** `InsertAdminAction` (F2): el repository de `ministerios` la llama con
`gendb.New(tx)` dentro de la transacción de cada mutación. No se escribe ningún `UPDATE`/`DELETE`
sobre `admin_actions` (solo lectura, FR-025 de F2).

## Escritura de la auditoría desde F6 (detalle)

- Mutación y fila de `admin_actions` comparten transacción (o ambos o ninguno). Si el INSERT del
  registro falla, **no se aplica** la operación (edge case "no aplicarse sin registro").
- Objetivo: `target_kind='content'`, `target_user_id`/`target_role_id` = NULL,
  `target_label` = `Ministerios · <nameEs>` / `Ministerios · Imagen · <fileName>` (máx. 254,
  `CHECK` de F2).
- `action` ∈ los 8 códigos nuevos (regla exacta en research R6-12): `ministry.create` en el alta
  aunque nazca publicado; `ministry.publish`/`ministry.unpublish` **solo** para cambios de estado;
  `ministry.leader.show`/`ministry.leader.hide` **solo** para cambios del interruptor;
  `ministry.update` para cualquier otro cambio de datos (incluido `sortOrder`); datos + estado /
  visibilidad en el mismo guardado → **dos filas**; `ministry.image.upload` en la subida.
  `result`: `success` en la transacción, `failure` en rechazos y `denied` en denegaciones
  (best-effort).
- `actor_user_id` siempre la cuenta del panel que ejecutó (el `CHECK` de F2 sigue vigente).

## Invariantes de negocio y dónde se verifican

| Invariante | Dónde |
|---|---|
| Solo `nameEs` y `descriptionEs` son obligatorios (FR-014) | `required` en los DTOs (`platform/validate`) + `NOT NULL` en la tabla |
| Español por campo: `*_en` exige su `*_es` (FR-007/Q6) | `CHECK` por par de columnas + regla del service (el formulario la refleja) |
| Encargado opcional; contacto opcional **aunque haya encargado** (Q8) | Columnas anulables; sin `CHECK` que exija contacto |
| Celular **o** correo, nunca ambos (Q8) | `CHECK (leader_phone IS NULL OR leader_email IS NULL)` + `400 invalid` con `details` en el service |
| El contacto exige nombre del encargado | `CHECK (leader_name_es IS NOT NULL OR (leader_phone IS NULL AND leader_email IS NULL))` |
| **FR-026: con `leader_visible = FALSE` ningún dato del encargado sale al público** | Consultas públicas con `CASE WHEN leader_visible` (la fila sale con `NULL`) + DTO público sin campos del encargado oculto + bloque sin campos de encargado + pruebas "0 apariciones" (service/handler/repository/e2e) |
| **«Oculto por defecto»: un ministerio nuevo nace con la ficha del encargado NO publicada** (decisión del humano 2026-10-13, puerta 2) | `leader_visible BOOLEAN NOT NULL DEFAULT FALSE` en `000010` (también para `INSERT` que omitan la columna) + `leaderVisible` con `default: false` en `MinisterioInput` (el service aplica `false` si falta) + interruptor apagado al crear + pruebas del default en las cuatro capas |
| Enlace solo `https` seguro (FR-015/Q3) | Tag `url` + regla de dominio (sin espacios ni control) en el service + `CHECK '^https://[^[:space:]]+$'` + prueba de `javascript:`/`data:`/`vbscript:` |
| El enlace se sirve tal cual (sin transformar el destino) | Sin normalización del valor en el service; prueba de igualdad `href` = valor guardado |
| Horario y lugar son **textos**: sin fechas ni recurrencia (Q6/FR-024) | No hay columnas de fecha ni tablas hijas: solo `TEXT`; ninguna consulta calcula fechas |
| Imagen opcional con alt obligatorio si hay imagen (FR-020) | `CHECK (image_file IS NULL OR image_alt_es IS NOT NULL)` + validación condicional en el DTO |
| Solo lo publicado es visible al visitante (FR-012) | Consultas `…Published` (única vía pública) + `IsFilePublished` para imágenes + pruebas "0 borradores expuestos" |
| Orden manual `sort_order, id` en sección y bloque (FR-004/FR-016/FR-027) | `ORDER BY sort_order, id` en `ListPublishedMinistries` + índice `ministries_sort_order_id_idx` + prueba de orden |
| Tope **5** del bloque, resuelto en el servidor (FR-027) | `maxPortadaItems = 5` (constante del service, con prueba) + `LIMIT :limit` en SQL; sin parámetro de límite en el contrato |
| Duplicado exacto rechazado (R6-11) | `ExistsMinistryDuplicate` (normalizado) → `409 conflict` |
| Enlace compartido permitido y señalado (R6-11) | `whatsapp_url_shared` (calculado en `GetMinistryByID`/`ListMinistries`) → `Notice` no bloqueante en el panel |
| Auditoría atómica con la mutación | Transacción única (`InsertAdminAction` sobre `tx`) + prueba de integración |
| Registro de auditoría de solo lectura | Sin consultas `UPDATE`/`DELETE` sobre `admin_actions` en `queries/` |
| Permisos: `ministerios` ya existe y **no cambia** el catálogo de F2 | Sin migración de permisos; `000002` no se toca; `AdminChain("ministerios")` en `routes.go` |

## Validación del modelo (qué comprobará la implementación)

1. `migrate up` limpio desde F4 (`000009`) y `down` completo de `000010`/`000011` sin errores
   (incluida la restauración del `CHECK` de `admin_actions`: up → down → up).
2. Pruebas de integración del repository (PostgreSQL real, `//go:build integration`):
   - el catálogo de `permissions` **sigue en 10 filas** y `ministerios` entre ellas (F6 no lo
     toca; el `down` de `000010`/`000011` no lo roza);
   - `CHECK` de exclusión: una fila con `leader_phone` **y** `leader_email` falla; con solo uno
     procede;
   - `CHECK` "contacto exige nombre": `leader_phone` sin `leader_name_es` falla;
   - `CHECK` de español por campo: `schedule_en` sin `schedule_es` falla (ídem `place_en`,
     `leader_name_en`, `image_alt_en`);
   - `CHECK` de imagen: `image_file` sin `image_alt_es` falla;
   - `CHECK` del enlace: `http://…`, `javascript:alert(1)` y un valor con espacio fallan;
     `https://chat.whatsapp.com/…` y `https://ejemplo.com/grupo` (dominio no de WhatsApp) pasan;
   - **«Oculto por defecto»** (decisión del humano 2026-10-13): un `INSERT` que omite
     `leader_visible` guarda `FALSE` (prueba del `DEFAULT` de `000010`); una fila con nombre y
     contacto de encargado y el interruptor en su default sale **sin** datos del encargado en la
     lectura pública y el service devuelve `leaderVisible: false`;
   - `ListPublishedMinistries`: **nunca** devuelve borradores; con `leader_visible = FALSE` las
     columnas del encargado salen **`NULL`** aunque la fila tenga datos (R6-2, la prueba más
     importante del módulo); el orden es `sort_order, id` y el `LIMIT` se respeta (5 en el bloque);
   - `GetMinistryByID` sí devuelve los datos del encargado con el interruptor apagado (panel) y
     `whatsapp_url_shared` es `TRUE` solo cuando otra fila comparte el enlace;
   - `ExistsMinistryDuplicate` detecta el duplicado exacto normalizado ("Alabanza" / " alabanza ")
     y **no** dispara por un enlace compartido;
   - `IsFilePublished`: `TRUE` con el ministerio publicado que referencia el archivo; `FALSE` con
     el mismo ministerio en borrador (y vuelve a `TRUE` al publicarlo), sin borrar el archivo;
   - cada mutación deja su fila en `admin_actions` en la misma transacción (y al forzar un fallo en
     el registro, la mutación no se aplica), con los 8 códigos y `target_label` correctos;
   - inserción de `admin_actions` con los 8 códigos dentro del `CHECK` ampliado (y rechazo de un
     código fuera del registro).
3. Tipos sqlc: `pgtype.*` traducido en `repository.go` (§8.1.6); `internal/db` regenerado y
   commiteado (`make sqlc-verify`).
