-- 000015_extend_admin_actions_for_media (down): restaura el CHECK de F7
-- (8 códigos de F2 + 15 de F3 + 15 de F4 + 8 de F6 + 8 de F7). Como en el down
-- de 000006/000008/000011/000013, es destructivo a propósito: las filas con
-- códigos F9 no caben en el CHECK anterior y se eliminan antes de restaurar.
-- Solo tiene sentido en desarrollo/rollback.
DELETE FROM admin_actions WHERE action IN (
    'media.sync', 'media.description.update', 'media.hide', 'media.show', 'media.merge',
    'media.mark.resolve'
);

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
