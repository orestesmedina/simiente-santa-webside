import { ApiError } from '../../api/client';
import { NO_PERMISSION_ERROR, SEARCH_TIMEOUT_ERROR, SYSTEM_ERROR } from './messages';

/**
 * Errores del módulo «Medios» del panel (patrón F4/F6/F7): el `403` del servidor
 * (FR-021) se distingue del resto para avisar sin detalles internos, y los
 * `details` de un `400 invalid` se traducen **por campo** para pintarlos junto a
 * su control (FR-023).
 */

/** `true` si el servidor respondió `403` (FR-021). */
export function isForbidden(error: unknown): boolean {
  return error instanceof ApiError && error.status === 403;
}

/**
 * Mensaje de un campo concreto del `details` del sobre de error (`400 invalid`,
 * FR-023); `undefined` si el servidor no señaló ese campo.
 */
export function detailMessage(error: unknown, field: string): string | undefined {
  if (error instanceof ApiError && error.details) {
    const value = error.details[field];
    return typeof value === 'string' && value.trim() !== '' ? value : undefined;
  }
  return undefined;
}

/** Mensaje de aviso para los errores que no dependen de un campo. */
export function generalErrorMessage(error: unknown): string {
  return isForbidden(error) ? NO_PERMISSION_ERROR : SYSTEM_ERROR;
}

/**
 * Mensaje del fallo de «Buscar episodios nuevos» (H1): el tiempo agotado —el
 * único desenlace que el aviso genérico no explicaba— se cuenta aparte, para que
 * el equipo sepa que la búsqueda se interrumpió a medias y que puede repetirla.
 */
export function searchErrorMessage(error: unknown): string {
  if (error instanceof ApiError && error.reason === 'timeout') {
    return SEARCH_TIMEOUT_ERROR;
  }
  return SYSTEM_ERROR;
}
