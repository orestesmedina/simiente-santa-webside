import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { server } from '../test/server';
import { API_BASE_URL, ApiError } from './client';
import {
  createService,
  createSocialLink,
  createWhatsappChannel,
  deleteService,
  deleteSocialLink,
  deleteWhatsappChannel,
  getPortada,
  getPortadaAdmin,
  updateAbout,
  updateContact,
  updateIdentity,
  updateService,
  updateSocialLink,
  updateWhatsappChannel,
  uploadImage,
  type AboutInput,
  type ContactInput,
  type IdentityInput,
  type PortadaAdmin,
  type PortadaPublica,
  type ScheduleItemAdmin,
  type SocialLinkAdmin,
  type WhatsappChannelAdmin,
} from './portada';

const publicUrl = `${API_BASE_URL}/api/v1/portada`;
const adminUrl = `${API_BASE_URL}/api/v1/admin/portada`;
const itemId = '11111111-1111-1111-1111-111111111111';

const productoPublico: PortadaPublica = {
  lang: 'es',
  identity: { name: 'Iglesia Simiente Santa' },
  schedule: [
    {
      id: itemId,
      dayOfWeek: 0,
      startTime: '10:00',
      endTime: '12:00',
      name: 'Culto dominical',
      place: 'Templo central',
    },
  ],
};

const identidadAdmin = {
  nameEs: 'Iglesia Simiente Santa',
  publicationState: 'draft' as const,
  createdAt: '2026-10-01T10:00:00Z',
  updatedAt: '2026-10-01T10:00:00Z',
};

const productoAdmin: PortadaAdmin = {
  identity: identidadAdmin,
  about: null,
  contact: null,
  schedule: { items: [] },
  whatsapp: { items: [] },
  socials: { items: [] },
};

describe('getPortada', () => {
  it('consulta la portada pública con el idioma pedido', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(publicUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(productoPublico);
      }),
    );

    await expect(getPortada('en')).resolves.toEqual(productoPublico);
    expect(consultada).toBe(`${publicUrl}?lang=en`);
  });

  it('propaga el ApiError del sobre con su `details`', async () => {
    server.use(
      http.get(publicUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Idioma no soportado',
              details: { lang: 'es|en' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    const error = (await getPortada('es').catch((cause: unknown) => cause)) as ApiError;
    expect(error.code).toBe('invalid');
    expect(error.status).toBe(400);
    expect(error.details).toEqual({ lang: 'es|en' });
  });
});

describe('getPortadaAdmin', () => {
  it('consulta el agregado del panel', async () => {
    let consultada: string | undefined;
    server.use(
      http.get(adminUrl, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(productoAdmin);
      }),
    );

    await expect(getPortadaAdmin()).resolves.toEqual(productoAdmin);
    expect(consultada).toBe(adminUrl);
  });
});

describe('singletons del panel', () => {
  it('updateIdentity hace PUT del reemplazo completo', async () => {
    let metodo: string | undefined;
    let cuerpo: unknown;
    server.use(
      http.put(`${adminUrl}/identidad`, async ({ request }) => {
        metodo = request.method;
        cuerpo = await request.json();
        return HttpResponse.json(identidadAdmin);
      }),
    );

    const input: IdentityInput = { nameEs: 'Iglesia Simiente Santa', publicationState: 'draft' };
    await expect(updateIdentity(input)).resolves.toEqual(identidadAdmin);
    expect(metodo).toBe('PUT');
    expect(cuerpo).toEqual(input);
  });

  it('updateAbout hace PUT en /quienes-somos', async () => {
    let consultada: string | undefined;
    server.use(
      http.put(`${adminUrl}/quienes-somos`, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ ...identidadAdmin, textEs: 'Somos una familia' });
      }),
    );

    const input: AboutInput = { textEs: 'Somos una familia', publicationState: 'published' };
    await updateAbout(input);
    expect(consultada).toBe(`${adminUrl}/quienes-somos`);
  });

  it('updateContact hace PUT del cuerpo tipado', async () => {
    let cuerpo: unknown;
    server.use(
      http.put(`${adminUrl}/contacto`, async ({ request }) => {
        cuerpo = await request.json();
        return HttpResponse.json({
          addressEs: 'Calle 1',
          email: 'hola@ejemplo.com',
          phone: '+50688888888',
          publicationState: 'published',
          createdAt: '2026-10-01T10:00:00Z',
          updatedAt: '2026-10-01T10:00:00Z',
        });
      }),
    );

    const input: ContactInput = {
      addressEs: 'Calle 1',
      email: 'hola@ejemplo.com',
      phone: '+50688888888',
      publicationState: 'published',
    };
    await updateContact(input);
    expect(cuerpo).toEqual(input);
  });
});

