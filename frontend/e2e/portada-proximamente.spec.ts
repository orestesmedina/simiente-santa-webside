import { expect, test, type BrowserContext, type Page } from '@playwright/test';
import { apiSend } from './helpers';
import { createAdminContext } from './publicSeed';
import { costaRicaToday, resetAgenda, shiftDays, uploadAgendaImage } from './agendaSeed';

/**
 * E2e del bloque «Lo que viene» de la portada (US7, quickstart §15): con 8
 * ítems próximos —eventos y actividades— la portada muestra **exactamente los 5
 * más cercanos** mezclados y etiquetados (SC-014), la **ventana de 2 meses con
 * respaldo** la resuelve el servidor y la vista no cambia (FR-025), el botón
 * «Ver más eventos» y los títulos de cada ítem llegan a `/eventos` (a su bloque)
 * en un solo paso (SC-015), sin nada publicado y próximo el bloque **se muestra
 * con su mensaje y sin botón** (FR-026, excepción de Q8), y el bloque es
 * bilingüe con fallback, responsivo a 320 px (SC-007) y navegable por teclado
 * (SC-008).
 *
 * Cómo se ejecuta (ver `agenda-publica.spec.ts`): los specs de la agenda
 * comparten estado, así que van en la misma tanda con `--workers=1`; en este
 * entorno el navegador se lanza dentro de la imagen oficial de Playwright.
 *
 *   docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
 *     -v "$PWD/..":/work -w /work/frontend mcr.microsoft.com/playwright:v1.63.0-jammy \
 *     npx playwright test --config e2e/playwright.config.ts portada-proximamente --workers=1
 */

test.describe.configure({ mode: 'serial' });

const HOME = '/';

interface SeedItem {
  kind: 'event' | 'activity';
  title: string;
  titleEn?: string;
  date: string;
  startTime?: string;
  imageFile?: string;
  imageAlt?: string;
  state?: 'draft' | 'published';
}

/** Crea un ítem de agenda por API (publicado y próximo salvo indicación). */
async function createItem(page: Page, item: SeedItem): Promise<string> {
  const common = {
    titleEs: item.title,
    titleEn: item.titleEn ?? null,
    detailEs: `Detalle de ${item.title}`,
    placeEs: `Lugar de ${item.title}`,
    publicationState: item.state ?? 'published',
    ...(item.imageFile
      ? { imageFile: item.imageFile, imageAltEs: item.imageAlt ?? `Imagen de ${item.title}` }
      : {}),
  };
  const path = item.kind === 'event' ? '/api/v1/admin/eventos' : '/api/v1/admin/actividades';
  const body =
    item.kind === 'event'
      ? {
          ...common,
          dates: [{ date: item.date, ...(item.startTime ? { startTime: item.startTime } : {}) }],
        }
      : { ...common, date: item.date, ...(item.startTime ? { startTime: item.startTime } : {}) };

  const response = await apiSend(page, 'POST', path, body);
  expect(response.ok(), `POST ${path} «${item.title}» → ${response.status()}`).toBe(true);
  return ((await response.json()) as { id: string }).id;
}

/** Sección del bloque en la portada. */
function block(page: Page): ReturnType<Page['locator']> {
  return page.locator('#proximamente');
}

/** Títulos del extracto, en el orden en que se pintan. */
async function blockTitles(page: Page): Promise<string[]> {
  const links = block(page).getByRole('heading', { level: 3 }).locator('a');
  return links.allTextContents();
}

