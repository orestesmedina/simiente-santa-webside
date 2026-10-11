import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { createSocialLink, type SocialLinkAdmin, type SocialLinkInput } from '../../../api/portada';
import { PORTADA_ADMIN_QUERY_KEY } from './usePortadaAdmin';

/** Crea el enlace de una red (FR-006); un segundo enlace de la misma red → 409. */
export function useCreateSocialLink() {
  const queryClient = useQueryClient();

  return useMutation<SocialLinkAdmin, ApiError, SocialLinkInput>({
    mutationFn: (input) => createSocialLink(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: PORTADA_ADMIN_QUERY_KEY });
    },
  });
}
