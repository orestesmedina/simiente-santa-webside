import { expect, test, type BrowserContext } from '@playwright/test';
import { createAdminContext, seedPublicPortada } from './publicSeed';
import { seedAgenda } from './agendaSeed';

/**
 * QA · Bugs reportados de 004-eventos-actividades (⟲ QA, no modifica código de
 * producción):
 *
 * - **Bug 1**: desde `/eventos`, los enlaces de secciones del menú (`<a>` nativo
 *   con `href="#whatsapp"`) resuelven contra la ruta actual (`/eventos#…`) y no
 *   llevan a la portada. Se espera que hagan click → `/` con `#whatsapp` y que
 *   la sección quede a la vista.
 * - **Bug 2**: el menú no indica la opción actual ni tiene hover.
 * - **Bug 3**: el modal de «Agregar evento» es estrecho (`max-w-lg` ≤ 512 px) y
 *   su contenido (`DatesEditor`, grid de 3 columnas + inputs nativos) desborda
 *   en horizontal (`overflow-y-auto` computa `overflow-x: auto`). Se exige
 *   ancho ≥ 640 px y `scrollWidth ≤ clientWidth` en el modal.
 *
 * Ejecutar con un worker (singlenton de sesión y rate-limit de login):
 *
 *   docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
 *     -v "$PWD/..":/work -w /work/frontend mcr.microsoft.com/playwright:v1.63.0-jammy \
 *     npx playwright test --config e2e/playwright.config.ts qa-bugs --workers=1
 */

test.describe.configure({ mode: 'serial' });

test.describe('QA de los 3 bugs reportados (004-eventos-actividades)', () => {
  test.setTimeout(180_000);

  test.beforeAll(async ({ browser }) => {
    const context = await createAdminContext(browser);
    try {
      const page = await context.newPage();
      // Evento publicado y próximo (así «Eventos y actividades» existe en la
      // barra) y portada con WhatsApp publicado (así el enlace es navegable).
      await seedAgenda(page);
      await seedPublicPortada(page);
    } finally {
      await context.close();
    }
  });

  async function visitante(browser: Parameters<typeof createAdminContext>[0]): Promise<{
    context: BrowserContext;
    close: () => Promise<void>;
  }> {
    const context = await browser.newContext();
    return {
      context,
      close: async () => {
        await context.close();
      },
    };
  }

  test('Bug 1: desde /eventos, «WhatsApp» lleva a la portada con su sección', async ({
    browser,
  }) => {
    const { context, close } = await visitante(browser);
    try {
      const page = await context.newPage();
      await page.goto('/eventos');

      const nav = page.getByRole('navigation', { name: 'Secciones de la portada' });
      const whatsapp = nav.getByRole('link', { name: 'WhatsApp' });
      await expect(whatsapp).toBeVisible();

      await whatsapp.click();

      // Con el href relativo el navegador simplemente añade el hash a /eventos:
      // eso es el bug. La corrección debe llevar al usuario a la portada.
      await expect(page).toHaveURL(/\/#whatsapp$/, { timeout: 5_000 });

      // La sección de WhatsApp existe en la portada y queda a la vista.
      const seccion = page.locator('#whatsapp');
      await expect(seccion).toBeVisible();
      await expect(seccion).toBeInViewport();
    } finally {
      await close();
    }
  });

  test('Bug 2: en /eventos la opción actual está marcada y los enlaces tienen hover', async ({
    browser,
  }) => {
    const { context, close } = await visitante(browser);
    try {
      const page = await context.newPage();
      await page.setViewportSize({ width: 1280, height: 900 });
      await page.goto('/eventos');

      const nav = page.getByRole('navigation', { name: 'Secciones de la portada' });
      const eventos = nav.getByRole('link', { name: 'Eventos y actividades' });
      await expect(eventos).toBeVisible();

      // Estado activo: aria-current="page" en la ruta actual.
      await expect(eventos).toHaveAttribute('aria-current', 'page', { timeout: 5_000 });

      // Hover: repintado real (background-color cambia al pasar el cursor).
      const antes = await eventos.evaluate((el) => getComputedStyle(el).backgroundColor);
      await eventos.hover();
      await page.waitForTimeout(350); // deja terminar cualquier transición
      const despues = await eventos.evaluate((el) => getComputedStyle(el).backgroundColor);
      expect(
        despues,
        `el hover debe cambiar el fondo (antes «${antes}», después «${despues}»)`,
      ).not.toBe(antes);
    } finally {
      await close();
    }
  });

  test('Bug 3: el modal de «Agregar evento» es ancho y sin scroll horizontal', async ({
    browser,
  }) => {
    const context = await createAdminContext(browser);
    try {
      const page = await context.newPage();
      await page.setViewportSize({ width: 1280, height: 900 });
      await page.goto('/panel/eventos');
      await expect(page.getByRole('heading', { name: 'Gestión de eventos' })).toBeVisible();

      await page.getByRole('button', { name: 'Agregar evento' }).click();
      const dialog = page.getByRole('dialog');
      await expect(dialog).toBeVisible();

      // El editor de fechas con varias filas es donde desborda el contenido.
      await dialog.getByRole('button', { name: 'Agregar fecha' }).click();
      await dialog.getByRole('button', { name: 'Agregar fecha' }).click();

      const medidas = await dialog.evaluate((el) => {
        const rect = el.getBoundingClientRect();
        return {
          width: rect.width,
          scrollWidth: el.scrollWidth,
          clientWidth: el.clientWidth,
          overflowX: getComputedStyle(el).overflowX,
        };
      });

      expect(
        medidas.width,
        `el modal debe tener ancho ≥ 640 px; midió ${medidas.width} (max-w-lg lo limita a 512)`,
      ).toBeGreaterThanOrEqual(640);

      expect(
        medidas.scrollWidth,
        `sin scroll horizontal: scrollWidth ${medidas.scrollWidth} > clientWidth ${medidas.clientWidth}`,
      ).toBeLessThanOrEqual(medidas.clientWidth);
    } finally {
      await context.close();
    }
  });
});
