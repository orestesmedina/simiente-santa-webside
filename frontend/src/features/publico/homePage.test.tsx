import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { PortadaPublica } from '../../api/portada';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { PublicLayout } from './components/PublicLayout';
import { HomePage } from './pages/HomePage';

const portadaUrl = `${API_BASE_URL}/api/v1/portada`;

const completa: PortadaPublica = {
  lang: 'es',
  identity: {
    name: 'Iglesia Simiente Santa',
    tagline: 'Un espacio para encontrarse con Dios',
    mission: 'Nuestra misión real',
    vision: 'Nuestra visión real',
    logoUrl: '/api/v1/media/img_logo.png',
    logoAlt: 'Logotipo de la iglesia',
    coverImageUrl: '/api/v1/media/img_cover.png',
    coverImageAlt: 'Imagen de la portada',
  },
  about: { text: 'Línea uno\nLínea dos' },
  schedule: [
    {
      id: 'srv-1',
      dayOfWeek: 0,
      startTime: '10:00',
      endTime: '12:00',
      name: 'Culto dominical',
      place: 'Templo central',
    },
    {
      id: 'srv-2',
      dayOfWeek: 3,
      startTime: '18:00',
      endTime: null,
      name: 'Oración',
      place: 'Salón',
    },
  ],
  whatsapp: [
    { id: 'wa-1', name: 'Escríbenos', kind: 'direct', url: 'https://wa.me/50688888888' },
    {
      id: 'wa-2',
      name: 'Grupo de la iglesia',
      kind: 'group',
      url: 'https://chat.whatsapp.com/ABCDEF',
    },
  ],
  contact: { address: 'San José, Costa Rica', email: 'hola@ejemplo.com', phone: '+506 8888 8888' },
  socials: [
    { id: 'so-1', network: 'facebook', url: 'https://facebook.com/simiente' },
    { id: 'so-2', network: 'instagram', url: 'https://instagram.com/simiente' },
  ],
};

function responder(data: PortadaPublica) {
  server.use(http.get(portadaUrl, () => HttpResponse.json(data)));
}

function renderHome() {
  return render(
    <AppProviders>
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route element={<PublicLayout />}>
            <Route index element={<HomePage />} />
          </Route>
        </Routes>
      </MemoryRouter>
    </AppProviders>,
  );
}

