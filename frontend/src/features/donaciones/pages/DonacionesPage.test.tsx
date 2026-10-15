import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { DonacionAdmin, DonacionTextoAdmin } from '../../../api/ofrendas';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { DonacionesPage } from './DonacionesPage';

const adminUrl = `${API_BASE_URL}/api/v1/admin/donaciones`;
const textosUrl = `${adminUrl}/textos`;

function makeDonacion(overrides: Partial<DonacionAdmin> = {}): DonacionAdmin {
  return {
    id: '11111111-1111-1111-1111-111111111111',
    kind: 'iban',
    number: 'CR05015202001234567890',
    institution: 'Banco Nacional de Costa Rica',
    holderEs: null,
    holderEn: null,
    currency: 'CRC',
    publicationState: 'draft',
    sortOrder: 0,
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function makeTexto(overrides: Partial<DonacionTextoAdmin> = {}): DonacionTextoAdmin {
  return {
    textKey: 'intro',
    textEs: 'Tu aporte va directo a la obra.',
    textEn: null,
    publicationState: 'draft',
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

/** Responde el listado de métodos y los textos con los valores indicados. */
function respond(
  items: DonacionAdmin[],
  intro: DonacionTextoAdmin | null,
  usage: DonacionTextoAdmin | null,
) {
  server.use(
    http.get(adminUrl, () =>
      HttpResponse.json({ items, total: items.length, limit: 100, offset: 0 }),
    ),
    http.get(textosUrl, () => HttpResponse.json({ intro, usage })),
  );
}

function renderPage() {
  return render(
    <AppProviders>
      <DonacionesPage />
    </AppProviders>,
  );
}

describe('DonacionesPage (ux.md §4.3/§5, US2/US3)', () => {
  it('muestra título, banda de ayuda y las dos zonas', async () => {
    respond([], null, null);

    renderPage();

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Gestión de donaciones' }),
    ).toBeInTheDocument();
    expect(screen.getByText(/Aquí administras las cuentas y números/)).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Textos de la sección' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Cuentas IBAN' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Números de SINPE Móvil' })).toBeInTheDocument();
  });

  it('un texto nunca escrito muestra «Todavía sin escribir.» con su acción (⟲ D-8)', async () => {
    respond([], null, null);

    renderPage();

    expect(await screen.findAllByText('Todavía sin escribir.')).toHaveLength(2);
    // Cada tarjeta de texto tiene su botón Editar (no es un hueco sin salida).
    expect(screen.getAllByRole('button', { name: 'Editar' })).toHaveLength(2);
  });

  it('un texto guardado muestra su estado propio y su contenido (Q10)', async () => {
    respond(
      [],
      makeTexto({ publicationState: 'published' }),
      makeTexto({ textKey: 'usage', textEs: 'Se usa en la obra social.' }),
    );

    renderPage();

    expect(await screen.findByText('Tu aporte va directo a la obra.')).toBeInTheDocument();
    expect(screen.getByText('Se usa en la obra social.')).toBeInTheDocument();
    expect(screen.getByText('Publicado')).toBeInTheDocument();
    expect(screen.getByText('Borrador')).toBeInTheDocument();
  });

  it('avisa cuando nada está publicado (métodos y textos en borrador)', async () => {
    respond(
      [makeDonacion({ publicationState: 'draft' })],
      makeTexto(),
      makeTexto({ textKey: 'usage' }),
    );

    renderPage();

    expect(await screen.findByText(/Nada está publicado ahora/)).toBeInTheDocument();
  });

  it('no avisa cuando hay algo publicado', async () => {
    respond(
      [makeDonacion({ publicationState: 'published' })],
      makeTexto(),
      makeTexto({ textKey: 'usage' }),
    );

    renderPage();

    await screen.findByRole('heading', { name: 'Cuentas IBAN' });
    expect(screen.queryByText(/Nada está publicado ahora/)).toBeNull();
  });

  it('la zona de textos muestra su error con «Reintentar»', async () => {
    let intentos = 0;
    server.use(
      http.get(adminUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
      http.get(textosUrl, () => {
        intentos += 1;
        if (intentos === 1) {
          return HttpResponse.json(
            { error: { code: 'internal', message: 'boom' } },
            { status: 500 },
          );
        }
        return HttpResponse.json({ intro: null, usage: null });
      }),
    );

    renderPage();

    const retry = await screen.findByRole('button', { name: 'Reintentar' });
    await userEvent.click(retry);
    expect(await screen.findAllByText('Todavía sin escribir.')).toHaveLength(2);
  });

  it('sin permiso avisa con el texto de F2 (FR-011)', async () => {
    server.use(
      http.get(adminUrl, () =>
        HttpResponse.json({ error: { code: 'forbidden', message: 'no' } }, { status: 403 }),
      ),
      http.get(textosUrl, () =>
        HttpResponse.json({ error: { code: 'forbidden', message: 'no' } }, { status: 403 }),
      ),
    );

    renderPage();

    expect(await screen.findByText(/No tienes acceso a esta sección/)).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Cuentas IBAN' })).toBeNull();
  });

  it('editar un texto abre su formulario y guarda con el aviso de su estado (Q10)', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    server.use(
      http.get(adminUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
      http.get(textosUrl, () => HttpResponse.json({ intro: makeTexto(), usage: null })),
      http.put(`${textosUrl}/intro`, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeTexto());
      }),
    );

    renderPage();

    const editButtons = await screen.findAllByRole('button', { name: 'Editar' });
    await userEvent.click(editButtons[0]);

    const form = await screen.findByRole('dialog');
    expect(within(form).getByRole('heading')).toHaveTextContent('Texto introductorio');
    await userEvent.click(within(form).getByRole('button', { name: 'Guardar cambios' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(await screen.findByText('Cambios guardados.')).toBeInTheDocument();
  });

  it('publicar un texto desde su tarjeta llama al PUT y avisa (Q10)', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    server.use(
      http.get(adminUrl, () => HttpResponse.json({ items: [], total: 0, limit: 100, offset: 0 })),
      http.get(textosUrl, () => HttpResponse.json({ intro: makeTexto(), usage: null })),
      http.put(`${textosUrl}/intro`, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeTexto({ publicationState: 'published' }));
      }),
    );

    renderPage();

    await userEvent.click(await screen.findByRole('button', { name: 'Publicar' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(cuerpo).toMatchObject({ publicationState: 'published' });
    expect(await screen.findByText('Publicado en "Ofrendas y diezmos".')).toBeInTheDocument();
  });
});
