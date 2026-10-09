import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { listAdminActions, type AdminActionList } from '../../../api/auditoria';

/** Clave raíz de TanStack Query del historial de acciones administrativas. */
export const ADMIN_ACTIONS_QUERY_KEY = ['audit', 'actions'] as const;

export interface UseAdminActionsParams {
  /** Filtro por cuenta involucrada (FR-024). */
  userId?: string;
  /** Inicio del rango, ISO-8601 inclusive. */
  from?: string;
  /** Fin del rango, ISO-8601 exclusivo. */
  to?: string;
  /** Página 1-based. */
  page?: number;
  limit?: number;
  /** Permite no consultar la pestaña inactiva. */
  enabled?: boolean;
}

/**
 * Historial paginado de acciones administrativas sensibles (FR-023/FR-024).
 * **Solo lectura**: no existe ninguna operación de escritura sobre el registro
 * (FR-025).
 */
export function useAdminActions({
  userId,
  from,
  to,
  page = 1,
  limit = 20,
  enabled = true,
}: UseAdminActionsParams = {}) {
  const offset = Math.max(0, (page - 1) * limit);

  return useQuery<AdminActionList, ApiError>({
    queryKey: [...ADMIN_ACTIONS_QUERY_KEY, { userId, from, to, limit, offset }],
    queryFn: () => listAdminActions({ userId, from, to, limit, offset }),
    placeholderData: keepPreviousData,
    enabled,
  });
}
