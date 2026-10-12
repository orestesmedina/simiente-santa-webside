import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { ProximamentePublica } from '../../../api/agenda';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { PublicLayout } from '../components/PublicLayout';
import { EventsPage } from './EventsPage';

const upcomingUrl = `${API_BASE_URL}/api/v1/proximamente`;
const portadaUrl = `${API_BASE_URL}/api/v1/portada`;

const evento = {
  id: 'ev-1',
  title: 'Vigilia de viernes',
  frequency: 'Todos los viernes',
  dates: [
    { date: '2026-10-17', startTime: '19:00', endTime: '21:00' },
    { date: '2026-10-24', startTime: '19:00' },
  ],
  place: 'Templo central',
  mapUrl: 'https://maps.google.com/?q=templo',
  detail: 'Noche de oración\nAbierta a todo el barrio',
  category: { id: 'cat-1', name: 'noche de oración' },
  image: { url: '/api/v1/media/img_vigilia.png', alt: 'Templo iluminado de noche' },
};

const actividad = {
  id: 'ac-1',
  title: 'Repartir comida',
  date: '2026-10-18',
  startTime: '08:00',
  endTime: '11:00',
  place: 'Comedor comunal',
  detail: 'Acción social del mes',
};

/** Sondea la ruta activa: prueba que la página **no** redirige (⟲ D-13). */
function LocationProbe() {
  const location = useLocation();
  return <p data-testid="pathname">{location.pathname}</p>;
}

function renderEventos() {
  return render(
    <AppProviders>
      <MemoryRouter initialEntries={['/eventos']}>
        <LocationProbe />
        <Routes>
          <Route element={<PublicLayout />}>
            <Route index element={<p>Portada</p>} />
            <Route path="eventos" element={<EventsPage />} />
          </Route>
        </Routes>
      </MemoryRouter>
    </AppProviders>,
  );
}

function responderProximamente(data: ProximamentePublica) {
  server.use(http.get(upcomingUrl, () => HttpResponse.json(data)));
}

function responderPortada() {
  server.use(http.get(portadaUrl, () => HttpResponse.json({ lang: 'es' })));
}

