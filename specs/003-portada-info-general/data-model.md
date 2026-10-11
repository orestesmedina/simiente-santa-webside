# Data Model — F3 Portada e información general

**Fecha**: 2026-10-09 · **Spec**: [spec.md](./spec.md) · **Decisiones**: [research.md](./research.md)
(R3-1, R3-3, R3-4, R3-5, R3-6, R3-7, R3-11) · **Plantilla de estilo**: el `data-model.md` de F2.

Fuente de verdad del esquema: **este documento** → migraciones `backend/migrations/` → consultas
`backend/internal/db/queries/home.sql` (sqlc). Convenciones: skill `postgres-db`, §VI de la
constitución y §8.1 de `docs/tecnico/arquitectura.md` (UUID, `created_at`/`updated_at`, mapeo
`pgtype` → dominio en el repository).

## Patrón transversal: contenido bilingüe (R3-1)

Todo campo traducible vive en **dos columnas**: `*_es` (**obligatorio**, idioma base) y `*_en`
(**opcional**). Los campos no traducibles (URLs, teléfonos, correo, destino de WhatsApp, día, horas,
archivo de imagen) no llevan sufijo. La regla "es obligatorio, inglés opcional" (Decisión 8) está en
el esquema: `*_es NOT NULL` + `*_en NULL`. **Normalización** *(analyze I6)*: `""` o solo espacios en
un `*_en` se guarda como **`NULL`** (trim en el service; los `CHECK` de longitud `BETWEEN 1 AND …`
rechazan `''`, que nunca debe llegar a la BD). El fallback `en → es` lo resuelve el service público
(R3-2); la BD guarda ambos valores crudos.

**Estado de publicación** (R3-3): toda fila de contenido lleva
`publication_state TEXT NOT NULL DEFAULT 'draft' CHECK (publication_state IN ('draft','published'))`.
La publicación es **por elemento**, también en los *singletons*: identidad, «quiénes somos» y contacto
son cada uno un elemento publicable **por separado** con su propio `publication_state` (una sección =
un elemento; no hay estado agrupado de sección ni de página). Las consultas del visitante filtran
`publication_state = 'published'`; una sección sin filas
publicadas se omite por completo en la respuesta pública (FR-013/Q10).

## Resumen de tablas

| Tabla | Cardinalidad | Qué guarda |
|---|---|---|
| `home_identity` | **singleton** (≤1 fila) | Nombre oficial, lema, misión, visión, logo, imagen de portada y sus textos alternativos (R3-4) |
| `home_about` | **singleton** (≤1 fila) | Texto plano de «quiénes somos» es/en (≤1.000 caracteres, FR-003) |
| `home_contact` | **singleton** (≤1 fila) | Dirección es/en, correo y teléfono (FR-007) |
| `home_services` | 0..N (≤50) | Servicios del horario: día, hora, nombre/descripción, lugar (FR-004, R3-5) |
| `home_whatsapp_channels` | 0..N (≤20) | Canales de WhatsApp: nombre/propósito, tipo y destino único (FR-005, R3-6) |
| `home_social_links` | 0..1 por red del catálogo | Un enlace por red del catálogo fijo (FR-006, R3-7) |
| `admin_actions` (**existente, F2**) | — | Se **amplía** el registro cerrado: 15 códigos `home.*` (incl. `home.image.upload`) y `target_kind='content'` (FR-017, R3-11) |

Ninguna tabla de F3 tiene FK hacia `users`: el "quién" de la edición vive en `admin_actions.actor_user_id`
(FR-017). Ningún contenido referencia a otro contenido.

---

## Migraciones

Dos migraciones nuevas (F2 terminó en `000004`): **`000005`** crea las tablas de contenido y
**`000006`** amplía el registro de auditoría de F2. Ninguna migración existente se toca (§VI).

### `000005_create_home_content.up.sql`

