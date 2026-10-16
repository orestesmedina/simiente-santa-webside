import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { DonacionTextoAdmin } from '../../../api/ofrendas';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { TextoForm } from './TextoForm';

const textosUrl = `${API_BASE_URL}/api/v1/admin/donaciones/textos`;

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

function renderForm(props: Partial<Parameters<typeof TextoForm>[0]> = {}) {
  const onDone = vi.fn();
  const onCancel = vi.fn();
  render(
    <AppProviders>
      <TextoForm textKey="intro" texto={null} onDone={onDone} onCancel={onCancel} {...props} />
    </AppProviders>,
  );
  return { onDone, onCancel };
}

const dialog = () => screen.getByRole('dialog');

describe('TextoForm · un componente para los dos textos', () => {
  it('sirve para el introductorio y para el de uso', () => {
    const { onDone } = renderForm({ textKey: 'usage', texto: makeTexto({ textKey: 'usage' }) });

    expect(onDone).not.toHaveBeenCalled();
    expect(within(dialog()).getByRole('heading')).toHaveTextContent(
      'Texto de uso de las donaciones',
    );
    expect(within(dialog()).getByLabelText(/Texto \(Español\)/)).toHaveValue(
      'Tu aporte va directo a la obra.',
    );
  });
});

describe('TextoForm · validación (FR-014/Decisión 8)', () => {
  it('el inglés sin español bloquea indicando qué falta', async () => {
    renderForm();

    await userEvent.click(within(dialog()).getByRole('button', { name: 'English (opcional)' }));
    await userEvent.type(within(dialog()).getByLabelText('Texto (English)'), 'Your gift matters.');
    await userEvent.click(within(dialog()).getByRole('button', { name: 'Guardar cambios' }));

    expect(
      (await within(dialog()).findAllByText(/Falta la versión en español/)).length,
    ).toBeGreaterThan(0);
  });

  it('2.001 caracteres lo indica con el límite y bloquea el guardado', async () => {
    renderForm();

    const textarea = within(dialog()).getByLabelText(/Texto \(Español\)/);
    fireEvent.change(textarea, { target: { value: 'a'.repeat(2001) } });

    expect(
      await within(dialog()).findByText('2001 de 2000 caracteres. Recorta 1 para poder guardar.'),
    ).toBeInTheDocument();
    expect(within(dialog()).getByRole('button', { name: 'Guardar cambios' })).toBeDisabled();
  });
});

describe('TextoForm · guardado (Q10/P7-17)', () => {
  it('crea un texto nunca escrito desde «Todavía sin escribir.»', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    server.use(
      http.put(`${textosUrl}/intro`, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeTexto({ textEs: 'Cómo donar.' }));
      }),
    );

    const { onDone } = renderForm({ texto: null });
    await userEvent.type(within(dialog()).getByLabelText(/Texto \(Español\)/), 'Cómo donar.');
    await userEvent.click(within(dialog()).getByRole('button', { name: 'Guardar cambios' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(cuerpo).toMatchObject({
      textEs: 'Cómo donar.',
      textEn: null,
      publicationState: 'draft',
    });
    expect(onDone).toHaveBeenCalledWith({ message: 'Cambios guardados.' });
  });

  it('NO cambia el estado del texto: conserva el publicado (Q10)', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    server.use(
      http.put(`${textosUrl}/intro`, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeTexto({ publicationState: 'published' }));
      }),
    );

    const { onDone } = renderForm({ texto: makeTexto({ publicationState: 'published' }) });

    // El formulario no ofrece publicar/retirar: eso vive en la tarjeta (⟲ D-11).
    expect(within(dialog()).queryByRole('button', { name: /Publicar|Retirar/ })).toBeNull();

    await userEvent.click(within(dialog()).getByRole('button', { name: 'Guardar cambios' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(cuerpo).toMatchObject({ publicationState: 'published' });
    expect(onDone).toHaveBeenCalledWith({
      message: 'Cambios guardados: ya visibles en la sección.',
    });
  });

  it('conserva saltos, tildes y emojis íntegros de ida y vuelta', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    server.use(
      http.put(`${textosUrl}/intro`, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeTexto());
      }),
    );

    const texto = 'Párrafo uno.\n\n¡Dona hoy! 🙏 ñandú';
    renderForm();

    fireEvent.change(within(dialog()).getByLabelText(/Texto \(Español\)/), {
      target: { value: texto },
    });
    await userEvent.click(within(dialog()).getByRole('button', { name: 'Guardar cambios' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(cuerpo?.textEs).toBe(texto);
  });
});
