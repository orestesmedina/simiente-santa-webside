import { useQuery } from '@tanstack/react-query';
import type { SystemStatus } from '../../../api/client';
import { getSystemStatus } from '../../../api/status';
import type { EstadoSistema } from '../estado';
import { toEstadoSistema } from '../toEstadoSistema';

interface UseSystemStatusResult {
  estado: EstadoSistema | undefined;
  /** Fecha de la última consulta con resultado (éxito o error). */
  fechaConsulta: Date | undefined;
  /** Hay una consulta en curso (la primera o una reconsulta). */
  consultando: boolean;
  /** Pide una nueva consulta; sustituye el resultado anterior. */
  refetch: () => void;
}

/**
 * Consulta el estado al montar (sin auto-refresco, D14) y expone un `refetch`
 * para el botón manual. Traduce siempre el resultado a `EstadoSistema`.
 */
export function useSystemStatus(): UseSystemStatusResult {
  const query = useQuery<SystemStatus, unknown>({
    queryKey: ['system-status'],
    queryFn: getSystemStatus,
    staleTime: 0,
    refetchOnWindowFocus: false,
    retry: false,
  });

  const estado = query.isSuccess
    ? toEstadoSistema({ ok: true, value: query.data })
    : query.isError
      ? toEstadoSistema({ ok: false, error: query.error })
      : undefined;

  const fechaConsulta =
    estado !== undefined && query.dataUpdatedAt > 0 ? new Date(query.dataUpdatedAt) : undefined;

  return {
    estado,
    fechaConsulta,
    consultando: query.isFetching,
    refetch: () => {
      void query.refetch();
    },
  };
}
