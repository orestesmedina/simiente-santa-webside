import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import {
  updateWhatsappChannel,
  type WhatsappChannelAdmin,
  type WhatsappChannelPatch,
} from '../../../api/portada';
import { PORTADA_ADMIN_QUERY_KEY } from './usePortadaAdmin';

/** Edita un canal y/o lo publica o retira (FR-005/FR-013). */
export function useUpdateWhatsappChannel() {
  const queryClient = useQueryClient();

  return useMutation<WhatsappChannelAdmin, ApiError, { id: string; input: WhatsappChannelPatch }>({
    mutationFn: ({ id, input }) => updateWhatsappChannel(id, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: PORTADA_ADMIN_QUERY_KEY });
    },
  });
}
