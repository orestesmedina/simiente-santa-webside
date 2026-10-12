import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { server } from '../test/server';
import {
  ACTIVITIES_BASE,
  CATEGORIES_BASE,
  EVENTS_BASE,
  createActividad,
  createCategoria,
  createEvento,
  deleteActividad,
  deleteCategoria,
  deleteEvento,
  getProximamente,
  listActividades,
  listCategorias,
  listEventos,
  updateActividad,
  updateCategoria,
  updateEvento,
  uploadActividadImage,
  uploadEventoImage,
  type ActividadAdmin,
  type CategoriaAdmin,
  type EventoAdmin,
  type ProximamentePublica,
} from './agenda';
import { API_BASE_URL, ApiError } from './client';

const publicUrl = `${API_BASE_URL}/api/v1/proximamente`;
const eventsUrl = `${API_BASE_URL}${EVENTS_BASE}`;
const activitiesUrl = `${API_BASE_URL}${ACTIVITIES_BASE}`;
const categoriesUrl = `${API_BASE_URL}${CATEGORIES_BASE}`;
const itemId = '11111111-1111-1111-1111-111111111111';

const proximamente: ProximamentePublica = {
  lang: 'es',
  events: [
    {
      id: itemId,
      title: 'Vigilia de viernes',
      frequency: 'Todos los viernes',
      dates: [{ date: '2026-10-17', startTime: '19:00', endTime: '21:00' }],
      place: 'Templo central',
      mapUrl: 'https://maps.google.com/?q=templo',
      detail: 'Noche de oración',
      category: { id: itemId, name: 'noche de oración' },
      image: { url: '/api/v1/media/img_x.png', alt: 'Templo iluminado' },
    },
  ],
  activities: [
    {
      id: itemId,
      title: 'Repartir comida',
      date: '2026-10-18',
      startTime: '08:00',
      endTime: '11:00',
      place: 'Comedor comunal',
      detail: 'Acción social',
    },
  ],
};

const evento: EventoAdmin = {
  id: itemId,
  titleEs: 'Vigilia',
  titleEn: null,
  frequencyEs: 'Todos los viernes',
  frequencyEn: null,
  detailEs: 'Noche de oración',
  detailEn: null,
  placeEs: 'Templo central',
  placeEn: null,
  mapUrl: null,
  organizerEs: null,
  organizerEn: null,
  categoryId: null,
  imageFile: null,
  imageAltEs: null,
  imageAltEn: null,
  publicationState: 'draft',
  dates: [{ id: itemId, date: '2026-10-17', startTime: '19:00', isPast: false }],
  nextDate: '2026-10-17',
  isPast: false,
  createdAt: '2026-10-01T10:00:00Z',
  updatedAt: '2026-10-01T10:00:00Z',
};

const actividad: ActividadAdmin = {
  id: itemId,
  titleEs: 'Repartir comida',
  titleEn: null,
  detailEs: 'Acción social',
  detailEn: null,
  placeEs: 'Comedor comunal',
  placeEn: null,
  mapUrl: null,
  organizerEs: null,
  organizerEn: null,
  categoryId: null,
  imageFile: null,
  imageAltEs: null,
  imageAltEn: null,
  date: '2026-10-18',
  startTime: '08:00',
  publicationState: 'draft',
  isPast: false,
  createdAt: '2026-10-01T10:00:00Z',
  updatedAt: '2026-10-01T10:00:00Z',
};

const categoria: CategoriaAdmin = {
  id: itemId,
  nameEs: 'acción social',
  nameEn: null,
  sortOrder: 0,
  createdAt: '2026-10-01T10:00:00Z',
  updatedAt: '2026-10-01T10:00:00Z',
};

describe('getProximamente', () => {
  it('consulta el documento público con el idioma pedido', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(publicUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(proximamente);
      }),
    );

    await expect(getProximamente('en')).resolves.toEqual(proximamente);
    expect(consultada).toBe(`${publicUrl}?lang=en`);
  });

  it('propaga el ApiError del sobre con su `details`', async () => {
    server.use(
      http.get(publicUrl, () =>
        HttpResponse.json(
          {
            error: { code: 'invalid', message: 'Idioma no soportado', details: { lang: 'es|en' } },
          },
          { status: 400 },
        ),
      ),
    );

    const error = (await getProximamente('es').catch((cause: unknown) => cause)) as ApiError;
    expect(error.code).toBe('invalid');
    expect(error.status).toBe(400);
    expect(error.details).toEqual({ lang: 'es|en' });
  });
});

