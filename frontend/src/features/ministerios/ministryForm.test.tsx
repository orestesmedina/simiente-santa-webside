import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { MinisterioAdmin } from '../../api/ministerios';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { MinistriesPage } from './pages/MinistriesPage';

/**
 * Formulario del módulo «Ministerios» (ux.md §4.4/§4.5, US2): alta/edición en
 * `Dialog` con pestañas de idioma, bloque Encargado (`ContactoEncargado` +
 * `ShowContactSwitch`) **apagado al crear** (D-13), enlace `https` estricto y
 * errores junto al campo. Nada se guarda a medias y nada se pierde al fallar.
 */

const listUrl = `${API_BASE_URL}/api/v1/admin/ministerios`;

function makeMinisterio(overrides: Partial<MinisterioAdmin> = {}): MinisterioAdmin {
  return {
    id: 'min-1',
    nameEs: 'Alabanza',
    nameEn: null,
    descriptionEs: 'Llevamos la adoración en los cultos.',
    descriptionEn: null,
    leaderNameEs: null,
    leaderNameEn: null,
    leaderPhone: null,
    leaderEmail: null,
    leaderVisible: false,
    // El contrato (analyze I1) admite `null` y el panel recibe la clave siempre:
    // `null` = sin enlace.
    whatsappUrl: null,
    whatsappUrlShared: false,
    scheduleEs: null,
    scheduleEn: null,
    placeEs: null,
    placeEn: null,
    imageFile: null,
    imageAltEs: null,
    imageAltEn: null,
    publicationState: 'draft',
    sortOrder: 0,
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function responderLista(items: MinisterioAdmin[] = []) {
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

/** Abre el `Dialog` de alta. */
async function abrirAlta() {
  await userEvent.click((await screen.findAllByRole('button', { name: 'Agregar ministerio' }))[0]);
  return screen.getByRole('dialog');
}

/** Rellena los obligatorios del español. */
async function completarObligatorios(
  dialog: HTMLElement,
  name = 'Alabanza',
  description = 'Adoración en los cultos.',
) {
  await userEvent.type(within(dialog).getByLabelText(/^Nombre \(Español\)/), name);
  await userEvent.type(within(dialog).getByLabelText(/^Descripción \(Español\)/), description);
}

describe('MinistryForm · alta con privacidad por defecto (D-13/US2)', () => {
  it('el interruptor nace APAGADO y el alta guarda `leaderVisible: false`', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    responderLista([]);
    server.use(
      http.post(listUrl, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeMinisterio({ leaderNameEs: 'Josué Ramírez' }), {
          status: 201,
        });
      }),
    );

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);

    // El encargado se guarda, pero la ficha no se publica sin un gesto expreso.
    const interruptor = within(dialog).getByRole('switch', {
      name: /Mostrar la ficha del encargado/,
    });
    expect(interruptor).toHaveAttribute('aria-checked', 'false');
    expect(within(dialog).getByText(/La ficha del encargado quedará oculta/)).toBeInTheDocument();

    await userEvent.type(
      within(dialog).getByLabelText(/^Nombre del encargado \(Español\)/),
      'Josué Ramírez',
    );
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(cuerpo).toMatchObject({
      nameEs: 'Alabanza',
      descriptionEs: 'Adoración en los cultos.',
      leaderNameEs: 'Josué Ramírez',
      leaderVisible: false,
    });
    expect(await screen.findByText('Ministerio creado.')).toBeInTheDocument();
    expect(screen.getByText(/Quedó en borrador/)).toBeInTheDocument();
  });

  it('encender y apagar el interruptor no pierde los datos (US2 esc. 10–11)', async () => {
    responderLista([]);

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);
    await userEvent.type(
      within(dialog).getByLabelText(/^Nombre del encargado \(Español\)/),
      'Josué Ramírez',
    );

    const interruptor = within(dialog).getByRole('switch', {
      name: /Mostrar la ficha del encargado/,
    });
    await userEvent.click(interruptor);
    expect(interruptor).toHaveAttribute('aria-checked', 'true');
    expect(within(dialog).getByText(/Se publicará:/)).toBeInTheDocument();

    await userEvent.click(interruptor);
    expect(interruptor).toHaveAttribute('aria-checked', 'false');

    // Los datos siguen ahí: apagar el interruptor solo cambia la publicación.
    expect(within(dialog).getByLabelText(/^Nombre del encargado \(Español\)/)).toHaveValue(
      'Josué Ramírez',
    );
  });

  it('con el interruptor encendido y sin nombre avisa de que falta el nombre', async () => {
    responderLista([]);

    renderPage();
    const dialog = await abrirAlta();

    await userEvent.click(
      within(dialog).getByRole('switch', { name: /Mostrar la ficha del encargado/ }),
    );

    expect(
      within(dialog).getByText('Para publicar la ficha, escribe primero el nombre del encargado.'),
    ).toBeInTheDocument();
  });

  it('editar precarga los datos y la última decisión del interruptor', async () => {
    responderLista([
      makeMinisterio({
        leaderNameEs: 'Ana Rosales',
        leaderVisible: true,
        scheduleEs: 'Todos los viernes, 7:00 p. m.',
      }),
    ]);

    renderPage();

    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Editar' }));

    const dialog = await screen.findByRole('dialog');
    expect(within(dialog).getByLabelText(/^Nombre \(Español\)/)).toHaveValue('Alabanza');
    expect(within(dialog).getByLabelText(/^Horario \(Español/)).toHaveValue(
      'Todos los viernes, 7:00 p. m.',
    );
    expect(within(dialog).getByLabelText(/^Nombre del encargado \(Español\)/)).toHaveValue(
      'Ana Rosales',
    );
    expect(
      within(dialog).getByRole('switch', { name: /Mostrar la ficha del encargado/ }),
    ).toHaveAttribute('aria-checked', 'true');
    expect(within(dialog).getByText(/Se publicará:/)).toBeInTheDocument();
  });
});

