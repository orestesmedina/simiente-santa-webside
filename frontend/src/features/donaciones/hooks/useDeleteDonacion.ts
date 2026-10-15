import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { deleteDonacion } from '../../../api/ofrendas';
import { useInvalidateDonaciones } from './useDonaciones';

/** Elimina un método (FR-010); queda registrado en la auditoría (FR-017). */
export function useDeleteDonacion() {
  const invalidate = useInvalidateDonaciones();

  return useMutation<void, ApiError, string>({
    mutationFn: (id) => deleteDonacion(id),
    onSuccess: invalidate,
  });
}
