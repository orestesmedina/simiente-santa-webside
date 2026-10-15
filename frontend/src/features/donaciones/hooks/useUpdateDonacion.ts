import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { updateDonacion, type DonacionAdmin, type DonacionPatch } from '../../../api/ofrendas';
import { useInvalidateDonaciones } from './useDonaciones';

/**
 * Edición parcial de un método (FR-013/FR-016): un `PATCH` con los campos
 * presentes (datos, `sortOrder` o `publicationState`) y el cambio visible en la
 * primera carga posterior. Cubre publicar/retirar y el reordenamiento dentro del
 * grupo (US2 esc. 10); un `null` explícito borra lo opcional (p. ej. `holderEs`).
 */
export function useUpdateDonacion() {
  const invalidate = useInvalidateDonaciones();

  return useMutation<DonacionAdmin, ApiError, { id: string; input: DonacionPatch }>({
    mutationFn: ({ id, input }) => updateDonacion(id, input),
    onSuccess: invalidate,
  });
}
