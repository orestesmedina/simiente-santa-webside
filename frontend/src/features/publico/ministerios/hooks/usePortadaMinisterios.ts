import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { ApiError } from '../../../../api/client';
import { getPortadaMinisterios, type PortadaMinisterios } from '../../../../api/ministerios';
import { useLanguage } from '../../i18n/useLanguage';

/**
 * Clave raíz de la consulta del bloque de la portada (una entrada por idioma).
 * Es una petición **propia** del bloque (P6-9), independiente de la consulta de
 * la sección `/ministerios` (`usePublicMinistries`): la presencia de la entrada
 * de navegación se deduce de **este** documento (bloque + barra + menú + pie
 * comparten caché), mientras que la redirección de la sección la decide la
 * página con su propia consulta.
 */
export const PORTADA_MINISTERIOS_QUERY_KEY = ['portada-ministerios'] as const;

/**
 * Bloque «Ministerios» de la portada (US6/FR-027/FR-028): el servidor entrega
 * hasta **5** ítems publicados (tope firme, sin parámetro de límite) con el DTO
 * mínimo —sin campos de encargado en ningún caso—, ya resueltos al idioma
 * activo. `items` se omite sin publicados: el bloque **no se dibuja** (sin
 * mensaje, diferencia deliberada con `ProximamenteBlock`).
 *
 * Al cambiar de idioma se conserva la respuesta anterior
 * (`placeholderData: keepPreviousData`) y, si la recarga falla, se mantiene la
 * última versión válida (ux.md §4.9: «error con ítems en caché»).
 */
export function usePortadaMinisterios() {
  const { lang } = useLanguage();

  const query = useQuery<PortadaMinisterios, ApiError>({
    queryKey: [...PORTADA_MINISTERIOS_QUERY_KEY, lang],
    queryFn: () => getPortadaMinisterios(lang),
    placeholderData: keepPreviousData,
  });

  const [retained, setRetained] = useState<PortadaMinisterios>();
  if (query.data && query.data !== retained) {
    setRetained(query.data);
  }

  return { ...query, data: query.data ?? retained };
}

/**
 * `true` si el documento del bloque trae al menos un ministerio publicado
 * (FR-005/SC-010). Con 0 publicados el servidor omite `items` y la entrada
 * «Ministerios» **no existe** en barra, menú móvil ni mapa del sitio.
 */
export function hasPortadaMinistryContent(data: PortadaMinisterios | undefined): boolean {
  return (data?.items?.length ?? 0) > 0;
}
