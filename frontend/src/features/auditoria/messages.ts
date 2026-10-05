/**
 * Textos de la Auditoría (ux.md §3.10/§4.g/§7). Es una vista de solo lectura:
 * no hay mensajes de acción en curso ni confirmaciones (FR-025).
 */

export const LOAD_AUDIT_ERROR =
  'No se pudo cargar el registro. Vuelve a intentarlo en unos minutos; si sigue, avísanos.';

/** Mismo texto que la página `/sin-permiso` cuando el servidor responde `403`. */
export const NO_PERMISSION_ERROR =
  'No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol.';

/** Cuenta que intentó entrar y no se pudo identificar (F-01, sin correo alguno). */
export const UNIDENTIFIED_ATTEMPT = 'Intento sin cuenta asociada';

/** Estado vacío (ux.md §4.g): con y sin filtros aplicados. */
export function emptyAuditMessage(hasFilters: boolean): string {
  return hasFilters
    ? 'Todavía no hay registros que coincidan con esos filtros. Prueba a quitar los filtros o usar un rango de fechas más amplio.'
    : 'Todavía no hay registros de actividad. Aparecerán cuando alguien entre al panel o se gestione una cuenta o un rol.';
}

/** Etiquetas en español de las acciones administrativas (FR-023). */
export const ACTION_LABELS: Record<string, string> = {
  'user.create': 'Creó una cuenta',
  'user.update': 'Editó una cuenta',
  'user.activate': 'Activó una cuenta',
  'user.deactivate': 'Desactivó una cuenta',
  'user.password_reset': 'Restableció la contraseña',
  'role.create': 'Creó un rol',
  'role.update': 'Editó un rol',
  'role.delete': 'Eliminó un rol',
};

export function actionLabel(action: string): string {
  return ACTION_LABELS[action] ?? action;
}