```sql
-- 000005_create_home_content (up)
-- F3 Portada e información general: contenido de la portada con versiones
-- es (obligatoria) / en (opcional) y estado de publicación por elemento
-- (research R3-1/R3-3). Convenciones: skill `postgres-db` y
-- specs/003-portada-info-general/data-model.md (fuente de verdad).

-- Patrón singleton (R3-4): `singleton` siempre TRUE + UNIQUE garantiza ≤1 fila
-- sin debilitar los NOT NULL/CHECK de los campos de negocio.

CREATE TABLE home_identity (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    singleton          BOOLEAN NOT NULL DEFAULT TRUE CHECK (singleton),
    name_es            TEXT NOT NULL CHECK (char_length(name_es) BETWEEN 1 AND 160),
    name_en            TEXT NULL CHECK (char_length(name_en) BETWEEN 1 AND 160),
    tagline_es         TEXT NULL CHECK (char_length(tagline_es) BETWEEN 1 AND 300),
    tagline_en         TEXT NULL CHECK (char_length(tagline_en) BETWEEN 1 AND 300),
    mission_es         TEXT NULL CHECK (char_length(mission_es) BETWEEN 1 AND 600),
    mission_en         TEXT NULL CHECK (char_length(mission_en) BETWEEN 1 AND 600),
    vision_es          TEXT NULL CHECK (char_length(vision_es) BETWEEN 1 AND 600),
    vision_en          TEXT NULL CHECK (char_length(vision_en) BETWEEN 1 AND 600),
    logo_file          TEXT NULL CHECK (char_length(logo_file) BETWEEN 1 AND 120),
    logo_alt_es        TEXT NULL CHECK (char_length(logo_alt_es) BETWEEN 1 AND 300),
    logo_alt_en        TEXT NULL CHECK (char_length(logo_alt_en) BETWEEN 1 AND 300),
    cover_image_file   TEXT NULL CHECK (char_length(cover_image_file) BETWEEN 1 AND 120),
    cover_image_alt_es TEXT NULL CHECK (char_length(cover_image_alt_es) BETWEEN 1 AND 300),
    cover_image_alt_en TEXT NULL CHECK (char_length(cover_image_alt_en) BETWEEN 1 AND 300),
    publication_state  TEXT NOT NULL DEFAULT 'draft'
                       CHECK (publication_state IN ('draft', 'published')),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (singleton),
    -- FR-019: toda imagen lleva texto alternativo en español (idioma base).
    CHECK (logo_file IS NULL OR logo_alt_es IS NOT NULL),
    CHECK (cover_image_file IS NULL OR cover_image_alt_es IS NOT NULL)
);

CREATE TABLE home_about (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    singleton         BOOLEAN NOT NULL DEFAULT TRUE CHECK (singleton),
    text_es           TEXT NOT NULL CHECK (char_length(text_es) BETWEEN 1 AND 1000),
    text_en           TEXT NULL CHECK (char_length(text_en) BETWEEN 1 AND 1000),
    publication_state TEXT NOT NULL DEFAULT 'draft'
                      CHECK (publication_state IN ('draft', 'published')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (singleton)
);

CREATE TABLE home_contact (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    singleton         BOOLEAN NOT NULL DEFAULT TRUE CHECK (singleton),
    address_es        TEXT NOT NULL CHECK (char_length(address_es) BETWEEN 1 AND 300),
    address_en        TEXT NULL CHECK (char_length(address_en) BETWEEN 1 AND 300),
    email             TEXT NOT NULL CHECK (char_length(email) BETWEEN 3 AND 254),
    phone             TEXT NOT NULL CHECK (char_length(phone) BETWEEN 7 AND 32),
    publication_state TEXT NOT NULL DEFAULT 'draft'
                      CHECK (publication_state IN ('draft', 'published')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (singleton),
    -- Normalización de F2 (P13): el dato normalizado es el que se guarda.
    CHECK (email = btrim(email)),
    CHECK (phone = btrim(phone))
);

-- Horario de servicios (FR-004, R3-5): día 0=domingo…6=sábado (selector localizado
-- por i18n, NUNCA texto libre), inicio "HH:MM" 24 h y fin opcional ("HH:MM", para
-- rangos tipo «10:00 a. m. − 12:00 m.»; CHECK: fin posterior al inicio).
-- Día y horas NO son traducibles (el nombre del día lo localiza el frontend y el
-- formato a.m./p.m. es presentación); nombre/descripción y lugar sí (patrón es/en).
CREATE TABLE home_services (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    day_of_week       SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time        TEXT NOT NULL CHECK (start_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
    end_time          TEXT NULL CHECK (end_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
    name_es           TEXT NOT NULL CHECK (char_length(name_es) BETWEEN 1 AND 160),
    name_en           TEXT NULL CHECK (char_length(name_en) BETWEEN 1 AND 160),
    description_es    TEXT NULL CHECK (char_length(description_es) BETWEEN 1 AND 400),
    description_en    TEXT NULL CHECK (char_length(description_en) BETWEEN 1 AND 400),
    place_es          TEXT NOT NULL CHECK (char_length(place_es) BETWEEN 1 AND 200),
    place_en          TEXT NULL CHECK (char_length(place_en) BETWEEN 1 AND 200),
    publication_state TEXT NOT NULL DEFAULT 'draft'
                      CHECK (publication_state IN ('draft', 'published')),
    sort_order        INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_time IS NULL OR end_time > start_time),
    CHECK (name_es = btrim(name_es)),
    CHECK (place_es = btrim(place_es))
);

CREATE INDEX home_services_sort_order_id_idx ON home_services (sort_order, id);

-- Canales de WhatsApp (FR-005, R3-6): kind=direct → destination es un teléfono
-- (se normaliza a dígitos con + opcional); kind=group → destination es una URL
-- https de chat.whatsapp.com o wa.me. Un canal exactamente igual (mismo kind,
-- destino y nombre normalizados) es duplicado y se rechaza (edge case de la spec).
CREATE TABLE home_whatsapp_channels (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind              TEXT NOT NULL CHECK (kind IN ('direct', 'group')),
    destination       TEXT NOT NULL CHECK (char_length(destination) BETWEEN 1 AND 320),
    name_es           TEXT NOT NULL CHECK (char_length(name_es) BETWEEN 1 AND 120),
    name_en           TEXT NULL CHECK (char_length(name_en) BETWEEN 1 AND 120),
    publication_state TEXT NOT NULL DEFAULT 'draft'
                      CHECK (publication_state IN ('draft', 'published')),
    sort_order        INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (kind, destination, name_es),
    CHECK (destination = btrim(destination)),
    CHECK (name_es = btrim(name_es))
);

CREATE INDEX home_whatsapp_channels_sort_order_id_idx ON home_whatsapp_channels (sort_order, id);

-- Redes sociales (FR-006/Q5, R3-7): catálogo fijo y UN enlace por red.
-- Ampliar el catálogo (p. ej. X) = migración nueva que extiende el CHECK + caso
-- de hosts en el service (ajuste menor acordado con el humano).
CREATE TABLE home_social_links (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    network           TEXT NOT NULL CHECK (network IN
                        ('facebook', 'instagram', 'youtube', 'tiktok', 'spotify')),
    url               TEXT NOT NULL CHECK (char_length(url) BETWEEN 1 AND 500),
    publication_state TEXT NOT NULL DEFAULT 'draft'
                      CHECK (publication_state IN ('draft', 'published')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (network),
    CHECK (url = btrim(url))
);
```

