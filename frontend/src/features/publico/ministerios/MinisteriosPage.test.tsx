import { QueryClientProvider } from '@tanstack/react-query';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import { createMemoryRouter, RouterProvider } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { MinisterioPublico, MinisteriosPublicos } from '../../../api/ministerios';
import { createQueryClientForTest } from '../../../test/queryClient';
import { server } from '../../../test/server';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { useLanguage } from '../i18n/useLanguage';
import { MinisteriosPage } from './MinisteriosPage';
import { PUBLIC_MINISTRIES_QUERY_KEY } from './hooks/usePublicMinistries';

/**
 * Página pública `/ministerios` (US1/US4/US5, ux.md §4.1): H1 + subtítulo + una
 * sola grilla en el orden manual del servidor, cada ficha con solo lo que tiene
 * (SC-017) y **redirección silenciosa a la portada** con 0 publicados (D-14) —
 * sin destello, sin `HiddenNotice`, sin aviso y sin bucle de historial—, además
 * de los estados de carga y de error (que **no** redirige).
 */

const publicUrl = `${API_BASE_URL}/api/v1/ministerios`;

function makeItem(overrides: Partial<MinisterioPublico> = {}): MinisterioPublico {
  return {
    id: 'min-1',
    name: 'Alabanza',
    description: 'Llevamos la adoración en los cultos.',
    ...overrides,
  };
}

