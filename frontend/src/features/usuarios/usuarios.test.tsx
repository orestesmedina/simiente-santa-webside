import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { UserItem } from '../../api/usuarios';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { UsersPage } from './pages/UsersPage';

const usersUrl = `${API_BASE_URL}/api/v1/admin/usuarios`;

function makeUser(overrides: Partial<UserItem> = {}): UserItem {
  return {
    id: '11111111-1111-1111-1111-111111111111',
    email: 'ana@ejemplo.com',
    firstName: 'Ana',
    lastName: 'Pérez',
    phone: '612345678',
    roleId: '22222222-2222-2222-2222-222222222222',
    roleName: 'Administración',
    isActive: true,
    mustChangePassword: false,
    lastLoginAt: '2026-10-03T17:42:00Z',
    lastLoginIp: '189.2.4.15',
    createdAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function renderUsers() {
  return render(
    <AppProviders>
      <UsersPage />
    </AppProviders>,
  );
}

function useList(items: UserItem[], total = items.length) {
  server.use(http.get(usersUrl, () => HttpResponse.json({ items, total, limit: 20, offset: 0 })));
}

describe('UsersPage · listado', () => {
  it('muestra estado, correo, rol y último acceso en cada fila (FR-019/FR-021)', async () => {
    const ana = makeUser();
    const beto = makeUser({
      id: '33333333-3333-3333-3333-333333333333',
      email: 'beto@ejemplo.com',
      firstName: 'Beto',
      lastName: 'López',
      roleName: 'Miembro',
      isActive: false,
      lastLoginAt: null,
      lastLoginIp: null,
    });
    useList([ana, beto], 2);

    renderUsers();
    const table = await screen.findByRole('table');

    expect(within(table).getByText('ana@ejemplo.com')).toBeInTheDocument();
    expect(within(table).getByText('Administración')).toBeInTheDocument();
    expect(within(table).getByText('Activo')).toBeInTheDocument();
    expect(within(table).getByText(/189\.2\.4\.15/)).toBeInTheDocument();
    expect(within(table).getByText('Inactivo')).toBeInTheDocument();
    expect(within(table).getByText('Miembro')).toBeInTheDocument();
  });

  it('la cuenta sin accesos indica que nunca ha entrado y no inventa fechas (FR-021)', async () => {
    const beto = makeUser({
      email: 'beto@ejemplo.com',
      isActive: false,
      lastLoginAt: null,
      lastLoginIp: null,
    });
    useList([beto], 1);

    renderUsers();
    const table = await screen.findByRole('table');

    expect(within(table).getByText('Nunca ha entrado al panel.')).toBeInTheDocument();
    expect(within(table).queryByText(/\d{2}\/\d{2}\/\d{4}/)).toBeNull();
  });

  it('pagina con limit/offset y conserva el estado al cambiar de página', async () => {
    const ana = makeUser();
    const beto = makeUser({
      id: '33333333-3333-3333-3333-333333333333',
      email: 'beto@ejemplo.com',
      firstName: 'Beto',
      lastName: 'López',
    });
    const urls: string[] = [];
    server.use(
      http.get(usersUrl, ({ request }) => {
        urls.push(request.url);
        const offset = Number(new URL(request.url).searchParams.get('offset') ?? '0');
        return HttpResponse.json({
          items: offset === 0 ? [ana] : [beto],
          total: 45,
          limit: 20,
          offset,
        });
      }),
    );

    renderUsers();
    const firstTable = await screen.findByRole('table');
    expect(within(firstTable).getByText('ana@ejemplo.com')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Anterior' })).toBeDisabled();

    await userEvent.click(screen.getByRole('button', { name: 'Siguiente' }));

    await waitFor(() =>
      expect(within(screen.getByRole('table')).getByText('beto@ejemplo.com')).toBeInTheDocument(),
    );
    expect(urls.some((url) => url.includes('limit=20') && url.includes('offset=20'))).toBe(true);
    expect(screen.getByRole('button', { name: 'Anterior' })).toBeEnabled();
  });

  it('no ofrece buscador de texto libre en el MVP (F-04)', async () => {
    useList([makeUser()], 1);

    renderUsers();
    await screen.findByRole('table');

    expect(screen.queryByRole('searchbox')).toBeNull();
    expect(screen.queryAllByRole('textbox')).toHaveLength(0);
    expect(screen.queryByPlaceholderText(/buscar/i)).toBeNull();
  });

  it('mientras carga muestra el indicador (ux §4.b)', async () => {
    server.use(
      http.get(usersUrl, async () => {
        await delay(50);
        return HttpResponse.json({ items: [makeUser()], total: 1, limit: 20, offset: 0 });
      }),
    );

    renderUsers();

    expect(screen.getByText('Cargando usuarios…')).toBeInTheDocument();
    expect(await screen.findByRole('table')).toBeInTheDocument();
  });

  it('sin cuentas muestra el estado vacío con la acción de crear', async () => {
    useList([], 0);

    renderUsers();

    expect(
      await screen.findByText('Aún no hay cuentas. Crea la primera para tu equipo.'),
    ).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: 'Crear usuario' }).length).toBeGreaterThan(0);
  });

  it('con error muestra el mensaje y permite reintentar', async () => {
    let calls = 0;
    server.use(
      http.get(usersUrl, () => {
        calls += 1;
        if (calls === 1) {
          return HttpResponse.json(
            { error: { code: 'internal', message: 'Error interno del servidor' } },
            { status: 500 },
          );
        }
        return HttpResponse.json({ items: [makeUser()], total: 1, limit: 20, offset: 0 });
      }),
    );

    renderUsers();

    expect(await screen.findByText('No se pudo cargar el listado.')).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Reintentar' }));
    expect(await screen.findByRole('table')).toBeInTheDocument();
  });
});
