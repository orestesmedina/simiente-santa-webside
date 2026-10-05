-- 000002_create_roles_and_permissions (up)
-- F2 Acceso y gestión de usuarios: catálogo de permisos, roles y su relación.
-- Cubre FR-014 (parte de datos), FR-015 (catálogo fijo de módulos) y FR-017
-- (un rol en uso no se elimina). Convenciones: skill `postgres-db` y
-- `specs/002-acceso-gestion-usuarios/data-model.md` (fuente de verdad).

CREATE TABLE permissions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code       TEXT NOT NULL UNIQUE CHECK (char_length(code) BETWEEN 1 AND 64),
    label      TEXT NOT NULL CHECK (char_length(label) BETWEEN 1 AND 120),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Catálogo fijo de FR-015: módulos del producto + administración de usuarios y roles.
-- Los `code` son los identificadores estables del catálogo (los usan el contrato y `AuthzByModule`);
-- `admin_usuarios_roles` es "el permiso de administrar usuarios y roles" de la spec (FR-015/FR-016/FR-024).
-- Los de F3–F9 quedan reservados: existen pero no dan acceso a nada hasta que se construyan.
INSERT INTO permissions (code, label) VALUES
    ('portada',             'Portada e información general'),
    ('eventos',             'Eventos'),
    ('actividades',         'Actividades'),
    ('grupos',              'Grupos de conexión'),
    ('ministerios',         'Ministerios'),
    ('donaciones',          'Donaciones'),
    ('noticias',            'Noticias y galería'),
    ('medios',              'Medios (prédicas y podcasts)'),
    ('admin_usuarios_roles','Administración de usuarios y roles');

CREATE TABLE roles (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Unicidad normalizada (Q5): solo difiere en mayúsculas = el mismo rol.
CREATE UNIQUE INDEX roles_name_lower_idx ON roles (lower(name));

CREATE TABLE role_permissions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE RESTRICT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (role_id, permission_id)
);

CREATE INDEX role_permissions_permission_id_idx ON role_permissions (permission_id);
