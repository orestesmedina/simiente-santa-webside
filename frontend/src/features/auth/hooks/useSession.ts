import { useQuery } from '@tanstack/react-query';
import { getSession, type SessionUser } from '../../../api/auth';
import { ApiError } from '../../../api/client';

/** Clave única de TanStack Query para la sesión (compartida por guards y panel). */
export const SESSION_QUERY_KEY = ['session'] as const;

/**
 * Consulta `GET /auth/session` y la cachea. Todos los guards, el layout del
 * panel y los hooks de auth comparten esta clave, así que la sesión se pide una
 * sola vez por navegación. Sin reintentos: un `401` debe resolverse de inmediato
 * en `/login`.
 */
export function useSessionQuery() {
  return useQuery<SessionUser, ApiError>({
    queryKey: SESSION_QUERY_KEY,
    queryFn: getSession,
    retry: false,
    staleTime: 30_000,
  });
}

/** `true` cuando el fallo es un `401` (sesión inexistente, expirada o cerrada). */
export function isUnauthenticated(error: unknown): boolean {
  return error instanceof ApiError && error.status === 401;
}
