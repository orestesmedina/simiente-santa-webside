import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { ItemProximoPortada, PortadaProximamente } from '../../../api/agenda';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { useLanguage } from '../i18n/useLanguage';
import { ProximamenteBlock } from './ProximamenteBlock';

/**
 * Bloque «Lo que viene» de la portada (US7/FR-025…FR-027, ux.md §4.10): hasta 5
 * tarjetas compactas con imagen y etiqueta de concepto, botón «Ver más eventos»
 * (solo con ítems) y el estado vacío con su mensaje.
 */

const url = `${API_BASE_URL}/api/v1/portada/proximamente`;

function makeItem(overrides: Partial<ItemProximoPortada> = {}): ItemProximoPortada {
  return {
    kind: 'event',
    id: 'ev-1',
    title: 'Vigilia de viernes',
    date: '2026-10-17',
    startTime: '19:00',
    endTime: '21:00',
    place: 'Templo central',
    ...overrides,
  };
}

/** Cinco ítems ya ordenados por cercanía, eventos y actividades mezclados. */
const cinco: ItemProximoPortada[] = [
  makeItem({ id: 'ev-1', title: 'Vigilia de viernes', date: '2026-10-17', startTime: '19:00' }),
  makeItem({
    kind: 'activity',
    id: 'ac-1',
    title: 'Repartir comida',
    date: '2026-10-18',
    startTime: '08:00',
    place: 'Comedor comunal',
    image: { url: '/api/v1/media/img_comida.png', alt: 'Voluntarios repartiendo' },
  }),
  makeItem({ id: 'ev-2', title: 'Ensayo del coro', date: '2026-10-20' }),
  makeItem({ kind: 'activity', id: 'ac-2', title: 'Visita a enfermos', date: '2026-10-22' }),
  makeItem({ id: 'ev-3', title: 'Noche de aposentos', date: '2026-10-24' }),
];

function responder(data: PortadaProximamente) {
  server.use(http.get(url, () => HttpResponse.json(data)));
}

/** Permite cambiar de idioma desde la prueba (como el selector de F10). */
function LanguageProbe() {
  const { lang, setLang } = useLanguage();
  return (
    <button type="button" onClick={() => setLang(lang === 'es' ? 'en' : 'es')}>
      cambiar idioma
    </button>
  );
}

function renderBlock() {
  return render(
    <AppProviders>
      <LanguageProvider>
        <LanguageProbe />
        <MemoryRouter initialEntries={['/']}>
          <ProximamenteBlock />
        </MemoryRouter>
      </LanguageProvider>
    </AppProviders>,
  );
}

