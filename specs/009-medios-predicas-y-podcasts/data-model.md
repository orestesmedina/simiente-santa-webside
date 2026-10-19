# Data Model — F9 Medios: prédicas y podcasts

**Fecha**: 2026-10-17 · **Spec**: [spec.md](./spec.md) · **Decisiones**: [research.md](./research.md)
(R9-1…R9-14) · **Plantilla de estilo**: el `data-model.md` de F7 (que sigue el de F6/F4).

Fuente de verdad del esquema: **este documento** → migraciones `backend/migrations/` → consultas
`backend/internal/db/queries/media_episodes.sql`, `media_episode_sources.sql`,
`media_sync_runs.sql` y la rama nueva de `media.sql` (sqlc). Convenciones: skill `postgres-db`,
§VI de la constitución y §8.1 de `docs/tecnico/arquitectura.md` (UUID, `created_at`/`updated_at`,
mapeo `pgtype` → dominio en el repository, consultas parametrizadas).

## Patrones transversales heredados de F3/F4/F6/F7 (no se reabren)

- **Bilingüismo por campo (R3-1/P7-9)**: el único campo traducible es la **descripción corta**
  (`description_es` base, `description_en` opcional; FR-025). `""`/espacios se guardan como `NULL`
  (trim en el service antes de persistir, patrón I6). `title`, `published_at`, `cover_file` y los
  enlaces **no** son traducibles (FR-026).
- **Estado de publicación — EXCEPCIÓN consciente (Consecuencia 5 de la spec)**: F9 **no** usa
  `publication_state ('draft','published')`. Su estado es **`visibility` con `'visible'/'hidden'`**
  (Q5/Q6: el contenido entra solo y no se borra). La sincronización **nunca** lo cambia salvo la
  **única excepción aprobada en la fase 5 (A3)**: la **indisponibilidad CONFIRMADA** (FR-017)
  oculta la ficha **aunque el equipo la hubiera dejado visible** (incluida una resolución previa
  «mantener visible» o «Mostrar de nuevo»); queda marcada y el equipo puede volver a mostrarla.
  Ninguna otra decisión del equipo se revierte (descripción, uniones, ocultar/mostrar por otras
  causas).
