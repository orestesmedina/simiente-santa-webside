import { delay, http, HttpResponse } from 'msw';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { server } from '../test/server';
import { API_BASE_URL, ApiError, apiFetch, readCookie, type ErrorEnvelope } from './client';
import { getSystemStatus } from './status';

const healthzUrl = `${API_BASE_URL}/healthz`;

function clearCsrfCookie() {
  document.cookie = 'csrf_token=; Max-Age=0; path=/';
}

const databaseUnavailable: ErrorEnvelope = {
  error: {
    code: 'database_unavailable',
    message: 'La base de datos no está conectada',
    details: { database: 'disconnected' },
  },
};

describe('apiFetch', () => {
  it('devuelve el DTO directo cuando la respuesta es 2xx', async () => {
    server.use(
      http.get(healthzUrl, () => HttpResponse.json({ status: 'ok', database: 'connected' })),
    );

    await expect(apiFetch('/healthz')).resolves.toEqual({
      status: 'ok',
      database: 'connected',
    });
  });

  it('convierte un 503 con sobre de error en ApiError sin perder code/message/details', async () => {
    server.use(http.get(healthzUrl, () => HttpResponse.json(databaseUnavailable, { status: 503 })));

    const error = await apiFetch('/healthz').catch((cause: unknown) => cause);

    expect(error).toBeInstanceOf(ApiError);
    const apiError = error as ApiError;
    expect(apiError.reason).toBe('http');
    expect(apiError.status).toBe(503);
    expect(apiError.code).toBe('database_unavailable');
    expect(apiError.message).toBe('La base de datos no está conectada');
    expect(apiError.details).toEqual({ database: 'disconnected' });
  });

  it('convierte un 500 con sobre de error en ApiError con code internal', async () => {
    server.use(
      http.get(healthzUrl, () =>
        HttpResponse.json(
          { error: { code: 'internal', message: 'Error interno del servidor' } },
          { status: 500 },
        ),
      ),
    );

    const error = (await apiFetch('/healthz').catch((cause: unknown) => cause)) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.reason).toBe('http');
    expect(error.code).toBe('internal');
  });

  it('convierte un 404 con sobre de error en ApiError con code not_found', async () => {
    server.use(
      http.get(healthzUrl, () =>
        HttpResponse.json(
          { error: { code: 'not_found', message: 'Recurso no encontrado' } },
          { status: 404 },
        ),
      ),
    );

    const error = (await apiFetch('/healthz').catch((cause: unknown) => cause)) as ApiError;

    expect(error.code).toBe('not_found');
  });

  it('trata un error con cuerpo que no cumple el sobre como respuesta no válida', async () => {
    server.use(
      http.get(healthzUrl, () => new HttpResponse('Error de texto plano', { status: 500 })),
    );

    const error = (await apiFetch('/healthz').catch((cause: unknown) => cause)) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.reason).toBe('invalid-response');
  });

  it('trata un JSON malformado como respuesta no válida', async () => {
    server.use(
      http.get(
        healthzUrl,
        () =>
          new HttpResponse('{ esto no es json', {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          }),
      ),
    );

    const error = (await apiFetch('/healthz').catch((cause: unknown) => cause)) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.reason).toBe('invalid-response');
  });

  it('trata un error de red como inaccesible', async () => {
    server.use(http.get(healthzUrl, () => HttpResponse.error()));

    const error = (await apiFetch('/healthz').catch((cause: unknown) => cause)) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.reason).toBe('network');
  });

  it('corta la consulta por timeout como inaccesible', async () => {
    server.use(
      http.get(healthzUrl, async () => {
        await delay('infinite');
        return HttpResponse.json({});
      }),
    );

    const error = (await apiFetch('/healthz', { timeoutMs: 20 }).catch(
      (cause: unknown) => cause,
    )) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.reason).toBe('timeout');
  });
});

