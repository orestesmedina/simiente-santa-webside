import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { RoleItem } from '../../api/roles';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { EMPTY_ROLES_MESSAGE, NO_PERMISSIONS_ERROR } from './messages';
import { RolesPage } from './pages/RolesPage';

const rolesUrl = `${API_BASE_URL}/api/v1/admin/roles`;
const permissionsUrl = `${API_BASE_URL}/api/v1/admin/permisos`;

const catalogo = {
  items: [
    { code: 'portada', label: 'Portada e información general' },
    { code: 'eventos', label: 'Eventos' },
    { code: 'noticias', label: 'Noticias y galería' },
    { code: 'admin_usuarios_roles', label: 'Administración de usuarios y roles' },
  ],
};

function makeRole(overrides: Partial<RoleItem> = {}): RoleItem {
  return {
    id: '33333333-3333-3333-3333-333333333333',
    name: 'Editores',
    permissions: ['eventos'],
    userCount: 0,
    createdAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function renderRoles() {
  return render(
    <AppProviders>
      <RolesPage />
    </AppProviders>,
  );
}

describe('RolesPage · listado', () => {
  it('muestra el resumen de permisos, el sello de módulo reservado y las cuentas asignadas', async () => {
    const role = makeRole({ permissions: ['eventos', 'admin_usuarios_roles'], userCount: 2 });
    server.use(
      http.get(rolesUrl, () =>
        HttpResponse.json({ items: [role], total: 1, limit: 20, offset: 0 }),
      ),
      http.get(permissionsUrl, () => HttpResponse.json(catalogo)),
    );

    renderRoles();
    const table = await screen.findByRole('table');

    expect(within(table).getByText('Editores')).toBeInTheDocument();
    expect(await within(table).findByText('Eventos')).toBeInTheDocument();
    expect(within(table).getByText(/Disponible más adelante/)).toBeInTheDocument();
    expect(within(table).getByText('2 cuentas')).toBeInTheDocument();
  });

  it('sin roles muestra el estado vacío con la acción de crear', async () => {
    server.use(
      http.get(rolesUrl, () => HttpResponse.json({ items: [], total: 0, limit: 20, offset: 0 })),
      http.get(permissionsUrl, () => HttpResponse.json(catalogo)),
    );

    renderRoles();

    expect(await screen.findByText(EMPTY_ROLES_MESSAGE)).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: 'Crear rol' }).length).toBeGreaterThan(0);
  });
});

describe('RoleForm · crear', () => {
  it('crea un rol con dos permisos, avisa "Rol creado." y refresca el listado', async () => {
    const roles: RoleItem[] = [makeRole()];
    let listCalls = 0;
    let postBody: unknown;
    server.use(
      http.get(rolesUrl, () => {
        listCalls += 1;
        return HttpResponse.json({ items: roles, total: roles.length, limit: 20, offset: 0 });
      }),
      http.get(permissionsUrl, () => HttpResponse.json(catalogo)),
      http.post(rolesUrl, async ({ request }) => {
        postBody = await request.json();
        const created = makeRole({ name: 'Editores', permissions: ['eventos', 'noticias'] });
        roles.push(created);
        return HttpResponse.json(created, { status: 201 });
      }),
    );

    renderRoles();
    await screen.findByRole('table');
    await userEvent.click(screen.getByRole('button', { name: 'Crear rol' }));
    const dialog = await screen.findByRole('dialog', { name: 'Crear rol' });

    await userEvent.type(within(dialog).getByLabelText(/^Nombre/), 'Editores');
    await userEvent.click(within(dialog).getByLabelText('Eventos'));
    await userEvent.click(within(dialog).getByLabelText('Noticias y galería'));
    await userEvent.click(within(dialog).getByRole('button', { name: 'Crear rol' }));

    expect(await screen.findByText('Rol creado.')).toBeInTheDocument();
    expect(postBody).toEqual({ name: 'Editores', permissions: ['eventos', 'noticias'] });
    await waitFor(() => expect(listCalls).toBeGreaterThanOrEqual(2));
  });

  it('sin permisos explica "al menos un permiso" y no envía nada (FR-014)', async () => {
    let postCount = 0;
    server.use(
      http.get(rolesUrl, () =>
        HttpResponse.json({ items: [makeRole()], total: 1, limit: 20, offset: 0 }),
      ),
      http.get(permissionsUrl, () => HttpResponse.json(catalogo)),
      http.post(rolesUrl, () => {
        postCount += 1;
        return HttpResponse.json(makeRole(), { status: 201 });
      }),
    );

    renderRoles();
    await screen.findByRole('table');
    await userEvent.click(screen.getByRole('button', { name: 'Crear rol' }));
    const dialog = await screen.findByRole('dialog', { name: 'Crear rol' });

    await userEvent.type(within(dialog).getByLabelText(/^Nombre/), 'Sin permisos');

    expect(within(dialog).getByText(NO_PERMISSIONS_ERROR)).toBeInTheDocument();
    const submit = within(dialog).getByRole('button', { name: 'Crear rol' });
    expect(submit).toBeDisabled();
    await userEvent.click(submit);
    expect(postCount).toBe(0);
  });

  it('con nombre duplicado normalizado muestra el mensaje de duplicado (SC-011)', async () => {
    let postCount = 0;
    server.use(
      http.get(rolesUrl, () =>
        HttpResponse.json({ items: [makeRole()], total: 1, limit: 20, offset: 0 }),
      ),
      http.get(permissionsUrl, () => HttpResponse.json(catalogo)),
      http.post(rolesUrl, () => {
        postCount += 1;
        return HttpResponse.json(
          { error: { code: 'conflict', message: 'Ya existe un rol con ese nombre' } },
          { status: 409 },
        );
      }),
    );

    renderRoles();
    await screen.findByRole('table');
    await userEvent.click(screen.getByRole('button', { name: 'Crear rol' }));
    const dialog = await screen.findByRole('dialog', { name: 'Crear rol' });

    await userEvent.type(within(dialog).getByLabelText(/^Nombre/), '  editores ');
    await userEvent.click(within(dialog).getByLabelText('Eventos'));
    await userEvent.click(within(dialog).getByRole('button', { name: 'Crear rol' }));

    expect(await screen.findByText('Ya existe un rol con ese nombre')).toBeInTheDocument();
    expect(postCount).toBe(1);
    expect(screen.queryByText('Rol creado.')).toBeNull();
  });
});

