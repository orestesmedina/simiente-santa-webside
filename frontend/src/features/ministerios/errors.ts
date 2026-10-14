import { ApiError } from '../../api/client';
import { NO_PERMISSION_ERROR, SYSTEM_ERROR } from './messages';

/**
 * Errores del módulo «Ministerios» del panel (patrón F4): el `403` del servidor
 * (FR-011) se distingue del resto para avisar sin detalles internos, y el `409`
 * (duplicado exacto, P6-12) se muestra con su propio texto.
 */

/** `true` si el servidor respondió `403` (FR-011). */
export function isForbidden(error: unknown): boolean {
  return error instanceof ApiError && error.status === 403;
}

/** `true` si el servidor respondió `409` (duplicado exacto, P6-12). */
export function isConflict(error: unknown): boolean {
  return error instanceof ApiError && error.status === 409;
}

/**
 * Mensaje de un campo concreto del `details` del sobre de error (`400 invalid`,
 * FR-014); `undefined` si el servidor no señaló ese campo.
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