### `000005_create_home_content.down.sql`

```sql
-- 000005_create_home_content (down) — revierte 000005 por completo (§VI).
DROP TABLE IF EXISTS home_social_links;
DROP TABLE IF EXISTS home_whatsapp_channels;
DROP TABLE IF EXISTS home_services;
DROP TABLE IF EXISTS home_contact;
DROP TABLE IF EXISTS home_about;
DROP TABLE IF EXISTS home_identity;
```

### `000006_extend_admin_actions_for_home_content.up.sql`

Amplía el registro de auditoría de F2 (FR-017/R3-11). Los CHECK de 000004 que se sustituyen llevan
nombres generados por PostgreSQL al declararlos sin nombre: los de columna son
`admin_actions_action_check` y `admin_actions_target_kind_check`, y el segundo CHECK de tabla (el de
coherencia de objetivo, línea tras el CHECK del actor) es `admin_actions_check1`. La migración los
referencia explícitamente y la prueba de integración de migraciones (up → down → up) lo verifica
(ver Riesgo RG3-2 del plan; alternativa considerada: bloque `DO` que localiza la restricción por su
definición, descartada por legibilidad).

```sql
-- 000006_extend_admin_actions_for_home_content (up)
-- F3: la auditoría de F2 se amplía a la portada (FR-017). El registro de
-- códigos de acción y de tipos de objetivo sigue siendo CERRADO y lo garantiza
-- la BD (research R3-11).

-- 1) Códigos de acción: los 8 de F2 + los 15 de F3 (home.*; incluye
--    'home.image.upload' por la subida de imágenes — analyze I8).
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

-- 2) Tipo de objetivo: + 'content' (elemento de la portada; sin FK: son 6 tablas).
ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_target_kind_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_target_kind_check
    CHECK (target_kind IN ('user', 'role', 'content'));

-- 3) Coherencia de objetivo: 'content' exige ambas FK en NULL y target_label.
ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_check1;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_target_coherence_check CHECK (
    (target_kind = 'user'    AND target_role_id IS NULL) OR
    (target_kind = 'role'    AND target_user_id IS NULL) OR
    (target_kind = 'content' AND target_user_id IS NULL AND target_role_id IS NULL
                              AND target_label IS NOT NULL)
);
```

### `000006_extend_admin_actions_for_home_content.down.sql`