describe('getSystemStatus', () => {
  it('devuelve el SystemStatus cuando el 200 trae el sobre de éxito completo', async () => {
    server.use(
      http.get(healthzUrl, () => HttpResponse.json({ status: 'ok', database: 'connected' })),
    );

    await expect(getSystemStatus()).resolves.toEqual({ status: 'ok', database: 'connected' });
  });

  it('trata un 200 con cuerpo incompleto como respuesta no válida (nunca conectado)', async () => {
    server.use(http.get(healthzUrl, () => HttpResponse.json({ status: 'ok' })));

    const error = (await getSystemStatus().catch((cause: unknown) => cause)) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.reason).toBe('invalid-response');
  });

  it('trata un 200 con database no reconocida como respuesta no válida', async () => {
    server.use(
      http.get(healthzUrl, () => HttpResponse.json({ status: 'ok', database: 'degraded' })),
    );

    const error = (await getSystemStatus().catch((cause: unknown) => cause)) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.reason).toBe('invalid-response');
  });

  it('propaga el 503 database_unavailable como ApiError del sobre', async () => {
    server.use(http.get(healthzUrl, () => HttpResponse.json(databaseUnavailable, { status: 503 })));

    const error = (await getSystemStatus().catch((cause: unknown) => cause)) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.reason).toBe('http');
    expect(error.code).toBe('database_unavailable');
    expect(error.details).toEqual({ database: 'disconnected' });
  });
});

describe('apiFetch credenciales y CSRF', () => {
  afterEach(() => {
    clearCsrfCookie();
    vi.restoreAllMocks();
  });

  it('envía siempre credenciales incluidas', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch');
    server.use(
      http.get(healthzUrl, () => HttpResponse.json({ status: 'ok', database: 'connected' })),
    );

    await apiFetch('/healthz');

    const init = fetchSpy.mock.calls[0]?.[1];
    expect(init?.credentials).toBe('include');
  });

  it('añade X-CSRF-Token en métodos no seguros leyendo la cookie', async () => {
    document.cookie = 'csrf_token=token-de-prueba';
    let cabecera: string | null = null;
    server.use(
      http.post(`${API_BASE_URL}/api/v1/auth/logout`, ({ request }) => {
        cabecera = request.headers.get('X-CSRF-Token');
        return HttpResponse.json({ loggedOut: true });
      }),
    );

    await apiFetch('/api/v1/auth/logout', { method: 'POST' });

    expect(cabecera).toBe('token-de-prueba');
  });

  it('no añade X-CSRF-Token en métodos seguros', async () => {
    document.cookie = 'csrf_token=token-de-prueba';
    let cabecera: string | null = 'no-consultada';
    server.use(
      http.get(`${API_BASE_URL}/api/v1/auth/session`, ({ request }) => {
        cabecera = request.headers.get('X-CSRF-Token');
        return HttpResponse.json({});
      }),
    );

    await apiFetch('/api/v1/auth/session');

    expect(cabecera).toBeNull();
  });

  it('respeta una cabecera X-CSRF-Token explícita del llamador', async () => {
    document.cookie = 'csrf_token=de-la-cookie';
    let cabecera: string | null = null;
    server.use(
      http.post(`${API_BASE_URL}/api/v1/auth/logout`, ({ request }) => {
        cabecera = request.headers.get('X-CSRF-Token');
        return HttpResponse.json({ loggedOut: true });
      }),
    );

    await apiFetch('/api/v1/auth/logout', {
      method: 'POST',
      headers: { 'X-CSRF-Token': 'explicita' },
    });

    expect(cabecera).toBe('explicita');
  });

  it('sin cookie no envía la cabecera CSRF', async () => {
    let cabecera: string | null = 'no-consultada';
    server.use(
      http.post(`${API_BASE_URL}/api/v1/auth/login`, ({ request }) => {
        cabecera = request.headers.get('X-CSRF-Token');
        return HttpResponse.json({});
      }),
    );

    await apiFetch('/api/v1/auth/login', { method: 'POST' });

    expect(cabecera).toBeNull();
  });
});

describe('readCookie', () => {
  afterEach(() => clearCsrfCookie());

  it('devuelve el valor de la cookie existente', () => {
    document.cookie = 'csrf_token=abc123';
    expect(readCookie('csrf_token')).toBe('abc123');
  });

  it('devuelve null si la cookie no existe', () => {
    expect(readCookie('csrf_token')).toBeNull();
  });
});
