import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { DonacionAdmin } from '../../../api/ofrendas';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { DonacionesListsByType } from './DonacionesListsByType';

const adminUrl = `${API_BASE_URL}/api/v1/admin/donaciones`;
const itemId = '11111111-1111-1111-1111-111111111111';

function makeDonacion(overrides: Partial<DonacionAdmin> = {}): DonacionAdmin {
  return {
    id: itemId,
    kind: 'iban',
    number: 'CR05015202001234567890',
    institution: 'Banco Nacional de Costa Rica',
    holderEs: 'Simiente Santa',
    holderEn: null,
    currency: 'CRC',
    publicationState: 'draft',
    sortOrder: 0,
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function renderList(props: Parameters<typeof DonacionesListsByType>[0]) {
  return render(
    <AppProviders>
      <DonacionesListsByType {...props} />
    </AppProviders>,
  );
}

describe('DonacionesListsByType (ux.md §4.3/§5.b, Q1/Q2)', () => {
  it('dibuja los dos bloques y la columna Moneda solo en IBAN (Q2)', () => {
    renderList({
      donaciones: [
        makeDonacion({ id: 'i-1', kind: 'iban' }),
        makeDonacion({
          id: 's-1',
          kind: 'sinpe',
          number: '+50688887777',
          currency: null,
          holderEs: null,
          institution: 'Banco Promerica',
        }),
      ],
    });

    expect(screen.getByRole('heading', { name: 'Cuentas IBAN' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Números de SINPE Móvil' })).toBeInTheDocument();

    const [ibanTable, sinpeTable] = screen.getAllByRole('table');
    expect(within(ibanTable).getByRole('columnheader', { name: 'Moneda' })).toBeInTheDocument();
    expect(within(sinpeTable).queryByRole('columnheader', { name: 'Moneda' })).toBeNull();
    // El titular vacío del SINPE se muestra con «—» (⟲ D-16).
    expect(within(sinpeTable).getAllByText('—').length).toBeGreaterThan(0);
  });

  it('sin cuentas IBAN no hay columna Moneda', () => {
    renderList({
      donaciones: [makeDonacion({ kind: 'sinpe', number: '+50688887777', currency: null })],
    });

    expect(screen.queryByText('Moneda')).toBeNull();
  });

  it('publica directamente con su aviso (US3)', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    server.use(
      http.patch(`${adminUrl}/${itemId}`, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeDonacion({ publicationState: 'published' }));
      }),
    );

    renderList({ donaciones: [makeDonacion()] });
    const tabla = screen.getByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Publicar' }));

    await waitFor(() => expect(cuerpo).toEqual({ publicationState: 'published' }));
    expect(await screen.findByText('Publicado en "Ofrendas y diezmos".')).toBeInTheDocument();
  });

  it('retira con confirmación suave y su aviso (US3)', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    server.use(
      http.patch(`${adminUrl}/${itemId}`, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeDonacion({ publicationState: 'draft' }));
      }),
    );

    renderList({ donaciones: [makeDonacion({ publicationState: 'published' })] });
    const tabla = screen.getByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Retirar de la sección' }));

    const confirm = await screen.findByRole('dialog');
    expect(confirm).toHaveTextContent(
      '"Banco Nacional de Costa Rica" dejará de ser visible al público',
    );
    await userEvent.click(within(confirm).getByRole('button', { name: 'Retirar' }));

    await waitFor(() => expect(cuerpo).toEqual({ publicationState: 'draft' }));
    expect(await screen.findByText(/Retirado de la sección\./)).toBeInTheDocument();
  });

  it('eliminar pide confirmación explícita y avisa (FR-017)', async () => {
    let borrado = false;
    server.use(
      http.delete(`${adminUrl}/${itemId}`, () => {
        borrado = true;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    renderList({ donaciones: [makeDonacion()] });
    const tabla = screen.getByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Eliminar' }));

    const confirm = await screen.findByRole('dialog');
    expect(confirm).toHaveTextContent('se eliminará para siempre');
    await userEvent.click(within(confirm).getByRole('button', { name: 'Eliminar' }));

    await waitFor(() => expect(borrado).toBe(true));
    expect(await screen.findByText('Eliminado.')).toBeInTheDocument();
  });

  it('muestra el estado de carga por bloque', () => {
    renderList({ donaciones: [], loading: true });
    expect(screen.getAllByText('Cargando las cuentas…')).toHaveLength(2);
  });

  it('ante un error ofrece «Reintentar» en cada bloque', async () => {
    const onRetry = vi.fn();
    renderList({ donaciones: [], error: true, onRetry });

    const retries = screen.getAllByRole('button', { name: 'Reintentar' });
    expect(retries).toHaveLength(2);
    await userEvent.click(retries[0]);
    expect(onRetry).toHaveBeenCalled();
  });

  it('el vacío de cada bloque trae su botón de agregar (ux.md §4.6)', () => {
    renderList({ donaciones: [] });

    expect(
      screen.getByText(
        'Todavía no hay cuentas IBAN. Agrega la primera para que el público la vea.',
      ),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Agregar cuenta' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Agregar número' })).toBeInTheDocument();
  });
});
