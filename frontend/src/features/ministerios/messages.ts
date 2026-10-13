/**
 * Catálogo de textos del módulo «Ministerios» del panel (ux.md §7.1). El panel
 * va **siempre en español** (FR-018): estas cadenas no pasan por el diccionario
 * del sitio público y no se traducen.
 */

/* ── Límites del contrato (espejo de las constantes del service, FR-019) ── */

export const NAME_MAX = 160;
export const DESCRIPTION_MAX = 2000;
export const SCHEDULE_MAX = 200;
export const PLACE_MAX = 200;
export const LEADER_NAME_MAX = 160;
export const PHONE_MAX = 32;
export const EMAIL_MAX = 254;
export const WHATSAPP_URL_MAX = 500;
export const IMAGE_ALT_MAX = 300;

/* ── Encabezado y estados de la vista ────────────────────────────────── */

export const TITLE = 'Gestión de ministerios';
export const HELP =
  'Aquí administras las áreas de servicio de la iglesia (p. ej., alabanza, sonido, acción social). Solo lo publicado es visible al público.';

export const LOAD_ERROR = 'No se pudo cargar los ministerios.';
export const NO_PERMISSION_ERROR =
  'No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol.';
export const SYSTEM_ERROR =
  'No se pudo completar la operación. Tus escritos están a salvo; vuelve a intentarlo en unos minutos.';
export const RETRY = 'Reintentar';

/* ── Guardado y publicación ──────────────────────────────────────────── */

export const SAVED = 'Cambios guardados.';
export const SAVED_PUBLISHED = 'Cambios guardados: ya visibles en la sección.';
export const CREATED = 'Ministerio creado.';
export const DRAFT_NOTICE = 'Quedó en borrador: no lo verá el público hasta que lo publiques.';
export const PUBLISHED = 'Publicado en "Ministerios".';
export const UNPUBLISHED =
  'Retirado de la sección. Estará disponible como borrador para publicarlo cuando quieras.';
export const SAVED_LEADER_HIDDEN =
  'Cambios guardados. La ficha del encargado quedó oculta: el público no verá su nombre ni su contacto.';
export const SAVED_LEADER_VISIBLE = 'Cambios guardados. La ficha del encargado está publicada.';
export const DELETED = 'Eliminado.';
export const SAVE = 'Guardar cambios';

/* ── Acciones y confirmaciones ───────────────────────────────────────── */

export const ADD = 'Agregar ministerio';
export const EDIT_ACTION = 'Editar';
export const DELETE_ACTION = 'Eliminar';

export const RETIRE_ACTION = 'Retirar de la sección';
export const RETIRE_CONFIRM_TITLE = 'Retirar de la sección';
export const RETIRE_CONFIRM_ACTION = 'Retirar';
export const RETIRE_CANCEL = 'Dejar como está';

export const DELETE_CONFIRM_TITLE = 'Eliminar';
export const DELETE_CONFIRM_ACTION = 'Eliminar';
export const DELETE_CANCEL = 'Cancelar';

/** «"{nombre}" dejará de ser visible al público…» (ux.md §5.c). */
export function retireConfirmDescription(name: string): string {
  return `"${name}" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.`;
}

/** «"{nombre}" se eliminará para siempre. Esto no se puede deshacer.» (ux.md §7.1). */
export function deleteConfirmDescription(name: string): string {
  return `"${name}" se eliminará para siempre. Esto no se puede deshacer.`;
}

/* ── Marcas y avisos del listado ─────────────────────────────────────── */

export const HIDDEN_LEADER_LABEL = 'Ficha oculta';
export const HIDDEN_LEADER_LEGEND =
  'Su ficha no es pública: el visitante no ve su nombre ni su contacto. Los datos quedan guardados aquí.';
export const EMPTY_DASH = '—';
export const NOTHING_PUBLISHED =
  'Nada está publicado ahora: el público no ve esta sección ni el bloque de la portada, y `/ministerios` redirige a la portada, hasta que publiques al menos un ministerio.';

/** «Todavía no hay ministerios…» (ux.md §4.6). */
export function emptyModuleMessage(): string {
  return 'Todavía no hay ministerios. Agrega el primero para que el público lo vea.';
}

/* ── Validaciones (espejo de FR-014/FR-015; la autoridad es el servidor) ── */

export const REQUIRED_NAME = 'Escribe el nombre del ministerio.';
export const REQUIRED_DESCRIPTION = 'Escribe la descripción del ministerio.';
export const REQUIRED_IMAGE_ALT = 'Escribe el texto alternativo de la imagen.';
export const LINK_INVALID =
  'Ese enlace no parece una dirección web. Pégalo completo; por ejemplo: https://chat.whatsapp.com/…';
