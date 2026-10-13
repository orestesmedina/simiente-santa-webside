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
CREATE INDEX agenda_events_category_id_idx ON agenda_events (category_id);

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
CREATE INDEX agenda_activities_category_id_idx ON agenda_activities (category_id);

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
