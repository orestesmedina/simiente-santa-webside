/**
 * Catálogo de textos del módulo de donaciones del panel (ux.md §7.1). El panel
 * va **siempre en español** (FR-018): estas cadenas no pasan por el diccionario
 * del sitio público y no se traducen.
 */

/* ── Límites del contrato (espejo de las constantes del service, FR-014) ── */

export const INSTITUTION_MAX = 160;
export const HOLDER_MAX = 160;
export const TEXT_MAX = 2000;
export const NUMBER_INPUT_MAX = 60;

/* ── Encabezado y estados de la vista ────────────────────────────────── */

export const TITLE = 'Gestión de donaciones';
export const HELP =
  'Aquí administras las cuentas y números donde la iglesia recibe ofrendas, y los textos que explican cómo donar y en qué se usan. Solo lo publicado es visible al público.';

export const LOAD_ERROR_ACCOUNTS = 'No se pudo cargar las cuentas y números.';
export const LOAD_ERROR_TEXTS = 'No se pudo cargar los textos.';
export const NO_PERMISSION_ERROR =
  'No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol.';
export const SYSTEM_ERROR =
  'No se pudo completar la operación. Tus escritos están a salvo; vuelve a intentarlo en unos minutos.';
export const RETRY = 'Reintentar';
export const LOADING_ACCOUNTS = 'Cargando las cuentas…';
export const LOADING_TEXTS = 'Cargando los textos…';

/* ── Guardado y publicación ──────────────────────────────────────────── */

export const CREATED_ACCOUNT = 'Cuenta creada.';
export const CREATED_SINPE = 'Número de SINPE creado.';
export const DRAFT_NOTICE_ACCOUNT =
  'Quedó en borrador: no la verá el público hasta que la publiques.';
export const DRAFT_NOTICE_SINPE =
  'Quedó en borrador: no lo verá el público hasta que lo publiques.';
export const SAVED = 'Cambios guardados.';
export const SAVED_PUBLISHED = 'Cambios guardados: ya visibles en la sección.';
export const PUBLISHED = 'Publicado en "Ofrendas y diezmos".';
export const UNPUBLISHED =
  'Retirado de la sección. Estará disponible como borrador para publicarlo cuando quieras.';
export const DELETED = 'Eliminado.';
export const SAVE = 'Guardar cambios';

/* ── Acciones y confirmaciones ───────────────────────────────────────── */

export const ADD_ACCOUNT = 'Agregar cuenta';
export const ADD_SINPE = 'Agregar número';
export const EDIT_ACTION = 'Editar';
export const DELETE_ACTION = 'Eliminar';

export const RETIRE_ACTION = 'Retirar de la sección';
export const RETIRE_CONFIRM_TITLE = 'Retirar de la sección';
export const RETIRE_CONFIRM_ACTION = 'Retirar';
export const RETIRE_CANCEL = 'Dejar como está';

export const DELETE_CONFIRM_TITLE = 'Eliminar';
export const DELETE_CONFIRM_ACTION = 'Eliminar';
export const DELETE_CANCEL = 'Cancelar';

/** «"{entidad}" dejará de ser visible al público…» (ux.md §5.d). */
export function retireConfirmDescription(entity: string): string {
  return `"${entity}" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.`;
}

/** «{tipo} {entidad} {número} se eliminará para siempre…» (ux.md §4.3/§7.1). */
export function deleteConfirmDescription(
  typeLabel: string,
  institution: string,
  number: string,
): string {
  return `${typeLabel} ${institution} ${number} se eliminará para siempre. Esto no se puede deshacer.`;
}

/* ── Tipos de método ─────────────────────────────────────────────────── */

export const TYPE_IBAN = 'Cuenta IBAN';
export const TYPE_SINPE = 'SINPE Móvil';

/* ── Zona de textos (ux.md §4.3) ─────────────────────────────────────── */

export const TEXTS_ZONE_TITLE = 'Textos de la sección';
export const TEXT_INTRO_TITLE = 'Introductorio';
export const TEXT_USAGE_TITLE = '¿En qué usamos las donaciones?';
export const TEXT_NOT_WRITTEN = 'Todavía sin escribir.';

/* ── Listado por grupos (ux.md §4.3) ─────────────────────────────────── */

export const GROUP_IBAN_TITLE = 'Cuentas IBAN';
export const GROUP_SINPE_TITLE = 'Números de SINPE Móvil';

export const COL_ORDER = 'Orden';
export const COL_NUMBER = 'Número';
export const COL_INSTITUTION = 'Entidad';
export const COL_HOLDER = 'Titular';
export const COL_CURRENCY = 'Moneda';
export const COL_STATE = 'Estado';
export const COL_ACTIONS = 'Acciones';

