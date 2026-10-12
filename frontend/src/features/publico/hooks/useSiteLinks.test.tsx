import { render, screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { PortadaPublica } from '../../../api/portada';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { PublicLayout } from '../components/PublicLayout';
import { PUBLIC_SECTIONS } from '../sections';
import { buildSiteLinks } from './useSiteLinks';

const portadaUrl = `${API_BASE_URL}/api/v1/portada`;
const upcomingUrl = `${API_BASE_URL}/api/v1/proximamente`;

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

  it('las anclas de la portada no cambian al añadir el ítem de ruta (F4)', () => {
    const conRuta = buildSiteLinks(completa, conEventos);
    const sinRuta = buildSiteLinks(completa);

    expect(conRuta.slice(0, sinRuta.length)).toEqual(sinRuta);
    for (const link of sinRuta) {
      expect(link.href.startsWith('#')).toBe(true);
      expect(PUBLIC_SECTIONS.some((section) => `#${section.anchor}` === link.href)).toBe(true);
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

    // El resto de enlaces de la portada sigue siendo un ancla existente.
    expect(within(nav).getByRole('link', { name: 'Quiénes somos' })).toHaveAttribute(
      'href',
      '#quienes-somos',
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
});
