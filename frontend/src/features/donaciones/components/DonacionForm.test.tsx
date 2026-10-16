import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { DonacionAdmin } from '../../../api/ofrendas';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { DonacionForm } from './DonacionForm';

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

function renderForm(props: Partial<Parameters<typeof DonacionForm>[0]> = {}) {
  const onDone = vi.fn();
  const onCancel = vi.fn();
  render(
    <AppProviders>
      <DonacionForm kind="iban" onDone={onDone} onCancel={onCancel} {...props} />
    </AppProviders>,
  );
  return { onDone, onCancel };
}

const dialog = () => screen.getByRole('dialog');

describe('DonacionForm · tipo fijo y campos por tipo (Q2/D-9)', () => {
  it('en SINPE no existe el campo Moneda ni un selector de tipo', () => {
    renderForm({ kind: 'sinpe' });

    expect(within(dialog()).getByText('SINPE Móvil')).toBeInTheDocument();
    expect(within(dialog()).queryByLabelText(/Moneda/)).toBeNull();
    expect(
      within(dialog()).getByLabelText(/Número de celular \(SINPE Móvil\)/),
    ).toBeInTheDocument();
    // Sin selector de tipo: el tipo es una etiqueta fija (D-9).
    expect(within(dialog()).queryAllByRole('combobox')).toHaveLength(0);
  });

  it('en IBAN la moneda es un Select obligatorio y el número se rotula como cuenta', () => {
    renderForm({ kind: 'iban' });

    expect(within(dialog()).getByText('Cuenta IBAN')).toBeInTheDocument();
    expect(within(dialog()).getByLabelText(/Moneda/)).toBeInTheDocument();
    expect(within(dialog()).getByLabelText(/Número de cuenta \(IBAN\)/)).toBeInTheDocument();
  });
});

