import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { listRoles, type RoleList } from '../../../api/roles';

/**
 * Clave raíz de TanStack Query para roles. Al invalidarla se refrescan tanto el
 * listado (`['roles', {…}]`) como el catálogo de opciones del formulario de
 * cuenta (`['roles', 'options']`, en `features/usuarios`), porque TanStack
 * Query invalida por prefijo.
 */
export const ROLES_QUERY_KEY = ['roles'] as const;

/** Tamaño de página por defecto del listado (coincide con el defecto del contrato). */
export const ROLES_PAGE_SIZE = 20;

export interface UseRolesParams {
  /** Página 1-based. */
  page?: number;
  limit?: number;
}

/**
 * Listado paginado de roles con sus permisos y cuentas asignadas (FR-017) con
 * `limit`/`offset`. Conserva la página anterior mientras llega la nueva para que
 * la tabla no parpadee al paginar (ux.md §4.d).
 */
export function useRoles({ page = 1, limit = ROLES_PAGE_SIZE }: UseRolesParams = {}) {
  const offset = Math.max(0, (page - 1) * limit);

  return useQuery<RoleList, ApiError>({
    queryKey: [...ROLES_QUERY_KEY, { limit, offset }],
    queryFn: () => listRoles({ limit, offset }),
    placeholderData: keepPreviousData,
  });
}
