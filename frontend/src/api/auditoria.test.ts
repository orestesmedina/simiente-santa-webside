import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { server } from '../test/server';
import { API_BASE_URL, ApiError } from './client';
import { listAccessEvents, listAdminActions } from './auditoria';

const accesosUrl = `${API_BASE_URL}/api/v1/admin/auditoria/accesos`;
const accionesUrl = `${API_BASE_URL}/api/v1/admin/auditoria/acciones`;

const acceso = {
  id: '44444444-4444-4444-4444-444444444444',
  userId: null,
  userEmail: null,
  userName: null,
  result: 'failure',
  ip: '189.2.4.15',
  createdAt: '2026-10-03T17:42:00Z',
};

const accion = {
  id: '55555555-5555-5555-5555-555555555555',
  actorId: '11111111-1111-1111-1111-111111111111',
  actorEmail: 'ana@ejemplo.com',
  actorName: 'Ana Pérez',
  action: 'user.deactivate',
  targetKind: 'user',
  targetId: '66666666-6666-6666-6666-666666666666',
  targetLabel: 'luis@ejemplo.com',
  result: 'success',
  createdAt: '2026-10-03T17:45:00Z',
};

describe('listAccessEvents', () => {
  it('sin filtros consulta la ruta base', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(accesosUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [acceso], total: 1, limit: 20, offset: 0 });
      }),
    );

    const resultado = await listAccessEvents();
    expect(consultada).toBe(accesosUrl);
    expect(resultado.items[0]?.userName).toBeNull();
  });

  it('construye la URL con cuenta, rango y paginación', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(accesosUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [], total: 0, limit: 20, offset: 0 });
      }),
    );

    await listAccessEvents({
      userId: '11111111-1111-1111-1111-111111111111',
      from: '2026-10-01T00:00:00Z',
      to: '2026-10-05T00:00:00Z',
      limit: 20,
      offset: 0,
    });

    const url = new URL(consultada ?? '');
    expect(url.searchParams.get('userId')).toBe('11111111-1111-1111-1111-111111111111');
    expect(url.searchParams.get('from')).toBe('2026-10-01T00:00:00Z');
    expect(url.searchParams.get('to')).toBe('2026-10-05T00:00:00Z');
    expect(url.searchParams.get('limit')).toBe('20');
    expect(url.searchParams.get('offset')).toBe('0');
  });

  it('propaga el ApiError del sobre', async () => {
    server.use(
      http.get(accesosUrl, () =>
        HttpResponse.json(
          { error: { code: 'forbidden', message: 'Sin permiso' } },
          { status: 403 },
        ),
      ),
    );

    const error = (await listAccessEvents().catch((cause: unknown) => cause)) as ApiError;
    expect(error.code).toBe('forbidden');
  });
});

describe('listAdminActions', () => {
  it('consulta la ruta de acciones y tipa el DTO', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(accionesUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [accion], total: 1, limit: 20, offset: 0 });
      }),
    );

    const resultado = await listAdminActions({ userId: 'abc' });

    expect(consultada).toBe(`${accionesUrl}?userId=abc`);
    expect(resultado.items[0]?.action).toBe('user.deactivate');
    expect(resultado.items[0]?.result).toBe('success');
  });
});