describe('eventos del panel', () => {
  it('listEventos envía la paginación y el filtro de estado solo cuando existen', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(eventsUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [evento], total: 1, limit: 20, offset: 0 });
      }),
    );

    const listado = await listEventos({ limit: 20, offset: 0, publicationState: 'published' });
    expect(listado.items).toHaveLength(1);
    expect(consultada).toBe(`${eventsUrl}?limit=20&offset=0&publicationState=published`);

    await listEventos();
    expect(consultada).toBe(eventsUrl);
  });

  it('createEvento hace POST del alta con sus fechas', async () => {
    let cuerpo: unknown;
    server.use(
      http.post(eventsUrl, async ({ request }) => {
        cuerpo = await request.json();
        return HttpResponse.json(evento, { status: 201 });
      }),
    );

    const input = {
      titleEs: 'Vigilia',
      detailEs: 'Noche de oración',
      placeEs: 'Templo central',
      dates: [{ date: '2026-10-17', startTime: '19:00' }],
    };
    await expect(createEvento(input)).resolves.toEqual(evento);
    expect(cuerpo).toEqual(input);
  });

  it('conserva el `409` de duplicado del módulo', async () => {
    server.use(
      http.post(eventsUrl, () =>
        HttpResponse.json(
          { error: { code: 'conflict', message: 'Ya existe un evento igual' } },
          { status: 409 },
        ),
      ),
    );

    const error = (await createEvento({
      titleEs: 'Vigilia',
      detailEs: 'x',
      placeEs: 'Templo',
      dates: [{ date: '2026-10-17' }],
    }).catch((cause: unknown) => cause)) as ApiError;
    expect(error.status).toBe(409);
    expect(error.code).toBe('conflict');
  });

  it('updateEvento hace PATCH sobre el id', async () => {
    let consultada: string | undefined;
    let metodo: string | undefined;
    server.use(
      http.patch(`${eventsUrl}/${itemId}`, ({ request }) => {
        consultada = request.url;
        metodo = request.method;
        return HttpResponse.json({ ...evento, publicationState: 'published' });
      }),
    );

    const actualizado = await updateEvento(itemId, { publicationState: 'published' });
    expect(consultada).toBe(`${eventsUrl}/${itemId}`);
    expect(metodo).toBe('PATCH');
    expect(actualizado.publicationState).toBe('published');
  });

  it('deleteEvento hace DELETE y resuelve sin cuerpo', async () => {
    let metodo: string | undefined;
    server.use(
      http.delete(`${eventsUrl}/${itemId}`, ({ request }) => {
        metodo = request.method;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    await expect(deleteEvento(itemId)).resolves.toBeUndefined();
    expect(metodo).toBe('DELETE');
  });

  it('uploadEventoImage envía multipart con el campo `file`', async () => {
    const visto: { cuerpo?: string } = {};
    server.use(
      http.post(`${eventsUrl}/imagenes`, async ({ request }) => {
        visto.cuerpo = await request.text();
        return HttpResponse.json(
          {
            fileName: 'img_aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.png',
            url: '/api/v1/media/img_aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.png',
            mimeType: 'image/png',
            sizeBytes: 1234,
          },
          { status: 201 },
        );
      }),
    );

    const resultado = await uploadEventoImage(new File(['x'], 'foto.png', { type: 'image/png' }));
    expect(visto.cuerpo).toContain('name="file"');
    expect(resultado.fileName).toBe('img_aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.png');
  });
});

describe('actividades del panel', () => {
  it('listActividades consulta el listado paginado', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(activitiesUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [actividad], total: 1, limit: 20, offset: 0 });
      }),
    );

    const listado = await listActividades({ publicationState: 'draft' });
    expect(listado.items).toHaveLength(1);
    expect(consultada).toBe(`${activitiesUrl}?publicationState=draft`);
  });

  it('createActividad hace POST de la fecha única', async () => {
    let cuerpo: unknown;
    server.use(
      http.post(activitiesUrl, async ({ request }) => {
        cuerpo = await request.json();
        return HttpResponse.json(actividad, { status: 201 });
      }),
    );

    const input = {
      titleEs: 'Repartir comida',
      detailEs: 'Acción social',
      placeEs: 'Comedor comunal',
      date: '2026-10-18',
    };
    await createActividad(input);
    expect(cuerpo).toEqual(input);
  });

  it('updateActividad hace PATCH y deleteActividad DELETE', async () => {
    const vistos: string[] = [];
    server.use(
      http.patch(`${activitiesUrl}/${itemId}`, ({ request }) => {
        vistos.push(request.method);
        return HttpResponse.json(actividad);
      }),
      http.delete(`${activitiesUrl}/${itemId}`, ({ request }) => {
        vistos.push(request.method);
        return new HttpResponse(null, { status: 204 });
      }),
    );

    await updateActividad(itemId, { titleEs: 'Repartir' });
    await deleteActividad(itemId);
    expect(vistos).toEqual(['PATCH', 'DELETE']);
  });

  it('uploadActividadImage usa su propio endpoint de módulo', async () => {
    let consultada: string | undefined;
    server.use(
      http.post(`${activitiesUrl}/imagenes`, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(
          {
            fileName: 'img_x.png',
            url: '/api/v1/media/img_x.png',
            mimeType: 'image/png',
            sizeBytes: 10,
          },
          { status: 201 },
        );
      }),
    );

    await uploadActividadImage(new File(['x'], 'foto.png', { type: 'image/png' }));
    expect(consultada).toBe(`${activitiesUrl}/imagenes`);
  });
});