```sql
-- 000006_extend_admin_actions_for_home_content (down).
-- DESTRUCTIVO a propósito (igual que el down de 000004, que tira el registro):
-- las filas con objetivo 'content' no caben en el CHECK anterior, así que se
-- eliminan antes de restaurar. Solo tiene sentido en desarrollo/rollback.
DELETE FROM admin_actions WHERE target_kind = 'content';

ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_target_coherence_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_check1 CHECK (
    (target_kind = 'user' AND target_role_id IS NULL) OR
    (target_kind = 'role' AND target_user_id IS NULL)
);

ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_target_kind_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_target_kind_check
    CHECK (target_kind IN ('user', 'role'));

ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_action_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_action_check CHECK (action IN (
    'user.create', 'user.update', 'user.activate', 'user.deactivate', 'user.password_reset',
    'role.create', 'role.update', 'role.delete'
));
```

---

## Consultas sqlc (`backend/internal/db/queries/home.sql`)

Se generan en `internal/db` (paquete compartido) y solo las usa el repository del dominio `portada`
(R2: SQL compartido vía consulta nombrada). Listas con `ORDER BY` explícito y determinista
(§8.1.7): `sort_order, id` en las colecciones ordenables y `network` en redes. Sin `SELECT *`.

| Consulta | Uso |
|---|---|
| `UpsertHomeIdentity` (`INSERT … ON CONFLICT (singleton) DO UPDATE … RETURNING`) | PUT de identidad (panel) |
| `GetHomeIdentity` / `GetHomeIdentityPublished` | Panel (todo) / público (solo `published`) |
| `UpsertHomeAbout`, `GetHomeAbout`, `GetHomeAboutPublished` | Ídem «quiénes somos» |
| `UpsertHomeContact`, `GetHomeContact`, `GetHomeContactPublished` | Ídem contacto |
| `InsertHomeService`, `UpdateHomeService`, `DeleteHomeService`, `GetHomeServiceByID` | Gestión del horario |
| `ListHomeServices` / `ListHomeServicesPublished` | Panel (todo) / público (`WHERE publication_state='published'`), `ORDER BY sort_order, id` |
| `InsertHomeWhatsappChannel`, `UpdateHomeWhatsappChannel`, `DeleteHomeWhatsappChannel`, `GetHomeWhatsappChannelByID` | Gestión de canales |
| `ListHomeWhatsappChannels` / `ListHomeWhatsappChannelsPublished` | Ídem, `ORDER BY sort_order, id` |
| `InsertHomeSocialLink`, `UpdateHomeSocialLink`, `DeleteHomeSocialLink`, `GetHomeSocialLinkByID` | Gestión de redes |
| `ListHomeSocialLinks` / `ListHomeSocialLinksPublished` | Ídem, `ORDER BY network` |
| `IsHomeFilePublished` | Descarga de imágenes (`GET /api/v1/media/{fileName}`, `analyze` C4): `TRUE` solo si el nombre está referenciado por un contenido **publicado** (en F3: `home_identity` con `publication_state='published'` y `logo_file`/`cover_image_file` = el nombre); F4–F9 amplían la consulta con sus tablas |

La auditoría **reutiliza** `InsertAdminAction` (ya generada por F2 en `queries/audit.sql`): el
repository de `portada` la llama con `gendb.New(tx)` **dentro de la transacción** de cada mutación
(R3-11). No se escribe ningún `UPDATE`/`DELETE` sobre `admin_actions` (solo lectura, FR-025 de F2).

**Separación público/panel en consultas distintas** (`…Published`): el filtro `published` no puede
olvidarse en una llamada futura — la única función que devuelve borradores es la del panel, que
exige el permiso `portada` (FR-013/SC-002).

## Escritura de la auditoría desde F3 (detalle)

- La mutación y su fila de `admin_actions` comparten transacción (o ambos o ninguno). Si el INSERT
  del registro falla, **no se aplica** la edición (edge case de la spec: "no aplicarse sin registro").
- Objetivo: `target_kind='content'`, `target_user_id`/`target_role_id` = NULL,
  `target_label` = `Portada · <Sección> · <nombre del elemento>` (p. ej.
  `Portada · Horario · Culto dominical`; sin nombre, el `id`). Máx. 254 caracteres (CHECK de F2).
- `action` ∈ los 15 códigos `home.*` (tabla y **regla exacta** en research R3-11.1: `create` en el
  alta aunque nazca publicada, `home.publish`/`home.unpublish` **solo** para cambios de estado,
  `home.image.upload` en la subida de imágenes); `result`:
  `success` en la transacción de la mutación, `failure` en intentos rechazados y `denied` en
  denegaciones de permiso (best-effort, fuera de transacción).