export const LINK_UNSAFE =
  'Ese enlace no es válido o no es seguro. Debe empezar por https:// y ser una dirección normal.';
export const PHONE_INVALID =
  'Ese número no parece un celular. Revisa los dígitos (y el + opcional del país).';
export const EMAIL_INVALID = 'Ese correo no parece un correo. Revísalo, por favor.';
/**
 * El contacto exige el nombre del encargado (regla del servidor: `leader_phone`/
 * `leader_email` con `CHECK` de nombre). `ux.md` §7.1 no trae un texto exacto
 * para esta regla —solo la leyenda del interruptor (`LEGEND_NEEDS_NAME`, que
 * habla de *publicar* la ficha)—, así que aquí se dice en las mismas palabras
 * del equipo qué falta para poder **guardar**.
 */
export const REQUIRED_LEADER_NAME_FOR_CONTACT =
  'Escribe primero el nombre del encargado para poder guardar su contacto.';
export const DUPLICATE =
  'Ya existe un ministerio igual (mismo nombre, misma descripción, mismo encargado y mismo enlace). Edita el que ya tienes o cambia algún dato.';
export const SHARED_LINK_NOTICE =
  'Otro ministerio ya usa este mismo enlace de WhatsApp. Revísalo si no es intencional; el ministerio sí se guardó.';

/** «Ese texto es demasiado largo (máximo {límite}).» (ux.md §7.1). */
export function textTooLong(limit: number): string {
  return `Ese texto es demasiado largo (máximo ${limit}).`;
}

/** «{n} de 2.000 caracteres. Recorta {m} para poder guardar.» (ux.md §7.1). */
export function descriptionLimitMessage(count: number): string {
  return `${count} de ${DESCRIPTION_MAX} caracteres. Recorta ${count - DESCRIPTION_MAX} para poder guardar.`;
}

/** Contador vivo de un campo («{n}/{máximo}»). */
export function counter(count: number, max: number): string {
  return `${count}/${max}`;
}

/** «El contenido en español es obligatorio… Falta la versión en español de {campo}.» */
export function spanishRequiredFor(field: string): string {
  return `El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de ${field}.`;
}

/* ── Bloque «Encargado» del formulario (§4.5) ────────────────────────── */

export const SWITCH_LABEL = 'Mostrar la ficha del encargado en el sitio público';
export const SWITCH_HELP =
  'La ficha del encargado está oculta hasta que actives este interruptor. Empieza apagado para que seas tú quien decida cuándo hacer pública su información.';
export const LEGEND_LEADER_VISIBLE =
  'Se publicará: el nombre y (si existe) el contacto del encargado aparecerán en la página de Ministerios.';
export const LEGEND_LEADER_HIDDEN =
  'La ficha del encargado quedará oculta: el público no verá ni su nombre ni su contacto, en ningún lugar del sitio. Tus datos quedan guardados aquí; puedes activar el interruptor otra vez cuando quieras publicarlos.';
export const LEGEND_NEEDS_NAME = 'Para publicar la ficha, escribe primero el nombre del encargado.';

export const CONTACT_NONE = '— Sin contacto —';
export const CONTACT_PHONE = 'Celular';
export const CONTACT_EMAIL = 'Correo';

/* ── Ayudas de formulario (§4.4/§4.5) ────────────────────────────────── */

export const ENGLISH_HELP =
  'Si lo dejas vacío, el público verá este contenido en su versión en español.';
export const SORT_HELP =
  'Los ministerios se muestran en orden: menor número primero. Con empate, el más antiguo va antes.';
export const SCHEDULE_HELP =
  'Escribe el horario tal como quieres que lo lean; por ejemplo: «Todos los viernes, 7:00 p. m.». No se calculan fechas ni calendarios.';
export const PLACE_HELP = 'Opcional. P. ej. «Salón principal» o «En la casa de la familia Pérez».';
export const CONTACT_HELP =
  'Puedes dejarlo sin contacto, o elegir celular o correo (uno solo). Al cambiar el tipo se borra el contacto anterior.';
export const LINK_HELP =
  'Pega la invitación del grupo; por ejemplo: https://chat.whatsapp.com/… Cualquier https válido.';
export const IMAGE_HELP =
  'Opcional. Añade el texto alternativo para personas que no la pueden ver. El logo de la iglesia respeta las reglas de marca.';
