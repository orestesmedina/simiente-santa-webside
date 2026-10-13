import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { CategoriaAdmin } from '../../api/agenda';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { CategoriesPage } from './pages/CategoriesPage';

const categoriesUrl = `${API_BASE_URL}/api/v1/admin/categorias`;

function makeCategoria(overrides: Partial<CategoriaAdmin> = {}): CategoriaAdmin {
  return {
    id: 'cat-1',
    nameEs: 'acción social',
    nameEn: 'social action',
    sortOrder: 0,
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function responderCategorias(items: CategoriaAdmin[]) {
  server.use(http.get(categoriesUrl, () => HttpResponse.json({ items })));
}

function renderPage() {
  return render(
    <AppProviders>
      <CategoriesPage />
    </AppProviders>,
  );
}

describe('CategoriesPage · catálogo (Q6/Q7, ux.md §4.6)', () => {
  it('lista los nombres es/en y marca los que no tienen inglés', async () => {
    responderCategorias([
      makeCategoria(),
      makeCategoria({ id: 'cat-2', nameEs: 'otro', nameEn: null, sortOrder: 9 }),
    ]);

    renderPage();

    const tabla = await screen.findByRole('table');
    expect(within(tabla).getByText('acción social')).toBeInTheDocument();
    expect(within(tabla).getByText('social action')).toBeInTheDocument();
    expect(within(tabla).getByText('Sin inglés')).toBeInTheDocument();
    expect(within(tabla).getByText('9')).toBeInTheDocument();
  });

  it('crea una categoría con nombre en español y sin inglés (opcional)', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    responderCategorias([]);
    server.use(
      http.post(categoriesUrl, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeCategoria(), { status: 201 });
      }),
    );

    renderPage();
    await userEvent.click(await screen.findByRole('button', { name: 'Agregar categoría' }));
    const dialog = screen.getByRole('dialog');
    await userEvent.type(within(dialog).getByLabelText(/^Nombre \(Español\)/), 'adoración');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar categoría' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(cuerpo).toEqual({ nameEs: 'adoración', nameEn: null });
    expect(await screen.findByText('Categoría creada.')).toBeInTheDocument();
  });

  it('exige el nombre en español y muestra el `409` de nombre duplicado', async () => {
    responderCategorias([]);
    server.use(
      http.post(categoriesUrl, () =>
        HttpResponse.json(
          { error: { code: 'conflict', message: 'Ya existe una categoría con ese nombre' } },
          { status: 409 },
        ),
      ),
    );

    renderPage();
    await userEvent.click(await screen.findByRole('button', { name: 'Agregar categoría' }));
    const dialog = screen.getByRole('dialog');

    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar categoría' }));
    expect(
      await within(dialog).findByText('Escribe el nombre de la categoría.'),
    ).toBeInTheDocument();

    await userEvent.type(within(dialog).getByLabelText(/^Nombre \(Español\)/), 'acción social');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar categoría' }));

    expect(
      await within(dialog).findByText('Ya existe una categoría con ese nombre.'),
    ).toBeInTheDocument();
  });

  it('editar renombra la categoría y el cambio es inmediato para los elementos', async () => {
    let parche: Record<string, unknown> | undefined;
    responderCategorias([makeCategoria()]);
    server.use(
      http.patch(`${categoriesUrl}/cat-1`, async ({ request }) => {
        parche = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeCategoria({ nameEs: 'ayuda social' }));
      }),
    );

    renderPage();
    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Editar' }));
    const dialog = screen.getByRole('dialog');
    await userEvent.clear(within(dialog).getByLabelText(/^Nombre \(Español\)/));
    await userEvent.type(within(dialog).getByLabelText(/^Nombre \(Español\)/), 'ayuda social');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Guardar cambios' }));

    await waitFor(() => expect(parche).toBeDefined());
    expect(parche?.nameEs).toBe('ayuda social');
    expect(await screen.findByText('Categoría guardada.')).toBeInTheDocument();
  });
});

describe('CategoriesPage · borrado con `409` en uso (⟲ D-16, decisión firme)', () => {
  it('una categoría en uso NO se puede borrar: el diálogo lo explica y solo deja cerrar', async () => {
    let borrados = 0;
    responderCategorias([makeCategoria()]);
    server.use(
      http.delete(`${categoriesUrl}/cat-1`, () => {
        borrados += 1;
        return HttpResponse.json(
          {
            error: {
              code: 'conflict',
              message: 'La categoría está en uso por 4 elementos',
            },
          },
          { status: 409 },
        );
      }),
    );

    renderPage();
    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Eliminar' }));

    const confirm = screen.getByRole('dialog');
    expect(
      within(confirm).getByText(
        '"acción social" se eliminará para siempre. Esto no se puede deshacer.',
      ),
    ).toBeInTheDocument();
    await userEvent.click(within(confirm).getByRole('button', { name: 'Eliminar' }));

    await waitFor(() => expect(borrados).toBe(1));
    const aviso = await screen.findByRole('dialog');
    expect(within(aviso).getByText('No se puede eliminar')).toBeInTheDocument();
    expect(
      within(aviso).getByText('Está en uso por 4 elementos: quítala primero de los elementos.'),
    ).toBeInTheDocument();

    // Nunca ofrece desligarla: las únicas acciones son cerrar (⟲ D-16).
    expect(
      within(aviso)
        .getAllByRole('button')
        .map((button) => button.textContent),
    ).toEqual(['Cerrar', 'Cancelar']);
    expect(within(aviso).queryByRole('button', { name: 'Eliminar' })).toBeNull();

    // El elemento sigue en el listado: nada se modificó.
    expect(within(tabla).getByText('acción social')).toBeInTheDocument();
  });

  it('sin uso, el borrado procede tras la confirmación', async () => {
    let borrados = 0;
    responderCategorias([makeCategoria()]);
    server.use(
      http.delete(`${categoriesUrl}/cat-1`, () => {
        borrados += 1;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    renderPage();
    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Eliminar' }));
    const confirm = screen.getByRole('dialog');
    await userEvent.click(within(confirm).getByRole('button', { name: 'Eliminar' }));

    await waitFor(() => expect(borrados).toBe(1));
    expect(await screen.findByText('Eliminado.')).toBeInTheDocument();
  });
});
