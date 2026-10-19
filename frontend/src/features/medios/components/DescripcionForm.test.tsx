import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { MedioAdmin } from '../../../api/medios';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { DescripcionForm } from './DescripcionForm';

/**
 * Dialog «Editar descripción» (ux.md §4.5, US3/FR-020/FR-023/FR-025): valor
 * vigente precargado, pestañas es/en, contador con el límite, regla del inglés
 * sin español, errores del servidor por campo y **la visibilidad intacta**
 * (⟲ D-24).
 */

const patchUrl = `${API_BASE_URL}/api/v1/admin/medios/med-1`;
const TEXTO_VIGENTE = 'La higuera estéril y lo que Dios espera de nosotros.';

function makeMedio(overrides: Partial<MedioAdmin> = {}): MedioAdmin {
  return {
    id: 'med-1',
    title: 'Potencial sin Resultados La Lección de la Higuera',
    descriptionEs: TEXTO_VIGENTE,
    descriptionEn: null,
    descriptionEdited: false,
    publishedAt: '2026-05-25T18:43:42Z',
    coverUrl: '/api/v1/media/img_higuera.jpg',
    videoUrl: 'https://www.youtube.com/watch?v=abc12345678',
    origin: 'video',
    visibility: 'visible',
    sources: [],
    createdAt: '2026-10-18T20:34:17Z',
    updatedAt: '2026-10-18T20:34:17Z',
    ...overrides,
  };
}

function renderForm(item = makeMedio(), onDone = () => {}, onCancel = () => {}) {
  return render(
    <AppProviders>
      <DescripcionForm item={item} onDone={onDone} onCancel={onCancel} />
    </AppProviders>,
  );
}

