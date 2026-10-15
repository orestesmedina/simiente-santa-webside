import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { ApiError } from '../../../../api/client';
import { getOfrendas, type OfrendasPublicas } from '../../../../api/ofrendas';
import { useLanguage } from '../../i18n/useLanguage';

/**
 * Clave raíz de la consulta de la sección pública (una entrada por idioma). La
 * comparten la página `/ofrendas` (T726) y la entrada de navegación «Ofrendas y
 * diezmos» (`useSiteLinks`, T727): la sección —y su entrada— existe si hay ≥1
 * método **o** ≥1 texto publicado (P7-8/P7-13). Las mutaciones del panel
 * invalidan esta clave para que el cambio se vea en la primera carga posterior
 * (SC-003).
 */
export const PUBLIC_OFRENDAS_QUERY_KEY = ['ofrendas'] as const;

/**
 * Sección pública `/ofrendas` (FR-002/FR-004): el servidor entrega solo lo
 * **publicado**, en el orden de la sección (`iban → sinpe`, `sortOrder`, `id`) y
 * ya resuelto al idioma activo (fallback `en → es` por campo). El cliente solo
 * presenta: **no** filtra, **no** reordena y **no** reimplementa el fallback ni
 * la regla de borradores. Con 0 publicados las claves `items`/`intro`/`usage`
 * vienen omitidas y la página decide redirigir a la portada (P7-21).
 *
 * Al cambiar de idioma se conserva la respuesta anterior
 * (`placeholderData: keepPreviousData`) y, si la recarga falla, se mantiene la
 * última versión válida (ux.md §4.1: «error con caché previa»).
 */
export function usePublicOfrendas() {
  const { lang } = useLanguage();

  const query = useQuery<OfrendasPublicas, ApiError>({
    queryKey: [...PUBLIC_OFRENDAS_QUERY_KEY, lang],
    queryFn: () => getOfrendas(lang),
    placeholderData: keepPreviousData,
  });

  // Última versión **válida** vista (mismo patrón que los hooks públicos de
  // F3/F4/F6): ajuste de estado durante el render, sin efecto ni render en cascada.
  const [retained, setRetained] = useState<OfrendasPublicas>();
  if (query.data && query.data !== retained) {
    setRetained(query.data);
  }

  return { ...query, data: query.data ?? retained };
}

/**
 * `true` si el documento de la sección trae **algo publicado**: al menos un
 * método (`items`) **o** un texto (`intro`/`usage`). Es la presencia en la
 * navegación y la condición de «la sección vive» (P7-8/P7-13, FR-005): una
 * sección con solo textos publicados y 0 métodos sigue existiendo.
 */
export function hasPublicOfrendasContent(data: OfrendasPublicas | undefined): boolean {
  return (data?.items?.length ?? 0) > 0 || Boolean(data?.intro) || Boolean(data?.usage);
}
