import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { server } from '../test/server';
import { changeMyPassword, getSession, login, logout } from './auth';
import { API_BASE_URL, ApiError } from './client';

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

const loginUrl = `${API_BASE_URL}/api/v1/auth/login`;
const logoutUrl = `${API_BASE_URL}/api/v1/auth/logout`;
const sessionUrl = `${API_BASE_URL}/api/v1/auth/session`;
const passwordUrl = `${API_BASE_URL}/api/v1/auth/password`;

describe('login', () => {
  it('devuelve el SessionUser cuando el 200 trae la identidad', async () => {
    server.use(http.post(loginUrl, () => HttpResponse.json(sessionUser)));

    await expect(login({ email: 'ana@ejemplo.com', password: 'secreta' })).resolves.toEqual(
      sessionUser,
    );
  });

  it('propaga el 401 genérico como ApiError con su code', async () => {
    server.use(
      http.post(loginUrl, () =>
        HttpResponse.json(
          { error: { code: 'unauthenticated', message: 'Correo o contraseña incorrectos' } },
          { status: 401 },
        ),
      ),
    );

    const error = (await login({ email: 'x@x.com', password: 'mala' }).catch(
      (cause: unknown) => cause,
    )) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.reason).toBe('http');
    expect(error.code).toBe('unauthenticated');
    expect(error.message).toBe('Correo o contraseña incorrectos');
  });

  it('conserva retryAfterSeconds y reason de un 429/403 en details', async () => {
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

    const error = (await login({ email: 'x@x.com', password: 'mala' }).catch(
      (cause: unknown) => cause,
    )) as ApiError;

    expect(error.code).toBe('rate_limited');
    expect(error.details).toEqual({ retryAfterSeconds: 900 });
  });

  it('expone reason=access_disabled al 403 de cuenta desactivada', async () => {
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

    const error = (await login({ email: 'x@x.com', password: 'buena' }).catch(
      (cause: unknown) => cause,
    )) as ApiError;

    expect(error.status).toBe(403);
    expect(error.details?.reason).toBe('access_disabled');
  });
});

describe('logout', () => {
  it('devuelve loggedOut true', async () => {
    server.use(http.post(logoutUrl, () => HttpResponse.json({ loggedOut: true })));

    await expect(logout()).resolves.toEqual({ loggedOut: true });
  });

  it('propaga el error del sobre', async () => {
    server.use(
      http.post(logoutUrl, () =>
        HttpResponse.json(
          { error: { code: 'unauthenticated', message: 'Sin sesión' } },
          { status: 401 },
        ),
      ),
    );

    const error = (await logout().catch((cause: unknown) => cause)) as ApiError;
    expect(error.code).toBe('unauthenticated');
  });
});

describe('getSession', () => {
  it('devuelve la sesión actual', async () => {
    server.use(http.get(sessionUrl, () => HttpResponse.json(sessionUser)));

    await expect(getSession()).resolves.toEqual(sessionUser);
  });

  it('propaga el 401 cuando no hay sesión', async () => {
    server.use(
      http.get(sessionUrl, () =>
        HttpResponse.json(
          { error: { code: 'unauthenticated', message: 'Sin sesión' } },
          { status: 401 },
        ),
      ),
    );

    const error = (await getSession().catch((cause: unknown) => cause)) as ApiError;
    expect(error.status).toBe(401);
  });
});

describe('changeMyPassword', () => {
  it('devuelve passwordChanged true', async () => {
    server.use(http.post(passwordUrl, () => HttpResponse.json({ passwordChanged: true })));

    await expect(
      changeMyPassword({ currentPassword: 'vieja', newPassword: 'Nueva123!' }),
    ).resolves.toEqual({ passwordChanged: true });
  });

  it('propaga un 400 con details de requisito incumplido', async () => {
    server.use(
      http.post(passwordUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Revisa la contraseña',
              details: { newPassword: 'no cumple la política' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    const error = (await changeMyPassword({ currentPassword: 'x', newPassword: 'y' }).catch(
      (cause: unknown) => cause,
    )) as ApiError;

    expect(error.code).toBe('invalid');
    expect(error.details).toEqual({ newPassword: 'no cumple la política' });
  });
});
