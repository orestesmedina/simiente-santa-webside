import type { MarcaDisponibilidad, MarcaRevision, Proveedor } from '../../api/medios';

/**
 * Catálogo de textos del módulo «Medios» del panel (ux.md §7.1). El panel va
 * **siempre en español** (FR-024): estas cadenas no pasan por el diccionario del
 * sitio público y no se traducen. Los nombres de los servicios se dicen como los
 * llama la iglesia —«YouTube», «el podcast»— y nunca con tecnicismos internos
 * («feed», «Atom», «RSS», «Anchor»: prohibidos por FR-018/D-21).
 */

/* ── Límites del contrato (espejo del service, FR-020) ───────────────── */

/** Límite de la descripción corta por idioma (default revisable #1). */
export const DESCRIPTION_MAX = 300;

/* ── Encabezado y banda de ayuda (ux.md §4.3) ────────────────────────── */

export const TITLE = 'Medios';
export const HELP =
  'Aquí miras y mantienes las grabaciones que el sitio trae desde el canal de YouTube y el feed del podcast. Pulsa “Buscar episodios nuevos” cuando se publique algo nuevo y revisa con calma lo que quede marcado; solo lo visible lo verá el público.';

/* ── Zona de búsqueda (ux.md §4.3/§4.4) ──────────────────────────────── */

export const SEARCH_ZONE_TITLE = 'Desde YouTube y el podcast';
export const SEARCH_ACTION = 'Buscar episodios nuevos';
export const SEARCH_SUPPORT =
  'Es seguro pulsarlo todas las veces que quieras: solo agrega lo que falte, nunca borra y nunca cambia lo que ya decidiste.';
export const SEARCHING = 'Buscando…';
export const SEARCHING_STATUS = 'Buscando episodios nuevos en YouTube y en el podcast…';

/* ── Resumen de la búsqueda (ux.md §4.4/§7.1) ────────────────────────── */

/**
 * «Se agregaron {N} episodios nuevos; {A} se unieron con su video y su audio. Ya
 * están visibles en “Medios”.» (ux.md §7.1). Un segmento **solo** se menciona si
 * su número es ≥ 1 (⟲ D-19: nunca «0 se unieron»).
 */
export function summaryImported(imported: number, joined: number): string {
  const joinedSegment = joined >= 1 ? `; ${joined} se unieron con su video y su audio` : '';
  return `Se agregaron ${imported} episodios nuevos${joinedSegment}. Ya están visibles en “Medios”.`;
}

/**
 * «Se agregaron {N} episodios nuevos; {A} se unieron con su video y su audio;
 * {D} quedaron marcados como posible duplicado. Revísalos abajo.» (ux.md §7.1,
 * ⟲ D-19: los segmentos en 0 se omiten).
 */
export function summaryImportedWithMarks(imported: number, joined: number, marked: number): string {
  const segments = [`Se agregaron ${imported} episodios nuevos`];
  if (joined >= 1) {
    segments.push(`${joined} se unieron con su video y su audio`);
  }
  if (marked >= 1) {
    segments.push(`${marked} quedaron marcados como posible duplicado`);
  }
  return `${segments.join('; ')}. Revísalos abajo.`;
}

/** «No hay episodios nuevos: nada cambió…» (FR-012, ⟲ D-20). */
export const SUMMARY_NO_NEWS =
  'No hay episodios nuevos: nada cambió. Cuando la iglesia publique algo más, vuelve a pulsar el botón.';

/** Nombre cotidiano de cada origen (FR-018/⟲ D-21: nunca «Atom/RSS/Anchor»). */
export function providerLabel(provider: Proveedor): string {
  return provider === 'youtube' ? 'YouTube' : 'el podcast';
}

/** El origen con su artículo, para frases («el feed de YouTube» / «el feed del podcast»). */
export function feedLabel(provider: Proveedor): string {
  return provider === 'youtube' ? 'feed de YouTube' : 'feed del podcast';
}

/**
 * «El feed de {YouTube / el podcast} no respondió. Se agregó lo que trajo {el
 * otro}; nada se ocultó ni se marcó por esa causa. Vuelve a pulsar la búsqueda en
 * un rato para recuperar lo que falte.» (FR-016).
 */
export function summaryFeedFailed(failed: Proveedor): string {
  const other = failed === 'youtube' ? 'el podcast' : 'YouTube';
  return `El ${feedLabel(failed)} no respondió. Se agregó lo que trajo ${other}; nada se ocultó ni se marcó por esa causa. Vuelve a pulsar la búsqueda en un rato para recuperar lo que falte.`;
}