describe('catálogo de categorías', () => {
  it('listCategorias no pagina y devuelve el catálogo completo', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(categoriesUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [categoria] });
      }),
    );

    const listado = await listCategorias();
    expect(listado.items).toHaveLength(1);
    expect(consultada).toBe(categoriesUrl);
  });

  it('createCategoria y updateCategoria hablan con el catálogo', async () => {
    const vistos: string[] = [];
    let cuerpo: unknown;
    server.use(
      http.post(categoriesUrl, async ({ request }) => {
        vistos.push(request.method);
        cuerpo = await request.json();
        return HttpResponse.json(categoria, { status: 201 });
      }),
      http.patch(`${categoriesUrl}/${itemId}`, ({ request }) => {
        vistos.push(request.method);
        return HttpResponse.json({ ...categoria, nameEs: 'ayuda social' });
      }),
    );

    await createCategoria({ nameEs: 'acción social' });
    const editada = await updateCategoria(itemId, { nameEs: 'ayuda social' });
    expect(vistos).toEqual(['POST', 'PATCH']);
    expect(cuerpo).toEqual({ nameEs: 'acción social' });
    expect(editada.nameEs).toBe('ayuda social');
  });

  it('deleteCategoria propaga el `409` de categoría en uso con su mensaje', async () => {
    server.use(
      http.delete(`${categoriesUrl}/${itemId}`, () =>
        HttpResponse.json(
          {
            error: {
              code: 'conflict',
              message: 'La categoría está en uso por 4 elementos',
              details: { categoryId: 'en uso' },
            },
          },
          { status: 409 },
        ),
      ),
    );

    const error = (await deleteCategoria(itemId).catch((cause: unknown) => cause)) as ApiError;
    expect(error.status).toBe(409);
    expect(error.code).toBe('conflict');
    expect(error.message).toBe('La categoría está en uso por 4 elementos');
    expect(error.details).toEqual({ categoryId: 'en uso' });
  });

  it('deleteCategoria resuelve sin cuerpo cuando el borrado procede', async () => {
    server.use(
      http.delete(`${categoriesUrl}/${itemId}`, () => new HttpResponse(null, { status: 204 })),
    );

    await expect(deleteCategoria(itemId)).resolves.toBeUndefined();
  });
});
