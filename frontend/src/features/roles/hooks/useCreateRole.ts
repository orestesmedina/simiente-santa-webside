import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { createRole, type RoleCreateInput, type RoleItem } from '../../../api/roles';
import { ROLES_QUERY_KEY } from './useRoles';

/**
 * Crea un rol con al menos un permiso (FR-014). Al terminar invalida todos los
 * listados de roles (y las opciones del formulario de cuenta) para que la fila
 * nueva aparezca de inmediato.
 */
export function useCreateRole() {
  const queryClient = useQueryClient();

  return useMutation<RoleItem, ApiError, RoleCreateInput>({
    mutationFn: (input) => createRole(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ROLES_QUERY_KEY });
    },
  });
}
