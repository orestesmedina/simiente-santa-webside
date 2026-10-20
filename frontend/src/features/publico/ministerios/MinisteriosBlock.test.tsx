import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { ItemMinisterioPortada, PortadaMinisterios } from '../../../api/ministerios';
import { AppProviders } from '../../../app/providers';
import { server } from '../../../test/server';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { useLanguage } from '../i18n/useLanguage';
import { MinisteriosBlock } from './MinisteriosBlock';

/**
 * Bloque «Ministerios» de la portada (US6/FR-027/FR-028, ux.md §4.9): hasta 5
 * tarjetas compactas con el botón «Ver más ministerios» y, **sin publicados, el
 * bloque no se dibuja** (sin título, sin mensaje, sin hueco — diferencia
 * deliberada con `ProximamenteBlock` de F4).
 */

const url = `${API_BASE_URL}/api/v1/portada/ministerios`;

function makeItem(overrides: Partial<ItemMinisterioPortada> = {}): ItemMinisterioPortada {
  return { id: 'min-1', name: 'Alabanza', ...overrides };
}

/** Ocho publicados en el orden manual del servidor: el bloque solo recibe 5. */
const ocho: ItemMinisterioPortada[] = Array.from({ length: 8 }, (_, index) =>
  makeItem({ id: `min-${index + 1}`, name: `Ministerio ${index + 1}` }),
);

function responder(data: PortadaMinisterios) {
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
          <div data-testid="bloque">
            <MinisteriosBlock />
          </div>
        </MemoryRouter>
      </LanguageProvider>
    </AppProviders>,
  );
}

