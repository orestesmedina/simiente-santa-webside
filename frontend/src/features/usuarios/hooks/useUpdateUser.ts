import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { updateUser, type UpdateUserInput, type UserItem } from '../../../api/usuarios';
import { USERS_QUERY_KEY } from './useUsers';

export interface UpdateUserVariables {
  id: string;
  input: UpdateUserInput;
}

/**
 * Edita los datos, el rol o el estado de una cuenta (FR-011/FR-012). Al
 * terminar, invalida el listado para que la fila refleje el cambio; el servidor
 * sigue siendo la autoridad del resultado (FR-016).
 */
export function useUpdateUser() {
  const queryClient = useQueryClient();

  return useMutation<UserItem, ApiError, UpdateUserVariables>({
    mutationFn: ({ id, input }) => updateUser(id, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: USERS_QUERY_KEY });
    },
  });
}
