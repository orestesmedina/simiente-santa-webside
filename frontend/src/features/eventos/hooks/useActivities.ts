import { useQuery } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import {
  listActividades,
  type ActividadAdminList,
  type PublicationState,
} from '../../../api/agenda';
import { AGENDA_ACTIVITIES_QUERY_KEY, AGENDA_PAGE_LIMIT } from './queryKeys';

export interface UseActivitiesOptions {
  /** Filtro por estado de publicación (FR-013); sin valor: todos. */
  publicationState?: PublicationState;
}

/**
 * Listado del panel de actividades (FR-011) con su estado real y la marca de
 * pasado por elemento. No pagina: pide el tope del contrato (⟲ D-17).
 */
export function useActivities(options: UseActivitiesOptions = {}) {
  const { publicationState } = options;

  return useQuery<ActividadAdminList, ApiError>({
    queryKey: [...AGENDA_ACTIVITIES_QUERY_KEY, publicationState ?? 'all'],
    queryFn: () => listActividades({ limit: AGENDA_PAGE_LIMIT, publicationState }),
  });
}
