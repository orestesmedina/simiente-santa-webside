import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { updateRole, type RoleItem, type RoleUpdateInput } from '../../../api/roles';
import { ROLES_QUERY_KEY } from './useRoles';

export interface UpdateRoleVariables {
  id: string;
  input: RoleUpdateInput;
}

/**
 * Edita nombre y/o permisos de un rol (FR-017/FR-018). Los cambios de permisos
 * surten efecto sin pasos adicionales en el servidor (FR-018/SC-009); al
 * terminar, invalida el listado para reflejarlos.
 */
export function useUpdateRole() {
  const queryClient = useQueryClient();

  return useMutation<RoleItem, ApiError, UpdateRoleVariables>({
    mutationFn: ({ id, input }) => updateRole(id, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ROLES_QUERY_KEY });
    },
  });
}
