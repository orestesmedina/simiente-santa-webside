import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type {
  PortadaAdmin,
  ScheduleItemAdmin,
  SocialLinkAdmin,
  WhatsappChannelAdmin,
} from '../../api/portada';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { ServicesList } from './components/ServicesList';
import { SocialsList } from './components/SocialsList';
import { WhatsAppList } from './components/WhatsAppList';
import {
  DRAFT_NOTICE,
  DELETED,
  EMPTY_SERVICES,
  END_TIME_INVALID,
  PUBLISHED,
  SERVICE_CREATED,
  UNPUBLISHED,
  URL_INVALID,
  WHATSAPP_CREATED,
  WHATSAPP_DUPLICATE,
  WHATSAPP_GROUP_INVALID,
  socialDuplicateMessage,
  socialSavedMessage,
} from './messages';

const adminUrl = `${API_BASE_URL}/api/v1/admin/portada`;
const scheduleUrl = `${adminUrl}/horario`;
const whatsappUrl = `${adminUrl}/whatsapp`;
const socialsUrl = `${adminUrl}/redes`;

function adminData(overrides: Partial<PortadaAdmin> = {}): PortadaAdmin {
  return {
    identity: null,
    about: null,
    contact: null,
    schedule: { items: [] },
    whatsapp: { items: [] },
    socials: { items: [] },
    ...overrides,
  };
}

