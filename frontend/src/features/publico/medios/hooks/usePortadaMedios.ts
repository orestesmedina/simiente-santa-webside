import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { ApiError } from '../../../../api/client';
import { getPortadaMedios, type PortadaMedios } from '../../../../api/medios';
import { useLanguage } from '../../i18n/useLanguage';

/**
 * Clave raíz de la consulta del bloque de la portada (una entrada por idioma).
 * Es una petición **propia** del bloque (FR-027), independiente de la consulta de
 * la sección `/medios` (`usePublicMedios`): el bloque solo mira hasta 5 fichas
 * publicables y **no** es fuente de la presencia en la navegación (P9-16/P7-8).
 */
export const PORTADA_MEDIOS_QUERY_KEY = ['portada-medios'] as const;

/**
 * Bloque «Medios» de la portada (US6/FR-027/SC-015): el servidor entrega hasta
 * **5** fichas publicables (tope firme, sin parámetro de límite) en el orden de
 * la sección, ya resueltas al idioma activo. `items` se **omite** sin fichas
 * publicables: el bloque **no se dibuja** (sin mensaje, patrón de F6/F7).
 *
 * Al cambiar de idioma se conserva la respuesta anterior
 * (`placeholderData: keepPreviousData`) y, si la recarga falla, se mantiene la
 * última versión válida (ux.md §4.9: «error con ítems en caché»).
 */
export function usePortadaMedios() {
  const { lang } = useLanguage();

  const query = useQuery<PortadaMedios, ApiError>({
    queryKey: [...PORTADA_MEDIOS_QUERY_KEY, lang],
    queryFn: () => getPortadaMedios(lang),
    placeholderData: keepPreviousData,
  });

  const [retained, setRetained] = useState<PortadaMedios>();
  if (query.data && query.data !== retained) {
    setRetained(query.data);
  }

  return { ...query, data: query.data ?? retained };
}

/**
 * `true` si el documento del bloque trae al menos una ficha publicable
 * (FR-027/SC-015). Con 0 publicables el servidor omite `items` y el bloque **no
 * se dibuja** (sin título ni mensaje).
 */
export function hasPortadaMediosContent(data: PortadaMedios | undefined): boolean {
  return (data?.items?.length ?? 0) > 0;
}
