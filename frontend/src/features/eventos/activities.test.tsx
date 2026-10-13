import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { ActividadAdmin, CategoriaAdmin } from '../../api/agenda';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { ActivitiesPage } from './pages/ActivitiesPage';

const activitiesUrl = `${API_BASE_URL}/api/v1/admin/actividades`;
const categoriesUrl = `${API_BASE_URL}/api/v1/admin/categorias`;

const categoria: CategoriaAdmin = {
  id: 'cat-1',
  nameEs: 'acción social',
  sortOrder: 0,
  createdAt: '2026-10-01T10:00:00Z',
  updatedAt: '2026-10-01T10:00:00Z',
};

function makeActividad(overrides: Partial<ActividadAdmin> = {}): ActividadAdmin {
  return {
    id: 'ac-1',
    titleEs: 'Repartir comida',
    titleEn: null,
    detailEs: 'Acción social del mes',
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
    endTime: '11:00',
    publicationState: 'draft',
    isPast: false,
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function responderActividades(items: ActividadAdmin[]) {
  server.use(
    http.get(activitiesUrl, () =>
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
      <ActivitiesPage />
    </AppProviders>,
  );
}

async function abrirAlta() {
  await userEvent.click((await screen.findAllByRole('button', { name: 'Agregar actividad' }))[0]);
  return screen.getByRole('dialog');
}

async function abrirAltaYCompletar() {
  const dialog = await abrirAlta();
  await userEvent.type(within(dialog).getByLabelText(/^Título \(Español\)/), 'Repartir comida');
  await userEvent.type(within(dialog).getByLabelText(/^Lugar \(Español\)/), 'Comedor comunal');
  await userEvent.type(within(dialog).getByLabelText(/^Detalle \(Español\)/), 'Acción social');
  return dialog;
}

describe('ActivitiesPage del panel · listado (US3, ux.md §4.5)', () => {
  it('muestra una sola fecha, su horario, el estado y la marca «Pasado»', async () => {
    responderCategorias([categoria]);
    responderActividades([
      makeActividad({ publicationState: 'published' }),
      makeActividad({
        id: 'ac-2',
        titleEs: 'Ensayo',
        date: '2026-10-11',
        startTime: undefined,
        endTime: undefined,
        isPast: true,
        publicationState: 'published',
      }),
    ]);

    renderPage();

    const tabla = await screen.findByRole('table');
    expect(within(tabla).getByText('Repartir comida')).toBeInTheDocument();
    expect(within(tabla).getByText('domingo 18 de octubre')).toBeInTheDocument();
    expect(within(tabla).getByText('08:00 – 11:00')).toBeInTheDocument();
    // Una actividad sin horario no dibuja «Hora: —» (⟲ D-12).
    expect(within(tabla).getByText('—')).toBeInTheDocument();
    expect(within(tabla).getAllByText('Publicado')).toHaveLength(2);
    expect(within(tabla).getByText('Pasado')).toBeInTheDocument();
  });

  it('sin actividades muestra el estado vacío con su acción', async () => {
    responderCategorias();
    responderActividades([]);

    renderPage();

    expect(
      await screen.findByText(
        'Todavía no hay actividades. Agrega el primero para que el público lo vea.',
      ),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Agregar actividad' })).toBeInTheDocument();
  });
});

describe('ActivityForm · validación y guardado (FR-011/FR-015)', () => {
  it('crea una actividad con fecha y horario válidos y avisa del borrador', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    responderCategorias([categoria]);
    responderActividades([]);
    server.use(
      http.post(activitiesUrl, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeActividad(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAltaYCompletar();

    fireEvent.change(within(dialog).getByLabelText(/^Fecha\s/), {
      target: { value: '2026-10-18' },
    });
    fireEvent.change(within(dialog).getByLabelText(/^Hora de inicio/), {
      target: { value: '08:00' },
    });
    fireEvent.change(within(dialog).getByLabelText(/^Hora de fin/), { target: { value: '11:00' } });
    await userEvent.selectOptions(within(dialog).getByLabelText(/^Categoría/), 'cat-1');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar actividad' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(cuerpo).toMatchObject({
      titleEs: 'Repartir comida',
      placeEs: 'Comedor comunal',
      detailEs: 'Acción social',
      date: '2026-10-18',
      startTime: '08:00',
      endTime: '11:00',
      categoryId: 'cat-1',
    });
    // El modelo de fecha única no tiene frecuencia ni lista de fechas (Decisión 3).
    expect(cuerpo).not.toHaveProperty('dates');
    expect(cuerpo).not.toHaveProperty('frequencyEs');

    expect(await screen.findByText('Actividad creada.')).toBeInTheDocument();
    expect(screen.getByText(/Quedó en borrador/)).toBeInTheDocument();
  });

  it('exige la fecha y aplica las reglas de horario (fin exige inicio y es posterior)', async () => {
    let peticiones = 0;
    responderCategorias();
    responderActividades([]);
    server.use(
      http.post(activitiesUrl, () => {
        peticiones += 1;
        return HttpResponse.json(makeActividad(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAltaYCompletar();

    fireEvent.change(within(dialog).getByLabelText(/^Hora de fin/), { target: { value: '09:00' } });
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar actividad' }));

    expect(
      (await within(dialog).findAllByText('Escoge la fecha de la actividad.')).length,
    ).toBeGreaterThan(0);
    expect(
      within(dialog).getAllByText('Escribe la hora de inicio para poder poner una hora de fin.')
        .length,
    ).toBeGreaterThan(0);
    expect(peticiones).toBe(0);

    fireEvent.change(within(dialog).getByLabelText(/^Fecha\s/), {
      target: { value: '2026-10-18' },
    });
    fireEvent.change(within(dialog).getByLabelText(/^Hora de inicio/), {
      target: { value: '10:00' },
    });
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar actividad' }));

    expect(
      (await within(dialog).findAllByText('La hora de fin debe ser posterior a la hora de inicio.'))
        .length,
    ).toBeGreaterThan(0);
    expect(peticiones).toBe(0);
  });

  it('el duplicado exacto del módulo muestra el mensaje del `409` sin perder lo escrito', async () => {
    responderCategorias();
    responderActividades([]);
    server.use(
      http.post(activitiesUrl, () =>
        HttpResponse.json(
          { error: { code: 'conflict', message: 'Ya existe una actividad igual' } },
          { status: 409 },
        ),
      ),
    );

    renderPage();
    const dialog = await abrirAltaYCompletar();
    fireEvent.change(within(dialog).getByLabelText(/^Fecha\s/), {
      target: { value: '2026-10-18' },
    });
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar actividad' }));

    expect(
      await within(dialog).findByText(/Ya existe una actividad igual \(mismo título/),
    ).toBeInTheDocument();
    expect(within(dialog).getByLabelText(/^Título \(Español\)/)).toHaveValue('Repartir comida');
  });
});

describe('ActivitiesList · publicar, retirar y eliminar (US4)', () => {
  it('publicar y retirar solo tocan la actividad', async () => {
    const parches: Array<Record<string, unknown>> = [];
    let estado: 'draft' | 'published' = 'draft';
    responderCategorias();
    server.use(
      http.get(activitiesUrl, () =>
        HttpResponse.json({
          items: [makeActividad({ publicationState: estado })],
          total: 1,
          limit: 100,
          offset: 0,
        }),
      ),
      http.patch(`${activitiesUrl}/ac-1`, async ({ request }) => {
        const body = (await request.json()) as Record<string, unknown>;
        parches.push(body);
        estado = body.publicationState === 'published' ? 'published' : 'draft';
        return HttpResponse.json(makeActividad({ publicationState: estado }));
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
    await userEvent.click(within(confirm).getByRole('button', { name: 'Retirar' }));

    await waitFor(() => expect(parches).toHaveLength(2));
    expect(parches[1]).toEqual({ publicationState: 'draft' });
  });

  it('editar una actividad publicada avisa que el cambio ya es visible (FR-014)', async () => {
    responderCategorias();
    responderActividades([makeActividad({ publicationState: 'published' })]);
    server.use(
      http.patch(`${activitiesUrl}/ac-1`, () =>
        HttpResponse.json(makeActividad({ publicationState: 'published' })),
      ),
    );

    renderPage();
    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Editar' }));
    const dialog = screen.getByRole('dialog');
    await userEvent.type(within(dialog).getByLabelText(/^Título \(Español\)/), ' del mes');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Guardar cambios' }));

    expect(
      await screen.findByText('Cambios guardados: ya visibles en la sección.'),
    ).toBeInTheDocument();
  });

  it('eliminar pide confirmación explícita', async () => {
    let borrados = 0;
    responderCategorias();
    responderActividades([makeActividad()]);
    server.use(
      http.delete(`${activitiesUrl}/ac-1`, () => {
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
        '"Repartir comida" se eliminará para siempre. Esto no se puede deshacer.',
      ),
    ).toBeInTheDocument();

    await userEvent.click(within(confirm).getByRole('button', { name: 'Eliminar' }));
    await waitFor(() => expect(borrados).toBe(1));
    expect(await screen.findByText('Eliminado.')).toBeInTheDocument();
  });
});
