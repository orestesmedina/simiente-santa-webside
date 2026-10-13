import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { deleteMinisterio } from '../../../api/ministerios';
import { useInvalidateMinistries } from './useMinistries';

/** Elimina un ministerio (FR-010); queda registrado en la auditoría (FR-017). */
export function useDeleteMinistry() {
  const invalidate = useInvalidateMinistries();

  return useMutation<void, ApiError, string>({
    mutationFn: (id) => deleteMinisterio(id),
    onSuccess: invalidate,
  });
}
