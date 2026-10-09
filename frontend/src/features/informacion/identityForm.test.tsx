import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { IdentityAdmin, PortadaAdmin } from '../../api/portada';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { IdentityForm } from './components/IdentityForm';
import {
  DRAFT_NOTICE,
  PUBLISHED,
  REQUIRED_ALT,
  SAVED,
  SAVED_PUBLISHED,
  UNPUBLISHED,
  imageInvalidMessage,
} from './messages';

const adminUrl = `${API_BASE_URL}/api/v1/admin/portada`;
const identityUrl = `${adminUrl}/identidad`;
const imagesUrl = `${adminUrl}/imagenes`;

function makeIdentity(overrides: Partial<IdentityAdmin> = {}): IdentityAdmin {
  return {
    nameEs: 'Iglesia Simiente Santa',
    publicationState: 'draft',
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function admin(identity: IdentityAdmin | null): PortadaAdmin {
  return {
    identity,
    about: null,
    contact: null,
    schedule: { items: [] },
    whatsapp: { items: [] },
    socials: { items: [] },
  };
}

function renderForm() {
  return render(
    <AppProviders>
      <IdentityForm />
    </AppProviders>,
  );
}

describe('IdentityForm · edición', () => {
  it('precarga la identidad y ofrece retirar cuando está publicada', async () => {
    server.use(
      http.get(adminUrl, () =>
        HttpResponse.json(admin(makeIdentity({ publicationState: 'published' }))),
      ),
    );

    renderForm();

    expect(await screen.findByDisplayValue('Iglesia Simiente Santa')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Retirar de la portada' })).toBeInTheDocument();
  });

  it('al editar un elemento publicado avisa que ya es visible (FR-014)', async () => {
    let current = admin(makeIdentity({ publicationState: 'published' }));
    let putBody: unknown;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.put(identityUrl, async ({ request }) => {
        putBody = await request.json();
        current = admin(
          makeIdentity({ publicationState: 'published', nameEs: 'Iglesia Simiente' }),
        );
        return HttpResponse.json(current.identity);
      }),
    );

    renderForm();
    const nombre = await screen.findByLabelText(/^Nombre oficial/);
    await waitFor(() => expect(nombre).toHaveValue('Iglesia Simiente Santa'));
    await userEvent.clear(nombre);
    await userEvent.type(nombre, 'Iglesia Simiente');
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(await screen.findByText(SAVED_PUBLISHED)).toBeInTheDocument();
    expect(putBody).toMatchObject({ nameEs: 'Iglesia Simiente', publicationState: 'published' });
  });

  it('crear la identidad la deja en borrador y lo avisa', async () => {
    let current = admin(null);
    let putBody: Record<string, unknown> | undefined;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.put(identityUrl, async ({ request }) => {
        putBody = (await request.json()) as Record<string, unknown>;
        current = admin(makeIdentity());
        return HttpResponse.json(current.identity);
      }),
    );

    renderForm();
    await userEvent.type(await screen.findByLabelText(/^Nombre oficial/), 'Iglesia Simiente Santa');
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(await screen.findByText(SAVED)).toBeInTheDocument();
    expect(screen.getByText(DRAFT_NOTICE)).toBeInTheDocument();
    expect(putBody).toMatchObject({ publicationState: 'draft' });
  });

  it('un campo English vacío se envía como null (analyze I6)', async () => {
    const current = admin(makeIdentity());
    let putBody: Record<string, unknown> | undefined;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.put(identityUrl, async ({ request }) => {
        putBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(current.identity);
      }),
    );

    renderForm();
    await screen.findByLabelText(/^Nombre oficial/);
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    await waitFor(() => expect(putBody).toBeDefined());
    expect(putBody?.nameEn).toBeNull();
    expect(putBody?.missionEn).toBeNull();
  });
});

