-- 000008_extend_admin_actions_for_agenda (up)
-- F4: la auditoría de F2 se amplía a eventos, actividades y categorías (FR-017).
-- El registro de códigos de acción sigue siendo CERRADO y lo garantiza la BD.
--
-- El `CHECK` que se sustituye es `admin_actions_action_check`, con nombre
-- propio desde la migración 000006 (research R4-10): 8 códigos de F2 + 15 de F3
-- + los 15 de F4 (6 event.*, 6 activity.* y los 3 category.* del catálogo de
-- categorías). El `target_kind` NO cambia: los elementos de F4 son `content`
-- (como los de F3: sin FK, `target_label` identifica el elemento).

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
    'category.create', 'category.update', 'category.delete'
));