function responder(data: MinisteriosPublicos) {
  server.use(http.get(publicUrl, () => HttpResponse.json(data)));
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

function renderPage(
  client = createQueryClientForTest(),
  initialEntries: string[] = ['/ministerios'],
) {
  const router = createMemoryRouter(
    [
      { path: '/', element: <p>portada de la iglesia</p> },
      { path: '/ministerios', element: <MinisteriosPage /> },
    ],
    { initialEntries },
  );

  const view = render(
    <QueryClientProvider client={client}>
      <LanguageProvider>
        <LanguageProbe />
        <RouterProvider router={router} />
      </LanguageProvider>
    </QueryClientProvider>,
  );

  return { ...view, router };
}

describe('MinisteriosPage · grilla pública (FR-002/FR-004)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('pinta H1, subtítulo y una sola grilla en el orden que entrega el servidor', async () => {
    responder({
      lang: 'es',
      items: [
        makeItem({ id: 'min-2', name: 'Escuela dominical' }),
        makeItem({ id: 'min-1', name: 'Alabanza' }),
        makeItem({ id: 'min-3', name: 'Acción social' }),
      ],
    });

    renderPage();

    expect(await screen.findByRole('heading', { level: 1, name: 'Ministerios' })).toBeVisible();
    expect(
      screen.getByText(
        'Las áreas de servicio de nuestra iglesia: conoce cada una y cómo participar.',
      ),
    ).toBeVisible();

    // Orden manual del servidor (`sortOrder, id`): el cliente no reordena.
    const fichas = screen.getAllByRole('listitem');
    expect(fichas).toHaveLength(3);
    expect(
      fichas.map((ficha) => within(ficha).getByRole('heading', { level: 3 }).textContent),
    ).toEqual(['Escuela dominical', 'Alabanza', 'Acción social']);
  });

  it('cada caso «sin dato» se resuelve solo, sin huecos ni líneas vacías (SC-017)', async () => {
    responder({
      lang: 'es',
      items: [
        // Ficha mínima: solo nombre y descripción (sin enlace, sin encargado publicado).
        makeItem({ id: 'min-1', name: 'Matrimonios', description: 'Acompañamos a las parejas.' }),
        // Ficha visible sin contacto: solo el nombre (Q8, no es error).
        makeItem({ id: 'min-2', name: 'Escuela dominical', leader: { name: 'Ana Rosales' } }),
        // Ficha con ficha oculta: el servidor no manda nada del encargado.
        makeItem({ id: 'min-3', name: 'Sonido' }),
        // Ficha completa.
        makeItem({
          id: 'min-4',
          name: 'Alabanza',
          schedule: 'Todos los viernes, 7:00 p. m.',
          place: 'Templo central',
          leader: { name: 'Josué Ramírez', phone: '+504 9999-9999' },
          whatsappUrl: 'https://chat.whatsapp.com/ABC123',
          image: { url: '/api/v1/media/img_alabanza.png', alt: 'Coro cantando' },
        }),
      ],
    });

    const { container } = renderPage();
    await screen.findByRole('heading', { level: 1, name: 'Ministerios' });

    const fichas = screen.getAllByRole('listitem');
    expect(fichas).toHaveLength(4);

    // 1 · solo nombre + descripción: nada más, y sin «—».
    const minima = fichas[0] as HTMLElement;
    expect(within(minima).getByText('Acompañamos a las parejas.')).toBeVisible();
    expect(minima.querySelector('img')).toBeNull();
    expect(minima.querySelector('a')).toBeNull();
    expect(within(minima).queryByText('Ana Rosales')).toBeNull();

    // 2 · ficha visible sin contacto: nombre y ninguna línea de contacto.
    const sinContacto = fichas[1] as HTMLElement;
    expect(within(sinContacto).getByText('Ana Rosales')).toBeVisible();
    expect(within(sinContacto).queryByText(/\+504/)).toBeNull();

    // 3 · ficha oculta: 0 apariciones del nombre y del contacto (SC-014).
    const oculta = fichas[2] as HTMLElement;
    expect(oculta.querySelectorAll('p')).toHaveLength(1);

    // 4 · completa: imagen con alt, horario, lugar, encargado y botón.
    const completa = fichas[3] as HTMLElement;
    expect(within(completa).getByRole('img', { name: 'Coro cantando' })).toBeVisible();
    expect(within(completa).getByText('Todos los viernes, 7:00 p. m.')).toBeVisible();
    expect(within(completa).getByText('Templo central')).toBeVisible();
    expect(
      within(completa).getByRole('link', { name: 'Unirse al grupo de WhatsApp de Alabanza' }),
    ).toHaveAttribute('href', 'https://chat.whatsapp.com/ABC123');

    // Cero marcadores de vacío en toda la página.
    expect(container.textContent).not.toContain('—');
    // Ningún botón dentro de las fichas (sin enlace → sin botón, SC-017).
    expect(within(screen.getByRole('list')).queryAllByRole('button')).toHaveLength(0);
  });

  it('el enlace se sirve tal cual y el horario es texto libre, sin `<time>` (FR-015/FR-024)', async () => {
    responder({
      lang: 'es',
      items: [
        makeItem({
          schedule: 'Una vez al mes, fecha variable',
          whatsappUrl: 'https://ejemplo.com/grupo?ref=1',
        }),
      ],
    });

    const { container } = renderPage();

    const enlace = await screen.findByRole('link', {
      name: 'Unirse al grupo de WhatsApp de Alabanza',
    });
    expect(enlace).toHaveAttribute('href', 'https://ejemplo.com/grupo?ref=1');
    expect(screen.getByText('Una vez al mes, fecha variable')).toBeVisible();
    expect(container.querySelector('time')).toBeNull();
  });

  it('en inglés la interfaz se traduce y el contenido sin inglés se ve en español (US4)', async () => {
    window.localStorage.setItem('ss.lang', 'en');
    responder({
      lang: 'en',
      items: [
        makeItem({
          name: 'Escuela dominical',
          description: 'Cada sábado los niños aprenden juntos.',
          schedule: 'Domingos, 9:00 a. m.',
          whatsappUrl: 'https://chat.whatsapp.com/ABC123',
        }),
      ],
    });

    renderPage();

    expect(await screen.findByRole('heading', { level: 1, name: 'Ministries' })).toBeVisible();
    expect(
      screen.getByText(
        'The service areas of our church: get to know each one and how to take part.',
      ),
    ).toBeVisible();
    expect(
      screen.getByRole('link', { name: 'Join the WhatsApp group of Escuela dominical' }),
    ).toBeVisible();
    // Contenido sin inglés: el servidor resolvió el fallback y el cliente lo pinta.
    expect(screen.getByText('Escuela dominical')).toBeVisible();
    expect(screen.getByText('Domingos, 9:00 a. m.')).toBeVisible();
  });
});

