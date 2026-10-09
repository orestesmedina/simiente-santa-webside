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
