import { QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import { createMemoryRouter, RouterProvider } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { MedioPublico, MediosPublicos } from '../../../api/medios';
import { createQueryClientForTest } from '../../../test/queryClient';
import { installIntersectionObserver } from '../../../test/intersectionObserver';
import { installMatchMedia } from '../../../test/matchMedia';
import { server } from '../../../test/server';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { useLanguage } from '../i18n/useLanguage';
import { MediosPage } from './MediosPage';
import { PUBLIC_MEDIOS_QUERY_KEY } from './hooks/usePublicMedios';

/**
 * Página pública `/medios` (US1/US4/US5, ux.md §4.1): H1 + subtítulo que anuncia
 * la salida del sitio, una sola lista del más reciente al más antiguo —cada
 * ficha con **solo lo que tiene** (carátula o marcador neutro, nombre, fecha,
 * descripción y enlaces externos anunciados)—, «Cargar más» por tandas,
 * **redirección silenciosa** con 0 publicables (Q10/P9-16) y **0 recursos de
 * terceros** (SC-005).
 */

const publicUrl = `${API_BASE_URL}/api/v1/medios`;

function makeItem(overrides: Partial<MedioPublico> = {}): MedioPublico {
  return {
    id: 'med-1',
    title: 'Potencial sin Resultados La Lección de la Higuera',
    description: 'La higuera estéril y lo que Dios espera de nosotros.',
    publishedAt: '2026-05-25T18:43:42Z',
    coverUrl: '/api/v1/media/img_higuera.jpg',
    videoUrl: 'https://www.youtube.com/watch?v=abc12345678',
    audioUrl: 'https://podcasters.spotify.com/pod/show/simiente/episodes/higuera',
    ...overrides,
  };
}

/** Responde una tanda de la sección respetando `limit` (como el servidor). */
function responder(items: MedioPublico[], total = items.length) {
  server.use(
    http.get(publicUrl, ({ request }) => {
      const url = new URL(request.url);
      const limit = Number(url.searchParams.get('limit') ?? '12');
      return HttpResponse.json({
        lang: (url.searchParams.get('lang') ?? 'es') as MediosPublicos['lang'],
        items: items.slice(0, limit),
        total,
        limit,
        offset: Number(url.searchParams.get('offset') ?? '0'),
      } satisfies MediosPublicos);
    }),
  );
}

/** Permite cambiar de idioma desde la prueba (como el selector de F10). */
function LanguageProbe() {
  const { lang, setLang } = useLanguage();
  return (
    <button type="button" onClick={() => setLang(lang === 'es' ? 'en' : 'es')}>
      cambiar idioma
    </button>
  );
}

function renderPage(client = createQueryClientForTest()) {
  const router = createMemoryRouter(
    [
      { path: '/', element: <p>portada de la iglesia</p> },
      { path: '/medios', element: <MediosPage /> },
    ],
    { initialEntries: ['/medios'] },
  );

  const view = render(
    <QueryClientProvider client={client}>
      <LanguageProvider>
        <LanguageProbe />
        <RouterProvider router={router} />
      </LanguageProvider>
    </QueryClientProvider>,
  );

  return { ...view, router };
}

describe('MediosPage · sección pública (US1, FR-002…FR-006)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('pinta H1, el subtítulo que anuncia la salida y la lista en el orden del servidor', async () => {
    responder([
      makeItem({ id: 'med-1', title: 'La Fe y la confianza en el Abba' }),
      makeItem({ id: 'med-2', title: 'A Mayor Fe, Mayor Presión' }),
      makeItem({ id: 'med-3', title: 'Fe como fruto del Espíritu' }),
    ]);

    renderPage();

    expect(await screen.findByRole('heading', { level: 1, name: 'Medios' })).toBeVisible();
    expect(
      screen.getByText(
        'Las prédicas y los podcasts de nuestra iglesia; se abren en YouTube y Spotify, en una pestaña nueva.',
      ),
    ).toBeVisible();

    // El orden es el del servidor (fecha desc. con desempate de importación):
    // el cliente NO reordena.
    const fichas = screen.getAllByRole('listitem');
    expect(fichas).toHaveLength(3);
    expect(
      fichas.map((ficha) => within(ficha).getByRole('heading', { level: 3 }).textContent),
    ).toEqual([
      'La Fe y la confianza en el Abba',
      'A Mayor Fe, Mayor Presión',
      'Fe como fruto del Espíritu',
    ]);
  });

  it('cada ficha muestra carátula, nombre, fecha y descripción, y solo los enlaces que existen (SC-010)', async () => {
    responder([
      // Un solo formato → un solo botón, sin líneas vacías.
      makeItem({
        id: 'med-1',
        title: 'A Mayor Fe, Mayor Presión: ¿Estás Listo para Crecer?',
        audioUrl: undefined,
      }),
      // Los dos formatos → los dos botones.
      makeItem({ id: 'med-2', title: 'La paternidad de Dios y la fe' }),
    ]);

    const { container } = renderPage();
    await screen.findByRole('heading', { level: 1, name: 'Medios' });

    const fichas = screen.getAllByRole('listitem');

    // Ficha 1 · solo video.
    const soloVideo = fichas[0] as HTMLElement;
    expect(within(soloVideo).getByRole('heading', { level: 3 })).toHaveTextContent(
      'A Mayor Fe, Mayor Presión: ¿Estás Listo para Crecer?',
    );
    expect(within(soloVideo).getByText('lunes 25 de mayo')).toBeVisible();
    expect(
      within(soloVideo).getByText('La higuera estéril y lo que Dios espera de nosotros.'),
    ).toBeVisible();
    expect(within(soloVideo).getAllByRole('link')).toHaveLength(1);
    expect(within(soloVideo).queryByText('Escuchar en Spotify')).toBeNull();

    // Ficha 2 · los dos enlaces.
    const dosEnlaces = fichas[1] as HTMLElement;
    expect(within(dosEnlaces).getAllByRole('link')).toHaveLength(2);
    expect(within(dosEnlaces).getByText('Ver en YouTube')).toBeVisible();
    expect(within(dosEnlaces).getByText('Escuchar en Spotify')).toBeVisible();

    // La carátula es una copia servida por el sitio (nunca un recurso externo).
    const caratula = within(fichas[0] as HTMLElement).getByRole('img', {
      name: 'Carátula del episodio A Mayor Fe, Mayor Presión: ¿Estás Listo para Crecer?',
    });
    expect(caratula).toHaveAttribute('src', `${API_BASE_URL}/api/v1/media/img_higuera.jpg`);

    // 0 recursos de terceros en toda la sección (SC-005).
    expect(container.querySelectorAll('iframe')).toHaveLength(0);
    for (const node of container.querySelectorAll('[src], [href]')) {
      const value = node.getAttribute('src') ?? node.getAttribute('href') ?? '';
      if (node.tagName === 'A') {
        continue;
      }
      expect(value).not.toMatch(/youtube\.com|spotify\.com|anchor\.fm/);
    }
  });

  it('sin carátula pinta el marcador neutro y sin descripción no deja esa línea (D-4/D-6, precisión 3)', async () => {
    responder([
      makeItem({ id: 'med-1', coverUrl: undefined, description: undefined }),
      makeItem({ id: 'med-2' }),
    ]);

    const { container } = renderPage();
    await screen.findByRole('heading', { level: 1, name: 'Medios' });

    const fichas = screen.getAllByRole('listitem');

    // Marcador neutro en el mismo marco 16:9 (nunca un hueco).
    const neutra = fichas[0] as HTMLElement;
    expect(neutra.querySelector('[data-testid="caratula-neutra"]')).not.toBeNull();
    expect(within(neutra).queryByRole('img')).toBeNull();

    // Sin descripción: solo la fecha queda como línea de texto (ni párrafo vacío).
    expect(neutra.querySelectorAll('p')).toHaveLength(1);
    expect(container.querySelector('[data-testid="caratula-neutra"]')).not.toBeNull();
    expect((fichas[1] as HTMLElement).querySelector('[data-testid="caratula-neutra"]')).toBeNull();
  });

  it('los enlaces externos se anuncian completos y abren en pestaña nueva (FR-005/SC-006)', async () => {
    responder([makeItem({ title: 'La Fe y la herida' })]);

    renderPage();
    await screen.findByRole('heading', { level: 1, name: 'Medios' });

    const youtube = screen.getByRole('link', {
      name: 'Abre YouTube en una pestaña nueva para ver La Fe y la herida; sales del sitio de la iglesia.',
    });
    expect(youtube).toHaveAttribute('target', '_blank');
    expect(youtube).toHaveAttribute('rel', 'noopener noreferrer');
    expect(youtube).toHaveAttribute('href', 'https://www.youtube.com/watch?v=abc12345678');

    const spotify = screen.getByRole('link', {
      name: 'Abre Spotify en una pestaña nueva para escuchar La Fe y la herida; sales del sitio de la iglesia.',
    });
    expect(spotify).toHaveAttribute('target', '_blank');
    expect(spotify).toHaveAttribute('rel', 'noopener noreferrer');

    // El indicador visible de salida viaja dentro del propio enlace (D-7).
    expect(within(youtube).getByText('(pestaña nueva)')).toBeInTheDocument();
    // El nombre no es un enlace: los únicos interactivos son los botones.
    expect(
      screen.getByRole('heading', { level: 3, name: 'La Fe y la herida' }).querySelector('a'),
    ).toBeNull();
  });

  it('«Cargar más» agrega la siguiente tanda y desaparece al agotarse (D-3/FR-006)', async () => {
    const items = Array.from({ length: 13 }, (_, index) =>
      makeItem({ id: `med-${index}`, title: `Episodio ${index}` }),
    );
    responder(items);

    renderPage();
    await screen.findByRole('heading', { level: 1, name: 'Medios' });

    expect(screen.getAllByRole('listitem')).toHaveLength(12);
    const cargarMas = screen.getByRole('button', { name: 'Cargar más' });

    await userEvent.click(cargarMas);

    await waitFor(() => expect(screen.getAllByRole('listitem')).toHaveLength(13));
    // Con todo mostrado el botón desaparece (sin mensaje «no hay más»).
    expect(screen.queryByRole('button', { name: 'Cargar más' })).toBeNull();
  });

  it('con 12 o menos publicables no hay «Cargar más»', async () => {
    responder([makeItem(), makeItem({ id: 'med-2' })]);

    renderPage();
    await screen.findByRole('heading', { level: 1, name: 'Medios' });

    expect(screen.queryByRole('button', { name: 'Cargar más' })).toBeNull();
  });

  it('en inglés se traduce la interfaz y el contenido sin inglés se ve en español (US4/FR-025/FR-026)', async () => {
    window.localStorage.setItem('ss.lang', 'en');
    responder([
      makeItem({
        title: 'Fe como fruto del Espíritu',
        description: 'El fruto del Espíritu se cultiva en comunidad.',
      }),
    ]);

    renderPage();

    expect(await screen.findByRole('heading', { level: 1, name: 'Media' })).toBeVisible();
    expect(
      screen.getByText(
        "Watch and listen to our church's sermons and podcast; they open in YouTube and Spotify in a new tab.",
      ),
    ).toBeVisible();
    expect(screen.getByText('Watch on YouTube')).toBeVisible();
    expect(screen.getByText('Listen on Spotify')).toBeVisible();

    // Datos no traducibles idénticos: nombre, fecha y enlaces (FR-026).
    expect(
      screen.getByRole('heading', { level: 3, name: 'Fe como fruto del Espíritu' }),
    ).toBeVisible();
    expect(screen.getByText('Monday, May 25')).toBeVisible();
    expect(
      screen.getByRole('link', {
        name: "Opens YouTube in a new tab to watch Fe como fruto del Espíritu; leaves the church's website.",
      }),
    ).toHaveAttribute('href', 'https://www.youtube.com/watch?v=abc12345678');

    // El contenido sin inglés se muestra en español, sin marca de «sin traducir».
    expect(screen.getByText('El fruto del Espíritu se cultiva en comunidad.')).toBeVisible();
    expect(screen.queryByText(/sin traducir|not translated/i)).toBeNull();
  });

  it('con movimiento reducido el contenido está íntegro y sin revelados pendientes (FR-028)', async () => {
    installMatchMedia({ '(prefers-reduced-motion: reduce)': true });
    installIntersectionObserver();
    responder([makeItem()]);

    const { container } = renderPage();
    await screen.findByRole('heading', { level: 1, name: 'Medios' });

    const revelados = container.querySelectorAll('[data-reveal]');
    expect(revelados.length).toBeGreaterThan(0);
    for (const nodo of revelados) {
      expect(nodo.getAttribute('data-reveal')).toBe('in');
    }
  });
});

