import { render, screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { PortadaPublica } from '../../../api/portada';
import type { OfrendasPublicas } from '../../../api/ofrendas';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { PublicLayout } from '../components/PublicLayout';
import { PUBLIC_SECTIONS } from '../sections';
import { buildSiteLinks } from './useSiteLinks';

const portadaUrl = `${API_BASE_URL}/api/v1/portada`;
const upcomingUrl = `${API_BASE_URL}/api/v1/proximamente`;
const ministriesUrl = `${API_BASE_URL}/api/v1/portada/ministerios`;
const ofrendasUrl = `${API_BASE_URL}/api/v1/ofrendas`;

const completa: PortadaPublica = {
  lang: 'es',
  about: { text: 'Somos una familia' },
  schedule: [{ id: 'srv-1', dayOfWeek: 0, startTime: '10:00', name: 'Culto', place: 'Templo' }],
  whatsapp: [{ id: 'wa-1', name: 'Escríbenos', kind: 'direct', url: 'https://wa.me/50688888888' }],
  contact: { address: 'San José', email: 'hola@ejemplo.com', phone: '+506 8888 8888' },
  socials: [{ id: 'so-1', network: 'facebook', url: 'https://facebook.com/simiente' }],
};

const conEventos = {
  lang: 'es' as const,
  events: [
    {
      id: 'ev-1',
      title: 'Vigilia',
      dates: [{ date: '2026-10-17' }],
      place: 'Templo',
      detail: 'Noche de oración',
    },
  ],
};

describe('useSiteLinks · fuente única de navegación (FR-012/FR-027)', () => {
  it('sin datos no hay ningún enlace', () => {
    expect(buildSiteLinks(undefined)).toEqual([]);
  });

  it('las secciones sin contenido publicado no generan enlace (0 enlaces rotos)', () => {
    const links = buildSiteLinks({ lang: 'es', about: { text: 'solo quiénes somos' } });

    expect(links.map((link) => link.id)).toEqual(['who']);
    expect(links.map((link) => link.href)).toEqual(['/#quienes-somos']);
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
      expect(link.href).toBe(`/#${section?.anchor}`);
    }
  });

  it('las anclas de la portada no cambian al añadir el ítem de ruta (F4)', () => {
    const conRuta = buildSiteLinks(completa, conEventos);
    const sinRuta = buildSiteLinks(completa);

    expect(conRuta.slice(0, sinRuta.length)).toEqual(sinRuta);
    for (const link of sinRuta) {
      // QA Bug 1: el ancla es absoluta a la portada (`/#ancla`), nunca relativa.
      expect(link.href.startsWith('/#')).toBe(true);
      expect(PUBLIC_SECTIONS.some((section) => `/#${section.anchor}` === link.href)).toBe(true);
    }
  });
});

describe('useSiteLinks · sección de eventos (FR-021/FR-022, ⟲ D-1/D-23)', () => {
  it('añade «Eventos y actividades» al final con href de ruta cuando hay contenido', () => {
    const links = buildSiteLinks(completa, conEventos);

    const ultimo = links[links.length - 1];
    expect(ultimo?.id).toBe('events');
    expect(ultimo?.href).toBe('/eventos');
    expect(ultimo?.messageKey).toBe('nav.sections.events');
  });

  it('sin nada publicado y próximo el ítem no existe (ni enlace roto ni huérfano)', () => {
    expect(buildSiteLinks(completa, { lang: 'es' }).map((link) => link.id)).not.toContain('events');
    expect(buildSiteLinks(completa, undefined).map((link) => link.id)).not.toContain('events');
    expect(buildSiteLinks(undefined, conEventos).map((link) => link.id)).toEqual(['events']);
  });
});

describe('useSiteLinks · sección de ministerios (FR-005/FR-006, T628)', () => {
  const conMinisterios = {
    lang: 'es' as const,
    items: [{ id: 'min-1', name: 'Alabanza', schedule: 'Todos los viernes' }],
  };

  it('añade «Ministerios» al final con href de ruta cuando el bloque trae ítems', () => {
    const links = buildSiteLinks(completa, conEventos, conMinisterios);

    const ultimo = links[links.length - 1];
    expect(ultimo?.id).toBe('ministries');
    expect(ultimo?.href).toBe('/ministerios');
    expect(ultimo?.messageKey).toBe('nav.sections.ministries');
    // Va **después** de «Eventos y actividades» (⟲ D-1: al final de la lista).
    expect(links.map((link) => link.id).indexOf('ministries')).toBeGreaterThan(
      links.map((link) => link.id).indexOf('events'),
    );
  });

  it('la presencia se deduce del documento del BLOQUE, no de la sección', () => {
    // `items` omitida (0 publicados) → sin enlace, en ninguna de las tres superficies.
    expect(
      buildSiteLinks(completa, conEventos, { lang: 'es' }).map((link) => link.id),
    ).not.toContain('ministries');
    expect(buildSiteLinks(completa, conEventos, undefined).map((link) => link.id)).not.toContain(
      'ministries',
    );
    // `items: []` tampoco cuenta como contenido.
    expect(
      buildSiteLinks(completa, undefined, { lang: 'es', items: [] }).map((link) => link.id),
    ).not.toContain('ministries');
    // Sin la respuesta del bloque, solo queda el ítem nuevo si hay ítems.
    expect(buildSiteLinks(undefined, undefined, conMinisterios).map((link) => link.id)).toEqual([
      'ministries',
    ]);
  });

  it('el resto de enlaces (anclas de portada y /eventos) no cambia', () => {
    const sinMinisterios = buildSiteLinks(completa, conEventos);
    const conMinisteriosLinks = buildSiteLinks(completa, conEventos, conMinisterios);

    expect(conMinisteriosLinks.slice(0, sinMinisterios.length)).toEqual(sinMinisterios);
  });
});

