import { render, screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { OfrendasPublicas } from '../../../api/ofrendas';
import { AppProviders } from '../../../app/providers';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { server } from '../../../test/server';
import { OfrendasPage } from './OfrendasPage';

const ofrendasUrl = `${API_BASE_URL}/api/v1/ofrendas`;

const completa: OfrendasPublicas = {
  lang: 'es',
  intro: { text: 'Tu aporte va directo a la obra.' },
  items: [
    {
      id: '11111111-1111-1111-1111-111111111111',
      kind: 'iban',
      number: 'CR05015202001234567890',
      institution: 'Banco Nacional de Costa Rica',
      currency: 'CRC',
      holder: 'Simiente Santa',
    },
    {
      id: '22222222-2222-2222-2222-222222222222',
      kind: 'iban',
      number: 'CR01010222003344556677',
      institution: 'BAC San José',
      currency: 'USD',
    },
    {
      id: '33333333-3333-3333-3333-333333333333',
      kind: 'sinpe',
      number: '+50688888888',
      institution: 'Banco Promerica',
    },
  ],
  usage: { text: 'Sostenemos la obra social y las misiones.' },
};

function responder(data: OfrendasPublicas) {
  server.use(http.get(ofrendasUrl, () => HttpResponse.json(data)));
}

function renderPage() {
  return render(
    <AppProviders>
      <LanguageProvider>
        <MemoryRouter initialEntries={['/ofrendas']}>
          <Routes>
            <Route path="/ofrendas" element={<OfrendasPage />} />
            <Route path="/" element={<p>portada</p>} />
          </Routes>
        </MemoryRouter>
      </LanguageProvider>
    </AppProviders>,
  );
}

describe('OfrendasPage · sección pública (US1/US4/US5)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('compone H1, introductorio, dos grupos (IBAN primero) y texto de uso', async () => {
    responder(completa);

    renderPage();

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Ofrendas y diezmos' }),
    ).toBeInTheDocument();
    expect(screen.getByText('Tu aporte va directo a la obra.')).toBeInTheDocument();

    const iban = screen.getByRole('heading', { level: 2, name: 'Cuentas IBAN' });
    const sinpe = screen.getByRole('heading', { level: 2, name: 'Números de SINPE Móvil' });
    // IBAN primero y luego SINPE (FR-004/Q1), como lo entrega el servidor.
    expect(iban.compareDocumentPosition(sinpe)).toBe(Node.DOCUMENT_POSITION_FOLLOWING);

    expect(screen.getByText('CR05 0152 0200 1234 5678 90')).toBeInTheDocument();
    expect(screen.getByText('+50688888888')).toBeInTheDocument();

    expect(
      screen.getByRole('heading', { level: 3, name: '¿En qué usamos las donaciones?' }),
    ).toBeInTheDocument();
    expect(screen.getByText('Sostenemos la obra social y las misiones.')).toBeInTheDocument();
  });

  it('sin cuentas IBAN publicadas el rótulo del grupo no existe (cero rótulo huérfano)', async () => {
    responder({
      lang: 'es',
      items: [completa.items![2]],
    });

    renderPage();

    await screen.findByRole('heading', { level: 1, name: 'Ofrendas y diezmos' });
    expect(screen.queryByRole('heading', { level: 2, name: 'Cuentas IBAN' })).toBeNull();
    expect(
      screen.getByRole('heading', { level: 2, name: 'Números de SINPE Móvil' }),
    ).toBeInTheDocument();
  });

  it('con SOLO textos publicados (0 métodos) la sección vive y NO redirige (P7-13)', async () => {
    responder({ lang: 'es', intro: { text: 'Cómo donar.' }, usage: { text: 'En qué se usa.' } });

    renderPage();

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Ofrendas y diezmos' }),
    ).toBeInTheDocument();
    expect(screen.getByText('Cómo donar.')).toBeInTheDocument();
    expect(screen.queryByRole('heading', { level: 2, name: 'Cuentas IBAN' })).toBeNull();
    // No hubo redirección a la portada.
    expect(screen.queryByText('portada')).toBeNull();
  });

  it('con nada publicado redirige en silencio a la portada (FR-005/Q9/P7-21)', async () => {
    responder({ lang: 'es' });

    renderPage();

    expect(await screen.findByText('portada')).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Ofrendas y diezmos' })).toBeNull();
  });

  it('ante error sin caché NO redirige: H1 + mensaje + «Volver a intentar»', async () => {
    server.use(
      http.get(ofrendasUrl, () =>
        HttpResponse.json(
          { error: { code: 'internal', message: 'Error interno' } },
          { status: 500 },
        ),
      ),
    );

    renderPage();

    expect(
      await screen.findByText(
        'No pudimos cargar la información de ofrendas. Revisa tu conexión y vuelve a intentarlo.',
      ),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Volver a intentar' })).toBeInTheDocument();
    expect(
      screen.getByRole('heading', { level: 1, name: 'Ofrendas y diezmos' }),
    ).toBeInTheDocument();
    // Sigue en /ofrendas (no se navegó a la portada).
    expect(screen.queryByText('portada')).toBeNull();
  });

  it('mientras decide muestra solo el esqueleto de carga (sin destello)', async () => {
    responder(completa);

    const { container } = renderPage();

    expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
    expect(within(container).queryByRole('heading', { name: 'Ofrendas y diezmos' })).toBeNull();
  });
});
