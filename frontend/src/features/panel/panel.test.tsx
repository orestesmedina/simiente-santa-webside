import { render, screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { SessionUser } from '../../api/auth';
import { AppProviders } from '../../app/providers';
import { ADMIN_USERS_ROLES, DONACIONES, MINISTERIOS } from '../../lib/permissions';
import { server } from '../../test/server';
import { NO_MODULES_MESSAGE, InicioPage } from './pages/InicioPage';

const sessionUrl = `${API_BASE_URL}/api/v1/auth/session`;

function makeSession(overrides: Partial<SessionUser> = {}): SessionUser {
  return {
    id: '11111111-1111-1111-1111-111111111111',
    email: 'ana@ejemplo.com',
    firstName: 'Ana',
    lastName: 'Pérez',
    phone: '612345678',
    roleId: '22222222-2222-2222-2222-222222222222',
    roleName: 'Administración',
    permissions: [ADMIN_USERS_ROLES],
    mustChangePassword: false,
    ...overrides,
  };
}

function useSession(session: SessionUser) {
  server.use(http.get(sessionUrl, () => HttpResponse.json(session)));
}

function renderInicio() {
  return render(
    <AppProviders>
      <MemoryRouter initialEntries={['/panel']}>
        <InicioPage />
      </MemoryRouter>
    </AppProviders>,
  );
}

describe('InicioPage · Mi cuenta', () => {
  it('muestra los datos de la sesión y el enlace de cambio de contraseña', async () => {
    useSession(makeSession());

    renderInicio();

    const card = await screen.findByRole('region', { name: 'Mi cuenta' });
    expect(within(card).getByText('Ana')).toBeInTheDocument();
    expect(within(card).getByText('Pérez')).toBeInTheDocument();
    expect(within(card).getByText('ana@ejemplo.com')).toBeInTheDocument();
    expect(within(card).getByText('Administración')).toBeInTheDocument();
    expect(within(card).getByRole('link', { name: 'Cambiar contraseña' })).toHaveAttribute(
      'href',
      '/cambiar-contrasena',
    );
  });

  it('mientras carga la sesión muestra el indicador', () => {
    server.use(http.get(sessionUrl, () => new Promise<never>(() => {})));

    renderInicio();

    expect(screen.getByText('Cargando tu cuenta…')).toBeInTheDocument();
  });
});

describe('InicioPage · accesos rápidos', () => {
  it('muestra solo las secciones autorizadas por permiso (FR-016)', async () => {
    useSession(makeSession());

    renderInicio();

    const nav = await screen.findByRole('navigation', { name: 'Accesos rápidos' });
    expect(within(nav).getByRole('link', { name: 'Usuarios' })).toHaveAttribute(
      'href',
      '/panel/usuarios',
    );
    expect(within(nav).getByRole('link', { name: 'Roles' })).toHaveAttribute(
      'href',
      '/panel/roles',
    );
    expect(within(nav).getByRole('link', { name: 'Auditoría' })).toHaveAttribute(
      'href',
      '/panel/auditoria',
    );
  });

  it('una cuenta sin permisos de módulo ve el aviso y ninguna sección (Edge Case)', async () => {
    useSession(makeSession({ permissions: [] }));

    renderInicio();

    expect(await screen.findByText(NO_MODULES_MESSAGE)).toBeInTheDocument();
    expect(screen.queryByRole('navigation', { name: 'Accesos rápidos' })).toBeNull();
    expect(screen.queryByRole('link', { name: 'Usuarios' })).toBeNull();
    expect(screen.queryByRole('link', { name: 'Roles' })).toBeNull();
    expect(screen.queryByRole('link', { name: 'Auditoría' })).toBeNull();
    // El resto del panel sigue usable: "Mi cuenta" y el cambio de contraseña.
    expect(screen.getByRole('region', { name: 'Mi cuenta' })).toBeInTheDocument();
  });

  it('muestra los tres accesos de agenda solo con sus permisos (F4, ux.md §4.2)', async () => {
    useSession(makeSession({ permissions: ['eventos'] }));

    renderInicio();

    const nav = await screen.findByRole('navigation', { name: 'Accesos rápidos' });
    expect(within(nav).getByRole('link', { name: 'Gestión de eventos' })).toHaveAttribute(
      'href',
      '/panel/eventos',
    );
    expect(within(nav).queryByRole('link', { name: 'Gestión de actividades' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Categorías' })).toBeNull();
    // Ni «eventos» ni «actividades» dan acceso al catálogo (FR-012).
    expect(within(nav).queryByRole('link', { name: 'Usuarios' })).toBeNull();
  });

  it('con solo el permiso «categorias» se ve solo esa tarjeta', async () => {
    useSession(makeSession({ permissions: ['categorias'] }));

    renderInicio();

    const nav = await screen.findByRole('navigation', { name: 'Accesos rápidos' });
    expect(within(nav).getByRole('link', { name: 'Categorías' })).toHaveAttribute(
      'href',
      '/panel/categorias',
    );
    expect(within(nav).queryByRole('link', { name: 'Gestión de eventos' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Gestión de actividades' })).toBeNull();
  });

  it('sin permisos de F4 no aparece ninguna tarjeta de agenda', async () => {
    useSession(makeSession());

    renderInicio();

    const nav = await screen.findByRole('navigation', { name: 'Accesos rápidos' });
    expect(within(nav).queryByRole('link', { name: 'Gestión de eventos' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Gestión de actividades' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Categorías' })).toBeNull();
  });

  it('con solo el permiso «ministerios» se ve solo su tarjeta (F6, FR-011)', async () => {
    useSession(makeSession({ permissions: [MINISTERIOS] }));

    renderInicio();

    const nav = await screen.findByRole('navigation', { name: 'Accesos rápidos' });
    expect(within(nav).getByRole('link', { name: 'Gestión de ministerios' })).toHaveAttribute(
      'href',
      '/panel/ministerios',
    );
    // Los módulos son independientes: «ministerios» no abre los demás (Decisión 5).
    expect(within(nav).queryByRole('link', { name: 'Usuarios' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Gestión de eventos' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Portada e información general' })).toBeNull();
  });

  it('sin el permiso «ministerios» su tarjeta no aparece', async () => {
    useSession(makeSession());

    renderInicio();

    const nav = await screen.findByRole('navigation', { name: 'Accesos rápidos' });
    expect(within(nav).queryByRole('link', { name: 'Gestión de ministerios' })).toBeNull();
  });

  it('con solo el permiso «donaciones» se ve solo su tarjeta (F7, FR-011)', async () => {
    useSession(makeSession({ permissions: [DONACIONES] }));

    renderInicio();

    const nav = await screen.findByRole('navigation', { name: 'Accesos rápidos' });
    expect(within(nav).getByRole('link', { name: 'Gestión de donaciones' })).toHaveAttribute(
      'href',
      '/panel/donaciones',
    );
    // Los módulos son independientes: «donaciones» no abre los demás (Decisión 5).
    expect(within(nav).queryByRole('link', { name: 'Usuarios' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Gestión de ministerios' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Portada e información general' })).toBeNull();
  });

  it('sin el permiso «donaciones» su tarjeta no aparece', async () => {
    useSession(makeSession());

    renderInicio();

    const nav = await screen.findByRole('navigation', { name: 'Accesos rápidos' });
    expect(within(nav).queryByRole('link', { name: 'Gestión de donaciones' })).toBeNull();
  });
});
