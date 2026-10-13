import { useQuery, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import {
  listMinisterios,
  type MinisterioAdminList,
  type PublicationState,
} from '../../../api/ministerios';
import { PORTADA_MINISTERIOS_QUERY_KEY } from '../../publico/ministerios/hooks/usePortadaMinisterios';
import { PUBLIC_MINISTRIES_QUERY_KEY } from '../../publico/ministerios/hooks/usePublicMinistries';

/**
 * Clave del listado del panel (por filtro de estado). Comparte el prefijo
 * `ministerios` con la sección pública: las mutaciones invalidan las tres
 * superficies (panel, sección y bloque) para que el cambio sea visible en la
 * primera carga posterior (SC-003).
 */
export const MINISTRIES_ADMIN_QUERY_KEY = ['ministerios', 'admin'] as const;

/**
 * Tope de página del panel. El contrato admite `limit` hasta 100 (§8.1.3) y la
 * UX no pagina el listado (ux.md §4.3): se pide el tope para mostrar «todos» los
 * ministerios en el orden manual (`sortOrder`, `id`) sin dejar ninguno fuera.
 */
export const MINISTRIES_PAGE_LIMIT = 100;

export interface UseMinistriesOptions {
  /** Filtro por estado de publicación (FR-012); sin valor: todos. */
  publicationState?: PublicationState;
}

/**
 * Listado del panel de ministerios (FR-010): ambos idiomas, la ficha del
 * encargado **íntegra** aunque no esté publicada (dato guardado ≠ dato
 * publicado) y `whatsappUrlShared` como aviso. Sin paginar: pide el tope.
 */
export function useMinistries(options: UseMinistriesOptions = {}) {
  const { publicationState } = options;

  return useQuery<MinisterioAdminList, ApiError>({
    queryKey: [...MINISTRIES_ADMIN_QUERY_KEY, publicationState ?? 'all'],
    queryFn: () => listMinisterios({ limit: MINISTRIES_PAGE_LIMIT, publicationState }),
  });
}

/**
 * Invalida el listado del panel **y** las dos superficies públicas (sección y
 * bloque de la portada) tras cada mutación: «se publica al guardar» (FR-013) y
 * el orden manual es el que ve el visitante (FR-016/SC-003).
 */
export function useInvalidateMinistries() {
  const queryClient = useQueryClient();
  return () => {
    void queryClient.invalidateQueries({ queryKey: MINISTRIES_ADMIN_QUERY_KEY });
    void queryClient.invalidateQueries({ queryKey: PUBLIC_MINISTRIES_QUERY_KEY });
    void queryClient.invalidateQueries({ queryKey: PORTADA_MINISTERIOS_QUERY_KEY });
  };
}