describe('MediosPage · redirección silenciosa con 0 publicables (Q10/P9-16/SC-016)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('redirige a la portada con `replace`, sin H1, sin lista, sin aviso ni mensaje', async () => {
    responder([], 0);

    const { router } = renderPage();

    expect(await screen.findByText('portada de la iglesia')).toBeVisible();
    expect(router.state.location.pathname).toBe('/');
    expect(router.state.historyAction).toBe('REPLACE');

    expect(screen.queryByRole('heading', { level: 1, name: 'Medios' })).toBeNull();
    expect(screen.queryByRole('list')).toBeNull();
    expect(screen.queryByRole('alert')).toBeNull();
    expect(screen.queryByText(/no hay|vací|no pudimos/i)).toBeNull();
  });

  it('con la respuesta ya en caché redirige en el primer render: sin esqueleto ni destello', async () => {
    const client = createQueryClientForTest();
    client.setQueryData<MediosPublicos>([...PUBLIC_MEDIOS_QUERY_KEY, 'es', 12, 0], {
      lang: 'es',
      items: [],
      total: 0,
      limit: 12,
      offset: 0,
    });
    responder([], 0);

    renderPage(client);

    expect(screen.getByText('portada de la iglesia')).toBeVisible();
    expect(screen.queryByRole('heading', { level: 1, name: 'Medios' })).toBeNull();
    expect(screen.queryByRole('status')).toBeNull();
    expect(screen.queryByRole('list')).toBeNull();
  });

  it('mientras carga solo se ve el esqueleto estándar (nunca la lista vacía)', async () => {
    server.use(
      http.get(publicUrl, async () => {
        await delay(50);
        return HttpResponse.json({
          lang: 'es',
          items: [makeItem()],
          total: 1,
          limit: 12,
          offset: 0,
        } satisfies MediosPublicos);
      }),
    );

    renderPage();

    const esqueleto = screen.getByRole('status');
    expect(esqueleto).toHaveAttribute('aria-busy', 'true');
    expect(screen.queryByRole('heading', { level: 1, name: 'Medios' })).toBeNull();
    expect(screen.queryByRole('list')).toBeNull();

    expect(await screen.findByRole('heading', { level: 1, name: 'Medios' })).toBeVisible();
  });

  it('ante error NO redirige: mensaje bilingüe y «Volver a intentar» (D-12)', async () => {
    let fallos = 1;
    server.use(
      http.get(publicUrl, () => {
        if (fallos > 0) {
          fallos -= 1;
          return HttpResponse.json(
            { error: { code: 'internal', message: 'Error' } },
            { status: 500 },
          );
        }
        return HttpResponse.json({
          lang: 'es',
          items: [makeItem()],
          total: 1,
          limit: 12,
          offset: 0,
        } satisfies MediosPublicos);
      }),
    );

    const { router } = renderPage();

    expect(
      await screen.findByText(
        'No pudimos cargar los medios. Revisa tu conexión y vuelve a intentarlo.',
      ),
    ).toBeVisible();
    expect(screen.getByRole('heading', { level: 1, name: 'Medios' })).toBeVisible();
    expect(screen.queryByText('portada de la iglesia')).toBeNull();
    expect(router.state.location.pathname).toBe('/medios');

    await userEvent.click(screen.getByRole('button', { name: 'Volver a intentar' }));
    expect(await screen.findByRole('listitem')).toBeVisible();
  });

  it('con error y contenido en caché conserva la lista y avisa en discreto', async () => {
    let peticiones = 0;
    server.use(
      http.get(publicUrl, () => {
        peticiones += 1;
        if (peticiones === 1) {
          return HttpResponse.json({
            lang: 'es',
            items: [makeItem()],
            total: 1,
            limit: 12,
            offset: 0,
          } satisfies MediosPublicos);
        }
        return HttpResponse.json(
          { error: { code: 'internal', message: 'Error' } },
          { status: 500 },
        );
      }),
    );

    renderPage();
    await screen.findByRole('listitem');

    // Un cambio de idioma fuerza una segunda petición que falla.
    await userEvent.click(screen.getByRole('button', { name: 'cambiar idioma' }));

    expect(await screen.findByText(/couldn't refresh/)).toBeVisible();
    expect(screen.getByRole('listitem')).toBeVisible();
  });
});
