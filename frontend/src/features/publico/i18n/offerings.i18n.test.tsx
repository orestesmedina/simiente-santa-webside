import { describe, expect, it } from 'vitest';
import { en } from './en';
import { es } from './es';
import { publicMessageKeys } from './messages';

/**
 * Claves i18n del sitio público de F7 (ux.md §7.2, SC-006): simetría es/en sin
 * huecos y textos literales del catálogo. Si falta una clave en un idioma, el
 * diccionario no compila (`Record<PublicMessageKey, string>`); esta prueba fija
 * además el texto para que no derive.
 */

const CLAVES_OFRENDAS = [
  'nav.sections.offerings',
  'offerings.page.title',
  'offerings.page.subtitle',
  'offerings.iban.heading',
  'offerings.sinpe.heading',
  'offerings.holder.label',
  'offerings.currency.label',
  'offerings.currency.colones',
  'offerings.currency.dolares',
  'offerings.copy',
  'offerings.copied',
  'offerings.copy.hint',
  'offerings.usage.heading',
  'offerings.loading',
  'offerings.error.load',
  'offerings.error.retry',
  'offerings.error.partial',
  'a11y.offerings.account',
  'a11y.offerings.sinpe',
  'home.offerings.title',
  'home.offerings.subtitle',
  'home.offerings.seeAll',
  'home.offerings.badge.iban',
  'home.offerings.badge.sinpe',
  'a11y.home.offerings.heading',
] as const;

describe('i18n de ofrendas (ux.md §7.2)', () => {
  it('las claves nuevas están en el catálogo y en los dos diccionarios, sin huecos', () => {
    for (const key of CLAVES_OFRENDAS) {
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
    expect(es['nav.sections.offerings']).toBe('Ofrendas y diezmos');
    expect(en['nav.sections.offerings']).toBe('Offerings & tithes');
    expect(es['offerings.page.title']).toBe('Ofrendas y diezmos');
    expect(en['offerings.page.title']).toBe('Offerings & tithes');
    expect(es['offerings.page.subtitle']).toBe(
      'Las formas de apoyar económicamente a nuestra iglesia.',
    );
    expect(en['offerings.page.subtitle']).toBe('The ways to support our church financially.');
    expect(es['offerings.iban.heading']).toBe('Cuentas IBAN');
    expect(en['offerings.iban.heading']).toBe('Bank accounts (IBAN)');
    expect(es['offerings.sinpe.heading']).toBe('Números de SINPE Móvil');
    expect(en['offerings.sinpe.heading']).toBe('SINPE Móvil numbers');
    expect(es['offerings.holder.label']).toBe('Titular');
    expect(en['offerings.holder.label']).toBe('Account holder');
    expect(es['offerings.currency.label']).toBe('Moneda');
    expect(en['offerings.currency.label']).toBe('Currency');
    expect(es['offerings.currency.colones']).toBe('colones');
    expect(en['offerings.currency.colones']).toBe('colones (CRC)');
    expect(es['offerings.currency.dolares']).toBe('dólares');
    expect(en['offerings.currency.dolares']).toBe('dollars (USD)');
    expect(es['offerings.copy']).toBe('Copiar número');
    expect(en['offerings.copy']).toBe('Copy number');
    expect(es['offerings.copied']).toBe('Número copiado.');
    expect(en['offerings.copied']).toBe('Number copied.');
    expect(es['offerings.copy.hint']).toBe(
      'Puedes copiarlo en un toque o escribirlo tal cual, bloque por bloque.',
    );
    expect(en['offerings.copy.hint']).toBe(
      'You can copy it with one tap or type it out block by block.',
    );
    expect(es['offerings.usage.heading']).toBe('¿En qué usamos las donaciones?');
    expect(en['offerings.usage.heading']).toBe('How we use the offerings');
    expect(es['offerings.loading']).toBe('Cargando…');
    expect(en['offerings.loading']).toBe('Loading…');
    expect(es['offerings.error.load']).toBe(
      'No pudimos cargar la información de ofrendas. Revisa tu conexión y vuelve a intentarlo.',
    );
    expect(en['offerings.error.load']).toBe(
      "We couldn't load the offering info. Check your connection and try again.",
    );
    expect(es['offerings.error.retry']).toBe('Volver a intentar');
    expect(en['offerings.error.retry']).toBe('Try again');
    expect(es['offerings.error.partial']).toBe(
      'No pudimos actualizar; ves la última versión disponible.',
    );
    expect(en['offerings.error.partial']).toBe(
      "We couldn't refresh; you're seeing the latest available version.",
    );
    expect(es['home.offerings.title']).toBe('Ofrendas y diezmos');
    expect(en['home.offerings.title']).toBe('Offerings & tithes');
    expect(es['home.offerings.subtitle']).toBe('Las formas de apoyar a la iglesia');
    expect(en['home.offerings.subtitle']).toBe('The ways to support the church');
    expect(es['home.offerings.seeAll']).toBe('Ver formas de ofrendar');
    expect(en['home.offerings.seeAll']).toBe('See ways to give');
    expect(es['home.offerings.badge.iban']).toBe('Cuenta IBAN');
    expect(en['home.offerings.badge.iban']).toBe('Bank account (IBAN)');
    expect(es['home.offerings.badge.sinpe']).toBe('SINPE Móvil');
    expect(en['home.offerings.badge.sinpe']).toBe('SINPE Móvil number');
  });

  it('las plantillas accesibles llevan sus marcadores de interpolación', () => {
    expect(es['a11y.offerings.account']).toContain('{entidad}');
    expect(es['a11y.offerings.account']).toContain('{número}');
    expect(en['a11y.offerings.account']).toContain('{entity}');
    expect(en['a11y.offerings.account']).toContain('{number}');
    expect(es['a11y.offerings.sinpe']).toContain('{entidad}');
    expect(en['a11y.offerings.sinpe']).toContain('{entity}');
    expect(es['a11y.home.offerings.heading']).toBeTruthy();
    expect(en['a11y.home.offerings.heading']).toBeTruthy();
  });
});