describe('RoleForm · editar', () => {
  it('refleja los permisos editados en el listado (FR-018)', async () => {
    const role = makeRole({ permissions: ['eventos'] });
    let listCalls = 0;
    let patchBody: Record<string, unknown> | undefined;
    server.use(
      http.get(rolesUrl, () => {
        listCalls += 1;
        const items =
          listCalls === 1 ? [role] : [{ ...role, permissions: ['eventos', 'noticias'] }];
        return HttpResponse.json({ items, total: 1, limit: 20, offset: 0 });
      }),
      http.get(permissionsUrl, () => HttpResponse.json(catalogo)),
      http.patch(`${rolesUrl}/${role.id}`, async ({ request }) => {
        patchBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json({ ...role, permissions: ['eventos', 'noticias'] });
      }),
    );

    renderRoles();
    const table = await screen.findByRole('table');
    await userEvent.click(within(table).getByRole('button', { name: 'Editar' }));
    const dialog = await screen.findByRole('dialog', { name: 'Editar rol' });

    await userEvent.click(within(dialog).getByLabelText('Noticias y galería'));
    await userEvent.click(within(dialog).getByRole('button', { name: 'Guardar cambios' }));

    expect(await screen.findByText('Cambios guardados.')).toBeInTheDocument();
    expect(patchBody).toEqual({ name: 'Editores', permissions: ['eventos', 'noticias'] });
    await waitFor(() =>
      expect(within(screen.getByRole('table')).getByText('Noticias y galería')).toBeInTheDocument(),
    );
  });
});

describe('RolesPage · eliminar', () => {
  it('elimina un rol sin uso tras confirmar y refresca el listado (FR-017)', async () => {
    const role = makeRole({ userCount: 0 });
    let deleteCalls = 0;
    server.use(
      http.get(rolesUrl, () =>
        HttpResponse.json({
          items: deleteCalls === 0 ? [role] : [],
          total: deleteCalls === 0 ? 1 : 0,
          limit: 20,
          offset: 0,
        }),
      ),
      http.get(permissionsUrl, () => HttpResponse.json(catalogo)),
      http.delete(`${rolesUrl}/${role.id}`, () => {
        deleteCalls += 1;
        return HttpResponse.json({ deleted: true });
      }),
    );

    renderRoles();
    const table = await screen.findByRole('table');
    await userEvent.click(within(table).getByRole('button', { name: 'Eliminar' }));

    const dialog = await screen.findByRole('dialog', { name: 'Eliminar rol' });
    expect(dialog).toHaveTextContent("Se borrará el rol 'Editores'");
    await userEvent.click(within(dialog).getByRole('button', { name: 'Eliminar' }));

    expect(await screen.findByText('Rol eliminado.')).toBeInTheDocument();
    expect(deleteCalls).toBe(1);
    expect(await screen.findByText(EMPTY_ROLES_MESSAGE)).toBeInTheDocument();
  });

  it('si el rol está en uso pide reasignar y NO llama a DELETE (US6 esc. 4–5)', async () => {
    const role = makeRole({ userCount: 3 });
    let deleteCalls = 0;
    server.use(
      http.get(rolesUrl, () =>
        HttpResponse.json({ items: [role], total: 1, limit: 20, offset: 0 }),
      ),
      http.get(permissionsUrl, () => HttpResponse.json(catalogo)),
      http.delete(`${rolesUrl}/${role.id}`, () => {
        deleteCalls += 1;
        return HttpResponse.json({ deleted: true });
      }),
    );

    renderRoles();
    const table = await screen.findByRole('table');
    await userEvent.click(within(table).getByRole('button', { name: 'Eliminar' }));

    const dialog = await screen.findByRole('dialog', { name: 'No se puede eliminar' });
    expect(dialog).toHaveTextContent('3 cuentas usan este rol');
    expect(dialog).toHaveTextContent('Primero cámbiales el rol');
    expect(deleteCalls).toBe(0);
  });
});
