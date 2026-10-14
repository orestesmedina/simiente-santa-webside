import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { ApiError } from '../../../../api/client';
import { getMinisterios, type MinisteriosPublicos } from '../../../../api/ministerios';
import { useLanguage } from '../../i18n/useLanguage';

/**
 * Clave raíz de la consulta de la sección pública (una entrada por idioma). La
 * usan también las mutaciones del panel para invalidar y que el cambio se vea en
 * la primera carga posterior (SC-003).
 */
export const PUBLIC_MINISTRIES_QUERY_KEY = ['ministerios'] as const;

/**
 * Sección pública `/ministerios` (FR-002/FR-004): el servidor entrega solo los
 * ministerios **publicados**, en el orden manual (`sortOrder`, `id`), ya
 * resueltos al idioma activo (fallback `en → es` por campo). El cliente solo
 * presenta: **no** filtra, **no** reordena y **no** reimplementa el fallback ni
 * la privacidad del encargado (FR-026). Con 0 publicados la clave `items` viene
 * omitida y la página decide redirigir a la portada (P6-21).
 *
 * Al cambiar de idioma se conserva la respuesta anterior
 * (`placeholderData: keepPreviousData`, patrón F3/F10) y, si la recarga falla,
 * se mantiene la última versión válida (ux.md §4.1: «error con caché previa»).
 */
export function usePublicMinistries() {
  const { lang } = useLanguage();

  const query = useQuery<MinisteriosPublicos, ApiError>({
    queryKey: [...PUBLIC_MINISTRIES_QUERY_KEY, lang],
    queryFn: () => getMinisterios(lang),
    placeholderData: keepPreviousData,
  });

  // Última versión **válida** vista (patrón de `useUpcomingEvents`): ajuste de
  // estado durante el render, sin efecto ni render en cascada.
  const [retained, setRetained] = useState<MinisteriosPublicos>();
  if (query.data && query.data !== retained) {
    setRetained(query.data);
  }

  return { ...query, data: query.data ?? retained };
}
