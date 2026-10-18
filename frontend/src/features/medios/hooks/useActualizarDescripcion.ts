import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { updateMedio, type MedioAdmin, type MedioPatch } from '../../../api/medios';
import { useInvalidateMedios } from './useMedios';

/**
 * Guarda la descripción corta (español y/o inglés) de una ficha (FR-020/FR-023):
 * `PATCH` con los campos presentes y el cambio visible en la primera carga
 * posterior (FR-015: se publica al guardar). El texto del equipo manda y la
 * sincronización nunca lo pisa (`descriptionEdited`, SC-007).
 */
export function useActualizarDescripcion() {
  const invalidate = useInvalidateMedios();

  return useMutation<MedioAdmin, ApiError, { id: string; input: MedioPatch }>({
    mutationFn: ({ id, input }) => updateMedio(id, input),
    onSuccess: invalidate,
  });
}
