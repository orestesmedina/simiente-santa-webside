import { ApiError, type SystemStatus } from '../../api/client';
import type { EstadoSistema } from './estado';

/**
 * Traduce el resultado de `getSystemStatus()` a un `EstadoSistema` (ux.md §3.1).
 * Cero ramas sin estado: cualquier fallo cae en uno de los estados definidos.
 */
export function toEstadoSistema(
  result: { ok: true; value: SystemStatus } | { ok: false; error: unknown },
): EstadoSistema {
  if (result.ok) {
    // El éxito ya exige `status === "ok"` y `database === "connected"` en
    // `getSystemStatus`; cualquier otro 200 llega aquí como error.
    return { kind: 'conectado' };
  }

  const error = result.error;

  if (error instanceof ApiError) {
    if (error.reason === 'http' && error.code === 'database_unavailable') {
      return { kind: 'bd-no-conectada' };
    }
    if (error.reason === 'network' || error.reason === 'timeout') {
      return { kind: 'inaccesible', motivo: 'sin-respuesta' };
    }
    return { kind: 'inaccesible', motivo: 'respuesta-inesperada' };
  }

  return { kind: 'inaccesible', motivo: 'respuesta-inesperada' };
}
