import { QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import { API_BASE_URL, type ErrorEnvelope } from '../../api/client';
import { server } from '../../test/server';
import { createQueryClientForTest } from '../../test/queryClient';
import { StatusPage } from './pages/StatusPage';

const healthzUrl = `${API_BASE_URL}/healthz`;
const INTRO = /Este sitio web está en construcción/;

function renderPage() {
  return render(
    <QueryClientProvider client={createQueryClientForTest()}>
      <StatusPage />
    </QueryClientProvider>,
  );
}

describe('StatusPage', () => {
  it('muestra el estado Consultando antes de recibir la primera respuesta', () => {
    server.use(
      http.get(healthzUrl, async () => {
        await new Promise(() => {
          /* nunca responde */
        });
      }),
    );

    renderPage();

    expect(screen.getByText('Consultando el estado del sistema…')).toBeInTheDocument();
    expect(screen.getByText(INTRO)).toBeInTheDocument();
  });

  it('muestra el estado conectado cuando la respuesta es 200 + SystemStatus válido', async () => {
    server.use(
      http.get(healthzUrl, () => HttpResponse.json({ status: 'ok', database: 'connected' })),
    );

    renderPage();

    expect(await screen.findByTestId('veredicto')).toHaveTextContent(
      'El sistema está funcionando.',
    );

    const detalle = screen.getByTestId('detalle');
    expect(within(detalle).getByText('Servidor:')).toBeInTheDocument();
    expect(within(detalle).getByText('en marcha')).toBeInTheDocument();
    expect(within(detalle).getByText('Base de datos:')).toBeInTheDocument();
    expect(within(detalle).getByText('conectada')).toBeInTheDocument();
    expect(screen.getByText(/Última consulta:/)).toBeInTheDocument();
  });

  it('muestra el error A cuando la respuesta es 503 + database_unavailable', async () => {
    const envelope: ErrorEnvelope = {
      error: {
        code: 'database_unavailable',
        message: 'La base de datos no está conectada',
        details: { database: 'disconnected' },
      },
    };
    server.use(http.get(healthzUrl, () => HttpResponse.json(envelope, { status: 503 })));

    renderPage();

    expect(await screen.findByTestId('veredicto')).toHaveTextContent(
      'El sistema está en marcha, pero la base de datos no está conectada.',
    );

    const detalle = screen.getByTestId('detalle');
    expect(within(detalle).getByText('en marcha')).toBeInTheDocument();
    expect(within(detalle).getByText('no conectada')).toBeInTheDocument();
    expect(
      screen.getByText(/la base de datos está detenida o todavía está arrancando/),
    ).toBeInTheDocument();
    // El mensaje y los detalles del sobre no se muestran frente a la persona.
    expect(screen.queryByText('La base de datos no está conectada')).not.toBeInTheDocument();
    expect(screen.queryByText('disconnected')).not.toBeInTheDocument();
    // ux.md §2/§6: la hora de la última consulta se muestra siempre que hay un
    // resultado, también en error (FR-003: nunca un estado memorizado).
    expect(screen.getByText(/Última consulta:/)).toBeInTheDocument();
  });

  it('muestra el error B (sin respuesta) cuando falla la red', async () => {
    server.use(http.get(healthzUrl, () => HttpResponse.error()));

    renderPage();

    expect(await screen.findByTestId('veredicto')).toHaveTextContent(
      'No se pudo consultar el estado del sistema.',
    );

    const detalle = screen.getByTestId('detalle');
    expect(within(detalle).getByText('sin respuesta')).toBeInTheDocument();
    expect(within(detalle).getByText('no se pudo comprobar')).toBeInTheDocument();
    expect(screen.getByText(/No hubo respuesta del servidor/)).toBeInTheDocument();
    // ux.md §6 (error B, sin respuesta): "Última consulta: [hora exacta del intento]".
    expect(screen.getByText(/Última consulta:/)).toBeInTheDocument();
  });

  it('muestra el error B (respuesta inesperada) ante un 500 con sobre válido de código no esperado', async () => {
    server.use(
      http.get(healthzUrl, () =>
        HttpResponse.json(
          { error: { code: 'internal', message: 'Error interno del servidor' } },
          { status: 500 },
        ),
      ),
    );

    renderPage();

    expect(await screen.findByTestId('veredicto')).toHaveTextContent(
      'No se pudo consultar el estado del sistema.',
    );

    const detalle = screen.getByTestId('detalle');
    // Servidor y Base de datos: "no se pudo comprobar" (nunca "sin respuesta" ni "no conectada").
    expect(within(detalle).getAllByText('no se pudo comprobar')).toHaveLength(2);
    expect(within(detalle).queryByText('sin respuesta')).not.toBeInTheDocument();
    expect(within(detalle).queryByText('no conectada')).not.toBeInTheDocument();
    expect(
      screen.getByText(/El servidor respondió, pero con algo que no pudimos interpretar/),
    ).toBeInTheDocument();
    expect(screen.queryByTestId('veredicto')).not.toHaveTextContent('El sistema está funcionando');
    // ux.md §6 (error B, respuesta inesperada): "Última consulta: [hora exacta del intento]".
    expect(screen.getByText(/Última consulta:/)).toBeInTheDocument();
  });

  it('muestra el error B (respuesta inesperada) ante un 200 con cuerpo no conforme', async () => {
    server.use(http.get(healthzUrl, () => HttpResponse.json({ status: 'ok' })));

    renderPage();

    expect(await screen.findByTestId('veredicto')).toHaveTextContent(
      'No se pudo consultar el estado del sistema.',
    );

    const detalle = screen.getByTestId('detalle');
    expect(within(detalle).getAllByText('no se pudo comprobar')).toHaveLength(2);
    expect(screen.queryByText('El sistema está funcionando.')).not.toBeInTheDocument();
    expect(screen.getByText(/Última consulta:/)).toBeInTheDocument();
  });

  it('vuelve a consultar al pulsar el botón y refleja el estado actual', async () => {
    let attempt = 0;
    server.use(
      http.get(healthzUrl, () => {
        attempt += 1;
        if (attempt === 1) {
          return HttpResponse.error();
        }
        return HttpResponse.json({ status: 'ok', database: 'connected' });
      }),
    );

    renderPage();

    expect(await screen.findByTestId('veredicto')).toHaveTextContent(
      'No se pudo consultar el estado del sistema.',
    );

    const button = screen.getByRole('button', { name: 'Volver a consultar el estado' });
    await userEvent.click(button);

    await waitFor(() =>
      expect(screen.getByTestId('veredicto')).toHaveTextContent('El sistema está funcionando.'),
    );
  });

  it('tras un éxito y una reconsulta fallida muestra la hora del intento, no la del éxito', async () => {
    // Controla el reloj del intento: la primera respuesta llega a las 10:00:00 y
    // la reconsulta fallida a las 11:00:00. Ambas horas quedan fijadas en el
    // `Date` que se usa para sellar el resultado de cada consulta.
    const intentos = [
      {
        hora: new Date('2026-10-03T10:00:00'),
        respuesta: () => HttpResponse.json({ status: 'ok', database: 'connected' }),
      },
      { hora: new Date('2026-10-03T11:00:00'), respuesta: () => HttpResponse.error() },
    ];
    let attempt = 0;
    const ahora = vi.spyOn(Date, 'now');
    server.use(
      http.get(healthzUrl, () => {
        const intento = intentos[attempt];
        attempt += 1;
        ahora.mockReturnValue(intento.hora.getTime());
        return intento.respuesta();
      }),
    );

    renderPage();

    expect(await screen.findByTestId('veredicto')).toHaveTextContent(
      'El sistema está funcionando.',
    );
    expect(screen.getByText(/Última consulta:/)).toHaveTextContent('10:00:00');

    const button = screen.getByRole('button', { name: 'Volver a consultar el estado' });
    await userEvent.click(button);

    await waitFor(() =>
      expect(screen.getByTestId('veredicto')).toHaveTextContent(
        'No se pudo consultar el estado del sistema.',
      ),
    );

    // La hora mostrada es la del intento fallido (11:00:00), nunca la del éxito
    // anterior (10:00:00): el estado no es una copia memorizada (FR-003).
    expect(screen.getByText(/Última consulta:/)).toHaveTextContent('11:00:00');
    expect(screen.getByText(/Última consulta:/)).not.toHaveTextContent('10:00:00');
  });

  it('es accesible: región en vivo, título y único botón', async () => {
    server.use(
      http.get(healthzUrl, () => HttpResponse.json({ status: 'ok', database: 'connected' })),
    );

    renderPage();

    await screen.findByTestId('veredicto');

    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(
      screen.getByRole('heading', { level: 1, name: 'Estado del sistema' }),
    ).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2, name: 'Estado actual' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Volver a consultar el estado' })).toBeEnabled();
  });
});
