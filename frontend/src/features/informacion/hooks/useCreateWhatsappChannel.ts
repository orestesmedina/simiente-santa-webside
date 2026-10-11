import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import {
  createWhatsappChannel,
  type WhatsappChannelAdmin,
  type WhatsappChannelInput,
} from '../../../api/portada';
import { PORTADA_ADMIN_QUERY_KEY } from './usePortadaAdmin';

/** Crea un canal de WhatsApp (FR-005). */
export function useCreateWhatsappChannel() {
  const queryClient = useQueryClient();

  return useMutation<WhatsappChannelAdmin, ApiError, WhatsappChannelInput>({
    mutationFn: (input) => createWhatsappChannel(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: PORTADA_ADMIN_QUERY_KEY });
    },
  });
}
