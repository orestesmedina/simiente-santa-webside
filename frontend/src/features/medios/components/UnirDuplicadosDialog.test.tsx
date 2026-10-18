import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { MedioAdmin } from '../../../api/medios';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { UnirDuplicadosDialog } from './UnirDuplicadosDialog';

/**
 * Dialog «Revisar duplicado» (ux.md §4.6, US3 esc. 1/2): las dos candidatas lado
 * a lado, la **vista previa** del resultado —con el nombre de YouTube y los dos
 * enlaces juntos—, la **doble confirmación** de la unión (irreversible) y «No son
 * el mismo sermón» sin confirmación extra.
 */

const adminUrl = `${API_BASE_URL}/api/v1/admin/medios`;
const TITULO_VIDEO = 'Potencial sin Resultados La Lección de la Higuera';
const TITULO_AUDIO = 'Potencial sin Resultados: La Lección de la Higuera';

/** Base del par: la entrada del **video** (default revisable #7). */
function makeVideo(overrides: Partial<MedioAdmin> = {}): MedioAdmin {
  return {
    id: 'med-1',
    title: TITULO_VIDEO,
    descriptionEs: 'La higuera estéril.',
    descriptionEn: null,
    descriptionEdited: false,
    publishedAt: '2026-05-25T18:43:42Z',
    coverUrl: '/api/v1/media/img_higuera.jpg',
    videoUrl: 'https://www.youtube.com/watch?v=abc12345678',
    origin: 'video',
    visibility: 'visible',
    duplicateOfId: 'med-2',
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

/** La otra ficha del par: la del podcast, que entró **oculta** (FR-014). */
function makeAudio(overrides: Partial<MedioAdmin> = {}): MedioAdmin {
  return {
    id: 'med-2',
    title: TITULO_AUDIO,
    descriptionEs: 'La higuera estéril.',
    descriptionEn: null,
    descriptionEdited: false,
    publishedAt: '2026-05-25T18:43:42Z',
    audioUrl: 'https://podcasters.spotify.com/pod/show/simiente/episodes/higuera',
    origin: 'audio',
    visibility: 'hidden',
    duplicateOfId: 'med-1',
    sources: [
      {
        provider: 'podcast',
        externalId: 'ep-higuera',
        feedTitle: 'Servicio dominical Iglesia Simiente Santa',
        url: 'https://podcasters.spotify.com/pod/show/simiente/episodes/higuera',
      },
    ],
    createdAt: '2026-10-18T20:34:18Z',
    updatedAt: '2026-10-18T20:34:18Z',
    ...overrides,
  };
}

function renderDialog(item = makeVideo(), pareja = makeAudio()) {
  const onDone = vi.fn();
  const onCancel = vi.fn();
  render(
    <AppProviders>
      <UnirDuplicadosDialog item={item} pareja={pareja} onDone={onDone} onCancel={onCancel} />
    </AppProviders>,
  );
  return { onDone, onCancel };
}

describe('UnirDuplicadosDialog · decisión antes que sorpresa (US3 esc. 1)', () => {
  it('muestra las dos candidatas y la vista previa con el nombre de YouTube', () => {
    renderDialog();

    const dialogo = screen.getByRole('dialog');

    // Las dos candidatas, cada una con su estado de visibilidad (FR-014 en pantalla).
    expect(within(dialogo).getByText(TITULO_AUDIO)).toBeInTheDocument();
    expect(within(dialogo).getAllByText(TITULO_VIDEO).length).toBeGreaterThanOrEqual(1);
    expect(within(dialogo).getByText('Visible')).toBeInTheDocument();
    expect(within(dialogo).getByText('Oculto')).toBeInTheDocument();
    expect(within(dialogo).getByText('De YouTube')).toBeInTheDocument();
    expect(within(dialogo).getByText('Del podcast')).toBeInTheDocument();

    // La vista previa: qué quedará, con el nombre que manda y los dos enlaces juntos.
    expect(within(dialogo).getByText('Así quedaría la ficha si las unes.')).toBeInTheDocument();
    expect(
      within(dialogo).getByText('Al unir, el nombre publicado es el de YouTube.'),
    ).toBeInTheDocument();
    expect(
      within(dialogo).getAllByRole('link', { name: 'https://www.youtube.com/watch?v=abc12345678' })
        .length,
    ).toBeGreaterThanOrEqual(1);
    expect(
      within(dialogo).getAllByRole('link', {
        name: 'https://podcasters.spotify.com/pod/show/simiente/episodes/higuera',
      }).length,
    ).toBeGreaterThanOrEqual(1);
  });

  it('la unión exige la confirmación danger y envía la ficha del video como base', async () => {
    const enviados: string[] = [];
    server.use(
      http.post(`${adminUrl}/med-1/unir`, async ({ request }) => {
        enviados.push(await request.text());
        return HttpResponse.json(makeVideo());
      }),
    );

    const { onDone } = renderDialog();

    await userEvent.click(screen.getByRole('button', { name: 'Unir' }));

    // Doble paso: primero la confirmación explícita de lo irreversible (⟲ D-26).
    expect(
      screen.getByText('Se unirán las dos entradas en una sola ficha. Esto no se puede deshacer.'),
    ).toBeInTheDocument();
    expect(enviados).toHaveLength(0);

    await userEvent.click(screen.getByRole('button', { name: 'Unir definitivamente' }));

    expect(enviados).toEqual(['{"otherId":"med-2"}']);
    expect(onDone).toHaveBeenCalledWith({
      message: 'Unidas en una sola ficha, visible en “Medios”.',
    });
  });

  it('si la fila que se mira es la del podcast, la base sigue siendo la del video', async () => {
    const enviados: string[] = [];
    server.use(
      http.post(`${adminUrl}/med-1/unir`, async ({ request }) => {
        enviados.push(await request.text());
        return HttpResponse.json(makeVideo());
      }),
    );

    renderDialog(makeAudio(), makeVideo());
    await userEvent.click(screen.getByRole('button', { name: 'Unir' }));
    await userEvent.click(screen.getByRole('button', { name: 'Unir definitivamente' }));

    // `{id}` es la del video (nombre, carátula y fecha) y la absorbida, la otra.
    expect(enviados).toEqual(['{"otherId":"med-2"}']);
  });

  it('«No son el mismo sermón» resuelve la marca del par sin confirmación extra (⟲ D-27)', async () => {
    const enviados: string[] = [];
    server.use(
      http.post(`${adminUrl}/med-1/marcas/resolver`, async ({ request }) => {
        enviados.push(await request.text());
        return HttpResponse.json(makeVideo({ duplicateMarkResolvedAt: '2026-10-18T21:00:00Z' }));
      }),
    );

    const { onDone } = renderDialog();

    await userEvent.click(screen.getByRole('button', { name: 'No son el mismo sermón' }));

    // Directo: sin `ConfirmDialog` de por medio y con `mark: possible_duplicate`.
    expect(screen.queryByText(/Esto no se puede deshacer/)).toBeNull();
    expect(enviados).toEqual(['{"mark":"possible_duplicate","visibility":"visible"}']);
    expect(onDone).toHaveBeenCalledWith({
      message: 'Listo: quedaron las dos como episodios distintos, visibles sin marca.',
    });
  });

  it('«Cancelar» cierra sin llamar al servidor', async () => {
    const { onCancel } = renderDialog();

    await userEvent.click(screen.getByRole('button', { name: 'Cancelar' }));

    expect(onCancel).toHaveBeenCalled();
  });

  it('un fallo del servidor no cierra el diálogo: avisa y deja reintentar', async () => {
    server.use(
      http.post(`${adminUrl}/med-1/unir`, () =>
        HttpResponse.json({ error: { code: 'internal', message: 'Error' } }, { status: 500 }),
      ),
    );

    const { onDone } = renderDialog();
    await userEvent.click(screen.getByRole('button', { name: 'Unir' }));
    await userEvent.click(screen.getByRole('button', { name: 'Unir definitivamente' }));

    expect(
      await screen.findByText(
        'No se pudo completar la operación. Tus episodios están a salvo: vuelve a intentarlo en unos minutos.',
      ),
    ).toBeInTheDocument();
    expect(onDone).not.toHaveBeenCalled();
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('una ficha sin carátula usa el marcador neutro en la candidata y en la previa', () => {
    renderDialog(makeVideo({ coverUrl: undefined }), makeAudio({ coverUrl: undefined }));

    expect(screen.getAllByTestId('candidata-sin-caratula')).toHaveLength(2);
    expect(screen.getByTestId('previa-sin-caratula')).toBeInTheDocument();
    expect(within(screen.getByRole('dialog')).getAllByText(TITULO_VIDEO).length).toBeGreaterThan(0);
  });
});
