import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { PortadaAdmin } from '../../api/portada';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { NO_PERMISSION_ERROR, INFO_LOAD_ERROR } from './messages';
import { InformationPage } from './pages/InformationPage';

const adminUrl = `${API_BASE_URL}/api/v1/admin/portada`;

const emptyData: PortadaAdmin = {
  identity: null,
  about: null,
  contact: null,
  schedule: { items: [] },
  whatsapp: { items: [] },
  socials: { items: [] },
};

const fullData: PortadaAdmin = {
  identity: {
    nameEs: 'Iglesia Simiente Santa',
    publicationState: 'published',
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
  },
  about: {
    textEs: 'Somos una familia',
    publicationState: 'draft',
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
  },
  contact: {
    addressEs: 'San José',
    email: 'hola@ejemplo.com',
    phone: '+506 8888 8888',
    publicationState: 'published',
    createdAt: '2026-10-01T10:00:00Z',
    updatedAt: '2026-10-01T10:00:00Z',
  },
  schedule: {
    items: [
      {
        id: 'srv-1',
        dayOfWeek: 0,
        startTime: '10:00',
        endTime: '12:00',
        nameEs: 'Culto dominical',
        placeEs: 'Templo central',
        publicationState: 'published',
        sortOrder: 0,
        createdAt: '2026-10-01T10:00:00Z',
        updatedAt: '2026-10-01T10:00:00Z',
      },
    ],
  },
  whatsapp: {
    items: [
      {
        id: 'wa-1',
        nameEs: 'Escríbenos',
        kind: 'direct',
        destination: '+506 8888 8888',
        publicationState: 'draft',
        sortOrder: 0,
        createdAt: '2026-10-01T10:00:00Z',
        updatedAt: '2026-10-01T10:00:00Z',
      },
    ],
  },
  socials: {
    items: [
      {
        id: 'so-1',
        network: 'facebook',
        url: 'https://facebook.com/simiente',
        publicationState: 'published',
        createdAt: '2026-10-01T10:00:00Z',
        updatedAt: '2026-10-01T10:00:00Z',
      },
    ],
  },
};

function responder(data: PortadaAdmin) {
  server.use(http.get(adminUrl, () => HttpResponse.json(data)));
}

function renderInformation() {
  return render(
    <AppProviders>
      <InformationPage />
    </AppProviders>,
  );
}

describe('InformationPage · pestañas (US3 esc. 7)', () => {
  it('muestra el título, la banda de ayuda y las seis pestañas', async () => {
    responder(emptyData);

    renderInformation();

    expect(
      await screen.findByRole('heading', { name: 'Portada e información general' }),
    ).toBeInTheDocument();
    expect(screen.getByText(/Solo lo publicado en cada sección es visible/)).toBeInTheDocument();

    const nav = screen.getByRole('navigation', { name: 'Secciones de la portada' });
    for (const label of [
      'Identidad',
      'Quiénes somos',
      'Horario de servicios',
      'WhatsApp',
      'Redes sociales',
      'Contacto',
    ]) {
      expect(within(nav).getByRole('button', { name: label })).toBeInTheDocument();
    }
  });

  it('navega por las seis piezas y muestra el contenido de cada una', async () => {
    responder(fullData);

    renderInformation();
    const nombre = await screen.findByLabelText(/^Nombre oficial/);
    expect(nombre).toHaveValue('Iglesia Simiente Santa');
    expect(screen.getByText('Publicado')).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Quiénes somos' }));
    expect(screen.getByText('Somos una familia')).toBeInTheDocument();
    expect(screen.getByText('Borrador')).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Horario de servicios' }));
    expect(screen.getByText(/Culto dominical/)).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'WhatsApp' }));
    expect(screen.getByText(/Escríbenos/)).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Redes sociales' }));
    expect(screen.getByText('Facebook')).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Contacto' }));
    expect(screen.getByText('San José')).toBeInTheDocument();
  });
});

describe('InformationPage · estados', () => {
  it('muestra el estado vacío de cada pieza', async () => {
    responder(emptyData);

    renderInformation();

    expect(await screen.findByLabelText(/^Nombre oficial/)).toHaveValue('');

    await userEvent.click(screen.getByRole('button', { name: 'Horario de servicios' }));
    expect(
      screen.getByText(
        'Todavía no hay servicios. Agrega el primero para que aparezca en la portada.',
      ),
    ).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'WhatsApp' }));
    expect(
      screen.getByText(
        'Todavía no hay canales. Agrega el primero para que aparezca en la portada.',
      ),
    ).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Redes sociales' }));
    expect(screen.getAllByText('Sin enlace todavía').length).toBe(5);
  });

  it('muestra el estado de carga', () => {
    responder(emptyData);

    renderInformation();

    expect(screen.getByText('Cargando información…')).toBeInTheDocument();
  });

  it('con error ofrece reintentar', async () => {
    let intentos = 0;
    server.use(
      http.get(adminUrl, () => {
        intentos += 1;
        if (intentos === 1) {
          return HttpResponse.json(
            { error: { code: 'internal', message: 'Error' } },
            { status: 500 },
          );
        }
        return HttpResponse.json(fullData);
      }),
    );

    renderInformation();

    expect(await screen.findByText(INFO_LOAD_ERROR)).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Reintentar' }));

    expect(await screen.findByDisplayValue('Iglesia Simiente Santa')).toBeInTheDocument();
  });

  it('un 403 del servidor se explica como sin permiso (FR-012)', async () => {
    server.use(
      http.get(adminUrl, () =>
        HttpResponse.json(
          { error: { code: 'forbidden', message: 'Sin permiso' } },
          { status: 403 },
        ),
      ),
    );

    renderInformation();

    expect(await screen.findByText(NO_PERMISSION_ERROR)).toBeInTheDocument();
  });
});
