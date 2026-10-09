import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { deleteWhatsappChannel } from '../../../api/portada';
import { PORTADA_ADMIN_QUERY_KEY } from './usePortadaAdmin';

/** Elimina un canal de WhatsApp (FR-005). */
export function useDeleteWhatsappChannel() {
  const queryClient = useQueryClient();

  return useMutation<void, ApiError, string>({
    mutationFn: (id) => deleteWhatsappChannel(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: PORTADA_ADMIN_QUERY_KEY });
    },
  });
}
