import { expect, test, type BrowserContext, type Page } from '@playwright/test';
import { apiSend } from './helpers';
import { createAdminContext } from './publicSeed';
import { resetOfrendas, seedOfrendas } from './ofrendasSeed';
import { resetMedios, seedMedios, type MediosSeed } from './mediosSeed';

/**
 * E2E del bloque «Medios» de la portada (US6/SC-015, quickstart §10): con **más
 * de 5** publicables muestra **exactamente los 5 más recientes** en el orden de la
 * sección (tope firme del servidor); el bloque de portada es un **carrusel de
 * 1/2/3 tarjetas** según el ancho (móvil/tableta/escritorio); con menos,
 * **exactos los que hay, sin huecos ni marcadores vacíos**; con **0 publicables el
 * bloque no existe** —sin título, sin mensaje, sin reservar espacio— y la portada
 * queda coherente (FR-027); su **posición** es inmediatamente **después del
 * bloque de ofrendas y diezmos** (Q12) y antes de la siguiente sección de F3;
 * cada ficha compacta lleva sus **enlaces externos anunciados** (FR-004/FR-005) y
 * **omite la fecha** (⟲ D-32); el botón **«Ver todo»** lleva a `/medios` en un solo
 * paso; **0 ocultos ni duplicados abiertos** aparecen por ninguna vía (SC-002); y
 * a 320 px no hay desplazamiento horizontal (SC-017) con recorrido por teclado y
 * foco visible (SC-008).
 *
 * Cómo se ejecutan (la base, el rate-limit de login y el navegador son
 * compartidos por toda la suite):
 *
 *   # Tanda 6 — medios (F9): las cuatro specs comparten la siembra/limpieza del
 *   # módulo (y la carátula subida), así que van juntas y con un solo worker.
 *   cd frontend && npx playwright test --config e2e/playwright.config.ts medios- portada-medios --workers=1
 *
 * (El reparto completo de la suite y el motivo de las tandas están en
 * `medios-publica.spec.ts`.) Para comprobar la **posición** el bloque de ofrendas
 * tiene que existir: se siembran dos métodos (patrón de `portada-ofrendas.spec.ts`
 * con ministerios) y se retiran al terminar.
 */
test.describe.configure({ mode: 'serial' });

const BLOCK = '#medios';
const CTA = 'Ver todo';
const BLOCK_LIST = 'Medios: mira los episodios recientes y entra a la sección completa';
const SUBTITLE = 'Las últimas prédicas y podcasts';
const LIST = 'Lista de episodios de medios';

/** Fichas compactas del bloque (una por episodio del extracto). */
function blockCards(page: Page) {
  return page.getByRole('list', { name: BLOCK_LIST }).getByRole('listitem');
}

/** `true` si el documento no tiene desplazamiento horizontal. */
async function hasNoHorizontalScroll(page: Page): Promise<boolean> {
  return page.evaluate(
    () => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1,
  );
}

/** `true` si el elemento enfocado pinta un indicador de foco visible. */
async function hasFocusIndicator(page: Page): Promise<boolean> {
  return page.evaluate(() => {
    const el = document.activeElement;
    if (!el || el === document.body) {
      return false;
    }
    const style = getComputedStyle(el);
    return (
      (style.outlineStyle !== 'none' && parseFloat(style.outlineWidth) > 0) ||
      style.boxShadow !== 'none'
    );
  });
}

/** Espera a que la portada esté pintada (hero/bloque de eventos de F4). */
async function waitForPortada(page: Page): Promise<void> {
  await expect(page.locator('#proximamente')).toBeVisible();
}

/**
 * Comprueba el orden de aparición en el DOM de una secuencia de `id`s: cada uno
 * sigue al anterior (`compareDocumentPosition` + `DOCUMENT_POSITION_FOLLOWING`).
 */
async function domOrder(page: Page, ids: string[]): Promise<boolean[]> {
  return page.evaluate((list) => {
    const nodes = list.map((id) => document.getElementById(id));
    return nodes.map((node, index) => {
      if (index === 0) {
        return node !== null;
      }
      const previous = nodes[index - 1];
      if (!previous || !node) {
        return false;
      }
      return Boolean(previous.compareDocumentPosition(node) & Node.DOCUMENT_POSITION_FOLLOWING);
    });
  }, ids);
}

