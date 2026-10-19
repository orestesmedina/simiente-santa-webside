import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { ApiError } from '../../../../api/client';
import { MEDIOS_PAGE_LIMIT, getMedios, type MediosPublicos } from '../../../../api/medios';
import { useLanguage } from '../../i18n/useLanguage';

/**
 * Clave raíz de la consulta de la sección pública (una entrada por idioma, tanda
 * y desplazamiento). La comparten la página `/medios` (T933) y la entrada de
 * navegación «Medios» (`useSiteLinks`, T934): la sección —y su entrada— existe
 * solo si el documento trae `total > 0` (P9-16/P7-8). Las mutaciones del panel
 * invalidan esta clave para que el cambio se vea en la primera carga posterior
 * (SC-003/FR-015).
 */
export const PUBLIC_MEDIOS_QUERY_KEY = ['medios'] as const;

/** Desplazamiento de la sección: la lista siempre empieza por el más reciente. */
const PUBLIC_MEDIOS_OFFSET = 0;

/**
 * Sección pública `/medios` (FR-002/FR-006): el servidor entrega solo lo
 * **publicable** —visibles, no absorbidas y sin par de duplicados pendiente
 * (FR-009)— del más reciente al más antiguo, ya resuelto al idioma activo
 * (fallback `en → es` solo en la descripción, FR-025). El cliente solo presenta:
 * **no** filtra, **no** reordena y **no** reimplementa el fallback ni la regla de
 * publicables.
 *
 * «Cargar más» (Q11/D-3) **agranda la ventana** de la misma consulta —`limit`
 * crece en tandas de 12— en lugar de mezclar tandas en el cliente: el orden y el
 * contenido siguen saliendo enteros del servidor, sin duplicados posibles y sin
 * que el visitante pierda su lugar. `hasMore` se calcula con `total`.
 *
 * Al cambiar de idioma se conserva la respuesta anterior
 * (`placeholderData: keepPreviousData`) y, si la recarga falla, se mantiene la
 * última versión válida (ux.md §4.1: «error con caché previa»).
 */
export function usePublicMedios() {
  const { lang } = useLanguage();
  const [limit, setLimit] = useState(MEDIOS_PAGE_LIMIT);

  const query = useQuery<MediosPublicos, ApiError>({
    queryKey: [...PUBLIC_MEDIOS_QUERY_KEY, lang, limit, PUBLIC_MEDIOS_OFFSET],
    queryFn: () => getMedios(lang, limit, PUBLIC_MEDIOS_OFFSET),
    placeholderData: keepPreviousData,
  });

  // Última versión **válida** vista (mismo patrón que los hooks públicos de
  // F3/F4/F6/F7): ajuste de estado durante el render, sin efecto ni cascada.
  const [retained, setRetained] = useState<MediosPublicos>();
  if (query.data && query.data !== retained) {
    setRetained(query.data);
  }
  const data = query.data ?? retained;

  return {
    ...query,
    data,
    /** `true` si el servidor todavía tiene fichas publicables sin mostrar (D-3). */
    hasMore: (data?.items?.length ?? 0) < (data?.total ?? 0),
    /** Agrega la siguiente tanda (12 fichas) al final de la misma lista (Q11/D-3). */
    loadMore: () => setLimit((current) => current + MEDIOS_PAGE_LIMIT),
  };
}

/**
 * `true` si el documento de la sección trae **algo publicable** (`total > 0`): es
 * la presencia en la navegación y la condición de «la sección vive» (P9-16,
 * FR-008). Con 0 publicables la entrada no existe y `/medios` redirige a la
 * portada.
 */
export function hasPublicMediosContent(data: MediosPublicos | undefined): boolean {
  return (data?.total ?? 0) > 0;
}
