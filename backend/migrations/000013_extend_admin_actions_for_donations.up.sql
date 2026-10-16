-- 000013_extend_admin_actions_for_donations (up)
-- F7: la auditoría de F2 se amplía a ofrendas y diezmos (FR-017). El registro de
-- códigos de acción sigue siendo CERRADO y lo garantiza la BD.
--
-- El `CHECK` que se sustituye es `admin_actions_action_check`, con nombre propio
-- desde la migración 000006 (F3, F4 y F6 ya hicieron este mismo ejercicio). El
-- `target_kind` NO cambia: los métodos y los textos son `content` (sin FK,
-- `target_label` identifica el elemento). Los 8 códigos nuevos son los de R7-10.

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
    'ministry.unpublish', 'ministry.leader.show', 'ministry.leader.hide', 'ministry.image.upload',
    'donation.create', 'donation.update', 'donation.delete', 'donation.publish',
    'donation.unpublish', 'donation.text.update', 'donation.text.publish', 'donation.text.unpublish'
));
