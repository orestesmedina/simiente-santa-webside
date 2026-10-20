import { expect, test, type BrowserContext, type Page } from '@playwright/test';
import { resetMedios, seedMedios } from './mediosSeed';
import { createAdminContext, seedPublicPortada, type PublicSeed } from './publicSeed';

/**
 * E2E de regresión de los bugs de la portada en móvil (2026-10-19, `fix/portada-movil`).
 *
 * **Bug 2 — pantalla en blanco en `/medios` (y en `/ministerios`) en móvil:**
 * `MediosGrid` envuelve la lista completa en un único `Reveal`; con un umbral de
 * intersección fraccional (0.15) un bloque más alto que el viewport (≈ 5.234 px
 * a 390×664, intersección máxima ≈ 11 %) nunca llega al umbral y queda
 * `opacity: 0` / `data-reveal="left"` para siempre. Regresión: **sin hacer
 * scroll**, el `.reveal-block` de la lista debe quedar revelado
 * (`data-reveal="in"`, `opacity: 1`) al cargar `/medios` a 390×844; el H1 es
 * visible, así que la aserción no depende de que la página entera reaccione.
 *
 * **Bug 3 — menú móvil sobre la página desplazada:** con `data-scrolled='true'`
 * la cabecera aplica `backdrop-filter: blur(12px)`, que crea *containing block*
 * para los descendientes `fixed`: el scrim (`mobile-menu-scrim`) queda reducido
 * a la cabecera (~68 px de alto) y tocar fuera del panel, en la zona inferior
 * del viewport, no lo cierra; además el panel `.public-nav-list[data-open]` se
 * pinta con `background: var(--surface-raised)` (alpha 0.08, translúcido).
 * Regresión: con la página desplazada y el menú abierto, el scrim debe cubrir el
 * viewport (≈ 844 px de alto), tocar en la zona inferior debe cerrar el menú y
 * el fondo del panel debe ser opaco (alpha 1).
 *
 * Cómo se ejecuta (rate-limit de login y base compartidos):
 *
 *   cd frontend && npx playwright test --config e2e/playwright.config.ts e2e/qa-portada-movil.spec.ts --workers=1
 */
test.describe.configure({ mode: 'serial' });

const MOBILE = { width: 390, height: 844 };
const LIST = 'Lista de episodios de medios';

/** Sección `/medios`: un solo `.reveal-block` envuelve la lista completa (MediosGrid). */
function revealDeLista(page: Page) {
  return page.locator('.reveal-block').filter({
    has: page.getByRole('list', { name: LIST }),
  });
}

test.describe('Bug 2 · `/medios` revela la lista sin scroll en móvil', () => {
  test.setTimeout(300_000);

  let adminContext: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    adminContext = await createAdminContext(browser);
    const adminPage = await adminContext.newPage();
    try {
      await seedMedios(adminPage);
    } finally {
      await adminPage.close();
    }
  });

  test.afterAll(async () => {
    resetMedios();
    await adminContext?.close();
  });

  test('a 390×844 el bloque de la lista está revelado nada más cargar (sin scroll)', async ({
    browser,
  }) => {
    const context = await browser.newContext({ viewport: MOBILE });
    const page = await context.newPage();
    try {
      await page.goto('/medios');
      await expect(page.getByRole('heading', { level: 1, name: 'Medios' })).toBeVisible();

      const bloque = revealDeLista(page);
      await expect(bloque).toHaveCount(1);
      // SIN scroll: si el umbral del observador no alcanza (bug 2), el bloque
      // queda pendiente (`data-reveal="left"`, `opacity: 0`) y la página entera
      // se ve en blanco.
      await expect(
        bloque,
        'el bloque de la lista se revela al cargar, aunque mida más que el viewport',
      ).toHaveAttribute('data-reveal', 'in', { timeout: 3000 });
      await expect(bloque).toHaveCSS('opacity', '1', { timeout: 3000 });
    } finally {
      await context.close();
    }
  });
});

test.describe('Bug 3 · menú móvil con la página desplazada: scrim a viewport y cierre al tocar fuera', () => {
  test.setTimeout(300_000);

  let seed: PublicSeed;

  test.beforeAll(async ({ browser }) => {
    // La semilla de portada queda publicada al terminar (estado normal de
    // desarrollo, igual que en `navegacion-movil.spec.ts`).
    const context = await createAdminContext(browser);
    try {
      const page = await context.newPage();
      seed = await seedPublicPortada(page);
    } finally {
      await context.close();
    }
  });

  test('a 390×844 desplazado: el scrim cubre el viewport, el fondo del panel es opaco y tocar fuera cierra', async ({
    browser,
  }) => {
    const context = await browser.newContext({ viewport: MOBILE });
    const page = await context.newPage();
    try {
      await page.goto('/');
      await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

      // Desplazar la página: la cabecera gana `data-scrolled='true'` (y con él
      // `backdrop-filter`, el containing block que aclara el bug).
      await page.mouse.wheel(0, 1200);
      await expect(page.locator('header')).toHaveAttribute('data-scrolled', 'true');
      expect(
        await page.evaluate(() => window.scrollY),
        'precondición: la página está desplazada',
      ).toBeGreaterThan(800);

      const burger = page.getByRole('button', { name: 'Abrir menú de secciones' });
      await burger.click();
      const panel = page.locator('ul.public-nav-list');
      await expect(panel).toHaveAttribute('data-open', 'true');

      // ── El scrim debe cubrir el viewport, no solo la cabecera ────────────
      // Con el bug, el `backdrop-filter` de la cabecera reduce el
      // `fixed inset-0` del scrim a la caja de la cabecera (~390×68), así que el
      // resto de la pantalla queda sin scrim (tampoco atenúa).
      const scrim = page.getByTestId('mobile-menu-scrim');
      const caja = await scrim.boundingBox();
      expect(
        caja?.height ?? 0,
        'el scrim cubre el alto del viewport (844, no 68)',
      ).toBeGreaterThanOrEqual(840);
      expect(caja?.width ?? 0, 'el scrim cubre el ancho del viewport').toBeGreaterThanOrEqual(390);

      // ── El fondo del panel debe ser opaco (alpha 1) ─────────────────────
      const alpha = await panel.evaluate((el) => {
        const color = getComputedStyle(el).backgroundColor;
        const partes = /rgba?\(([^)]+)\)/.exec(color)?.[1].split(',');
        if (!partes) {
          return Number.NaN;
        }
        return partes.length === 4 ? Number(partes[3]) : 1;
      });
      expect(
        alpha,
        `el fondo del panel es opaco (backgroundColor alpha, hoy ${alpha})`,
      ).toBeGreaterThanOrEqual(0.99);

      // ── Tocar la zona inferior (fuera del panel) cierra el menú ──────────
      // Con el bug, en esa zona no está el scrim (solo la cabecera): el toque
      // va a la página y el menú sigue abierto.
      await page.mouse.click(MOBILE.width / 2, MOBILE.height - 30);
      await expect(burger).toHaveAttribute('aria-expanded', 'false');
    } finally {
      await context.close();
    }
  });
});
