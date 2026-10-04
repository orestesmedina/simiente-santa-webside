import { delay, http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { server } from '../test/server';
import { API_BASE_URL, ApiError, apiFetch, type ErrorEnvelope } from './client';
import { getSystemStatus } from './status';

const healthzUrl = `${API_BASE_URL}/healthz`;

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
