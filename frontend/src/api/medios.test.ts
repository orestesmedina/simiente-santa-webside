import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { server } from '../test/server';
import { API_BASE_URL } from './client';
import {
  MEDIOS_ADMIN_PAGE_LIMIT,
  findMedio,
  listMedios,
  type MedioAdmin,
  type MedioAdminList,
} from './medios';

/**
 * Cliente del módulo «Medios» (F9, contrato 0.9.0): la búsqueda de una ficha
 * por identificador (B3 de la revisión) recorre las páginas del listado del
 * panel porque el contrato no tiene `GET /admin/medios/{id}`, de modo que la
 * unión de duplicados no depende de la página cargada.
 */

const adminUrl = `${API_BASE_URL}/api/v1/admin/medios`;

function makeMedio(id: string, title = `Episodio ${id}`): MedioAdmin {
  return {
    id,
    title,
    descriptionEs: null,
    descriptionEn: null,
    descriptionEdited: false,
    publishedAt: '2026-05-25T18:43:42Z',
    origin: 'video',
    visibility: 'visible',
    sources: [],
    createdAt: '2026-10-18T20:34:17Z',
    updatedAt: '2026-10-18T20:34:17Z',
  };
}

/** Responde el listado paginado como el servidor (`limit`/`offset`/`total`). */
function responderPaginado(items: MedioAdmin[]) {
  const consultas: string[] = [];
  server.use(
    http.get(adminUrl, ({ request }) => {
      consultas.push(request.url);
      const url = new URL(request.url);
      const limit = Number(url.searchParams.get('limit') ?? MEDIOS_ADMIN_PAGE_LIMIT);
      const offset = Number(url.searchParams.get('offset') ?? 0);
      const pagina: MedioAdminList = {
        items: items.slice(offset, offset + limit),
        total: items.length,
        limit,
        offset,
      };
      return HttpResponse.json(pagina);
    }),
  );
  return consultas;
}

describe('findMedio (B3: la pareja no depende de la página cargada)', () => {
  it('encuentra una ficha de la segunda página y pide las páginas con el tope del contrato', async () => {
    // 100 fichas en la primera página y la buscada en la segunda (101 en total).
    const primera = Array.from({ length: MEDIOS_ADMIN_PAGE_LIMIT }, (_, indice) =>
      makeMedio(`med-${indice}`),
    );
    const buscada = makeMedio('med-100', 'La ficha que quedó fuera del tope');
    const consultas = responderPaginado([...primera, buscada]);

    const encontrada = await findMedio('med-100');

    expect(encontrada?.title).toBe('La ficha que quedó fuera del tope');
    expect(consultas).toEqual([
      `${adminUrl}?limit=${MEDIOS_ADMIN_PAGE_LIMIT}&offset=0`,
      `${adminUrl}?limit=${MEDIOS_ADMIN_PAGE_LIMIT}&offset=${MEDIOS_ADMIN_PAGE_LIMIT}`,
    ]);
  });

  it('encuentra en la primera página sin pedir la segunda', async () => {
    const consultas = responderPaginado([makeMedio('med-1'), makeMedio('med-2')]);

    const encontrada = await findMedio('med-2');

    expect(encontrada?.id).toBe('med-2');
    expect(consultas).toHaveLength(1);
  });

  it('devuelve `undefined` cuando la ficha ya no existe, sin dar vueltas de más', async () => {
    const consultas = responderPaginado([makeMedio('med-1')]);

    await expect(findMedio('med-99')).resolves.toBeUndefined();
    expect(consultas).toHaveLength(1);
  });
});

describe('listMedios (contrato del panel)', () => {
  it('pide el listado con los parámetros del panel', async () => {
    const consultas = responderPaginado([makeMedio('med-1')]);

    await listMedios({ limit: MEDIOS_ADMIN_PAGE_LIMIT, offset: 0 });

    expect(consultas).toEqual([`${adminUrl}?limit=${MEDIOS_ADMIN_PAGE_LIMIT}&offset=0`]);
  });
});