describe('MinistryForm · encargado y contacto (Q8)', () => {
  it('un solo medio de contacto: al cambiar de tipo se borra el anterior', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    responderLista([]);
    server.use(
      http.post(listUrl, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeMinisterio(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);
    await userEvent.type(
      within(dialog).getByLabelText(/^Nombre del encargado \(Español\)/),
      'Ana Rosales',
    );

    // Celular primero…
    await userEvent.selectOptions(within(dialog).getByLabelText(/^Medio de contacto/), 'phone');
    await userEvent.type(within(dialog).getByLabelText(/^Número de celular/), '+504 9999-9999');

    // …y al pasar a correo el campo anterior desaparece (nunca los dos, Q8).
    await userEvent.selectOptions(within(dialog).getByLabelText(/^Medio de contacto/), 'email');
    expect(within(dialog).queryByLabelText(/^Número de celular/)).toBeNull();
    await userEvent.type(within(dialog).getByLabelText(/^Correo electrónico/), 'ana@ejemplo.com');

    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    // El medio anterior viaja como `null` (nunca los dos con valor, Q8).
    expect(cuerpo).toMatchObject({ leaderEmail: 'ana@ejemplo.com', leaderPhone: null });
  });

  it('el contacto exige el nombre del encargado (regla del servidor)', async () => {
    let peticiones = 0;
    responderLista([]);
    server.use(
      http.post(listUrl, () => {
        peticiones += 1;
        return HttpResponse.json(makeMinisterio(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);

    await userEvent.selectOptions(within(dialog).getByLabelText(/^Medio de contacto/), 'phone');
    await userEvent.type(within(dialog).getByLabelText(/^Número de celular/), '+504 9999-9999');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    expect(
      (await within(dialog).findAllByText(/Escribe primero el nombre del encargado/)).length,
    ).toBeGreaterThan(0);
    expect(peticiones).toBe(0);
  });

  it('un celular con formato inválido se rechaza sin guardar', async () => {
    let peticiones = 0;
    responderLista([]);
    server.use(
      http.post(listUrl, () => {
        peticiones += 1;
        return HttpResponse.json(makeMinisterio(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);
    await userEvent.type(
      within(dialog).getByLabelText(/^Nombre del encargado \(Español\)/),
      'Ana Rosales',
    );
    await userEvent.selectOptions(within(dialog).getByLabelText(/^Medio de contacto/), 'phone');
    await userEvent.type(within(dialog).getByLabelText(/^Número de celular/), 'abc');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    expect(
      (await within(dialog).findAllByText(/Ese número no parece un celular/)).length,
    ).toBeGreaterThan(0);
    expect(peticiones).toBe(0);
  });
});

describe('MinistryForm · validación espejo (FR-014/FR-015)', () => {
  it('rechaza un enlace no seguro (`javascript:`) sin guardar (SC-009)', async () => {
    let peticiones = 0;
    responderLista([]);
    server.use(
      http.post(listUrl, () => {
        peticiones += 1;
        return HttpResponse.json(makeMinisterio(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);
    await userEvent.type(
      within(dialog).getByLabelText(/^Enlace del grupo de WhatsApp/),
      'javascript:alert(1)',
    );
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    expect(
      (
        await within(dialog).findAllByText(
          'Ese enlace no es válido o no es seguro. Debe empezar por https:// y ser una dirección normal.',
        )
      ).length,
    ).toBeGreaterThan(0);
    expect(peticiones).toBe(0);
  });

  it('el inglés sin su versión en español bloquea indicando qué falta (Q6)', async () => {
    let peticiones = 0;
    responderLista([]);
    server.use(
      http.post(listUrl, () => {
        peticiones += 1;
        return HttpResponse.json(makeMinisterio(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAlta();
    // Solo el nombre en español: la descripción queda vacía (Q6: un campo vacío
    // no obliga a nada) y su versión inglesa es la que bloquea.
    await userEvent.type(within(dialog).getByLabelText(/^Nombre \(Español\)/), 'Alabanza');

    await userEvent.click(within(dialog).getByRole('button', { name: 'English (opcional)' }));
    await userEvent.type(within(dialog).getByLabelText(/^Descripción \(English\)/), 'Worship.');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    expect(
      (await within(dialog).findAllByText(/Falta la versión en español de la descripción/)).length,
    ).toBeGreaterThan(0);
    expect(peticiones).toBe(0);
  });

  it('un campo vacío no obliga a nada: el inglés vacío no bloquea el guardado', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    responderLista([]);
    server.use(
      http.post(listUrl, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeMinisterio(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);
    await userEvent.click(within(dialog).getByRole('button', { name: 'English (opcional)' }));
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    expect(cuerpo).toMatchObject({ nameEn: null, descriptionEn: null, scheduleEn: null });
  });

  it('los campos obligatorios vacíos se señalan y el foco va al primero', async () => {
    responderLista([]);

    renderPage();
    const dialog = await abrirAlta();
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    expect(await within(dialog).findByText('Revisa estos campos:')).toBeInTheDocument();
    expect(within(dialog).getAllByText('Escribe el nombre del ministerio.').length).toBeGreaterThan(
      0,
    );
    await waitFor(() => expect(document.activeElement).toHaveAttribute('id', 'ministry-name-es'));
  });

  it('los contadores están vivos y el límite se explica (FR-019)', async () => {
    responderLista([]);

    renderPage();
    const dialog = await abrirAlta();

    // «0/160» aparece también en el nombre del encargado (mismo límite): se
    // comprueba que el contador del nombre existe y que crece al escribir.
    expect(within(dialog).getAllByText('0/160').length).toBeGreaterThan(0);
    await userEvent.type(within(dialog).getByLabelText(/^Nombre \(Español\)/), 'Alabanza');
    expect(within(dialog).getByText('8/160')).toBeInTheDocument();

    expect(within(dialog).getByText('0/2000')).toBeInTheDocument();
  });

  it('la imagen exige el texto alternativo en español (FR-020)', async () => {
    let peticiones = 0;
    let subidas = 0;
    responderLista([]);
    server.use(
      http.post(listUrl, () => {
        peticiones += 1;
        return HttpResponse.json(makeMinisterio(), { status: 201 });
      }),
      http.post(`${listUrl}/imagenes`, () => {
        subidas += 1;
        return HttpResponse.json({ fileName: 'img_nueva.png' }, { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);

    const grupo = within(dialog).getByRole('group', { name: 'Imagen del ministerio' });
    await userEvent.upload(
      within(grupo).getByLabelText(/^Archivo/),
      new File(['imagen'], 'foto.png', { type: 'image/png' }),
    );
    await waitFor(() => expect(subidas).toBe(1));

    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    expect(
      (await within(dialog).findAllByText('Escribe el texto alternativo de la imagen.')).length,
    ).toBeGreaterThan(0);
    expect(peticiones).toBe(0);
  });

  it('un archivo que no es imagen avisa sin romper el formulario (FR-020)', async () => {
    responderLista([]);

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);

    const grupo = within(dialog).getByRole('group', { name: 'Imagen del ministerio' });
    // `applyAccept: false`: se simula el archivo que el equipo eligió a mano
    // (el `accept` del input no es una garantía; el cliente valida igual).
    await userEvent.upload(
      within(grupo).getByLabelText(/^Archivo/),
      new File(['texto'], 'notas.txt', { type: 'text/plain' }),
      { applyAccept: false },
    );

    expect(
      await within(grupo).findByText(/Ese archivo no es una imagen válida/),
    ).toBeInTheDocument();
    // El resto del formulario queda íntegro.
    expect(within(dialog).getByLabelText(/^Nombre \(Español\)/)).toHaveValue('Alabanza');
    expect(within(dialog).getByLabelText(/^Descripción \(Español\)/)).toHaveValue(
      'Adoración en los cultos.',
    );
  });
});

describe('MinistryForm · enlace de WhatsApp: vacío es `null`, nunca `""` (I1)', () => {
  it('crear sin enlace manda `null` (no manda `""`, que no cumple el `pattern`)', async () => {
    let cuerpo: Record<string, unknown> | undefined;
    responderLista([]);
    server.use(
      http.post(listUrl, async ({ request }) => {
        cuerpo = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeMinisterio(), { status: 201 });
      }),
    );

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    await waitFor(() => expect(cuerpo).toBeDefined());
    // Sin enlace la clave viaja explícita en `null`; nunca como cadena vacía
    // (analyze I1: el contrato declara `type: [string, "null"]` + `pattern`).
    expect(cuerpo).toHaveProperty('whatsappUrl', null);
    expect(cuerpo?.whatsappUrl).not.toBe('');
  });

  it('quitar el enlace al editar manda `null` (borrado explícito), nunca `""`', async () => {
    let parche: Record<string, unknown> | undefined;
    const enlace = 'https://chat.whatsapp.com/ABC123';
    responderLista([makeMinisterio({ whatsappUrl: enlace, publicationState: 'published' })]);
    server.use(
      http.patch(`${listUrl}/min-1`, async ({ request }) => {
        parche = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeMinisterio({ publicationState: 'published' }));
      }),
    );

    renderPage();
    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Editar' }));

    const dialog = await screen.findByRole('dialog');
    const campoEnlace = within(dialog).getByLabelText(/^Enlace del grupo de WhatsApp/);
    expect(campoEnlace).toHaveValue(enlace);

    // El equipo borra el enlace y guarda: el ministerio queda sin enlace, así
    // que el sitio público deja de mostrar el botón de unirse (FR-003/SC-009).
    await userEvent.clear(campoEnlace);
    await userEvent.click(within(dialog).getByRole('button', { name: 'Guardar cambios' }));

    await waitFor(() => expect(parche).toBeDefined());
    expect(parche).toHaveProperty('whatsappUrl', null);
    expect(parche?.whatsappUrl).not.toBe('');
  });

  it('editar sin tocar el enlace lo conserva (no lo borra)', async () => {
    let parche: Record<string, unknown> | undefined;
    const enlace = 'https://chat.whatsapp.com/ABC123';
    responderLista([makeMinisterio({ whatsappUrl: enlace })]);
    server.use(
      http.patch(`${listUrl}/min-1`, async ({ request }) => {
        parche = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeMinisterio({ whatsappUrl: enlace }));
      }),
    );

    renderPage();
    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Editar' }));

    const dialog = await screen.findByRole('dialog');
    fireEvent.change(within(dialog).getByLabelText(/^Nombre \(Español\)/), {
      target: { value: 'Alabanza y adoración' },
    });
    await userEvent.click(within(dialog).getByRole('button', { name: 'Guardar cambios' }));

    await waitFor(() => expect(parche).toBeDefined());
    expect(parche?.whatsappUrl).toBe(enlace);
  });
});

describe('MinistryForm · errores del servidor y avisos (D-7)', () => {
  it('muestra el `409` de duplicado sin perder lo escrito', async () => {
    responderLista([]);
    server.use(
      http.post(listUrl, () =>
        HttpResponse.json(
          { error: { code: 'conflict', message: 'Ya existe un ministerio igual' } },
          { status: 409 },
        ),
      ),
    );

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    expect(
      await within(dialog).findByText(/Ya existe un ministerio igual \(mismo nombre/),
    ).toBeInTheDocument();
    expect(within(dialog).getByLabelText(/^Nombre \(Español\)/)).toHaveValue('Alabanza');
  });

  it('los `details` del servidor se muestran junto al campo', async () => {
    responderLista([]);
    server.use(
      http.post(listUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Datos inválidos',
              details: { placeEs: 'El lugar es demasiado largo' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    // El mensaje sale junto al campo y en el resumen del formulario.
    expect(
      (await within(dialog).findAllByText('El lugar es demasiado largo')).length,
    ).toBeGreaterThan(0);
  });

  it('el enlace compartido es un aviso no bloqueante tras guardar (D-7)', async () => {
    responderLista([]);
    server.use(
      http.post(listUrl, () =>
        HttpResponse.json(makeMinisterio({ whatsappUrlShared: true }), { status: 201 }),
      ),
    );

    renderPage();
    const dialog = await abrirAlta();
    await completarObligatorios(dialog);
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar ministerio' }));

    expect(
      await screen.findByText(
        'Otro ministerio ya usa este mismo enlace de WhatsApp. Revísalo si no es intencional; el ministerio sí se guardó.',
      ),
    ).toBeInTheDocument();
    // El guardado salió adelante: el aviso no bloquea.
    expect(screen.getByText('Ministerio creado.')).toBeInTheDocument();
  });

  it('editar un ministerio publicado avisa que ya es visible (FR-013)', async () => {
    let parche: Record<string, unknown> | undefined;
    responderLista([makeMinisterio({ publicationState: 'published' })]);
    server.use(
      http.patch(`${listUrl}/min-1`, async ({ request }) => {
        parche = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeMinisterio({ publicationState: 'published' }));
      }),
    );

    renderPage();
    const tabla = await screen.findByRole('table');
    await userEvent.click(within(tabla).getByRole('button', { name: 'Editar' }));

    const dialog = await screen.findByRole('dialog');
    fireEvent.change(within(dialog).getByLabelText(/^Nombre \(Español\)/), {
      target: { value: 'Alabanza y adoración' },
    });
    await userEvent.click(within(dialog).getByRole('button', { name: 'Guardar cambios' }));

    await waitFor(() => expect(parche).toBeDefined());
    expect(parche).toMatchObject({ nameEs: 'Alabanza y adoración' });
    expect(
      await screen.findByText('Cambios guardados: ya visibles en la sección.'),
    ).toBeInTheDocument();
  });
});
