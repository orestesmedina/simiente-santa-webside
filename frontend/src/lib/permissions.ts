/**
 * Código del permiso de administración de usuarios y roles (FR-015/FR-024).
 * Es el único permiso de módulo que usa F2 en la navegación del panel.
 */
export const ADMIN_USERS_ROLES = 'admin_usuarios_roles';

/**
 * Código del permiso del módulo «Portada e información general» (F3, FR-012).
 * Ya existía en el catálogo de F2; F3 lo activa en la interfaz.
 */
export const PORTADA = 'portada';

/**
 * Permisos de los módulos de agenda (F4, FR-012): **independientes entre sí**.
 * Tener «eventos» no da acceso a las actividades ni al catálogo, y el catálogo
 * se autoriza con su permiso propio `categorias` (10.º del catálogo de F2, ⟲ D-6).
 */
export const EVENTOS = 'eventos';
export const ACTIVIDADES = 'actividades';
export const CATEGORIAS = 'categorias';

/**
 * Permiso del módulo «Ministerios» (F6, FR-011). Ya existía en el catálogo de
 * F2 (`ministerios`); F6 lo **activa** en la interfaz: no se crea ningún permiso
 * nuevo y el catálogo del backend sigue en 10.
 */
export const MINISTERIOS = 'ministerios';

/**
 * Permiso del módulo «Donaciones» (F7, FR-011). Ya existía en el catálogo de
 * F2 (`donaciones`); F7 lo **activa** en la interfaz (el nombre **no** cambia,
 * Q8: el módulo del panel y el permiso se llaman «donaciones», la sección
 * pública es «Ofrendas y diezmos»). No se crea ningún permiso nuevo y el
 * catálogo del backend sigue en 10.
 */
export const DONACIONES = 'donaciones';

/**
 * Permisos de módulo construidos hasta hoy (F2, F3, F4, F6 y F7). Lo usa el
 * inicio del panel para decidir si hay algún acceso rápido que ofrecer y la
 * lista de roles para saber qué módulos existen de verdad (FR-015 de F2/FR-012
 * de F4).
 */
export const MODULE_PERMISSIONS = [
  ADMIN_USERS_ROLES,
  PORTADA,
  EVENTOS,
  ACTIVIDADES,
  CATEGORIAS,
  MINISTERIOS,
  DONACIONES,
] as const;

/** Forma mínima de sesión que necesita la comprobación de permisos. */
export interface SessionLike {
  permissions: readonly string[];
}

/**
 * Comprueba si la sesión tiene un permiso del catálogo. Sin sesión, con
 * `permissions` no disponible o con la lista vacía devuelve `false` (Edge Case
 * válido: una cuenta sin permisos de módulo ve solo Inicio). La autoridad real
 * sigue siendo el servidor; esto solo decide la interfaz (FR-016).
 */
export function hasPermission(session: SessionLike | null | undefined, code: string): boolean {
  if (!session || !Array.isArray(session.permissions)) {
    return false;
  }
  return session.permissions.includes(code);
}

/** `true` si la sesión tiene al menos uno de los permisos indicados. */
export function hasAnyPermission(
  session: SessionLike | null | undefined,
  codes: readonly string[],
): boolean {
  return codes.some((code) => hasPermission(session, code));
}
