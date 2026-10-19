import { describe, expect, it } from 'vitest';
import { ApiError } from '../../api/client';
import { generalErrorMessage, isForbidden, searchErrorMessage } from './errors';
import { NO_PERMISSION_ERROR, SEARCH_TIMEOUT_ERROR, SYSTEM_ERROR } from './messages';

/**
 * Errores del módulo «Medios»: el `403` se avisa sin detalles internos (FR-021)
 * y el fallo de «Buscar episodios nuevos» por **tiempo agotado** se cuenta
 * aparte (H1 de la revisión): es el desenlace que dejaba al equipo con un
 * «error del sistema» genérico y sin saber que la búsqueda se interrumpió.
 */
describe('errores del panel de medios', () => {
  it('el `403` del servidor se distingue del resto (FR-021)', () => {
    const forbidden = ApiError.fromEnvelope(403, { code: 'forbidden', message: 'No autorizado' });

    expect(isForbidden(forbidden)).toBe(true);
    expect(isForbidden(ApiError.network())).toBe(false);
    expect(generalErrorMessage(forbidden)).toBe(NO_PERMISSION_ERROR);
    expect(generalErrorMessage(ApiError.network())).toBe(SYSTEM_ERROR);
    expect(generalErrorMessage(undefined)).toBe(SYSTEM_ERROR);
  });

  it('la búsqueda agotada por tiempo lo dice con claridad y no promete un resumen (H1)', () => {
    expect(searchErrorMessage(ApiError.timeout())).toBe(SEARCH_TIMEOUT_ERROR);
    expect(SEARCH_TIMEOUT_ERROR).toContain('se interrumpió antes de dar el resumen');
    expect(SEARCH_TIMEOUT_ERROR).toContain('Nada se borró');
  });

  it('los demás fallos de la búsqueda usan el aviso genérico', () => {
    expect(searchErrorMessage(ApiError.network())).toBe(SYSTEM_ERROR);
    expect(
      searchErrorMessage(ApiError.fromEnvelope(500, { code: 'internal', message: 'Falla' })),
    ).toBe(SYSTEM_ERROR);
    expect(searchErrorMessage(undefined)).toBe(SYSTEM_ERROR);
  });
});
