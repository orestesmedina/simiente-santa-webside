-- 000014_create_media_content (up)
-- F9 Medios: una ficha por sermón (FR-001) con sus fuentes de importación —una
-- fila por elemento de feed, UNIQUE (provider, external_id): idempotencia de la
-- búsqueda, R9-11— y el registro de cada «Buscar episodios nuevos» (FR-018).
-- Estado = visibility 'visible'/'hidden' (Consecuencia 5: F9 no usa
-- borrador/publicado). Convenciones: skill `postgres-db` y
-- specs/009-medios-predicas-y-podcasts/data-model.md (fuente de verdad).
-- Decisiones: research R9-4…R9-12.
--
-- Nota de transcripción: en `data-model.md` el bloque de columnas
-- `availability_checked_at` / `availability_result` está escrito dos veces
-- (repetido idéntico dentro de `media_items`); aquí se declara UNA sola vez, en
-- la posición del primero, porque PostgreSQL no admite dos columnas con el
-- mismo nombre. El resto es literal al modelo.

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
    duplicate_of_id            UUID NULL REFERENCES media_items(id) ON DELETE RESTRICT,
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
    merged_into_id    UUID NULL REFERENCES media_items(id) ON DELETE RESTRICT,

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
    item_id      UUID NOT NULL REFERENCES media_items(id) ON DELETE RESTRICT,
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
    actor_user_id               UUID NULL REFERENCES users(id) ON DELETE RESTRICT,
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
