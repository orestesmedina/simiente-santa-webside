import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { server } from '../test/server';
import { API_BASE_URL, ApiError } from './client';
import { createUser, getUser, listUsers, resetUserPassword, updateUser } from './usuarios';

const user = {
  id: '11111111-1111-1111-1111-111111111111',
  email: 'ana@ejemplo.com',
  firstName: 'Ana',
  lastName: 'Pérez',
  phone: '612345678',
  roleId: '22222222-2222-2222-2222-222222222222',
  roleName: 'Administración',
  isActive: true,
  mustChangePassword: false,
  lastLoginAt: null,
  lastLoginIp: null,
  createdAt: '2026-10-01T10:00:00Z',
};

const listUrl = `${API_BASE_URL}/api/v1/admin/usuarios`;
const userUrl = `${listUrl}/${user.id}`;

describe('listUsers', () => {
  it('sin parámetros no añade query', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(listUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [user], total: 1, limit: 20, offset: 0 });
      }),
    );

    const resultado = await listUsers();

    expect(consultada).toBe(listUrl);
    expect(resultado.items).toHaveLength(1);
    expect(resultado.total).toBe(1);
  });

  it('construye la URL con limit y offset', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(listUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [], total: 0, limit: 20, offset: 40 });
      }),
    );

    await listUsers({ limit: 20, offset: 40 });

    expect(consultada).toBe(`${listUrl}?limit=20&offset=40`);
  });

  it('propaga el ApiError del sobre', async () => {
    server.use(
      http.get(listUrl, () =>
        HttpResponse.json(
          { error: { code: 'forbidden', message: 'Sin permiso' } },
          { status: 403 },
        ),
      ),
    );

    const error = (await listUsers().catch((cause: unknown) => cause)) as ApiError;
    expect(error.code).toBe('forbidden');
    expect(error.status).toBe(403);
  });
});

describe('getUser', () => {
  it('consulta el detalle por id', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(userUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(user);
      }),
    );

    await expect(getUser(user.id)).resolves.toEqual(user);
    expect(consultada).toBe(userUrl);
  });
});

describe('createUser', () => {
  it('hace POST con el cuerpo tipado y devuelve la cuenta creada', async () => {
    let metodo: string | undefined;
    let cuerpo: unknown;
    server.use(
      http.post(listUrl, async ({ request }) => {
        metodo = request.method;
        cuerpo = await request.json();
        return HttpResponse.json(user, { status: 201 });
      }),
    );

    const input = {
      firstName: 'Ana',
      lastName: 'Pérez',
      email: 'ana@ejemplo.com',
      phone: '612345678',
      roleId: user.roleId,
      password: 'Nueva123!',
    };
    await expect(createUser(input)).resolves.toEqual(user);
    expect(metodo).toBe('POST');
    expect(cuerpo).toEqual(input);
  });

  it('conserva details.phone de un 400', async () => {
    server.use(
      http.post(listUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Revisa los datos',
              details: { phone: 'formato incorrecto' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    const error = (await createUser({} as never).catch((cause: unknown) => cause)) as ApiError;
    expect(error.details).toEqual({ phone: 'formato incorrecto' });
  });
});

describe('updateUser', () => {
  it('hace PATCH sobre el id', async () => {
    let metodo: string | undefined;
    let cuerpo: unknown;
    server.use(
      http.patch(userUrl, async ({ request }) => {
        metodo = request.method;
        cuerpo = await request.json();
        return HttpResponse.json({ ...user, isActive: false });
      }),
    );

    const actualizado = await updateUser(user.id, { isActive: false });
    expect(metodo).toBe('PATCH');
    expect(cuerpo).toEqual({ isActive: false });
    expect(actualizado.isActive).toBe(false);
  });
});

describe('resetUserPassword', () => {
  it('hace POST en /password y devuelve passwordReset', async () => {
    let consultada: string | undefined;
    server.use(
      http.post(`${userUrl}/password`, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ passwordReset: true });
      }),
    );

    await expect(resetUserPassword(user.id, { password: 'Nueva123!' })).resolves.toEqual({
      passwordReset: true,
    });
    expect(consultada).toBe(`${userUrl}/password`);
  });
});
