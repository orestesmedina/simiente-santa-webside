/**
 * Catálogo de textos del módulo de agenda del panel (ux.md §7.1). El panel va
 * **siempre en español** (FR-016): estas cadenas no pasan por el diccionario del
 * sitio público y no se traducen.
 */

/* ── Límites del contrato (espejo de las constantes del service, FR-015) ── */

export const TITLE_MAX = 160;
export const FREQUENCY_MAX = 160;
export const PLACE_MAX = 200;
export const DETAIL_MAX = 2000;
export const ORGANIZER_MAX = 160;
export const MAP_URL_MAX = 500;
export const IMAGE_ALT_MAX = 300;
export const MAX_DATES = 50;

/* ── Estados de la vista ─────────────────────────────────────────────── */

export const EVENTS_LOAD_ERROR = 'No se pudo cargar los eventos.';
export const ACTIVITIES_LOAD_ERROR = 'No se pudo cargar las actividades.';
export const CATEGORIES_LOAD_ERROR = 'No se pudo cargar el catálogo de categorías.';
export const NO_PERMISSION_ERROR =
  'No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol.';
export const SYSTEM_ERROR =
  'No se pudo completar la operación. Tus escritos están a salvo; vuelve a intentarlo en unos minutos.';
export const RETRY = 'Reintentar';

/* ── Guardado y publicación ──────────────────────────────────────────── */

export const SAVED = 'Cambios guardados.';
export const SAVED_PUBLISHED = 'Cambios guardados: ya visibles en la sección.';
export const PUBLISHED = 'Publicado en "Eventos y actividades".';
export const UNPUBLISHED =
  'Retirado de la sección. Estará disponible como borrador para publicarlo cuando quieras.';
export const DRAFT_NOTICE = 'Quedó en borrador: no lo verá el público hasta que lo publiques.';
export const EVENT_CREATED = 'Evento creado.';
export const ACTIVITY_CREATED = 'Actividad creada.';
export const CATEGORY_CREATED = 'Categoría creada.';
export const DELETED = 'Eliminado.';
export const SAVE = 'Guardar cambios';

export const ADD_EVENT = 'Agregar evento';
export const ADD_ACTIVITY = 'Agregar actividad';
export const ADD_CATEGORY = 'Agregar categoría';
export const EDIT_ACTION = 'Editar';
export const DELETE_ACTION = 'Eliminar';

/* ── Marcas y avisos del listado ─────────────────────────────────────── */

export const PAST_LABEL = 'Pasado';
export const PAST_LEGEND =
  'Sus fechas ya pasaron: el público no lo ve hasta que cargues una próxima fecha.';
export const NOTHING_PUBLISHED =
  'Nada está publicado ahora: el público no ve esta sección hasta que publiques al menos un elemento.';
export const ALL_PAST_NOTICE =
  'El público no ve esta sección porque sus fechas ya pasaron; carga la próxima fecha.';
export const NO_NEXT_DATE = '—';

/* ── Confirmaciones ──────────────────────────────────────────────────── */

export const RETIRE_ACTION = 'Retirar de la sección';
export const RETIRE_CONFIRM_TITLE = 'Retirar de la sección';
export const RETIRE_CONFIRM_ACTION = 'Retirar';
export const RETIRE_CANCEL = 'Dejar como está';

/** «"{nombre}" dejará de ser visible al público…» (ux.md §5.c). */
export function retireConfirmDescription(name: string): string {
  return `"${name}" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.`;
}

export const DELETE_CONFIRM_TITLE = 'Eliminar';
export const DELETE_CONFIRM_ACTION = 'Eliminar';
export const DELETE_CANCEL = 'Cancelar';

/** «"{nombre}" se eliminará para siempre. Esto no se puede deshacer.» (ux.md §7.1). */
export function deleteConfirmDescription(name: string): string {
  return `"${name}" se eliminará para siempre. Esto no se puede deshacer.`;
}

/* ── Validaciones (espejo de FR-015; la autoridad es el servidor) ────── */

