import { useQuery } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { listRoles, type RoleItem } from '../../../api/roles';

/**
 * Clave de TanStack Query para el catálogo de roles que alimenta el selector
 * del formulario de cuenta. Independiente de la clave del listado de roles
 * (T249) para que ambas cachés no interfieran.
 */
export const ROLE_OPTIONS_QUERY_KEY = ['roles', 'options'] as const;

/** Roles disponibles para asignar a una cuenta (FR-009). */
export function useRoleOptions() {
  return useQuery<RoleItem[], ApiError>({
    queryKey: ROLE_OPTIONS_QUERY_KEY,
    queryFn: async () => (await listRoles({ limit: 100 })).items,
    staleTime: 30_000,
  });
}
