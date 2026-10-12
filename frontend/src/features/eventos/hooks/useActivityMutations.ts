import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  createActividad,
  deleteActividad,
  updateActividad,
  type ActividadAdmin,
  type ActividadInput,
  type ActividadPatch,
} from '../../../api/agenda';
import { ApiError } from '../../../api/client';
import { AGENDA_ACTIVITIES_QUERY_KEY, UPCOMING_QUERY_KEY } from './queryKeys';

/** Invalida el listado del panel y la sección pública compartida (SC-003). */
function useInvalidateActivities() {
  const queryClient = useQueryClient();
  return () => {
    void queryClient.invalidateQueries({ queryKey: AGENDA_ACTIVITIES_QUERY_KEY });
    void queryClient.invalidateQueries({ queryKey: UPCOMING_QUERY_KEY });
  };
}

/** Alta de actividad puntual (FR-011/FR-015); nace en borrador. */
export function useCreateActivity() {
  const invalidate = useInvalidateActivities();

  return useMutation<ActividadAdmin, ApiError, ActividadInput>({
    mutationFn: (input) => createActividad(input),
    onSuccess: invalidate,
  });
}

/** Edición de actividad (FR-014: «se publica al guardar»). */
export function useUpdateActivity() {
  const invalidate = useInvalidateActivities();

  return useMutation<ActividadAdmin, ApiError, { id: string; input: ActividadPatch }>({
    mutationFn: ({ id, input }) => updateActividad(id, input),
    onSuccess: invalidate,
  });
}

/** Elimina una actividad (FR-011); queda registrado en la auditoría. */
export function useDeleteActivity() {
  const invalidate = useInvalidateActivities();

  return useMutation<void, ApiError, string>({
    mutationFn: (id) => deleteActividad(id),
    onSuccess: invalidate,
  });
}
