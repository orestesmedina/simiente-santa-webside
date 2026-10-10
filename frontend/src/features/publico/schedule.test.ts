import { describe, expect, it } from 'vitest';
import { en } from './i18n/en';
import { es } from './i18n/es';
import type { Language, PublicMessageKey } from './i18n/messages';
import { formatClockTime, formatTimeRange, scheduleDayKey } from './schedule';

/** Traductor de prueba: resuelve una clave en el idioma pedido. */
const translate = (lang: Language) => {
  const dictionary = { es, en }[lang];
  return (key: PublicMessageKey) => dictionary[key];
};

describe('scheduleDayKey', () => {
  it('mapea 0–6 a las claves i18n de los días', () => {
    expect(scheduleDayKey(0)).toBe('schedule.day.0');
    expect(scheduleDayKey(3)).toBe('schedule.day.3');
    expect(scheduleDayKey(6)).toBe('schedule.day.6');
  });

  it('un valor fuera de rango cae al domingo sin romper', () => {
    expect(scheduleDayKey(9)).toBe('schedule.day.0');
    expect(scheduleDayKey(-1)).toBe('schedule.day.0');
  });
});

describe('formatTimeRange (analyze C2, ux.md §4.6/D-3)', () => {
  it('localiza el rango a a.m./p.m. en español', () => {
    const t = translate('es');
    expect(formatTimeRange('10:00', '12:00', t)).toBe('10:00 a. m. − 12:00 m.');
    expect(formatTimeRange('18:00', null, t)).toBe('6:00 p. m.');
    expect(formatTimeRange('00:00', '13:30', t)).toBe('12:00 a. m. − 1:30 p. m.');
    expect(formatTimeRange('11:30', '12:45', t)).toBe('11:30 a. m. − 12:45 p. m.');
  });

  it('reserva «m.» para el mediodía exacto y usa «p. m.» en toda la franja de las 12', () => {
    const t = translate('es');
    expect(formatClockTime('12:00', t)).toBe('12:00 m.');
    expect(formatClockTime('12:45', t)).toBe('12:45 p. m.');
    expect(formatClockTime('13:00', t)).toBe('1:00 p. m.');
    expect(formatClockTime('00:00', t)).toBe('12:00 a. m.');
  });

  it('localiza el rango a AM/PM en inglés', () => {
    const t = translate('en');
    expect(formatTimeRange('10:00', '12:00', t)).toBe('10:00 AM − 12:00 PM');
    expect(formatTimeRange('18:00', null, t)).toBe('6:00 PM');
    expect(formatTimeRange('00:00', '13:30', t)).toBe('12:00 AM − 1:30 PM');
  });

  it('en inglés toda la franja de las 12 es PM, con medianoche AM', () => {
    const t = translate('en');
    expect(formatClockTime('12:00', t)).toBe('12:00 PM');
    expect(formatClockTime('12:45', t)).toBe('12:45 PM');
    expect(formatClockTime('13:00', t)).toBe('1:00 PM');
    expect(formatClockTime('00:00', t)).toBe('12:00 AM');
  });

  it('sin hora de fin (null o ausente) muestra solo la hora de inicio', () => {
    const t = translate('es');
    expect(formatTimeRange('18:00', undefined, t)).toBe('6:00 p. m.');
    expect(formatTimeRange('18:00', null, t)).toBe('6:00 p. m.');
  });

  it('un valor fuera del contrato «HH:MM» se muestra tal cual', () => {
    const t = translate('es');
    expect(formatClockTime('mediodía', t)).toBe('mediodía');
    expect(formatClockTime('24:00', t)).toBe('24:00');
  });
});
