import { describe, expect, it } from 'vitest';
import { formatTimeRange, scheduleDayKey } from './schedule';

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

describe('formatTimeRange (analyze C2)', () => {
  it('con hora de fin muestra el rango', () => {
    expect(formatTimeRange('10:00', '12:00')).toBe('10:00 – 12:00');
  });

  it('sin hora de fin (null o ausente) muestra solo la hora de inicio', () => {
    expect(formatTimeRange('18:00')).toBe('18:00');
    expect(formatTimeRange('18:00', null)).toBe('18:00');
  });
});
