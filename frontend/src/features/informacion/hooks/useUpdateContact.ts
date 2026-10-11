import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { updateContact, type ContactAdmin, type ContactInput } from '../../../api/portada';
import { PORTADA_ADMIN_QUERY_KEY } from './usePortadaAdmin';

/** Guarda los datos de contacto (FR-007) y refresca el agregado del panel. */
export function useUpdateContact() {
  const queryClient = useQueryClient();

  return useMutation<ContactAdmin, ApiError, ContactInput>({
    mutationFn: (input) => updateContact(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: PORTADA_ADMIN_QUERY_KEY });
    },
  });
}