/** «Ninguno de los dos feeds respondió… todo lo que tenías está a salvo.» */
export const SUMMARY_BOTH_FEEDS_FAILED =
  'Ninguno de los dos feeds respondió. No se agregó nada y no se cambió nada: todo lo que tenías está a salvo. Prueba más tarde.';

/**
 * «La búsqueda revisó {M} episodios para ver si sus enlaces siguen funcionando:
 * mira las marcas.» (FR-017). Solo se muestra con al menos una comprobación.
 */
export function summaryAvailabilityChecked(checked: number): string {
  return `La búsqueda revisó ${checked} episodios para ver si sus enlaces siguen funcionando: mira las marcas.`;
}

/* ── Listado del catálogo (ux.md §4.3) ───────────────────────────────── */

export const COL_TITLE = 'Episodio';
export const COL_FORMATS = 'Formatos';
export const COL_STATE = 'Estado';
export const COL_MARK = 'Marca';
export const COL_ACTIONS = 'Acciones';

/** Chips de formato por fila (⟲ D-15: el formato se ve, no se deduce). */
export const FORMAT_VIDEO = 'Video';
export const FORMAT_AUDIO = 'Audio';

/** Píldoras de estado (ux.md §1: «Visible» / «Oculto», siempre con palabra). */
export const STATE_VISIBLE = 'Visible';
export const STATE_HIDDEN = 'Oculto';

/** Origen de cada entrada (FR-019; ux.md §4.6: «De YouTube» / «Del podcast»). */
export const ORIGIN_VIDEO = 'De YouTube';
export const ORIGIN_AUDIO = 'Del podcast';
export const ORIGIN_BOTH = 'De YouTube y del podcast';

/** Etiqueta de la fecha de importación de la fila (FR-019, ⟲ A2). */
export function importedOn(formattedDate: string): string {
  return `Importado el ${formattedDate}`;
}

/** Acciones de fila (ux.md §4.3/§4.6). Ninguna es solo icono. */
export const ACTION_EDIT_DESCRIPTION = 'Editar descripción';
export const ACTION_HIDE = 'Ocultar';
export const ACTION_SHOW = 'Mostrar';
export const ACTION_REVIEW_DUPLICATE = 'Revisar duplicado';

/* ── Estados vacíos y de carga (ux.md §4.8/§5.b) ─────────────────────── */

export const EMPTY_TITLE = 'Todavía no hay episodios.';
export const EMPTY_HELP =
  'Pulsa “Buscar episodios nuevos” para que el sitio lea por primera vez el canal de YouTube y el feed del podcast.';
export const FILTER_EMPTY = 'No hay nada pendiente de revisión: todo está al día.';
export const LOADING = 'Cargando los episodios…';
export const LOAD_ERROR = 'No se pudo cargar los episodios.';
export const RETRY = 'Reintentar';

/* ── Filtro de marcas (⟲ D-10) ───────────────────────────────────────── */

export const FILTER_ALL = 'Todo';
export const FILTER_PENDING = 'Solo lo pendiente de revisión';
export const FILTER_HELP =
  'Las marcas las pone el sistema cuando algo necesita tu revisión; se resuelven en cada fila.';

/* ── Avisos de las operaciones (ux.md §5.d/§7.1) ─────────────────────── */

export const HIDDEN_NOTICE = 'Quedó oculto: el público no lo ve.';
export const SHOWN_NOTICE = 'Ahora está visible en “Medios”.';
export const DESCRIPTION_SAVED_PUBLISHED = 'Cambios guardados: ya visibles en la sección.';
export const DESCRIPTION_SAVED = 'Cambios guardados.';
export const MERGED_NOTICE = 'Unidas en una sola ficha, visible en “Medios”.';
export const NOT_SAME_NOTICE =
  'Listo: quedaron las dos como episodios distintos, visibles sin marca.';
export const MARK_CLOSED = 'La marca se cerró y no se vuelve a abrir por esta misma causa.';
export const NO_PERMISSION_ERROR =
  'No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol.';
export const SYSTEM_ERROR =
  'No se pudo completar la operación. Tus episodios están a salvo: vuelve a intentarlo en unos minutos.';

/* ── Dialog «Editar descripción» (ux.md §4.5/§7.1) ───────────────────── */

export const DESCRIPTION_FORM_TITLE = 'Editar descripción';
export const DESCRIPTION_CONTEXT = 'Estás ajustando la descripción de:';
export const DESCRIPTION_HELP =
  'Se muestra tal como la escribas, en la ficha del episodio. La descripción completa del feed no se publica.';
