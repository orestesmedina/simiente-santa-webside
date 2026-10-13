import { render, screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../api/client';
import { ADMIN_USERS_ROLES, PORTADA } from '../lib/permissions';
import { server } from '../test/server';
import { AppProviders } from './providers';
import { AppLayout, PanelLayout, PanelSurfaceLayout } from './layout';

describe('AppLayout', () => {
  it('expone un layout semántico accesible con main y navegación', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route element={<AppLayout />}>
            <Route index element={<p>contenido de la ruta</p>} />
          </Route>
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole('main')).toBeInTheDocument();
    expect(screen.getByRole('navigation', { name: 'Navegación principal' })).toBeInTheDocument();
    expect(screen.getByText('contenido de la ruta')).toBeInTheDocument();
  });
});

describe('PanelSurfaceLayout', () => {
  it('pinta la superficie clara del panel y su contenido (bug del fondo negro)', () => {
    const { container } = render(
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route element={<PanelSurfaceLayout />}>
            <Route index element={<p>contenido de la ruta</p>} />
          </Route>
        </Routes>
      </MemoryRouter>,
    );

    // El panel es siempre claro (FR-016): la superficie no depende del tema del
    // sitio público, de modo que el lienzo oscuro del navegador nunca se ve.
    const shell = container.querySelector('#root > div') ?? container.firstElementChild;
    expect(shell).not.toBeNull();
    expect((shell as HTMLElement).className).toContain('bg-white');
    expect((shell as HTMLElement).className).toContain('text-panel-ink');
    expect(screen.getByText('contenido de la ruta')).toBeInTheDocument();
  });

  it('acepta contenido explícito además del Outlet (404 fuera del panel)', () => {
    const { container } = render(
      <PanelSurfaceLayout>
        <main>página no encontrada</main>
      </PanelSurfaceLayout>,
    );

    expect(screen.getByText('página no encontrada')).toBeInTheDocument();
    const shell = container.firstElementChild as HTMLElement;
    expect(shell.className).toContain('bg-white');
  });
});

describe('PanelLayout', () => {
  const session = {
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

  it('muestra la navegación autorizada, el nombre, el rol y Salir', async () => {
    server.use(http.get(`${API_BASE_URL}/api/v1/auth/session`, () => HttpResponse.json(session)));

    render(
      <AppProviders>
        <MemoryRouter initialEntries={['/panel']}>
          <Routes>
            <Route path="/panel" element={<PanelLayout />}>
              <Route index element={<p>contenido</p>} />
            </Route>
          </Routes>
        </MemoryRouter>
      </AppProviders>,
    );

    expect(
      await screen.findByRole('navigation', { name: 'Navegación del panel' }),
    ).toBeInTheDocument();
    expect(await screen.findByText('Ana Pérez · Administración')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Salir' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Menú' })).toHaveAttribute('aria-expanded', 'false');
  });

  it('muestra la entrada de la portada solo con el permiso portada', async () => {
    server.use(
      http.get(`${API_BASE_URL}/api/v1/auth/session`, () =>
        HttpResponse.json({ ...session, permissions: [PORTADA] }),
      ),
    );

    render(
      <AppProviders>
        <MemoryRouter initialEntries={['/panel']}>
          <Routes>
            <Route path="/panel" element={<PanelLayout />}>
              <Route index element={<p>contenido</p>} />
            </Route>
          </Routes>
        </MemoryRouter>
      </AppProviders>,
    );

    const nav = await screen.findByRole('navigation', { name: 'Navegación del panel' });
    expect(
      await within(nav).findByRole('link', { name: 'Portada e información general' }),
    ).toBeInTheDocument();
    expect(within(nav).queryByRole('link', { name: 'Usuarios' })).not.toBeInTheDocument();
  });
});
