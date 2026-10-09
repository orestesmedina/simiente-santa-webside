-- 000002_create_roles_and_permissions (down)
-- Revierte la migración por completo, en orden inverso a la creación.

DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS permissions;
