-- 000015_extend_admin_actions_for_media (up)
-- F9: la auditoría de F2 se amplía a medios (FR-022). El registro de códigos de
-- acción sigue siendo CERRADO y lo garantiza la BD.
--
-- El `CHECK` que se sustituye es `admin_actions_action_check`, con nombre propio
-- desde la migración 000006 (F3, F4, F6 y F7 ya hicieron este mismo ejercicio). El
-- `target_kind` NO cambia: las fichas de F9 son `content` (sin FK, `target_label`
-- identifica el elemento: «Medios · <title>»). Los 6 códigos nuevos son los de
-- R9-12/P9-14: F9 no tiene `create`/`delete` porque no hay alta manual ni borrado
-- de episodios (FR-019).

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
    'donation.unpublish', 'donation.text.update', 'donation.text.publish', 'donation.text.unpublish',
    'media.sync', 'media.description.update', 'media.hide', 'media.show', 'media.merge',
    'media.mark.resolve'
));
