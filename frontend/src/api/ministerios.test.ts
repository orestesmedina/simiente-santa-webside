import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { server } from '../test/server';
import { API_BASE_URL, ApiError } from './client';
import {
  MINISTRIES_BASE,
  createMinisterio,
  deleteMinisterio,
  getMinisterios,
  getPortadaMinisterios,
  listMinisterios,
  updateMinisterio,
  uploadMinistryImage,
  type MinisterioAdmin,
  type MinisterioPublico,
  type MinisteriosPublicos,
  type PortadaMinisterios,
} from './ministerios';

/**
 * Cliente del módulo «Ministerios» (F6, contrato 0.7.0): peticiones, sobre de
 * error y `details` por campo. Ningún componente llama a `fetch`: todo pasa por
 * aquí y los tipos salen **solo** del contrato vivo (`schema.d.ts`).
 */

const publicUrl = `${API_BASE_URL}/api/v1/ministerios`;
const portadaUrl = `${API_BASE_URL}/api/v1/portada/ministerios`;
const adminUrl = `${API_BASE_URL}${MINISTRIES_BASE}`;
const itemId = '11111111-1111-1111-1111-111111111111';

const ministerioPublico: MinisterioPublico = {
  id: itemId,
  name: 'Alabanza',
  description: 'Llevamos la adoración en los cultos.',
  schedule: 'Todos los viernes, 7:00 p. m.',
  place: 'Templo central',
  leader: { name: 'Josué Ramírez', phone: '+504 9999-9999' },
  whatsappUrl: 'https://chat.whatsapp.com/abc',
  image: { url: '/api/v1/media/img_x.png', alt: 'Equipo de alabanza' },
};

const ministerioAdmin: MinisterioAdmin = {
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
  scheduleEs: 'Todos los viernes, 7:00 p. m.',
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

describe('getMinisterios (público)', () => {
  it('consulta la sección pública con el idioma pedido', async () => {
    let consultada: string | undefined;
    const documento: MinisteriosPublicos = { lang: 'es', items: [ministerioPublico] };
    server.use(
      http.get(publicUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(documento);
      }),
    );

    await expect(getMinisterios('es')).resolves.toEqual(documento);
    expect(consultada).toBe(`${publicUrl}?lang=es`);
  });

  it('acepta el documento sin `items` (clave omitida con 0 publicados)', async () => {
    server.use(http.get(publicUrl, () => HttpResponse.json({ lang: 'en' })));

    const documento = await getMinisterios('en');
    expect(documento.items).toBeUndefined();
  });

  it('propaga el ApiError del sobre con su `details` (lang inválido)', async () => {
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

    const error = (await getMinisterios('es').catch((cause: unknown) => cause)) as ApiError;
    expect(error.code).toBe('invalid');
    expect(error.status).toBe(400);
    expect(error.details).toEqual({ lang: 'es|en' });
  });
});

describe('getPortadaMinisterios (bloque de la portada)', () => {
  it('consulta el bloque con el idioma pedido y admite `items` omitida', async () => {
    let consultada: string | undefined;
    const documento: PortadaMinisterios = {
      lang: 'en',
      items: [{ id: itemId, name: 'Worship', schedule: 'Fridays, 7:00 p. m.' }],
    };
    server.use(
      http.get(portadaUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(documento);
      }),
    );

    await expect(getPortadaMinisterios('en')).resolves.toEqual(documento);
    expect(consultada).toBe(`${portadaUrl}?lang=en`);

    server.use(http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })));
    await expect(getPortadaMinisterios('es')).resolves.toEqual({ lang: 'es' });
  });
});

describe('ministerios del panel', () => {
  it('listMinisterios envía la paginación y el filtro solo cuando existen', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(adminUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ items: [ministerioAdmin], total: 1, limit: 20, offset: 0 });
      }),
    );

    const listado = await listMinisterios({ limit: 20, offset: 0, publicationState: 'published' });
    expect(listado.items).toHaveLength(1);
    expect(consultada).toBe(`${adminUrl}?limit=20&offset=0&publicationState=published`);

    await listMinisterios();
    expect(consultada).toBe(adminUrl);
  });

  it('createMinisterio hace POST del alta con `leaderVisible` false', async () => {
    let cuerpo: unknown;
    server.use(
      http.post(adminUrl, async ({ request }) => {
        cuerpo = await request.json();
        return HttpResponse.json(ministerioAdmin, { status: 201 });
      }),
    );

    const input = {
      nameEs: 'Alabanza',
      descriptionEs: 'Llevamos la adoración en los cultos.',
      leaderNameEs: 'Josué Ramírez',
      leaderVisible: false,
    };
    await expect(createMinisterio(input)).resolves.toEqual(ministerioAdmin);
    expect(cuerpo).toEqual(input);
  });

  it('conserva el `409` de duplicado exacto del módulo', async () => {
    server.use(
      http.post(adminUrl, () =>
        HttpResponse.json(
          { error: { code: 'conflict', message: 'Ya existe un ministerio igual' } },
          { status: 409 },
        ),
      ),
    );

    const error = (await createMinisterio({
      nameEs: 'Alabanza',
      descriptionEs: 'x',
      leaderVisible: false,
    }).catch((cause: unknown) => cause)) as ApiError;
    expect(error.status).toBe(409);
    expect(error.code).toBe('conflict');
  });

  it('updateMinisterio hace PATCH parcial sobre el id', async () => {
    let consultada: string | undefined;
    let metodo: string | undefined;
    let cuerpo: unknown;
    server.use(
      http.patch(`${adminUrl}/${itemId}`, async ({ request }) => {
        consultada = request.url;
        metodo = request.method;
        cuerpo = await request.json();
        return HttpResponse.json({ ...ministerioAdmin, leaderVisible: true });
      }),
    );

    const actualizado = await updateMinisterio(itemId, { leaderVisible: true });
    expect(consultada).toBe(`${adminUrl}/${itemId}`);
    expect(metodo).toBe('PATCH');
    expect(cuerpo).toEqual({ leaderVisible: true });
    expect(actualizado.leaderVisible).toBe(true);
  });

  it('propaga el `404` de un id inexistente', async () => {
    server.use(
      http.patch(`${adminUrl}/${itemId}`, () =>
        HttpResponse.json({ error: { code: 'not_found', message: 'No existe' } }, { status: 404 }),
      ),
    );

    const error = (await updateMinisterio(itemId, { sortOrder: 1 }).catch(
      (cause: unknown) => cause,
    )) as ApiError;
    expect(error.status).toBe(404);
  });

  it('deleteMinisterio hace DELETE y resuelve sin cuerpo', async () => {
    let metodo: string | undefined;
    server.use(
      http.delete(`${adminUrl}/${itemId}`, ({ request }) => {
        metodo = request.method;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    await expect(deleteMinisterio(itemId)).resolves.toBeUndefined();
    expect(metodo).toBe('DELETE');
  });

  it('uploadMinistryImage envía multipart con el campo `file`', async () => {
    const visto: { cuerpo?: string } = {};
    server.use(
      http.post(`${adminUrl}/imagenes`, async ({ request }) => {
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

    const resultado = await uploadMinistryImage(new File(['x'], 'foto.png', { type: 'image/png' }));
    expect(visto.cuerpo).toContain('name="file"');
    expect(resultado.fileName).toBe('img_aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.png');
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

    const error = (await listMinisterios().catch((cause: unknown) => cause)) as ApiError;
    expect(error.status).toBe(403);
    expect(error.code).toBe('forbidden');
  });
});
