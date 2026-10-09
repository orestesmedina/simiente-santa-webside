/**
 * Catálogo de textos del módulo «Portada e información general» (ux.md §7.1).
 * El panel va **siempre en español** (FR-016): estas cadenas no pasan por el
 * diccionario del sitio público y no se traducen.
 */

/* ── Estados de la vista ─────────────────────────────────────────────── */

export const INFO_LOAD_ERROR = 'No se pudo cargar la información.';
export const NO_PERMISSION_ERROR =
  'No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol.';
export const SYSTEM_ERROR =
  'No se pudo completar la operación. Tus escritos están a salvo; vuelve a intentarlo en unos minutos.';

/* ── Guardado y publicación ──────────────────────────────────────────── */

export const SAVED = 'Cambios guardados.';
export const SAVED_PUBLISHED = 'Cambios guardados: ya visibles en la portada.';
export const PUBLISHED = 'Publicado en la portada.';
export const UNPUBLISHED =
  'Retirado de la portada. Estará disponible como borrador para publicarlo cuando quieras.';
export const DRAFT_NOTICE = 'Quedó en borrador: no lo verá el público hasta que lo publiques.';

export const SERVICE_CREATED = 'Servicio creado.';
export const WHATSAPP_CREATED = 'Canal de WhatsApp creado.';

/** «Enlace de {red} guardado.» (ux.md §7.1). */
export function socialSavedMessage(network: string): string {
  return `Enlace de ${network} guardado.`;
}

/* ── Confirmaciones ──────────────────────────────────────────────────── */

export const UNPUBLISH_CONFIRM_TITLE = 'Retirar de la portada';
/** «"{nombre}" dejará de ser visible al público…» (ux.md §7.1). */
export function unpublishConfirmDescription(name: string): string {
  return `"${name}" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.`;
}
export const UNPUBLISH_CONFIRM_ACTION = 'Retirar';
export const UNPUBLISH_CANCEL = 'Dejar como está';

export const DELETE_CONFIRM_TITLE = 'Eliminar';
export function deleteConfirmDescription(name: string): string {
  return `"${name}" se eliminará del panel y de la portada. Esta acción no se puede deshacer.`;
}
export const DELETE_CONFIRM_ACTION = 'Eliminar';

/* ── Validaciones (espejo de FR-015; la autoridad es el servidor) ────── */

export const REQUIRED_NAME = 'Escribe el nombre.';
export const REQUIRED_ABOUT = 'Escribe el texto de quiénes somos.';
export const REQUIRED_ADDRESS = 'Escribe la dirección.';
export const REQUIRED_EMAIL = 'Escribe el correo.';
export const REQUIRED_PHONE = 'Escribe el teléfono.';
export const REQUIRED_SERVICE_NAME = 'Escribe el nombre del servicio.';
export const REQUIRED_PLACE = 'Escribe el lugar.';
export const REQUIRED_START_TIME = 'Escribe la hora de inicio.';
export const REQUIRED_WHATSAPP_NAME = 'Escribe el nombre o propósito del canal.';
export const REQUIRED_DESTINATION = 'Escribe el destino del canal.';
export const REQUIRED_URL = 'Escribe el enlace.';
export const REQUIRED_ALT = 'Escribe el texto alternativo de la imagen.';

export const EMAIL_INVALID = 'Este correo no tiene el formato correcto.';
export const PHONE_INVALID =
  'Escribe un número de teléfono válido con código del país si corresponde.';
export const WHATSAPP_GROUP_INVALID =
  'Ese enlace no parece de WhatsApp. Pega la invitación del grupo: empieza por https://chat.whatsapp.com.';
export const URL_INVALID = 'Escribe un enlace que empiece por https://.';

/** «Ese enlace no corresponde a la red seleccionada…» (ux.md §7.1). */
export function socialInvalidMessage(network: string, example: string): string {
  return `Ese enlace no corresponde a la red seleccionada. Pega el perfil de la iglesia en ${network}, por ejemplo ${example}.`;
}

/** «Cada red admite un solo enlace…» (ux.md §7.1). */
export function socialDuplicateMessage(network: string): string {
  return `Cada red admite un solo enlace. Ya existe un enlace para ${network}: edítalo o retíralo antes de cambiarlo.`;
}

export const WHATSAPP_DUPLICATE =
  'Ya existe un canal igual (mismo nombre y mismo destino). Edita el que ya tienes o cámbiale el nombre.';

/** «{n} de 1.000 caracteres. Recorta {m} para poder guardar.» (ux.md §7.1). */
export function aboutLimitMessage(count: number): string {
  return `${count} de 1.000 caracteres. Recorta ${count - 1000} para poder guardar.`;
}

export const SPANISH_REQUIRED =
  'El contenido en español es obligatorio: es el idioma base de la portada.';
export const END_TIME_INVALID = 'La hora de fin debe ser posterior a la hora de inicio.';
export const START_TIME_INVALID = 'Escribe la hora en formato de 24 horas, por ejemplo 10:00.';

/** «Ese archivo no es una imagen válida…» (ux.md §7.1/M3). */
export function imageInvalidMessage(maxSize: string): string {
  return `Ese archivo no es una imagen válida. Usa JPG, PNG o WebP de menos de ${maxSize} y prueba de nuevo.`;
}

/* ── Ayudas de formulario ────────────────────────────────────────────── */

export const LOGO_HELP =
  'El logotipo no se deforma, no cambia de color, no se gira y no lleva efectos: sube el archivo original.';
export const COVER_HELP =
  'Aparece detrás del título en la portada. Se recorta ligeramente según la pantalla; usa una imagen horizontal.';
export const ENGLISH_HELP =
  'Si lo dejas vacío, el público verá este contenido en su versión en español.';
export const ABOUT_HELP = 'Escríbelo con tus palabras; se muestran los saltos de línea tal cual.';
export const WHATSAPP_PHONE_HELP = 'Con código de país: +506 8888 8888.';
export const WHATSAPP_GROUP_HELP =
  'Pega la invitación del grupo: empieza por https://chat.whatsapp.com.';
export const SOCIAL_URL_HELP = 'Pega el perfil de la iglesia en esa red (empieza por https://).';

/* ── Estados vacíos ──────────────────────────────────────────────────── */

export const EMPTY_IDENTITY =
  'Todavía no hay información de identidad. Complétala para que aparezca en la portada.';
export const EMPTY_ABOUT =
  'Todavía no hay texto de «quiénes somos». Escríbelo para que aparezca en la portada.';
export const EMPTY_CONTACT =
  'Todavía no hay datos de contacto. Complétalos para que aparezcan en la portada.';
export const EMPTY_SERVICES =
  'Todavía no hay servicios. Agrega el primero para que aparezca en la portada.';
export const EMPTY_WHATSAPP =
  'Todavía no hay canales. Agrega el primero para que aparezca en la portada.';
export const NO_SOCIAL_LINK = 'Sin enlace todavía';

/* ── Etiquetas de las pestañas (ux.md §4.3) ──────────────────────────── */

export const TAB_IDENTITY = 'Identidad';
export const TAB_ABOUT = 'Quiénes somos';
export const TAB_SCHEDULE = 'Horario de servicios';
export const TAB_WHATSAPP = 'WhatsApp';
export const TAB_SOCIALS = 'Redes sociales';
export const TAB_CONTACT = 'Contacto';

export const MODULE_TITLE = 'Portada e información general';
export const MODULE_HELP =
  'Lo que guardes aquí se muestra en la portada pública de la iglesia. Solo lo publicado en cada sección es visible al público.';
