import { ApiError, apiFetch, type SystemStatus } from './client';

/**
 * Único lugar que consulta el estado del sistema (`GET /healthz`). Devuelve el
 * DTO tipado del contrato o lanza `ApiError`. La traducción a los estados de
 * la interfaz (`EstadoSistema`) la hace la feature `status` (T024).
 */
export async function getSystemStatus(): Promise<SystemStatus> {
  const body = await apiFetch<unknown>('/healthz');

  if (!isSystemStatus(body)) {
    throw ApiError.invalidResponse();
  }

  return body;
}

/**
 * El éxito exige el sobre completo: `status === "ok"` y
 * `database === "connected"`. Cualquier otro cuerpo (vacío, incompleto o con
 * un valor no reconocido) se trata como respuesta inesperada (ux.md §3.1).
 */
function isSystemStatus(value: unknown): value is SystemStatus {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const candidate = value as Record<string, unknown>;
  return candidate.status === 'ok' && candidate.database === 'connected';
}
