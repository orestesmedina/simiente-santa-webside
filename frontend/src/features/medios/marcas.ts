import type { MarcaDisponibilidad, MarcaRevision, MedioAdmin, Proveedor } from '../../api/medios';

/**
 * Reglas de marcas de revisión del panel (ux.md §4.7, P9-19).
 *
 * Las marcas viven en **dos ejes independientes** que **conviven**: el eje de
 * **duplicados** (`duplicateOfId` + `duplicateMarkResolvedAt`; un par abierto
 * **bloquea la publicación**, FR-009/SC-002) y el eje de **disponibilidad**
 * (`availabilityMark` + `availabilityMarkResolvedAt`). «Pendiente» significa
 * **sin resolver** en su eje: es lo que el filtro de marcas pide al servidor con
 * `reviewMark` y lo que la fila ofrece resolver.
 */

/** `true` si la ficha tiene un **par de duplicados abierto** (pendiente). */
export function parDuplicadoAbierto(item: MedioAdmin): boolean {
  return Boolean(item.duplicateOfId) && !item.duplicateMarkResolvedAt;
}

/** Marca de disponibilidad **sin resolver**; `undefined` si no hay o ya se cerró. */
export function marcaDisponibilidadPendiente(item: MedioAdmin): MarcaDisponibilidad | undefined {
  if (!item.availabilityMark || item.availabilityMarkResolvedAt) {
    return undefined;
  }
  return item.availabilityMark;
}

/** Todas las marcas **pendientes** de una ficha (una por eje, en orden de catálogo). */
export function marcasPendientes(item: MedioAdmin): MarcaRevision[] {
  const marks: MarcaRevision[] = [];
  if (parDuplicadoAbierto(item)) {
    marks.push('possible_duplicate');
  }
  const disponibilidad = marcaDisponibilidadPendiente(item);
  if (disponibilidad) {
    marks.push(disponibilidad);
  }
  return marks;
}

/** Conteo de pendientes por clase de marca (para las pastillas del filtro, ⟲ D-10). */
export function contarPendientes(items: MedioAdmin[]): Record<MarcaRevision, number> {
  const counts: Record<MarcaRevision, number> = {
    possible_duplicate: 0,
    unavailable: 0,
    missing_from_feed: 0,
  };
  for (const item of items) {
    for (const mark of marcasPendientes(item)) {
      counts[mark] += 1;
    }
  }
  return counts;
}

/**
 * Servicio que nombra el explicador de «ya no aparece en el feed» (ux.md §4.7).
 *
 * El DTO del panel no guarda **de qué** feed salió la entrada, así que se usa el
 * proveedor de su primera fuente —la del video si el sermón está en los dos
 * lados, que es la base de la ficha al unir (P9-10)— y, si la ficha aún no tiene
 * fuentes, el que sugiere su formato.
 */
export function proveedorDelEpisodio(item: MedioAdmin): Proveedor {
  const primera = item.sources[0];
  if (primera) {
    return primera.provider;
  }
  return item.origin === 'audio' ? 'podcast' : 'youtube';
}

/** `true` si la última comprobación confirmó que el enlace sigue funcionando. */
export function enlaceDisponible(item: MedioAdmin): boolean {
  return item.availabilityResult === 'available';
}
