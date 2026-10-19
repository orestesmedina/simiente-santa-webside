import { describe, expect, it } from 'vitest';
import { en } from './en';
import { es } from './es';
import { publicMessageKeys } from './messages';

/**
 * Claves i18n del sitio público de F9 (ux.md §7.2, SC-006/SC-011): simetría
 * es/en sin huecos y textos literales del catálogo. Si falta una clave en un
 * idioma, el diccionario no compila (`Record<PublicMessageKey, string>`); esta
 * prueba fija además el texto para que no derive. El **contenido** (nombre,
 * descripción con su inglés, fecha y enlaces) no vive aquí: llega resuelto del
 * servidor y es idéntico entre idiomas (FR-025/FR-026).
 */

const CLAVES_MEDIOS = [
  'nav.sections.media',
  'media.page.title',
  'media.page.subtitle',
  'media.loadMore',
  'media.loading',
  'media.error.load',
  'media.error.retry',
  'media.error.partial',
  'media.link.youtube',
  'media.link.spotify',
  'media.link.newTab',
  'a11y.media.link.youtube',
  'a11y.media.link.spotify',
  'a11y.media.image',
  'a11y.media.card.heading',
  'a11y.media.card.newTab',
  'home.media.title',
  'home.media.subtitle',
  'home.media.seeAll',
  'a11y.home.media.heading',
] as const;

describe('i18n de medios (ux.md §7.2)', () => {
  it('las claves nuevas están en el catálogo y en los dos diccionarios, sin huecos', () => {
    for (const key of CLAVES_MEDIOS) {
      expect(publicMessageKeys).toContain(key);
      expect(es[key], `es.${key}`).toBeTruthy();
      expect(en[key], `en.${key}`).toBeTruthy();
      expect(es[key].trim().length, `es.${key} no vacía`).toBeGreaterThan(0);
      expect(en[key].trim().length, `en.${key} no vacía`).toBeGreaterThan(0);
    }
  });

  it('los dos diccionarios siguen teniendo exactamente las claves del catálogo', () => {
    expect(Object.keys(es).sort()).toEqual([...publicMessageKeys].sort());
    expect(Object.keys(en).sort()).toEqual([...publicMessageKeys].sort());
  });

  it('los textos son los del catálogo de ux.md §7.2', () => {
    expect(es['nav.sections.media']).toBe('Medios');
    expect(en['nav.sections.media']).toBe('Media');
    expect(es['media.page.title']).toBe('Medios');
    expect(en['media.page.title']).toBe('Media');
    expect(es['media.page.subtitle']).toBe(
      'Las prédicas y los podcasts de nuestra iglesia; se abren en YouTube y Spotify, en una pestaña nueva.',
    );
    expect(en['media.page.subtitle']).toBe(
      "Watch and listen to our church's sermons and podcast; they open in YouTube and Spotify in a new tab.",
    );
    expect(es['media.loadMore']).toBe('Cargar más');
    expect(en['media.loadMore']).toBe('Load more');
    expect(es['media.loading']).toBe('Cargando…');
    expect(en['media.loading']).toBe('Loading…');
    expect(es['media.error.load']).toBe(
      'No pudimos cargar los medios. Revisa tu conexión y vuelve a intentarlo.',
    );
    expect(en['media.error.load']).toBe(
      "We couldn't load the media. Check your connection and try again.",
    );
    expect(es['media.error.retry']).toBe('Volver a intentar');
    expect(en['media.error.retry']).toBe('Try again');
    expect(es['media.error.partial']).toBe(
      'No pudimos actualizar; ves la última versión disponible.',
    );
    expect(en['media.error.partial']).toBe(
      "We couldn't refresh; you're seeing the latest available version.",
    );
    expect(es['media.link.youtube']).toBe('Ver en YouTube');
    expect(en['media.link.youtube']).toBe('Watch on YouTube');
    expect(es['media.link.spotify']).toBe('Escuchar en Spotify');
    expect(en['media.link.spotify']).toBe('Listen on Spotify');
    expect(es['media.link.newTab']).toBe('pestaña nueva');
    expect(en['media.link.newTab']).toBe('new tab');
    expect(es['home.media.title']).toBe('Medios');
    expect(en['home.media.title']).toBe('Media');
    expect(es['home.media.subtitle']).toBe('Las últimas prédicas y podcasts');
    expect(en['home.media.subtitle']).toBe('The latest sermons and episodes');
    expect(es['home.media.seeAll']).toBe('Ver todo');
    expect(en['home.media.seeAll']).toBe('See all');
    expect(es['a11y.media.image']).toBe('Carátula del episodio {nombre}');
    expect(en['a11y.media.image']).toBe('Cover of the episode {name}');
    expect(es['a11y.media.card.heading']).toBe('Lista de episodios de medios');
    expect(en['a11y.media.card.heading']).toBe('List of media episodes');
    expect(es['a11y.media.card.newTab']).toBe('(pestaña nueva)');
    expect(en['a11y.media.card.newTab']).toBe('(new tab)');
  });

  it('el anuncio de los enlaces dice el servicio, la pestaña nueva y la salida del sitio (FR-005/SC-012)', () => {
    for (const key of ['a11y.media.link.youtube', 'a11y.media.link.spotify'] as const) {
      const servicioEs = key.endsWith('youtube') ? 'YouTube' : 'Spotify';
      const servicioEn = servicioEs;
      expect(es[key], `es.${key} nombra el servicio`).toContain(servicioEs);
      expect(en[key], `en.${key} nombra el servicio`).toContain(servicioEn);
      expect(es[key], `es.${key} anuncia la pestaña nueva`).toContain('pestaña nueva');
      expect(en[key], `en.${key} anuncia la pestaña nueva`).toContain('new tab');
      expect(es[key], `es.${key} anuncia que se sale del sitio`).toContain('sales del sitio');
      expect(en[key], `en.${key} anuncia que se sale del sitio`).toContain("leaves the church's");
      // El nombre del episodio se interpola en los dos idiomas.
      expect(es[key]).toContain('{nombre}');
      expect(en[key]).toContain('{name}');
    }
  });

  it('los encabezados accesibles de la sección y del bloque están traducidos', () => {
    expect(es['a11y.home.media.heading']).toBe(
      'Medios: mira los episodios recientes y entra a la sección completa',
    );
    expect(en['a11y.home.media.heading']).toBe(
      'Media: quick look at recent episodes and link to the full section',
    );
  });
});
