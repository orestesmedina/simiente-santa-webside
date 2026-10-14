import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { MinisterioAdmin } from '../../api/ministerios';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { MinistriesPage } from './pages/MinistriesPage';

/**
 * Módulo «Gestión de ministerios» del panel (`/panel/ministerios`, ux.md
 * §4.3/§5.a): título, banda de ayuda, listado con Orden/Nombre/Horario-Lugar/
 * Encargado (con la marca «Ficha oculta»)/Estado/Acciones, publicar/retirar/
 * eliminar con sus confirmaciones y los estados de carga, vacío, error y éxito.
 */

const listUrl = `${API_BASE_URL}/api/v1/admin/ministerios`;

function makeMinisterio(overrides: Partial<MinisterioAdmin> = {}): MinisterioAdmin {
  return {
    id: 'min-1',
    nameEs: 'Alabanza',
    nameEn: null,
    descriptionEs: 'Llevamos la adoración en los cultos.',
    descriptionEn: null,
    leaderNameEs: 'Josué Ramírez',
    leaderNameEn: null,
    leaderPhone: '+504 9999-9999',
    leaderEmail: null,
    leaderVisible: true,
    whatsappUrl: 'https://chat.whatsapp.com/ABC123',
    whatsappUrlShared: false,
    scheduleEs: 'Todos los viernes, 7:00 p. m.',
    scheduleEn: null,
    placeEs: 'Templo central',
    placeEn: null,
    imageFile: null,
    imageAltEs: null,
    imageAltEn: null,
    publicationState: 'published',
    sortOrder: 1,
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function responderLista(items: MinisterioAdmin[]) {
  server.use(
    http.get(listUrl, () =>
      HttpResponse.json({ items, total: items.length, limit: 100, offset: 0 }),
    ),
  );
}

function renderPage() {
  return render(
    <AppProviders>
      <MinistriesPage />
    </AppProviders>,
  );
}

describe('MinistriesPage · listado del panel (US2/US3, ux.md §4.3)', () => {
  it('muestra Orden, Nombre, Horario/Lugar, Encargado, Estado y Acciones en el orden del servidor', async () => {
    responderLista([
      makeMinisterio({
        id: 'min-2',
        nameEs: 'Escuela dominical',
        sortOrder: 1,
        leaderNameEs: null,
        leaderPhone: null,
        leaderVisible: false,
        scheduleEs: null,
        placeEs: null,
      }),
      makeMinisterio({ id: 'min-1', nameEs: 'Alabanza', sortOrder: 2, publicationState: 'draft' }),
    ]);

    renderPage();

    const tabla = await screen.findByRole('table');
    for (const header of [
      'Orden',
      'Nombre',
      'Horario / Lugar',
      'Encargado',
      'Estado',
      'Acciones',
    ]) {
      expect(within(tabla).getByRole('columnheader', { name: header })).toBeInTheDocument();
    }

    // Orden del servidor (`sortOrder`, `id`): el panel muestra el que ve el público.
    const filas = within(tabla).getAllByRole('row').slice(1);
    expect(filas).toHaveLength(2);
    expect(within(filas[0] as HTMLElement).getByText('Escuela dominical')).toBeInTheDocument();
    expect(within(filas[1] as HTMLElement).getByText('Alabanza')).toBeInTheDocument();

    // Orden numérico (D-6) y «—» cuando no hay horario ni lugar.
    expect(within(filas[0] as HTMLElement).getByText('1')).toBeInTheDocument();
    expect(within(filas[0] as HTMLElement).getAllByText('—').length).toBeGreaterThan(0);
    expect(
      within(filas[1] as HTMLElement).getByText('Todos los viernes, 7:00 p. m. · Templo central'),
    ).toBeInTheDocument();

    // Estados: píldora con palabra, sin marca «Pasado» (nada vence en F6).
    expect(within(filas[0] as HTMLElement).getByText('Publicado')).toBeInTheDocument();
    expect(within(filas[1] as HTMLElement).getByText('Borrador')).toBeInTheDocument();
    expect(screen.queryByText('Pasado')).toBeNull();
  });

  it('la marca «Ficha oculta» aparece solo con el interruptor apagado', async () => {
    responderLista([
      makeMinisterio({ id: 'min-1', nameEs: 'Alabanza', leaderVisible: true }),
      makeMinisterio({
        id: 'min-2',
        nameEs: 'Sonido',
        sortOrder: 2,
        leaderVisible: false,
        leaderNameEs: 'Ana Rosales',
      }),
    ]);

    renderPage();

    const tabla = await screen.findByRole('table');
    const filas = within(tabla).getAllByRole('row').slice(1);

    // El panel SIEMPRE ve el dato guardado (dato guardado ≠ dato publicado).
    expect(within(filas[0] as HTMLElement).getByText('Josué Ramírez')).toBeInTheDocument();
    expect(within(filas[0] as HTMLElement).queryByText('Ficha oculta')).toBeNull();

    expect(within(filas[1] as HTMLElement).getByText('Ana Rosales')).toBeInTheDocument();
    expect(within(filas[1] as HTMLElement).getByText('Ficha oculta')).toBeInTheDocument();
  });

  it('publica y retira con confirmación suave, con sus mensajes', async () => {
    let actual: MinisterioAdmin = makeMinisterio({ publicationState: 'draft' });
    server.use(
      http.get(listUrl, () =>
        HttpResponse.json({ items: [actual], total: 1, limit: 100, offset: 0 }),
      ),
      http.patch(`${listUrl}/${actual.id}`, async ({ request }) => {
        const patch = (await request.json()) as Partial<MinisterioAdmin>;
        actual = { ...actual, ...patch };
        return HttpResponse.json(actual);
      }),
    );

    renderPage();

    // La `Table` pinta tabla (escritorio) y tarjetas (móvil): se opera la tabla.
    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Publicar' }));
    expect(await screen.findByText('Publicado en "Ministerios".')).toBeInTheDocument();

    // Retirar pide confirmación suave (los datos se conservan).
    await userEvent.click(within(tabla).getByRole('button', { name: 'Retirar de la sección' }));
    const confirmacion = await screen.findByRole('dialog');
    expect(
      within(confirmacion).getByText(
        '"Alabanza" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.',
      ),
    ).toBeInTheDocument();
    await userEvent.click(within(confirmacion).getByRole('button', { name: 'Retirar' }));

    expect(
      await screen.findByText(
        'Retirado de la sección. Estará disponible como borrador para publicarlo cuando quieras.',
      ),
    ).toBeInTheDocument();
  });

  it('elimina con confirmación explícita de no recuperable', async () => {
    let borrado = false;
    const ministerio = makeMinisterio();
    server.use(
      http.get(listUrl, () =>
        HttpResponse.json({ items: [ministerio], total: 1, limit: 100, offset: 0 }),
      ),
      http.delete(`${listUrl}/${ministerio.id}`, () => {
        borrado = true;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    renderPage();

    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Eliminar' }));
    const confirmacion = await screen.findByRole('dialog');
    expect(
      within(confirmacion).getByText(
        '"Alabanza" se eliminará para siempre. Esto no se puede deshacer.',
      ),
    ).toBeInTheDocument();
    await userEvent.click(within(confirmacion).getByRole('button', { name: 'Eliminar' }));

    expect(await screen.findByText('Eliminado.')).toBeInTheDocument();
    expect(borrado).toBe(true);
  });
});

describe('MinistriesPage · estados de la vista (ux.md §5.a)', () => {
  it('mientras carga anuncia el estado y no pinta la tabla', async () => {
    server.use(
      http.get(listUrl, async () => {
        await delay(50);
        return HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 });
      }),
    );

    renderPage();

    expect(screen.getByText('Cargando ministerios…')).toBeInTheDocument();
    expect(screen.queryByRole('table')).toBeNull();
    expect(
      await screen.findByText(
        'Todavía no hay ministerios. Agrega el primero para que el público lo vea.',
      ),
    ).toBeInTheDocument();
  });

  it('sin ministerios ofrece el estado vacío con su acción', async () => {
    responderLista([]);

    renderPage();

    expect(
      await screen.findByText(
        'Todavía no hay ministerios. Agrega el primero para que el público lo vea.',
      ),
    ).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: 'Agregar ministerio' }).length).toBeGreaterThan(0);
  });

  it('con todo en borrador avisa que la sección y la portada no se ven (FR-005/FR-028)', async () => {
    responderLista([
      makeMinisterio({ publicationState: 'draft' }),
      makeMinisterio({ id: 'min-2', nameEs: 'Sonido', publicationState: 'draft' }),
    ]);

    renderPage();

    expect(await screen.findByText(/Nada está publicado ahora/)).toBeInTheDocument();
    expect(screen.getByText(/redirige a la portada/)).toBeInTheDocument();
    // El listado completo sigue a la vista (US3 esc. 7).
    expect(screen.getByRole('table')).toBeInTheDocument();
  });

  it('el error de carga ofrece «Reintentar» y recupera el listado', async () => {
    let fallos = 1;
    server.use(
      http.get(listUrl, () => {
        if (fallos > 0) {
          fallos -= 1;
          return HttpResponse.json(
            { error: { code: 'internal', message: 'Error' } },
            { status: 500 },
          );
        }
        return HttpResponse.json({
          items: [makeMinisterio()],
          total: 1,
          limit: 100,
          offset: 0,
        });
      }),
    );

    renderPage();

    expect(await screen.findByText('No se pudo cargar los ministerios.')).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Reintentar' }));

    expect(await screen.findByRole('table')).toBeInTheDocument();
  });

  it('sin permiso en el servidor (403) avisa sin detalles internos', async () => {
    server.use(
      http.get(listUrl, () =>
        HttpResponse.json(
          { error: { code: 'forbidden', message: 'Sin permiso' } },
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
    expect(screen.queryByRole('table')).toBeNull();
  });

  it('con contenido en caché y una recarga fallida conserva el listado y avisa', async () => {
    let peticiones = 0;
    server.use(
      http.get(listUrl, () => {
        peticiones += 1;
        if (peticiones === 1) {
          return HttpResponse.json({
            items: [makeMinisterio({ publicationState: 'draft' })],
            total: 1,
            limit: 100,
            offset: 0,
          });
        }
        return HttpResponse.json(
          { error: { code: 'internal', message: 'Error' } },
          { status: 500 },
        );
      }),
      http.patch(`${listUrl}/min-1`, () =>
        HttpResponse.json(makeMinisterio({ publicationState: 'published' })),
      ),
    );

    renderPage();
    const tabla = await screen.findByRole('table');

    // Publicar invalida el listado y la recarga falla: se conserva el último
    // listado válido y se añade un aviso discreto (ux.md §5.a).
    await userEvent.click(within(tabla).getByRole('button', { name: 'Publicar' }));
    await waitFor(() => expect(peticiones).toBeGreaterThan(1));

    expect(screen.getByRole('table')).toBeInTheDocument();
    expect(await screen.findByText('No se pudo cargar los ministerios.')).toBeInTheDocument();
  });
});
