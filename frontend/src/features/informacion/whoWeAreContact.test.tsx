import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { AboutAdmin, ContactAdmin, PortadaAdmin } from '../../api/portada';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { ContactForm } from './components/ContactForm';
import { WhoWeAreForm } from './components/WhoWeAreForm';
import {
  DRAFT_NOTICE,
  EMAIL_INVALID,
  PHONE_INVALID,
  PUBLISHED,
  REQUIRED_ABOUT,
  SAVED,
  UNPUBLISHED,
  aboutLimitMessage,
} from './messages';

const adminUrl = `${API_BASE_URL}/api/v1/admin/portada`;
const aboutUrl = `${adminUrl}/quienes-somos`;
const contactUrl = `${adminUrl}/contacto`;

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

function makeAbout(overrides: Partial<AboutAdmin> = {}): AboutAdmin {
  return {
    textEs: 'Somos una familia',
    publicationState: 'draft',
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function makeContact(overrides: Partial<ContactAdmin> = {}): ContactAdmin {
  return {
    addressEs: 'San José',
    email: 'hola@ejemplo.com',
    phone: '+506 8888 8888',
    publicationState: 'draft',
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

describe('WhoWeAreForm · «quiénes somos» (FR-003)', () => {
  it('el contador avisa y bloquea el guardado al superar 1.000 caracteres', async () => {
    server.use(http.get(adminUrl, () => HttpResponse.json(adminData())));

    render(
      <AppProviders>
        <WhoWeAreForm />
      </AppProviders>,
    );

    const textarea = await screen.findByLabelText(/^Texto \(Español\)/);
    fireEvent.change(textarea, { target: { value: 'a'.repeat(1001) } });

    expect(screen.getByText('1001 de 1.000 caracteres')).toBeInTheDocument();
    expect(screen.getByText(aboutLimitMessage(1001))).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeDisabled();
  });

  it('exige el texto en español y no guarda nada si falta', async () => {
    let putCount = 0;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(adminData())),
      http.put(aboutUrl, () => {
        putCount += 1;
        return HttpResponse.json(makeAbout());
      }),
    );

    render(
      <AppProviders>
        <WhoWeAreForm />
      </AppProviders>,
    );

    await screen.findByLabelText(/^Texto \(Español\)/);
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(await screen.findByText(REQUIRED_ABOUT)).toBeInTheDocument();
    expect(putCount).toBe(0);
  });

  it('crear deja la sección en borrador, avisa y guarda el inglés vacío como null (I6)', async () => {
    let current = adminData();
    let putBody: Record<string, unknown> | undefined;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.put(aboutUrl, async ({ request }) => {
        putBody = (await request.json()) as Record<string, unknown>;
        current = adminData({ about: makeAbout() });
        return HttpResponse.json(current.about);
      }),
    );

    render(
      <AppProviders>
        <WhoWeAreForm />
      </AppProviders>,
    );

    await userEvent.type(
      await screen.findByLabelText(/^Texto \(Español\)/),
      'Simiente Santa es una familia',
    );
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(await screen.findByText(SAVED)).toBeInTheDocument();
    expect(screen.getByText(DRAFT_NOTICE)).toBeInTheDocument();
    expect(putBody).toMatchObject({
      textEs: 'Simiente Santa es una familia',
      textEn: null,
      publicationState: 'draft',
    });
  });

  it('publica y retira la sección por separado (FR-013)', async () => {
    let current = adminData({ about: makeAbout() });
    const bodies: Array<Record<string, unknown>> = [];
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.put(aboutUrl, async ({ request }) => {
        const body = (await request.json()) as Record<string, unknown>;
        bodies.push(body);
        current = adminData({
          about: makeAbout({ publicationState: body.publicationState as 'draft' | 'published' }),
        });
        return HttpResponse.json(current.about);
      }),
    );

    render(
      <AppProviders>
        <WhoWeAreForm />
      </AppProviders>,
    );

    await userEvent.click(await screen.findByRole('button', { name: 'Publicar' }));
    expect(await screen.findByText(PUBLISHED)).toBeInTheDocument();
    expect(bodies[0]).toMatchObject({ publicationState: 'published' });

    await userEvent.click(await screen.findByRole('button', { name: 'Retirar de la portada' }));
    const dialog = await screen.findByRole('dialog', { name: 'Retirar de la portada' });
    await userEvent.click(within(dialog).getByRole('button', { name: 'Retirar' }));
    expect(await screen.findByText(UNPUBLISHED)).toBeInTheDocument();
    expect(bodies[1]).toMatchObject({ publicationState: 'draft' });
  });
});