describe('HomePage · secciones publicadas (US1, SC-012)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('compone solo las secciones presentes; las vacías no dejan rastro', async () => {
    responder({
      lang: 'es',
      identity: { name: 'Iglesia Simiente Santa' },
      about: { text: 'Somos una familia' },
      schedule: [{ id: 'srv-1', dayOfWeek: 0, startTime: '10:00', name: 'Culto', place: 'Templo' }],
    });

    renderHome();

    expect(
      await screen.findByRole('heading', { name: 'Iglesia Simiente Santa' }),
    ).toBeInTheDocument();
    expect(screen.getByText('Somos una familia')).toBeInTheDocument();
    expect(screen.getAllByText('Culto').length).toBeGreaterThan(0);

    // WhatsApp, contacto y redes no aparecen (ni sección ni ancla).
    expect(screen.queryByRole('heading', { name: 'WhatsApp' })).toBeNull();
    expect(screen.queryByRole('heading', { name: 'Contacto' })).toBeNull();
    expect(screen.queryByRole('heading', { name: 'Redes sociales' })).toBeNull();

    const nav = screen.getByRole('navigation', { name: 'Secciones de la portada' });
    expect(within(nav).getByRole('link', { name: 'Quiénes somos' })).toBeInTheDocument();
    expect(within(nav).getByRole('link', { name: 'Horario de servicios' })).toBeInTheDocument();
    expect(within(nav).queryByRole('link', { name: 'WhatsApp' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Contacto' })).toBeNull();
  });

  it('con la respuesta vacía no pinta ninguna sección de F3; el bloque de F4 sí (SC-012)', async () => {
    responder({ lang: 'es' });

    renderHome();

    expect(
      await screen.findByRole('navigation', { name: 'Secciones de la portada' }),
    ).toBeInTheDocument();
    // El bloque «Lo que viene» es la **única excepción** a «ocultar secciones
    // vacías» (Q8/FR-026): se pinta siempre, con su mensaje cuando no hay ítems.
    expect(
      await screen.findByRole('heading', { level: 2, name: 'Lo que viene' }),
    ).toBeInTheDocument();
    // Espera al dato: el mensaje solo aparece cuando la consulta se resuelve.
    expect(
      await screen.findByText('Por el momento no hay eventos ni actividades.'),
    ).toBeInTheDocument();
    expect(
      screen.getAllByRole('heading', { level: 2 }).map((heading) => heading.textContent),
    ).toEqual(['Lo que viene']);
  });

  it('pone el bloque «Lo que viene» tras el hero y antes de «quiénes somos» sin alterar F3 (F4/US7)', async () => {
    responder(completa);
    server.use(
      http.get(`${API_BASE_URL}/api/v1/portada/proximamente`, () =>
        HttpResponse.json({
          lang: 'es',
          items: [
            {
              kind: 'event',
              id: 'ev-1',
              title: 'Vigilia de viernes',
              date: '2026-10-17',
              startTime: '19:00',
              place: 'Templo central',
            },
            {
              kind: 'activity',
              id: 'ac-1',
              title: 'Repartir comida',
              date: '2026-10-18',
              place: 'Comedor',
            },
          ],
        }),
      ),
    );

    renderHome();

    await screen.findByRole('link', { name: 'Evento: Vigilia de viernes' });

    // Ubicación (ajuste del 2026-10-12): inmediatamente después del hero y
    // **antes** de «quiénes somos»; el resto del orden de F3 no cambia.
    const hero = screen.getByRole('heading', { level: 1, name: 'Iglesia Simiente Santa' });
    const bloque = screen.getByRole('heading', { level: 2, name: 'Lo que viene' });
    const quienesSomos = screen.getByRole('heading', { level: 2, name: 'Quiénes somos' });
    const horario = screen.getByRole('heading', { level: 2, name: 'Horario de servicios' });
    expect(hero.compareDocumentPosition(bloque)).toBe(Node.DOCUMENT_POSITION_FOLLOWING);
    expect(bloque.compareDocumentPosition(quienesSomos)).toBe(Node.DOCUMENT_POSITION_FOLLOWING);
    expect(quienesSomos.compareDocumentPosition(horario)).toBe(Node.DOCUMENT_POSITION_FOLLOWING);

    // Los ítems del extracto, etiquetados y enlazados a su bloque de /eventos.
    expect(screen.getByText('Evento')).toBeInTheDocument();
    expect(screen.getByText('Actividad')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Actividad: Repartir comida' })).toHaveAttribute(
      'href',
      '/eventos#actividades',
    );
    // Botón único «paso 2» del bloque (SC-015).
    expect(screen.getByRole('link', { name: 'Ver más eventos' })).toHaveAttribute(
      'href',
      '/eventos',
    );

    // Las secciones de F3 siguen renderizándose igual que antes.
    expect(screen.getByText(/Línea uno\s+Línea dos/)).toBeInTheDocument();
    expect(screen.getByRole('table')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Escribir por WhatsApp/ })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'hola@ejemplo.com' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Ver en Facebook/ })).toBeInTheDocument();
  });

  it('inserta el bloque «Ministerios» tras el de eventos y antes de «quiénes somos» (F6/US6)', async () => {
    responder(completa);
    server.use(
      http.get(`${API_BASE_URL}/api/v1/portada/proximamente`, () =>
        HttpResponse.json({
          lang: 'es',
          items: [
            { kind: 'event', id: 'ev-1', title: 'Vigilia', date: '2026-10-17', place: 'Templo' },
          ],
        }),
      ),
      http.get(`${API_BASE_URL}/api/v1/portada/ministerios`, () =>
        HttpResponse.json({
          lang: 'es',
          items: [{ id: 'min-1', name: 'Alabanza', schedule: 'Todos los viernes' }],
        }),
      ),
    );

    renderHome();

    await screen.findByRole('link', { name: 'Ministerio: Alabanza' });

    // Una sola inserción: entre el bloque de F4 y «quiénes somos», sin tocar el
    // resto de la secuencia de F3 (FR-021/FR-029).
    const eventos = screen.getByRole('heading', { level: 2, name: 'Lo que viene' });
    const ministerios = screen.getByRole('heading', { level: 2, name: 'Ministerios' });
    const quienesSomos = screen.getByRole('heading', { level: 2, name: 'Quiénes somos' });
    const horario = screen.getByRole('heading', { level: 2, name: 'Horario de servicios' });
    expect(eventos.compareDocumentPosition(ministerios)).toBe(Node.DOCUMENT_POSITION_FOLLOWING);
    expect(ministerios.compareDocumentPosition(quienesSomos)).toBe(
      Node.DOCUMENT_POSITION_FOLLOWING,
    );
    expect(quienesSomos.compareDocumentPosition(horario)).toBe(Node.DOCUMENT_POSITION_FOLLOWING);

    // Único CTA del bloque (paso 2 hacia la sección completa).
    expect(screen.getByRole('link', { name: 'Ver más ministerios' })).toHaveAttribute(
      'href',
      '/ministerios',
    );

    // Las secciones de F3 siguen renderizándose igual que antes.
    expect(screen.getByText(/Línea uno\s+Línea dos/)).toBeInTheDocument();
    expect(screen.getByRole('table')).toBeInTheDocument();
  });

  it('sin ministerios publicados la portada no gana sección ni mensaje (FR-028)', async () => {
    responder(completa);
    let peticiones = 0;
    server.use(
      http.get(`${API_BASE_URL}/api/v1/portada/ministerios`, () => {
        peticiones += 1;
        return HttpResponse.json({ lang: 'es' });
      }),
    );

    renderHome();
    await screen.findByRole('heading', { name: 'Iglesia Simiente Santa' });
    // Se espera a que el bloque haya respondido: aun así no dibuja nada.
    await waitFor(() => expect(peticiones).toBe(1));
    // La portada ya está pintada por completo y el bloque no repite la petición:
    // una sola petición paga bloque + barra + menú + pie (P6-9).
    await screen.findByRole('table');
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(peticiones).toBe(1);

    expect(screen.queryByRole('heading', { level: 2, name: 'Ministerios' })).toBeNull();
    expect(screen.queryByRole('link', { name: 'Ver más ministerios' })).toBeNull();

    // Y las demás secciones de la portada se renderizan igual que antes.
    expect(screen.getByRole('heading', { level: 2, name: 'Quiénes somos' })).toBeInTheDocument();
    expect(screen.getByRole('table')).toBeInTheDocument();
  });
});

