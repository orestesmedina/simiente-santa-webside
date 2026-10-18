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

/** Etiquetas en español de las acciones administrativas (FR-023, ux.md §4.9). */
export const ACTION_LABELS: Record<string, string> = {
  'user.create': 'Creó una cuenta',
  'user.update': 'Editó una cuenta',
  'user.activate': 'Activó una cuenta',
  'user.deactivate': 'Desactivó una cuenta',
  'user.password_reset': 'Restableció la contraseña',
  'role.create': 'Creó un rol',
  'role.update': 'Editó un rol',
  'role.delete': 'Eliminó un rol',
  // F4 (migración `000008`, R4-10): 15 códigos de la agenda, para que el panel
  // no muestre códigos crudos. Los `home.*` de F3 siguen sin etiqueta (deuda
  // heredada, fuera del alcance de F4).
  'event.create': 'Creó un evento',
  'event.update': 'Modificó un evento',
  'event.delete': 'Eliminó un evento',
  'event.publish': 'Publicó un evento',
  'event.unpublish': 'Retiró un evento',
  'event.image.upload': 'Subió la imagen de un evento',
  'activity.create': 'Creó una actividad',
  'activity.update': 'Modificó una actividad',
  'activity.delete': 'Eliminó una actividad',
  'activity.publish': 'Publicó una actividad',
  'activity.unpublish': 'Retiró una actividad',
  'activity.image.upload': 'Subió la imagen de una actividad',
  'category.create': 'Creó una categoría',
  'category.update': 'Modificó una categoría',
  'category.delete': 'Eliminó una categoría',
  // F6 (migración `000011`, R6-12): 8 códigos del módulo de ministerios, para
  // que el panel no muestre códigos crudos (ux.md §4.8/§7.1).
  'ministry.create': 'Creó el ministerio',
  'ministry.update': 'Modificó el ministerio',
  'ministry.delete': 'Eliminó el ministerio',
  'ministry.publish': 'Publicó el ministerio',
  'ministry.unpublish': 'Retiró el ministerio',
  'ministry.leader.show': 'Publicó la ficha del encargado del ministerio',
  'ministry.leader.hide': 'Ocultó la ficha del encargado del ministerio',
  'ministry.image.upload': 'Subió la imagen de un ministerio',
  // F7 (migración `000013`, R7-10): 8 códigos del módulo de donaciones, para
  // que el panel no muestre códigos crudos (ux.md §4.8/§7.1).
  'donation.create': 'Creó una cuenta o número de donaciones',
  'donation.update': 'Modificó una cuenta o número de donaciones',
  'donation.delete': 'Eliminó una cuenta o número de donaciones',
  'donation.publish': 'Publicó una cuenta o número de donaciones',
  'donation.unpublish': 'Retiró una cuenta o número de donaciones',
  'donation.text.update': 'Modificó el texto introductorio o el de uso',
  'donation.text.publish': 'Publicó el texto introductorio o el de uso',
  'donation.text.unpublish': 'Retiró el texto introductorio o el de uso',
  // F9 (migración `000015`, P9-14): 6 códigos del módulo de medios, para que el
  // panel no muestre códigos crudos (ux.md §7.1). F9 no tiene `create`/`delete`:
  // no hay alta manual ni borrado (FR-019).
  'media.sync': 'Buscó episodios nuevos',
  'media.description.update': 'Modificó la descripción de un episodio',
  'media.hide': 'Ocultó un episodio',
  'media.show': 'Mostró un episodio',
  'media.merge': 'Unió dos entradas en una sola ficha',
  'media.mark.resolve': 'Resolvió la marca de un episodio',
};

export function actionLabel(action: string): string {
  return ACTION_LABELS[action] ?? action;
}
