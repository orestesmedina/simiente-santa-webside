/**
 * Textos de la gestión de roles (ux.md §4.d/§7). Viven en un solo sitio para que
 * listado, formulario y diálogos compartan los mismos mensajes en español.
 */

export const SYSTEM_ERROR =
  'No se pudo completar la operación. Vuelve a intentarlo en unos minutos; si sigue, avísanos.';

export const SESSION_EXPIRED_ERROR =
  'Tu sesión terminó. Entra de nuevo y vuelve a hacer la acción; no se guardó a medias.';

export const LOAD_ROLES_ERROR = 'No se pudo cargar el listado.';
export const EMPTY_ROLES_MESSAGE =
  'Aún no hay roles. Crea el primero para organizar los accesos de tu equipo.';

export const ROLE_CREATED = 'Rol creado.';
export const CHANGES_SAVED = 'Cambios guardados.';
export const ROLE_DELETED = 'Rol eliminado.';

/** Nombre de rol duplicado (normalizado, Q5/SC-011). Fallback si el servidor no da texto. */
export const ROLE_NAME_IN_USE = 'Ya existe un rol con ese nombre. Elige otro nombre.';

/** Al menos un permiso por rol (FR-014/US4 esc. 4). */
export const NO_PERMISSIONS_ERROR = 'Un rol necesita al menos un permiso. Marca al menos uno.';

/** Confirmación destructiva de eliminar un rol sin uso (ux.md §7). */
export function deleteConfirmDescription(name: string, permissionCount: number): string {
  return `Se borrará el rol '${name}' con sus ${permissionCount} permisos. Las cuentas que lo usen no se verán afectadas si antes les asignas otro.`;
}

/** Rol en uso: se impide eliminar y se pide reasignar primero (FR-017/US6 esc. 4–5). */
export function roleInUseMessage(userCount: number): string {
  return `No se puede eliminar: ${userCount} cuentas usan este rol. Primero cámbiales el rol y después podrás eliminarlo.`;
}
