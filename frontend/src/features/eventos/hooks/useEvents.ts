import { useQuery } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { listEventos, type EventoAdminList, type PublicationState } from '../../../api/agenda';
import { AGENDA_EVENTS_QUERY_KEY, AGENDA_PAGE_LIMIT } from './queryKeys';

export interface UseEventsOptions {
  /** Filtro por estado de publicación (FR-013); sin valor: todos. */
  publicationState?: PublicationState;
}

/**
 * Listado del panel de eventos (FR-010) con su estado de publicación real
 * (borrador/publicado) y la marca de pasado por elemento. No pagina: pide el
 * tope del contrato (⟲ D-17).
 */
export function useEvents(options: UseEventsOptions = {}) {
  const { publicationState } = options;

  return useQuery<EventoAdminList, ApiError>({
    queryKey: [...AGENDA_EVENTS_QUERY_KEY, publicationState ?? 'all'],
    queryFn: () => listEventos({ limit: AGENDA_PAGE_LIMIT, publicationState }),
  });
}
