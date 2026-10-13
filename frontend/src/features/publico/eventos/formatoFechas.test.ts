import { describe, expect, it } from 'vitest';
import { en } from '../i18n/en';
import { es } from '../i18n/es';
import { publicMessageKeys } from '../i18n/messages';
import { datetimeValue, formatLongDate } from './formatoFechas';

describe('formatLongDate (FR-009/R4-3/D-22)', () => {
  it('escribe la fecha en palabras en español, sin desplazar el día', () => {
    expect(formatLongDate('2026-10-17', 'es')).toBe('sábado 17 de octubre');
    expect(formatLongDate('2026-01-01', 'es')).toBe('jueves 1 de enero');
    expect(formatLongDate('2026-12-31', 'es')).toBe('jueves 31 de diciembre');
  });

  it('escribe la fecha en palabras en inglés', () => {
    expect(formatLongDate('2026-10-17', 'en')).toBe('Saturday, October 17');
    expect(formatLongDate('2026-01-01', 'en')).toBe('Thursday, January 1');
  });

  it('el dato no cambia entre idiomas: mismo día, distinta presentación', () => {
    // El caso que rompía `new Date('YYYY-MM-DD')` en zonas UTC−x (Costa Rica):
    // el 1 de enero es 1 de enero en ambos idiomas, nunca el 31 de diciembre.
    expect(formatLongDate('2026-01-01', 'es')).toContain('1');
    expect(formatLongDate('2026-01-01', 'en')).toContain('1');
    expect(formatLongDate('2026-01-01', 'es')).not.toContain('31');
    expect(formatLongDate('2026-01-01', 'en')).not.toContain('December');
  });

  it('devuelve el valor tal cual si no es una fecha válida del contrato', () => {
    expect(formatLongDate('2026-13-32', 'es')).toBe('2026-13-32');
    expect(formatLongDate('', 'es')).toBe('');
    expect(formatLongDate('17/10/2026', 'en')).toBe('17/10/2026');
  });
});

describe('datetimeValue (FR-009)', () => {
  it('transporta el día y, si existe, la hora de inicio en 24 h', () => {
    expect(datetimeValue('2026-10-17', '19:00')).toBe('2026-10-17T19:00');
  });

  it('sin hora lleva solo el día (una actividad sin horario es válida, D-12)', () => {
    expect(datetimeValue('2026-10-18')).toBe('2026-10-18');
    expect(datetimeValue('2026-10-18', undefined)).toBe('2026-10-18');
  });
});

describe('claves i18n de la agenda (ux.md §7.2, SC-006)', () => {
  const NUEVAS = [
    'nav.sections.events',
    'events.page.title',
    'events.page.subtitle',
    'events.block.scheduled',
    'events.block.activities',
    'events.location.label',
    'events.location.link',
    'events.date.next',
    'events.date.more',
    'events.empty.hidden',
    'events.category.aria',
    'events.loading',
    'events.error.load',
    'events.error.retry',
    'events.error.partial',
    'a11y.events.heading',
  ] as const;

  it('existen en los dos diccionarios y ninguna queda vacía', () => {
    for (const key of NUEVAS) {
      expect(publicMessageKeys, `catálogo incluye ${key}`).toContain(key);
      expect(es[key], `es.${key}`).toBeTruthy();
      expect(en[key], `en.${key}`).toBeTruthy();
      expect(es[key].trim().length, `es.${key} no vacía`).toBeGreaterThan(0);
      expect(en[key].trim().length, `en.${key} no vacía`).toBeGreaterThan(0);
    }
  });

  it('usa los textos del catálogo de ux.md §7.2', () => {
    expect(es['nav.sections.events']).toBe('Eventos y actividades');
    expect(en['nav.sections.events']).toBe('Events & activities');
    expect(es['events.page.title']).toBe('Eventos y actividades');
    expect(es['events.block.scheduled']).toBe('Próximos eventos');
    expect(en['events.block.scheduled']).toBe('Upcoming events');
    expect(es['events.block.activities']).toBe('Próximas actividades');
    expect(en['events.block.activities']).toBe('Upcoming activities');
    expect(es['events.location.link']).toBe('Ver ubicación');
    expect(en['events.location.link']).toBe('See location');
    expect(es['events.date.next']).toBe('Próxima fecha');
    expect(es['events.date.more']).toBe('Más fechas');
    expect(es['events.error.retry']).toBe('Volver a intentar');
    expect(en['events.error.retry']).toBe('Try again');
    expect(es['a11y.events.heading']).toBe('Eventos y actividades de la iglesia');
    expect(en['a11y.events.heading']).toBe('Church events and activities');
  });

  it('el aviso de sección vacía es la copia del humano (tasks.md, alineación 6)', () => {
    expect(es['events.empty.hidden']).toBe('Por ahora no hay eventos ni actividades');
  });

  it('la clave de categoría espera el nombre interpolado', () => {
    expect(es['events.category.aria']).toContain('{name}');
    expect(en['events.category.aria']).toContain('{name}');
  });
});
