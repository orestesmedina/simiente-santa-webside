import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { API_BASE_URL, ApiError } from '../../../api/client';
import type { DonacionAdmin, DonacionTextoAdmin } from '../../../api/ofrendas';
import { createQueryClientForTest } from '../../../test/queryClient';
import { server } from '../../../test/server';
import { PORTADA_OFRENDAS_QUERY_KEY } from '../../publico/ofrendas/hooks/usePortadaOfrendas';
import { PUBLIC_OFRENDAS_QUERY_KEY } from '../../publico/ofrendas/hooks/usePublicOfrendas';
import { useCreateDonacion } from './useCreateDonacion';
import { useDeleteDonacion } from './useDeleteDonacion';
import { useDonacionTextos } from './useDonacionTextos';
import { DONACIONES_ADMIN_QUERY_KEY, DONACIONES_PAGE_LIMIT, useDonaciones } from './useDonaciones';
import { DONACIONES_TEXTOS_QUERY_KEY } from './queryKeys';
import { usePutDonacionTexto } from './usePutDonacionTexto';
import { useUpdateDonacion } from './useUpdateDonacion';

const adminUrl = `${API_BASE_URL}/api/v1/admin/donaciones`;
const textosUrl = `${adminUrl}/textos`;
const itemId = '11111111-1111-1111-1111-111111111111';

const donacion: DonacionAdmin = {
  id: itemId,
  kind: 'iban',
  number: 'CR05015202001234567890',
  institution: 'Banco Nacional de Costa Rica',
  holderEs: 'Simiente Santa',
  holderEn: null,
  currency: 'CRC',
  publicationState: 'draft',
  sortOrder: 0,
  createdAt: '2026-10-01T10:00:00Z',
  updatedAt: '2026-10-01T10:00:00Z',
};

const texto: DonacionTextoAdmin = {
  textKey: 'intro',
  textEs: 'Tu aporte va directo a la obra.',
  textEn: null,
  publicationState: 'draft',
  createdAt: '2026-10-01T10:00:00Z',
  updatedAt: '2026-10-01T10:00:00Z',
};

function makeWrapper() {
  const queryClient = createQueryClientForTest();
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return { queryClient, wrapper };
}

