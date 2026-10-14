-- 000011_extend_admin_actions_for_ministries (up)
-- F6: la auditoría de F2 se amplía a ministerios (FR-017). El registro de
-- códigos de acción sigue siendo CERRADO y lo garantiza la BD.
--
-- El `CHECK` que se sustituye es `admin_actions_action_check`, con nombre
-- propio desde la migración 000006 (research R4-10): 8 códigos de F2 + 15 de F3
-- + 15 de F4 + los 8 de F6 (R6-12). El `target_kind` NO cambia: los ministerios
-- son `content` (como los elementos de F3/F4: sin FK, `target_label`
-- identifica el elemento).

ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_action_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_action_check CHECK (action IN (
    'user.create', 'user.update', 'user.activate', 'user.deactivate', 'user.password_reset',
    'role.create', 'role.update', 'role.delete',
    'home.identity.update', 'home.about.update', 'home.contact.update',
    'home.schedule.create', 'home.schedule.update', 'home.schedule.delete',
    'home.whatsapp.create', 'home.whatsapp.update', 'home.whatsapp.delete',
    'home.social.create', 'home.social.update', 'home.social.delete',
    'home.image.upload',
    'home.publish', 'home.unpublish',
    'event.create', 'event.update', 'event.delete', 'event.publish', 'event.unpublish',
    'event.image.upload',
    'activity.create', 'activity.update', 'activity.delete', 'activity.publish',
    'activity.unpublish', 'activity.image.upload',
    'category.create', 'category.update', 'category.delete',
    'ministry.create', 'ministry.update', 'ministry.delete', 'ministry.publish',
    'ministry.unpublish', 'ministry.leader.show', 'ministry.leader.hide', 'ministry.image.upload'
));