function makeService(overrides: Partial<ScheduleItemAdmin> = {}): ScheduleItemAdmin {
  return {
    id: 'srv-1',
    dayOfWeek: 0,
    startTime: '10:00',
    endTime: null,
    nameEs: 'Culto dominical',
    placeEs: 'Templo central',
    publicationState: 'draft',
    sortOrder: 0,
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function makeChannel(overrides: Partial<WhatsappChannelAdmin> = {}): WhatsappChannelAdmin {
  return {
    id: 'wa-1',
    nameEs: 'Escríbenos',
    kind: 'direct',
    destination: '+506 8888 8888',
    publicationState: 'draft',
    sortOrder: 0,
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function makeSocial(overrides: Partial<SocialLinkAdmin> = {}): SocialLinkAdmin {
  return {
    id: 'so-1',
    network: 'facebook',
    url: 'https://facebook.com/simiente',
    publicationState: 'draft',
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function renderList(node: React.ReactElement) {
  return render(<AppProviders>{node}</AppProviders>);
}

describe('ServicesList · horario (FR-004, analyze C2)', () => {
  it('vacío muestra el estado propio y ofrece agregar el primero', async () => {
    server.use(http.get(adminUrl, () => HttpResponse.json(adminData())));

    renderList(<ServicesList />);

    expect(await screen.findByText(EMPTY_SERVICES)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Agregar servicio' })).toBeInTheDocument();
  });

  it('el Select de día ofrece las 7 opciones localizadas', async () => {
    server.use(http.get(adminUrl, () => HttpResponse.json(adminData())));
    renderList(<ServicesList />);
    await screen.findByText(EMPTY_SERVICES);
    await userEvent.click(screen.getByRole('button', { name: 'Agregar servicio' }));

    const dialog = await screen.findByRole('dialog', { name: 'Agregar servicio' });
    const options = within(dialog).getByLabelText(/^Día/).querySelectorAll('option');
    expect(Array.from(options).map((option) => option.textContent)).toEqual([
      'Domingo',
      'Lunes',
      'Martes',
      'Miércoles',
      'Jueves',
      'Viernes',
      'Sábado',
    ]);
  });

  it('el alta válida nace en borrador y lo avisa', async () => {
    let current = adminData();
    let postBody: Record<string, unknown> | undefined;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.post(scheduleUrl, async ({ request }) => {
        postBody = (await request.json()) as Record<string, unknown>;
        current = adminData({ schedule: { items: [makeService()] } });
        return HttpResponse.json(makeService(), { status: 201 });
      }),
    );

    renderList(<ServicesList />);
    await screen.findByText(EMPTY_SERVICES);
    await userEvent.click(screen.getByRole('button', { name: 'Agregar servicio' }));
    const dialog = await screen.findByRole('dialog', { name: 'Agregar servicio' });

    await userEvent.selectOptions(within(dialog).getByLabelText(/^Día/), '0');
    fireEvent.change(within(dialog).getByLabelText(/^Hora de inicio/), {
      target: { value: '10:00' },
    });
    await userEvent.type(within(dialog).getByLabelText(/^Nombre \(Español\)/), 'Culto dominical');
    await userEvent.type(within(dialog).getByLabelText(/^Lugar \(Español\)/), 'Templo central');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar servicio' }));

    expect(await screen.findByText(SERVICE_CREATED)).toBeInTheDocument();
    expect(screen.getByText(DRAFT_NOTICE)).toBeInTheDocument();
    expect(postBody).toMatchObject({
      dayOfWeek: 0,
      startTime: '10:00',
      endTime: null,
      nameEs: 'Culto dominical',
      placeEs: 'Templo central',
    });
    expect(postBody).not.toHaveProperty('publicationState');
  });

  it('una hora de fin anterior a la de inicio se rechaza (C2)', async () => {
    let postCount = 0;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(adminData())),
      http.post(scheduleUrl, () => {
        postCount += 1;
        return HttpResponse.json(makeService(), { status: 201 });
      }),
    );

    renderList(<ServicesList />);
    await screen.findByText(EMPTY_SERVICES);
    await userEvent.click(screen.getByRole('button', { name: 'Agregar servicio' }));
    const dialog = await screen.findByRole('dialog', { name: 'Agregar servicio' });

    fireEvent.change(within(dialog).getByLabelText(/^Hora de inicio/), {
      target: { value: '10:00' },
    });
    fireEvent.change(within(dialog).getByLabelText(/^Hora de fin/), { target: { value: '09:00' } });
    await userEvent.type(within(dialog).getByLabelText(/^Nombre \(Español\)/), 'Culto');
    await userEvent.type(within(dialog).getByLabelText(/^Lugar \(Español\)/), 'Templo');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar servicio' }));

    expect(await within(dialog).findByText(END_TIME_INVALID)).toBeInTheDocument();
    expect(postCount).toBe(0);
  });

  it('publicar, retirar y eliminar por elemento (FR-013)', async () => {
    let current = adminData({ schedule: { items: [makeService()] } });
    const patchBodies: Array<Record<string, unknown>> = [];
    let deleted = false;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.patch(`${scheduleUrl}/srv-1`, async ({ request }) => {
        const body = (await request.json()) as Record<string, unknown>;
        patchBodies.push(body);
        current = adminData({
          schedule: {
            items: [
              makeService({ publicationState: body.publicationState as 'draft' | 'published' }),
            ],
          },
        });
        return HttpResponse.json(makeService());
      }),
      http.delete(`${scheduleUrl}/srv-1`, () => {
        deleted = true;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    renderList(<ServicesList />);
    await userEvent.click((await screen.findAllByRole('button', { name: 'Publicar' }))[0]);

    expect(await screen.findByText(PUBLISHED)).toBeInTheDocument();
    expect(patchBodies[0]).toMatchObject({ publicationState: 'published' });
    expect((await screen.findAllByText('Publicado')).length).toBeGreaterThan(0);

    await userEvent.click(screen.getAllByRole('button', { name: 'Retirar de la portada' })[0]);
    const confirm = await screen.findByRole('dialog', { name: 'Retirar de la portada' });
    await userEvent.click(within(confirm).getByRole('button', { name: 'Retirar' }));
    expect(await screen.findByText(UNPUBLISHED)).toBeInTheDocument();

    await userEvent.click(screen.getAllByRole('button', { name: 'Eliminar' })[0]);
    const deleteDialog = await screen.findByRole('dialog', { name: 'Eliminar' });
    await userEvent.click(within(deleteDialog).getByRole('button', { name: 'Eliminar' }));
    expect(await screen.findByText(DELETED)).toBeInTheDocument();
    await waitFor(() => expect(deleted).toBe(true));
  });
});

describe('WhatsAppList · canales (FR-005)', () => {
  it('el alta de un mensaje directo guarda y avisa', async () => {
    let current = adminData();
    let postBody: Record<string, unknown> | undefined;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.post(whatsappUrl, async ({ request }) => {
        postBody = (await request.json()) as Record<string, unknown>;
        current = adminData({ whatsapp: { items: [makeChannel()] } });
        return HttpResponse.json(makeChannel(), { status: 201 });
      }),
    );

    renderList(<WhatsAppList />);
    await screen.findByText(
      'Todavía no hay canales. Agrega el primero para que aparezca en la portada.',
    );
    await userEvent.click(screen.getByRole('button', { name: 'Agregar canal' }));
    const dialog = await screen.findByRole('dialog', { name: 'Agregar canal' });

    await userEvent.type(
      within(dialog).getByLabelText(/^Nombre o propósito \(Español\)/),
      'Escríbenos',
    );
    await userEvent.type(
      within(dialog).getByRole('textbox', { name: /^Número/ }),
      '+506 8888 8888',
    );
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar canal' }));

    expect(await screen.findByText(WHATSAPP_CREATED)).toBeInTheDocument();
    expect(postBody).toMatchObject({
      nameEs: 'Escríbenos',
      kind: 'direct',
      destination: '+506 8888 8888',
    });
  });

  it('un enlace de grupo mal formado se rechaza antes de enviar', async () => {
    let postCount = 0;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(adminData())),
      http.post(whatsappUrl, () => {
        postCount += 1;
        return HttpResponse.json(makeChannel(), { status: 201 });
      }),
    );

    renderList(<WhatsAppList />);
    await screen.findByText(
      'Todavía no hay canales. Agrega el primero para que aparezca en la portada.',
    );
    await userEvent.click(screen.getByRole('button', { name: 'Agregar canal' }));
    const dialog = await screen.findByRole('dialog', { name: 'Agregar canal' });

    await userEvent.click(within(dialog).getByRole('radio', { name: 'Enlace de grupo' }));
    await userEvent.type(within(dialog).getByLabelText(/^Nombre o propósito \(Español\)/), 'Grupo');
    await userEvent.type(
      within(dialog).getByLabelText(/^Enlace del grupo/),
      'http://otro-sitio.com/x',
    );
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar canal' }));

    expect(await within(dialog).findByText(WHATSAPP_GROUP_INVALID)).toBeInTheDocument();
    expect(postCount).toBe(0);
  });

  it('un canal duplicado (409) se explica con el texto del catálogo', async () => {
    server.use(
      http.get(adminUrl, () => HttpResponse.json(adminData())),
      http.post(whatsappUrl, () =>
        HttpResponse.json({ error: { code: 'conflict', message: 'Ya existe' } }, { status: 409 }),
      ),
    );

    renderList(<WhatsAppList />);
    await screen.findByText(
      'Todavía no hay canales. Agrega el primero para que aparezca en la portada.',
    );
    await userEvent.click(screen.getByRole('button', { name: 'Agregar canal' }));
    const dialog = await screen.findByRole('dialog', { name: 'Agregar canal' });
    await userEvent.type(
      within(dialog).getByLabelText(/^Nombre o propósito \(Español\)/),
      'Escríbenos',
    );
    await userEvent.type(
      within(dialog).getByRole('textbox', { name: /^Número/ }),
      '+506 8888 8888',
    );
    await userEvent.click(within(dialog).getByRole('button', { name: 'Agregar canal' }));

    expect(await within(dialog).findByText(WHATSAPP_DUPLICATE)).toBeInTheDocument();
  });
});

describe('SocialsList · redes (FR-006)', () => {
  it('muestra el catálogo completo con las redes sin enlace', async () => {
    server.use(http.get(adminUrl, () => HttpResponse.json(adminData())));

    renderList(<SocialsList />);

    expect((await screen.findAllByText('Facebook')).length).toBeGreaterThan(0);
    expect(screen.getAllByText('Sin enlace todavía').length).toBeGreaterThanOrEqual(5);
    expect(screen.getAllByRole('button', { name: 'Agregar enlace' }).length).toBeGreaterThanOrEqual(
      5,
    );
  });

  it('agregar el enlace de Facebook lo guarda y avisa', async () => {
    let current = adminData();
    let postBody: Record<string, unknown> | undefined;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.post(socialsUrl, async ({ request }) => {
        postBody = (await request.json()) as Record<string, unknown>;
        current = adminData({ socials: { items: [makeSocial()] } });
        return HttpResponse.json(makeSocial(), { status: 201 });
      }),
    );

    renderList(<SocialsList />);
    await screen.findAllByText('Facebook');
    await userEvent.click(screen.getAllByRole('button', { name: 'Agregar enlace' })[0]);
    const dialog = await screen.findByRole('dialog', { name: 'Agregar enlace de Facebook' });

    await userEvent.type(
      within(dialog).getByLabelText(/^Enlace de Facebook/),
      'https://facebook.com/simiente',
    );
    await userEvent.click(within(dialog).getByRole('button', { name: 'Guardar' }));

    expect(await screen.findByText(socialSavedMessage('Facebook'))).toBeInTheDocument();
    expect(postBody).toMatchObject({ network: 'facebook', url: 'https://facebook.com/simiente' });
  });

  it('un enlace que no es https se rechaza antes de enviar', async () => {
    let postCount = 0;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(adminData())),
      http.post(socialsUrl, () => {
        postCount += 1;
        return HttpResponse.json(makeSocial(), { status: 201 });
      }),
    );

    renderList(<SocialsList />);
    await screen.findAllByText('Facebook');
    await userEvent.click(screen.getAllByRole('button', { name: 'Agregar enlace' })[0]);
    const dialog = await screen.findByRole('dialog', { name: 'Agregar enlace de Facebook' });
    await userEvent.type(
      within(dialog).getByLabelText(/^Enlace de Facebook/),
      'facebook.com/simiente',
    );
    await userEvent.click(within(dialog).getByRole('button', { name: 'Guardar' }));

    expect(await within(dialog).findByText(URL_INVALID)).toBeInTheDocument();
    expect(postCount).toBe(0);
  });

  it('un segundo enlace de la misma red (409) explica la regla', async () => {
    server.use(
      http.get(adminUrl, () => HttpResponse.json(adminData())),
      http.post(socialsUrl, () =>
        HttpResponse.json({ error: { code: 'conflict', message: 'Ya existe' } }, { status: 409 }),
      ),
    );

    renderList(<SocialsList />);
    await screen.findAllByText('Facebook');
    await userEvent.click(screen.getAllByRole('button', { name: 'Agregar enlace' })[0]);
    const dialog = await screen.findByRole('dialog', { name: 'Agregar enlace de Facebook' });
    await userEvent.type(
      within(dialog).getByLabelText(/^Enlace de Facebook/),
      'https://facebook.com/x',
    );
    await userEvent.click(within(dialog).getByRole('button', { name: 'Guardar' }));

    expect(await within(dialog).findByText(socialDuplicateMessage('Facebook'))).toBeInTheDocument();
  });
});
