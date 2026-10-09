import { render, screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../api/client';
import { ADMIN_USERS_ROLES, PORTADA } from '../lib/permissions';
import { server } from '../test/server';
import { AppProviders } from './providers';
import { AppRoutes } from './router';

const sessionUrl = `${API_BASE_URL}/api/v1/auth/session`;
const portadaUrl = `${API_BASE_URL}/api/v1/portada`;
const portadaAdminUrl = `${API_BASE_URL}/api/v1/admin/portada`;

const sessionAdmin = {
  id: '11111111-1111-1111-1111-111111111111',
  email: 'ana@ejemplo.com',
  firstName: 'Ana',
  lastName: 'Pérez',
  phone: '612345678',
  roleId: '22222222-2222-2222-2222-222222222222',
  roleName: 'Administración',
  permissions: [ADMIN_USERS_ROLES],
  mustChangePassword: false,
};

function renderApp(entry: string) {
  return render(
    <AppProviders>
      <MemoryRouter initialEntries={[entry]}>
        <AppRoutes />
      </MemoryRouter>
    </AppProviders>,
  );
}

describe('AppRoutes', () => {
  it('la portada pública vive en / (sin sesión)', async () => {
    server.use(
      http.get(portadaUrl, () =>
        HttpResponse.json({ lang: 'es', identity: { name: 'Iglesia Simiente Santa' } }),
      ),
    );

    renderApp('/');

    expect(
      await screen.findByRole('heading', { name: 'Iglesia Simiente Santa' }),
    ).toBeInTheDocument();
  });

  it('/health muestra «Estado del sistema» dentro del layout', () => {
    renderApp('/health');

    expect(screen.getByRole('heading', { name: 'Estado del sistema' })).toBeInTheDocument();
    expect(screen.getByRole('main')).toBeInTheDocument();
  });

  it('el catch-all muestra Página no encontrada', () => {
    renderApp('/ruta-que-no-existe');

    expect(screen.getByRole('heading', { name: 'Página no encontrada' })).toBeInTheDocument();
  });

  it('sin sesión, el panel redirige al acceso', async () => {
    server.use(
      http.get(sessionUrl, () =>
        HttpResponse.json(
          { error: { code: 'unauthenticated', message: 'Sin sesión' } },
          { status: 401 },
        ),
      ),
    );

    renderApp('/panel');

    expect(await screen.findByRole('heading', { name: 'Entrar al panel' })).toBeInTheDocument();
  });

  it('con permiso, la navegación muestra las secciones autorizadas', async () => {
    server.use(http.get(sessionUrl, () => HttpResponse.json(sessionAdmin)));

    renderApp('/panel');

    const nav = await screen.findByRole('navigation', { name: 'Navegación del panel' });
    expect(within(nav).getByRole('link', { name: 'Inicio' })).toBeInTheDocument();
    expect(within(nav).getByRole('link', { name: 'Usuarios' })).toBeInTheDocument();
    expect(within(nav).getByRole('link', { name: 'Roles' })).toBeInTheDocument();
    expect(within(nav).getByRole('link', { name: 'Auditoría' })).toBeInTheDocument();
  });

  it('sin permiso de módulo, la navegación solo muestra Inicio', async () => {
    server.use(http.get(sessionUrl, () => HttpResponse.json({ ...sessionAdmin, permissions: [] })));

    renderApp('/panel');

    const nav = await screen.findByRole('navigation', { name: 'Navegación del panel' });
    expect(within(nav).getByRole('link', { name: 'Inicio' })).toBeInTheDocument();
    expect(within(nav).queryByRole('link', { name: 'Usuarios' })).not.toBeInTheDocument();
  });

  it('con sesión pero sin permiso, /panel/usuarios muestra sin permiso', async () => {
    server.use(http.get(sessionUrl, () => HttpResponse.json({ ...sessionAdmin, permissions: [] })));

    renderApp('/panel/usuarios');

    expect(
      await screen.findByRole('heading', { name: 'No tienes acceso a esta sección' }),
    ).toBeInTheDocument();
  });

  it('sin sesión, /panel/informacion redirige al acceso', async () => {
    server.use(
      http.get(sessionUrl, () =>
        HttpResponse.json(
          { error: { code: 'unauthenticated', message: 'Sin sesión' } },
          { status: 401 },
        ),
      ),
    );

    renderApp('/panel/informacion');

    expect(await screen.findByRole('heading', { name: 'Entrar al panel' })).toBeInTheDocument();
  });

  it('con el permiso portada, /panel/informacion muestra el módulo', async () => {
    server.use(
      http.get(sessionUrl, () => HttpResponse.json({ ...sessionAdmin, permissions: [PORTADA] })),
      http.get(portadaAdminUrl, () =>
        HttpResponse.json({
          identity: null,
          about: null,
          contact: null,
          schedule: { items: [] },
          whatsapp: { items: [] },
          socials: { items: [] },
        }),
      ),
    );

    renderApp('/panel/informacion');

    expect(
      await screen.findByRole('heading', { name: 'Portada e información general' }),
    ).toBeInTheDocument();
  });

  it('con sesión pero sin permiso portada, /panel/informacion muestra sin permiso', async () => {
    server.use(http.get(sessionUrl, () => HttpResponse.json({ ...sessionAdmin, permissions: [] })));

    renderApp('/panel/informacion');

    expect(
      await screen.findByRole('heading', { name: 'No tienes acceso a esta sección' }),
    ).toBeInTheDocument();
  });

  it('con el permiso portada, el menú muestra la entrada del módulo y no las de F2', async () => {
    server.use(
      http.get(sessionUrl, () => HttpResponse.json({ ...sessionAdmin, permissions: [PORTADA] })),
    );

    renderApp('/panel');

    const nav = await screen.findByRole('navigation', { name: 'Navegación del panel' });
    expect(
      await within(nav).findByRole('link', { name: 'Portada e información general' }),
    ).toBeInTheDocument();
    expect(within(nav).queryByRole('link', { name: 'Usuarios' })).not.toBeInTheDocument();
  });

  it('mustChangePassword obliga al cambio antes del panel', async () => {
    server.use(
      http.get(sessionUrl, () => HttpResponse.json({ ...sessionAdmin, mustChangePassword: true })),
    );

    renderApp('/panel');

    expect(await screen.findByRole('heading', { name: 'Cambiar contraseña' })).toBeInTheDocument();
  });
});