describe('useDonaciones (listado del panel, FR-010)', () => {
  it('pide el tope del contrato sin paginar (ux.md §4.3)', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(adminUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [donacion], total: 1, limit: 100, offset: 0 });
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useDonaciones(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(consultada).toBe(`${adminUrl}?limit=${DONACIONES_PAGE_LIMIT}`);
    expect(result.current.data?.items[0]?.institution).toBe('Banco Nacional de Costa Rica');
  });

  it('envía el filtro de estado cuando se indica', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(adminUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 });
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useDonaciones({ publicationState: 'published' }), {
      wrapper,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(consultada).toBe(
      `${adminUrl}?limit=${DONACIONES_PAGE_LIMIT}&publicationState=published`,
    );
  });

  it('propaga el `403` del servidor sin el permiso `donaciones` (FR-011)', async () => {
    server.use(
      http.get(adminUrl, () =>
        HttpResponse.json(
          { error: { code: 'forbidden', message: 'No tienes acceso a esta sección' } },
          { status: 403 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useDonaciones(), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as ApiError).status).toBe(403);
  });

  it('propaga el `401` cuando la sesión no es válida', async () => {
    server.use(
      http.get(adminUrl, () =>
        HttpResponse.json(
          { error: { code: 'unauthenticated', message: 'Sesión requerida' } },
          { status: 401 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useDonaciones(), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as ApiError).status).toBe(401);
  });
});

describe('useDonacionTextos (FR-021)', () => {
  it('devuelve los dos singletons, `null` hasta el primer guardado', async () => {
    server.use(http.get(textosUrl, () => HttpResponse.json({ intro: texto, usage: null })));

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useDonacionTextos(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.intro?.textEs).toBe('Tu aporte va directo a la obra.');
    expect(result.current.data?.usage).toBeNull();
  });
});

describe('invalidación de caché tras cada mutación (SC-003)', () => {
  it('el alta de un método invalida panel, textos, sección y bloque públicos', async () => {
    server.use(
      http.post(adminUrl, () => HttpResponse.json(donacion, { status: 201 })),
      http.get(adminUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
    );

    const { queryClient, wrapper } = makeWrapper();
    const spy = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useCreateDonacion(), { wrapper });

    result.current.mutate({
      kind: 'iban',
      number: 'CR05015202001234567890',
      institution: 'Banco Nacional',
      currency: 'CRC',
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith({ queryKey: DONACIONES_ADMIN_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: DONACIONES_TEXTOS_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: PUBLIC_OFRENDAS_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: PORTADA_OFRENDAS_QUERY_KEY });
  });

  it('editar y eliminar refrescan también las superficies públicas', async () => {
    server.use(
      http.patch(`${adminUrl}/${itemId}`, () => HttpResponse.json(donacion)),
      http.delete(`${adminUrl}/${itemId}`, () => new HttpResponse(null, { status: 204 })),
    );

    const { queryClient, wrapper } = makeWrapper();
    const spy = vi.spyOn(queryClient, 'invalidateQueries');

    const update = renderHook(() => useUpdateDonacion(), { wrapper });
    update.result.current.mutate({ id: itemId, input: { sortOrder: 3 } });
    await waitFor(() => expect(update.result.current.isSuccess).toBe(true));

    const remove = renderHook(() => useDeleteDonacion(), { wrapper });
    remove.result.current.mutate(itemId);
    await waitFor(() => expect(remove.result.current.isSuccess).toBe(true));

    expect(spy).toHaveBeenCalledWith({ queryKey: DONACIONES_ADMIN_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: PUBLIC_OFRENDAS_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: PORTADA_OFRENDAS_QUERY_KEY });
  });

  it('guardar un texto invalida la zona de textos y las superficies públicas', async () => {
    server.use(http.put(`${textosUrl}/intro`, () => HttpResponse.json({ ...texto, textEs: 'x' })));

    const { queryClient, wrapper } = makeWrapper();
    const spy = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => usePutDonacionTexto(), { wrapper });

    result.current.mutate({ textKey: 'intro', input: { textEs: 'x' } });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith({ queryKey: DONACIONES_TEXTOS_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: PUBLIC_OFRENDAS_QUERY_KEY });
  });
});

describe('errores de las mutaciones del panel (FR-014/FR-015/FR-016)', () => {
  it('el `400 invalid` conserva los `details` por campo', async () => {
    server.use(
      http.post(adminUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Datos no válidos',
              details: { number: 'Ese número no parece una cuenta IBAN válida.' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useCreateDonacion(), { wrapper });

    result.current.mutate({
      kind: 'iban',
      number: 'CR1',
      institution: 'Banco Nacional',
      currency: 'CRC',
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    const error = result.current.error as ApiError;
    expect(error.status).toBe(400);
    expect(error.details).toEqual({ number: 'Ese número no parece una cuenta IBAN válida.' });
  });

  it('el duplicado responde `409 conflict`', async () => {
    server.use(
      http.post(adminUrl, () =>
        HttpResponse.json(
          { error: { code: 'conflict', message: 'Ya existe esa misma cuenta o número' } },
          { status: 409 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useCreateDonacion(), { wrapper });

    result.current.mutate({
      kind: 'sinpe',
      number: '+50688888888',
      institution: 'Banco Promerica',
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as ApiError).status).toBe(409);
  });

  it('el `404 not_found` del `PATCH` propaga su código', async () => {
    server.use(
      http.patch(`${adminUrl}/${itemId}`, () =>
        HttpResponse.json(
          { error: { code: 'not_found', message: 'Recurso no encontrado' } },
          { status: 404 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useUpdateDonacion(), { wrapper });

    result.current.mutate({ id: itemId, input: { institution: 'Otro banco' } });

    await waitFor(() => expect(result.current.isError).toBe(true));
    const error = result.current.error as ApiError;
    expect(error.status).toBe(404);
    expect(error.code).toBe('not_found');
  });
});
