import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import { RequirePasswordChange } from '../../app/guards';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import {
  ChangePasswordPage,
  MANDATORY_PASSWORD_MESSAGE,
  PASSWORD_CHANGED_MESSAGE,
} from './pages/ChangePasswordPage';

const sessionUrl = `${API_BASE_URL}/api/v1/auth/session`;
const passwordUrl = `${API_BASE_URL}/api/v1/auth/password`;

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

function renderPage({ mandatory = false } = {}) {
  server.use(
    http.get(sessionUrl, () =>
      HttpResponse.json({ ...sessionUser, mustChangePassword: mandatory }),
    ),
  );

  return render(
    <MemoryRouter initialEntries={['/cambiar-contrasena']}>
      <AppProviders>
        <Routes>
          <Route path="/cambiar-contrasena" element={<ChangePasswordPage />} />
          <Route element={<RequirePasswordChange />}>
            <Route path="/panel" element={<p>panel del equipo</p>} />
          </Route>
        </Routes>
        <LocationProbe />
      </AppProviders>
    </MemoryRouter>,
  );
}

async function fillForm(
  currentPassword = 'vieja',
  newPassword = 'Nueva.2026',
  confirmPassword = 'Nueva.2026',
) {
  await userEvent.type(screen.getByLabelText(/^Contraseña actual/), currentPassword);
  await userEvent.type(screen.getByLabelText(/^Contraseña nueva/), newPassword);
  await userEvent.type(screen.getByLabelText(/^Confirmar/), confirmPassword);
}

function usePasswordHandler(response: () => Response | Promise<Response>) {
  server.use(http.post(passwordUrl, response));
}

describe('ChangePasswordPage', () => {
  it('con el cambio válido avisa del éxito y sale del modo obligatorio (US7)', async () => {
    renderPage({ mandatory: true });
    usePasswordHandler(() => HttpResponse.json({ passwordChanged: true }));

    await fillForm();
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(await screen.findByText(PASSWORD_CHANGED_MESSAGE)).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Ir al panel' }));
    expect(await screen.findByText('panel del equipo')).toBeInTheDocument();
    expect(screen.getByTestId('location')).toHaveTextContent('/panel');
  });

  it('muestra el requisito incumplido que devuelve el servidor (US7 esc. 3)', async () => {
    renderPage();
    usePasswordHandler(() =>
      HttpResponse.json(
        {
          error: {
            code: 'invalid',
            message: 'La contraseña no cumple la política de seguridad',
            details: { newPassword: 'La contraseña debe incluir al menos una letra mayúscula' },
          },
        },
        { status: 400 },
      ),
    );

    await fillForm();
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(
      await screen.findByText('La contraseña debe incluir al menos una letra mayúscula'),
    ).toBeInTheDocument();
  });

  it('avisa cuando la contraseña actual no coincide (US7 esc. 2)', async () => {
    renderPage();
    usePasswordHandler(() =>
      HttpResponse.json(
        {
          error: {
            code: 'invalid',
            message: 'Tu contraseña actual no coincide',
            details: { currentPassword: 'La contraseña actual no es correcta' },
          },
        },
        { status: 400 },
      ),
    );

    await fillForm('mala');
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(
      await screen.findByText('Tu contraseña actual no coincide. Vuelve a escribirla.'),
    ).toBeInTheDocument();
  });

  it('con la sesión expirada muestra su mensaje', async () => {
    renderPage();
    usePasswordHandler(() =>
      HttpResponse.json(
        { error: { code: 'unauthenticated', message: 'Tu sesión terminó' } },
        { status: 401 },
      ),
    );

    await fillForm();
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(
      await screen.findByText(
        'Tu sesión terminó. Entra de nuevo y vuelve a hacer la acción; no se guardó a medias.',
      ),
    ).toBeInTheDocument();
  });

  it('con un 403 muestra un error del sistema sin detalles internos', async () => {
    renderPage();
    usePasswordHandler(() =>
      HttpResponse.json(
        { error: { code: 'forbidden', message: 'No tienes permiso' } },
        { status: 403 },
      ),
    );

    await fillForm();
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(
      await screen.findByText(
        'No se pudo completar la operación. Vuelve a intentarlo en unos minutos; si sigue, avísanos.',
      ),
    ).toBeInTheDocument();
  });

  it('en el modo obligatorio no muestra navegación y pide el cambio (US7 esc. 4)', async () => {
    renderPage({ mandatory: true });

    expect(await screen.findByText(MANDATORY_PASSWORD_MESSAGE)).toBeInTheDocument();
    expect(screen.queryByRole('navigation')).toBeNull();
    expect(screen.queryByRole('link', { name: 'Volver al panel' })).toBeNull();
  });

  it('marca en vivo la política y deshabilita guardar hasta cumplirla (FR-010)', async () => {
    renderPage();

    await userEvent.type(screen.getByLabelText(/^Contraseña nueva/), 'debil');

    const lista = await screen.findByRole('list', { name: 'Requisitos de la contraseña' });
    expect(lista).toHaveTextContent('Al menos 8 caracteres');
    expect(lista).toHaveTextContent('Una letra mayúscula');
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeDisabled();
  });
});