describe('EventsPage · composición de la sección pública (US1, SC-012)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('compone los dos bloques con sus títulos, en el orden del servidor', async () => {
    responderPortada();
    responderProximamente({ lang: 'es', events: [evento], activities: [actividad] });

    renderEventos();

    const eventos = await screen.findByRole('heading', { name: 'Próximos eventos' });
    const actividades = screen.getByRole('heading', { name: 'Próximas actividades' });
    expect(eventos).toBeInTheDocument();
    expect(actividades).toBeInTheDocument();

    // El bloque de eventos va antes que el de actividades en el DOM (FR-021).
    expect(eventos.compareDocumentPosition(actividades)).toBe(Node.DOCUMENT_POSITION_FOLLOWING);

    // Un `h1` por página y los títulos de bloque como `h2` (ux.md §8.1).
    expect(
      screen.getByRole('heading', { level: 1, name: 'Eventos y actividades' }),
    ).toBeInTheDocument();
    expect(eventos.tagName).toBe('H2');
    expect(
      screen.getByRole('heading', { level: 3, name: 'Vigilia de viernes' }),
    ).toBeInTheDocument();
  });

  it('con un solo concepto el otro bloque no se dibuja (⟲ D-5: cero rótulos huérfanos)', async () => {
    responderPortada();
    responderProximamente({ lang: 'es', events: [evento] });

    renderEventos();

    expect(await screen.findByRole('heading', { name: 'Próximos eventos' })).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Próximas actividades' })).toBeNull();
    expect(screen.queryByText('Repartir comida')).toBeNull();
  });

  it('muestra la próxima fecha escrita, su horario localizado y el `<time datetime>` real', async () => {
    responderPortada();
    responderProximamente({ lang: 'es', events: [evento] });

    renderEventos();

    const fecha = await screen.findByText('sábado 17 de octubre');
    expect(fecha.tagName).toBe('TIME');
    // El dato viaja intacto: `YYYY-MM-DD` + hora, sin conversión de zona (⟲ D-22).
    expect(fecha).toHaveAttribute('datetime', '2026-10-17T19:00');

    expect(screen.getByTestId('hora')).toHaveTextContent('7:00 p. m. − 9:00 p. m.');

    // La frecuencia y las fechas restantes (⟲ D-11) también están a la vista.
    expect(screen.getByText('Todos los viernes')).toBeInTheDocument();
    expect(screen.getByText('Más fechas')).toBeInTheDocument();
    const siguiente = document.querySelector('time[datetime="2026-10-24T19:00"]');
    expect(siguiente).not.toBeNull();
    expect(siguiente).toHaveTextContent('sábado 24 de octubre');
  });

  it('la imagen lleva el alt del equipo y el enlace de mapa el nombre accesible completo (D-20)', async () => {
    responderPortada();
    responderProximamente({ lang: 'es', events: [evento], activities: [actividad] });

    renderEventos();

    const imagen = await screen.findByRole('img', { name: 'Templo iluminado de noche' });
    expect(imagen).toHaveAttribute('src', `${API_BASE_URL}/api/v1/media/img_vigilia.png`);

    const mapa = screen.getByRole('link', { name: 'Ver ubicación de "Vigilia de viernes"' });
    expect(mapa).toHaveAttribute('href', 'https://maps.google.com/?q=templo');
    expect(mapa).toHaveAttribute('target', '_blank');
    expect(mapa).toHaveAttribute('rel', 'noopener noreferrer');

    // La categoría se anuncia con contexto (⟲ D-2) y no es un color semántico.
    expect(screen.getByLabelText('Categoría: noche de oración')).toBeInTheDocument();
  });

  it('un elemento sin imagen y sin categoría se ve completo (FR-020)', async () => {
    responderPortada();
    responderProximamente({ lang: 'es', activities: [actividad] });

    renderEventos();

    const titulo = await screen.findByRole('heading', { level: 3, name: 'Repartir comida' });
    const tarjeta = titulo.closest('article');
    expect(tarjeta).not.toBeNull();
    expect(within(tarjeta as HTMLElement).getByText('domingo 18 de octubre')).toBeInTheDocument();
    expect(within(tarjeta as HTMLElement).getByText('Acción social del mes')).toBeInTheDocument();
    expect(within(tarjeta as HTMLElement).getByText('Comedor comunal')).toBeInTheDocument();
    expect(within(tarjeta as HTMLElement).queryByRole('img')).toBeNull();
    expect(within(tarjeta as HTMLElement).queryByRole('link')).toBeNull();
  });

  it('una actividad sin horario no dibuja bloque de horario (⟲ D-12)', async () => {
    responderPortada();
    responderProximamente({
      lang: 'es',
      activities: [
        { id: 'ac-2', title: 'Ensayo del coro', date: '2026-10-20', place: 'Salón', detail: 'x' },
      ],
    });

    renderEventos();

    await screen.findByRole('heading', { level: 3, name: 'Ensayo del coro' });
    expect(screen.queryByTestId('hora')).toBeNull();
    expect(screen.getByText('martes 20 de octubre')).toHaveAttribute('datetime', '2026-10-20');
  });

  it('en inglés, la interfaz cambia y el contenido sin inglés se ve en español (US5)', async () => {
    window.localStorage.setItem('ss.lang', 'en');
    responderPortada();
    responderProximamente({ lang: 'en', events: [evento], activities: [actividad] });

    renderEventos();

    expect(await screen.findByRole('heading', { name: 'Upcoming events' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Upcoming activities' })).toBeInTheDocument();
    // El servidor resolvió el fallback `en → es`: el cliente lo pinta tal cual.
    expect(
      screen.getByRole('heading', { level: 3, name: 'Vigilia de viernes' }),
    ).toBeInTheDocument();
    // Y las fechas se reescriben en formato inglés sin cambiar el día real.
    const fecha = screen.getByText('Saturday, October 17');
    expect(fecha).toHaveAttribute('datetime', '2026-10-17T19:00');
    expect(screen.getAllByTestId('hora')[0]).toHaveTextContent('7:00 PM − 9:00 PM');
  });
});

describe('EventsPage · sin nada publicado y próximo (FR-013/FR-022, ⟲ D-13)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('pinta el aviso de sección vacía sin redirigir a la portada', async () => {
    responderPortada();
    responderProximamente({ lang: 'es' });

    renderEventos();

    expect(await screen.findByText('Por ahora no hay eventos ni actividades')).toBeInTheDocument();
    expect(screen.getByTestId('hidden-notice')).toBeInTheDocument();

    // La URL se conserva: no hay `Navigate` ni `useNavigate` (decisión firme).
    expect(screen.getByTestId('pathname')).toHaveTextContent('/eventos');
    expect(screen.queryByText('Portada')).toBeNull();

    // Sin bloques: ni rótulos ni datos (cero secciones vacías).
    expect(screen.queryByRole('heading', { level: 2 })).toBeNull();
    expect(screen.queryByRole('heading', { level: 3 })).toBeNull();
    // La página sigue siendo la sección: H1 presente.
    expect(
      screen.getByRole('heading', { level: 1, name: 'Eventos y actividades' }),
    ).toBeInTheDocument();
  });
});