- **Protección de las decisiones del equipo (FR-015/SC-007)**: `description_edited` fija si la
  descripción corta es del equipo; la búsqueda no la toca. Las uniones confirmadas viven como
  `merged_into_id` y no se deshacen (R9-12). Una marca **resuelta** no se reabre por la **misma**
  causa (default #6): cambiar de **clase** de marca sí resetea la resolución y la marca nueva nace
  abierta (ver «Modelo de marcas en dos ejes» más abajo).
- **Auditoría (R7-10)**: las mutaciones insertan `admin_actions` con `InsertAdminAction` **dentro
  de su transacción** (`target_kind='content'`, FK en `NULL`, `target_label` descriptivo); registro
  de solo lectura. La búsqueda completa deja **una** fila `media.sync` por ejecución (el detalle
  numérico vive en `media_sync_runs`).
- **Normalización (F2/P13)**: el dato normalizado es el que se guarda —texto recortado
  (`btrim`), enlaces validados por host/esquema antes de persistir (R9-8), carátula copiada con
  nombre generado por el servidor (`img_<uuid>.<ext>`, `platform/storage`).

## Resumen de tablas

| Tabla | Cardinalidad | Qué guarda |
|---|---|---|
| `media_items` (**nueva**) | 0..N | **Una ficha por sermón** (FR-001): `title` (del feed, no editable), `published_at`, `description_es`/`description_en` (corta, editable por el equipo), `description_edited`, `cover_file` (copia de la carátula), `origin` (`video`/`audio`/`both`), **`visibility`** (`visible`/`hidden`) y **dos ejes de marcas independientes** (corrección C1 de la fase 5): **duplicados** (`duplicate_of_id` = par recíproco + `duplicate_mark_resolved_at`) y **disponibilidad** (`availability_mark` + `availability_mark_resolved_at`, sobre `availability_checked_at`/`availability_result`) y el puntero de unión `merged_into_id` |
| `media_episode_sources` (**nueva**) | 1..2 por ficha | **Una fila por elemento de feed importado** (R9-11): `provider` (`youtube`/`podcast`), `external_id` (**UNIQUE por proveedor**), `feed_title` (procedencia), `url` (enlace validado) y su `item_id`. Es la garantía de idempotencia y el vínculo video↔audio |
| `media_sync_runs` (**nueva**) | 1 por búsqueda | **Registro de cada «Buscar episodios nuevos»** (FR-018): quién, cuándo, cuántos entraron/se unieron/se marcaron/ocultaron, comprobaciones de disponibilidad y qué feeds fallaron |
| `admin_actions` (**existente, F2**) | — | Se **amplía** el registro cerrado: 6 códigos `media.*` (FR-022, R9-12 de `plan`) |
| `permissions` (**existente, F2**) | — | **SIN CAMBIOS**: `('medios', 'Medios (prédicas y podcasts)')` ya está sembrado por `000002`; F9 **no** crea permisos nuevos (FR-021: el catálogo sigue en **10**) |
| `roles` / `role_permissions` / `users` (**existentes, F2**) | — | **SIN CAMBIOS**; `media_sync_runs.actor_user_id` es la única FK nueva hacia `users` (quién pulsó el botón; las cuentas no se eliminan —decisión 9—) |
| `home_identity`, `agenda_*`, `ministries` | — | **SIN CAMBIOS**; solo se **amplía la consulta** `IsFilePublished` (`queries/media.sql`) con la rama de F9 |

**Ninguna tabla de F9 guarda el audio ni el video** (FR-001: el sitio no aloja ni transmite la
grabación). Lo único que se conserva del exterior es **una copia de la carátula** (`cover_file`
en `platform/storage`, servida por `GET /api/v1/media/{fileName}`), que por eso necesita su rama
en `IsFilePublished`. La **descripción completa del feed no se guarda** (FR-020: no se publica y
R9-6 descarta la regeneración).

---

## Migraciones

Dos migraciones nuevas (F7 terminó en `000013`): **`000014`** crea `media_items`,
`media_episode_sources` y `media_sync_runs`; **`000015`** amplía el registro de auditoría con los
6 códigos `media.*`. **No hay migración de permisos** (el permiso `medios` ya existe) y ninguna
migración existente se toca (§VI). Las pruebas de migración usarán
`testutil.SerializeMigrations` (decisión 25 del roadmap: los ciclos destructivos se serializan).

### `000014_create_media_content.up.sql`

```sql
-- 000014_create_media_content (up)
-- F9 Medios: una ficha por sermón (FR-001) con sus fuentes de importación —una
-- fila por elemento de feed, UNIQUE (provider, external_id): idempotencia de la
-- búsqueda, R9-11— y el registro de cada «Buscar episodios nuevos» (FR-018).
-- Estado = visibility 'visible'/'hidden' (Consecuencia 5: F9 no usa
-- borrador/publicado). Convenciones: skill `postgres-db` y
-- specs/009-medios-predicas-y-podcasts/data-model.md (fuente de verdad).
-- Decisiones: research R9-4…R9-12.

CREATE TABLE media_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Nombre del episodio: el del feed, NO editable (FR-003). Al unir, manda el
    -- del feed del video (default revisable #7). Se conserva tal cual (sin
    -- truncar por debajo de 300: los feeds reales van muy por debajo).
    title             TEXT NOT NULL CHECK (char_length(title) BETWEEN 1 AND 300),
    -- Fecha de publicación del feed (se guarda en UTC; el corte de "mismo día"
    -- y el orden público se calculan en APP_TZ = America/Costa_Rica, FR-006 y
    -- R9-5: el límite de medianoche local cambia el día).
    published_at      TIMESTAMPTZ NOT NULL,

    -- Descripción corta (FR-020): recorte de la descripción del feed al
    -- importar y ÚNICO campo ajustable por el equipo (con versión en inglés
    -- opcional, FR-025). description_edited protege la versión del equipo
    -- (FR-015/SC-007: la sincronización jamás la pisa). NULL solo cuando el
    -- feed no trae descripción (hueco documentado en research R9-14.3): la
    -- ficha omite la línea (SC-010: sin huecos).
    description_es    TEXT NULL CHECK (char_length(description_es) BETWEEN 1 AND 300),
    description_en    TEXT NULL CHECK (char_length(description_en) BETWEEN 1 AND 300),
    description_edited BOOLEAN NOT NULL DEFAULT FALSE,

    -- Carátula COPIADA al sitio al importar (FR-004/SC-005: 0 recursos de
    -- terceros en el navegador del visitante). Nombre generado por el servidor
    -- (platform/storage: img_<uuid>.<jpg|png|webp>); NULL → marcador neutro
    -- (default revisable #4). No se actualiza (default #10).
    cover_file        TEXT NULL CHECK (char_length(cover_file) BETWEEN 1 AND 120),

    -- Formatos que ofrece la ficha, derivados de sus fuentes (FR-003):
    -- 'video' (solo YouTube), 'audio' (solo podcast) o 'both'. Lo mantiene el
    -- service en la misma transacción que las fuentes (invariante probado).
    origin            TEXT NOT NULL CHECK (origin IN ('video', 'audio', 'both')),

    -- Visibilidad (Q5/FR-015): decide el equipo y la sincronización NUNCA la
    -- cambia, salvo la ÚNICA excepción aprobada en la fase 5 (A3): la
    -- indisponibilidad CONFIRMADA (FR-017) oculta la ficha aunque el equipo la
    -- hubiera dejado visible; queda marcada y el equipo puede «Mostrar de
    -- nuevo».
    visibility        TEXT NOT NULL DEFAULT 'visible'
                      CHECK (visibility IN ('visible', 'hidden')),

    -- EJE DE DUPLICADOS (FR-014) — corrección C1 de la fase 5: el par y la
    -- resolución de SU marca viven aquí, INDEPENDIENTES del eje de
    -- disponibilidad (antes una única review_mark con CHECK de simetría hacía
    -- imposible marcar «no disponible» sobre una ficha con par sin romper la
    -- pareja recíproca). duplicate_of_id IS NOT NULL = hay par (cada miembro
    -- apunta al otro; lo escribe siempre el service, con prueba del recíproco:
    -- no cabe en un CHECK porque mira otra fila). duplicate_mark_resolved_at
    -- IS NOT NULL = «posible duplicado» RESUELTO («no son duplicados»); NULL =
    -- pendiente. El bloqueo público es SOLO el par con resolución NULL
    -- (síntesis de FR-009/SC-002/FR-017, research R9-7).
    duplicate_of_id            UUID NULL REFERENCES media_items(id),
    duplicate_mark_resolved_at TIMESTAMPTZ NULL,

    -- EJE DE DISPONIBILIDAD (FR-017, R9-2) — estado propio de la ficha, no una
    -- marca que compite con la de duplicado: los dos ejes pueden coexistir
    -- (par abierto + enlace confirmado muerto es un estado legal). La marca
    -- registra su CLASE ('unavailable' = la comprobación confirmó 404/410;
    -- 'missing_from_feed' = todas sus fuentes salieron de la ventana del feed
    -- sin confirmación de muerte) y su resolución. availability_checked_at /
    -- availability_result son el dato que la motivó: resultado AGREGADO de las
    -- comprobaciones de sus fuentes (el peor gana — ver «Regla de agregación»);
    -- 'unavailable' solo se fija con 404/410 confirmados; 'inconclusive' nunca
    -- oculta nada (regla a prueba de fallos).
    availability_checked_at       TIMESTAMPTZ NULL,
    availability_result           TEXT NULL CHECK (availability_result IN
                                     ('available', 'unavailable', 'inconclusive')),
    availability_mark             TEXT NULL CHECK (availability_mark IN
                                     ('unavailable', 'missing_from_feed')),
    availability_mark_resolved_at TIMESTAMPTZ NULL,

    -- Unión confirmada (R9-12): la ficha absorbida apunta a la resultante y
    -- desaparece de TODAS las consultas (merged_into_id IS NULL). Sin DELETE en
    -- F9 (FR-019: la acción es ocultar; la unión deja puntero).
    merged_into_id    UUID NULL REFERENCES media_items(id),

    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Español base por campo (FR-025/Decisión 8): description_en exige
    -- description_es.
    CHECK (description_en IS NULL OR description_es IS NOT NULL),
    -- Resolución exige marca en SU eje; sin marca no hay resolución (cada eje
    -- se resuelve por separado — la operación indica cuál —).
    CHECK (duplicate_mark_resolved_at IS NULL OR duplicate_of_id IS NOT NULL),
    CHECK (availability_mark_resolved_at IS NULL OR availability_mark IS NOT NULL),
    -- Normalización de F2 (P13): textos cortos guardados ya recortados
    -- (precedente literal de 000010/000012).
    CHECK (title = btrim(title)),
    CHECK (description_es IS NULL OR description_es = btrim(description_es)),
    CHECK (description_en IS NULL OR description_en = btrim(description_en))
);

-- Orden público y de panel (FR-006): published_at DESC y, ante la misma fecha,
-- orden de importación FIFO (created_at ASC, id) — default revisable #11,
-- research R9-14.5. El filtro habitual es visibility + no-absorbida.
CREATE INDEX media_items_order_idx ON media_items (published_at, created_at, id);
CREATE INDEX media_items_visibility_idx ON media_items (visibility);
CREATE INDEX media_items_duplicate_of_idx ON media_items (duplicate_of_id);
CREATE INDEX media_items_availability_mark_idx ON media_items (availability_mark);
CREATE INDEX media_items_merged_into_idx ON media_items (merged_into_id);

-- Una fila por elemento de feed importado (research R9-11). UNIQUE
-- (provider, external_id) es la garantía de que una búsqueda repetida NUNCA
-- duplica ni reimporta, incluso con dos búsquedas simultáneas (default #12).
CREATE TABLE media_episode_sources (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Ficha a la que pertenece la fuente. Al unir, la fuente se REPUNTA a la
    -- ficha resultante (R9-12); nunca se borra (conserva la procedencia).
    item_id      UUID NOT NULL REFERENCES media_items(id),
    -- Proveedor del feed (FR-010): solo YouTube y Spotify for Creators/Anchor.
    provider     TEXT NOT NULL CHECK (provider IN ('youtube', 'podcast')),
    -- Identificador externo del elemento del feed (R9-5): yt:videoId en
    -- YouTube, <guid> en el podcast. Es la llave de idempotencia.
    external_id  TEXT NOT NULL CHECK (char_length(external_id) BETWEEN 1 AND 120),
    -- Título tal como lo trae ESTE feed (procedencia, FR-003/default #7: al
    -- unir, el título del podcast se conserva como dato de origen).
    feed_title   TEXT NOT NULL CHECK (char_length(feed_title) BETWEEN 1 AND 300),
    -- Enlace del elemento, validado contra el host esperado del proveedor
    -- ANTES de persistir (R9-8/seguridad): solo https y hosts del proveedor.
    url          TEXT NOT NULL CHECK (char_length(url) BETWEEN 1 AND 500),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (provider, external_id),
    -- Defensa en profundidad: el service valida el enlace completo; la BD
    -- exige la forma esperada por proveedor (patrón de los CHECK de 000012).
    CHECK (url ~ '^https://'),
    CHECK ((provider = 'youtube'  AND url ~ '^https://(www\.)?youtube\.com/watch\?v=[A-Za-z0-9_-]{11}$')
        OR (provider = 'podcast' AND url ~ '^https://podcasters\.spotify\.com/')),
    CHECK (external_id = btrim(external_id)),
    CHECK (feed_title = btrim(feed_title))
);

-- Recuperar las fuentes de una ficha (panel, unión) y detectar lo que ya no
-- está en la ventana del feed (FR-017: por proveedor).
CREATE INDEX media_episode_sources_item_idx ON media_episode_sources (item_id);
CREATE INDEX media_episode_sources_provider_idx ON media_episode_sources (provider, external_id);

-- Registro de cada «Buscar episodios nuevos» (FR-018/R9-9). Una fila por
-- ejecución completa (con conteos parciales si el presupuesto de tiempo se
-- agotó o un feed falló: los fallos se informan, no se esconden).
CREATE TABLE media_sync_runs (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Quién pulsó el botón (FR-022; las cuentas de F2 no se eliminan).
    actor_user_id               UUID NULL REFERENCES users(id),
    started_at                  TIMESTAMPTZ NOT NULL,
    finished_at                 TIMESTAMPTZ NOT NULL,
    -- Resumen de FR-018 (más los contadores de disponibilidad de FR-017).
    imported_count              INTEGER NOT NULL DEFAULT 0 CHECK (imported_count >= 0),
    joined_count                INTEGER NOT NULL DEFAULT 0 CHECK (joined_count >= 0),
    -- marked_count cuenta SOLO las marcas «posible duplicado» nuevas de ESTA
    -- búsqueda (literal de FR-018, criterio B2 de la fase 5). Las marcas de
    -- disponibilidad nuevas NO cuentan aquí: las que ocultan van en
    -- hidden_count y su huella numérica en availability_*; el detalle por
    -- ficha se ve en el listado del panel.
    marked_count                INTEGER NOT NULL DEFAULT 0 CHECK (marked_count >= 0),
    hidden_count                INTEGER NOT NULL DEFAULT 0 CHECK (hidden_count >= 0),
    availability_checked_count  INTEGER NOT NULL DEFAULT 0
                                CHECK (availability_checked_count >= 0),
    availability_inconclusive_count INTEGER NOT NULL DEFAULT 0
                                CHECK (availability_inconclusive_count >= 0),
    -- Estado de cada feed en ESTA ejecución (FR-016): 'failed' nunca oculta ni
    -- marca nada por esa causa.
    youtube_status              TEXT NOT NULL CHECK (youtube_status IN ('ok', 'failed')),
    podcast_status              TEXT NOT NULL CHECK (podcast_status IN ('ok', 'failed')),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (finished_at >= started_at)
);

-- "Última búsqueda" del panel: ORDER BY started_at DESC, id.
CREATE INDEX media_sync_runs_started_at_idx ON media_sync_runs (started_at DESC, id);
```

### `000014_create_media_content.down.sql`

```sql
-- 000014_create_media_content (down) — revierte 000014 por completo (§VI).
DROP TABLE IF EXISTS media_sync_runs;
DROP TABLE IF EXISTS media_episode_sources;
DROP TABLE IF EXISTS media_items;
```

### `000015_extend_admin_actions_for_media.up.sql`

Amplía el registro de auditoría de F2 (FR-022/R9-12 de plan). El `CHECK` que se sustituye es
`admin_actions_action_check` con nombre propio desde `000006` (F3, F4, F6 y F7 hicieron este mismo
ejercicio y su prueba de integración lo verifica con up → down → up). El `target_kind` **no**
cambia: las fichas son `content` como los elementos de F3/F4/F6/F7. Los **6 códigos** nuevos:
`media.sync`, `media.description.update`, `media.hide`, `media.show`, `media.merge`,
`media.mark.resolve` (patrón de los 8 `donation.*` de F7; F9 no tiene `create`/`delete` porque no
hay alta ni borrado —FR-019—).

```sql
-- 000015_extend_admin_actions_for_media (up)
-- F9: la auditoría de F2 se amplía a medios (FR-022). El registro de códigos de
-- acción sigue siendo CERRADO y lo garantiza la BD.

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
    'ministry.unpublish', 'ministry.leader.show', 'ministry.leader.hide', 'ministry.image.upload',
    'donation.create', 'donation.update', 'donation.delete', 'donation.publish',
    'donation.unpublish', 'donation.text.update', 'donation.text.publish', 'donation.text.unpublish',
    'media.sync', 'media.description.update', 'media.hide', 'media.show', 'media.merge',
    'media.mark.resolve'
));
```

### `000015_extend_admin_actions_for_media.down.sql`

```sql
-- 000015_extend_admin_actions_for_media (down): restaura el CHECK de F7
-- (8 códigos de F2 + 15 de F3 + 15 de F4 + 8 de F6 + 8 de F7). Como en el down
-- de 000006/000008/000011/000013, es destructivo a propósito: las filas con
-- códigos F9 no caben en el CHECK anterior y se eliminan antes de restaurar.
-- Solo tiene sentido en desarrollo/rollback.
DELETE FROM admin_actions WHERE action IN (
    'media.sync', 'media.description.update', 'media.hide', 'media.show', 'media.merge',
    'media.mark.resolve'
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
    'category.create', 'category.update', 'category.delete',
    'ministry.create', 'ministry.update', 'ministry.delete', 'ministry.publish',
    'ministry.unpublish', 'ministry.leader.show', 'ministry.leader.hide', 'ministry.image.upload',
    'donation.create', 'donation.update', 'donation.delete', 'donation.publish',
    'donation.unpublish', 'donation.text.update', 'donation.text.publish', 'donation.text.unpublish'
));
```

**Notas de diseño.**

- **`media_episode_sources` con `UNIQUE (provider, external_id)`** (R9-11): es la pieza que hace
  que «Buscar episodios nuevos» sea **idempotente por construcción** —el `INSERT … ON CONFLICT DO
  NOTHING` descarta lo ya visto sin leer antes— y que la unión de dos fichas **no** pueda
  reimportar lo absorbido (los identificadores viven en la ficha resultante).
- **`origin` es columna mantenida, no generada**: Postgres no puede derivarla de otra tabla en una
  columna generada; el service la recalcula en la misma transacción que toca fuentes (alta,
  unión) y hay prueba de invariante (`origin='both'` ⇔ la ficha tiene fuente `youtube` **y**
  `podcast`).
- **Modelo de marcas en dos ejes (corrección C1 de la fase 5; `plan.md` P9-19)**: la marca
  «posible duplicado» pertenece al **flujo de duplicados** (`duplicate_of_id` +
  `duplicate_mark_resolved_at`) y la disponibilidad tiene **estado propio**
  (`availability_mark` + `availability_mark_resolved_at`, sobre `availability_*`). **Nunca
  compiten**: una ficha puede estar a la vez en un par dudoso y marcada «no disponible» (y de hecho
  es uno de los casos obligados de prueba: par abierto + enlace confirmado muerto → oculta + dos
  marcas, par intacto para «Revisar duplicado»/«Unir»). La **clase** de la marca solo cambia dentro
  del eje de disponibilidad (`unavailable` ↔ `missing_from_feed`): **cambiar de clase resetea la
  resolución** y la marca nueva nace **abierta**; con la **misma** clase, una marca resuelta **no se
  reabre** (default #6).
- **Precedencia (aprobada por el humano en la fase 5, A3)**: **la indisponibilidad confirmada
  manda**. Una comprobación que confirma 404/410 oculta la ficha y deja la marca «no disponible»
  **abierta** aunque el equipo hubiera resuelto antes «mantener visible» o «Mostrar de nuevo» (es la
  promesa del `ConfirmDialog` de D-28 y lo que fija SC-007 —precisión de la fase 5—). La
  no-reapertura protege contra el re-marcado **sin evidencia nueva**, no contra una confirmación
  nueva.
- **Cola de comprobación por búsqueda**: solo se comprueba lo que salió de la ventana del feed y
  pertenece a fichas **`visibility='visible'` no absorbidas** (tope 30). De ahí salen dos
  consecuencias buscadas: una ficha oculta por el equipo **no** se re-comprueba (nada de lo que
  decide el equipo se reescribe) y «Mostrar de nuevo» devuelve la ficha a la cola — si el enlace
  sigue muerto, la próxima búsqueda la oculta otra vez (D-28) —.
- **Regla de agregación de `availability_*` (corrección M2 de la fase 5)**: la comprobación es **por
  fuente** (una por elemento de feed que salió de la ventana; hasta 2 por ficha unida) y el
  resultado de la ficha es **el peor** de los resultados de las fuentes comprobadas en esa pasada
  (`unavailable` peor que `inconclusive`, peor que `available`). Consecuencias: (a) un enlace
  confirmado muerto **oculta la ficha entera** aunque el otro responda 200 (precisión 2); (b) la
  marca `missing_from_feed` solo se pone cuando **todas** las fuentes de la ficha están fuera de sus
  ventanas —si solo salió la del podcast y el video sigue en el suyo, **no hay marca** (decir «ya no
  aparece en el feed» sería falso); (c) con feeds caídos no se comprueba nada de ese proveedor
  (FR-016) y su fuente no cuenta como «fuera de ventana».
- **`duplicate_of_id` sin `CHECK` de simetría** (corrección C1): la pareja recíproca (cada miembro
  apunta al otro) es **invariante de service con prueba** (T914/T920), no un `CHECK`: un `CHECK` de
  simetría contra la clase de marca era precisamente lo que impedía marcar disponibilidad sobre un
  par. Lo que sí exige la BD: `duplicate_mark_resolved_at` solo con par y
  `availability_mark_resolved_at` solo con marca.
- **La pareja es de dos — limitación aceptada (decisión del humano, 2026-10-19)**: `duplicate_of_id`
  modela **un par recíproco**, no grupos. Un tercer elemento **dudoso** (solo título o solo día)
  contra una ficha que **ya tiene pareja** no puede entrar en un par nuevo (el candidato se excluye
  en la rama dudosa) y **entra suelto y visible**, sin marca; solo es un problema si resulta ser el
  mismo sermón —mitigación: el equipo lo oculta a mano—. La **unión automática segura** (título
  normalizado + mismo día, FR-013) **sí** funciona contra una ficha emparejada. Sin cambios en el
  modelo aprobado; **mejora futura** posible: extender el eje de duplicados a grupos.
- **`availability_*` separados de las marcas**: la marca es para el equipo («revisa esto») y el
  resultado de disponibilidad es el **dato** que la motivó; así el panel puede mostrar «ya no
  aparece en el feed · última comprobación: inconclusa (12:03)» sin inventar marcas nuevas.
- **`merged_into_id` sin `DELETE`** (R9-12/FR-019): literal «no hay borrado de episodios». El
  índice permite excluir absorbidas con coste mínimo. La alternativa (borrar la fila vacía) está
  documentada y descartada en research R9-12. La unión **despeja el eje de duplicados** de ambas
  fichas (par y resolución) y **no toca** el eje de disponibilidad (es independiente; si la
  resultante tuviera una marca de disponibilidad abierta, sigue).
- **`media_sync_runs.actor_user_id` con FK a `users`**: única FK nueva hacia `users`; las cuentas
  de F2 no se eliminan (decisión 9), de modo que el «quién» del FR-018/FR-022 es estable. En
  `admin_actions` el actor sigue viajando como hace F2 (`actor_user_id`).
- **Sin `sort_order`** (a diferencia de F3/F4/F6/F7): el orden de F9 es **por fecha de
  publicación** con desempate por importación (FR-006/default #11), no manual.
- **Sin `deleted_at` ni historial**: fuera de alcance (spec: «sin historial de versiones ni
  papelera»).

---

## Consultas sqlc

### `backend/internal/db/queries/media_episodes.sql` (dominio `medios`, nuevo)

Listas con `ORDER BY` explícito y determinista (§8.1.7):
`ORDER BY published_at DESC, created_at ASC, id`. Sin `SELECT *`. Todo parametrizado (§IV).
**Separación público/panel en consultas distintas** (la única que ve fichas no publicables es la
del panel).

| Consulta | Uso |
|---|---|
| `ListPublishedMediaItems` (`:limit`, `:offset`) | **Público** (sección `/medios`): `WHERE visibility='visible' AND merged_into_id IS NULL AND (duplicate_of_id IS NULL OR duplicate_mark_resolved_at IS NOT NULL)` — el predicado público completo de R9-7 —, orden de sección, `LIMIT/OFFSET` (tandas de 12, «Cargar más») |
| `CountPublishedMediaItems` | Total del listado público (sobre de paginación y testigo de «nada visible» → redirección) |
| `ListPortadaMediaItems` (`:limit`) | **Bloque de la portada**: mismo predicado, `LIMIT 5` en el servidor (FR-027/SC-015) |
| `ListMediaItems` (`:limit`, `:offset`, `sqlc.narg('visibility')`, `sqlc.narg('review_mark')`, `sqlc.narg('include_merged')`) | Panel paginado (por defecto **sin** absorbidas), el mismo orden público; filtros opcionales para revisar marcas. Semántica de `review_mark` (corrección de la fase 5): filtra por la marca de esa clase **pendiente** (sin resolver) — `possible_duplicate` → `duplicate_of_id IS NOT NULL AND duplicate_mark_resolved_at IS NULL`; `unavailable`/`missing_from_feed` → `availability_mark = <clase> AND availability_mark_resolved_at IS NULL` |
| `CountMediaItems` (mismos filtros) | Total del listado del panel |
| `GetMediaItemByID` | Panel (detalle, respuestas de escritura) |
| `InsertMediaItem` / `UpdateMediaItem` | Alta al importar y escritura completa ya fusionada por el service (descripción, visibilidad, marcas, `merged_into_id`, disponibilidad) — patrón `UpdateDonationMethod` de F7 |
| `RepointMediaItemSources` (`:from_item_id`, `:to_item_id`) | Unión: mueve las fuentes de la ficha absorbida a la resultante (R9-12) |
| `ListMediaItemSourcesByItem` (`:item_id`) | Fuentes de una ficha (panel: procedencia; unión) |

### `backend/internal/db/queries/media_episode_sources.sql` (dominio `medios`, nuevo)

| Consulta | Uso |
|---|---|
| `InsertMediaEpisodeSource` (`ON CONFLICT (provider, external_id) DO NOTHING`) | Importación idempotente (R9-11); devuelve la fila insertada o vacío si ya existía |
| `ListMediaEpisodeSourcesByProvider` (`:provider`) | Búsqueda: qué identificadores hay importados de cada proveedor (para detectar «ya no está en la ventana del feed», FR-017) |
| `GetMediaEpisodeSource` (`:provider`, `:external_id`) | Existe-ya (salto de importación) y detalle para las comprobaciones |

### `backend/internal/db/queries/media_sync_runs.sql` (dominio `medios`, nuevo)

| Consulta | Uso |
|---|---|
| `InsertMediaSyncRun` | Fin de cada búsqueda (el resumen de FR-018 también se persiste) |
| `GetLastMediaSyncRun` | «Última búsqueda» del panel (`ORDER BY started_at DESC, id LIMIT 1`) |

### Rama nueva en `backend/internal/db/queries/media.sql` (dominio `media`, existente)

`IsFilePublished` (la consulta compartida de la descarga pública de imágenes, R4-6/R6-7) **crece
con una rama aditiva** —mismo gesto que F6—:

```sql
        UNION ALL

        -- F9 (R9-8): la carátula copiada de una ficha PUBLICABLE. Rama
        -- aditiva: una ficha oculta, absorbida o con «posible duplicado»
        -- abierto NO sirve su carátula y el archivo no se borra (al volver a
        -- ser publicable, la misma URL responde 200).
        SELECT m.id
        FROM media_items m
        WHERE m.cover_file = sqlc.arg('file_name')::text
          AND m.visibility = 'visible'
          AND m.merged_into_id IS NULL
          AND (m.duplicate_of_id IS NULL
               OR m.duplicate_mark_resolved_at IS NOT NULL)
```

## Escritura de la auditoría desde F9 (detalle; patrón R7-10 sin cambios)

- Mutación y fila de `admin_actions` comparten transacción (o ambos o ninguno).
- Objetivo: `target_kind='content'`, `target_user_id`/`target_role_id` = NULL,
  `target_label` = `Medios · <title>` (máx. 254) o `Medios · Búsqueda` para `media.sync`.
- Regla exacta de códigos (**6**): `media.sync` por cada búsqueda **completada** (una fila; el
  resumen numérico vive en `media_sync_runs` y un extracto —p. ej. «3 entraron, 1 unió, 2
  marcaron»— viaja en `target_label`); `media.description.update` al guardar la descripción corta
  (incluye el primer ajuste); `media.hide`/`media.show` **solo** para cambios de `visibility`
  (si un guardado cambia datos **y** visibilidad: **dos filas**, como F7); `media.merge` por unión
  confirmada (target de la ficha resultante; el id de la absorbida va en el `target_label`);
  `media.mark.resolve` por resolución de marca (**con la marca concreta que se resuelve** — la
  operación indica el eje — y la visibilidad elegida en el `target_label`).
  El alta **nunca** emite `media.*` extra: lo importa la búsqueda (`media.sync`).
- `result`: `success` en la transacción, `failure` en rechazos y `denied` en denegaciones
  (best-effort, `audit.Recorder` del dominio, como `donaciones/service_audit.go`).
- La inserción de `admin_actions` de una mutación **es condición de la operación** (si el registro
  falla, no se aplica — edge case «nunca aplicarse sin registro» de la spec —). La fila de
  `media_sync_runs` es parte de la misma escritura final de la búsqueda.

## Invariantes de negocio y dónde se verifican

| Invariante | Dónde |
|---|---|
| Una ficha = una entrada por sermón con 1–2 fuentes y `origin` coherente (FR-001/Q18) | `origin` recalculado en la transacción + prueba de invariante (`'both'` ⇔ fuentes `youtube` **y** `podcast`) |
| **Nunca** dos veces el mismo elemento de feed (FR-012) | `UNIQUE (provider, external_id)` + `INSERT … ON CONFLICT DO NOTHING` + prueba con búsqueda repetida y con dos búsquedas simultáneas |
| La búsqueda **no** cambia visibilidad, descripción ajustada ni uniones (FR-015/SC-007) | Update de la búsqueda solo toca los ejes de marcas, `availability_*` y `visibility` **solo por FR-017** (excepción aprobada en la fase 5: la indisponibilidad confirmada oculta aunque el equipo hubiera dejado visible — A3 —); prueba que fija cada campo antes y después de sincronizar |
| Solo lo publicable sale al visitante (FR-009/SC-002/R9-7) | Predicado en `ListPublished*`/`ListPortada*` (única vía pública) + prueba «0 ocultos, 0 absorbidas, 0 duplicados abiertos expuestos» |
| **Dos ejes de marcas independientes** (C1): par y disponibilidad nunca compiten | Sin `CHECK` que ate `availability_mark` a `duplicate_of_id`; prueba de integración de la matriz **par abierto/resuelto × enlace vivo/muerto/inconcluso** (T914) y de service (T921); par recíproco (`A→B` y `B→A`) escrito siempre por el service + prueba |
| Una marca resuelta no se reabre por la misma causa (default #6) | Misma clase + sin confirmación nueva → el sync no escribe; **cambiar de clase resetea la resolución** (marca nueva abierta) y **la indisponibilidad confirmada manda siempre** (A3: nueva marca abierta aunque hubiera resolución previa); prueba en T921 |
| «No disponible» solo con confirmación (FR-017/SC-014) | Solo `404/410` fijan `unavailable` y `visibility='hidden'`; `inconclusive` nunca oculta; **agregación por ficha: el peor de sus fuentes comprobadas gana** y `missing_from_feed` solo cuando **todas** sus fuentes están fuera de ventana (M2); tabla de casos del probador + prueba |
| La cola de comprobación solo mira lo que salió del feed y está visible | Solo fuentes fuera de ventana de fichas `visibility='visible'` no absorbidas (B4: un enlace muerto **dentro** de la ventana se publica hasta que salga de ella — matiz documentado en `quickstart.md` §8 —); prueba |
| Un feed caído no oculta ni marca nada (FR-016) | `…/busca-nuevos` con un `FeedSource` que falla: 0 escrituras de esa rama; resumen con `*_status='failed'` |
| Unión = una sola ficha con los dos enlaces, sin DELETE (R9-12) | `RepointMediaItemSources` + `merged_into_id` + `origin='both'` en la transacción; prueba + auditoría `media.merge` |
| Carátula solo de hosts permitidos y servida solo por contenido publicable (R9-8) | validación de URL/allowlist/DNS/LimitReader en `caratula.go` + rama de `IsFilePublished` + pruebas |
| La descripción corta del equipo sobrevive (FR-020) | `description_edited`; la búsqueda no la escribe; prueba (SC-007) |
| Español por campo: `description_en` exige `description_es` (FR-023/Decisión 8) | `CHECK` + `400 invalid` con `details.descriptionEs` |
| Texto plano, nunca código (FR-023/§IV) | descripciones como texto plano (HTML del feed **descartado** al importar); sin `dangerouslySetInnerHTML`; prueba con `<script>`, emojis, tildes intactas |
| Auditoría atómica con la mutación y registro de solo lectura | transacción única + sin `UPDATE`/`DELETE` sobre `admin_actions` |
| Permisos: `medios` ya existe y el catálogo sigue en 10 (FR-021) | Sin migración de permisos; `AdminChain("medios")`; prueba de integración que fija el catálogo en 10 |
| Orden público `published_at DESC, created_at ASC, id` (FR-006/default #11) | `ORDER BY` + índice `media_items_order_idx` + prueba de orden con fechas iguales |

## Validación del modelo (qué comprobará la implementación)

1. `migrate up` limpio desde F7 (`000013`) y `down` completo de `000014`/`000015` sin errores
   (incluida la restauración del `CHECK` de `admin_actions`: up → down → up, con
   `testutil.SerializeMigrations`).
2. Pruebas de integración del repository (PostgreSQL real, `//go:build integration`):
   - el catálogo de `permissions` **sigue en 10 filas** y `medios` entre ellas;
   - `UNIQUE (provider, external_id)` rechaza el doble insert (y el `ON CONFLICT` no falla);
     dos fuentes del **mismo** proveedor **no** caben en la misma ficha (regla de service +
     prueba: una ficha = 1 video y/o 1 audio);
   - `CHECK` de `url` por proveedor: una URL de YouTube con `javascript:`, con host ajeno o sin
     `v=<11>` falla en la BD; una de podcast que no sea `podcasters.spotify.com` falla;
   - `CHECK` de marcas: `duplicate_mark_resolved_at` sin `duplicate_of_id` falla;
     `availability_mark_resolved_at` sin `availability_mark` falla; `availability_mark` fuera del
     catálogo falla; los dos ejes **conviven** (par abierto + `availability_mark='unavailable'` es
     un estado legal — matriz C1 —) y el **recíproco** del par lo garantiza el service (prueba con
     la pareja escrita en los dos sentidos; no hay `CHECK` de simetría a propósito, C1);
   - matriz de marcas **par abierto/resuelto × enlace vivo/muerto/inconcluso** (C1/A3): los seis
     casos dejan la ficha en el estado que fija `plan.md` P9-19 (el par nunca se rompe al marcar
     disponibilidad; la indisponibilidad confirmada oculta aunque hubiera resolución previa);
   - `visibility` fuera de `visible`/`hidden` falla; `origin` fuera de `video`/`audio`/`both`
     falla; `description_en` sin `description_es` falla; textos con espacios sobrantes fallan;
   - `ListPublishedMediaItems`: **nunca** devuelve `hidden`, absorbidas ni un **par abierto**
     (`duplicate_of_id` sin resolver); sí devuelve marcas de disponibilidad sin resolver
     («ya no aparece en el feed»: FR-017) y pares **resueltos**; el orden es
     `published_at DESC, created_at ASC, id` y `LIMIT/OFFSET` de las tandas funciona (12 por
     defecto);
   - `ListPortadaMediaItems`: exactamente 5 con 8 publicables, exactos con menos;
   - `InsertMediaEpisodeSource` con `ON CONFLICT` no duplica ni actualiza (instantánea, R9-6);
   - `RepointMediaItemSources` mueve las fuentes y la ficha resultante queda con las dos;
   - cada mutación deja su fila en `admin_actions` en la misma transacción (y al forzar un fallo
     en el registro, la mutación no se aplica), con los 6 códigos y `target_label` correctos;
     inserción de `admin_actions` con los 6 códigos dentro del `CHECK` ampliado (y rechazo de un
     código fuera del registro).
3. Tipos sqlc: `pgtype.*` traducido en `repository.go` (§8.1.6); `internal/db` regenerado y
   commiteado (`make sqlc-verify`).
