-- 000004_create_login_events_and_admin_actions (up)
-- Cambio de alcance: auditoría (US8, FR-021…FR-026). Añade la proyección del
-- último acceso en `users` y las dos tablas duraderas del registro. Se añade
-- como migración propia (no editando 000003) para aislar el cambio de alcance;
-- la numeración queda sin huecos: 000001 (F1) → 000002 → 000003 → 000004.
-- Convenciones: skill `postgres-db` y `specs/002-acceso-gestion-usuarios/data-model.md`.

-- Proyección del último acceso exitoso (FR-021): dos columnas en users para que la ficha y el
-- listado la muestren sin consultar el historial (ver "Último acceso por cuenta" más abajo).
-- Solo las escribe el login exitoso, en el mismo paso que inserta su fila en login_events.
ALTER TABLE users
    ADD COLUMN last_login_at TIMESTAMPTZ NULL,
    ADD COLUMN last_login_ip TEXT NULL CHECK (char_length(last_login_ip) BETWEEN 3 AND 45);

-- Historial de intentos de inicio de sesión (FR-022): una fila por intento, exitoso o fallido.
-- user_id queda NULL cuando el correo no corresponde a ninguna cuenta: el intento se registra
-- igual pero no se asocia a nada y no se crea ninguna cuenta (FR-003, sin "cuentas fantasma").
CREATE TABLE login_events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NULL REFERENCES users(id) ON DELETE RESTRICT,
    result     TEXT NOT NULL CHECK (result IN ('success', 'failure')),
    ip         TEXT NOT NULL CHECK (char_length(ip) BETWEEN 3 AND 45),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Orden por defecto del historial (§8.1.7) y filtro por cuenta + rango de fechas (FR-024).
CREATE INDEX login_events_created_at_id_idx ON login_events (created_at DESC, id DESC);
CREATE INDEX login_events_user_created_at_id_idx ON login_events (user_id, created_at DESC, id DESC);

-- Historial de acciones administrativas sensibles (FR-023): quién, qué, sobre qué, cuándo y con
-- qué resultado —incluidos los intentos que no se completan y los denegados por falta de permiso.
CREATE TABLE admin_actions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id  UUID NULL REFERENCES users(id) ON DELETE RESTRICT,
    action         TEXT NOT NULL CHECK (action IN (
                       'user.create', 'user.update', 'user.activate', 'user.deactivate',
                       'user.password_reset', 'role.create', 'role.update', 'role.delete')),
    target_kind    TEXT NOT NULL CHECK (target_kind IN ('user', 'role')),
    target_user_id UUID NULL REFERENCES users(id) ON DELETE RESTRICT,
    target_role_id UUID NULL REFERENCES roles(id) ON DELETE SET NULL,
    target_label   TEXT NULL CHECK (char_length(target_label) BETWEEN 1 AND 254),
    result         TEXT NOT NULL CHECK (result IN ('success', 'failure', 'denied')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Sin actor solo puede quedar la inicialización del sistema (FR-007), que cuenta como
    -- creación de cuenta y por eso se registra como 'user.create'.
    CHECK (actor_user_id IS NOT NULL OR action = 'user.create'),
    -- Cada objetivo usa la FK de su tipo. Ambas son anulables: una creación rechazada no llega a
    -- tener id, y un rol eliminado (FR-017) deja target_role_id en NULL conservando target_label.
    CHECK ((target_kind = 'user' AND target_role_id IS NULL) OR
           (target_kind = 'role' AND target_user_id IS NULL))
);

CREATE INDEX admin_actions_created_at_id_idx ON admin_actions (created_at DESC, id DESC);
CREATE INDEX admin_actions_actor_created_at_id_idx ON admin_actions (actor_user_id, created_at DESC, id DESC);
CREATE INDEX admin_actions_target_user_created_at_id_idx ON admin_actions (target_user_id, created_at DESC, id DESC);
CREATE INDEX admin_actions_target_role_id_idx ON admin_actions (target_role_id);
