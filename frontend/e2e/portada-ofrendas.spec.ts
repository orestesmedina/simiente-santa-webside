import { expect, test, type BrowserContext, type Page } from '@playwright/test';
import { apiSend } from './helpers';
import { createAdminContext } from './publicSeed';
import { resetMinisterios, seedMinisterios } from './ministeriosSeed';
import { resetOfrendas, seedOfrendas, type OfrendasSeed } from './ofrendasSeed';

/**
 * E2E del bloque «Ofrendas y diezmos» de la portada (US6, SC-014, quickstart
 * §14): con 8 métodos publicados el bloque muestra **exactamente los 5 primeros**
 * en el orden de la sección (`iban` antes que `sinpe`; `sortOrder`, `id` dentro de
 * cada grupo), con su chip de tipo («Cuenta IBAN»/«SINPE Móvil»), su entidad y su
 * número **íntegro** (nunca elipsis); con menos muestra **exactos** los que hay,
 * **sin huecos ni marcadores**; con **0 métodos publicados el bloque no existe**
 * —sin título, sin mensaje, sin hueco, a diferencia del bloque «Lo que viene» de
 * F4— y la portada queda coherente (FR-022/SC-010); el bloque está
 * **inmediatamente después del bloque de ministerios** (F6) y antes de «Quiénes
 * somos» (Q7); cada ficha lleva su chip y **no es clicable** (el único
 * interactivo es el botón «Ver formas de ofrendar», que lleva a `/ofrendas` en un
 * solo paso); **0 borradores** por ninguna vía (SC-002); a 320 px no hay
 * desplazamiento horizontal (SC-007) y el bloque se recorre con teclado con foco
 * visible (SC-008).
 *
 * Cómo se ejecutan (la base, el rate-limit de login y el navegador son
 * compartidos por toda la suite):
 *
 *   # Tanda 5 — ofrendas (F7): las cuatro specs comparten la siembra/limpieza del
 *   # módulo, así que van juntas y con un solo worker.
 *   npx playwright test --config e2e/playwright.config.ts ofrendas- portada-ofrendas --workers=1
 *
 * La suite completa se reparte en **cinco tandas** porque `POST
 * /api/v1/auth/login` limita a 20 intentos por minuto y por IP: con todo en una
 * sola corrida, los últimos specs reciben `429` y no llegan a entrar.
 *
 *   # Tanda 1 — F2 (acceso y auditoría: la que más inicios de sesión hace)
 *   npx playwright test --config e2e/playwright.config.ts acceso auditoria --workers=1
 *   # Tanda 2 — F4 (agenda)
 *   npx playwright test --config e2e/playwright.config.ts agenda- --workers=1
 *   # Tanda 3 — F3 y F10 (portada, sistema de diseño, tema y estado)
 *   npx playwright test --config e2e/playwright.config.ts portada- diseno- movimiento navegacion- tema- qa- status --workers=1
 *   # Tanda 4 — F6 (ministerios y su bloque de portada)
 *   npx playwright test --config e2e/playwright.config.ts ministerios- portada-ministerios --workers=1
 *   # Tanda 5 — F7 (ofrendas)
 *   npx playwright test --config e2e/playwright.config.ts ofrendas- portada-ofrendas --workers=1
 *
 * `--workers=1` es obligatorio en las cinco (singletons compartidos y contenido
 * sembrado por spec). En este entorno el navegador no arranca desde el host
 * —faltan las librerías del sistema (`libnspr4`)—, así que los comandos se
 * ejecutan dentro de la imagen oficial de Playwright, con red del host para
 * llegar al stack de `make up`:
 *
 *   docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
 *     -v "$PWD/..":/work -w /work/frontend mcr.microsoft.com/playwright:v1.63.0-jammy \
 *     npx playwright test --config e2e/playwright.config.ts ofrendas- portada-ofrendas --workers=1
 *
 * (El stack debe estar reconstruido con F7: `docker compose up -d --build` y
 * `make db-migrate`.)
 */
test.describe.configure({ mode: 'serial' });

const BLOCK = '#ofrendas';
const CTA = 'Ver formas de ofrendar';
const IBAN_BADGE = 'Cuenta IBAN';
const SINPE_BADGE = 'SINPE Móvil';