describe('DescripcionForm · edición de la descripción corta (US3, FR-020)', () => {
  it('precarga el valor vigente (también el recortado del feed) y el contexto de la ficha', () => {
    renderForm();

    // Contexto: qué episodio se está ajustando (⟲ D-23).
    expect(screen.getByText('Estás ajustando la descripción de:')).toBeInTheDocument();
    expect(
      screen.getByText('Potencial sin Resultados La Lección de la Higuera'),
    ).toBeInTheDocument();
    expect(screen.getByText('lunes 25 de mayo')).toBeInTheDocument();

    expect(screen.getByLabelText('Español')).toHaveValue(TEXTO_VIGENTE);
    expect(screen.getByText(`${TEXTO_VIGENTE.length}/300`)).toBeInTheDocument();
    // La ayuda del catálogo, tal cual.
    expect(
      screen.getByText(
        'Se muestra tal como la escribas, en la ficha del episodio. La descripción completa del feed no se publica.',
      ),
    ).toBeInTheDocument();
  });

  it('guarda solo las descripciones: la visibilidad no se cambia aquí (⟲ D-24)', async () => {
    const enviados: string[] = [];
    server.use(
      http.patch(patchUrl, async ({ request }) => {
        enviados.push(await request.text());
        return HttpResponse.json(makeMedio({ descriptionEs: 'La higuera estéril.' }));
      }),
    );
    const onDone = vi.fn();

    renderForm(makeMedio(), onDone);
    const campo = screen.getByLabelText('Español');
    await userEvent.clear(campo);
    await userEvent.type(campo, 'La higuera estéril.');
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    await waitFor(() => expect(enviados).toHaveLength(1));
    expect(JSON.parse(enviados[0] as string)).toEqual({
      descriptionEs: 'La higuera estéril.',
      descriptionEn: '',
    });
    // El aviso dice que se publica al guardar (FR-015).
    await waitFor(() =>
      expect(onDone).toHaveBeenCalledWith({
        message: 'Cambios guardados: ya visibles en la sección.',
      }),
    );
  });

  it('en una ficha oculta el aviso no promete publicación (FR-015)', async () => {
    server.use(http.patch(patchUrl, () => HttpResponse.json(makeMedio({ visibility: 'hidden' }))));
    const onDone = vi.fn();

    renderForm(makeMedio({ visibility: 'hidden' }), onDone);
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    await waitFor(() => expect(onDone).toHaveBeenCalledWith({ message: 'Cambios guardados.' }));
  });

  it('el inglés sin español bloquea el guardado y dice qué falta (FR-023/FR-025)', async () => {
    const peticiones: string[] = [];
    server.use(
      http.patch(patchUrl, async ({ request }) => {
        peticiones.push(await request.text());
        return HttpResponse.json(makeMedio());
      }),
    );

    renderForm(makeMedio({ descriptionEs: '' }));
    await userEvent.click(screen.getByRole('button', { name: 'English (opcional)' }));
    expect(
      screen.getByText(
        'Si lo dejas vacío, el público verá este contenido en su versión en español.',
      ),
    ).toBeInTheDocument();
    await userEvent.type(screen.getByLabelText('English (opcional)'), 'The fig tree.');

    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(
      await screen.findAllByText(
        'El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de la descripción corta.',
      ),
    ).not.toHaveLength(0);
    // No se llamó al servidor (la regla es la misma que aplica él).
    expect(peticiones).toHaveLength(0);
    // El error va junto al campo que se rellenó sin su base y el foco se queda
    // en él (patrón F6/F7 de la misma regla, FR-023).
    await waitFor(() => expect(screen.getByLabelText('English (opcional)')).toHaveFocus());
    expect(screen.getByLabelText('English (opcional)')).toHaveAttribute('aria-invalid', 'true');
  });

  it('301 caracteres lo indica con el límite y deshabilita «Guardar» (FR-020)', async () => {
    renderForm(makeMedio({ descriptionEs: '' }));

    const campo = screen.getByLabelText('Español');
    fireEvent.change(campo, { target: { value: 'a'.repeat(301) } });

    expect(
      screen.getByText('301 de 300 caracteres. Recorta 1 para poder guardar.'),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeDisabled();

    // Al recortar, el contador vuelve a la cuenta viva y el botón se habilita.
    fireEvent.change(campo, { target: { value: 'a'.repeat(300) } });
    expect(screen.getByText('300/300')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeEnabled();
  });

  it('el contador vivo cuenta runas: 300 caracteres reales con emojis no desbordan (B1)', () => {
    renderForm(makeMedio({ descriptionEs: '' }));

    // 40 familias (secuencia ZWJ) + 20 letras = 300 runas, 460 unidades UTF-16.
    const conEmojis = '👩‍👩‍👧‍👦'.repeat(40) + 'a'.repeat(20);
    fireEvent.change(screen.getByLabelText('Español'), { target: { value: conEmojis } });

    expect(screen.getByText('300/300')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeEnabled();
  });

  it('un `400` del servidor con `details` se pinta junto a su campo (FR-023)', async () => {
    server.use(
      http.patch(patchUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Datos inválidos',
              details: {
                descriptionEs: 'El contenido en español no puede superar los 300 caracteres.',
              },
            },
          },
          { status: 400 },
        ),
      ),
    );

    renderForm();
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(
      await screen.findAllByText('El contenido en español no puede superar los 300 caracteres.'),
    ).not.toHaveLength(0);
    expect(screen.getByLabelText('Español')).toHaveAttribute('aria-invalid', 'true');
  });

  it('tildes, ñ, ¿¡ y emojis viajan íntegros de ida y vuelta (Edge Case)', async () => {
    const enviados: string[] = [];
    server.use(
      http.patch(patchUrl, async ({ request }) => {
        enviados.push(await request.text());
        return HttpResponse.json(makeMedio());
      }),
    );

    renderForm(makeMedio({ descriptionEs: '' }));
    const texto = '¿Qué espera Dios? ¡Mucho! La lección de la higuera 🌿 — ñandú.';
    fireEvent.change(screen.getByLabelText('Español'), { target: { value: texto } });
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    await waitFor(() => expect(enviados).toHaveLength(1));
    expect(JSON.parse(enviados[0] as string).descriptionEs).toBe(texto);
  });

  it('mientras se guarda el modal no se cierra: `Esc` no pierde lo escrito (⟲ D-18)', async () => {
    server.use(
      http.patch(patchUrl, async () => {
        await delay(80);
        return HttpResponse.json(makeMedio());
      }),
    );
    const onCancel = vi.fn();

    renderForm(makeMedio(), () => {}, onCancel);
    fireEvent.change(screen.getByLabelText('Español'), { target: { value: 'Texto en curso' } });
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));
    await screen.findByRole('button', { name: 'Guardando…' });

    await userEvent.keyboard('{Escape}');

    // El diálogo sigue abierto con el texto y sin cancelar nada.
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByLabelText('Español')).toHaveValue('Texto en curso');
    expect(onCancel).not.toHaveBeenCalled();
  });

  it('«Cancelar» cierra sin enviar nada al servidor', async () => {
    const peticiones: string[] = [];
    server.use(
      http.patch(patchUrl, async ({ request }) => {
        peticiones.push(await request.text());
        return HttpResponse.json(makeMedio());
      }),
    );
    const onCancel = vi.fn();

    renderForm(makeMedio(), () => {}, onCancel);
    fireEvent.change(screen.getByLabelText('Español'), { target: { value: 'otro texto' } });
    await userEvent.click(screen.getByRole('button', { name: 'Cancelar' }));

    expect(onCancel).toHaveBeenCalled();
    expect(peticiones).toHaveLength(0);
  });
});
