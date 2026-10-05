import { describe, expect, it } from 'vitest';
import { formatDate, formatDateTime, formatLastLogin, formatTime } from './format';

// Fechas construidas en hora local para que el resultado no dependa de la TZ de CI.
const fecha = new Date(2026, 9, 3, 17, 42);

describe('format', () => {
  it('formatea la fecha en español', () => {
    expect(formatDate(fecha)).toBe('03/10/2026');
  });

  it('formatea la hora en 24 h', () => {
    expect(formatTime(fecha)).toBe('17:42');
  });

  it('formatea fecha y hora para los historiales', () => {
    expect(formatDateTime(fecha)).toBe('03/10/2026, 17:42');
  });

  it('acepta la fecha en ISO del contrato', () => {
    expect(formatDateTime(fecha.toISOString())).toBe('03/10/2026, 17:42');
  });

  it('indica "sin accesos" sin inventar fechas', () => {
    expect(formatLastLogin(null)).toBe('Nunca ha entrado al panel.');
    expect(formatLastLogin(undefined, '189.2.4.15')).toBe('Nunca ha entrado al panel.');
  });

  it('compone el último acceso con fecha, hora e IP', () => {
    expect(formatLastLogin(fecha.toISOString(), '189.2.4.15')).toBe(
      '03/10/2026, a las 17:42, desde 189.2.4.15',
    );
  });
});
