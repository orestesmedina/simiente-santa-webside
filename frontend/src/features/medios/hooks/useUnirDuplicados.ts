import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { mergeMedio, type MedioAdmin } from '../../../api/medios';
import { useInvalidateMedios } from './useMedios';

/**
 * Une dos entradas del mismo sermón (FR-014/US3 esc. 1): `{id}` se conserva
 * —nombre, carátula, fecha y descripción— y adopta el enlace que faltaba,
 * mientras `otherId` queda absorbida y sale del listado. Es **definitiva**
 * (default revisable #8), así que la interfaz confirma el resultado antes de
 * llamar (`UnirDuplicadosDialog` + `ConfirmDialog`).
 */
export function useUnirDuplicados() {
  const invalidate = useInvalidateMedios();

  return useMutation<MedioAdmin, ApiError, { id: string; otherId: string }>({
    mutationFn: ({ id, otherId }) => mergeMedio(id, { otherId }),
    onSuccess: invalidate,
  });
}