describe('DonacionForm · validación espejo (FR-014/FR-015/FR-016)', () => {
  it('exige la moneda en IBAN con su mensaje', async () => {
    renderForm({ kind: 'iban' });

    await userEvent.type(within(dialog()).getByLabelText(/Entidad financiera/), 'Banco Nacional');
    await userEvent.type(
      within(dialog()).getByLabelText(/Número de cuenta \(IBAN\)/),
      'CR05015202001234567890',
    );
    await userEvent.click(within(dialog()).getByRole('button', { name: 'Agregar cuenta' }));

    expect(
      await within(dialog()).findAllByText('Elige la moneda de la cuenta: colones o dólares.'),
    ).not.toHaveLength(0);
  });

  it('acepta un IBAN con espacios y lo envía tal cual (normaliza el servidor)', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    server.use(
      http.post(adminUrl, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeDonacion(), { status: 201 });
      }),
    );

    renderForm({ kind: 'iban' });
    await userEvent.type(within(dialog()).getByLabelText(/Entidad financiera/), 'Banco Nacional');
    await userEvent.type(
      within(dialog()).getByLabelText(/Número de cuenta \(IBAN\)/),
      'CR05 0152 0200 1234 5678 90',
    );
    await userEvent.selectOptions(within(dialog()).getByLabelText(/Moneda/), 'CRC');
    await userEvent.click(within(dialog()).getByRole('button', { name: 'Agregar cuenta' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(cuerpo).toMatchObject({
      kind: 'iban',
      number: 'CR05 0152 0200 1234 5678 90',
      institution: 'Banco Nacional',
      currency: 'CRC',
    });
  });

  it('rechaza un IBAN mal formado con el error del campo', async () => {
    renderForm({ kind: 'iban' });
    await userEvent.type(within(dialog()).getByLabelText(/Entidad financiera/), 'Banco');
    await userEvent.type(within(dialog()).getByLabelText(/Número de cuenta \(IBAN\)/), 'CR1234');
    await userEvent.selectOptions(within(dialog()).getByLabelText(/Moneda/), 'CRC');
    await userEvent.click(within(dialog()).getByRole('button', { name: 'Agregar cuenta' }));

    expect(
      await within(dialog()).findAllByText(/no parece una cuenta IBAN válida/),
    ).not.toHaveLength(0);
  });

  it('rechaza un SINPE que no es celular costarricense', async () => {
    renderForm({ kind: 'sinpe' });
    await userEvent.type(within(dialog()).getByLabelText(/Entidad financiera/), 'Banco');
    await userEvent.type(
      within(dialog()).getByLabelText(/Número de celular \(SINPE Móvil\)/),
      '12345',
    );
    await userEvent.click(within(dialog()).getByRole('button', { name: 'Agregar número' }));

    expect(
      await within(dialog()).findAllByText(/no parece un celular de Costa Rica/),
    ).not.toHaveLength(0);
  });

  it('el titular en inglés sin español bloquea indicando qué falta', async () => {
    renderForm({ kind: 'iban' });
    await userEvent.click(within(dialog()).getByRole('button', { name: 'English (opcional)' }));
    await userEvent.type(
      within(dialog()).getByLabelText(/Titular de la cuenta \(English\)/),
      'Simiente Santa',
    );
    await userEvent.type(within(dialog()).getByLabelText(/Entidad financiera/), 'Banco');
    await userEvent.type(
      within(dialog()).getByLabelText(/Número de cuenta \(IBAN\)/),
      'CR05015202001234567890',
    );
    await userEvent.selectOptions(within(dialog()).getByLabelText(/Moneda/), 'CRC');
    await userEvent.click(within(dialog()).getByRole('button', { name: 'Agregar cuenta' }));

    expect(
      await within(dialog()).findAllByText(/Falta la versión en español de el titular/),
    ).not.toHaveLength(0);
  });

  it('sin titular guarda sin problema (US2 esc. 8)', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    server.use(
      http.post(adminUrl, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeDonacion({ holderEs: null }), { status: 201 });
      }),
    );

    renderForm({ kind: 'iban' });
    await userEvent.type(within(dialog()).getByLabelText(/Entidad financiera/), 'Banco');
    await userEvent.type(
      within(dialog()).getByLabelText(/Número de cuenta \(IBAN\)/),
      'CR05015202001234567890',
    );
    await userEvent.selectOptions(within(dialog()).getByLabelText(/Moneda/), 'CRC');
    await userEvent.click(within(dialog()).getByRole('button', { name: 'Agregar cuenta' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(cuerpo).toMatchObject({ holderEs: null, holderEn: null });
  });

  it('el duplicado (409) se avisa sin perder lo escrito', async () => {
    server.use(
      http.post(adminUrl, () =>
        HttpResponse.json({ error: { code: 'conflict', message: 'Ya existe' } }, { status: 409 }),
      ),
    );

    renderForm({ kind: 'iban' });
    const institution = within(dialog()).getByLabelText(/Entidad financiera/);
    await userEvent.type(institution, 'Banco Nacional');
    await userEvent.type(
      within(dialog()).getByLabelText(/Número de cuenta \(IBAN\)/),
      'CR05015202001234567890',
    );
    await userEvent.selectOptions(within(dialog()).getByLabelText(/Moneda/), 'CRC');
    await userEvent.click(within(dialog()).getByRole('button', { name: 'Agregar cuenta' }));

    expect(
      await within(dialog()).findByText(/Ya existe esa misma cuenta o número/),
    ).toBeInTheDocument();
    expect(institution).toHaveValue('Banco Nacional');
  });
});

describe('DonacionForm · edición (D-9)', () => {
  it('no ofrece selector de tipo y el PATCH conserva el tipo (sin enviarlo)', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    server.use(
      http.patch(`${adminUrl}/${itemId}`, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeDonacion({ institution: 'Otro banco' }));
      }),
    );

    renderForm({ kind: 'iban', donacion: makeDonacion() });

    expect(within(dialog()).getByText('Cuenta IBAN')).toBeInTheDocument();
    expect(within(dialog()).queryAllByRole('combobox')).toHaveLength(1); // solo Moneda.

    const institution = within(dialog()).getByLabelText(/Entidad financiera/);
    await userEvent.clear(institution);
    await userEvent.type(institution, 'Otro banco');
    await userEvent.click(within(dialog()).getByRole('button', { name: 'Guardar cambios' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(cuerpo).toMatchObject({ institution: 'Otro banco' });
    expect(cuerpo).not.toHaveProperty('kind');
  });
});
