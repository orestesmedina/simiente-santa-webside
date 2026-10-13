-- 000009_seed_permission_categorias (down): retira la fila del permiso.
-- role_permissions.permission_id es ON DELETE RESTRICT (000002): primero se
-- retiran las asignaciones de ESTE permiso (solo las suyas; el resto de los
-- permisos de cada rol no se toca) y luego la fila. Destructivo a propósito,
-- como el down de 000006/000008: solo tiene sentido en desarrollo/rollback.
DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE code = 'categorias');
DELETE FROM permissions WHERE code = 'categorias';
