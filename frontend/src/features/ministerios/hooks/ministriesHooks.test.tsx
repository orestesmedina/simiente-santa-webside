import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { API_BASE_URL, ApiError } from '../../../api/client';
import type { MinisterioAdmin } from '../../../api/ministerios';
import { createQueryClientForTest } from '../../../test/queryClient';
import { server } from '../../../test/server';
import { PORTADA_MINISTERIOS_QUERY_KEY } from '../../publico/ministerios/hooks/usePortadaMinisterios';
import { PUBLIC_MINISTRIES_QUERY_KEY } from '../../publico/ministerios/hooks/usePublicMinistries';
import { useCreateMinistry } from './useCreateMinistry';
import { useDeleteMinistry } from './useDeleteMinistry';
import { MINISTRIES_ADMIN_QUERY_KEY, MINISTRIES_PAGE_LIMIT, useMinistries } from './useMinistries';
import { useUpdateMinistry } from './useUpdateMinistry';
import { useUploadMinistryImage } from './useUploadMinistryImage';

const adminUrl = `${API_BASE_URL}/api/v1/admin/ministerios`;
const itemId = '11111111-1111-1111-1111-111111111111';

const ministerio: MinisterioAdmin = {
  id: itemId,
  nameEs: 'Alabanza',
  nameEn: null,
  descriptionEs: 'Llevamos la adoración en los cultos.',
  descriptionEn: null,
  leaderNameEs: 'Josué Ramírez',
  leaderNameEn: null,
  leaderPhone: '+504 9999-9999',
  leaderEmail: null,
  leaderVisible: false,
  whatsappUrlShared: false,
  scheduleEs: 'Todos los viernes',
  scheduleEn: null,
  placeEs: 'Templo central',
  placeEn: null,
  imageFile: null,
  imageAltEs: null,
  imageAltEn: null,
  publicationState: 'draft',
  sortOrder: 0,
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

describe('useMinistries (listado del panel, FR-010)', () => {
  it('pide el tope del contrato sin paginar (ux.md §4.3)', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(adminUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [ministerio], total: 1, limit: 100, offset: 0 });
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useMinistries(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(consultada).toBe(`${adminUrl}?limit=${MINISTRIES_PAGE_LIMIT}`);
    expect(result.current.data?.items[0]?.nameEs).toBe('Alabanza');
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
    const { result } = renderHook(() => useMinistries({ publicationState: 'published' }), {
      wrapper,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(consultada).toBe(
      `${adminUrl}?limit=${MINISTRIES_PAGE_LIMIT}&publicationState=published`,
    );
  });

  it('propaga el `403` del servidor sin el permiso `ministerios`', async () => {
    server.use(
      http.get(adminUrl, () =>
        HttpResponse.json(
          { error: { code: 'forbidden', message: 'No tienes acceso a esta sección' } },
          { status: 403 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useMinistries(), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as ApiError).status).toBe(403);
  });
});

describe('invalidación de caché tras cada mutación (SC-003)', () => {
  it('el alta invalida el listado del panel, la sección y el bloque públicos', async () => {
    server.use(
      http.post(adminUrl, () => HttpResponse.json(ministerio, { status: 201 })),
      http.get(adminUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
    );

    const { queryClient, wrapper } = makeWrapper();
    const spy = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useCreateMinistry(), { wrapper });

    result.current.mutate({ nameEs: 'Alabanza', descriptionEs: 'x', leaderVisible: false });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith({ queryKey: MINISTRIES_ADMIN_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: PUBLIC_MINISTRIES_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: PORTADA_MINISTERIOS_QUERY_KEY });
  });

  it('editar y eliminar refrescan también las superficies públicas', async () => {
    server.use(
      http.patch(`${adminUrl}/${itemId}`, () => HttpResponse.json(ministerio)),
      http.delete(`${adminUrl}/${itemId}`, () => new HttpResponse(null, { status: 204 })),
    );

    const { queryClient, wrapper } = makeWrapper();
    const spy = vi.spyOn(queryClient, 'invalidateQueries');

    const update = renderHook(() => useUpdateMinistry(), { wrapper });
    update.result.current.mutate({ id: itemId, input: { leaderVisible: true } });
    await waitFor(() => expect(update.result.current.isSuccess).toBe(true));

    const remove = renderHook(() => useDeleteMinistry(), { wrapper });
    remove.result.current.mutate(itemId);
    await waitFor(() => expect(remove.result.current.isSuccess).toBe(true));

    expect(spy).toHaveBeenCalledWith({ queryKey: MINISTRIES_ADMIN_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: PUBLIC_MINISTRIES_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: PORTADA_MINISTERIOS_QUERY_KEY });
  });
});

describe('useUploadMinistryImage (FR-020)', () => {
  it('sube el archivo y devuelve el nombre generado', async () => {
    server.use(
      http.post(`${adminUrl}/imagenes`, () =>
        HttpResponse.json(
          {
            fileName: 'img_aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.webp',
            url: '/api/v1/media/img_aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.webp',
            mimeType: 'image/webp',
            sizeBytes: 10,
          },
          { status: 201 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useUploadMinistryImage(), { wrapper });

    result.current.mutate(new File(['x'], 'foto.webp', { type: 'image/webp' }));

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.fileName).toBe('img_aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.webp');
  });

  it('expone el error de tipo/tamaño del servidor con su `details`', async () => {
    server.use(
      http.post(`${adminUrl}/imagenes`, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Archivo no válido',
              details: { file: 'Usa JPG, PNG o WebP.' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useUploadMinistryImage(), { wrapper });

    result.current.mutate(new File(['x'], 'nota.txt', { type: 'text/plain' }));

    await waitFor(() => expect(result.current.isError).toBe(true));
    const error = result.current.error as ApiError;
    expect(error.status).toBe(400);
    expect(error.details).toEqual({ file: 'Usa JPG, PNG o WebP.' });
  });
});
