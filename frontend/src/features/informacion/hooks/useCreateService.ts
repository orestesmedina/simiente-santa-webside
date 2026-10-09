import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import {
  createService,
  type ScheduleItemAdmin,
  type ScheduleItemInput,
} from '../../../api/portada';
import { PORTADA_ADMIN_QUERY_KEY } from './usePortadaAdmin';

/** Crea un servicio del horario (FR-004); nace en borrador salvo estado explícito. */
export function useCreateService() {
  const queryClient = useQueryClient();

  return useMutation<ScheduleItemAdmin, ApiError, ScheduleItemInput>({
    mutationFn: (input) => createService(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: PORTADA_ADMIN_QUERY_KEY });
    },
  });
}
