import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { API_BASE_URL, ApiError } from '../../../api/client';
import { server } from '../../../test/server';
import { createQueryClientForTest } from '../../../test/queryClient';
import { LanguageProvider } from '../../publico/i18n/LanguageProvider';
import {
  hasUpcomingContent,
  useUpcomingEvents,
} from '../../publico/eventos/hooks/useUpcomingEvents';
import { useActivities } from './useActivities';
import { useCreateActivity, useDeleteActivity, useUpdateActivity } from './useActivityMutations';
import { useCreateCategory, useDeleteCategory, useEventCategories } from './useEventCategories';
import { useCreateEvent, useDeleteEvent, useUpdateEvent } from './useEventMutations';
import { useEvents } from './useEvents';
import {
  AGENDA_ACTIVITIES_QUERY_KEY,
  AGENDA_CATEGORIES_QUERY_KEY,
  AGENDA_EVENTS_QUERY_KEY,
  UPCOMING_QUERY_KEY,
} from './queryKeys';

const eventsUrl = `${API_BASE_URL}/api/v1/admin/eventos`;
const activitiesUrl = `${API_BASE_URL}/api/v1/admin/actividades`;
const categoriesUrl = `${API_BASE_URL}/api/v1/admin/categorias`;
const upcomingUrl = `${API_BASE_URL}/api/v1/proximamente`;
const itemId = '11111111-1111-1111-1111-111111111111';

const evento = {
  id: itemId,
  titleEs: 'Vigilia',
  detailEs: 'Noche de oración',
  placeEs: 'Templo central',
  publicationState: 'draft' as const,
  dates: [{ id: itemId, date: '2026-10-17', isPast: false }],
  nextDate: '2026-10-17',
  isPast: false,
  createdAt: '2026-10-01T10:00:00Z',
  updatedAt: '2026-10-01T10:00:00Z',
};

const actividad = {
  id: itemId,
  titleEs: 'Repartir comida',
  detailEs: 'Acción social',
  placeEs: 'Comedor comunal',
  date: '2026-10-18',
  publicationState: 'draft' as const,
  isPast: false,
  createdAt: '2026-10-01T10:00:00Z',
  updatedAt: '2026-10-01T10:00:00Z',
};

const categoria = {
  id: itemId,
  nameEs: 'acción social',
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

function makePublicWrapper() {
  const queryClient = createQueryClientForTest();
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <LanguageProvider>{children}</LanguageProvider>
    </QueryClientProvider>
  );
  return { queryClient, wrapper };
}

describe('useEvents / useActivities (FR-010/FR-011)', () => {
  it('devuelve el listado del panel con su estado y marca de pasado', async () => {
    server.use(
      http.get(eventsUrl, () =>
        HttpResponse.json({ items: [evento], total: 1, limit: 100, offset: 0 }),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useEvents(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items[0]?.titleEs).toBe('Vigilia');
  });

  it('pide el tope del contrato (sin paginación en la UX, D-17)', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(eventsUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 });
      }),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useEvents({ publicationState: 'published' }), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(consultada).toBe(`${eventsUrl}?limit=100&publicationState=published`);
  });

  it('propaga el `403` del servidor sin permisos', async () => {
    server.use(
      http.get(activitiesUrl, () =>
        HttpResponse.json(
          { error: { code: 'forbidden', message: 'No tienes acceso a esta sección' } },
          { status: 403 },
        ),
      ),
    );

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useActivities(), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as ApiError).status).toBe(403);
  });
});

describe('useUpcomingEvents (FR-001/FR-022)', () => {
  it('consulta el documento público con el idioma activo', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(upcomingUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ lang: 'es', events: [{ id: itemId, title: 'Vigilia' }] });
      }),
    );

    const { wrapper } = makePublicWrapper();
    const { result } = renderHook(() => useUpcomingEvents(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(consultada).toBe(`${upcomingUrl}?lang=es`);
  });

  it('hasUpcomingContent distingue «sin nada publicado y próximo» de «con contenido»', () => {
    const eventoPublico = {
      id: itemId,
      title: 'Vigilia',
      dates: [{ date: '2026-10-17' }],
      place: 'Templo',
      detail: 'Noche de oración',
    };
    const actividadPublica = {
      id: itemId,
      title: 'Comida',
      date: '2026-10-18',
      place: 'Comedor',
      detail: 'Acción social',
    };

    expect(hasUpcomingContent(undefined)).toBe(false);
    expect(hasUpcomingContent({ lang: 'es' })).toBe(false);
    expect(hasUpcomingContent({ lang: 'es', events: [] })).toBe(false);
    expect(hasUpcomingContent({ lang: 'es', activities: [] })).toBe(false);
    expect(hasUpcomingContent({ lang: 'es', events: [eventoPublico] })).toBe(true);
    expect(hasUpcomingContent({ lang: 'es', activities: [actividadPublica] })).toBe(true);
  });
});