describe('EventsPage · estados de carga y error (ux.md §4.1)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('mientras carga muestra esqueletos con anuncio', async () => {
    responderPortada();
    server.use(
      http.get(upcomingUrl, async () => {
        await delay(50);
        return HttpResponse.json({ lang: 'es' });
      }),
    );

    renderEventos();

    expect(screen.getAllByText('Cargando…').length).toBeGreaterThan(0);
    expect(await screen.findByText('Por ahora no hay eventos ni actividades')).toBeInTheDocument();
  });

  it('sin nada en caché ofrece «Volver a intentar»', async () => {
    responderPortada();
    let fallos = 1;
    server.use(
      http.get(upcomingUrl, () => {
        if (fallos > 0) {
          fallos -= 1;
          return HttpResponse.json(
            { error: { code: 'internal', message: 'Error' } },
            { status: 500 },
          );
        }
        return HttpResponse.json({ lang: 'es', events: [evento] });
      }),
    );

    renderEventos();

    expect(
      await screen.findByText(/No pudimos cargar los eventos y actividades/),
    ).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Volver a intentar' }));

    expect(await screen.findByRole('heading', { name: 'Próximos eventos' })).toBeInTheDocument();
  });

  it('con caché previa mantiene el contenido y avisa de la actualización fallida', async () => {
    responderPortada();
    let respuestas = 0;
    server.use(
      http.get(upcomingUrl, () => {
        respuestas += 1;
        if (respuestas === 1) {
          return HttpResponse.json({ lang: 'es', events: [evento] });
        }
        return HttpResponse.json(
          { error: { code: 'internal', message: 'Error' } },
          { status: 500 },
        );
      }),
    );

    renderEventos();
    await screen.findByRole('heading', { name: 'Próximos eventos' });

    // Un cambio de idioma fuerza una segunda petición que falla: se conserva lo
    // último válido y se avisa sin vaciar la página.
    await userEvent.click(screen.getAllByRole('radio', { name: 'Inglés' })[0]);

    await waitFor(() => expect(screen.getByText(/couldn't refresh/)).toBeInTheDocument());
    expect(
      screen.getByRole('heading', { level: 3, name: 'Vigilia de viernes' }),
    ).toBeInTheDocument();
  });
});
