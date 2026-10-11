import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { updateAbout, type AboutAdmin, type AboutInput } from '../../../api/portada';
import { PORTADA_ADMIN_QUERY_KEY } from './usePortadaAdmin';

/** Guarda «quiénes somos» (FR-003) y refresca el agregado del panel. */
export function useUpdateWhoWeAre() {
  const queryClient = useQueryClient();

  return useMutation<AboutAdmin, ApiError, AboutInput>({
    mutationFn: (input) => updateAbout(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: PORTADA_ADMIN_QUERY_KEY });
    },
  });
}
