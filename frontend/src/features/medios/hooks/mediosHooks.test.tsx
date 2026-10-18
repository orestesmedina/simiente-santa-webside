import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { API_BASE_URL, ApiError } from '../../../api/client';
import type { MedioAdmin, ResumenBusqueda } from '../../../api/medios';
import { createQueryClientForTest } from '../../../test/queryClient';
import { server } from '../../../test/server';
import { PORTADA_MEDIOS_QUERY_KEY } from '../../publico/medios/hooks/usePortadaMedios';
import { PUBLIC_MEDIOS_QUERY_KEY } from '../../publico/medios/hooks/usePublicMedios';
import { MEDIOS_ADMIN_QUERY_KEY, MEDIOS_PAGE_LIMIT, useMedios } from './useMedios';
import { useActualizarDescripcion } from './useActualizarDescripcion';
import { useBuscarEpisodios } from './useBuscarEpisodios';
import { useOcultarMostrarEpisodio } from './useOcultarMostrarEpisodio';
import { useResolverMarca } from './useResolverMarca';
import { useUnirDuplicados } from './useUnirDuplicados';

const adminUrl = `${API_BASE_URL}/api/v1/admin/medios`;
const itemId = '11111111-1111-1111-1111-111111111111';
const otherId = '22222222-2222-2222-2222-222222222222';

const medio: MedioAdmin = {
  id: itemId,
  title: 'Potencial sin Resultados La Lección de la Higuera',
  descriptionEs: 'La lección de la higuera.',
  descriptionEn: null,
  descriptionEdited: false,
  publishedAt: '2026-07-03T01:22:48Z',
  coverUrl: '/api/v1/media/img_1111.jpg',
  videoUrl: 'https://www.youtube.com/watch?v=abcdefghijk',
  audioUrl: 'https://podcasters.spotify.com/pod/show/simiente/episodes/1',
  origin: 'both',
  visibility: 'visible',
  sources: [
    {
      provider: 'youtube',
      externalId: 'abcdefghijk',
      feedTitle: 'Potencial sin Resultados La Lección de la Higuera',
      url: 'https://www.youtube.com/watch?v=abcdefghijk',
    },
  ],
  createdAt: '2026-10-17T10:00:00Z',
  updatedAt: '2026-10-17T10:00:00Z',
};

const resumen: ResumenBusqueda = {
  startedAt: '2026-10-17T10:00:00Z',
  finishedAt: '2026-10-17T10:00:12Z',
  importedCount: 10,
  joinedCount: 8,
  markedCount: 0,
  hiddenCount: 0,
  availabilityCheckedCount: 0,
  availabilityInconclusiveCount: 0,
  youtube: 'ok',
  podcast: 'ok',
};

function makeWrapper() {
  const queryClient = createQueryClientForTest();
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return { queryClient, wrapper };
}

