/**
 * Textos de la gestión de cuentas (ux.md §7). Viven en un solo sitio para que
 * listado, formulario y diálogos compartan los mismos mensajes en español.
 */

export const SYSTEM_ERROR =
  'No se pudo completar la operación. Vuelve a intentarlo en unos minutos; si sigue, avísanos.';

export const SESSION_EXPIRED_ERROR =
  'Tu sesión terminó. Entra de nuevo y vuelve a hacer la acción; no se guardó a medias.';

/** Último administrador activo (FR-008): desactivar o degradar sin reemplazo. */
export const LAST_ADMIN_ERROR =
  'No puedes dejar el panel sin un administrador activo. Activa otra cuenta con permiso de administración primero.';

/** Correo duplicado (US3 esc. 2, Q5/SC-011). Fallback si el servidor no da texto. */
export const EMAIL_IN_USE_ERROR = 'Ya existe una cuenta con ese correo. Prueba otro correo.';

/** Rol inexistente al guardar (US3 esc. 5). */
export const ROLE_NOT_FOUND_ERROR = 'El rol elegido ya no existe; elige otro.';

export const PHONE_INVALID_ERROR = 'Escribe un número de teléfono válido.';

export const LOAD_USERS_ERROR = 'No se pudo cargar el listado.';
export const EMPTY_USERS_MESSAGE = 'Aún no hay cuentas. Crea la primera para tu equipo.';

export const CUENTA_CREADA = 'Cuenta creada.';
export const CAMBIOS_GUARDADOS = 'Cambios guardados.';
export const CUENTA_ACTIVADA = 'Cuenta activada.';
export const CUENTA_DESACTIVADA = 'Cuenta desactivada.';
export const PASSWORD_RESET = 'Contraseña restablecida.';

/** Confirmaciones (ux.md §7): explican que los datos se conservan. */
export const DEACTIVATE_CONFIRM_DESCRIPTION =
  'La persona perderá el acceso de inmediato, aunque tenga sesión abierta; sus datos se conservan.';
export const ACTIVATE_CONFIRM_DESCRIPTION =
  'La persona podrá volver a entrar al panel; todos sus datos se conservan tal y como están.';
