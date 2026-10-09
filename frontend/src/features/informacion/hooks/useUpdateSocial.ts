import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { updateSocialLink, type SocialLinkAdmin, type SocialLinkPatch } from '../../../api/portada';
import { PORTADA_ADMIN_QUERY_KEY } from './usePortadaAdmin';

/** Edita el enlace de una red y/o lo publica o retira (FR-006/FR-013). */
export function useUpdateSocial() {
  const queryClient = useQueryClient();

  return useMutation<SocialLinkAdmin, ApiError, { id: string; input: SocialLinkPatch }>({
    mutationFn: ({ id, input }) => updateSocialLink(id, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: PORTADA_ADMIN_QUERY_KEY });
    },
  });
}
