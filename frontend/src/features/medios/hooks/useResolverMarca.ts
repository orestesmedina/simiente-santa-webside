import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { resolveMedioMark, type MedioAdmin, type ResolverMarcaInput } from '../../../api/medios';
import { useInvalidateMedios } from './useMedios';

/**
 * Resuelve la marca de revisión indicada en `mark` y aplica la visibilidad
 * elegida (FR-017/US3 esc. 6). Los dos ejes —duplicados y disponibilidad— se
 * resuelven por separado; «posible duplicado» resuelto como *no son el mismo
 * sermón* deja **ambas** fichas visibles y sin marca. Una marca resuelta no se
 * reabre por la misma causa.
 */
export function useResolverMarca() {
  const invalidate = useInvalidateMedios();

  return useMutation<MedioAdmin, ApiError, { id: string; input: ResolverMarcaInput }>({
    mutationFn: ({ id, input }) => resolveMedioMark(id, input),
    onSuccess: invalidate,
  });
}
