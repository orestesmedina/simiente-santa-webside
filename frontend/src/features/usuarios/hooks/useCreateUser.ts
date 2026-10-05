import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { createUser, type CreateUserInput, type UserItem } from '../../../api/usuarios';
import { USERS_QUERY_KEY } from './useUsers';

/**
 * Crea una cuenta con su contraseña inicial (FR-009/FR-010). Al terminar,
 * refresca el listado para que la fila nueva aparezca arriba (orden
 * `createdAt DESC` del contrato).
 */
export function useCreateUser() {
  const queryClient = useQueryClient();

  return useMutation<UserItem, ApiError, CreateUserInput>({
    mutationFn: (input) => createUser(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: USERS_QUERY_KEY });
    },
  });
}