export const REQUIRED_TITLE = 'Escribe el título.';
export const REQUIRED_PLACE = 'Escribe el lugar.';
export const REQUIRED_DETAIL = 'Escribe el detalle.';
export const REQUIRED_IMAGE_ALT = 'Escribe el texto alternativo de la imagen.';
export const REQUIRED_EVENT_DATES = 'Agrega al menos una fecha para el evento.';
export const REQUIRED_ACTIVITY_DATE = 'Escoge la fecha de la actividad.';
export const DATE_INVALID = 'Ese día no existe o no está bien escrito. Escógelo del calendario.';
export const DUPLICATE_DATE = 'Esa fecha ya está en la lista del evento.';
export const TOO_MANY_DATES = `Un evento admite hasta ${MAX_DATES} fechas.`;
export const END_TIME_INVALID = 'La hora de fin debe ser posterior a la hora de inicio.';
export const TIME_INVALID = 'Escribe la hora en formato de 24 horas, por ejemplo 19:00.';
export const END_WITHOUT_START = 'Escribe la hora de inicio para poder poner una hora de fin.';
export const MAP_URL_INVALID =
  'Ese enlace no parece una dirección web. Pégalo completo; por ejemplo: https://maps.google.com/?q=…';
export const SPANISH_REQUIRED =
  'El contenido en español es obligatorio: es el idioma base del sitio.';
export const CATEGORY_DUPLICATE = 'Ya existe una categoría con ese nombre.';
export const REQUIRED_CATEGORY_NAME = 'Escribe el nombre de la categoría.';

/** «Ese texto es demasiado largo (máximo {límite}).» (ux.md §7.1). */
export function textTooLong(limit: number): string {
  return `Ese texto es demasiado largo (máximo ${limit}).`;
}

/** «{n} de 2.000 caracteres. Recorta {m} para poder guardar.» (ux.md §7.1). */
export function detailLimitMessage(count: number): string {
  return `${count} de ${DETAIL_MAX} caracteres. Recorta ${count - DETAIL_MAX} para poder guardar.`;
}

/** Contador vivo de un campo («{n}/{máximo}»). */
export function counter(count: number, max: number): string {
  return `${count}/${max}`;
}

/** «Ya existe un {evento/actividad} igual (mismo título, misma fecha y mismo lugar)…». */
export function duplicateMessage(concept: 'evento' | 'actividad'): string {
  return `Ya existe un ${concept} igual (mismo título, misma fecha y mismo lugar). Edita el que ya tienes o cambia algún dato.`;
}

/** «3 fechas futuras» / «1 fecha futura» (ux.md §4.3). */
export function futureDatesLabel(count: number): string {
  return count === 1 ? '1 fecha futura' : `${count} fechas futuras`;
}

/** «Todavía no hay {eventos/actividades}…» (ux.md §4.7). */
export function emptyModuleMessage(concept: 'eventos' | 'actividades'): string {
  return `Todavía no hay ${concept}. Agrega el primero para que el público lo vea.`;
}

/** «Está en uso por {n} elementos: quítala primero de los elementos.» (⟲ D-16). */
export function categoryInUseMessage(count: number): string {
  return `Está en uso por ${count} ${count === 1 ? 'elemento' : 'elementos'}: quítala primero de los elementos.`;
}

/* ── Ayudas de formulario ────────────────────────────────────────────── */

export const ENGLISH_HELP =
  'Si lo dejas vacío, el público verá este contenido en su versión en español.';
export const DATES_HELP =
  'Cada fila es una fecha concreta en que se repite el encuentro. La sección siempre muestra la más cercana.';
export const MAP_HELP =
  'Pega el enlace de Google Maps o similar; por ejemplo: https://maps.google.com/?q=…';
export const IMAGE_HELP =
  'Opcional. Añade el texto alternativo para personas que no la pueden ver.';
export const ORGANIZER_HELP = 'Opcional: quién organiza o está a cargo.';

/* ── Encabezados del panel ───────────────────────────────────────────── */

export const EVENTS_TITLE = 'Gestión de eventos';
export const EVENTS_HELP =
  'Aquí administras los encuentros periódicos de la iglesia (p. ej., noche de aposentos, vigilia). Solo lo publicado y con fecha futura es visible al público.';
export const ACTIVITIES_TITLE = 'Gestión de actividades';
export const ACTIVITIES_HELP =
  'Aquí administras las acciones puntuales de la iglesia. Solo lo publicado y con fecha futura es visible al público.';
export const CATEGORIES_TITLE = 'Categorías de agenda';
export const CATEGORIES_HELP =
  'El catálogo de etiquetas que acompañan a los eventos y las actividades. Al renombrar una categoría, los elementos muestran el nombre actual.';