/** Tarjetas del bloque (una por método del extracto). */
function blockCards(page: Page) {
  return page.locator(`${BLOCK} ul > li`);
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

test.describe('Bloque «Ofrendas y diezmos» de la portada (US6, SC-014)', () => {
  test.setTimeout(300_000);

  let adminContext: BrowserContext;
  let adminPage: Page;
  let seed: OfrendasSeed;

  test.beforeAll(async ({ browser }) => {
    adminContext = await createAdminContext(browser);
    adminPage = await adminContext.newPage();
    // El bloque de ofrendas va **después** del de ministerios (Q7): se publican
    // ministerios para que ese bloque exista y la posición sea comprobable.
    await seedMinisterios(adminPage, { publishedCount: 2, withDrafts: false });
    await apiSend(adminPage, 'PUT', '/api/v1/admin/portada/quienes-somos', {
      textEs: 'Somos la iglesia del barrio.',
      publicationState: 'published',
    });
    seed = await seedOfrendas(adminPage);
  });

  test.afterAll(async () => {
    if (adminPage) {
      await resetOfrendas(adminPage);
      await resetMinisterios(adminPage);
    }
    await adminContext?.close();
  });

  test('con 8 publicados muestra exactamente los 5 primeros, en orden y en su posición (SC-014)', async ({
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
      await expect(
        page.getByRole('heading', { level: 2, name: 'Ofrendas y diezmos' }),
      ).toBeVisible();
      const cards = blockCards(page);
      await expect(cards).toHaveCount(5);

      // Orden de la sección: los 5 primeros del listado publicado (iban → sinpe).
      expect(await cards.locator('h3').allInnerTexts()).toEqual(
        seed.portada.map((item) => item.institution),
      );
      // El sexto en adelante no aparece.
      for (const item of seed.published.slice(5)) {
        await expect(page.locator(BLOCK).getByText(item.institution)).toHaveCount(0);
      }
      // 0 borradores por ninguna vía (SC-002).
      const html = await page.evaluate(() => document.documentElement.outerHTML);
      for (const draft of seed.drafts) {
        expect(html, `el borrador «${draft.key}» no debe aparecer`).not.toContain(
          draft.institution,
        );
        expect(html).not.toContain(draft.number);
      }

      // ── Chip de tipo, entidad y número íntegro; sin copia por ficha ──────
      for (const item of seed.portada) {
        const card = cards.filter({ hasText: item.institution });
        await expect(
          card.getByText(item.kind === 'iban' ? IBAN_BADGE : SINPE_BADGE, {
            exact: true,
          }),
        ).toBeVisible();
        // La ficha presenta datos; el único interactivo del bloque es su botón.
        await expect(card.getByRole('link')).toHaveCount(0);
        await expect(card.getByRole('button')).toHaveCount(0);
        await expect(card.getByText('—', { exact: true })).toHaveCount(0);
        // El número aparece íntegro (agrupado en IBAN, normalizado en SINPE).
        const presentado =
          item.kind === 'iban' ? item.number.replace(/(.{4})/g, '$1 ').trim() : item.number;
        await expect(card.getByText(presentado)).toBeVisible();
        if (item.holderEs) {
          await expect(card).toContainText(item.holderEs);
        }
        if (item.kind === 'sinpe') {
          await expect(card).not.toContainText('Moneda');
          await expect(card).not.toContainText('Dólares');
          await expect(card).not.toContainText('Colones');
        }
      }

      // ── Posición: después de ministerios y antes de «Quiénes somos» ──────
      expect(
        await domOrder(page, ['ministerios', 'ofrendas', 'quienes-somos']),
        'el bloque va inmediatamente después de ministerios y antes de «quiénes somos»',
      ).toEqual([true, true, true]);

      // ── Un solo paso a `/ofrendas` desde el botón ────────────────────────
      const cta = page.locator(BLOCK).getByRole('link', { name: CTA });
      await expect(cta).toHaveAttribute('href', '/ofrendas');
      await cta.click();
      await expect(page).toHaveURL(/\/ofrendas$/);
      await expect(
        page.getByRole('heading', { level: 1, name: 'Ofrendas y diezmos' }),
      ).toBeVisible();
    } finally {
      await context.close();
    }
  });

  test('con menos publicados muestra exactos, sin huecos ni marcadores', async ({ browser }) => {
    seed = await seedOfrendas(adminPage, { publishedCount: 3, withDrafts: false });

    const context = await browser.newContext();
    const page = await context.newPage();
    try {
      await page.goto('/');
      await waitForPortada(page);

      const cards = blockCards(page);
      await expect(cards).toHaveCount(3);
      expect(await cards.locator('h3').allInnerTexts()).toEqual(
        seed.published.map((item) => item.institution),
      );

      // Sin marcadores de relleno en ninguna tarjeta.
      for (let index = 0; index < 3; index += 1) {
        await expect(cards.nth(index).getByText('—', { exact: true })).toHaveCount(0);
      }
      // El CTA sigue presente y la posición no cambia.
      await expect(page.locator(BLOCK).getByRole('link', { name: CTA })).toBeVisible();
      expect(await domOrder(page, ['ministerios', 'ofrendas', 'quienes-somos'])).toEqual([
        true,
        true,
        true,
      ]);
    } finally {
      await context.close();
    }
  });

  test('sin métodos publicados el bloque no existe y la portada queda coherente (FR-022)', async ({
    browser,
  }) => {
    // Solo textos publicados: el bloque **no** se dibuja (solo mira métodos).
    seed = await seedOfrendas(adminPage, {
      publishedCount: 0,
      withDrafts: false,
      texts: 'published',
    });
    expect(seed.published).toHaveLength(0);

    const context = await browser.newContext();
    const page = await context.newPage();
    try {
      await page.goto('/');
      await waitForPortada(page);

      // Sin título, sin mensaje y sin hueco: el bloque **no se dibuja**.
      await expect(page.locator(BLOCK)).toHaveCount(0);
      await expect(page.getByRole('heading', { level: 2, name: 'Ofrendas y diezmos' })).toHaveCount(
        0,
      );
      await expect(page.getByRole('link', { name: CTA })).toHaveCount(0);
      await expect(page.getByText('Las formas de apoyar a la iglesia')).toHaveCount(0);
      // A diferencia de «Lo que viene» (F4), que sí mantiene su mensaje.
      await expect(page.locator('#proximamente')).toBeVisible();

      // La portada sigue coherente y navegable.
      await expect(page.locator('#quienes-somos')).toBeVisible();
      await expect(page.getByRole('navigation', { name: 'Secciones de la portada' })).toBeVisible();
      expect(await hasNoHorizontalScroll(page)).toBe(true);
    } finally {
      await context.close();
    }
  });

  test('320 px sin scroll horizontal y recorrido por teclado con foco visible (SC-007/SC-008)', async ({
    browser,
  }) => {
    seed = await seedOfrendas(adminPage);

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

      // El bloque se apila en una columna a 320 px.
      const columns = await page.evaluate(
        () =>
          getComputedStyle(document.querySelector('#ofrendas ul') as HTMLElement)
            .gridTemplateColumns,
      );
      expect(columns.trim().split(/\s+/).length, 'una sola columna a 320 px').toBe(1);

      // El número más largo (IBAN internacional) se lee completo, sin elipsis.
      const internacional = seed.portada.find((item) => item.number.length === 24)!;
      const numero = page
        .locator(BLOCK)
        .getByText(internacional.number.replace(/(.{4})/g, '$1 ').trim())
        .first();
      const box = await numero.evaluate((el) => ({
        scrollWidth: el.scrollWidth,
        clientWidth: el.clientWidth,
        textOverflow: getComputedStyle(el).textOverflow,
      }));
      expect(box.textOverflow, 'sin elipsis a 320 px').not.toBe('ellipsis');
      expect(box.scrollWidth).toBeLessThanOrEqual(box.clientWidth + 1);

      // Foco visible en el único interactivo del bloque.
      const cta = page.locator(BLOCK).getByRole('link', { name: CTA });
      await cta.focus();
      expect(await hasFocusIndicator(page), 'foco visible en el CTA del bloque').toBe(true);
    } finally {
      await context.close();
    }
  });
});
