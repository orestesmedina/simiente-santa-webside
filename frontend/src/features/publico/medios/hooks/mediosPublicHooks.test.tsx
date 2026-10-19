import { QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL, ApiError } from '../../../../api/client';
import type { MediosPublicos, PortadaMedios } from '../../../../api/medios';
import { createQueryClientForTest } from '../../../../test/queryClient';
import { server } from '../../../../test/server';
import { LanguageProvider } from '../../i18n/LanguageProvider';
import { hasPortadaMediosContent, usePortadaMedios } from './usePortadaMedios';
import { hasPublicMediosContent, usePublicMedios } from './usePublicMedios';

const publicUrl = `${API_BASE_URL}/api/v1/medios`;
const portadaUrl = `${API_BASE_URL}/api/v1/portada/medios`;
const itemId = '11111111-1111-1111-1111-111111111111';

function makeItem(index: number) {
  return {
    id: `1111111${index}-1111-1111-1111-111111111111`,
    title: `Sermón ${index}`,
    publishedAt: '2026-07-03T01:22:48Z',
    videoUrl: 'https://www.youtube.com/watch?v=abcdefghijk',
  };
}

function makeTanda(count: number, total: number, offset = 0): MediosPublicos {
  return {
    lang: 'es',
    items: Array.from({ length: count }, (_, i) => makeItem(offset + i)),
    total,
    limit: 12,
    offset,
  };
}

const bloque: PortadaMedios = {
  lang: 'es',
  items: [{ id: itemId, title: 'Potencial sin Resultados', publishedAt: '2026-07-03T01:22:48Z' }],
};

function makeWrapper() {
  const queryClient = createQueryClientForTest();
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <LanguageProvider>{children}</LanguageProvider>
    </QueryClientProvider>
  );
  return { queryClient, wrapper };
}

describe('usePublicMedios (sección pública, FR-002/FR-006)', () => {
  it('consulta la primera tanda con el idioma activo, 12 fichas y offset 0', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(publicUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(makeTanda(2, 2));
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePublicMedios(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(consultada).toBe(`${publicUrl}?lang=es&limit=12&offset=0`);
    expect(result.current.data?.items).toHaveLength(2);
    expect(result.current.hasMore).toBe(false);
  });

  it('«Cargar más» acumula la siguiente tanda y `hasMore` se agota con `total` (Q11/D-3)', async () => {
    const consultadas: string[] = [];
    server.use(
      http.get(publicUrl, ({ request }) => {
        consultadas.push(request.url);
        const limit = Number(new URL(request.url).searchParams.get('limit'));
        // Con 13 publicables: la primera tanda (12) deja una pendiente y la
        // segunda (24) ya las trae todas: `hasMore` pasa a `false` y la lista
        // acumula sin que el cliente reordene ni mezcle nada.
        return HttpResponse.json(limit < 13 ? makeTanda(12, 13) : makeTanda(13, 13));
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePublicMedios(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items).toHaveLength(12);
    expect(result.current.hasMore).toBe(true);

    act(() => result.current.loadMore());

    await waitFor(() => expect(result.current.data?.items).toHaveLength(13));
    expect(consultadas[1]).toBe(`${publicUrl}?lang=es&limit=24&offset=0`);
    expect(result.current.hasMore).toBe(false);
  });

  it('ante un error conserva la última respuesta válida (error con caché, ux.md §4.1)', async () => {
    let llamadas = 0;
    server.use(
      http.get(publicUrl, () => {
        llamadas += 1;
        if (llamadas === 1) {
          return HttpResponse.json(makeTanda(2, 2));
        }
        return HttpResponse.json(
          { error: { code: 'internal', message: 'Falla del servidor' } },
          { status: 500 },
        );
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePublicMedios(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    await result.current.refetch();

    await waitFor(() => expect(result.current.isError).toBe(true));
    // El contenido previo no desaparece: la página muestra el aviso discreto.
    expect(result.current.data?.items).toHaveLength(2);
  });

  it('propaga el `400` del servidor (idioma no soportado)', async () => {
    server.use(
      http.get(publicUrl, () =>
        HttpResponse.json(
          { error: { code: 'invalid', message: 'Idioma no soportado' } },
          { status: 400 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePublicMedios(), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as ApiError).status).toBe(400);
  });
});

describe('hasPublicMediosContent (FR-008/P9-16)', () => {
  it('la sección existe solo con `total > 0`', () => {
    expect(hasPublicMediosContent(undefined)).toBe(false);
    expect(hasPublicMediosContent({ lang: 'es', items: [], total: 0, limit: 12, offset: 0 })).toBe(
      false,
    );
    expect(hasPublicMediosContent(makeTanda(1, 1))).toBe(true);
  });
});

describe('usePortadaMedios (bloque de la portada, FR-027)', () => {
  it('consulta el bloque con el idioma activo y sin parámetro de límite', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(portadaUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(bloque);
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePortadaMedios(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(consultada).toBe(`${portadaUrl}?lang=es`);
    expect(result.current.data?.items).toHaveLength(1);
  });

  it('con 0 publicables expone el documento sin `items` (clave omitida)', async () => {
    server.use(http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })));

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePortadaMedios(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items).toBeUndefined();
  });
});

describe('hasPortadaMediosContent (SC-015)', () => {
  it('distingue «sin publicables» de «con fichas»', () => {
    expect(hasPortadaMediosContent(undefined)).toBe(false);
    expect(hasPortadaMediosContent({ lang: 'es' })).toBe(false);
    expect(hasPortadaMediosContent({ lang: 'es', items: [] })).toBe(false);
    expect(hasPortadaMediosContent(bloque)).toBe(true);
  });
});
