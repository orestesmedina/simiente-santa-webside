import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { UserItem } from '../../api/usuarios';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { LAST_ADMIN_ERROR } from './messages';
import { UsersPage } from './pages/UsersPage';

const usersUrl = `${API_BASE_URL}/api/v1/admin/usuarios`;
const rolesUrl = `${API_BASE_URL}/api/v1/admin/roles`;

const role = {
  id: '22222222-2222-2222-2222-222222222222',
  name: 'Administración',
  permissions: ['admin_usuarios_roles'],
  userCount: 1,
  createdAt: '2026-10-01T10:00:00Z',
};

const anaId = '11111111-1111-1111-1111-111111111111';
const anaUrl = `${usersUrl}/${anaId}`;
const anaPasswordUrl = `${anaUrl}/password`;

function makeUser(overrides: Partial<UserItem> = {}): UserItem {
  return {
    id: anaId,
    email: 'ana@ejemplo.com',
    firstName: 'Ana',
    lastName: 'Pérez',
    phone: '612345678',
    roleId: role.id,
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

function useRoles() {
  server.use(
    http.get(rolesUrl, () => HttpResponse.json({ items: [role], total: 1, limit: 100, offset: 0 })),
  );
}

async function openCreate() {
  await userEvent.click(screen.getByRole('button', { name: 'Crear usuario' }));
  return screen.findByRole('dialog', { name: 'Crear usuario' });
}

async function openEdit(email = 'ana@ejemplo.com') {
  const table = await screen.findByRole('table');
  const cell = within(table).getByText(email);
  const row = cell.closest('tr');
  if (!row) {
    throw new Error('No se encontró la fila de la cuenta');
  }
  await userEvent.click(within(row).getByRole('button', { name: 'Editar' }));
  return screen.findByRole('dialog', { name: 'Editar cuenta' });
}

async function fillCreateDialog(dialog: HTMLElement) {
  await userEvent.type(within(dialog).getByLabelText(/^Nombre/), 'Nuevo');
  await userEvent.type(within(dialog).getByLabelText(/^Apellidos/), 'Usuario');
  await userEvent.type(within(dialog).getByLabelText(/^Correo/), 'nuevo@ejemplo.com');
  await userEvent.type(within(dialog).getByLabelText(/^Teléfono/), '612345678');
  await userEvent.selectOptions(within(dialog).getByLabelText(/^Rol/), role.id);
  await userEvent.type(within(dialog).getByLabelText(/^Contraseña inicial/), 'Nueva.2026');
}

describe('UserForm · crear', () => {
  it('envía la contraseña inicial, avisa "Cuenta creada." y refresca el listado', async () => {
    const created = makeUser({
      email: 'nuevo@ejemplo.com',
      firstName: 'Nuevo',
      lastName: 'Usuario',
      lastLoginAt: null,
      lastLoginIp: null,
    });
    let listCalls = 0;
    let postBody: unknown;
    server.use(
      http.get(usersUrl, () => {
        listCalls += 1;
        return HttpResponse.json({ items: [makeUser()], total: 1, limit: 20, offset: 0 });
      }),
      http.get(rolesUrl, () =>
        HttpResponse.json({ items: [role], total: 1, limit: 100, offset: 0 }),
      ),
      http.post(usersUrl, async ({ request }) => {
        postBody = await request.json();
        return HttpResponse.json(created, { status: 201 });
      }),
    );

    renderUsers();
    await screen.findByRole('table');
    const dialog = await openCreate();
    await fillCreateDialog(dialog);
    await userEvent.click(within(dialog).getByRole('button', { name: 'Crear cuenta' }));

    expect(await screen.findByText('Cuenta creada.')).toBeInTheDocument();
    expect(postBody).toEqual({
      firstName: 'Nuevo',
      lastName: 'Usuario',
      email: 'nuevo@ejemplo.com',
      phone: '612345678',
      roleId: role.id,
      password: 'Nueva.2026',
    });
    await waitFor(() => expect(listCalls).toBeGreaterThanOrEqual(2));
  });

  it('con correo duplicado muestra el error en el campo y no crea nada (SC-011)', async () => {
    let postCount = 0;
    server.use(
      http.get(usersUrl, () =>
        HttpResponse.json({ items: [makeUser()], total: 1, limit: 20, offset: 0 }),
      ),
      http.get(rolesUrl, () =>
        HttpResponse.json({ items: [role], total: 1, limit: 100, offset: 0 }),
      ),
      http.post(usersUrl, () => {
        postCount += 1;
        return HttpResponse.json(
          { error: { code: 'conflict', message: 'Ese correo ya está en uso' } },
          { status: 409 },
        );
      }),
    );

    renderUsers();
    await screen.findByRole('table');
    const dialog = await openCreate();
    await fillCreateDialog(dialog);
    await userEvent.click(within(dialog).getByRole('button', { name: 'Crear cuenta' }));

    expect(await screen.findByText('Ese correo ya está en uso')).toBeInTheDocument();
    expect(postCount).toBe(1);
    expect(screen.queryByText('Cuenta creada.')).toBeNull();
  });

  it('muestra details.phone del servidor junto al teléfono', async () => {
    server.use(
      http.get(usersUrl, () =>
        HttpResponse.json({ items: [makeUser()], total: 1, limit: 20, offset: 0 }),
      ),
      http.get(rolesUrl, () =>
        HttpResponse.json({ items: [role], total: 1, limit: 100, offset: 0 }),
      ),
      http.post(usersUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Revisa los datos',
              details: { phone: 'El teléfono no es válido' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    renderUsers();
    await screen.findByRole('table');
    const dialog = await openCreate();
    await fillCreateDialog(dialog);
    await userEvent.click(within(dialog).getByRole('button', { name: 'Crear cuenta' }));

    expect(await screen.findByText('El teléfono no es válido')).toBeInTheDocument();
  });

  it('muestra details.roleId del servidor junto al rol (US3 esc. 5)', async () => {
    server.use(
      http.get(usersUrl, () =>
        HttpResponse.json({ items: [makeUser()], total: 1, limit: 20, offset: 0 }),
      ),
      http.get(rolesUrl, () =>
        HttpResponse.json({ items: [role], total: 1, limit: 100, offset: 0 }),
      ),
      http.post(usersUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Revisa los datos',
              details: { roleId: 'El rol seleccionado no existe' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    renderUsers();
    await screen.findByRole('table');
    const dialog = await openCreate();
    await fillCreateDialog(dialog);
    await userEvent.click(within(dialog).getByRole('button', { name: 'Crear cuenta' }));

    expect(await screen.findByText('El rol seleccionado no existe')).toBeInTheDocument();
  });
});

describe('UserForm · editar', () => {
  it('precarga la ficha, muestra el último acceso y no muestra contraseña (FR-021/FR-026)', async () => {
    useList([makeUser()], 1);
    useRoles();

    renderUsers();
    const dialog = await openEdit();

    expect(within(dialog).getByLabelText(/^Nombre/)).toHaveValue('Ana');
    expect(within(dialog).getByLabelText(/^Correo/)).toHaveValue('ana@ejemplo.com');
    expect(within(dialog).queryByLabelText(/Contraseña/)).toBeNull();
    expect(within(dialog).getByText(/189\.2\.4\.15/)).toBeInTheDocument();
  });

  it('la cuenta sin accesos indica que nunca ha entrado (US8 esc. 6)', async () => {
    useList([makeUser({ email: 'beto@ejemplo.com', lastLoginAt: null, lastLoginIp: null })], 1);
    useRoles();

    renderUsers();
    const dialog = await openEdit('beto@ejemplo.com');

    expect(within(dialog).getByText('Nunca ha entrado al panel.')).toBeInTheDocument();
  });

  it('al guardar no reenvía ningún campo de contraseña (FR-003/FR-026)', async () => {
    let patchBody: Record<string, unknown> | undefined;
    server.use(
      http.get(usersUrl, () =>
        HttpResponse.json({ items: [makeUser()], total: 1, limit: 20, offset: 0 }),
      ),
      http.get(rolesUrl, () =>
        HttpResponse.json({ items: [role], total: 1, limit: 100, offset: 0 }),
      ),
      http.patch(anaUrl, async ({ request }) => {
        patchBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json({ ...makeUser(), firstName: 'Ana María' });
      }),
    );

    renderUsers();
    const dialog = await openEdit();
    const firstName = within(dialog).getByLabelText(/^Nombre/);
    await userEvent.clear(firstName);
    await userEvent.type(firstName, 'Ana María');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Guardar cambios' }));

    expect(await screen.findByText('Cambios guardados.')).toBeInTheDocument();
    expect(patchBody).toEqual({
      firstName: 'Ana María',
      lastName: 'Pérez',
      email: 'ana@ejemplo.com',
      phone: '612345678',
      roleId: role.id,
    });
    expect(patchBody).not.toHaveProperty('password');
  });
});

describe('UsersPage · activar/desactivar', () => {
  it('desactivar pide confirmación, explica que los datos se conservan y refleja el estado', async () => {
    let active = true;
    server.use(
      http.get(usersUrl, () =>
        HttpResponse.json({
          items: [makeUser({ isActive: active })],
          total: 1,
          limit: 20,
          offset: 0,
        }),
      ),
      http.patch(anaUrl, async ({ request }) => {
        const body = (await request.json()) as { isActive: boolean };
        active = body.isActive;
        return HttpResponse.json(makeUser({ isActive: active }));
      }),
    );

    renderUsers();
    const table = await screen.findByRole('table');
    const row = within(table).getByText('ana@ejemplo.com').closest('tr');
    if (!row) {
      throw new Error('No se encontró la fila');
    }
    await userEvent.click(within(row).getByRole('button', { name: 'Desactivar' }));

    const dialog = await screen.findByRole('dialog', { name: 'Desactivar cuenta' });
    expect(dialog).toHaveTextContent('sus datos se conservan');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Desactivar' }));

    expect(await screen.findByText('Cuenta desactivada.')).toBeInTheDocument();
    await waitFor(() =>
      expect(within(screen.getByRole('table')).getByText('Inactivo')).toBeInTheDocument(),
    );
  });

  it('la regla del último administrador se explica como aviso (FR-008)', async () => {
    server.use(
      http.get(usersUrl, () =>
        HttpResponse.json({ items: [makeUser()], total: 1, limit: 20, offset: 0 }),
      ),
      http.patch(anaUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'conflict',
              message: 'No se puede dejar el panel sin administración',
              details: { reason: 'admin_required' },
            },
          },
          { status: 409 },
        ),
      ),
    );

    renderUsers();
    const table = await screen.findByRole('table');
    const row = within(table).getByText('ana@ejemplo.com').closest('tr');
    if (!row) {
      throw new Error('No se encontró la fila');
    }
    await userEvent.click(within(row).getByRole('button', { name: 'Desactivar' }));
    const dialog = await screen.findByRole('dialog', { name: 'Desactivar cuenta' });
    await userEvent.click(within(dialog).getByRole('button', { name: 'Desactivar' }));

    expect(await screen.findByText(LAST_ADMIN_ERROR)).toBeInTheDocument();
  });
});

describe('ResetPasswordDialog', () => {
  it('exige la política FR-010, restablece y no devuelve la contraseña', async () => {
    let resetBody: unknown;
    server.use(
      http.get(usersUrl, () =>
        HttpResponse.json({ items: [makeUser()], total: 1, limit: 20, offset: 0 }),
      ),
      http.get(rolesUrl, () =>
        HttpResponse.json({ items: [role], total: 1, limit: 100, offset: 0 }),
      ),
      http.post(anaPasswordUrl, async ({ request }) => {
        resetBody = await request.json();
        return HttpResponse.json({ passwordReset: true });
      }),
    );

    renderUsers();
    const editDialog = await openEdit();
    await userEvent.click(
      within(editDialog).getByRole('button', { name: 'Definir una contraseña nueva' }),
    );

    const dialog = await screen.findByRole('dialog', {
      name: 'Definir una contraseña nueva',
    });
    const passwordInput = within(dialog).getByLabelText(/^Contraseña nueva/);
    await userEvent.type(passwordInput, 'debil');

    expect(
      within(dialog).getByRole('list', { name: 'Requisitos de la contraseña' }),
    ).toHaveTextContent('Al menos 8 caracteres');
    expect(within(dialog).getByRole('button', { name: 'Restablecer contraseña' })).toBeDisabled();

    await userEvent.clear(passwordInput);
    await userEvent.type(passwordInput, 'Nueva.2026');
    await userEvent.type(
      within(dialog).getByLabelText(/^Confirmar contraseña nueva/),
      'Nueva.2026',
    );
    await userEvent.click(within(dialog).getByRole('button', { name: 'Restablecer contraseña' }));

    expect(await screen.findByText('Contraseña restablecida.')).toBeInTheDocument();
    expect(resetBody).toEqual({ password: 'Nueva.2026' });
  });
});
