-- 000006_extend_admin_actions_for_home_content (up)
-- F3: la auditoría de F2 se amplía a la portada (FR-017). El registro de
-- códigos de acción y de tipos de objetivo sigue siendo CERRADO y lo garantiza
-- la BD (research R3-11).

-- 1) Códigos de acción: los 8 de F2 + los 15 de F3 (home.*; incluye
--    'home.image.upload' por la subida de imágenes — analyze I8).
ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_action_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_action_check CHECK (action IN (
    'user.create', 'user.update', 'user.activate', 'user.deactivate', 'user.password_reset',
    'role.create', 'role.update', 'role.delete',
    'home.identity.update', 'home.about.update', 'home.contact.update',
    'home.schedule.create', 'home.schedule.update', 'home.schedule.delete',
    'home.whatsapp.create', 'home.whatsapp.update', 'home.whatsapp.delete',
    'home.social.create', 'home.social.update', 'home.social.delete',
    'home.image.upload',
    'home.publish', 'home.unpublish'
));

-- 2) Tipo de objetivo: + 'content' (elemento de la portada; sin FK: son 6 tablas).
ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_target_kind_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_target_kind_check
    CHECK (target_kind IN ('user', 'role', 'content'));

-- 3) Coherencia de objetivo: 'content' exige ambas FK en NULL y target_label.
ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_check1;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_target_coherence_check CHECK (
    (target_kind = 'user'    AND target_role_id IS NULL) OR
    (target_kind = 'role'    AND target_user_id IS NULL) OR
    (target_kind = 'content' AND target_user_id IS NULL AND target_role_id IS NULL
                              AND target_label IS NOT NULL)
);
