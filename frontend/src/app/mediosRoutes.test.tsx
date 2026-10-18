import { render, screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../api/client';
import { server } from '../test/server';
import { AppProviders } from './providers';
import { AppRoutes } from './router';

/**
 * Ruta pública de F9 (T933): `/medios` es una **ruta real** sin sesión (FR-002)
 * y, sin nada publicable, redirige en silencio a la portada (Q10/P9-16/SC-016).
 * La entrada de navegación la cubre `useSiteLinks` (T934).
 */

const portadaUrl = `${API_BASE_URL}/api/v1/portada`;
const mediosUrl = `${API_BASE_URL}/api/v1/medios`;

function renderApp(entry: string) {
  return render(
    <AppProviders>
      <MemoryRouter initialEntries={[entry]}>
        <AppRoutes />
      </MemoryRouter>
    </AppProviders>,
  );
}

describe('AppRoutes · ruta /medios (F9)', () => {
  it('la sección pública vive en /medios y no requiere sesión', async () => {
    server.use(
      http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })),
      http.get(mediosUrl, () =>
        HttpResponse.json({
          lang: 'es',
          items: [
            {
              id: '11111111-1111-1111-1111-111111111111',
              title: 'La Fe y la confianza en el Abba',
              description: 'Confiar es un acto diario.',
              publishedAt: '2026-07-03T01:58:52Z',
              coverUrl: '/api/v1/media/img_abba.jpg',
              videoUrl: 'https://www.youtube.com/watch?v=abc12345678',
            },
          ],
          total: 1,
          limit: 12,
          offset: 0,
        }),
      ),
    );

    renderApp('/medios');

    expect(await screen.findByRole('heading', { level: 1, name: 'Medios' })).toBeInTheDocument();
    expect(screen.getByText('Confiar es un acto diario.')).toBeInTheDocument();
  });

  it('sin nada publicable la URL directa redirige a la portada (Q10)', async () => {
    server.use(
      http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })),
      http.get(mediosUrl, () =>
        HttpResponse.json({ lang: 'es', items: [], total: 0, limit: 12, offset: 0 }),
      ),
    );

    renderApp('/medios');

    // Llega a la portada y en ningún momento pinta un H1 de medios.
    expect(await screen.findByRole('heading', { name: 'Lo que viene' })).toBeInTheDocument();
    expect(screen.queryByRole('heading', { level: 1, name: 'Medios' })).toBeNull();
  });
});
