import { useMutation, useQueryClient } from '@tanstack/react-query';
import { login, type LoginInput, type SessionUser } from '../../../api/auth';
import { ApiError } from '../../../api/client';
import { SESSION_QUERY_KEY } from './useSession';

/**
 * Inicia sesión (FR-001). Al llegar el `SessionUser` lo publica en la caché de
 * TanStack Query para que los guards y el layout lo compartan sin volver a
 * pedirlo. La navegación posterior la decide la pantalla de acceso.
 */
export function useLogin() {
  const queryClient = useQueryClient();

  return useMutation<SessionUser, ApiError, LoginInput>({
    mutationFn: (input) => login(input),
    onSuccess: (session) => {
      queryClient.setQueryData(SESSION_QUERY_KEY, session);
    },
  });
}
