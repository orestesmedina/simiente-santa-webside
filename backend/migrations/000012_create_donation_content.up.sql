-- 000012_create_donation_content (up)
-- F7 Ofrendas y diezmos: el listado administrable de métodos de donación
-- (cuentas IBAN y números de SINPE Móvil, agrupados por tipo con orden manual
-- dentro de cada grupo — Q1) y los dos textos libres de la sección, uno por
-- clave, cada uno con su propio estado de publicación (Q5/Q10/Q11).
-- Convenciones: skill `postgres-db` y specs/007-donaciones/data-model.md
-- (fuente de verdad). Decisiones: research R7-1…R7-5.

CREATE TABLE donation_methods (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Tipo de método (Q1/FR-019): 'iban' = cuenta bancaria, 'sinpe' = SINPE Móvil.
    kind              TEXT NOT NULL CHECK (kind IN ('iban', 'sinpe')),
    -- Número ya NORMALIZADO (research R7-2/R7-3): el dato normalizado es el que
    -- se guarda. IBAN = solo los caracteres del número, en mayúsculas
    -- (sin espacios ni guiones; longitud exacta por país la valida el service).
    -- SINPE = '+506' + 8 dígitos. El CHECK por tipo refuerza la forma aunque el
    -- service ya la valida (defensa en profundidad).
    number            TEXT NOT NULL CHECK (char_length(number) BETWEEN 1 AND 34),
    -- Entidad financiera (nombre del banco), obligatoria (Q2). NO traducible
    -- (FR-007/US4 esc. 5/SC-006; research R7-5 señala la contradicción de FR-019).
    institution       TEXT NOT NULL CHECK (char_length(institution) BETWEEN 1 AND 160),
    -- Titular de la cuenta, OPCIONAL (Q2: "si se rellena, se publica") y único
    -- campo traducible del método (FR-007).
    holder_es         TEXT NULL CHECK (char_length(holder_es) BETWEEN 1 AND 160),
    holder_en         TEXT NULL CHECK (char_length(holder_en) BETWEEN 1 AND 160),
    -- Moneda SOLO en cuentas IBAN (Q2/FR-014): obligatoria en IBAN y prohibida
    -- en SINPE (los dos CHECK de abajo). Código CRC/USD; la etiqueta visible la
    -- localiza el frontend (patrón day_of_week de 000005; no es contenido
    -- traducible, SC-006).
    currency          TEXT NULL CHECK (currency IN ('CRC', 'USD')),
    publication_state TEXT NOT NULL DEFAULT 'draft'
                      CHECK (publication_state IN ('draft', 'published')),
    -- Orden manual DENTRO del grupo de tipo (Q1/FR-004); el orden entre grupos
    -- lo fija el CASE del ORDER BY (iban primero, sinpe después).
    sort_order        INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Formato exacto por tipo (el dato normalizado es el que se guarda):
    --   IBAN  → 2 letras de país + 2 dígitos de control + BBAN alfanumérico
    --            (la longitud exacta por país la valida el service: R7-2)
    --   SINPE → '+506' + 8 dígitos (celular costarricense: R7-3)
    CHECK ((kind = 'iban'  AND number ~ '^[A-Z]{2}[0-9]{2}[A-Z0-9]{10,30}$')
        OR (kind = 'sinpe' AND number ~ '^\+506[0-9]{8}$')),
    -- Moneda obligatoria en IBAN y prohibida en SINPE (Q2/FR-014).
    CHECK ((kind = 'iban' AND currency IS NOT NULL)
        OR (kind = 'sinpe' AND currency IS NULL)),
    -- Español base por campo: holder_en exige holder_es (FR-007/Decisión 8).
    CHECK (holder_en IS NULL OR holder_es IS NOT NULL),
    -- Normalización de F2 (P13): el dato guardado ya viene recortado
    -- (precedente literal de 000010; los anulables con IS NULL OR … = btrim(…)).
    CHECK (institution = btrim(institution)),
    CHECK (holder_es IS NULL OR holder_es = btrim(holder_es)),
    CHECK (holder_en IS NULL OR holder_en = btrim(holder_en))
);

-- Orden de presentación dentro del grupo y filtro público por estado.
CREATE INDEX donation_methods_kind_sort_order_id_idx
    ON donation_methods (kind, sort_order, id);
CREATE INDEX donation_methods_publication_state_idx
    ON donation_methods (publication_state);

-- Los dos textos libres de la sección (FR-021, Q5/Q11): un renglón por clave
-- ('intro' = «cómo donar», 'usage' = «en qué utilizamos esas donaciones»),
-- con su PROPIO estado de publicación (Q10). Es el patrón "singleton por clave":
-- UNIQUE (text_key) garantiza ≤1 fila por clave y las filas se crean en el
-- primer guardado (upsert del PUT), como los singletons de 000005; la migración
-- NO siembra contenido.
CREATE TABLE donation_texts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text_key          TEXT NOT NULL UNIQUE CHECK (text_key IN ('intro', 'usage')),
    body_es           TEXT NOT NULL CHECK (char_length(body_es) BETWEEN 1 AND 2000),
    body_en           TEXT NULL CHECK (char_length(body_en) BETWEEN 1 AND 2000),
    publication_state TEXT NOT NULL DEFAULT 'draft'
                      CHECK (publication_state IN ('draft', 'published')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Español base por campo (FR-007/Decisión 8): body_en exige body_es
    -- (redundante con body_es NOT NULL; se deja como documentación viva de la
    -- regla, igual que los CHECK de español por campo de 000010).
    CHECK (body_en IS NULL OR body_es IS NOT NULL)
);
