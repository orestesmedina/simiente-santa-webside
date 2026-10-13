import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL, ApiError } from '../../../../api/client';
import type { MinisteriosPublicos, PortadaMinisterios } from '../../../../api/ministerios';
import { createQueryClientForTest } from '../../../../test/queryClient';
import { server } from '../../../../test/server';
import { LanguageProvider } from '../../i18n/LanguageProvider';
import { usePublicMinistries } from './usePublicMinistries';
import { hasPortadaMinistryContent, usePortadaMinisterios } from './usePortadaMinisterios';

const publicUrl = `${API_BASE_URL}/api/v1/ministerios`;
const portadaUrl = `${API_BASE_URL}/api/v1/portada/ministerios`;
const itemId = '11111111-1111-1111-1111-111111111111';

const seccion: MinisteriosPublicos = {
  lang: 'es',
  items: [{ id: itemId, name: 'Alabanza', description: 'Adoración en los cultos.' }],
};

const bloque: PortadaMinisterios = {
  lang: 'es',
  items: [{ id: itemId, name: 'Alabanza', schedule: 'Todos los viernes' }],
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

describe('usePublicMinistries (sección pública, FR-002)', () => {
  it('consulta la sección con el idioma activo y devuelve los publicados', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(publicUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(seccion);
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePublicMinistries(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items).toHaveLength(1);
    expect(consultada).toBe(`${publicUrl}?lang=es`);
  });

  it('con 0 publicados expone el documento sin `items` (clave omitida)', async () => {
    server.use(http.get(publicUrl, () => HttpResponse.json({ lang: 'es' })));

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePublicMinistries(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items).toBeUndefined();
  });

  it('ante un error conserva la última respuesta válida (error con caché, ux.md §4.1)', async () => {
    const respuestas: Array<MinisteriosPublicos | { error: unknown }> = [];
    server.use(
      http.get(publicUrl, () => {
        if (respuestas.length === 0) {
          respuestas.push(seccion);
          return HttpResponse.json(seccion);
        }
        return HttpResponse.json(
          { error: { code: 'internal', message: 'Falla del servidor' } },
          { status: 500 },
        );
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePublicMinistries(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    await result.current.refetch();

    await waitFor(() => expect(result.current.isError).toBe(true));
    // El contenido previo no desaparece: la página muestra aviso discreto.
    expect(result.current.data?.items).toHaveLength(1);
  });

  it('propaga el `400` del idioma inválido', async () => {
    server.use(
      http.get(publicUrl, () =>
        HttpResponse.json(
          { error: { code: 'invalid', message: 'Idioma no soportado' } },
          { status: 400 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePublicMinistries(), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as ApiError).status).toBe(400);
  });
});

describe('usePortadaMinisterios (bloque de la portada, FR-027)', () => {
  it('consulta el bloque con el idioma activo', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(portadaUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(bloque);
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePortadaMinisterios(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items).toHaveLength(1);
    expect(consultada).toBe(`${portadaUrl}?lang=es`);
  });

  it('con 0 publicados expone el documento sin `items`', async () => {
    server.use(http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })));

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePortadaMinisterios(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items).toBeUndefined();
  });
});

describe('hasPortadaMinistryContent (FR-005/SC-010)', () => {
  it('distingue «sin publicados» de «con contenido»', () => {
    expect(hasPortadaMinistryContent(undefined)).toBe(false);
    expect(hasPortadaMinistryContent({ lang: 'es' })).toBe(false);
    expect(hasPortadaMinistryContent({ lang: 'es', items: [] })).toBe(false);
    expect(
      hasPortadaMinistryContent({ lang: 'es', items: [{ id: itemId, name: 'Alabanza' }] }),
    ).toBe(true);
  });
});
