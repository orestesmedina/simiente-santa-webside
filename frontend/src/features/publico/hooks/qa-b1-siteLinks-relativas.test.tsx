import { describe, expect, it } from 'vitest';
import type { PortadaPublica } from '../../../api/portada';
import { buildSiteLinks } from './useSiteLinks';

/**
 * QA · Bug 1 (⟲ 004-eventos-actividades): desde `/eventos` los enlaces del menú
 * no llevan a las secciones de la portada. Causa raíz: `buildSiteLinks` produce
 * `href` **relativos** (`#whatsapp`); al renderizarse como `<a>` nativos desde
 * una ruta distinta de `/`, el navegador los resuelve contra `/eventos` y el
 * usuario queda en `/eventos#whatsapp` (no navega). Lot de este test: cada ancla
 * de portada debe incluir el origen (`/#whatsapp`), de modo que siempre apunte
 * a la portada aunque el enlace se pinte en cualquier ruta.
 */

const portadaConTodo: PortadaPublica = {
  lang: 'es',
  about: { text: 'Somos una familia' },
  schedule: [{ id: 'srv-1', dayOfWeek: 0, startTime: '10:00', name: 'Culto', place: 'Templo' }],
  whatsapp: [{ id: 'wa-1', name: 'Escríbenos', kind: 'direct', url: 'https://wa.me/50688888888' }],
  contact: { address: 'San José', email: 'hola@ejemplo.com', phone: '+506 8888 8888' },
  socials: [{ id: 'so-1', network: 'facebook', url: 'https://facebook.com/simiente' }],
};

describe('useSiteLinks · QA Bug 1: los anclas de portada son absolutos a la raíz', () => {
  it('cada enlace de sección apunta a `/#ancla`, no a un ancla relativo', () => {
    const links = buildSiteLinks(portadaConTodo);

    const whatsapp = links.find((link) => link.id === 'whatsapp');
    expect(whatsapp, 'existe el enlace de WhatsApp').toBeDefined();
    expect(whatsapp?.href).toBe('/#whatsapp');

    const quienes = links.find((link) => link.id === 'who');
    expect(quienes?.href).toBe('/#quienes-somos');
  });

  it('ningún href de portada queda como ancla relativa (se resolvería contra la ruta actual)', () => {
    const links = buildSiteLinks(portadaConTodo);

    for (const link of links.filter((candidate) => candidate.id !== 'events')) {
      expect(
        link.href.startsWith('/#'),
        `el enlace ${link.id} debe incluir la ruta de la portada; recibió «${link.href}»`,
      ).toBe(true);
    }
  });

  it('el ítem de agenda sigue siendo la ruta /eventos (sin doble barra)', () => {
    const links = buildSiteLinks(portadaConTodo, {
      lang: 'es',
      events: [
        {
          id: 'ev-1',
          title: 'Vigilia',
          dates: [{ date: '2026-10-17' }],
          place: 'Templo',
          detail: 'Noche de oración',
        },
      ],
    });

    expect(links.find((link) => link.id === 'events')?.href).toBe('/eventos');
  });
});