export const TAB_ES = 'Español';
export const TAB_EN = 'English (opcional)';
export const ENGLISH_HELP =
  'Si lo dejas vacío, el público verá este contenido en su versión en español.';
export const SPANISH_REQUIRED =
  'El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de la descripción corta.';
export const SAVE = 'Guardar';
export const SAVING = 'Guardando…';
export const CANCEL = 'Cancelar';

/** «{n} de {L} caracteres. Recorta {m} para poder guardar.» (ux.md §7.1). */
export function descriptionLimitMessage(count: number): string {
  return `${count} de ${DESCRIPTION_MAX} caracteres. Recorta ${count - DESCRIPTION_MAX} para poder guardar.`;
}

/** Contador vivo del campo («{n}/{L}»). */
export function counter(count: number): string {
  return `${count}/${DESCRIPTION_MAX}`;
}

/* ── Dialog «Revisar duplicado» (ux.md §4.6/§7.1) ────────────────────── */

export const MERGE_DIALOG_TITLE = 'Revisar duplicado';
export const MERGE_PREVIEW_TITLE = 'Así quedaría la ficha si las unes.';
export const MERGE_PREVIEW_NOTE = 'Al unir, el nombre publicado es el de YouTube.';
export const MERGE_ACTION = 'Unir';
export const MERGE_CONFIRM_TITLE = 'Unir las dos entradas';
export const MERGE_CONFIRM =
  'Se unirán las dos entradas en una sola ficha. Esto no se puede deshacer.';
export const MERGE_CONFIRM_ACTION = 'Unir definitivamente';
export const NOT_SAME_ACTION = 'No son el mismo sermón';

/* ── Marcas de revisión y sus resoluciones (ux.md §4.7/§7.1) ─────────── */

export const MARK_POSSIBLE_DUPLICATE = 'Posible duplicado';
export const MARK_UNAVAILABLE = 'No disponible';
export const MARK_MISSING_FROM_FEED = 'Ya no aparece en el feed';

/** Etiqueta de la píldora de una marca de revisión (siempre con palabra, FR-030). */
export function markLabel(mark: MarcaRevision): string {
  switch (mark) {
    case 'possible_duplicate':
      return MARK_POSSIBLE_DUPLICATE;
    case 'unavailable':
      return MARK_UNAVAILABLE;
    case 'missing_from_feed':
      return MARK_MISSING_FROM_FEED;
  }
}

/** Explicador de una línea que acompaña a la píldora (ux.md §4.7). */
export function markExplainer(
  mark: MarcaRevision,
  provider: Proveedor,
  linkAvailable?: boolean,
): string {
  if (mark === 'possible_duplicate') {
    return 'El sistema no pudo estar seguro de si estas dos entradas son el mismo sermón.';
  }
  if (mark === 'unavailable') {
    return 'Se comprobó: el enlace de esta grabación ya no funciona (borrada o privada). El público no la ve.';
  }
  const service = feedLabel(provider);
  return linkAvailable
    ? `Salió de la ventana del ${service} (así es: la ventana es corta). Su enlace sigue funcionando y sigue visible para el público.`
    : `Salió de la ventana del ${service}. No se pudo comprobar si su enlace sigue funcionando; mientras tanto, sigue visible para el público.`;
}

/** Acciones de resolución de cada marca (ux.md §4.7). */
export const RESOLVE_SHOW_AGAIN = 'Mostrar de nuevo';
export const RESOLVE_KEEP_HIDDEN = 'Mantener oculto';
export const RESOLVE_KEEP_VISIBLE = 'Mantener visible';

/** «Si el enlace de “{nombre}” sigue sin funcionar, volverá a quedar oculto…» (⟲ D-28). */
export function showAgainConfirm(name: string): string {
  return `Si el enlace de “${name}” sigue sin funcionar, volverá a quedar oculto la próxima vez que se revise.`;
}

export const SHOW_AGAIN_ACTION = 'Mostrar';
export const SHOW_AGAIN_CANCEL = 'Dejarlo como está';

/** Etiqueta de una marca ya resuelta (la lista la conserva como registro). */
export const MARK_RESOLVED = 'Resuelta';

/** `true` si la marca pertenece al eje de disponibilidad (P9-19). */
export function isAvailabilityMark(mark: MarcaRevision): mark is MarcaDisponibilidad {
  return mark === 'unavailable' || mark === 'missing_from_feed';
}
