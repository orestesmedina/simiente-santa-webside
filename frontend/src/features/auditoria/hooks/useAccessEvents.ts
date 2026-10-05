import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { listAccessEvents, type AccessEventList } from '../../../api/auditoria';

/** Clave raíz de TanStack Query del historial de accesos. */
export const ACCESS_EVENTS_QUERY_KEY = ['audit', 'access'] as const;

/** Tamaño de página por defecto de los historiales (coincide con el contrato). */
export const AUDIT_PAGE_SIZE = 20;

export interface UseAccessEventsParams {
  /** Filtro por cuenta (FR-024). */
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
 * Historial paginado de intentos de inicio de sesión (FR-022/FR-024). La clave
 * incluye filtros y página: al cambiarlos, la consulta se refresca y la
 * `placeholderData` mantiene la página anterior para no parpadear.
 */
export function useAccessEvents({
  userId,
  from,
  to,
  page = 1,
  limit = AUDIT_PAGE_SIZE,
  enabled = true,
}: UseAccessEventsParams = {}) {
  const offset = Math.max(0, (page - 1) * limit);

  return useQuery<AccessEventList, ApiError>({
    queryKey: [...ACCESS_EVENTS_QUERY_KEY, { userId, from, to, limit, offset }],
    queryFn: () => listAccessEvents({ userId, from, to, limit, offset }),
    placeholderData: keepPreviousData,
    enabled,
  });
}
