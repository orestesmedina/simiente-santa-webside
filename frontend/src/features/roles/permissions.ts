import type { PermissionItem } from '../../api/roles';
import { MODULE_PERMISSIONS } from '../../lib/permissions';

/**
 * Sello informativo (nunca error) de los módulos aún no construidos —hoy F5 y
 * F8— (ux.md §3.7/§4.d). Se muestra en el formulario y en el resumen del listado.
 */
export const PENDING_PERMISSION_NOTE = 'Disponible más adelante';

/**
 * `true` solo para los permisos de módulos que existen hoy. En F2 el único
 * módulo construido era la administración de usuarios y roles; F3 sumó
 * «Portada e información general» (`portada`, FR-012), F4 suma los tres de la
 * agenda (`eventos`, `actividades` y `categorias`, ⟲ D-6), F6 suma
 * «Ministerios» (`ministerios`, FR-011), F7 suma «Donaciones» (`donaciones`,
 * FR-011) y F9 suma «Medios» (`medios`, FR-021). Los permisos de F6, F7 y F9 ya
 * estaban sembrados por F2: se ofrecen **sin sello**. El resto del catálogo (F5
 * y F8) se puede asignar a un rol pero no da acceso a nada todavía (FR-015: es
 * válido y esperado).
 */
export function isPermissionAvailable(code: string): boolean {
  return (MODULE_PERMISSIONS as readonly string[]).includes(code);
}

/** Etiqueta del catálogo para un código; cae al propio código si no está. */
export function permissionLabel(
  code: string,
  catalog: readonly PermissionItem[] | undefined,
): string {
  return catalog?.find((item) => item.code === code)?.label ?? code;
}
