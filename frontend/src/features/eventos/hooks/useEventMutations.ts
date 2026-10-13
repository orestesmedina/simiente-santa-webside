import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  createEvento,
  deleteEvento,
  updateEvento,
  type EventoAdmin,
  type EventoInput,
  type EventoPatch,
} from '../../../api/agenda';
import { ApiError } from '../../../api/client';
import { AGENDA_EVENTS_QUERY_KEY, UPCOMING_QUERY_KEY } from './queryKeys';

/** Invalida el listado del panel y la sección pública compartida (SC-003). */
function useInvalidateEvents() {
  const queryClient = useQueryClient();
  return () => {
    void queryClient.invalidateQueries({ queryKey: AGENDA_EVENTS_QUERY_KEY });
    void queryClient.invalidateQueries({ queryKey: UPCOMING_QUERY_KEY });
  };
}

/** Alta de evento (FR-010/FR-015); nace en borrador salvo estado explícito. */
export function useCreateEvent() {
  const invalidate = useInvalidateEvents();

  return useMutation<EventoAdmin, ApiError, EventoInput>({
    mutationFn: (input) => createEvento(input),
    onSuccess: invalidate,
  });
}

/**
 * Edición de evento (FR-014: «se publica al guardar»). Un `PATCH` con `dates`
 * reemplaza la lista completa (R4-11).
 */
export function useUpdateEvent() {
  const invalidate = useInvalidateEvents();

  return useMutation<EventoAdmin, ApiError, { id: string; input: EventoPatch }>({
    mutationFn: ({ id, input }) => updateEvento(id, input),
    onSuccess: invalidate,
  });
}

/** Elimina un evento y sus fechas (FR-010); queda registrado en la auditoría. */
export function useDeleteEvent() {
  const invalidate = useInvalidateEvents();

  return useMutation<void, ApiError, string>({
    mutationFn: (id) => deleteEvento(id),
    onSuccess: invalidate,
  });
}