describe('useSiteLinks · superficies de la misma lista (FR-022)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  function renderPublico() {
    return render(
      <AppProviders>
        <MemoryRouter initialEntries={['/']}>
          <Routes>
            <Route element={<PublicLayout />}>
              <Route index element={<p>portada</p>} />
            </Route>
          </Routes>
        </MemoryRouter>
      </AppProviders>,
    );
  }

  it('con contenido el ítem aparece en la barra y en el mapa del pie', async () => {
    server.use(
      http.get(portadaUrl, () => HttpResponse.json(completa)),
      http.get(upcomingUrl, () => HttpResponse.json(conEventos)),
    );

    renderPublico();

    const nav = await screen.findByRole('navigation', { name: 'Secciones de la portada' });
    const enBarra = await within(nav).findByRole('link', { name: 'Eventos y actividades' });
    expect(enBarra).toHaveAttribute('href', '/eventos');

    const sitemap = screen.getByRole('navigation', { name: 'Mapa del sitio' });
    expect(within(sitemap).getByRole('link', { name: 'Eventos y actividades' })).toHaveAttribute(
      'href',
      '/eventos',
    );

    // La lista del menú móvil es la MISMA del `<nav>` (una sola lista en DOM).
    expect(document.querySelectorAll('a[href="/eventos"]')).toHaveLength(2);

    // El resto de enlaces de la portada sigue siendo un ancla existente (y
    // absoluta a la portada, QA Bug 1).
    expect(within(nav).getByRole('link', { name: 'Quiénes somos' })).toHaveAttribute(
      'href',
      '/#quienes-somos',
    );
  });

  it('sin contenido el ítem no está en ninguna superficie', async () => {
    let peticiones = 0;
    server.use(
      http.get(portadaUrl, () => HttpResponse.json(completa)),
      http.get(upcomingUrl, () => {
        peticiones += 1;
        return HttpResponse.json({ lang: 'es' });
      }),
    );

    renderPublico();

    const nav = await screen.findByRole('navigation', { name: 'Secciones de la portada' });
    await within(nav).findByRole('link', { name: 'Quiénes somos' });
    // Espera a que la consulta pública de agenda haya respondido (sin bloques).
    await waitFor(() => expect(peticiones).toBeGreaterThan(0));

    expect(screen.queryByRole('link', { name: 'Eventos y actividades' })).toBeNull();
    expect(document.querySelectorAll('a[href="/eventos"]')).toHaveLength(0);
  });

  it('con ministerios publicados el ítem aparece en la barra y en el mapa del pie (T628)', async () => {
    server.use(
      http.get(portadaUrl, () => HttpResponse.json(completa)),
      http.get(ministriesUrl, () =>
        HttpResponse.json({ lang: 'es', items: [{ id: 'min-1', name: 'Alabanza' }] }),
      ),
    );

    renderPublico();

    const nav = await screen.findByRole('navigation', { name: 'Secciones de la portada' });
    expect(await within(nav).findByRole('link', { name: 'Ministerios' })).toHaveAttribute(
      'href',
      '/ministerios',
    );

    const sitemap = screen.getByRole('navigation', { name: 'Mapa del sitio' });
    expect(within(sitemap).getByRole('link', { name: 'Ministerios' })).toHaveAttribute(
      'href',
      '/ministerios',
    );

    // Barra + mapa del pie: la lista del menú móvil es la MISMA del `<nav>`.
    expect(document.querySelectorAll('a[href="/ministerios"]')).toHaveLength(2);
  });

  it('sin ministerios publicados el ítem no está en ninguna superficie (FR-005/SC-010)', async () => {
    let peticiones = 0;
    server.use(
      http.get(portadaUrl, () => HttpResponse.json(completa)),
      http.get(ministriesUrl, () => {
        peticiones += 1;
        return HttpResponse.json({ lang: 'es' });
      }),
    );

    renderPublico();

    const nav = await screen.findByRole('navigation', { name: 'Secciones de la portada' });
    await within(nav).findByRole('link', { name: 'Quiénes somos' });
    await waitFor(() => expect(peticiones).toBeGreaterThan(0));

    expect(screen.queryByRole('link', { name: 'Ministerios' })).toBeNull();
    expect(document.querySelectorAll('a[href="/ministerios"]')).toHaveLength(0);
  });
});

