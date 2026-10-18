import { render, screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { MedioPublico } from '../../../api/medios';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { MediosBlock } from './MediosBlock';

/**
 * Bloque «Medios» de la portada (US6/FR-027/SC-015, ux.md §4.9): hasta **5**
 * fichas compactas en el orden de la sección —con sus enlaces externos
 * anunciados y **sin fecha** (⟲ D-32)— y el botón «Ver todo» hacia `/medios`;
 * **sin publicables el bloque no se dibuja** (sin título, sin mensaje, sin
 * hueco) y tampoco ante error sin caché.
 */

const portadaUrl = `${API_BASE_URL}/api/v1/portada/medios`;

function item(id: string, overrides: Partial<MedioPublico> = {}): MedioPublico {
  return {
    id,
    title: `Episodio ${id}`,
    description: `Descripción del episodio ${id}.`,
    publishedAt: '2026-05-25T18:43:42Z',
    coverUrl: `/api/v1/media/img_${id}.jpg`,
    videoUrl: `https://www.youtube.com/watch?v=${id.padEnd(11, 'x')}`,
    audioUrl: 'https://podcasters.spotify.com/pod/show/simiente/episodes/x',
    ...overrides,
  };
}

function renderBlock() {
  return render(
    <AppProviders>
      <LanguageProvider>
        <MediosBlock />
      </LanguageProvider>
    </AppProviders>,
  );
}

describe('MediosBlock · bloque «Medios» de la portada (US6/FR-027/SC-015)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('con 0 publicables el bloque NO se dibuja: sin título ni mensaje', async () => {
    server.use(http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })));

    const { container } = renderBlock();

    await waitFor(() => expect(container).toBeEmptyDOMElement());
  });

  it('con 8 publicables muestra exactamente los 5 del servidor, en su orden (tope del servidor)', async () => {
    // El servidor aplica el tope 5; el cliente NO recorta ni reordena.
    const items = Array.from({ length: 8 }, (_, index) => item(`med-${index}`)).slice(0, 5);
    server.use(http.get(portadaUrl, () => HttpResponse.json({ lang: 'es', items })));

    renderBlock();

    expect(await screen.findByRole('heading', { level: 2, name: 'Medios' })).toBeInTheDocument();
    const fichas = screen.getAllByRole('listitem');
    expect(fichas).toHaveLength(5);
    expect(
      fichas.map((ficha) => within(ficha).getByRole('heading', { level: 3 }).textContent),
    ).toEqual([
      'Episodio med-0',
      'Episodio med-1',
      'Episodio med-2',
      'Episodio med-3',
      'Episodio med-4',
    ]);
    expect(screen.getByRole('link', { name: 'Ver todo' })).toHaveAttribute('href', '/medios');
  });

  it('con menos de 5 muestra exactamente los que hay, sin huecos ni marcadores vacíos', async () => {
    server.use(
      http.get(portadaUrl, () =>
        HttpResponse.json({ lang: 'es', items: [item('uno'), item('dos'), item('tres')] }),
      ),
    );

    renderBlock();

    await screen.findByRole('heading', { level: 2, name: 'Medios' });
    expect(screen.getAllByRole('listitem')).toHaveLength(3);
    expect(screen.getByText('Descripción del episodio uno.')).toBeInTheDocument();
  });

  it('cada ficha lleva sus enlaces anunciados, sin fecha y sin recursos de terceros (D-32/SC-005)', async () => {
    server.use(
      http.get(portadaUrl, () =>
        HttpResponse.json({
          lang: 'es',
          items: [
            item('uno'),
            // Sin carátula y sin audio: solo el botón de YouTube y el marcador neutro.
            item('dos', { coverUrl: undefined, audioUrl: undefined, description: undefined }),
          ],
        }),
      ),
    );

    const { container } = renderBlock();
    await screen.findByRole('heading', { level: 2, name: 'Medios' });

    const fichas = screen.getAllByRole('listitem');
    expect(
      within(fichas[0] as HTMLElement).getByRole('link', {
        name: 'Abre YouTube en una pestaña nueva para ver Episodio uno; sales del sitio de la iglesia.',
      }),
    ).toHaveAttribute('target', '_blank');
    expect(
      within(fichas[0] as HTMLElement).getByRole('link', {
        name: 'Abre Spotify en una pestaña nueva para escuchar Episodio uno; sales del sitio de la iglesia.',
      }),
    ).toBeInTheDocument();

    // La fecha se omite en el bloque (⟲ D-32): ni la fecha ni el año.
    expect(screen.queryByText('lunes 25 de mayo')).toBeNull();

    // Ficha mínima: marcador neutro, sin descripción y con un solo enlace.
    const minima = fichas[1] as HTMLElement;
    expect(minima.querySelector('[data-testid="caratula-neutra"]')).not.toBeNull();
    expect(within(minima).getAllByRole('link')).toHaveLength(1);
    expect(within(minima).queryByText(/Descripción/)).toBeNull();

    expect(container.querySelectorAll('iframe')).toHaveLength(0);
    for (const nodo of container.querySelectorAll('[src]')) {
      expect(nodo.getAttribute('src')).not.toMatch(/youtube\.com|spotify\.com|anchor\.fm/);
    }
  });

  it('en inglés la interfaz se traduce y el contenido sin inglés se ve en español (US6 esc. 7)', async () => {
    window.localStorage.setItem('ss.lang', 'en');
    server.use(
      http.get(portadaUrl, () =>
        HttpResponse.json({
          lang: 'en',
          items: [item('uno', { title: 'Fe como fruto del Espíritu' })],
        }),
      ),
    );

    renderBlock();

    expect(await screen.findByRole('heading', { level: 2, name: 'Media' })).toBeInTheDocument();
    expect(screen.getByText('The latest sermons and episodes')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'See all' })).toBeInTheDocument();
    expect(
      screen.getByRole('link', {
        name: "Opens YouTube in a new tab to watch Fe como fruto del Espíritu; leaves the church's website.",
      }),
    ).toBeInTheDocument();
  });

  it('mientras carga muestra un esqueleto con área de carátula', () => {
    server.use(http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })));

    renderBlock();

    expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
  });

  it('ante error sin caché el bloque no se dibuja (ni título huérfano)', async () => {
    server.use(
      http.get(portadaUrl, () =>
        HttpResponse.json({ error: { code: 'internal', message: 'x' } }, { status: 500 }),
      ),
    );

    const { container } = renderBlock();

    await waitFor(() => expect(container).toBeEmptyDOMElement());
  });
});
