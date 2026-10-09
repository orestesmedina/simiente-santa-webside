-- 000006_extend_admin_actions_for_home_content (down).
-- DESTRUCTIVO a propósito (igual que el down de 000004, que tira el registro):
-- las filas con objetivo 'content' no caben en el CHECK anterior, así que se
-- eliminan antes de restaurar. Solo tiene sentido en desarrollo/rollback.
DELETE FROM admin_actions WHERE target_kind = 'content';

ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_target_coherence_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_check1 CHECK (
    (target_kind = 'user' AND target_role_id IS NULL) OR
    (target_kind = 'role' AND target_user_id IS NULL)
);

ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_target_kind_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_target_kind_check
    CHECK (target_kind IN ('user', 'role'));

ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_action_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_action_check CHECK (action IN (
    'user.create', 'user.update', 'user.activate', 'user.deactivate', 'user.password_reset',
    'role.create', 'role.update', 'role.delete'
));
