import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { MarcaRevision, MedioAdmin, ResumenBusqueda } from '../../../api/medios';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { marcasPendientes } from '../marcas';
import { MediosPage } from './MediosPage';

/**
 * Módulo «Gestión de medios» del panel (ux.md §4.3/§4.4/§5, US2/US3): zona de
 * búsqueda con su «durante» (`role="status"`), resumen de los seis casos, filtro
 * de marcas con conteos, listado con origen/fecha de importación/marcas y sus
 * acciones directas, y los estados de carga, vacío, error y sin permiso.
 *
 * `Table` pinta la **misma** información como tabla (escritorio) y como tarjetas
 * etiquetadas (móvil), las dos en el DOM: las aserciones de fila se acotan a la
 * tabla para no duplicarse.
 */

const adminUrl = `${API_BASE_URL}/api/v1/admin/medios`;
const buscarUrl = `${adminUrl}/buscar-nuevos`;

function makeMedio(overrides: Partial<MedioAdmin> = {}): MedioAdmin {
  return {
    id: 'med-1',
    title: 'Potencial sin Resultados La Lección de la Higuera',
    descriptionEs: 'La higuera estéril y lo que Dios espera de nosotros.',
    descriptionEn: null,
    descriptionEdited: false,
    publishedAt: '2026-05-25T18:43:42Z',
    coverUrl: '/api/v1/media/img_higuera.jpg',
    videoUrl: 'https://www.youtube.com/watch?v=abc12345678',
    audioUrl: 'https://podcasters.spotify.com/pod/show/simiente/episodes/higuera',
    origin: 'both',
    visibility: 'visible',
    sources: [
      {
        provider: 'youtube',
        externalId: 'abc12345678',
        feedTitle: 'Canal Iglesia Simiente Santa',
        url: 'https://www.youtube.com/watch?v=abc12345678',
      },
    ],
    createdAt: '2026-10-18T20:34:17Z',
    updatedAt: '2026-10-18T20:34:17Z',
    ...overrides,
  };
}

function resumen(overrides: Partial<ResumenBusqueda> = {}): ResumenBusqueda {
  return {
    startedAt: '2026-10-18T20:34:00Z',
    finishedAt: '2026-10-18T20:34:20Z',
    importedCount: 0,
    joinedCount: 0,
    markedCount: 0,
    hiddenCount: 0,
    availabilityCheckedCount: 0,
    availabilityInconclusiveCount: 0,
    youtube: 'ok',
    podcast: 'ok',
    ...overrides,
  };
}

/** Responde el listado aplicando `reviewMark` como el servidor (pendiente de esa clase). */
function responder(items: MedioAdmin[]) {
  server.use(
    http.get(adminUrl, ({ request }) => {
      const reviewMark = new URL(request.url).searchParams.get(
        'reviewMark',
      ) as MarcaRevision | null;
      const filtrados = reviewMark
        ? items.filter((item) => marcasPendientes(item).includes(reviewMark))
        : items;
      return HttpResponse.json({
        items: filtrados,
        total: filtrados.length,
        limit: 100,
        offset: 0,
      });
    }),
  );
}

function renderPage() {
  return render(
    <AppProviders>
      <MediosPage />
    </AppProviders>,
  );
}

/** Tabla del listado (la representación de escritorio; la móvil es una lista). */
function tabla() {
  return screen.getByRole('table');
}

