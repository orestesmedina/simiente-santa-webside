import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { server } from '../test/server';
import { API_BASE_URL, ApiError } from './client';
import { createRole, deleteRole, getRole, listPermissions, listRoles, updateRole } from './roles';

const role = {
  id: '33333333-3333-3333-3333-333333333333',
  name: 'Editores',
  permissions: ['eventos', 'noticias'],
  userCount: 2,
  createdAt: '2026-10-01T10:00:00Z',
};

const listUrl = `${API_BASE_URL}/api/v1/admin/roles`;
const roleUrl = `${listUrl}/${role.id}`;
const permisosUrl = `${API_BASE_URL}/api/v1/admin/permisos`;

describe('listRoles', () => {
  it('construye la URL con paginación', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(listUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [role], total: 1, limit: 10, offset: 10 });
      }),
    );

    const resultado = await listRoles({ limit: 10, offset: 10 });

    expect(consultada).toBe(`${listUrl}?limit=10&offset=10`);
    expect(resultado.items[0]?.userCount).toBe(2);
  });
});

describe('getRole', () => {
  it('consulta el detalle por id', async () => {
    server.use(http.get(roleUrl, () => HttpResponse.json(role)));
    await expect(getRole(role.id)).resolves.toEqual(role);
  });
});

describe('createRole', () => {
  it('hace POST con nombre y permisos', async () => {
    let cuerpo: unknown;
    server.use(
      http.post(listUrl, async ({ request }) => {
        cuerpo = await request.json();
        return HttpResponse.json(role, { status: 201 });
      }),
    );

    await createRole({ name: 'Editores', permissions: ['eventos', 'noticias'] });
    expect(cuerpo).toEqual({ name: 'Editores', permissions: ['eventos', 'noticias'] });
  });
});

describe('updateRole', () => {
  it('hace PATCH sobre el id', async () => {
    let metodo: string | undefined;
    server.use(
      http.patch(roleUrl, ({ request }) => {
        metodo = request.method;
        return HttpResponse.json({ ...role, name: 'Editores 2' });
      }),
    );

    const actualizado = await updateRole(role.id, { name: 'Editores 2' });
    expect(metodo).toBe('PATCH');
    expect(actualizado.name).toBe('Editores 2');
  });
});

describe('deleteRole', () => {
  it('hace DELETE y devuelve deleted', async () => {
    let metodo: string | undefined;
    server.use(
      http.delete(roleUrl, ({ request }) => {
        metodo = request.method;
        return HttpResponse.json({ deleted: true });
      }),
    );

    await expect(deleteRole(role.id)).resolves.toEqual({ deleted: true });
    expect(metodo).toBe('DELETE');
  });

  it('propaga el 409 con el mensaje de rol en uso', async () => {
    server.use(
      http.delete(roleUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'conflict',
              message: 'No se puede eliminar: 2 cuentas usan este rol',
              details: { reason: 'role_in_use' },
            },
          },
          { status: 409 },
        ),
      ),
    );

    const error = (await deleteRole(role.id).catch((cause: unknown) => cause)) as ApiError;
    expect(error.details?.reason).toBe('role_in_use');
  });
});

describe('listPermissions', () => {
  it('devuelve el catálogo de permisos', async () => {
    server.use(
      http.get(permisosUrl, () =>
        HttpResponse.json({ items: [{ code: 'eventos', label: 'Eventos' }] }),
      ),
    );

    await expect(listPermissions()).resolves.toEqual({
      items: [{ code: 'eventos', label: 'Eventos' }],
    });
  });
});
