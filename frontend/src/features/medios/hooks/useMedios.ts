import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { ApiError } from '../../../api/client';
import {
  listMedios,
  type MarcaRevision,
  type MedioAdminList,
  type Visibilidad,
} from '../../../api/medios';
import {
  MEDIOS_ADMIN_QUERY_KEY,
  MEDIOS_PAGE_LIMIT,
  PORTADA_MEDIOS_QUERY_KEY,
  PUBLIC_MEDIOS_QUERY_KEY,
} from './queryKeys';

export { MEDIOS_ADMIN_QUERY_KEY, MEDIOS_PAGE_LIMIT };

export interface UseMediosOptions {
  /** Filtro por visibilidad (FR-015); sin valor: todas. */
  visibility?: Visibilidad;
  /** Filtro por la marca **pendiente** de esa clase (FR-014/FR-017); sin valor: todas. */
  reviewMark?: MarcaRevision;
}

/**
 * Listado del panel de medios (FR-019): **todo** el catálogo importado —visibles,
 * ocultos y marcados— en el mismo orden que ve el visitante, con el origen de
 * cada entrada (`origin` y `createdAt`, FR-019), las `sources` de cada enlace y
 * `lastSync` si ya se ejecutó alguna búsqueda (FR-018). Sin paginar: pide el tope
 * del contrato.
 */
export function useMedios(options: UseMediosOptions = {}) {
  const { visibility, reviewMark } = options;

  const query = useQuery<MedioAdminList, ApiError>({
    queryKey: [...MEDIOS_ADMIN_QUERY_KEY, visibility ?? 'all', reviewMark ?? 'all'],
    queryFn: () => listMedios({ limit: MEDIOS_PAGE_LIMIT, visibility, reviewMark }),
  });

  // Última versión **válida** vista (ux.md §5.b: «error con caché previa»): el
  // panel conserva el último listado y añade un aviso discreto en lugar de
  // vaciarse. Ajuste de estado durante el render, sin efecto ni cascada.
  const [retained, setRetained] = useState<MedioAdminList>();
  if (query.data && query.data !== retained) {
    setRetained(query.data);
  }

  return { ...query, data: query.data ?? retained };
}

/**
 * Invalida el listado del panel **y** las dos superficies públicas (sección
 * `/medios` y bloque de la portada) tras cada mutación: lo guardado se publica al
 * instante (FR-015) y lo importado entra en su lugar por fecha (SC-003).
 */
export function useInvalidateMedios() {
  const queryClient = useQueryClient();
  return () => {
    void queryClient.invalidateQueries({ queryKey: MEDIOS_ADMIN_QUERY_KEY });
    void queryClient.invalidateQueries({ queryKey: PUBLIC_MEDIOS_QUERY_KEY });
    void queryClient.invalidateQueries({ queryKey: PORTADA_MEDIOS_QUERY_KEY });
  };
}