describe('useMedios (listado del panel, FR-019)', () => {
  it('pide el tope del contrato sin paginar y trae el origen y la última búsqueda', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(adminUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({
          items: [medio],
          total: 1,
          limit: 100,
          offset: 0,
          lastSync: { ...resumen, id: otherId, actorUserId: itemId },
        });
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useMedios(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(consultada).toBe(`${adminUrl}?limit=${MEDIOS_PAGE_LIMIT}`);
    expect(result.current.data?.items[0]?.origin).toBe('both');
    expect(result.current.data?.items[0]?.createdAt).toBe('2026-10-17T10:00:00Z');
    expect(result.current.data?.lastSync?.importedCount).toBe(10);
  });

  it('envía los filtros de visibilidad y de marca pendiente (FR-015/FR-017)', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(adminUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 });
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(
      () => useMedios({ visibility: 'hidden', reviewMark: 'possible_duplicate' }),
      { wrapper },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(consultada).toBe(
      `${adminUrl}?limit=${MEDIOS_PAGE_LIMIT}&visibility=hidden&reviewMark=possible_duplicate`,
    );
  });

  it('ante un error conserva el último listado válido (ux.md §5.b)', async () => {
    let llamadas = 0;
    server.use(
      http.get(adminUrl, () => {
        llamadas += 1;
        if (llamadas === 1) {
          return HttpResponse.json({ items: [medio], total: 1, limit: 100, offset: 0 });
        }
        return HttpResponse.json(
          { error: { code: 'internal', message: 'Falla del servidor' } },
          { status: 500 },
        );
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useMedios(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    await result.current.refetch();

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.data?.items).toHaveLength(1);
  });

  it('propaga el `403` sin el permiso `medios` y el `401` sin sesión (FR-021)', async () => {
    server.use(
      http.get(adminUrl, () =>
        HttpResponse.json(
          { error: { code: 'forbidden', message: 'No tienes acceso a esta sección' } },
          { status: 403 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const first = renderHook(() => useMedios(), { wrapper });
    await waitFor(() => expect(first.result.current.isError).toBe(true));
    expect((first.result.current.error as ApiError).status).toBe(403);

    server.use(
      http.get(adminUrl, () =>
        HttpResponse.json(
          { error: { code: 'unauthenticated', message: 'Sesión requerida' } },
          { status: 401 },
        ),
      ),
    );

    const second = renderHook(() => useMedios(), { wrapper });
    await waitFor(() => expect(second.result.current.isError).toBe(true));
    expect((second.result.current.error as ApiError).status).toBe(401);
  });
});

describe('useBuscarEpisodios (FR-011/FR-018)', () => {
  it('queda «en vuelo» mientras corre y refresca panel, sección y bloque', async () => {
    let resolver: (() => void) | undefined;
    server.use(
      http.post(`${adminUrl}/buscar-nuevos`, async () => {
        await new Promise<void>((resolve) => {
          resolver = resolve;
        });
        return HttpResponse.json(resumen);
      }),
      http.get(adminUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
    );

    const { queryClient, wrapper } = makeWrapper();
    const spy = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useBuscarEpisodios(), { wrapper });

    result.current.mutate();

    await waitFor(() => expect(result.current.isPending).toBe(true));
    expect(result.current.data).toBeUndefined();

    await waitFor(() => expect(resolver).toBeDefined());
    resolver?.();

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.importedCount).toBe(10);
    expect(spy).toHaveBeenCalledWith({ queryKey: MEDIOS_ADMIN_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: PUBLIC_MEDIOS_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: PORTADA_MEDIOS_QUERY_KEY });
  });

  it('un feed caído no es error: el resumen viaja igual (FR-016)', async () => {
    server.use(
      http.post(`${adminUrl}/buscar-nuevos`, () =>
        HttpResponse.json({ ...resumen, youtube: 'failed' }),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useBuscarEpisodios(), { wrapper });

    result.current.mutate();

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.youtube).toBe('failed');
  });

  it('propaga el `500` del servidor', async () => {
    server.use(
      http.post(`${adminUrl}/buscar-nuevos`, () =>
        HttpResponse.json(
          { error: { code: 'internal', message: 'Falla del servidor' } },
          { status: 500 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useBuscarEpisodios(), { wrapper });

    result.current.mutate();

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as ApiError).status).toBe(500);
  });
});

describe('mutations del catálogo (FR-015/FR-020/FR-022)', () => {
  it('guardar la descripción hace `PATCH` con los dos idiomas y refresca todo', async () => {
    let metodo: string | undefined;
    let cuerpo: unknown;
    server.use(
      http.patch(`${adminUrl}/${itemId}`, async ({ request }) => {
        metodo = request.method;
        cuerpo = await request.json();
        return HttpResponse.json({ ...medio, descriptionEdited: true });
      }),
      http.get(adminUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
    );

    const { queryClient, wrapper } = makeWrapper();
    const spy = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useActualizarDescripcion(), { wrapper });

    result.current.mutate({
      id: itemId,
      input: { descriptionEs: 'Nuevo texto.', descriptionEn: 'New text.' },
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(metodo).toBe('PATCH');
    expect(cuerpo).toEqual({ descriptionEs: 'Nuevo texto.', descriptionEn: 'New text.' });
    expect(spy).toHaveBeenCalledWith({ queryKey: PUBLIC_MEDIOS_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: PORTADA_MEDIOS_QUERY_KEY });
  });

  it('el `400 invalid` conserva los `details` por campo (FR-023)', async () => {
    server.use(
      http.patch(`${adminUrl}/${itemId}`, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Datos no válidos',
              details: { descriptionEn: 'Falta la versión en español de la descripción corta.' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useActualizarDescripcion(), { wrapper });

    result.current.mutate({ id: itemId, input: { descriptionEn: 'Only English.' } });

    await waitFor(() => expect(result.current.isError).toBe(true));
    const error = result.current.error as ApiError;
    expect(error.status).toBe(400);
    expect(error.details).toEqual({
      descriptionEn: 'Falta la versión en español de la descripción corta.',
    });
  });

  it('ocultar y mostrar mandan solo `visibility`', async () => {
    const cuerpos: unknown[] = [];
    server.use(
      http.patch(`${adminUrl}/${itemId}`, async ({ request }) => {
        const cuerpo = await request.json();
        cuerpos.push(cuerpo);
        return HttpResponse.json({
          ...medio,
          visibility: (cuerpo as { visibility: string }).visibility,
        });
      }),
      http.get(adminUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useOcultarMostrarEpisodio(), { wrapper });

    result.current.mutate({ id: itemId, visibility: 'hidden' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.visibility).toBe('hidden');

    result.current.mutate({ id: itemId, visibility: 'visible' });
    await waitFor(() => expect(result.current.data?.visibility).toBe('visible'));
    expect(cuerpos).toEqual([{ visibility: 'hidden' }, { visibility: 'visible' }]);
  });

  it('unir manda `otherId` y el `409` de estado inconsistente se propaga (FR-014)', async () => {
    let cuerpo: unknown;
    server.use(
      http.post(`${adminUrl}/${itemId}/unir`, async ({ request }) => {
        cuerpo = await request.json();
        return HttpResponse.json({ ...medio, origin: 'both' });
      }),
      http.get(adminUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useUnirDuplicados(), { wrapper });

    result.current.mutate({ id: itemId, otherId });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(cuerpo).toEqual({ otherId });

    server.use(
      http.post(`${adminUrl}/${itemId}/unir`, () =>
        HttpResponse.json(
          { error: { code: 'conflict', message: 'Estado inconsistente' } },
          { status: 409 },
        ),
      ),
    );

    result.current.mutate({ id: itemId, otherId });
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as ApiError).status).toBe(409);
  });

  it('resolver la marca manda `mark` y `visibility`; el `404` se propaga (FR-017)', async () => {
    let cuerpo: unknown;
    server.use(
      http.post(`${adminUrl}/${itemId}/marcas/resolver`, async ({ request }) => {
        cuerpo = await request.json();
        return HttpResponse.json({ ...medio, visibility: 'hidden' });
      }),
      http.get(adminUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useResolverMarca(), { wrapper });

    result.current.mutate({ id: itemId, input: { mark: 'unavailable', visibility: 'hidden' } });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(cuerpo).toEqual({ mark: 'unavailable', visibility: 'hidden' });

    server.use(
      http.post(`${adminUrl}/${itemId}/marcas/resolver`, () =>
        HttpResponse.json(
          { error: { code: 'not_found', message: 'Recurso no encontrado' } },
          { status: 404 },
        ),
      ),
    );

    result.current.mutate({
      id: itemId,
      input: { mark: 'missing_from_feed', visibility: 'visible' },
    });
    await waitFor(() => expect(result.current.isError).toBe(true));
    const error = result.current.error as ApiError;
    expect(error.status).toBe(404);
    expect(error.code).toBe('not_found');
  });
});