describe('useSiteLinks · sección de ofrendas (P7-8/P7-13, T727)', () => {
  const conMetodos: OfrendasPublicas = {
    lang: 'es',
    items: [
      {
        id: 'don-1',
        kind: 'iban',
        number: 'CR05015202001234567890',
        institution: 'Banco Nacional de Costa Rica',
        currency: 'CRC',
      },
    ],
  };
  const soloTextos: OfrendasPublicas = { lang: 'es', intro: { text: 'Cómo donar.' } };
  const conMinisterios = {
    lang: 'es' as const,
    items: [{ id: 'min-1', name: 'Alabanza', schedule: 'Todos los viernes' }],
  };

  it('con métodos publicados añade «Ofrendas y diezmos» al final, tras Ministerios', () => {
    const links = buildSiteLinks(completa, conEventos, conMinisterios, conMetodos);

    const ultimo = links[links.length - 1];
    expect(ultimo?.id).toBe('offerings');
    expect(ultimo?.href).toBe('/ofrendas');
    expect(ultimo?.messageKey).toBe('nav.sections.offerings');
    // Después de «Eventos y actividades» y de «Ministerios» (⟲ D-1).
    expect(links.map((link) => link.id).indexOf('offerings')).toBeGreaterThan(
      links.map((link) => link.id).indexOf('ministries'),
    );
  });

  it('con SOLO textos publicados (0 métodos) el ítem también existe: la sección vive (P7-13)', () => {
    expect(
      buildSiteLinks(completa, conEventos, conMinisterios, soloTextos).map((l) => l.id),
    ).toContain('offerings');
    // Con `usage` en vez de `intro` también; y sin ningún otro contenido.
    expect(
      buildSiteLinks(undefined, undefined, undefined, {
        lang: 'es',
        usage: { text: 'En qué se usa.' },
      }).map((l) => l.id),
    ).toEqual(['offerings']);
  });

  it('sin nada publicado (documento solo `lang` o ausente) el ítem no existe', () => {
    expect(
      buildSiteLinks(completa, conEventos, conMinisterios, { lang: 'es' }).map((l) => l.id),
    ).not.toContain('offerings');
    expect(
      buildSiteLinks(completa, conEventos, conMinisterios, undefined).map((l) => l.id),
    ).not.toContain('offerings');
    // `items: []` tampoco cuenta como contenido.
    expect(
      buildSiteLinks(completa, undefined, undefined, { lang: 'es', items: [] }).map((l) => l.id),
    ).not.toContain('offerings');
  });

  it('el resto de enlaces (anclas, /eventos y /ministerios) no cambia al añadir ofrendas', () => {
    const sinOfrendas = buildSiteLinks(completa, conEventos, conMinisterios);
    const conOfrendas = buildSiteLinks(completa, conEventos, conMinisterios, conMetodos);

    expect(conOfrendas.slice(0, sinOfrendas.length)).toEqual(sinOfrendas);
  });

  it('con contenido el ítem aparece en la barra y en el mapa del pie', async () => {
    server.use(
      http.get(portadaUrl, () => HttpResponse.json(completa)),
      http.get(ofrendasUrl, () => HttpResponse.json(conMetodos)),
    );

    render(
      <AppProviders>
        <MemoryRouter initialEntries={['/']}>
          <Routes>
            <Route element={<PublicLayout />}>
              <Route index element={<p>portada</p>} />
            </Route>
          </Routes>
        </MemoryRouter>
      </AppProviders>,
    );

    const nav = await screen.findByRole('navigation', { name: 'Secciones de la portada' });
    expect(await within(nav).findByRole('link', { name: 'Ofrendas y diezmos' })).toHaveAttribute(
      'href',
      '/ofrendas',
    );

    const sitemap = screen.getByRole('navigation', { name: 'Mapa del sitio' });
    expect(within(sitemap).getByRole('link', { name: 'Ofrendas y diezmos' })).toHaveAttribute(
      'href',
      '/ofrendas',
    );

    // Barra + mapa del pie: sin duplicar la lista del menú móvil.
    expect(document.querySelectorAll('a[href="/ofrendas"]')).toHaveLength(2);
  });

  it('sin nada publicado el ítem no está en ninguna superficie (FR-005/SC-010)', async () => {
    let peticiones = 0;
    server.use(
      http.get(portadaUrl, () => HttpResponse.json(completa)),
      http.get(ofrendasUrl, () => {
        peticiones += 1;
        return HttpResponse.json({ lang: 'es' });
      }),
    );

    render(
      <AppProviders>
        <MemoryRouter initialEntries={['/']}>
          <Routes>
            <Route element={<PublicLayout />}>
              <Route index element={<p>portada</p>} />
            </Route>
          </Routes>
        </MemoryRouter>
      </AppProviders>,
    );

    const nav = await screen.findByRole('navigation', { name: 'Secciones de la portada' });
    await within(nav).findByRole('link', { name: 'Quiénes somos' });
    await waitFor(() => expect(peticiones).toBeGreaterThan(0));

    expect(screen.queryByRole('link', { name: 'Ofrendas y diezmos' })).toBeNull();
    expect(document.querySelectorAll('a[href="/ofrendas"]')).toHaveLength(0);
  });
});
