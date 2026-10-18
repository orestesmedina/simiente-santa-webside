import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { updateMedio, type MedioAdmin, type Visibilidad } from '../../../api/medios';
import { useInvalidateMedios } from './useMedios';

/**
 * Oculta o muestra una ficha (FR-015/D-8): un `PATCH` con **solo** `visibility`.
 * Ocultar es directo y reversible; mostrar devuelve el episodio a su lugar por
 * fecha (nunca al final). La sincronización posterior **no** cambia esta decisión
 * salvo una indisponibilidad confirmada (FR-017, precisión de la fase 5).
 */
export function useOcultarMostrarEpisodio() {
  const invalidate = useInvalidateMedios();

  return useMutation<MedioAdmin, ApiError, { id: string; visibility: Visibilidad }>({
    mutationFn: ({ id, visibility }) => updateMedio(id, { visibility }),
    onSuccess: invalidate,
  });
}
