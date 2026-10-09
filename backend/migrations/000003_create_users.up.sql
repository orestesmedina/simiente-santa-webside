-- 000003_create_users (up)
-- Cuenta del panel (FR-009…FR-013, FR-017, FR-019). Un solo rol por cuenta
-- (`users.role_id`, Q4); un rol en uso no se puede eliminar (`ON DELETE
-- RESTRICT`, FR-017). Sin borrado de cuentas (FR-013): no existe `DELETE`,
-- retirar acceso es `is_active = false`. Convenciones: skill `postgres-db` y
-- `specs/002-acceso-gestion-usuarios/data-model.md`.

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
