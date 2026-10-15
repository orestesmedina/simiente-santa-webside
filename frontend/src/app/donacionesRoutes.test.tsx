import { render, screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../api/client';
import type { SessionUser } from '../api/auth';
import { DONACIONES } from '../lib/permissions';
import { server } from '../test/server';
import { AppProviders } from './providers';
import { AppRoutes } from './router';

/**
 * Ruta del panel de F7 (T729): el módulo vive en `/panel/donaciones`, protegido
 * por el permiso **`donaciones`** (F2/FR-011, Q8). Sin él, la URL forzada cae en
 * «sin permiso».
 */

const sessionUrl = `${API_BASE_URL}/api/v1/auth/session`;
const adminUrl = `${API_BASE_URL}/api/v1/admin/donaciones`;

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
    http.get(`${adminUrl}/textos`, () => HttpResponse.json({ intro: null, usage: null })),
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

describe('AppRoutes · ruta del panel de donaciones (F7/T729)', () => {
  it('con el permiso «donaciones» la ruta del panel abre su página', async () => {
    usePanelApi();
    useSession([DONACIONES]);

    renderApp('/panel/donaciones');

    expect(
      await screen.findByRole('heading', { name: 'Gestión de donaciones' }),
    ).toBeInTheDocument();
  });

  it('sin el permiso «donaciones» la ruta forzada cae en «sin permiso» de F2', async () => {
    usePanelApi();
    useSession([]);

    renderApp('/panel/donaciones');

    expect(
      await screen.findByRole('heading', { name: 'No tienes acceso a esta sección' }),
    ).toBeInTheDocument();
  });
});
