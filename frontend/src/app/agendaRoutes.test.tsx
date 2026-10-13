import { render, screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../api/client';
import type { SessionUser } from '../api/auth';
import { ACTIVIDADES, ADMIN_USERS_ROLES, CATEGORIAS, EVENTOS } from '../lib/permissions';
import { server } from '../test/server';
import { AppProviders } from './providers';
import { AppRoutes } from './router';

const sessionUrl = `${API_BASE_URL}/api/v1/auth/session`;
const eventsUrl = `${API_BASE_URL}/api/v1/admin/eventos`;
const activitiesUrl = `${API_BASE_URL}/api/v1/admin/actividades`;
const categoriesUrl = `${API_BASE_URL}/api/v1/admin/categorias`;
const portadaUrl = `${API_BASE_URL}/api/v1/portada`;

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
    http.get(eventsUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
    http.get(activitiesUrl, () =>
      HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 }),
    ),
    http.get(categoriesUrl, () => HttpResponse.json({ items: [] })),
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

describe('AppRoutes · rutas de agenda (F4 T430)', () => {
  it('la sección pública vive en /eventos y no requiere sesión', async () => {
    server.use(
      http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })),
      http.get(`${API_BASE_URL}/api/v1/proximamente`, () => HttpResponse.json({ lang: 'es' })),
    );

    renderApp('/eventos');

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Eventos y actividades' }),
    ).toBeInTheDocument();
  });

  it('con el permiso correspondiente, cada módulo del panel abre su página', async () => {
    usePanelApi();
    useSession([EVENTOS]);

    renderApp('/panel/eventos');
    expect(await screen.findByRole('heading', { name: 'Gestión de eventos' })).toBeInTheDocument();
  });

  it('sin el permiso del módulo, la ruta forzada cae en «sin permiso» de F2', async () => {
    usePanelApi();
    useSession([EVENTOS]);

    renderApp('/panel/actividades');

    expect(
      await screen.findByRole('heading', { name: 'No tienes acceso a esta sección' }),
    ).toBeInTheDocument();
  });

  it('el catálogo de categorías exige su permiso propio, no «eventos» (FR-012)', async () => {
    usePanelApi();
    useSession([EVENTOS, ACTIVIDADES]);

    renderApp('/panel/categorias');

    expect(
      await screen.findByRole('heading', { name: 'No tienes acceso a esta sección' }),
    ).toBeInTheDocument();
  });

  it('con el permiso «categorias» el catálogo abre y el resto de módulos no', async () => {
    usePanelApi();
    useSession([CATEGORIAS]);

    renderApp('/panel/categorias');
    expect(
      await screen.findByRole('heading', { name: 'Categorías de agenda' }),
    ).toBeInTheDocument();
  });

  it('el menú del panel filtra cada entrada por su permiso (FR-012)', async () => {
    usePanelApi();
    useSession([ACTIVIDADES]);

    renderApp('/panel');

    const nav = await screen.findByRole('navigation', { name: 'Navegación del panel' });
    expect(within(nav).getByRole('link', { name: 'Gestión de actividades' })).toBeInTheDocument();
    expect(within(nav).queryByRole('link', { name: 'Gestión de eventos' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Categorías' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Usuarios' })).toBeNull();
  });

  it('una cuenta sin permisos de F4 no ve ninguna entrada de agenda', async () => {
    usePanelApi();
    useSession([ADMIN_USERS_ROLES]);

    renderApp('/panel');

    const nav = await screen.findByRole('navigation', { name: 'Navegación del panel' });
    expect(within(nav).queryByRole('link', { name: 'Gestión de eventos' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Gestión de actividades' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Categorías' })).toBeNull();
  });
});
