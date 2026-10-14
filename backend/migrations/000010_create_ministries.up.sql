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
    CHECK (whatsapp_url IS NULL OR whatsapp_url = btrim(whatsapp_url)),
    -- Normalización de F2 (P13): el dato normalizado es el que se guarda.
    -- Precedente literal de 000005 (home_contact: CHECK (email = btrim(email))
    -- y CHECK (phone = btrim(phone))); aquí leader_phone/leader_email son
    -- anulables y por eso van con IS NULL OR (A6-9).
    CHECK (leader_phone IS NULL OR leader_phone = btrim(leader_phone)),
    CHECK (leader_email IS NULL OR leader_email = btrim(leader_email))
);

-- Orden de presentación (Q5/FR-004): sort_order, id, como home_services y
-- agenda_categories.
CREATE INDEX ministries_sort_order_id_idx ON ministries (sort_order, id);

-- WHERE publication_state en la lectura pública y en el filtro del panel.
CREATE INDEX ministries_publication_state_idx ON ministries (publication_state);
