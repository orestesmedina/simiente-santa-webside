import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import {
  createMinisterio,
  type MinisterioAdmin,
  type MinisterioInput,
} from '../../../api/ministerios';
import { useInvalidateMinistries } from './useMinistries';

/**
 * Alta de ministerio (FR-010/FR-014). `leaderVisible` viaja explícitamente con
 * la decisión del equipo: el formulario lo envía en `false` por defecto
 * («Oculto por defecto», D-13). Tras el alta invalida panel, sección y bloque.
 */
export function useCreateMinistry() {
  const invalidate = useInvalidateMinistries();

  return useMutation<MinisterioAdmin, ApiError, MinisterioInput>({
    mutationFn: (input) => createMinisterio(input),
    onSuccess: invalidate,
  });
}