describe('IdentityForm · imágenes (FR-019)', () => {
  it('subir una imagen exige el texto alternativo en español antes de guardar', async () => {
    const current = admin(null);
    let putCount = 0;
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.post(imagesUrl, () =>
        HttpResponse.json(
          {
            fileName: 'img_logo.png',
            url: '/api/v1/media/img_logo.png',
            mimeType: 'image/png',
            sizeBytes: 100,
          },
          { status: 201 },
        ),
      ),
      http.put(identityUrl, () => {
        putCount += 1;
        return HttpResponse.json(current.identity);
      }),
    );

    renderForm();
    await userEvent.type(await screen.findByLabelText(/^Nombre oficial/), 'Iglesia Simiente Santa');

    const logoGroup = screen.getByRole('group', { name: 'Logotipo' });
    const file = new File(['contenido'], 'logo.png', { type: 'image/png' });
    await userEvent.upload(within(logoGroup).getByLabelText(/^Archivo/), file);

    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(await within(logoGroup).findByText(REQUIRED_ALT)).toBeInTheDocument();
    expect(putCount).toBe(0);
  });

  it('un archivo inválido muestra el aviso sin llamar a la subida', async () => {
    server.use(
      http.get(adminUrl, () => HttpResponse.json(admin(null))),
      http.post(imagesUrl, () => {
        throw new Error('no debería subirse');
      }),
    );

    renderForm();
    await screen.findByLabelText(/^Nombre oficial/);

    const logoGroup = screen.getByRole('group', { name: 'Logotipo' });
    const file = new File(['<svg/>'], 'logo.svg', { type: 'image/svg+xml' });
    fireEvent.change(within(logoGroup).getByLabelText(/^Archivo/), { target: { files: [file] } });

    expect(await within(logoGroup).findByText(imageInvalidMessage('8 MB'))).toBeInTheDocument();
  });

  it('muestra junto al campo el error del servidor sobre el alt (FR-015)', async () => {
    const current = admin(null);
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.post(imagesUrl, () =>
        HttpResponse.json(
          {
            fileName: 'img_logo.png',
            url: '/api/v1/media/img_logo.png',
            mimeType: 'image/png',
            sizeBytes: 100,
          },
          { status: 201 },
        ),
      ),
      http.put(identityUrl, () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid',
              message: 'Revisa los datos',
              details: { logoAltEs: 'Escribe el texto alternativo de la imagen.' },
            },
          },
          { status: 400 },
        ),
      ),
    );

    renderForm();
    await userEvent.type(await screen.findByLabelText(/^Nombre oficial/), 'Iglesia Simiente Santa');
    const logoGroup = screen.getByRole('group', { name: 'Logotipo' });
    await userEvent.upload(
      within(logoGroup).getByLabelText(/^Archivo/),
      new File(['contenido'], 'logo.png', { type: 'image/png' }),
    );
    await userEvent.type(
      within(logoGroup).getByLabelText(/^Texto alternativo \(Español\)/),
      'Logotipo de la iglesia',
    );

    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }));

    expect(
      await within(logoGroup).findByText('Escribe el texto alternativo de la imagen.'),
    ).toBeInTheDocument();
  });
});

describe('IdentityForm · publicar y retirar (FR-013)', () => {
  it('publica, avisa y ofrece retirar; retirar confirma y avisa', async () => {
    let current = admin(makeIdentity({ publicationState: 'draft' }));
    const bodies: Array<Record<string, unknown>> = [];
    server.use(
      http.get(adminUrl, () => HttpResponse.json(current)),
      http.put(identityUrl, async ({ request }) => {
        const body = (await request.json()) as Record<string, unknown>;
        bodies.push(body);
        const state = body.publicationState as 'draft' | 'published';
        current = admin(makeIdentity({ publicationState: state }));
        return HttpResponse.json(current.identity);
      }),
    );

    renderForm();
    await userEvent.click(await screen.findByRole('button', { name: 'Publicar' }));

    expect(await screen.findByText(PUBLISHED)).toBeInTheDocument();
    expect(bodies[0]).toMatchObject({ publicationState: 'published' });

    await userEvent.click(await screen.findByRole('button', { name: 'Retirar de la portada' }));
    const dialog = await screen.findByRole('dialog', { name: 'Retirar de la portada' });
    expect(dialog).toHaveTextContent('dejará de ser visible al público');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Retirar' }));

    expect(await screen.findByText(UNPUBLISHED)).toBeInTheDocument();
    expect(bodies[1]).toMatchObject({ publicationState: 'draft' });
  });
});
