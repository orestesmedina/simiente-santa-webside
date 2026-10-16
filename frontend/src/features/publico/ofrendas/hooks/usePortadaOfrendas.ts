import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { ApiError } from '../../../../api/client';
import { getPortadaOfrendas, type PortadaOfrendas } from '../../../../api/ofrendas';
import { useLanguage } from '../../i18n/useLanguage';

/**
 * Clave raíz de la consulta del bloque de la portada (una entrada por idioma).
 * Es una petición **propia** del bloque (FR-022), independiente de la consulta de
 * la sección `/ofrendas` (`usePublicOfrendas`): el bloque solo mira los métodos
 * publicados y **no** es fuente de la presencia en la navegación (P7-8).
 */
export const PORTADA_OFRENDAS_QUERY_KEY = ['portada-ofrendas'] as const;

/**
 * Bloque «Ofrendas y diezmos» de la portada (US6/FR-022/SC-014): el servidor
 * entrega hasta **5** métodos publicados (tope firme, sin parámetro de límite) con
 * el mismo DTO de ítem que la sección, ya resueltos al idioma activo. `items` se
 * omite sin métodos publicados: el bloque **no se dibuja** (sin mensaje).
 *
 * Al cambiar de idioma se conserva la respuesta anterior
 * (`placeholderData: keepPreviousData`) y, si la recarga falla, se mantiene la
 * última versión válida (ux.md §4.9: «error con ítems en caché»).
 */
export function usePortadaOfrendas() {
  const { lang } = useLanguage();

  const query = useQuery<PortadaOfrendas, ApiError>({
    queryKey: [...PORTADA_OFRENDAS_QUERY_KEY, lang],
    queryFn: () => getPortadaOfrendas(lang),
    placeholderData: keepPreviousData,
    // El documento lo consumen la portada (`OfrendasBlock`, T728) y, por caché,
    // la navegación puede reutilizarlo; con `staleTime` corto el segundo montaje
    // **reutiliza** el documento recién traído en lugar de repetir la petición.
    // Las mutaciones del panel invalidan la clave (T722), así que publicar sigue
    // viéndose al instante en la primera carga posterior.
    staleTime: 30_000,
  });

  const [retained, setRetained] = useState<PortadaOfrendas>();
  if (query.data && query.data !== retained) {
    setRetained(query.data);
  }

  return { ...query, data: query.data ?? retained };
}

/**
 * `true` si el documento del bloque trae al menos un método publicado
 * (FR-022/SC-014). Con 0 métodos publicados el servidor omite `items` y el bloque
 * **no se dibuja** —aunque haya textos publicados: el bloque solo mira métodos—.
 */
export function hasPortadaOfrendasContent(data: PortadaOfrendas | undefined): boolean {
  return (data?.items?.length ?? 0) > 0;
}
