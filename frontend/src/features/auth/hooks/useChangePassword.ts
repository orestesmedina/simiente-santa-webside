import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  changeMyPassword,
  type ChangePasswordInput,
  type PasswordChangedResponse,
  type SessionUser,
} from '../../../api/auth';
import { ApiError } from '../../../api/client';
import { SESSION_QUERY_KEY } from './useSession';

/**
 * Cambia la contraseña de la propia cuenta (FR-020). Al completarse, marca la
 * sesión cacheada como sin cambio pendiente (`mustChangePassword = false`)
 * para que `RequirePasswordChange` deje pasar al panel (US7 esc. 4). Revoca las
 * demás sesiones en el servidor; la actual sigue viva.
 */
export function useChangePassword() {
  const queryClient = useQueryClient();

  return useMutation<PasswordChangedResponse, ApiError, ChangePasswordInput>({
    mutationFn: (input) => changeMyPassword(input),
    onSuccess: () => {
      queryClient.setQueryData<SessionUser | undefined>(SESSION_QUERY_KEY, (session) =>
        session ? { ...session, mustChangePassword: false } : session,
      );
    },
  });
}
