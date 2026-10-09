import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { LoginPage } from './pages/LoginPage';

const loginUrl = `${API_BASE_URL}/api/v1/auth/login`;

const sessionUser = {
  id: '11111111-1111-1111-1111-111111111111',
  email: 'ana@ejemplo.com',
  firstName: 'Ana',
  lastName: 'Pérez',
  phone: '612345678',
  roleId: '22222222-2222-2222-2222-222222222222',
  roleName: 'Administración',
  permissions: ['admin_usuarios_roles'],
  mustChangePassword: false,
};

function LocationProbe() {
  const location = useLocation();
  return <output data-testid="location">{`${location.pathname}${location.search}`}</output>;
}

function renderLogin(entry = '/login') {
  return render(
    <MemoryRouter initialEntries={[entry]}>
      <AppProviders>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/panel" element={<p>panel del equipo</p>} />
          <Route path="/panel/usuarios" element={<p>usuarios del equipo</p>} />
        </Routes>
        <LocationProbe />
      </AppProviders>
    </MemoryRouter>,
  );
}

async function submitLogin(email: string, password: string) {
  await userEvent.type(screen.getByLabelText(/^Correo/), email);
  await userEvent.type(screen.getByLabelText(/^Contraseña/), password);
  await userEvent.click(screen.getByRole('button', { name: 'Entrar' }));
}

describe('LoginPage', () => {
  it('con credenciales válidas entra al panel (SC-002)', async () => {
    server.use(http.post(loginUrl, () => HttpResponse.json(sessionUser)));

    renderLogin();
    await submitLogin('ana@ejemplo.com', 'Secreta.123');

    expect(await screen.findByText('panel del equipo')).toBeInTheDocument();
    expect(screen.getByTestId('location')).toHaveTextContent('/panel');
  });

  it('vuelve al destino indicado tras entrar', async () => {
    server.use(http.post(loginUrl, () => HttpResponse.json(sessionUser)));

    renderLogin('/login?destino=%2Fpanel%2Fusuarios');
    await submitLogin('ana@ejemplo.com', 'Secreta.123');

    expect(await screen.findByText('usuarios del equipo')).toBeInTheDocument();
    expect(screen.getByTestId('location')).toHaveTextContent('/panel/usuarios');
  });

  // SC-008: el mensaje debe ser idéntico exista o no la cuenta.
  it('con un correo inexistente muestra el error genérico', async () => {
    server.use(
      http.post(loginUrl, () =>
        HttpResponse.json(
          { error: { code: 'unauthenticated', message: 'Correo o contraseña incorrectos' } },
          { status: 401 },
        ),
      ),
    );

    renderLogin();
    await submitLogin('noexiste@ejemplo.com', 'loquesea');

    expect(await screen.findByRole('alert')).toHaveTextContent('Correo o contraseña incorrectos.');
  });

  it('con una contraseña errónea muestra el mismo error genérico (SC-008)', async () => {
    server.use(
      http.post(loginUrl, () =>
        HttpResponse.json(
          { error: { code: 'unauthenticated', message: 'Correo o contraseña incorrectos' } },
          { status: 401 },
        ),
      ),
    );

    renderLogin();
    await submitLogin('ana@ejemplo.com', 'equivocada');

    expect(await screen.findByRole('alert')).toHaveTextContent('Correo o contraseña incorrectos.');
  });

  it('con la cuenta desactivada explica que el acceso está desactivado (US1 esc. 3)', async () => {
    server.use(
      http.post(loginUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'forbidden',
              message: 'Ese acceso está desactivado',
              details: { reason: 'access_disabled' },
            },
          },
          { status: 403 },
        ),
      ),
    );

    renderLogin();
    await submitLogin('ana@ejemplo.com', 'Secreta.123');

    expect(await screen.findByRole('alert')).toHaveTextContent(/Ese acceso está desactivado/);
  });

  it('con el bloqueo temporal muestra los minutos y deshabilita el formulario (FR-006)', async () => {
    server.use(
      http.post(loginUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'rate_limited',
              message: 'Demasiados intentos',
              details: { retryAfterSeconds: 900 },
            },
          },
          { status: 429 },
        ),
      ),
    );

    renderLogin();
    await submitLogin('ana@ejemplo.com', 'Secreta.123');

    const notice = await screen.findByRole('alert');
    expect(notice).toHaveTextContent(/se han superado los intentos permitidos/);
    expect(notice).toHaveTextContent(/quedan 15 minutos/);
    expect(screen.getByLabelText(/^Correo/)).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Entrar' })).toBeDisabled();
  });

  it('avisa cuando la sesión terminó (motivo=expirada, FR-005)', () => {
    renderLogin('/login?motivo=expirada');

    expect(
      screen.getByText('Tu sesión terminó. Vuelve a entrar para continuar.'),
    ).toBeInTheDocument();
  });

  it('valida los campos obligatorios con mensajes junto al campo (ux §4.a)', async () => {
    renderLogin();
    await userEvent.click(screen.getByRole('button', { name: 'Entrar' }));

    expect(await screen.findByText('Escribe tu correo.')).toBeInTheDocument();
    expect(screen.getByText('Escribe tu contraseña.')).toBeInTheDocument();
  });

  it('es accesible y no ofrece registro ni recuperación (Out of Scope)', () => {
    renderLogin();

    expect(screen.getByLabelText(/^Correo/)).toHaveAttribute('autocomplete', 'email');
    expect(screen.getByLabelText(/^Contraseña/)).toHaveAttribute(
      'autocomplete',
      'current-password',
    );
    expect(screen.getByRole('button', { name: 'Mostrar contraseña' })).toBeInTheDocument();
    expect(screen.queryByText(/regístrate|olvidé mi contraseña|recuperar contraseña/i)).toBeNull();
  });
});
