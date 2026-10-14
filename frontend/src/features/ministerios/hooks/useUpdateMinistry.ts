import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import {
  updateMinisterio,
  type MinisterioAdmin,
  type MinisterioPatch,
} from '../../../api/ministerios';
import { useInvalidateMinistries } from './useMinistries';

/**
 * Edición parcial de un ministerio (FR-013/FR-016): un `PATCH` con los campos
 * presentes (o `publicationState`/`leaderVisible`) y el cambio visible en la
 * primera carga posterior. Cubre publicar/retirar y el interruptor de la ficha
 * del encargado (US2 esc. 10–11).
 */
export function useUpdateMinistry() {
  const invalidate = useInvalidateMinistries();

  return useMutation<MinisterioAdmin, ApiError, { id: string; input: MinisterioPatch }>({
    mutationFn: ({ id, input }) => updateMinisterio(id, input),
    onSuccess: invalidate,
  });
}
