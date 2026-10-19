import { render, screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../api/client';
import type { SessionUser } from '../api/auth';
import { MEDIOS } from '../lib/permissions';
import { server } from '../test/server';
import { AppProviders } from './providers';
import { AppRoutes } from './router';

/**
 * Rutas de F9 (T933/T936): la sección pública es `/medios` —una **ruta real**
 * sin sesión (FR-002) que, sin nada publicable, redirige en silencio a la
 * portada (Q10/P9-16/SC-016)— y el módulo del panel vive en `/panel/medios`,
 * protegido por el permiso **`medios`** (F2/FR-021, Q13). La entrada de
 * navegación la cubre `useSiteLinks` (T934).
 */

const portadaUrl = `${API_BASE_URL}/api/v1/portada`;
const mediosUrl = `${API_BASE_URL}/api/v1/medios`;
const sessionUrl = `${API_BASE_URL}/api/v1/auth/session`;
const adminUrl = `${API_BASE_URL}/api/v1/admin/medios`;

function makeSession(permissions: string[]): SessionUser {
  return {
    id: '11111111-1111-1111-1111-111111111111',
    email: 'ana@ejemplo.com',
    firstName: 'Ana',
    lastName: 'Pérez',
    phone: '612345678',
    roleId: '22222222-2222-2222-2222-222222222222',
    roleName: 'Administración',
    permissions,
    mustChangePassword: false,
  };
}

function useSession(permissions: string[]) {
  server.use(http.get(sessionUrl, () => HttpResponse.json(makeSession(permissions))));
}

function usePanelApi() {
  server.use(
    http.get(adminUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
  );
}

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

describe('AppRoutes · ruta del panel de medios (F9/T936)', () => {
  it('con el permiso «medios» la ruta del panel abre su página', async () => {
    usePanelApi();
    useSession([MEDIOS]);

    renderApp('/panel/medios');

    expect(await screen.findByRole('heading', { level: 1, name: 'Medios' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Buscar episodios nuevos' })).toBeInTheDocument();
  });

  it('sin el permiso «medios» la ruta forzada cae en «sin permiso» de F2', async () => {
    usePanelApi();
    useSession([]);

    renderApp('/panel/medios');

    expect(
      await screen.findByRole('heading', { name: 'No tienes acceso a esta sección' }),
    ).toBeInTheDocument();
  });
});