describe('MinisteriosBlock · extracto de la portada (US6/FR-027)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => window.localStorage.clear());

  it('pinta exactamente los 5 primeros ítems que entrega el servidor, en orden', async () => {
    responder({ lang: 'es', items: ocho.slice(0, 5) });

    renderBlock();

    expect(
      await screen.findByRole('heading', { level: 2, name: 'Ministerios' }),
    ).toBeInTheDocument();
    expect(screen.getByText('Las áreas donde puedes servir')).toBeVisible();

    const fila = await screen.findByRole('list');
    const items = within(fila).getAllByRole('listitem');
    expect(items).toHaveLength(5);
    expect(
      items.map((item) => within(item).getByRole('heading', { level: 3 }).textContent),
    ).toEqual(['Ministerio 1', 'Ministerio 2', 'Ministerio 3', 'Ministerio 4', 'Ministerio 5']);
    // El 6.º no aparece ni con paginación: el tope es del servidor (FR-027).
    expect(screen.queryByText('Ministerio 6')).toBeNull();
  });

  it('la fila es un carrusel accesible con controles ligados a la lista', async () => {
    responder({ lang: 'es', items: ocho.slice(0, 2) });

    renderBlock();

    const lista = await screen.findByRole('list', {
      name: 'Ministerios de la iglesia: mira el resumen y entra a la sección completa',
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
    responder({ lang: 'es', items: ocho.slice(0, 3) });

    const { container } = renderBlock();

    const fila = await screen.findByRole('list');
    expect(within(fila).getAllByRole('listitem')).toHaveLength(3);
    expect(container.textContent).not.toContain('—');
    expect(screen.queryByText(/próximamente|sin datos/i)).toBeNull();
  });

  it('el botón «Ver más ministerios» lleva a /ministerios y solo aparece con ítems', async () => {
    responder({ lang: 'es', items: ocho.slice(0, 1) });

    renderBlock();

    expect(await screen.findByRole('link', { name: 'Ver más ministerios' })).toHaveAttribute(
      'href',
      '/ministerios',
    );
  });

  it('con 0 publicados el bloque no se dibuja: sin título, sin mensaje, sin hueco (FR-028)', async () => {
    let peticiones = 0;
    server.use(
      http.get(url, () => {
        peticiones += 1;
        return HttpResponse.json({ lang: 'es' });
      }),
    );

    renderBlock();

    // Se espera a que la consulta haya respondido: con el dato en mano el bloque
    // sigue sin dibujar nada (ni título, ni mensaje, ni hueco).
    await waitFor(() => expect(peticiones).toBe(1));
    await waitFor(() => expect(screen.getByTestId('bloque')).toBeEmptyDOMElement());
    expect(screen.queryByRole('heading')).toBeNull();
    expect(screen.queryByRole('list')).toBeNull();
    expect(screen.queryByRole('link', { name: 'Ver más ministerios' })).toBeNull();
    expect(screen.queryByRole('status')).toBeNull();
    expect(screen.queryByText(/no hay|vací|no pudimos/i)).toBeNull();
  });

  it('las tarjetas no muestran encargado ni botón de unirse (D-12)', async () => {
    responder({
      lang: 'es',
      items: [
        makeItem({ name: 'Alabanza', schedule: 'Todos los viernes', place: 'Templo' }),
        makeItem({ id: 'min-2', name: 'Escuela dominical' }),
      ],
    });

    renderBlock();

    const fila = await screen.findByRole('list');
    const items = within(fila).getAllByRole('listitem');
    for (const item of items) {
      // Un solo enlace por tarjeta (el título); ningún botón ni enlace externo.
      expect(within(item).getAllByRole('link')).toHaveLength(1);
      expect(within(item).queryByText(/Unirse|Encargado/)).toBeNull();
    }
  });

  it('en inglés la interfaz se traduce y el contenido sin inglés se ve en español (US6 esc. 7)', async () => {
    window.localStorage.setItem('ss.lang', 'en');
    responder({
      lang: 'en',
      items: [makeItem({ name: 'Escuela dominical', schedule: 'Domingos, 9:00 a. m.' })],
    });

    renderBlock();

    expect(
      await screen.findByRole('heading', { level: 2, name: 'Ministries' }),
    ).toBeInTheDocument();
    expect(screen.getByText('The areas where you can serve')).toBeVisible();
    expect(screen.getByRole('link', { name: 'See more ministries' })).toHaveAttribute(
      'href',
      '/ministerios',
    );
    expect(screen.getByText('Ministry')).toBeVisible();
    // Fallback del servidor: el nombre y el horario sin inglés se ven en español.
    expect(screen.getByRole('link', { name: 'Ministry: Escuela dominical' })).toBeVisible();
    expect(screen.getByText('Domingos, 9:00 a. m.')).toBeVisible();
  });

  it('mientras carga muestra su esqueleto propio y acotado', async () => {
    server.use(
      http.get(url, async () => {
        await delay(50);
        return HttpResponse.json({ lang: 'es', items: ocho.slice(0, 2) });
      }),
    );

    renderBlock();

    const esqueleto = screen.getByRole('status');
    expect(esqueleto).toHaveAttribute('aria-busy', 'true');
    expect(screen.getByText('Cargando…')).toBeInTheDocument();
    expect(screen.queryByRole('heading')).toBeNull();

    expect(
      await screen.findByRole('heading', { level: 2, name: 'Ministerios' }),
    ).toBeInTheDocument();
  });

  it('con error y sin caché no dibuja nada (ni título huérfano, D-10)', async () => {
    let peticiones = 0;
    server.use(
      http.get(url, () => {
        peticiones += 1;
        return HttpResponse.json(
          { error: { code: 'internal', message: 'Error' } },
          { status: 500 },
        );
      }),
    );

    renderBlock();

    await waitFor(() => expect(peticiones).toBe(1));
    await waitFor(() => expect(screen.getByTestId('bloque')).toBeEmptyDOMElement());
    expect(screen.queryByRole('heading')).toBeNull();
    expect(screen.queryByRole('alert')).toBeNull();
  });

  it('con ítems en caché conserva lo último válido y avisa en discreto', async () => {
    let peticiones = 0;
    server.use(
      http.get(url, () => {
        peticiones += 1;
        if (peticiones === 1) {
          return HttpResponse.json({ lang: 'es', items: ocho.slice(0, 1) });
        }
        return HttpResponse.json(
          { error: { code: 'internal', message: 'Error' } },
          { status: 500 },
        );
      }),
    );

    renderBlock();
    await screen.findByRole('listitem');

    await userEvent.click(screen.getByRole('button', { name: 'cambiar idioma' }));

    expect(await screen.findByText(/couldn't refresh/)).toBeVisible();
    expect(screen.getByRole('listitem')).toBeVisible();
  });
});
