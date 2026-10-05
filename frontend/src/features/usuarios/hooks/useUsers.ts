import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { listUsers, type UserList } from '../../../api/usuarios';

/** Clave raíz de TanStack Query para el listado de cuentas (invalida todas las páginas). */
export const USERS_QUERY_KEY = ['users'] as const;

/** Tamaño de página por defecto del listado (coincide con el defecto del contrato). */
export const USERS_PAGE_SIZE = 20;

export interface UseUsersParams {
  /** Página 1-based. */
  page?: number;
  limit?: number;
}

/**
 * Listado paginado de cuentas (FR-019) con `limit`/`offset`. La clave incluye
 * la página para cachear cada tramo por separado; sin buscador de texto libre
 * (F-04: fuera del MVP).
 */
export function useUsers({ page = 1, limit = USERS_PAGE_SIZE }: UseUsersParams = {}) {
  const offset = Math.max(0, (page - 1) * limit);

  return useQuery<UserList, ApiError>({
    queryKey: [...USERS_QUERY_KEY, { limit, offset }],
    queryFn: () => listUsers({ limit, offset }),
    // Conserva la página anterior mientras llega la nueva: la tabla no parpadea
    // al paginar (ux.md §4.b).
    placeholderData: keepPreviousData,
  });
}
