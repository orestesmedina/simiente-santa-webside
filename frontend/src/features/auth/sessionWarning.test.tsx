import { act, fireEvent, render, screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { SessionWarning } from './components/SessionWarning';

const sessionUrl = `${API_BASE_URL}/api/v1/auth/session`;

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
  return <p data-testid="location">{`${location.pathname}${location.search}`}</p>;
}

interface WarningOptions {
  inactivityMs: number;
  warningMs: number;
  absoluteMs: number;
  tickMs: number;
}

function renderWarning(props: WarningOptions) {
  server.use(http.get(sessionUrl, () => HttpResponse.json(sessionUser)));

  return render(
    <MemoryRouter initialEntries={['/panel']}>
      <AppProviders>
        <Routes>
          <Route path="/panel" element={<SessionWarning {...props} />} />
          <Route path="/login" element={<p>pantalla de acceso</p>} />
        </Routes>
        <LocationProbe />
      </AppProviders>
    </MemoryRouter>,
  );
}

const fastTicks: WarningOptions = {
  inactivityMs: 3000,
  warningMs: 1000,
  absoluteMs: 600_000,
  tickMs: 100,
};

describe('SessionWarning', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('avisa 2 minutos antes del cierre y "Sigo aquí" lo descarta (FR-005)', async () => {
    renderWarning(fastTicks);

    await act(async () => {
      vi.advanceTimersByTime(2200);
    });

    expect(screen.getByRole('alertdialog', { name: '¿Sigues ahí?' })).toBeInTheDocument();
    expect(screen.getByRole('status')).toHaveTextContent(/Tu sesión se cerrará/);

    fireEvent.click(screen.getByRole('button', { name: 'Sigo aquí' }));

    await act(async () => {
      vi.advanceTimersByTime(100);
    });
    expect(screen.queryByRole('alertdialog')).toBeNull();
  });

  it('al expirar por inactividad redirige a /login con su aviso (FR-005)', async () => {
    renderWarning({ ...fastTicks, inactivityMs: 2000, warningMs: 500 });

    await act(async () => {
      vi.advanceTimersByTime(2200);
    });

    expect(screen.getByTestId('location')).toHaveTextContent('/login?motivo=expirada');
  });
});
