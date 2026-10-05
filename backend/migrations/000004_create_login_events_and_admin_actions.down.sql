-- 000004_create_login_events_and_admin_actions (down)
-- Revierte la migración por completo, incluida la proyección del último acceso.

DROP TABLE IF EXISTS admin_actions;
DROP TABLE IF EXISTS login_events;
ALTER TABLE users
    DROP COLUMN IF EXISTS last_login_ip,
    DROP COLUMN IF EXISTS last_login_at;
