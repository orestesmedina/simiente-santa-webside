import { describe, expect, it } from 'vitest';
import { ApiError } from '../../api/client';
import { detailMessage, generalErrorMessage, isConflict, isForbidden } from './errors';

function httpError(status: number, details?: Record<string, unknown>): ApiError {
  return new ApiError('http', 'error', { status, details });
}

describe('errores del módulo de donaciones', () => {
  it('reconoce el 403 y el 409', () => {
    expect(isForbidden(httpError(403))).toBe(true);
    expect(isForbidden(httpError(409))).toBe(false);
    expect(isConflict(httpError(409))).toBe(true);
    expect(isConflict(httpError(403))).toBe(false);
    expect(isForbidden(new Error('otro'))).toBe(false);
  });

  it('traduce el `details` por campo', () => {
    const error = httpError(400, { number: 'IBAN inválido', institution: '  ' });
    expect(detailMessage(error, 'number')).toBe('IBAN inválido');
    expect(detailMessage(error, 'institution')).toBeUndefined();
    expect(detailMessage(error, 'otro')).toBeUndefined();
  });

  it('el mensaje general distingue el 403 del resto', () => {
    expect(generalErrorMessage(httpError(403))).toMatch(/No tienes acceso/);
    expect(generalErrorMessage(httpError(500))).toMatch(/No se pudo completar la operación/);
    expect(generalErrorMessage(httpError(409))).toMatch(/No se pudo completar la operación/);
  });
});
