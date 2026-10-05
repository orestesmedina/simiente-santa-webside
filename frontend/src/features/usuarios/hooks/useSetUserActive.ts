import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { updateUser, type UserItem } from '../../../api/usuarios';
import { USERS_QUERY_KEY } from './useUsers';

export interface SetUserActiveVariables {
  user: UserItem;
  /** `false` desactiva (FR-012); `true` reactiva (US5 esc. 3). */
  isActive: boolean;
}

/**
 * Activa o desactiva una cuenta (FR-012/FR-013). Desactivar bloquea el acceso
 * de inmediato y revoca sus sesiones; los datos se conservan siempre. Al
 * terminar invalida el listado para reflejar el estado nuevo.
 */
export function useSetUserActive() {
  const queryClient = useQueryClient();

  return useMutation<UserItem, ApiError, SetUserActiveVariables>({
    mutationFn: ({ user, isActive }) => updateUser(user.id, { isActive }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: USERS_QUERY_KEY });
    },
  });
}
