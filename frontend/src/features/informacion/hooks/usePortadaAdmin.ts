import { useQuery } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { getPortadaAdmin, type PortadaAdmin } from '../../../api/portada';

/** Clave raíz de la consulta del agregado del panel (FR-011). */
export const PORTADA_ADMIN_QUERY_KEY = ['portada', 'admin'] as const;

/**
 * Estado completo del módulo para el panel (`GET /api/v1/admin/portada`,
 * T340/FR-011): ambos idiomas y `publicationState` por elemento, incluidos los
 * borradores. Es la única llamada con la que el módulo precarga todas las
 * pestañas; cada mutación la invalida para refrescar píldoras y formularios.
 */
export function usePortadaAdmin() {
  return useQuery<PortadaAdmin, ApiError>({
    queryKey: PORTADA_ADMIN_QUERY_KEY,
    queryFn: getPortadaAdmin,
  });
}