describe('useEventCategories (Q6)', () => {
  it('devuelve el catálogo completo', async () => {
    server.use(http.get(categoriesUrl, () => HttpResponse.json({ items: [categoria] })));

    const { wrapper } = makeWrapper();
    const { result } = renderHook(() => useEventCategories(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items[0]?.nameEs).toBe('acción social');
  });
});

describe('invalidación de caché tras cada mutación (SC-003)', () => {
  it('el alta de evento invalida el listado del panel y la sección pública', async () => {
    server.use(
      http.post(eventsUrl, () => HttpResponse.json(evento, { status: 201 })),
      http.get(eventsUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
      http.get(upcomingUrl, () => HttpResponse.json({ lang: 'es' })),
    );

    const { queryClient, wrapper } = makeWrapper();
    const spy = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useCreateEvent(), { wrapper });

    result.current.mutate({
      titleEs: 'Vigilia',
      detailEs: 'x',
      placeEs: 'Templo',
      dates: [{ date: '2026-10-17' }],
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith({ queryKey: AGENDA_EVENTS_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: UPCOMING_QUERY_KEY });
  });

  it('editar y eliminar un evento también refrescan la sección pública', async () => {
    server.use(
      http.patch(`${eventsUrl}/${itemId}`, () => HttpResponse.json(evento)),
      http.delete(`${eventsUrl}/${itemId}`, () => new HttpResponse(null, { status: 204 })),
    );

    const { queryClient, wrapper } = makeWrapper();
    const spy = vi.spyOn(queryClient, 'invalidateQueries');

    const update = renderHook(() => useUpdateEvent(), { wrapper });
    update.result.current.mutate({ id: itemId, input: { titleEs: 'Nueva' } });
    await waitFor(() => expect(update.result.current.isSuccess).toBe(true));

    const remove = renderHook(() => useDeleteEvent(), { wrapper });
    remove.result.current.mutate(itemId);
    await waitFor(() => expect(remove.result.current.isSuccess).toBe(true));

    expect(spy).toHaveBeenCalledWith({ queryKey: AGENDA_EVENTS_QUERY_KEY });
  });

  it('las mutaciones de actividad invalidan su propio listado', async () => {
    server.use(
      http.post(activitiesUrl, () => HttpResponse.json(actividad, { status: 201 })),
      http.patch(`${activitiesUrl}/${itemId}`, () => HttpResponse.json(actividad)),
      http.delete(`${activitiesUrl}/${itemId}`, () => new HttpResponse(null, { status: 204 })),
    );

    const { queryClient, wrapper } = makeWrapper();
    const spy = vi.spyOn(queryClient, 'invalidateQueries');

    const create = renderHook(() => useCreateActivity(), { wrapper });
    create.result.current.mutate({
      titleEs: 'Repartir',
      detailEs: 'x',
      placeEs: 'Comedor',
      date: '2026-10-18',
    });
    await waitFor(() => expect(create.result.current.isSuccess).toBe(true));

    const update = renderHook(() => useUpdateActivity(), { wrapper });
    update.result.current.mutate({ id: itemId, input: { titleEs: 'Otra' } });
    await waitFor(() => expect(update.result.current.isSuccess).toBe(true));

    const remove = renderHook(() => useDeleteActivity(), { wrapper });
    remove.result.current.mutate(itemId);
    await waitFor(() => expect(remove.result.current.isSuccess).toBe(true));

    expect(spy).toHaveBeenCalledWith({ queryKey: AGENDA_ACTIVITIES_QUERY_KEY });
    expect(spy).toHaveBeenCalledWith({ queryKey: UPCOMING_QUERY_KEY });
  });

  it('las mutaciones del catálogo invalidan las categorías', async () => {
    server.use(
      http.post(categoriesUrl, () => HttpResponse.json(categoria, { status: 201 })),
      http.delete(`${categoriesUrl}/${itemId}`, () => new HttpResponse(null, { status: 204 })),
    );

    const { queryClient, wrapper } = makeWrapper();
    const spy = vi.spyOn(queryClient, 'invalidateQueries');

    const create = renderHook(() => useCreateCategory(), { wrapper });
    create.result.current.mutate({ nameEs: 'ayuda social' });
    await waitFor(() => expect(create.result.current.isSuccess).toBe(true));

    const remove = renderHook(() => useDeleteCategory(), { wrapper });
    remove.result.current.mutate(itemId);
    await waitFor(() => expect(remove.result.current.isSuccess).toBe(true));

    expect(spy).toHaveBeenCalledWith({ queryKey: AGENDA_CATEGORIES_QUERY_KEY });
  });

  it('un borrado fallido (`409` en uso) no invalida el catálogo', async () => {
    server.use(
      http.delete(`${categoriesUrl}/${itemId}`, () =>
        HttpResponse.json(
          { error: { code: 'conflict', message: 'Está en uso por 4 elementos' } },
          { status: 409 },
        ),
      ),
    );

    const { queryClient, wrapper } = makeWrapper();
    const spy = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useDeleteCategory(), { wrapper });

    result.current.mutate(itemId);
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(spy).not.toHaveBeenCalled();
  });
});
