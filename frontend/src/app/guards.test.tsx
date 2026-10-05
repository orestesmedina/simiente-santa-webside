import { render, screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../api/client';
import { ADMIN_USERS_ROLES } from '../lib/permissions';
import { server } from '../test/server';
import { AppProviders } from './providers';
import { RequireAuth, RequirePasswordChange, RequirePermission } from './guards';

const sessionUrl = `${API_BASE_URL}/api/v1/auth/session`;

function makeSession(overrides: Record<string, unknown> = {}) {
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

function LocationProbe() {
  const location = useLocation();
  return <output data-testid="location">{`${location.pathname}${location.search}`}</output>;
}

function renderGuards(entry: string) {
  return render(
    <AppProviders>
      <MemoryRouter initialEntries={[entry]}>
        <Routes>
          <Route element={<RequireAuth />}>
            <Route path="/protegida" element={<p>contenido protegido</p>} />
            <Route path="/cambiar-contrasena" element={<p>cambio de contraseña</p>} />
            <Route element={<RequirePasswordChange />}>
              <Route path="/panel" element={<p>contenido del panel</p>} />
            </Route>
            <Route element={<RequirePermission code={ADMIN_USERS_ROLES} />}>
              <Route path="/seccion" element={<p>sección de administración</p>} />
            </Route>
          </Route>
          <Route path="/login" element={<p>pantalla de acceso</p>} />
          <Route path="/sin-permiso" element={<p>pantalla sin permiso</p>} />
        </Routes>
        <LocationProbe />
      </MemoryRouter>
    </AppProviders>,
  );
}

describe('RequireAuth', () => {
  it('sin sesión (401) redirige a /login conservando el destino', async () => {
    server.use(
      http.get(sessionUrl, () =>
        HttpResponse.json(
          { error: { code: 'unauthenticated', message: 'Sin sesión' } },
          { status: 401 },
        ),
      ),
    );

    renderGuards('/protegida');

    expect(await screen.findByText('pantalla de acceso')).toBeInTheDocument();
    expect(screen.queryByText('contenido protegido')).not.toBeInTheDocument();
    const location = screen.getByTestId('location').textContent ?? '';
    expect(location).toContain('/login');
    expect(location).toContain('motivo=sin-sesion');
    expect(location).toContain('destino=%2Fprotegida');
  });

  it('con sesión permite el contenido', async () => {
    server.use(http.get(sessionUrl, () => HttpResponse.json(makeSession())));

    renderGuards('/protegida');

    expect(await screen.findByText('contenido protegido')).toBeInTheDocument();
  });

  it('un fallo que no es 401 muestra el aviso con Reintentar', async () => {
    server.use(
      http.get(sessionUrl, () =>
        HttpResponse.json(
          { error: { code: 'internal', message: 'Error interno del servidor' } },
          { status: 500 },
        ),
      ),
    );

    renderGuards('/protegida');

    expect(await screen.findByText(/No se pudo verificar tu acceso/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Reintentar' })).toBeInTheDocument();
  });
});

describe('RequirePasswordChange', () => {
  it('con mustChangePassword redirige al cambio obligatorio', async () => {
    server.use(
      http.get(sessionUrl, () => HttpResponse.json(makeSession({ mustChangePassword: true }))),
    );

    renderGuards('/panel');

    expect(await screen.findByText('cambio de contraseña')).toBeInTheDocument();
    expect(screen.queryByText('contenido del panel')).not.toBeInTheDocument();
  });

  it('sin mustChangePassword permite el panel', async () => {
    server.use(http.get(sessionUrl, () => HttpResponse.json(makeSession())));

    renderGuards('/panel');

    expect(await screen.findByText('contenido del panel')).toBeInTheDocument();
  });
});

describe('RequirePermission', () => {
  it('sin el permiso redirige a /sin-permiso', async () => {
    server.use(http.get(sessionUrl, () => HttpResponse.json(makeSession({ permissions: [] }))));

    renderGuards('/seccion');

    expect(await screen.findByText('pantalla sin permiso')).toBeInTheDocument();
    expect(screen.queryByText('sección de administración')).not.toBeInTheDocument();
  });

  it('con el permiso permite la sección', async () => {
    server.use(http.get(sessionUrl, () => HttpResponse.json(makeSession())));

    renderGuards('/seccion');

    expect(await screen.findByText('sección de administración')).toBeInTheDocument();
  });
});