describe('MinisteriosPage · redirección silenciosa con 0 publicados (D-14/P6-21)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('redirige a la portada con `replace`, sin H1, sin grilla, sin aviso ni mensaje', async () => {
    responder({ lang: 'es' });

    const { router } = renderPage();

    expect(await screen.findByText('portada de la iglesia')).toBeVisible();
    expect(router.state.location.pathname).toBe('/');
    // `replace`: la portada no deja `/ministerios` en el historial (sin bucle).
    expect(router.state.historyAction).toBe('REPLACE');

    expect(screen.queryByRole('heading', { level: 1, name: 'Ministerios' })).toBeNull();
    expect(screen.queryByRole('list')).toBeNull();
    expect(screen.queryByRole('listitem')).toBeNull();
    // Ni `HiddenNotice` ni ningún aviso de sección vacía (diferencia con F4).
    expect(screen.queryByRole('alert')).toBeNull();
    expect(screen.queryByText(/no hay|vací|no pudimos/i)).toBeNull();
    expect(screen.queryByRole('button', { name: /Volver a intentar/ })).toBeNull();
  });

  it('con la respuesta ya en caché redirige en el primer render: sin esqueleto ni destello', async () => {
    const client = createQueryClientForTest();
    client.setQueryData<MinisteriosPublicos>([...PUBLIC_MINISTRIES_QUERY_KEY, 'es'], {
      lang: 'es',
    });
    responder({ lang: 'es' });

    renderPage(client);

    // El primer render ya está en la portada: nunca se pintó H1, ni grilla, ni esqueleto.
    expect(screen.getByText('portada de la iglesia')).toBeVisible();
    expect(screen.queryByRole('heading', { level: 1, name: 'Ministerios' })).toBeNull();
    expect(screen.queryByRole('status')).toBeNull();
    expect(screen.queryByRole('list')).toBeNull();
  });

  it('mientras carga solo se ve el esqueleto estándar (nunca la grilla vacía)', async () => {
    server.use(
      http.get(publicUrl, async () => {
        await delay(50);
        return HttpResponse.json({ lang: 'es', items: [makeItem()] });
      }),
    );

    renderPage();

    const esqueleto = screen.getByRole('status');
    expect(esqueleto).toHaveAttribute('aria-busy', 'true');
    expect(screen.queryByRole('heading', { level: 1, name: 'Ministerios' })).toBeNull();
    expect(screen.queryByRole('list')).toBeNull();
    expect(screen.queryByText('portada de la iglesia')).toBeNull();

    expect(await screen.findByRole('heading', { level: 1, name: 'Ministerios' })).toBeVisible();
  });

  it('ante error NO redirige: muestra el estado de error con «Volver a intentar»', async () => {
    let fallos = 1;
    server.use(
      http.get(publicUrl, () => {
        if (fallos > 0) {
          fallos -= 1;
          return HttpResponse.json(
            { error: { code: 'internal', message: 'Error' } },
            { status: 500 },
          );
        }
        return HttpResponse.json({ lang: 'es', items: [makeItem()] });
      }),
    );

    const { router } = renderPage();

    expect(
      await screen.findByText(
        'No pudimos cargar los ministerios. Revisa tu conexión y vuelve a intentarlo.',
      ),
    ).toBeVisible();
    expect(screen.getByRole('heading', { level: 1, name: 'Ministerios' })).toBeVisible();
    expect(screen.queryByText('portada de la iglesia')).toBeNull();
    expect(router.state.location.pathname).toBe('/ministerios');

    await userEvent.click(screen.getByRole('button', { name: 'Volver a intentar' }));
    expect(await screen.findByRole('listitem')).toBeVisible();
  });

  it('con error y contenido en caché conserva la grilla y avisa en discreto', async () => {
    let peticiones = 0;
    server.use(
      http.get(publicUrl, () => {
        peticiones += 1;
        if (peticiones === 1) {
          return HttpResponse.json({ lang: 'es', items: [makeItem()] });
        }
        return HttpResponse.json(
          { error: { code: 'internal', message: 'Error' } },
          { status: 500 },
        );
      }),
    );

    renderPage();
    await screen.findByRole('listitem');

    // Un cambio de idioma fuerza una segunda petición que falla.
    await userEvent.click(screen.getByRole('button', { name: 'cambiar idioma' }));

    expect(await screen.findByText(/couldn't refresh/)).toBeVisible();
    expect(screen.getByRole('listitem')).toBeVisible();
  });
});