describe('ContactForm · contacto (FR-007)', () => {
  it('precarga los datos de contacto', async () => {
    server.use(http.get(adminUrl, () => HttpResponse.json(adminData({ contact: makeContact() }))));

    render(
      <AppProviders>
        <ContactForm />
      </AppProviders>,
    );

    expect(await screen.findByDisplayValue('San José')).toBeInTheDocument();
    expect(screen.getByLabelText(/^Correo/)).toHaveValue('hola@ejemplo.com');
    expect(screen.getByLabelText(/^Teléfono/)).toHaveValue('+506 8888 8888');
  });

  it('muestra los errores de correo y teléfono junto al campo y no guarda (FR-015)', async () => {
    let putCount = 0;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(adminData())),
      http.put(contactUrl, () => {
        putCount += 1;
        return HttpResponse.json(makeContact());
      }),
    );

    render(
      <AppProviders>
        <ContactForm />
      </AppProviders>,
    );

    await userEvent.type(await screen.findByLabelText(/^Dirección \(Español\)/), 'San José');
    await userEvent.type(screen.getByLabelText(/^Correo/), 'no-es-correo');
    await userEvent.type(screen.getByLabelText(/^Teléfono/), '12');
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(await screen.findByText(EMAIL_INVALID)).toBeInTheDocument();
    expect(screen.getByText(PHONE_INVALID)).toBeInTheDocument();
    expect(putCount).toBe(0);
  });

  it('guarda dirección/correo/teléfono y refresca la sección', async () => {
    let current = adminData();
    let putBody: Record<string, unknown> | undefined;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.put(contactUrl, async ({ request }) => {
        putBody = (await request.json()) as Record<string, unknown>;
        current = adminData({ contact: makeContact() });
        return HttpResponse.json(current.contact);
      }),
    );

    render(
      <AppProviders>
        <ContactForm />
      </AppProviders>,
    );

    await userEvent.type(await screen.findByLabelText(/^Dirección \(Español\)/), 'San José');
    await userEvent.type(screen.getByLabelText(/^Correo/), 'hola@ejemplo.com');
    await userEvent.type(screen.getByLabelText(/^Teléfono/), '+506 8888 8888');
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(await screen.findByText(SAVED)).toBeInTheDocument();
    await waitFor(() => expect(putBody).toBeDefined());
    expect(putBody).toMatchObject({
      addressEs: 'San José',
      addressEn: null,
      email: 'hola@ejemplo.com',
      phone: '+506 8888 8888',
    });
  });

  it('muestra junto al campo el error del servidor sobre el teléfono', async () => {
    server.use(
      http.get(adminUrl, () => HttpResponse.json(adminData())),
      http.put(contactUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Revisa los datos',
              details: { phone: 'El teléfono no es válido' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    render(
      <AppProviders>
        <ContactForm />
      </AppProviders>,
    );

    await userEvent.type(await screen.findByLabelText(/^Dirección \(Español\)/), 'San José');
    await userEvent.type(screen.getByLabelText(/^Correo/), 'hola@ejemplo.com');
    await userEvent.type(screen.getByLabelText(/^Teléfono/), '+506 8888 8888');
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(await screen.findByText('El teléfono no es válido')).toBeInTheDocument();
  });
});
