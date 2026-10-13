import { describe, expect, it } from 'vitest';
import { en } from './en';
import { es } from './es';
import { publicMessageKeys } from './messages';

/**
 * Claves i18n del sitio público de F6 (ux.md §7.2, SC-006): simetría es/en sin
 * huecos y textos literales del catálogo. Si falta una clave en un idioma, el
 * diccionario no compila (`Record<PublicMessageKey, string>`); esta prueba fija
 * además el texto para que no derive.
 */

const CLAVES_MINISTERIOS = [
  'nav.sections.ministries',
  'ministries.page.title',
  'ministries.page.subtitle',
  'ministries.leader.label',
  'ministries.join',
  'a11y.ministries.join',
  'ministries.loading',
  'ministries.error.load',
  'ministries.error.retry',
  'ministries.error.partial',
  'a11y.ministries.heading',
  'home.ministries.title',
  'home.ministries.subtitle',
  'home.ministries.seeAll',
  'home.ministries.badge',
  'a11y.home.ministries.heading',
] as const;

describe('i18n de ministerios (ux.md §7.2)', () => {
  it('las claves nuevas están en el catálogo y en los dos diccionarios, sin huecos', () => {
    for (const key of CLAVES_MINISTERIOS) {
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
    expect(es['nav.sections.ministries']).toBe('Ministerios');
    expect(en['nav.sections.ministries']).toBe('Ministries');
    expect(es['ministries.page.title']).toBe('Ministerios');
    expect(en['ministries.page.title']).toBe('Ministries');
    expect(es['ministries.page.subtitle']).toBe(
      'Las áreas de servicio de nuestra iglesia: conoce cada una y cómo participar.',
    );
    expect(en['ministries.page.subtitle']).toBe(
      'The service areas of our church: get to know each one and how to take part.',
    );
    expect(es['ministries.leader.label']).toBe('Encargado');
    expect(en['ministries.leader.label']).toBe('Led by');
    expect(es['ministries.join']).toBe('Unirse al grupo de WhatsApp');
    expect(en['ministries.join']).toBe('Join the WhatsApp group');
    expect(es['ministries.loading']).toBe('Cargando…');
    expect(en['ministries.loading']).toBe('Loading…');
    expect(es['home.ministries.title']).toBe('Ministerios');
    expect(en['home.ministries.title']).toBe('Ministries');
    expect(es['home.ministries.subtitle']).toBe('Las áreas donde puedes servir');
    expect(en['home.ministries.subtitle']).toBe('The areas where you can serve');
    expect(es['home.ministries.seeAll']).toBe('Ver más ministerios');
    expect(en['home.ministries.seeAll']).toBe('See more ministries');
    expect(es['home.ministries.badge']).toBe('Ministerio');
    expect(en['home.ministries.badge']).toBe('Ministry');
  });

  it('el nombre accesible del botón de unirse interpola el nombre del ministerio', () => {
    // La plantilla lleva el marcador `{nombre}` (es) / `{name}` (en).
    expect(es['a11y.ministries.join']).toContain('{nombre}');
    expect(en['a11y.ministries.join']).toContain('{name}');
  });
});