describe('horario', () => {
  const servicio: ScheduleItemAdmin = {
    id: itemId,
    dayOfWeek: 0,
    startTime: '10:00',
    endTime: null,
    nameEs: 'Culto dominical',
    placeEs: 'Templo central',
    publicationState: 'draft',
    sortOrder: 0,
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
  };

  it('createService hace POST del alta', async () => {
    let metodo: string | undefined;
    let cuerpo: unknown;
    server.use(
      http.post(`${adminUrl}/horario`, async ({ request }) => {
        metodo = request.method;
        cuerpo = await request.json();
        return HttpResponse.json(servicio, { status: 201 });
      }),
    );

    const input = {
      dayOfWeek: 0,
      startTime: '10:00',
      nameEs: 'Culto dominical',
      placeEs: 'Templo central',
    };
    await expect(createService(input)).resolves.toEqual(servicio);
    expect(metodo).toBe('POST');
    expect(cuerpo).toEqual(input);
  });

  it('updateService hace PATCH sobre el id', async () => {
    let consultada: string | undefined;
    server.use(
      http.patch(`${adminUrl}/horario/${itemId}`, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json({ ...servicio, publicationState: 'published' });
      }),
    );

    const actualizado = await updateService(itemId, { publicationState: 'published' });
    expect(consultada).toBe(`${adminUrl}/horario/${itemId}`);
    expect(actualizado.publicationState).toBe('published');
  });

  it('deleteService hace DELETE y resuelve sin cuerpo', async () => {
    let metodo: string | undefined;
    server.use(
      http.delete(`${adminUrl}/horario/${itemId}`, ({ request }) => {
        metodo = request.method;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    await expect(deleteService(itemId)).resolves.toBeUndefined();
    expect(metodo).toBe('DELETE');
  });
});

describe('whatsapp', () => {
  const canal: WhatsappChannelAdmin = {
    id: itemId,
    nameEs: 'Atención',
    kind: 'direct',
    destination: '+50688888888',
    publicationState: 'draft',
    sortOrder: 0,
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
  };

  it('createWhatsappChannel hace POST con el cuerpo', async () => {
    let cuerpo: unknown;
    server.use(
      http.post(`${adminUrl}/whatsapp`, async ({ request }) => {
        cuerpo = await request.json();
        return HttpResponse.json(canal, { status: 201 });
      }),
    );

    const input = { nameEs: 'Atención', kind: 'direct' as const, destination: '+50688888888' };
    await createWhatsappChannel(input);
    expect(cuerpo).toEqual(input);
  });

  it('updateWhatsappChannel hace PATCH sobre el id', async () => {
    let consultada: string | undefined;
    server.use(
      http.patch(`${adminUrl}/whatsapp/${itemId}`, ({ request }) => {
        consultada = request.url;
        return HttpResponse.json(canal);
      }),
    );

    await updateWhatsappChannel(itemId, { nameEs: 'Atención' });
    expect(consultada).toBe(`${adminUrl}/whatsapp/${itemId}`);
  });

  it('deleteWhatsappChannel hace DELETE', async () => {
    let metodo: string | undefined;
    server.use(
      http.delete(`${adminUrl}/whatsapp/${itemId}`, ({ request }) => {
        metodo = request.method;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    await deleteWhatsappChannel(itemId);
    expect(metodo).toBe('DELETE');
  });

  it('conserva details.network de un 400 de red (panel)', async () => {
    server.use(
      http.post(`${adminUrl}/redes`, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Red no válida',
              details: { network: 'catálogo fijo' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    const error = (await createSocialLink({ network: 'facebook', url: 'https://x' }).catch(
      (cause: unknown) => cause,
    )) as ApiError;
    expect(error.details).toEqual({ network: 'catálogo fijo' });
  });
});

describe('redes', () => {
  const enlace: SocialLinkAdmin = {
    id: itemId,
    network: 'facebook',
    url: 'https://facebook.com/simiente',
    publicationState: 'draft',
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
  };

  it('createSocialLink hace POST con el cuerpo', async () => {
    let cuerpo: unknown;
    server.use(
      http.post(`${adminUrl}/redes`, async ({ request }) => {
        cuerpo = await request.json();
        return HttpResponse.json(enlace, { status: 201 });
      }),
    );

    await createSocialLink({ network: 'facebook', url: 'https://facebook.com/simiente' });
    expect(cuerpo).toEqual({ network: 'facebook', url: 'https://facebook.com/simiente' });
  });

  it('updateSocialLink hace PATCH sobre el id', async () => {
    let modoPeticion: string | undefined;
    server.use(
      http.patch(`${adminUrl}/redes/${itemId}`, ({ request }) => {
        modoPeticion = request.method;
        return HttpResponse.json({ ...enlace, publicationState: 'published' });
      }),
    );

    const actualizado = await updateSocialLink(itemId, { publicationState: 'published' });
    expect(modoPeticion).toBe('PATCH');
    expect(actualizado.publicationState).toBe('published');
  });

  it('deleteSocialLink hace DELETE', async () => {
    let metodo: string | undefined;
    server.use(
      http.delete(`${adminUrl}/redes/${itemId}`, ({ request }) => {
        metodo = request.method;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    await deleteSocialLink(itemId);
    expect(metodo).toBe('DELETE');
  });
});

describe('uploadImage', () => {
  it('envía multipart/form-data con el campo `file`', async () => {
    let metodo: string | undefined;
    const visto: { cuerpo?: string } = {};
    server.use(
      http.post(`${adminUrl}/imagenes`, async ({ request }) => {
        metodo = request.method;
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

    const file = new File(['contenido'], 'logo.png', { type: 'image/png' });
    const resultado = await uploadImage(file);

    expect(metodo).toBe('POST');
    expect(visto.cuerpo).toContain('name="file"');
    expect(visto.cuerpo).toContain('Content-Type: image/png');
    expect(resultado.fileName).toBe('img_aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.png');
  });
});
