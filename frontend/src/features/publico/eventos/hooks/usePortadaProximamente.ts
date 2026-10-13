import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { ApiError } from '../../../../api/client';
import { getPortadaProximamente, type PortadaProximamente } from '../../../../api/agenda';
import { useLanguage } from '../../i18n/useLanguage';

/**
 * Clave raíz de la consulta del extracto de la portada (una entrada por idioma).
 * Es una petición **propia** del bloque (⟲ D-26), independiente de la consulta
 * de la sección `/eventos` (`useUpcomingEvents`): el bloque no comparte datos
 * con la página completa.
 */
export const PORTADA_PROXIMAMENTE_QUERY_KEY = ['portada-proximamente'] as const;

/**
 * Extracto «Lo que viene» de la portada (US7/FR-025…FR-027): el servidor
 * entrega hasta **5** ítems próximos —eventos y actividades mezclados por
 * cercanía, con la ventana de 2 meses y su respaldo ya aplicados— resueltos al
 * idioma activo (fallback `en → es` por campo, FR-008). El cliente solo
 * presenta: no filtra, no reordena ni reimplementa el fallback.
 *
 * Al cambiar de idioma se conserva la respuesta anterior
 * (`placeholderData: keepPreviousData`, patrón de F3/F10) y, si la recarga
 * falla, se mantiene la última versión válida (ux.md §4.10: «error con ítems en
 * caché») para que la portada no parpadee ni se quede muda.
 */
export function usePortadaProximamente() {
  const { lang } = useLanguage();

  const query = useQuery<PortadaProximamente, ApiError>({
    queryKey: [...PORTADA_PROXIMAMENTE_QUERY_KEY, lang],
    queryFn: () => getPortadaProximamente(lang),
    placeholderData: keepPreviousData,
  });

  // Ajuste de estado durante el render (patrón de React para estado derivado,
  // el mismo de `useUpcomingEvents`): sin efecto ni render en cascada.
  const [retained, setRetained] = useState<PortadaProximamente>();
  if (query.data && query.data !== retained) {
    setRetained(query.data);
  }

  return { ...query, data: query.data ?? retained };
}