describe('MediosPage · zona de búsqueda (US2, ux.md §4.3/§4.4)', () => {
  it('la zona de búsqueda está disponible antes que el listado (no depende de él)', async () => {
    server.use(
      http.get(adminUrl, async () => {
        await delay(50);
        return HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 });
      }),
    );

    renderPage();

    expect(screen.getByRole('heading', { name: 'Desde YouTube y el podcast' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Buscar episodios nuevos' })).toBeEnabled();
    expect(
      screen.getByText(
        'Es seguro pulsarlo todas las veces que quieras: solo agrega lo que falte, nunca borra y nunca cambia lo que ya decidiste.',
      ),
    ).toBeInTheDocument();

    // El listado, mientras tanto, muestra su propio estado de carga.
    expect(screen.getByText('Cargando los episodios…')).toBeInTheDocument();

    expect(await screen.findByText(/Todavía no hay episodios/)).toBeInTheDocument();
  });

  it('durante la búsqueda el botón queda deshabilitado en «Buscando…» con `role="status"`', async () => {
    responder([makeMedio()]);
    server.use(
      http.post(buscarUrl, async () => {
        await delay(50);
        return HttpResponse.json(resumen({ importedCount: 1, joinedCount: 1 }));
      }),
    );

    renderPage();
    await screen.findByRole('table');

    await userEvent.click(screen.getByRole('button', { name: 'Buscar episodios nuevos' }));

    const buscando = screen.getByRole('button', { name: 'Buscando…' });
    expect(buscando).toBeDisabled();
    // El contexto del «durante» se anuncia solo (⟲ D-17).
    expect(
      screen.getByText('Buscando episodios nuevos en YouTube y en el podcast…'),
    ).toHaveAttribute('role', 'status');
    // Las acciones de fila quedan deshabilitadas mientras corre (⟲ D-16).
    expect(within(tabla()).getByRole('button', { name: 'Ocultar' })).toBeDisabled();

    // Y al terminar, el resumen de la búsqueda (⟲ D-22: permanece, no es toast).
    expect(
      await screen.findByText(
        'Se agregaron 1 episodios nuevos; 1 se unieron con su video y su audio. Ya están visibles en “Medios”.',
      ),
    ).toBeInTheDocument();
  });

  it('el resumen cubre los seis casos del catálogo y no menciona segmentos en 0 (⟲ D-19)', async () => {
    responder([]);

    for (const caso of [
      {
        resumen: resumen({ importedCount: 2, joinedCount: 1 }),
        esperado:
          'Se agregaron 2 episodios nuevos; 1 se unieron con su video y su audio. Ya están visibles en “Medios”.',
      },
      {
        resumen: resumen({ importedCount: 3, joinedCount: 1, markedCount: 2 }),
        esperado:
          'Se agregaron 3 episodios nuevos; 1 se unieron con su video y su audio; 2 quedaron marcados como posible duplicado. Revísalos abajo.',
      },
      {
        resumen: resumen(),
        esperado:
          'No hay episodios nuevos: nada cambió. Cuando la iglesia publique algo más, vuelve a pulsar el botón.',
      },
      {
        resumen: resumen({ youtube: 'failed' }),
        esperado:
          'El feed de YouTube no respondió. Se agregó lo que trajo el podcast; nada se ocultó ni se marcó por esa causa. Vuelve a pulsar la búsqueda en un rato para recuperar lo que falte.',
      },
      {
        resumen: resumen({ youtube: 'failed', podcast: 'failed' }),
        esperado:
          'Ninguno de los dos feeds respondió. No se agregó nada y no se cambió nada: todo lo que tenías está a salvo. Prueba más tarde.',
      },
      {
        resumen: resumen({ availabilityCheckedCount: 4 }),
        esperado:
          'La búsqueda revisó 4 episodios para ver si sus enlaces siguen funcionando: mira las marcas.',
      },
    ]) {
      server.use(http.post(buscarUrl, () => HttpResponse.json(caso.resumen)));
      const { unmount } = renderPage();
      await userEvent.click(screen.getByRole('button', { name: 'Buscar episodios nuevos' }));
      expect(await screen.findByText(caso.esperado)).toBeInTheDocument();
      unmount();
    }
  });

  it('un feed caído sin novedades no añade «0 episodios nuevos» (⟲ D-19)', async () => {
    responder([]);
    server.use(http.post(buscarUrl, () => HttpResponse.json(resumen({ podcast: 'failed' }))));

    renderPage();
    await userEvent.click(screen.getByRole('button', { name: 'Buscar episodios nuevos' }));

    expect(await screen.findByText(/El feed del podcast no respondió/)).toBeInTheDocument();
    expect(screen.queryByText(/Se agregaron 0/)).toBeNull();
    expect(screen.queryByText(/No hay episodios nuevos/)).toBeNull();
  });
});

describe('MediosPage · listado del catálogo (US3, ux.md §4.3/§5.b)', () => {
  it('la fila muestra carátula, nombre, fecha, formatos, origen, fecha de importación, estado y marca', async () => {
    responder([
      makeMedio({
        duplicateOfId: 'med-2',
        availabilityMark: 'missing_from_feed',
        availabilityResult: 'available',
      }),
    ]);

    renderPage();
    await screen.findByRole('table');

    const fila = tabla();
    const celda = within(fila)
      .getByText('Potencial sin Resultados La Lección de la Higuera')
      .closest('td') as HTMLElement;

    // Carátula miniatura y fecha de publicación (en la zona de la iglesia).
    expect(celda.querySelector('img')).toHaveAttribute(
      'src',
      `${API_BASE_URL}/api/v1/media/img_higuera.jpg`,
    );
    expect(within(celda).getByText('lunes 25 de mayo')).toBeInTheDocument();
    // Origen (FR-019/⟲ A2) y fecha de importación.
    expect(within(celda).getByText('De YouTube y del podcast')).toBeInTheDocument();
    expect(within(celda).getByText('Importado el 18/10/2026')).toBeInTheDocument();
    // Fuentes de la fila: proveedor, título del feed y enlace.
    expect(
      within(celda).getByRole('link', { name: 'https://www.youtube.com/watch?v=abc12345678' }),
    ).toHaveAttribute('target', '_blank');

    // Chips de formato por palabra (⟲ D-15) y estado con palabra.
    expect(within(fila).getAllByText('Video').length).toBeGreaterThan(0);
    expect(within(fila).getAllByText('Audio').length).toBeGreaterThan(0);
    expect(within(fila).getByText('Visible')).toBeInTheDocument();

    // Los dos ejes conviven: dos píldoras con su explicador (P9-19).
    expect(within(fila).getByText('Posible duplicado')).toBeInTheDocument();
    expect(
      within(fila).getByText(
        'El sistema no pudo estar seguro de si estas dos entradas son el mismo sermón.',
      ),
    ).toBeInTheDocument();
    expect(within(fila).getByText('Ya no aparece en el feed')).toBeInTheDocument();
    expect(
      within(fila).getByText(
        'Salió de la ventana del feed de YouTube (así es: la ventana es corta). Su enlace sigue funcionando y sigue visible para el público.',
      ),
    ).toBeInTheDocument();
  });

  it('sin carátula la fila usa el marcador neutro y sin marcas no hay píldora', async () => {
    responder([makeMedio({ coverUrl: undefined, origin: 'audio' })]);

    const { container } = renderPage();
    await screen.findByRole('table');

    // Tabla + tarjeta móvil: una miniatura neutra por representación.
    expect(container.querySelectorAll('[data-testid="miniatura-neutra"]')).toHaveLength(2);
    expect(within(tabla()).getByText('Del podcast')).toBeInTheDocument();
    expect(within(tabla()).queryByText('Posible duplicado')).toBeNull();
    expect(within(tabla()).queryByText('No disponible')).toBeNull();
  });

  it('«Ocultar» es directo (sin confirmación) y su aviso dice el efecto real', async () => {
    const peticiones: string[] = [];
    responder([makeMedio()]);
    server.use(
      http.patch(`${adminUrl}/med-1`, async ({ request }) => {
        peticiones.push(await request.text());
        return HttpResponse.json(makeMedio({ visibility: 'hidden' }));
      }),
    );

    renderPage();
    await screen.findByRole('table');

    await userEvent.click(within(tabla()).getByRole('button', { name: 'Ocultar' }));

    expect(await screen.findByText('Quedó oculto: el público no lo ve.')).toBeInTheDocument();
    expect(peticiones).toEqual(['{"visibility":"hidden"}']);
    // Directo: ningún diálogo de confirmación por medio.
    expect(screen.queryByRole('dialog')).toBeNull();
  });

  it('las resoluciones de la marca de disponibilidad viajan con SU `mark` (P9-19)', async () => {
    const peticiones: string[] = [];
    responder([makeMedio({ availabilityMark: 'missing_from_feed' })]);
    server.use(
      http.post(`${adminUrl}/med-1/marcas/resolver`, async ({ request }) => {
        peticiones.push(await request.text());
        return HttpResponse.json(makeMedio());
      }),
    );

    renderPage();
    await screen.findByRole('table');

    // La fila no duplica el botón genérico: sus resoluciones cubren las dos direcciones.
    expect(within(tabla()).queryByRole('button', { name: 'Mostrar' })).toBeNull();
    await userEvent.click(within(tabla()).getByRole('button', { name: 'Mantener visible' }));

    expect(
      await screen.findByText('La marca se cerró y no se vuelve a abrir por esta misma causa.'),
    ).toBeInTheDocument();
    expect(peticiones).toEqual(['{"mark":"missing_from_feed","visibility":"visible"}']);
  });

  it('la marca «No disponible» ofrece «Mantener oculto» directo y resuelve su eje', async () => {
    const peticiones: string[] = [];
    responder([makeMedio({ availabilityMark: 'unavailable', visibility: 'hidden' })]);
    server.use(
      http.post(`${adminUrl}/med-1/marcas/resolver`, async ({ request }) => {
        peticiones.push(await request.text());
        return HttpResponse.json(makeMedio({ visibility: 'hidden' }));
      }),
    );

    renderPage();
    await screen.findByRole('table');

    expect(within(tabla()).getByText('Oculto')).toBeInTheDocument();
    await userEvent.click(within(tabla()).getByRole('button', { name: 'Mantener oculto' }));

    expect(
      await screen.findByText('La marca se cerró y no se vuelve a abrir por esta misma causa.'),
    ).toBeInTheDocument();
    expect(peticiones).toEqual(['{"mark":"unavailable","visibility":"hidden"}']);
  });
});

describe('MediosPage · filtro de marcas (⟲ D-10)', () => {
  it('muestra las pastillas con sus conteos y filtra por el parámetro real', async () => {
    const conPar = makeMedio({ id: 'med-1', duplicateOfId: 'med-2' });
    const sinMarca = makeMedio({ id: 'med-3', title: 'Fe como fruto del Espíritu' });
    const conFeed = makeMedio({
      id: 'med-4',
      title: 'La Fe y la actitud',
      availabilityMark: 'missing_from_feed',
    });
    responder([conPar, sinMarca, conFeed]);

    renderPage();
    await screen.findByRole('table');

    expect(screen.getByRole('button', { name: 'Todo' })).toHaveAttribute('aria-pressed', 'true');
    const pastillaPar = screen.getByRole('button', { name: 'Posible duplicado (1)' });
    expect(
      screen.getByRole('button', { name: 'Ya no aparece en el feed (1)' }),
    ).toBeInTheDocument();
    // Las clases sin pendientes no se pintan.
    expect(screen.queryByRole('button', { name: /No disponible/ })).toBeNull();

    await userEvent.click(pastillaPar);

    await waitFor(() =>
      expect(within(tabla()).queryByText('Fe como fruto del Espíritu')).toBeNull(),
    );
    expect(
      within(tabla()).getByText('Potencial sin Resultados La Lección de la Higuera'),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Posible duplicado (1)' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
  });

  it('el filtro sin resultados dice que todo está al día (no es un hueco, §4.8)', async () => {
    const conFeed = makeMedio({ availabilityMark: 'missing_from_feed' });
    // El servidor responde vacío para esa clase (p. ej. otra persona ya la resolvió).
    server.use(
      http.get(adminUrl, ({ request }) => {
        const reviewMark = new URL(request.url).searchParams.get('reviewMark');
        const items = reviewMark ? [] : [conFeed];
        return HttpResponse.json({ items, total: items.length, limit: 100, offset: 0 });
      }),
    );

    renderPage();
    await screen.findByRole('table');

    await userEvent.click(screen.getByRole('button', { name: 'Ya no aparece en el feed (1)' }));

    expect(
      await screen.findByText('No hay nada pendiente de revisión: todo está al día.'),
    ).toBeInTheDocument();
    // Sin duplicar la puerta de importación (⟲ D-33): el único botón de la página
    // sigue siendo el de la zona de búsqueda.
    expect(screen.getAllByRole('button', { name: 'Buscar episodios nuevos' })).toHaveLength(1);
  });
});

describe('MediosPage · estados (ux.md §5.b, FR-021)', () => {
  it('el catálogo vacío remite a la zona de importación sin duplicar el botón (⟲ D-33)', async () => {
    responder([]);

    renderPage();

    expect(await screen.findByText(/Todavía no hay episodios\./)).toBeInTheDocument();
    expect(
      screen.getByText(
        /Pulsa “Buscar episodios nuevos” para que el sitio lea por primera vez el canal de YouTube y el feed del podcast\./,
      ),
    ).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: 'Buscar episodios nuevos' })).toHaveLength(1);
  });

  it('el error de carga ofrece «Reintentar» y recupera el listado', async () => {
    let fallos = 1;
    server.use(
      http.get(adminUrl, () => {
        if (fallos > 0) {
          fallos -= 1;
          return HttpResponse.json(
            { error: { code: 'internal', message: 'Error' } },
            { status: 500 },
          );
        }
        return HttpResponse.json({ items: [makeMedio()], total: 1, limit: 100, offset: 0 });
      }),
    );

    renderPage();

    expect(await screen.findByText('No se pudo cargar los episodios.')).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Reintentar' }));

    expect(await screen.findByRole('table')).toBeInTheDocument();
    expect(
      within(tabla()).getByText('Potencial sin Resultados La Lección de la Higuera'),
    ).toBeInTheDocument();
  });

  it('sin permiso del servidor avisa sin detalles internos (FR-021)', async () => {
    server.use(
      http.get(adminUrl, () =>
        HttpResponse.json(
          { error: { code: 'forbidden', message: 'No autorizado' } },
          { status: 403 },
        ),
      ),
    );

    renderPage();

    expect(
      await screen.findByText(
        'No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol.',
      ),
    ).toBeInTheDocument();
    // Sin permiso no se ofrece la búsqueda (la autoridad es el servidor).
    expect(screen.queryByRole('button', { name: 'Buscar episodios nuevos' })).toBeNull();
  });
});
