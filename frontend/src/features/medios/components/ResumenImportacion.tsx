import type { ResumenBusqueda } from '../../../api/medios';
import { Notice } from '../../../components/Notice';
import {
  SUMMARY_BOTH_FEEDS_FAILED,
  SUMMARY_NO_NEWS,
  summaryAvailabilityChecked,
  summaryFeedFailed,
  summaryImported,
  summaryImportedWithMarks,
} from '../messages';

export interface ResumenImportacionProps {
  /** Resumen devuelto por «Buscar episodios nuevos» (FR-018). */
  resumen: ResumenBusqueda;
}

/**
 * Resumen de la última búsqueda (ux.md §4.4, FR-018): aparece entre la zona de
 * búsqueda y el listado ya refrescado, y **permanece visible hasta la próxima
 * búsqueda o recarga** —no es un toast de 3 s (⟲ D-22: el equipo necesita poder
 * contar el resultado a otros)—.
 *
 * Los seis casos del catálogo, cada uno con su tono:
 *  1. nuevos importados, todo unido y sin marcas → **éxito**;
 *  2. nuevos importados **y** algunos marcados → **éxito** (el segmento de
 *     marcas solo aparece si su número es ≥ 1, ⟲ D-19);
 *  3. sin novedades (FR-012) → **neutro**;
 *  4. un feed falló (FR-016) → **aviso**, nombrando el servicio («YouTube» / «el
 *     podcast», nunca tecnicismos, ⟲ D-21) y recordando que nada se ocultó ni se
 *     marcó por esa causa;
 *  5. ambos feeds fallaron → **error**, con el catálogo a salvo;
 *  6. comprobaciones de disponibilidad (FR-017) → **informativo**, remitiendo a
 *     las marcas de las filas.
 *
 * Un feed caído **no** se mezcla con un mensaje de «0 episodios nuevos»: si no
 * se importó nada porque el otro feed tampoco traía novedades, se dice la verdad
 * con el aviso del feed y sin segmentos en 0.
 */
export function ResumenImportacion({ resumen }: ResumenImportacionProps) {
  const ambosFallaron = resumen.youtube === 'failed' && resumen.podcast === 'failed';
  const feedCaido =
    resumen.youtube === 'failed' ? 'youtube' : resumen.podcast === 'failed' ? 'podcast' : undefined;
  const hayNuevos = resumen.importedCount >= 1;

  return (
    <div className="space-y-2">
      {ambosFallaron && <Notice variant="error">{SUMMARY_BOTH_FEEDS_FAILED}</Notice>}

      {!ambosFallaron && feedCaido && (
        <Notice variant="warning">{summaryFeedFailed(feedCaido)}</Notice>
      )}

      {!ambosFallaron && hayNuevos && (
        <Notice variant="success">
          {resumen.markedCount >= 1
            ? summaryImportedWithMarks(
                resumen.importedCount,
                resumen.joinedCount,
                resumen.markedCount,
              )
            : summaryImported(resumen.importedCount, resumen.joinedCount)}
        </Notice>
      )}

      {!ambosFallaron && !hayNuevos && !feedCaido && (
        <Notice variant="info">{SUMMARY_NO_NEWS}</Notice>
      )}

      {resumen.availabilityCheckedCount >= 1 && (
        <Notice variant="info">
          {summaryAvailabilityChecked(resumen.availabilityCheckedCount)}
        </Notice>
      )}
    </div>
  );
}