describe('HomePage · contenido y enlaces (US1, SC-009)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('muestra identidad, misión/visión/lema y las imágenes con su alt (FR-019)', async () => {
    responder(completa);

    renderHome();
    await screen.findByRole('heading', { name: 'Iglesia Simiente Santa' });

    const hero = screen.getByRole('region', { name: 'Iglesia Simiente Santa' });
    expect(within(hero).getByRole('heading', { name: 'Nuestra misión' })).toBeInTheDocument();
    expect(within(hero).getByText('Nuestra misión real')).toBeInTheDocument();
    expect(within(hero).getByText('Nuestra visión real')).toBeInTheDocument();
    expect(within(hero).getByText('Un espacio para encontrarse con Dios')).toBeInTheDocument();
    const logo = within(hero).getByRole('img', { name: 'Logotipo de la iglesia' });
    expect(logo).toBeInTheDocument();
    // Las URLs relativas de la API se resuelven contra la base de la API (bug F3).
    expect(logo).toHaveAttribute('src', `${API_BASE_URL}/api/v1/media/img_logo.png`);
    const cover = within(hero).getByRole('img', { name: 'Imagen de la portada' });
    expect(cover).toBeInTheDocument();
    expect(cover).toHaveAttribute('src', `${API_BASE_URL}/api/v1/media/img_cover.png`);
  });

  it('el horario muestra el rango con endTime y solo la hora de inicio sin él (C2)', async () => {
    responder(completa);

    renderHome();
    await screen.findByRole('table');

    const tabla = screen.getByRole('table');
    // Presentación localizada a.m./p.m. a partir del dato de 24 h (ux.md §4.6/D-3).
    expect(within(tabla).getByText('10:00 a. m. − 12:00 m.')).toBeInTheDocument();
    expect(within(tabla).getByText('6:00 p. m.')).toBeInTheDocument();
    // El día se localiza desde dayOfWeek (domingo = 0, miércoles = 3).
    expect(within(tabla).getByText('Domingo')).toBeInTheDocument();
    expect(within(tabla).getByText('Miércoles')).toBeInTheDocument();
  });

  it('los enlaces de WhatsApp y redes apuntan al destino y abren fuera (SC-009)', async () => {
    responder(completa);

    renderHome();
    await screen.findByRole('heading', { name: 'WhatsApp' });

    const dm = screen.getByRole('link', { name: /Escribir por WhatsApp/ });
    expect(dm).toHaveAttribute('href', 'https://wa.me/50688888888');
    expect(dm).toHaveAttribute('target', '_blank');
    expect(dm).toHaveAttribute('rel', 'noopener noreferrer');

    const grupo = screen.getByRole('link', { name: /Entrar al grupo/ });
    expect(grupo).toHaveAttribute('href', 'https://chat.whatsapp.com/ABCDEF');

    const facebook = screen.getByRole('link', { name: /Ver en Facebook/ });
    expect(facebook).toHaveAttribute('href', 'https://facebook.com/simiente');
    expect(facebook).toHaveAttribute('target', '_blank');
  });

  it('el contacto ofrece enlaces nativos de correo y teléfono', async () => {
    responder(completa);

    renderHome();
    await screen.findByRole('heading', { name: 'Contacto' });

    expect(screen.getByRole('link', { name: 'hola@ejemplo.com' })).toHaveAttribute(
      'href',
      'mailto:hola@ejemplo.com',
    );
    expect(screen.getByRole('link', { name: '+506 8888 8888' })).toHaveAttribute(
      'href',
      'tel:+50688888888',
    );
  });
});