- `actor_user_id` siempre la cuenta del panel que ejecutó (nunca NULL en F3: el CHECK de F2
  `actor_user_id IS NOT NULL OR action='user.create'` sigue vigente sin cambios).

## Invariantes de negocio y dónde se verifican

| Invariante | Dónde |
|---|---|
| Español obligatorio, inglés opcional | Columnas `*_es NOT NULL` / `*_en NULL` + `required`/`omitempty` en los DTOs (`platform/validate`) |
| `""`/espacios en un `*_en` → `NULL` (analyze I6) | Normalización de trim en el service (T317) antes de persistir; pruebas en T319/T320 (los `CHECK BETWEEN 1 AND …` rechazan `''`) |
| «Quiénes somos» ≤ 1.000 caracteres por idioma | `CHECK` de `home_about` + `max=1000` en el DTO (mensaje por campo) |
| Horario: inicio obligario "HH:MM" y fin opcional posterior | `CHECK` de patrón + `CHECK (end_time IS NULL OR end_time > start_time)`; día 0–6 (selector localizado, nunca texto libre) |
| Un solo enlace por red y red dentro del catálogo | `UNIQUE (network)` + `CHECK (network IN …)` + validación de host en el service (R3-7) |
| Canal de WhatsApp duplicado exacto rechazado | `UNIQUE (kind, destination, name_es)` sobre valores normalizados + `409 conflict` en el service (R3-6) |
| Destino de WhatsApp coherente con su tipo | Service (teléfono con tag `phone` si `direct`; URL https de dominio WhatsApp si `group`) — la BD guarda el destino normalizado |
| Toda imagen lleva texto alternativo en español | `CHECK (logo_file IS NULL OR logo_alt_es IS NOT NULL)` (+ idem cover) y `required` condicional en el DTO |
| Solo lo publicado es visible al visitante | Consultas `…Published` (única vía pública) + pruebas de "0 borradores expuestos"; **la descarga de imágenes también**: `IsHomeFilePublished` (analyze C4) → `404` si el archivo no está referenciado por contenido publicado |
| Sección sin elementos publicados → se omite | Service público (no publica claves vacías) + prueba de contrato |
| Máx. 1 fila en identidad/quiénes somos/contacto | `UNIQUE (singleton)` + upsert en transacción con advisory lock (R3-4) |
| Colecciones acotadas (≤50 servicios, ≤20 canales) | Service (constantes revisables) — el agregado del panel es un documento, no un listado paginado |
| Auditoría atómica con la mutación | Transacción única repository (`InsertAdminAction` sobre `tx`) + prueba de integración |
| Registro de auditoría de solo lectura | Sin consultas `UPDATE`/`DELETE` sobre `admin_actions` en `queries/` (FR-025 de F2) |

## Validación del modelo (qué comprobará la implementación)

1. `migrate up` limpio desde F2 (`000004`) y `down` completo de `000005`/`000006` sin errores
   (incluida la restauración de los CHECK de `admin_actions`: up → down → up, evidencia de RG3-2).
2. Pruebas de integración del repository (PostgreSQL real, `//go:build integration`):
   - upsert de singletons: dos `UpsertHome*` concurrentes dejan **una** fila con la última escritura;
   - `UNIQUE (network)`: segundo enlace para la misma red → error de restricción traducido a
     `conflict`;
   - `UNIQUE (kind, destination, name_es)` de WhatsApp → `conflict` solo en el duplicado exacto;
   - `CHECK` de `publication_state`, de `day_of_week`, de `start_time`/`end_time` (patrón y
     `end_time > start_time`) y de longitudes;
   - lecturas `…Published` nunca devuelven filas en `draft` (tampoco con texto `en` presente);
   - `IsHomeFilePublished`: `TRUE` solo con la identidad publicada que referencia el archivo;
     al retirar la identidad pasa a `FALSE` (la descarga responde `404`, analyze C4);
   - cada mutación deja su fila en `admin_actions` **en la misma transacción** (y al forzar un fallo
     en el registro, la mutación no se aplica);
   - inserción de `admin_actions` con `target_kind='content'`, `target_label` obligatorio y ambas FK
     en NULL (y rechazo del `CHECK` de coherencia con una FK rellena).
3. Tipos sqlc: los `NULL` llegan como `pgtype.Text`/`pgtype.UUID` y se traducen en `repository.go`
   (§8.1.6); `internal/db` regenerado y commiteado (`make sqlc-verify`).