test.describe('Bloque «Medios» de la portada (US6, SC-015)', () => {
  test.setTimeout(300_000);

  let adminContext: BrowserContext;
  let adminPage: Page;
  let seed: MediosSeed;

  test.beforeAll(async ({ browser }) => {
    adminContext = await createAdminContext(browser);
    adminPage = await adminContext.newPage();
    // El bloque de medios va **después** del de ofrendas (Q12): se publican dos
    // métodos para que ese bloque exista y la posición sea comprobable.
    await seedOfrendas(adminPage, { publishedCount: 2, withDrafts: false });
    await apiSend(adminPage, 'PUT', '/api/v1/admin/portada/quienes-somos', {
      textEs: 'Somos la iglesia del barrio.',
      publicationState: 'published',
    });
    seed = await seedMedios(adminPage);
  });

  test.afterAll(async () => {
    if (adminPage) {
      resetMedios();
      await resetOfrendas(adminPage);
    }
    await adminContext?.close();
  });

  test('con más de 5 publicables muestra exactamente los 5 primeros, en orden y en su posición (SC-015)', async ({
    browser,
  }) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    try {
      await page.setViewportSize({ width: 1280, height: 900 });
      await page.goto('/');
      await waitForPortada(page);

      // ── Exactamente 5 (tope firme del servidor) ──────────────────────────
      await expect(page.locator(BLOCK)).toBeVisible();
      await expect(page.getByRole('heading', { level: 2, name: 'Medios' })).toBeVisible();
      await expect(page.getByText(SUBTITLE)).toBeVisible();
      const cards = blockCards(page);
      await expect(cards).toHaveCount(5);

      // Orden de la sección: los 5 primeros del listado publicable.
      expect(await cards.locator('h3').allInnerTexts()).toEqual(
        seed.portada.map((item) => item.title),
      );
      // El sexto en adelante no aparece.
      for (const item of seed.publicables.slice(5)) {
        await expect(page.locator(BLOCK).getByText(item.title)).toHaveCount(0);
      }

      // ── 0 ocultos ni duplicados abiertos por ninguna vía (SC-002) ───────
      const html = await page.evaluate(() => document.documentElement.outerHTML);
      for (const oculta of seed.ocultas) {
        expect(html, `la ficha «${oculta.key}» no debe aparecer`).not.toContain(oculta.title);
      }

      // ── Cada ficha: carátula (o marcador neutro), nombre, descripción y
      //    sus enlaces externos anunciados; **sin** fecha (⟲ D-32) ─────────
      for (const item of seed.portada) {
        const card = cards.filter({ hasText: item.title });
        await expect(card.locator('h3')).toHaveText(item.title);
        if (item.hasCover) {
          const cover = card.locator('img[alt^="Carátula del episodio"]');
          await expect(cover).toHaveCount(1);
          await expect(cover).toHaveAttribute('alt', `Carátula del episodio ${item.title}`);
        } else {
          await expect(card.getByTestId('caratula-neutra')).toBeVisible();
        }
        // La descripción corta sí; la fecha **no** (el bloque la omite: un solo
        // `p` en la ficha).
        await expect(card.locator('p')).toHaveCount(1);
        if (item.descriptionEs) {
          await expect(card).toContainText(item.descriptionEs);
        }
        // Enlaces externos: solo los formatos que la ficha tiene, anunciados.
        await expect(card.getByRole('link')).toHaveCount(item.sources.length);
        for (const source of item.sources) {
          const servicio = source.provider === 'youtube' ? 'YouTube' : 'Spotify';
          const link = card.getByRole('link', { name: new RegExp(`^Abre ${servicio}`) });
          await expect(link).toHaveAttribute('href', source.url);
          await expect(link).toHaveAttribute('target', '_blank');
          await expect(link).toHaveAttribute('rel', 'noopener noreferrer');
          const label = (await link.getAttribute('aria-label')) ?? '';
          expect(label).toContain(servicio);
          expect(label).toContain('en una pestaña nueva');
          expect(label).toContain('sales del sitio');
        }
        await expect(card.getByText('—', { exact: true })).toHaveCount(0);
      }

      // ── Posición: después de ofrendas y antes de «Quiénes somos» ─────────
      expect(
        await domOrder(page, ['ofrendas', 'medios', 'quienes-somos']),
        'el bloque va inmediatamente después de ofrendas y antes de «quiénes somos»',
      ).toEqual([true, true, true]);

      // ── Un solo paso a `/medios` desde «Ver todo» ────────────────────────
      const cta = page.locator(BLOCK).getByRole('link', { name: CTA });
      await expect(cta).toHaveAttribute('href', '/medios');
      await cta.click();
      await expect(page).toHaveURL(/\/medios$/);
      await expect(page.getByRole('heading', { level: 1, name: 'Medios' })).toBeVisible();
      await expect(page.getByRole('list', { name: LIST })).toBeVisible();
    } finally {
      await context.close();
    }
  });

  test('con menos publicables muestra exactos, sin huecos ni marcadores', async ({ browser }) => {
    seed = await seedMedios(adminPage, { publishedCount: 3 });

    const context = await browser.newContext();
    const page = await context.newPage();
    try {
      await page.goto('/');
      await waitForPortada(page);

      const cards = blockCards(page);
      await expect(cards).toHaveCount(3);
      expect(await cards.locator('h3').allInnerTexts()).toEqual(
        seed.publicables.map((item) => item.title),
      );

      // Sin marcadores de relleno en ninguna tarjeta.
      for (let index = 0; index < 3; index += 1) {
        await expect(cards.nth(index).getByText('—', { exact: true })).toHaveCount(0);
      }
      // El CTA sigue presente y la posición no cambia.
      await expect(page.locator(BLOCK).getByRole('link', { name: CTA })).toBeVisible();
      expect(await domOrder(page, ['ofrendas', 'medios', 'quienes-somos'])).toEqual([
        true,
        true,
        true,
      ]);
    } finally {
      await context.close();
    }
  });

  test('sin nada publicable el bloque no existe y la portada queda coherente (FR-027)', async ({
    browser,
  }) => {
    seed = await seedMedios(adminPage, { publishedCount: 0 });
    expect(seed.publicables).toHaveLength(0);

    const context = await browser.newContext();
    const page = await context.newPage();
    try {
      await page.goto('/');
      await waitForPortada(page);

      // Sin título, sin mensaje y sin hueco: el bloque **no se dibuja**.
      await expect(page.locator(BLOCK)).toHaveCount(0);
      await expect(page.getByRole('heading', { level: 2, name: 'Medios' })).toHaveCount(0);
      await expect(page.getByRole('link', { name: CTA })).toHaveCount(0);
      await expect(page.getByText(SUBTITLE)).toHaveCount(0);
      // A diferencia de «Lo que viene» (F4), que sí mantiene su mensaje.
      await expect(page.locator('#proximamente')).toBeVisible();

      // La portada sigue coherente y navegable (y la entrada «Medios» no está).
      await expect(page.locator('#quienes-somos')).toBeVisible();
      await expect(page.getByRole('navigation', { name: 'Secciones de la portada' })).toBeVisible();
      await expect(page.locator('a[href="/medios"]')).toHaveCount(0);
      expect(await hasNoHorizontalScroll(page)).toBe(true);
    } finally {
      await context.close();
    }
  });

  test('320 px sin scroll horizontal y recorrido por teclado con foco visible (SC-017/SC-008)', async ({
    browser,
  }) => {
    seed = await seedMedios(adminPage);

    const context = await browser.newContext();
    const page = await context.newPage();
    try {
      await page.setViewportSize({ width: 320, height: 900 });
      await page.goto('/');
      await waitForPortada(page);
      await expect(page.locator(BLOCK)).toBeVisible();
      expect(await hasNoHorizontalScroll(page), 'sin desplazamiento horizontal a 320 px').toBe(
        true,
      );

      // El carrusel muestra una tarjeta por página a 320 px y las demás quedan
      // dentro del propio carrusel (el desborde nunca llega al documento).
      const carrusel = await page.evaluate(() => {
        const lista = document.querySelector('#medios ul.public-carousel') as HTMLElement;
        const tarjeta = lista.querySelector('li') as HTMLElement;
        return {
          visible: lista.clientWidth,
          tarjeta: tarjeta.getBoundingClientRect().width,
          total: lista.scrollWidth,
        };
      });
      expect(
        Math.abs(carrusel.tarjeta - carrusel.visible),
        'una tarjeta visible a 320 px',
      ).toBeLessThanOrEqual(2);
      expect(carrusel.total, 'las tarjetas restantes viven dentro del carrusel').toBeGreaterThan(
        carrusel.visible,
      );

      // Foco visible en el CTA del bloque.
      const cta = page.locator(BLOCK).getByRole('link', { name: CTA });
      await cta.focus();
      expect(await hasFocusIndicator(page), 'foco visible en el CTA del bloque').toBe(true);

      // El bloque se recorre con teclado: la travesía alcanza sus enlaces.
      await page.setViewportSize({ width: 1280, height: 900 });
      await page.goto('/');
      await waitForPortada(page);
      // Espera a que el bloque (y su carrusel) estén montados antes de recorrer
      // el orden de tabulación: si no, el foco puede pasar de largo mientras la
      // consulta sigue en vuelo (el bloque se inserta detrás del foco).
      await expect(page.getByRole('list', { name: BLOCK_LIST })).toBeVisible();
      const alcanzados: string[] = [];
      for (let step = 0; step < 120; step += 1) {
        await page.keyboard.press('Tab');
        const label = await page.evaluate(() => {
          const el = document.activeElement as HTMLElement | null;
          if (!el || el === document.body) return null;
          return (el.innerText || el.getAttribute('aria-label') || '').trim().slice(0, 60);
        });
        if (label === null) {
          break;
        }
        alcanzados.push(label);
        if (label.includes(CTA)) {
          break;
        }
      }
      const recorrido = alcanzados.join(' | ');
      expect(recorrido, 'la travesía por teclado alcanza «Ver todo»').toContain(CTA);
      expect(recorrido, 'y los enlaces externos de las fichas').toContain('Ver en YouTube');
    } finally {
      await context.close();
    }
  });
});
