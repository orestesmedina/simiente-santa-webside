import type { PermissionItem } from '../../api/roles';
import { ADMIN_USERS_ROLES } from '../../lib/permissions';

/**
 * Sello informativo (nunca error) de los módulos de F3–F9 aún no construidos
 * (ux.md §3.7/§4.d). Se muestra en el formulario y en el resumen del listado.
 */
export const PENDING_PERMISSION_NOTE = 'Disponible más adelante';

/**
 * `true` solo para los permisos de módulos que existen hoy. En F2 el único
 * módulo construido es la administración de usuarios y roles; el resto del
 * catálogo (F3–F9) se puede asignar a un rol pero no da acceso a nada todavía
 * (FR-015: es válido y esperado).
 */
export function isPermissionAvailable(code: string): boolean {
  return code === ADMIN_USERS_ROLES;
}

/** Etiqueta del catálogo para un código; cae al propio código si no está. */
export function permissionLabel(
  code: string,
  catalog: readonly PermissionItem[] | undefined,
): string {
  return catalog?.find((item) => item.code === code)?.label ?? code;
}
