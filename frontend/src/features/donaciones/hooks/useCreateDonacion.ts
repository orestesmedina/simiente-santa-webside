import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { createDonacion, type DonacionAdmin, type DonacionInput } from '../../../api/ofrendas';
import { useInvalidateDonaciones } from './useDonaciones';

/**
 * Alta de un método de donación (FR-010/FR-014). El input lleva el número tal
 * como lo escribe el equipo (IBAN con espacios/guiones, SINPE con o sin `+506`)
 * y el servidor lo normaliza y valida (FR-015/FR-016). Tras el alta invalida
 * panel, textos, sección y bloque.
 */
export function useCreateDonacion() {
  const invalidate = useInvalidateDonaciones();

  return useMutation<DonacionAdmin, ApiError, DonacionInput>({
    mutationFn: (input) => createDonacion(input),
    onSuccess: invalidate,
  });
}
