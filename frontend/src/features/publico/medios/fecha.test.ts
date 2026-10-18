import { describe, expect, it } from 'vitest';
import { CHURCH_TIME_ZONE, episodeDay, formatEpisodeDate } from './fecha';

/**
 * Presentación de la fecha de publicación de un episodio (FR-006/FR-026): el
 * día se calcula en la zona de la iglesia (`America/Costa_Rica`), no en UTC ni
 * en la del navegador, y la palabra del formato sigue el idioma de la interfaz
 * mientras el **dato** es único entre idiomas.
 */
describe('fecha del episodio (FR-006)', () => {
  it('usa la zona de la iglesia, no UTC (el caso real de los feeds)', () => {
    // «Fri, 03 Jul 2026 01:58:52 GMT» es el **2 de julio** a las 19:58 en
    // America/Costa_Rica (UTC−6): el mismo caso que fija el «mismo día» de la
    // unión en el backend (T907).
    expect(CHURCH_TIME_ZONE).toBe('America/Costa_Rica');
    expect(episodeDay('2026-07-03T01:58:52Z')).toBe('2026-07-02');
    expect(formatEpisodeDate('2026-07-03T01:58:52Z', 'es')).toBe('jueves 2 de julio');
    expect(formatEpisodeDate('2026-07-03T01:58:52Z', 'en')).toBe('Thursday, July 2');
  });

  it('el dato es único entre idiomas: cambia la palabra, nunca el día', () => {
    const iso = '2026-05-25T18:43:42Z';
    expect(episodeDay(iso)).toBe('2026-05-25');
    expect(formatEpisodeDate(iso, 'es')).toContain('25');
    expect(formatEpisodeDate(iso, 'en')).toContain('25');
    expect(formatEpisodeDate(iso, 'es')).toBe('lunes 25 de mayo');
    expect(formatEpisodeDate(iso, 'en')).toBe('Monday, May 25');
  });

  it('un valor que no es una fecha del contrato se devuelve tal cual (la vista no se rompe)', () => {
    expect(episodeDay('no-es-fecha')).toBe('no-es-fecha');
    expect(formatEpisodeDate('no-es-fecha', 'es')).toBe('no-es-fecha');
    expect(formatEpisodeDate('', 'es')).toBe('');
  });
});
