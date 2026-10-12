import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { CategoriaAdmin, EventoAdmin } from '../../api/agenda';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { EventsPage } from './pages/EventsPage';

const eventsUrl = `${API_BASE_URL}/api/v1/admin/eventos`;
const categoriesUrl = `${API_BASE_URL}/api/v1/admin/categorias`;

const categoria: CategoriaAdmin = {
  id: 'cat-1',
  nameEs: 'acción social',
  sortOrder: 0,
  createdAt: '2026-10-01T10:00:00Z',
  updatedAt: '2026-10-01T10:00:00Z',
};

function makeEvento(overrides: Partial<EventoAdmin> = {}): EventoAdmin {
  return {
    id: 'ev-1',
    titleEs: 'Vigilia de viernes',
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
    dates: [{ id: 'd-1', date: '2026-10-17', startTime: '19:00', isPast: false }],
    nextDate: '2026-10-17',
    isPast: false,
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function responderEventos(items: EventoAdmin[]) {
  server.use(
    http.get(eventsUrl, () =>
      HttpResponse.json({ items, total: items.length, limit: 100, offset: 0 }),
    ),
  );
}

function responderCategorias(items: CategoriaAdmin[] = []) {
  server.use(http.get(categoriesUrl, () => HttpResponse.json({ items })));
}

function renderPage() {
  return render(
    <AppProviders>
      <EventsPage />
    </AppProviders>,
  );
}

/** Abre el `Dialog` de alta. */
async function abrirAlta() {
  await userEvent.click((await screen.findAllByRole('button', { name: 'Agregar evento' }))[0]);
  return screen.getByRole('dialog');
}

/** Abre el `Dialog` de alta y rellena los campos obligatorios del español. */
async function abrirAltaYCompletar(titulo = 'Vigilia de viernes') {
  const dialog = await abrirAlta();
  await userEvent.type(within(dialog).getByLabelText(/^Título \(Español\)/), titulo);
  await userEvent.type(within(dialog).getByLabelText(/^Lugar \(Español\)/), 'Templo central');
  await userEvent.type(within(dialog).getByLabelText(/^Detalle \(Español\)/), 'Noche de oración');
  return dialog;
}

describe('EventsPage del panel · listado (US2, ux.md §4.3)', () => {
  it('muestra próxima fecha escrita, frecuencia, estado y la marca «Pasado»', async () => {
    responderCategorias([categoria]);
    responderEventos([
      makeEvento({ publicationState: 'published' }),
      makeEvento({
        id: 'ev-2',
        titleEs: 'Vigilia pasada',
        publicationState: 'published',
        nextDate: null,
        isPast: true,
        dates: [{ id: 'd-2', date: '2026-10-03', isPast: true }],
      }),
    ]);

    renderPage();

    // La `Table` pinta tabla (escritorio) y tarjetas (móvil): se comprueba la tabla.
    const tabla = await screen.findByRole('table');
    expect(within(tabla).getByText('Vigilia de viernes')).toBeInTheDocument();
    expect(within(tabla).getByText('sábado 17 de octubre · 19:00')).toBeInTheDocument();
    expect(within(tabla).getByText('Todos los viernes · 1 fecha futura')).toBeInTheDocument();
    // Sin fechas futuras solo se muestra la frecuencia: la marca «Pasado» ya lo dice.
    expect(within(tabla).getByText('Todos los viernes')).toBeInTheDocument();
    expect(within(tabla).getAllByText('Publicado')).toHaveLength(2);

    // La marca «Pasado» es una etiqueta aparte, no un cuarto estado (⟲ D-8).
    expect(within(tabla).getByText('Pasado')).toBeInTheDocument();
    expect(within(tabla).getByText(/Sus fechas ya pasaron/)).toBeInTheDocument();
    expect(within(tabla).getByText('—')).toBeInTheDocument();
  });

  it('con todo en borrador avisa que el público no ve la sección', async () => {
    responderCategorias();
    responderEventos([makeEvento(), makeEvento({ id: 'ev-2', titleEs: 'Otra' })]);

    renderPage();

    expect(await screen.findByText(/Nada está publicado ahora/)).toBeInTheDocument();
  });

  it('sin eventos muestra el estado vacío con su acción', async () => {
    responderCategorias();
    responderEventos([]);

    renderPage();

    expect(
      await screen.findByText(
        'Todavía no hay eventos. Agrega el primero para que el público lo vea.',
      ),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Agregar evento' })).toBeInTheDocument();
  });
});

describe('EventForm · validación (FR-015)', () => {
  it('crea un evento con fechas válidas y avisa que quedó en borrador', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    responderCategorias([categoria]);
    responderEventos([]);
    server.use(
      http.post(eventsUrl, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeEvento(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAltaYCompletar();

    fireEvent.change(within(dialog).getByLabelText(/^Fecha 1/), {
      target: { value: '2026-10-17' },
    });
    fireEvent.change(within(dialog).getByLabelText(/^Hora de inicio/), {
      target: { value: '19:00' },
    });
    await userEvent.selectOptions(within(dialog).getByLabelText(/^Categoría/), 'cat-1');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar evento' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(cuerpo).toMatchObject({
      titleEs: 'Vigilia de viernes',
      placeEs: 'Templo central',
      detailEs: 'Noche de oración',
      categoryId: 'cat-1',
      dates: [{ date: '2026-10-17', startTime: '19:00' }],
    });

    expect(await screen.findByText('Evento creado.')).toBeInTheDocument();
    expect(screen.getByText(/Quedó en borrador/)).toBeInTheDocument();
  });

  it('la hora de fin sin hora de inicio muestra el error del campo (decisión firme)', async () => {
    let peticiones = 0;
    responderCategorias();
    responderEventos([]);
    server.use(
      http.post(eventsUrl, () => {
        peticiones += 1;
        return HttpResponse.json(makeEvento(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAltaYCompletar();

    fireEvent.change(within(dialog).getByLabelText(/^Fecha 1/), {
      target: { value: '2026-10-17' },
    });
    fireEvent.change(within(dialog).getByLabelText(/^Hora de fin/), { target: { value: '21:00' } });
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar evento' }));

    // Aparece junto a la fila de la fecha y en el resumen del formulario.
    expect(
      (
        await within(dialog).findAllByText(
          'Escribe la hora de inicio para poder poner una hora de fin.',
        )
      ).length,
    ).toBeGreaterThan(0);
    expect(peticiones).toBe(0);
  });

  it('rechaza una fecha repetida y una hora de fin anterior a la de inicio', async () => {
    let peticiones = 0;
    responderCategorias();
    responderEventos([]);
    server.use(
      http.post(eventsUrl, () => {
        peticiones += 1;
        return HttpResponse.json(makeEvento(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAltaYCompletar();

    fireEvent.change(within(dialog).getByLabelText(/^Fecha 1/), {
      target: { value: '2026-10-17' },
    });
    fireEvent.change(within(dialog).getByLabelText(/^Hora de inicio/), {
      target: { value: '19:00' },
    });
    fireEvent.change(within(dialog).getByLabelText(/^Hora de fin/), { target: { value: '18:00' } });
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar fecha' }));
    fireEvent.change(within(dialog).getByLabelText(/^Fecha 2/), {
      target: { value: '2026-10-17' },
    });
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar evento' }));

    expect(
      (await within(dialog).findAllByText('Esa fecha ya está en la lista del evento.')).length,
    ).toBeGreaterThan(0);
    expect(
      within(dialog).getAllByText('La hora de fin debe ser posterior a la hora de inicio.').length,
    ).toBeGreaterThan(0);
    expect(peticiones).toBe(0);
  });

  it('sin fechas no guarda y pide agregar al menos una (FR-015)', async () => {
    let peticiones = 0;
    responderCategorias();
    responderEventos([]);
    server.use(
      http.post(eventsUrl, () => {
        peticiones += 1;
        return HttpResponse.json(makeEvento(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAltaYCompletar();
    await userEvent.click(within(dialog).getByRole('button', { name: 'Quitar la fecha 1' }));
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar evento' }));

    expect(
      (await within(dialog).findAllByText('Agrega al menos una fecha para el evento.')).length,
    ).toBeGreaterThan(0);
    expect(peticiones).toBe(0);
  });

  it('el resumen lleva el foco al primer campo inválido (FR-015)', async () => {
    responderCategorias();
    responderEventos([]);

    renderPage();
    const dialog = await abrirAlta();
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar evento' }));

    expect(await within(dialog).findByText('Revisa estos campos:')).toBeInTheDocument();
    await waitFor(() => expect(document.activeElement).toHaveAttribute('id', 'agenda-title-es'));
  });

  it('muestra el `409` de duplicado sin perder lo escrito', async () => {
    responderCategorias();
    responderEventos([]);
    server.use(
      http.post(eventsUrl, () =>
        HttpResponse.json(
          { error: { code: 'conflict', message: 'Ya existe un evento igual' } },
          { status: 409 },
        ),
      ),
    );

    renderPage();
    const dialog = await abrirAltaYCompletar();

    fireEvent.change(within(dialog).getByLabelText(/^Fecha 1/), {
      target: { value: '2026-10-17' },
    });
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar evento' }));

    expect(
      await within(dialog).findByText(/Ya existe un evento igual \(mismo título/),
    ).toBeInTheDocument();
    // Nada se pierde: los valores permanecen (ux.md §2.6).
    expect(within(dialog).getByLabelText(/^Título \(Español\)/)).toHaveValue('Vigilia de viernes');
    expect(within(dialog).getByLabelText(/^Fecha 1/)).toHaveValue('2026-10-17');
  });
});

describe('EventsList · publicar, retirar y eliminar (US4)', () => {
  it('editar un evento publicado avisa que el cambio ya es visible (FR-014)', async () => {
    let parche: Record<string, unknown> | undefined;
    responderCategorias();
    responderEventos([makeEvento({ publicationState: 'published' })]);
    server.use(
      http.patch(`${eventsUrl}/ev-1`, async ({ request }) => {
        parche = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeEvento({ publicationState: 'published' }));
      }),
    );

    renderPage();
    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Editar' }));
    const dialog = screen.getByRole('dialog');
    await userEvent.type(within(dialog).getByLabelText(/^Título \(Español\)/), ' ampliada');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Guardar cambios' }));

    await waitFor(() => expect(parche).toBeDefined());
    expect(parche?.titleEs).toBe('Vigilia de viernes ampliada');
    expect(
      await screen.findByText('Cambios guardados: ya visibles en la sección.'),
    ).toBeInTheDocument();
  });

  it('publicar es una acción directa y retirar pide confirmación (ux.md §5.c)', async () => {
    const parches: Array<Record<string, unknown>> = [];
    let estado: 'draft' | 'published' = 'draft';
    responderCategorias();
    server.use(
      http.get(eventsUrl, () =>
        HttpResponse.json({
          items: [makeEvento({ publicationState: estado })],
          total: 1,
          limit: 100,
          offset: 0,
        }),
      ),
      http.patch(`${eventsUrl}/ev-1`, async ({ request }) => {
        const body = (await request.json()) as Record<string, unknown>;
        parches.push(body);
        estado = body.publicationState === 'published' ? 'published' : 'draft';
        return HttpResponse.json(makeEvento({ publicationState: estado }));
      }),
    );

    renderPage();
    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Publicar' }));
    expect(await screen.findByText('Publicado en "Eventos y actividades".')).toBeInTheDocument();
    expect(parches[0]).toEqual({ publicationState: 'published' });

    await userEvent.click(
      await within(tabla).findByRole('button', { name: 'Retirar de la sección' }),
    );
    const confirm = screen.getByRole('dialog');
    expect(within(confirm).getByText(/dejará de ser visible al público/)).toBeInTheDocument();
    await userEvent.click(within(confirm).getByRole('button', { name: 'Retirar' }));

    await waitFor(() => expect(parches).toHaveLength(2));
    expect(parches[1]).toEqual({ publicationState: 'draft' });
    expect(await screen.findByText(/Retirado de la sección/)).toBeInTheDocument();
  });

  it('eliminar pide confirmación explícita y no deja papelera (FR-015)', async () => {
    let borrados = 0;
    responderCategorias();
    responderEventos([makeEvento()]);
    server.use(
      http.delete(`${eventsUrl}/ev-1`, () => {
        borrados += 1;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    renderPage();
    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Eliminar' }));
    const confirm = screen.getByRole('dialog');
    expect(
      within(confirm).getByText(
        '"Vigilia de viernes" se eliminará para siempre. Esto no se puede deshacer.',
      ),
    ).toBeInTheDocument();

    await userEvent.click(within(confirm).getByRole('button', { name: 'Eliminar' }));

    await waitFor(() => expect(borrados).toBe(1));
    expect(await screen.findByText('Eliminado.')).toBeInTheDocument();
  });
});
