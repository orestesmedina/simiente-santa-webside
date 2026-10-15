import { render, screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../api/client';
import { server } from '../test/server';
import { AppProviders } from './providers';
import { AppRoutes } from './router';

/**
 * Ruta pública de F7 (T726): `/ofrendas` es una **ruta real** sin sesión
 * (FR-006) y, sin nada publicado, redirige en silencio a la portada (Q9/P7-21).
 * La entrada de navegación la cubre `useSiteLinks` (T727).
 */

const portadaUrl = `${API_BASE_URL}/api/v1/portada`;
const ofrendasUrl = `${API_BASE_URL}/api/v1/ofrendas`;

function renderApp(entry: string) {
  return render(
    <AppProviders>
      <MemoryRouter initialEntries={[entry]}>
        <AppRoutes />
      </MemoryRouter>
    </AppProviders>,
  );
}

describe('AppRoutes · ruta /ofrendas (F7)', () => {
  it('la sección pública vive en /ofrendas y no requiere sesión', async () => {
    server.use(
      http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })),
      http.get(ofrendasUrl, () =>
        HttpResponse.json({
          lang: 'es',
          intro: { text: 'Cómo donar.' },
          items: [
            {
              id: '11111111-1111-1111-1111-111111111111',
              kind: 'iban',
              number: 'CR05015202001234567890',
              institution: 'Banco Nacional de Costa Rica',
              currency: 'CRC',
            },
          ],
        }),
      ),
    );

    renderApp('/ofrendas');

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Ofrendas y diezmos' }),
    ).toBeInTheDocument();
    expect(screen.getByText('Cómo donar.')).toBeInTheDocument();
  });

  it('sin nada publicado la URL directa redirige a la portada (Q9)', async () => {
    server.use(
      http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })),
      http.get(ofrendasUrl, () => HttpResponse.json({ lang: 'es' })),
    );

    renderApp('/ofrendas');

    // Llega a la portada y en ningún momento pinta un H1 de ofrendas.
    expect(await screen.findByRole('heading', { name: 'Lo que viene' })).toBeInTheDocument();
    expect(screen.queryByRole('heading', { level: 1, name: 'Ofrendas y diezmos' })).toBeNull();
  });
});