describe('ProximamenteBlock · fila de tarjetas (FR-025/FR-026)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('pinta exactamente los ítems recibidos, en orden y con su etiqueta de concepto', async () => {
    responder({
      lang: 'es',
      items: [
        ...cinco,
        makeItem({ id: 'ev-4', title: 'Fuera del tope', date: '2026-11-01' }),
      ].slice(0, 5),
    });

    renderBlock();

    expect(
      await screen.findByRole('heading', { level: 2, name: 'Lo que viene' }),
    ).toBeInTheDocument();
    // Espera al dato (el encabezado se pinta también durante la carga).
    await screen.findByRole('link', { name: 'Evento: Vigilia de viernes' });

    const fila = screen.getByRole('list');
    const items = within(fila).getAllByRole('listitem');
    expect(items).toHaveLength(5);
    // Orden de lectura = orden del servidor (por cercanía).
    expect(
      items.map((item) => within(item).getByRole('heading', { level: 3 }).textContent),
    ).toEqual([
      'Vigilia de viernes',
      'Repartir comida',
      'Ensayo del coro',
      'Visita a enfermos',
      'Noche de aposentos',
    ]);
    // Etiqueta de concepto por ítem (eventos y actividades mezclados).
    expect(within(items[0] as HTMLElement).getByText('Evento')).toBeInTheDocument();
    expect(within(items[1] as HTMLElement).getByText('Actividad')).toBeInTheDocument();
    // La imagen del ítem con su `alt` (FR-020) y la fecha con su `<time>` real.
    expect(
      within(items[1] as HTMLElement).getByRole('img', { name: 'Voluntarios repartiendo' }),
    ).toBeVisible();
    expect(within(items[0] as HTMLElement).getByText('sábado 17 de octubre')).toHaveAttribute(
      'datetime',
      '2026-10-17T19:00',
    );
    expect(within(items[0] as HTMLElement).getByText('7:00 p. m. − 9:00 p. m.')).toBeVisible();
    expect(within(items[0] as HTMLElement).getByText('Templo central')).toBeVisible();
  });

  it('la fila es un carrusel accesible (1/2/3 tarjetas) con controles ligados a la lista', async () => {
    responder({ lang: 'es', items: cinco.slice(0, 2) });

    renderBlock();

    const lista = await screen.findByRole('list', {
      name: 'Lo que viene: próximos eventos y actividades de la iglesia',
    });
    expect(lista).toHaveClass('public-carousel');
    expect(screen.getByRole('button', { name: 'Anterior' })).toHaveAttribute(
      'aria-controls',
      lista.id,
    );
    expect(screen.getByRole('button', { name: 'Siguiente' })).toHaveAttribute(
      'aria-controls',
      lista.id,
    );
  });

  it('con menos de 5 no deja huecos ni marcadores vacíos', async () => {
    responder({ lang: 'es', items: cinco.slice(0, 3) });

    renderBlock();

    await screen.findByRole('link', { name: 'Evento: Vigilia de viernes' });
    const fila = screen.getByRole('list');
    expect(within(fila).getAllByRole('listitem')).toHaveLength(3);
    expect(within(fila).queryByText(/próximamente|—/)).toBeNull();
  });

  it('el título de cada ítem enlaza a su bloque de /eventos con nombre accesible completo', async () => {
    responder({ lang: 'es', items: cinco.slice(0, 2) });

    renderBlock();

    const enlaceEvento = await screen.findByRole('link', {
      name: 'Evento: Vigilia de viernes',
    });
    expect(enlaceEvento).toHaveAttribute('href', '/eventos#eventos');

    const enlaceActividad = screen.getByRole('link', { name: 'Actividad: Repartir comida' });
    expect(enlaceActividad).toHaveAttribute('href', '/eventos#actividades');
  });

  it('el botón «Ver más eventos» lleva a /eventos y solo aparece con ítems', async () => {
    responder({ lang: 'es', items: cinco.slice(0, 1) });

    renderBlock();

    expect(await screen.findByRole('link', { name: 'Ver más eventos' })).toHaveAttribute(
      'href',
      '/eventos',
    );
  });

  it('sin ningún ítem muestra el mensaje (sin tarjetas y sin botón)', async () => {
    responder({ lang: 'es' });

    renderBlock();

    expect(
      await screen.findByText('Por el momento no hay eventos ni actividades.'),
    ).toBeInTheDocument();
    expect(screen.getByRole('status')).toHaveTextContent(
      'Por el momento no hay eventos ni actividades.',
    );
    expect(screen.queryByRole('list')).toBeNull();
    expect(screen.queryByRole('listitem')).toBeNull();
    expect(screen.queryByRole('link', { name: 'Ver más eventos' })).toBeNull();
    // El bloque nunca desaparece: encabezado presente (excepción a SC-012).
    expect(screen.getByRole('heading', { level: 2, name: 'Lo que viene' })).toBeInTheDocument();
  });

  it('en inglés la interfaz se traduce y el contenido sin inglés se ve en español', async () => {
    window.localStorage.setItem('ss.lang', 'en');
    responder({
      lang: 'en',
      items: [
        makeItem({ id: 'ev-1', title: 'Friday vigil', date: '2026-10-17', startTime: '19:00' }),
        makeItem({
          kind: 'activity',
          id: 'ac-1',
          title: 'Repartir comida',
          date: '2026-10-18',
          startTime: undefined,
          endTime: undefined,
        }),
      ],
    });

    renderBlock();

    expect(
      await screen.findByRole('heading', { level: 2, name: "What's coming" }),
    ).toBeInTheDocument();
    await screen.findByRole('link', { name: 'Event: Friday vigil' });
    expect(screen.getByRole('link', { name: 'See more events' })).toHaveAttribute(
      'href',
      '/eventos',
    );
    expect(screen.getByText('Event')).toBeInTheDocument();
    expect(screen.getByText('Activity')).toBeInTheDocument();
    // Contenido sin inglés: el servidor resolvió el fallback y el cliente lo pinta.
    expect(screen.getByRole('link', { name: 'Activity: Repartir comida' })).toBeInTheDocument();
    // Fechas en el idioma elegido, con el dato real intacto.
    expect(screen.getByText('Saturday, October 17')).toHaveAttribute(
      'datetime',
      '2026-10-17T19:00',
    );
    expect(screen.getByText('7:00 PM − 9:00 PM')).toBeVisible();
  });

  it('mientras carga muestra su esqueleto de altura acotada', async () => {
    server.use(
      http.get(url, async () => {
        await delay(50);
        return HttpResponse.json({ lang: 'es', items: cinco.slice(0, 1) });
      }),
    );

    renderBlock();

    expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
    expect(
      await screen.findByRole('link', { name: 'Evento: Vigilia de viernes' }),
    ).toBeInTheDocument();
  });

  it('con error y sin caché muestra el mensaje, el aviso y «Volver a intentar»', async () => {
    let fallos = 1;
    server.use(
      http.get(url, () => {
        if (fallos > 0) {
          fallos -= 1;
          return HttpResponse.json(
            { error: { code: 'internal', message: 'Error' } },
            { status: 500 },
          );
        }
        return HttpResponse.json({ lang: 'es', items: cinco.slice(0, 1) });
      }),
    );

    renderBlock();

    expect(await screen.findByText('No pudimos cargar lo que viene.')).toBeInTheDocument();
    expect(screen.getByText('Por el momento no hay eventos ni actividades.')).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Volver a intentar' }));

    expect(
      await screen.findByRole('link', { name: 'Evento: Vigilia de viernes' }),
    ).toBeInTheDocument();
  });

  it('con ítems en caché conserva lo último válido y avisa en discreto', async () => {
    let peticiones = 0;
    server.use(
      http.get(url, () => {
        peticiones += 1;
        if (peticiones === 1) {
          return HttpResponse.json({ lang: 'es', items: cinco.slice(0, 1) });
        }
        return HttpResponse.json(
          { error: { code: 'internal', message: 'Error' } },
          { status: 500 },
        );
      }),
    );

    renderBlock();
    await screen.findByRole('link', { name: 'Evento: Vigilia de viernes' });

    // Un cambio de idioma fuerza una segunda petición que falla.
    await userEvent.click(screen.getByRole('button', { name: 'cambiar idioma' }));

    // El aviso va en el idioma elegido (la interfaz cambió a inglés).
    expect(await screen.findByText(/couldn't refresh/)).toBeInTheDocument();
    // El contenido retenido sigue ahí, con la interfaz ya en inglés.
    expect(screen.getByRole('link', { name: 'Event: Vigilia de viernes' })).toBeInTheDocument();
  });
});
