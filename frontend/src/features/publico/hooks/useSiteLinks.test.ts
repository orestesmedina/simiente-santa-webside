import { describe, expect, it } from 'vitest';
import type { PortadaPublica } from '../../../api/portada';
import { PUBLIC_SECTIONS } from '../sections';
import { buildSiteLinks } from './useSiteLinks';

const completa: PortadaPublica = {
  lang: 'es',
  about: { text: 'Somos una familia' },
  schedule: [{ id: 'srv-1', dayOfWeek: 0, startTime: '10:00', name: 'Culto', place: 'Templo' }],
  whatsapp: [{ id: 'wa-1', name: 'Escríbenos', kind: 'direct', url: 'https://wa.me/50688888888' }],
  contact: { address: 'San José', email: 'hola@ejemplo.com', phone: '+506 8888 8888' },
  socials: [{ id: 'so-1', network: 'facebook', url: 'https://facebook.com/simiente' }],
};

describe('useSiteLinks · fuente única de navegación (FR-012/FR-027)', () => {
  it('sin datos no hay ningún enlace', () => {
    expect(buildSiteLinks(undefined)).toEqual([]);
  });

  it('las secciones sin contenido publicado no generan enlace (0 enlaces rotos)', () => {
    const links = buildSiteLinks({ lang: 'es', about: { text: 'solo quiénes somos' } });

    expect(links.map((link) => link.id)).toEqual(['who']);
    expect(links.map((link) => link.href)).toEqual(['#quienes-somos']);
  });

  it('el orden es el de la página y los nombres salen de las claves de la cabecera', () => {
    const links = buildSiteLinks(completa);

    expect(links.map((link) => link.id)).toEqual([
      'who',
      'schedule',
      'whatsapp',
      'contact',
      'social',
    ]);

    for (const link of links) {
      const section = PUBLIC_SECTIONS.find((candidate) => candidate.id === link.id);
      expect(section, `la sección ${link.id} existe en PUBLIC_SECTIONS`).toBeDefined();
      expect(link.messageKey).toBe(section?.messageKey);
      expect(link.anchor).toBe(section?.anchor);
      expect(link.href).toBe(`#${section?.anchor}`);
    }
  });

  it('no inventa enlaces a rutas F4–F9: todo href es un ancla existente de la portada', () => {
    for (const link of buildSiteLinks(completa)) {
      expect(link.href.startsWith('#')).toBe(true);
      expect(PUBLIC_SECTIONS.some((section) => `#${section.anchor}` === link.href)).toBe(true);
    }
  });
});
