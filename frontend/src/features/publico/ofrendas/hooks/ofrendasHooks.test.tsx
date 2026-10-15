import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL, ApiError } from '../../../../api/client';
import type { OfrendasPublicas, PortadaOfrendas } from '../../../../api/ofrendas';
import { createQueryClientForTest } from '../../../../test/queryClient';
import { server } from '../../../../test/server';
import { LanguageProvider } from '../../i18n/LanguageProvider';
import { hasPublicOfrendasContent, usePublicOfrendas } from './usePublicOfrendas';
import { hasPortadaOfrendasContent, usePortadaOfrendas } from './usePortadaOfrendas';

const seccionUrl = `${API_BASE_URL}/api/v1/ofrendas`;
const portadaUrl = `${API_BASE_URL}/api/v1/portada/ofrendas`;
const itemId = '11111111-1111-1111-1111-111111111111';

const seccion: OfrendasPublicas = {
  lang: 'es',
  items: [
    {
      id: itemId,
      kind: 'iban',
      number: 'CR05015202001234567890',
      institution: 'Banco Nacional de Costa Rica',
      currency: 'CRC',
      holder: 'Simiente Santa',
    },
  ],
  intro: { text: 'Tu aporte va directo a la obra.' },
  usage: { text: 'Sostenemos la obra social y las misiones.' },
};

const bloque: PortadaOfrendas = {
  lang: 'es',
  items: [
    {
      id: itemId,
      kind: 'sinpe',
      number: '+50688888888',
      institution: 'Banco Promerica',
    },
  ],
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

describe('usePublicOfrendas (sección pública, FR-002)', () => {
  it('consulta la sección con el idioma activo y devuelve lo publicado', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(seccionUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(seccion);
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePublicOfrendas(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items).toHaveLength(1);
    expect(result.current.data?.intro?.text).toBe('Tu aporte va directo a la obra.');
    expect(consultada).toBe(`${seccionUrl}?lang=es`);
  });

  it('con 0 publicados expone el documento solo con `lang` (claves omitidas)', async () => {
    server.use(http.get(seccionUrl, () => HttpResponse.json({ lang: 'es' })));

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePublicOfrendas(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items).toBeUndefined();
    expect(result.current.data?.intro).toBeUndefined();
    expect(result.current.data?.usage).toBeUndefined();
  });

  it('ante un error conserva la última respuesta válida (error con caché, ux.md §4.1)', async () => {
    let calls = 0;
    server.use(
      http.get(seccionUrl, () => {
        calls += 1;
        if (calls === 1) {
          return HttpResponse.json(seccion);
        }
        return HttpResponse.json(
          { error: { code: 'internal', message: 'Falla del servidor' } },
          { status: 500 },
        );
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePublicOfrendas(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    await result.current.refetch();

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.data?.items).toHaveLength(1);
  });

  it('propaga el `400` del idioma inválido', async () => {
    server.use(
      http.get(seccionUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Idioma no soportado',
              details: { lang: 'Idioma no soportado' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePublicOfrendas(), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
    const error = result.current.error as ApiError;
    expect(error.status).toBe(400);
    expect(error.details).toEqual({ lang: 'Idioma no soportado' });
  });
});

describe('hasPublicOfrendasContent (P7-8/P7-13 · FR-005)', () => {
  it('la sección vive con métodos o con textos publicados', () => {
    expect(hasPublicOfrendasContent(undefined)).toBe(false);
    expect(hasPublicOfrendasContent({ lang: 'es' })).toBe(false);
    expect(hasPublicOfrendasContent({ lang: 'es', items: [] })).toBe(false);
    expect(hasPublicOfrendasContent({ lang: 'es', intro: { text: 'hola' } })).toBe(true);
    expect(hasPublicOfrendasContent({ lang: 'es', usage: { text: 'hola' } })).toBe(true);
    expect(
      hasPublicOfrendasContent({
        lang: 'es',
        items: [{ id: itemId, kind: 'iban', number: 'CR05015202001234567890', institution: 'x' }],
      }),
    ).toBe(true);
  });
});

describe('usePortadaOfrendas (bloque de la portada, FR-022)', () => {
  it('consulta el bloque con el idioma activo', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(portadaUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(bloque);
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePortadaOfrendas(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items).toHaveLength(1);
    expect(consultada).toBe(`${portadaUrl}?lang=es`);
  });

  it('con 0 métodos publicados expone el documento sin `items`', async () => {
    server.use(http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })));

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePortadaOfrendas(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items).toBeUndefined();
  });
});

describe('hasPortadaOfrendasContent (FR-022 · solo métodos)', () => {
  it('solo cuenta métodos publicados', () => {
    expect(hasPortadaOfrendasContent(undefined)).toBe(false);
    expect(hasPortadaOfrendasContent({ lang: 'es' })).toBe(false);
    expect(hasPortadaOfrendasContent({ lang: 'es', items: [] })).toBe(false);
    expect(
      hasPortadaOfrendasContent({
        lang: 'es',
        items: [{ id: itemId, kind: 'sinpe', number: '+50688888888', institution: 'x' }],
      }),
    ).toBe(true);
  });
});
