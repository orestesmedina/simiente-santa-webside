import { render, screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../api/client';
import type { SessionUser } from '../api/auth';
import { MINISTERIOS } from '../lib/permissions';
import { server } from '../test/server';
import { AppProviders } from './providers';
import { AppRoutes } from './router';

/**
 * Rutas de F6 (T627/T630): la sección pública vive en `/ministerios` —sin sesión—
 * y el módulo del panel en `/panel/ministerios`, protegido por el permiso
 * **`ministerios`** (F2/FR-011). Sin él, la URL forzada cae en «sin permiso».
 */

const sessionUrl = `${API_BASE_URL}/api/v1/auth/session`;
const ministriesUrl = `${API_BASE_URL}/api/v1/admin/ministerios`;

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
    http.get(ministriesUrl, () =>
      HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 }),
    ),
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

describe('AppRoutes · rutas de ministerios (F6)', () => {
  it('la sección pública vive en /ministerios y no requiere sesión', async () => {
    server.use(
      http.get(`${API_BASE_URL}/api/v1/portada`, () => HttpResponse.json({ lang: 'es' })),
      http.get(`${API_BASE_URL}/api/v1/ministerios`, () =>
        HttpResponse.json({
          lang: 'es',
          items: [{ id: 'min-1', name: 'Alabanza', description: 'Adoración en los cultos.' }],
        }),
      ),
    );

    renderApp('/ministerios');

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Ministerios' }),
    ).toBeInTheDocument();
  });

  it('con el permiso «ministerios» la ruta del panel abre su página', async () => {
    usePanelApi();
    useSession([MINISTERIOS]);

    renderApp('/panel/ministerios');

    expect(
      await screen.findByRole('heading', { name: 'Gestión de ministerios' }),
    ).toBeInTheDocument();
  });

  it('sin el permiso «ministerios» la ruta forzada cae en «sin permiso» de F2', async () => {
    usePanelApi();
    useSession([]);

    renderApp('/panel/ministerios');

    expect(
      await screen.findByRole('heading', { name: 'No tienes acceso a esta sección' }),
    ).toBeInTheDocument();
  });
});
