import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import {
  updateService,
  type ScheduleItemAdmin,
  type ScheduleItemPatch,
} from '../../../api/portada';
import { PORTADA_ADMIN_QUERY_KEY } from './usePortadaAdmin';

/** Edita un servicio y/o lo publica o retira (FR-004/FR-013). */
export function useUpdateService() {
  const queryClient = useQueryClient();

  return useMutation<ScheduleItemAdmin, ApiError, { id: string; input: ScheduleItemPatch }>({
    mutationFn: ({ id, input }) => updateService(id, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: PORTADA_ADMIN_QUERY_KEY });
    },
  });
}
