import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { deleteRole, type RoleDeletedResponse } from '../../../api/roles';
import { ROLES_QUERY_KEY } from './useRoles';

export interface DeleteRoleVariables {
  id: string;
}

/**
 * Elimina un rol **solo cuando no está en uso** (FR-017): el servidor devuelve
 * `409` si tiene cuentas asignadas. Al terminar invalida el listado.
 */
export function useDeleteRole() {
  const queryClient = useQueryClient();

  return useMutation<RoleDeletedResponse, ApiError, DeleteRoleVariables>({
    mutationFn: ({ id }) => deleteRole(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ROLES_QUERY_KEY });
    },
  });
}