export const CURRENCY_COLONES = 'Colones';
export const CURRENCY_DOLARES = 'Dólares';
/** El listado del panel sí usa «—» para el titular vacío (⟲ D-16). */
export const EMPTY_DASH = '—';

/* ── Formulario (ux.md §4.4/§4.5) ────────────────────────────────────── */

export const FIELD_INSTITUTION = 'Entidad financiera';
export const FIELD_HOLDER = 'Titular de la cuenta';
export const FIELD_NUMBER_IBAN = 'Número de cuenta (IBAN)';
export const FIELD_NUMBER_SINPE = 'Número de celular (SINPE Móvil)';
export const FIELD_CURRENCY = 'Moneda';
export const FIELD_ORDER = 'Orden';
export const TAB_ES = 'Español';
export const TAB_EN = 'English (opcional)';

export const ENGLISH_HELP =
  'Si lo dejas vacío, el público verá este contenido en su versión en español.';
export const INSTITUTION_HELP = 'El nombre del banco; por ejemplo: Banco Nacional de Costa Rica.';
export const HOLDER_HELP =
  'Opcional. Solo si quieres que el público vea a nombre de quién está la cuenta.';
export const ORDER_HELP =
  'Las cuentas (o los números, dentro de su grupo) se muestran en orden: menor número primero. Con empate, el más antiguo va antes.';
export const NUMBER_IBAN_HELP =
  'Puedes pegarlo tal como te lo dieron, con espacios o guiones: se guardará limpio y el público lo verá agrupado en bloques.';
export const NUMBER_SINPE_HELP =
  'El número de celular a donde llega la donación; acepta +506 o no — se guarda con su prefijo normalizado.';
export const TEXT_HELP = 'Uno o varios párrafos. Aparecerá tal como lo escribas, con sus saltos.';
export const TYPE_FIXED_HELP =
  'El tipo no se cambia aquí: si necesitas el número como cuenta o como celular, crea el nuevo y elimina el otro.';
export const CURRENCY_HELP =
  'Solo se elige en las cuentas IBAN: colones o dólares, según la cuenta del banco.';

/* ── Validaciones (espejo de FR-014/FR-015/FR-016; la autoridad es el servidor) ── */

export const REQUIRED_INSTITUTION = 'Escribe el nombre del banco.';
export const REQUIRED_NUMBER =
  'Escribe el número de la cuenta (o el número de celular para SINPE Móvil).';
export const IBAN_INVALID =
  'Ese número no parece una cuenta IBAN válida. Revisa la cantidad de letras y dígitos: vale tanto el IBAN de Costa Rica como uno internacional.';
export const SINPE_INVALID =
  'Ese número no parece un celular de Costa Rica: son 8 dígitos, con o sin el prefijo +506.';
export const REQUIRED_CURRENCY = 'Elige la moneda de la cuenta: colones o dólares.';
export const DUPLICATE =
  'Ya existe esa misma cuenta o número (mismo tipo, entidad y número). Edita el que ya tienes o cambia algún dato.';
export const NOTHING_PUBLISHED =
  'Nada está publicado ahora: el público no ve esta sección ni el bloque de la portada, y `/ofrendas` redirige a la portada, hasta que publiques al menos un método (cuenta o número) o un texto.';

/** «Ese texto es demasiado largo (máximo {límite}).» (ux.md §7.1). */
export function textTooLong(limit: number): string {
  return `Ese texto es demasiado largo (máximo ${limit}).`;
}

/** «{n} de 2.000 caracteres. Recorta {m} para poder guardar.» (ux.md §7.1). */
export function textLimitMessage(count: number): string {
  return `${count} de ${TEXT_MAX} caracteres. Recorta ${count - TEXT_MAX} para poder guardar.`;
}

/** Contador vivo de un campo («{n}/{máximo}»). */
export function counter(count: number, max: number): string {
  return `${count}/${max}`;
}

/** «El contenido en español es obligatorio… Falta la versión en español de {campo}.» */
export function spanishRequiredFor(field: string): string {
  return `El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de ${field}.`;
}

/** Estado vacío de cada bloque del listado (ux.md §4.6). */
export function emptyGroupMessage(kind: 'iban' | 'sinpe'): string {
  return kind === 'iban'
    ? 'Todavía no hay cuentas IBAN. Agrega la primera para que el público la vea.'
    : 'Todavía no hay números de SINPE Móvil. Agrega el primero para que el público lo vea.';
}
