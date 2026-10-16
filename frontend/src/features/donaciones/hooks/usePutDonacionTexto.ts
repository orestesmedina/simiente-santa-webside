import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import {
  putDonacionTexto,
  type DonacionTextoAdmin,
  type DonacionTextoInput,
  type DonacionTextKey,
} from '../../../api/ofrendas';
import { useInvalidateDonaciones } from './useDonaciones';

/**
 * Guarda (crea o reemplaza) un texto de la sección (FR-021): `textEs`
 * obligatorio, `textEn` opcional y el estado propio del texto. Un cambio sobre
 * un texto publicado es visible de inmediato (FR-013, mismo que los métodos).
 */
export function usePutDonacionTexto() {
  const invalidate = useInvalidateDonaciones();

  return useMutation<
    DonacionTextoAdmin,
    ApiError,
    { textKey: DonacionTextKey; input: DonacionTextoInput }
  >({
    mutationFn: ({ textKey, input }) => putDonacionTexto(textKey, input),
    onSuccess: invalidate,
  });
}
