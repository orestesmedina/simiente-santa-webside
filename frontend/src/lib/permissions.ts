/**
 * Código del permiso de administración de usuarios y roles (FR-015/FR-024).
 * Es el único permiso de módulo que usa F2 en la navegación del panel.
 */
export const ADMIN_USERS_ROLES = 'admin_usuarios_roles';

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