test.describe('Bloque «Lo que viene» de la portada (US7)', () => {
  test.setTimeout(240_000);

  let adminContext: BrowserContext;
  let adminPage: Page;
  let visitor: BrowserContext;
  let page: Page;

  test.beforeAll(async ({ browser }) => {
    adminContext = await createAdminContext(browser);
    adminPage = await adminContext.newPage();
    await resetAgenda(adminPage);
    visitor = await browser.newContext();
    page = await visitor.newPage();
  });

  test.afterAll(async () => {
    await visitor.close();
    await adminContext.close();
  });

  test('los 5 más cercanos, mezclados y etiquetados; botón y enlaces en un paso (SC-014/SC-015)', async () => {
    const today = costaRicaToday();
    const eventImage = await uploadAgendaImage(adminPage, 'eventos');
    const activityImage = await uploadAgendaImage(adminPage, 'actividades');

    // Ocho ítems próximos, eventos y actividades alternándose por fecha.
    const octavo = `Octavo ${Date.now().toString(36)}`;
    const items: SeedItem[] = [
      {
        kind: 'activity',
        title: 'Comida del mes',
        date: shiftDays(today, 3),
        startTime: '08:00',
        imageFile: activityImage,
        imageAlt: 'Voluntarios repartiendo comida',
      },
      {
        kind: 'event',
        title: 'Vigilia de viernes',
        date: shiftDays(today, 5),
        startTime: '19:00',
        imageFile: eventImage,
        imageAlt: 'Templo iluminado',
      },
      { kind: 'activity', title: 'Visita a enfermos', date: shiftDays(today, 7) },
      { kind: 'event', title: 'Ensayo del coro', date: shiftDays(today, 9) },
      { kind: 'activity', title: 'Taller de jóvenes', date: shiftDays(today, 11) },
      { kind: 'event', title: 'Noche de aposentos', date: shiftDays(today, 13) },
      { kind: 'activity', title: 'Feria de la iglesia', date: shiftDays(today, 15) },
      { kind: 'event', title: octavo, date: shiftDays(today, 17) },
    ];
    const ids: string[] = [];
    for (const item of items) {
      ids.push(await createItem(adminPage, item));
    }

    await page.goto(HOME);

    // Encabezado del bloque y exactamente 5 tarjetas (FR-025: tope firme).
    await expect(page.getByRole('heading', { level: 2, name: 'Lo que viene' })).toBeVisible();
    await expect(block(page).getByRole('listitem')).toHaveCount(5);
    expect(await blockTitles(page)).toEqual(items.slice(0, 5).map((item) => item.title));

    // Mezclados pero identificados, con la imagen y su `alt` (FR-020/FR-026).
    await expect(block(page).getByText('Evento').first()).toBeVisible();
    await expect(block(page).getByText('Actividad').first()).toBeVisible();
    await expect(
      block(page).getByRole('img', { name: 'Voluntarios repartiendo comida' }),
    ).toBeVisible();
    await expect(block(page).getByRole('img', { name: 'Templo iluminado' })).toBeVisible();

    // Los dos más lejanos de los 8 no caben en el extracto.
    await expect(block(page).getByText(items[7].title)).toHaveCount(0);
    await expect(block(page).getByText(items[6].title)).toHaveCount(0);

    // Al retirar el segundo más cercano, el siguiente próximo ocupa su lugar.
    expect((await apiSend(adminPage, 'DELETE', `/api/v1/admin/eventos/${ids[1]}`)).status()).toBe(
      204,
    );
    await page.reload();
    await expect(block(page).getByRole('listitem')).toHaveCount(5);
    expect(await blockTitles(page)).toEqual(
      [items[0], items[2], items[3], items[4], items[5]].map((item) => item.title),
    );
    await expect(block(page).getByText(items[1].title)).toHaveCount(0);
  });

  test('ventana de 2 meses con respaldo y mensaje sin ítems (FR-025/FR-026)', async () => {
    await resetAgenda(adminPage);
    const today = costaRicaToday();

    // Un ítem dentro de la ventana y otro muy lejano (más de 2 meses).
    await createItem(adminPage, { kind: 'event', title: 'Cercano', date: shiftDays(today, 10) });
    await createItem(adminPage, { kind: 'event', title: 'Lejano', date: shiftDays(today, 100) });

    await page.goto(HOME);
    await expect(block(page).getByRole('listitem')).toHaveCount(1);
    expect(await blockTitles(page)).toEqual(['Cercano']);

    // Sin nada en la ventana, el respaldo trae los siguientes próximos.
    const listado = (await (
      await apiSend(adminPage, 'GET', '/api/v1/admin/eventos?limit=100')
    ).json()) as { items: { id: string; titleEs: string }[] };
    const cercano = listado.items.find((item) => item.titleEs === 'Cercano');
    expect(cercano, 'el evento cercano recién creado').toBeDefined();
    expect(
      (await apiSend(adminPage, 'DELETE', `/api/v1/admin/eventos/${cercano?.id}`)).status(),
    ).toBe(204);

    await page.reload();
    await expect(block(page).getByRole('listitem')).toHaveCount(1);
    expect(await blockTitles(page)).toEqual(['Lejano']);
  });

  test('sin nada publicado y próximo: mensaje sin botón y navegación coherente (FR-026/FR-022)', async () => {
    await resetAgenda(adminPage);
    const today = costaRicaToday();

    // Un borrador próximo y un publicado ya vencido: ninguno debe verse.
    await createItem(adminPage, {
      kind: 'event',
      title: 'Borrador próximo',
      date: shiftDays(today, 4),
      state: 'draft',
    });
    await createItem(adminPage, {
      kind: 'activity',
      title: 'Actividad vencida',
      date: shiftDays(today, -3),
    });

    await page.goto(HOME);

    // El bloque se muestra con su mensaje (excepción de Q8) y sin botón.
    await expect(page.getByRole('heading', { level: 2, name: 'Lo que viene' })).toBeVisible();
    await expect(block(page).getByRole('status')).toHaveText(
      'Por el momento no hay eventos ni actividades.',
    );
    await expect(block(page).getByRole('listitem')).toHaveCount(0);
    await expect(block(page).getByRole('link', { name: 'Ver más eventos' })).toHaveCount(0);
    await expect(page.getByText('Borrador próximo')).toHaveCount(0);
    await expect(page.getByText('Actividad vencida')).toHaveCount(0);

    // Coherencia con SC-012: sin nada publicado y próximo, la entrada de
    // navegación de la sección sigue oculta (barra y mapa del sitio del pie).
    await expect(page.getByRole('link', { name: 'Eventos y actividades' })).toHaveCount(0);
  });

  test('inglés con fallback, 320 px sin scroll lateral y teclado con foco visible', async () => {
    await resetAgenda(adminPage);
    const today = costaRicaToday();

    await createItem(adminPage, {
      kind: 'event',
      title: 'Vigilia de viernes',
      titleEn: 'Friday vigil',
      date: shiftDays(today, 6),
      startTime: '19:00',
    });
    // Sin inglés: el servidor resuelve el fallback y el cliente lo pinta tal cual.
    await createItem(adminPage, {
      kind: 'activity',
      title: 'Repartir comida',
      date: shiftDays(today, 8),
    });

    // ── Inglés ──────────────────────────────────────────────────────────────
    await page.goto(HOME);
    await expect(block(page).getByRole('listitem')).toHaveCount(2);
    await page.locator('header').getByRole('radio', { name: 'Inglés' }).check();

    await expect(page.getByRole('heading', { level: 2, name: "What's coming" })).toBeVisible();
    await expect(block(page).getByRole('link', { name: 'See more events' })).toHaveAttribute(
      'href',
      '/eventos',
    );
    await expect(block(page).getByRole('link', { name: 'Event: Friday vigil' })).toBeVisible();
    await expect(
      block(page).getByRole('link', { name: 'Activity: Repartir comida' }),
    ).toBeVisible();
    await expect(page.locator('html')).toHaveAttribute('lang', 'en');

    await page.locator('header').getByRole('radio', { name: 'Spanish' }).check();
    await expect(page.locator('html')).toHaveAttribute('lang', 'es');

    // ── 320 px: carrusel de una tarjeta por página, sin scroll lateral (SC-007) ─
    await page.setViewportSize({ width: 320, height: 900 });
    await page.reload();
    await expect(block(page).getByRole('listitem')).toHaveCount(2);
    const [primera, segunda] = await block(page).getByRole('listitem').all();
    const cajaPrimera = await primera.boundingBox();
    const cajaSegunda = await segunda.boundingBox();
    expect(cajaPrimera).not.toBeNull();
    expect(cajaSegunda).not.toBeNull();
    // Las tarjetas van en la misma fila, una por página: la segunda queda a la
    // derecha (dentro del carrusel), no apilada debajo.
    expect(Math.round(cajaSegunda!.y)).toBe(Math.round(cajaPrimera!.y));
    expect(cajaSegunda!.x).toBeGreaterThan(cajaPrimera!.x + cajaPrimera!.width);
    const medidas = await block(page)
      .locator('ul.public-carousel')
      .evaluate((el) => {
        const tarjeta = el.querySelector('li') as HTMLElement;
        return {
          visible: el.clientWidth,
          tarjeta: tarjeta.getBoundingClientRect().width,
          total: el.scrollWidth,
        };
      });
    expect(Math.abs(medidas.tarjeta - medidas.visible), 'una tarjeta visible').toBeLessThanOrEqual(
      2,
    );
    expect(medidas.total, 'la segunda tarjeta vive dentro del carrusel').toBeGreaterThan(
      medidas.visible,
    );
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth),
      'sin scroll horizontal a 320 px',
    ).toBeLessThanOrEqual(1);

    // ── Teclado: los enlaces del bloque reciben foco visible (SC-008) ───────
    await page.keyboard.press('Tab');
    await expect(page.getByRole('link', { name: 'Saltar al contenido principal' })).toBeFocused();

    const enlaceEvento = block(page).getByRole('link', { name: 'Evento: Vigilia de viernes' });
    await enlaceEvento.focus();
    await expect(enlaceEvento).toBeFocused();
    const estilo = await enlaceEvento.evaluate((element) => {
      const style = getComputedStyle(element);
      return {
        outlineStyle: style.outlineStyle,
        outlineWidth: parseFloat(style.outlineWidth),
      };
    });
    expect(
      estilo.outlineStyle !== 'none' && estilo.outlineWidth > 0,
      'el título-enlace del ítem muestra el anillo de foco',
    ).toBe(true);

    // ── En un solo paso: el título lleva a su bloque de /eventos (SC-015) ───
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.goto(HOME);
    await block(page).getByRole('link', { name: 'Actividad: Repartir comida' }).click();
    await expect(page).toHaveURL(/\/eventos#actividades$/);
    await expect(page.locator('#actividades')).toBeVisible();
    await expect(page.locator('#actividades')).toBeInViewport();

    // Y el botón lleva a la sección completa.
    await page.goto(HOME);
    await block(page).getByRole('link', { name: 'Ver más eventos' }).click();
    await expect(page).toHaveURL(/\/eventos$/);
    await expect(
      page.getByRole('heading', { level: 1, name: 'Eventos y actividades' }),
    ).toBeVisible();
  });
});
