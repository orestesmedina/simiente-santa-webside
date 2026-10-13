import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { API_BASE_URL, ApiError } from '../../../../api/client';
import type { PortadaProximamente } from '../../../../api/agenda';
import { server } from '../../../../test/server';
import { createQueryClientForTest } from '../../../../test/queryClient';
import { LanguageProvider } from '../../i18n/LanguageProvider';
import { useLanguage } from '../../i18n/useLanguage';
import { usePortadaProximamente } from './usePortadaProximamente';

const url = `${API_BASE_URL}/api/v1/portada/proximamente`;

const enEspanol: PortadaProximamente = {
  lang: 'es',
  items: [
    {
      kind: 'event',
      id: 'ev-1',
      title: 'Vigilia de viernes',
      date: '2026-10-17',
      startTime: '19:00',
      endTime: '21:00',
      place: 'Templo central',
      image: { url: '/api/v1/media/img_vigilia.png', alt: 'Templo iluminado' },
    },
    {
      kind: 'activity',
      id: 'ac-1',
      title: 'Repartir comida',
      date: '2026-10-18',
      place: 'Comedor comunal',
    },
  ],
};

/** Permite cambiar de idioma desde la prueba (como el selector de F10). */
function LanguageProbe() {
  const { lang, setLang } = useLanguage();
  return (
    <button type="button" onClick={() => setLang(lang === 'es' ? 'en' : 'es')}>
      cambiar idioma
    </button>
  );
}

function makeWrapper() {
  const queryClient = createQueryClientForTest();
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <LanguageProvider>
        <LanguageProbe />
        {children}
      </LanguageProvider>
    </QueryClientProvider>
  );
  return { queryClient, wrapper };
}

describe('usePortadaProximamente (US7/FR-025)', () => {
  // El idioma elegido se recuerda en `localStorage` (F3/F10): sin limpiarlo, la
  // prueba de cambio de idioma dejaría el español de partida en inglés.
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('consulta el extracto con el idioma activo y devuelve los ítems resueltos', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(url, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(enEspanol);
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePortadaProximamente(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(consultada).toBe(`${url}?lang=es`);
    expect(result.current.data?.items?.map((item) => item.title)).toEqual([
      'Vigilia de viernes',
      'Repartir comida',
    ]);
    // El cliente no filtra ni reordena: presenta lo que el servidor ya resolvió.
    expect(result.current.data?.items?.[0]?.kind).toBe('event');
    expect(result.current.data?.items?.[1]?.kind).toBe('activity');
  });

  it('una respuesta sin ítems deja la lista vacía (clave omitida, nunca null)', async () => {
    server.use(http.get(url, () => HttpResponse.json({ lang: 'es' })));

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePortadaProximamente(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({ lang: 'es' });
    expect(result.current.data?.items).toBeUndefined();
  });

  it('propaga el error del servidor sin inventar contenido', async () => {
    server.use(
      http.get(url, () =>
        HttpResponse.json({ error: { code: 'internal', message: 'Error' } }, { status: 500 }),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePortadaProximamente(), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as ApiError).status).toBe(500);
    expect(result.current.data).toBeUndefined();
  });

  it('retiene la última respuesta válida cuando una consulta posterior falla', async () => {
    let peticiones = 0;
    server.use(
      http.get(url, () => {
        peticiones += 1;
        if (peticiones === 1) {
          return HttpResponse.json(enEspanol);
        }
        return HttpResponse.json(
          { error: { code: 'internal', message: 'Error' } },
          { status: 500 },
        );
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePortadaProximamente(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    // El cambio de idioma dispara una consulta nueva que falla: el bloque no se
    // queda vacío, conserva lo último válido (patrón de F3/F10).
    await userEvent.click(screen.getByRole('button', { name: 'cambiar idioma' }));

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.data?.items?.[0]?.title).toBe('Vigilia de viernes');
  });

  it('al cambiar de idioma vuelve a consultar con lang=en y usa la respuesta nueva', async () => {
    const urls: string[] = [];
    server.use(
      http.get(url, ({ request }) => {
        urls.push(request.url);
        const lang = new URL(request.url).searchParams.get('lang');
        return HttpResponse.json({
          lang,
          items: [
            {
              kind: 'event',
              id: 'ev-1',
              title: lang === 'en' ? 'Friday vigil' : 'Vigilia de viernes',
              date: '2026-10-17',
              place: 'Templo central',
            },
          ],
        });
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => usePortadaProximamente(), { wrapper });

    await waitFor(() => expect(result.current.data?.items?.[0]?.title).toBe('Vigilia de viernes'));

    await userEvent.click(screen.getByRole('button', { name: 'cambiar idioma' }));

    await waitFor(() => expect(result.current.data?.lang).toBe('en'));
    expect(result.current.data?.items?.[0]?.title).toBe('Friday vigil');
    expect(urls).toContain(`${url}?lang=en`);
  });
});