describe('HomePage · idioma (US4, SC-006)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('cambia la interfaz al instante y pide los contenidos en el idioma activo', async () => {
    const urls: string[] = [];
    server.use(
      http.get(portadaUrl, ({ request }) => {
        urls.push(request.url);
        return HttpResponse.json({
          lang: 'es',
          about: { text: 'Contenido solo en español' },
        });
      }),
    );

    renderHome();
    await screen.findByText('Contenido solo en español');

    await userEvent.click(screen.getAllByRole('radio', { name: 'Inglés' })[0]);

    // Interfaz en inglés (etiqueta de la navegación)…
    expect(await screen.findByRole('navigation', { name: 'Page sections' })).toBeInTheDocument();
    // …y el contenido sin inglés se muestra en español, tal como lo resolvió el
    // servidor (analyze I2: el cliente no reimplementa el fallback).
    expect(screen.getByText('Contenido solo en español')).toBeInTheDocument();
    expect(urls.some((url) => url.includes('lang=en'))).toBe(true);
  });
});

describe('HomePage · estados', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('muestra esqueletos con anuncio mientras carga', async () => {
    server.use(
      http.get(portadaUrl, async () => {
        await delay(50);
        return HttpResponse.json({ lang: 'es', identity: { name: 'Iglesia Simiente Santa' } });
      }),
    );

    renderHome();

    expect(screen.getAllByText('Cargando…').length).toBeGreaterThan(0);
    expect(
      await screen.findByRole('heading', { name: 'Iglesia Simiente Santa' }),
    ).toBeInTheDocument();
  });

  it('sin contenido previo, el error ofrece volver a intentar', async () => {
    let fallos = 1;
    server.use(
      http.get(portadaUrl, () => {
        if (fallos > 0) {
          fallos -= 1;
          return HttpResponse.json(
            { error: { code: 'internal', message: 'Error' } },
            { status: 500 },
          );
        }
        return HttpResponse.json({ lang: 'es', identity: { name: 'Iglesia Simiente Santa' } });
      }),
    );

    renderHome();

    expect(
      await screen.findByText(/No pudimos cargar la información de la iglesia/),
    ).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Volver a intentar' }));

    expect(
      await screen.findByRole('heading', { name: 'Iglesia Simiente Santa' }),
    ).toBeInTheDocument();
  });
});

describe('HomePage · movimiento (F10 T1026/FR-005/FR-008)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('monta cada sección dentro de un bloque de revelado', async () => {
    responder(completa);

    const { container } = renderHome();
    await screen.findByRole('heading', { name: 'Iglesia Simiente Santa' });

    for (const anchor of ['quienes-somos', 'horario', 'whatsapp', 'contacto', 'redes']) {
      const section = container.querySelector(`#${anchor}`);
      expect(section, `sección ${anchor}`).not.toBeNull();
      expect(section?.closest('.reveal-block'), `${anchor} dentro de un bloque`).not.toBeNull();
    }
    expect(container.querySelector('.reveal-group')).not.toBeNull();
  });

  it('el hero tiene su propia entrada y no se re-anima al cambiar de idioma', async () => {
    responder(completa);

    const { container } = renderHome();
    await screen.findByRole('heading', { name: 'Iglesia Simiente Santa' });
    expect(
      container.querySelector('section[aria-labelledby="portada-nombre"]')?.className,
    ).toContain('hero-enter');

    const who = container.querySelector('#quienes-somos');
    expect(who).not.toBeNull();

    await userEvent.click(screen.getAllByRole('radio', { name: 'Inglés' })[0]);
    expect(await screen.findByRole('navigation', { name: 'Page sections' })).toBeInTheDocument();

    // Claves React estables: React reutiliza el mismo nodo (no re-monta → no re-anima).
    expect(container.querySelector('#quienes-somos')).toBe(who);
  });
});
