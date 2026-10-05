import { useMutation, useQueryClient } from '@tanstack/react-query';
import { logout, type LogoutResponse } from '../../../api/auth';
import { ApiError } from '../../../api/client';
import { SESSION_QUERY_KEY } from './useSession';

/**
 * Cierra la sesión actual (FR-004). Al terminar, borra la sesión de la caché
 * para que los guards vuelvan a pedirla. Cada pantalla decide a dónde navega
 * después (el panel a `/login`; el aviso de sesión, igualmente a `/login`).
 */
export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation<LogoutResponse, ApiError, void>({
    mutationFn: () => logout(),
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: SESSION_QUERY_KEY });
    },
  });
}
