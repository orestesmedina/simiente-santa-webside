import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import {
  resetUserPassword,
  type PasswordResetResponse,
  type ResetPasswordInput,
} from '../../../api/usuarios';
import { USERS_QUERY_KEY } from './useUsers';

export interface ResetPasswordVariables {
  id: string;
  input: ResetPasswordInput;
}

/**
 * Restablece la contraseña de una cuenta (FR-010): su titular deberá cambiarla
 * al entrar (`mustChangePassword = true`) y sus sesiones abiertas se revocan.
 * Nunca devuelve ni guarda el valor de la contraseña (FR-026).
 */
export function useResetPassword() {
  const queryClient = useQueryClient();

  return useMutation<PasswordResetResponse, ApiError, ResetPasswordVariables>({
    mutationFn: ({ id, input }) => resetUserPassword(id, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: USERS_QUERY_KEY });
    },
  });
}
