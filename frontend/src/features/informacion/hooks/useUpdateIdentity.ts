import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { updateIdentity, type IdentityAdmin, type IdentityInput } from '../../../api/portada';
import { PORTADA_ADMIN_QUERY_KEY } from './usePortadaAdmin';

/**
 * Guarda (crea o reemplaza) la identidad (FR-002) y refresca el agregado del
 * panel para que la píldora de estado y los formularios queden al día.
 */
export function useUpdateIdentity() {
  const queryClient = useQueryClient();

  return useMutation<IdentityAdmin, ApiError, IdentityInput>({
    mutationFn: (input) => updateIdentity(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: PORTADA_ADMIN_QUERY_KEY });
    },
  });
}
