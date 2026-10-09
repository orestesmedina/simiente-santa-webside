import { ApiError } from '../../api/client';
import { NO_PERMISSION_ERROR, SYSTEM_ERROR } from './messages';

/** `true` si el servidor respondió `403` (FR-012/SC-011). */
export function isForbidden(error: unknown): boolean {
  return error instanceof ApiError && error.status === 403;
}

/** `true` si el servidor respondió `404` (elemento inexistente). */
export function isNotFound(error: unknown): boolean {
  return error instanceof ApiError && error.status === 404;
}

/**
 * Mensaje de un campo concreto del `details` del sobre de error (`400 invalid`,
 * FR-015); `undefined` si el servidor no señaló ese campo.
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
