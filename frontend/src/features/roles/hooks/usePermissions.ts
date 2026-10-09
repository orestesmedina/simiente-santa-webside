import { useQuery } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { listPermissions, type PermissionList } from '../../../api/roles';

/** Clave de TanStack Query del catálogo fijo de permisos (FR-015). */
export const PERMISSIONS_QUERY_KEY = ['permissions'] as const;

/**
 * Catálogo fijo de permisos por módulo (FR-015). Cambia muy poco, así que se
 * mantiene fresco durante varios minutos.
 */
export function usePermissions() {
  return useQuery<PermissionList, ApiError>({
    queryKey: PERMISSIONS_QUERY_KEY,
    queryFn: listPermissions,
    staleTime: 5 * 60_000,
  });
}
