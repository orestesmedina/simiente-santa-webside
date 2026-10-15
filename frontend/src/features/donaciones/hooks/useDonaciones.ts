import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { ApiError } from '../../../api/client';
import {
  listDonaciones,
  type DonacionAdminList,
  type PublicationState,
} from '../../../api/ofrendas';
import {
  DONACIONES_ADMIN_QUERY_KEY,
  DONACIONES_PAGE_LIMIT,
  DONACIONES_TEXTOS_QUERY_KEY,
  PORTADA_OFRENDAS_QUERY_KEY,
  PUBLIC_OFRENDAS_QUERY_KEY,
} from './queryKeys';

export { DONACIONES_ADMIN_QUERY_KEY, DONACIONES_PAGE_LIMIT };

export interface UseDonacionesOptions {
  /** Filtro por estado de publicación (FR-012); sin valor: todos. */
  publicationState?: PublicationState;
}

/**
 * Listado del panel de donaciones (FR-010): ambos idiomas del titular, el estado
 * y el número **normalizado** (FR-015), en el mismo orden que ve el visitante
 * (`iban` primero; `sortOrder`, `id` en cada grupo). Sin paginar: pide el tope.
 */
export function useDonaciones(options: UseDonacionesOptions = {}) {
  const { publicationState } = options;

  const query = useQuery<DonacionAdminList, ApiError>({
    queryKey: [...DONACIONES_ADMIN_QUERY_KEY, publicationState ?? 'all'],
    queryFn: () => listDonaciones({ limit: DONACIONES_PAGE_LIMIT, publicationState }),
  });

  // Última versión **válida** vista (ux.md §5.b: «error con contenido en caché»):
  // el panel conserva el último listado y añade un aviso discreto en lugar de
  // vaciarse. Ajuste de estado durante el render, sin efecto ni cascada.
  const [retained, setRetained] = useState<DonacionAdminList>();
  if (query.data && query.data !== retained) {
    setRetained(query.data);
  }

  return { ...query, data: query.data ?? retained };
}

/**
 * Invalida el listado del panel, la zona de textos **y** las dos superficies
 * públicas (sección `/ofrendas` y bloque de la portada) tras cada mutación: «se
 * publica al guardar» (FR-013) y el orden manual es el que ve el visitante
 * (FR-004/SC-003).
 */
export function useInvalidateDonaciones() {
  const queryClient = useQueryClient();
  return () => {
    void queryClient.invalidateQueries({ queryKey: DONACIONES_ADMIN_QUERY_KEY });
    void queryClient.invalidateQueries({ queryKey: DONACIONES_TEXTOS_QUERY_KEY });
    void queryClient.invalidateQueries({ queryKey: PUBLIC_OFRENDAS_QUERY_KEY });
    void queryClient.invalidateQueries({ queryKey: PORTADA_OFRENDAS_QUERY_KEY });
  };
}
