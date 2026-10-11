import { expect, test } from '@playwright/test';
import { createAdminContext, seedPublicPortada, type PublicSeed } from './publicSeed';

/**
 * E2E de QA para los hallazgos de medida abiertos de F10:
 * - **M4** (de `analyze.md`, se cierra en QA): el objetivo táctil y la
 *   legibilidad del selector compacto se comprueban sobre **estilos
 *   computados reales** (no sobre clases) con el umbral acordado de
 *   FR-004/FR-022 — texto ≥ 14 px objetivo táctil ≥ 44 px — y su huella a 320 px.
 * - **FR-014** (botón del menú siempre alcanzable): el hamburguesa sigue
 *   visible y operable en el **final** de la página en móvil, y en escritorio
 *   los enlaces de la barra siguen a la vista en el mismo punto.
 *
 * Ejecutar con `--workers=1` (singletons compartidos, rate-limit por IP).
 */

test.describe('QA de medidas del sistema y navegación (F10)', () => {
  test.setTimeout(120_000);

  let seed: PublicSeed;

  test.beforeAll(async ({ browser }) => {
    const context = await createAdminContext(browser);
    try {
      const page = await context.newPage();
      seed = await seedPublicPortada(page);
    } finally {
      await context.close();
    }
  });

  test('texto ≥ 14 px, objetivo táctil ≥ 44 px y huella compacta', async ({ page }) => {
    await page.setViewportSize({ width: 320, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    const group = page.locator('header').getByRole('group', { name: 'Idioma' });
    await expect(group).toBeVisible();

    const labels = group.locator('label');
    await expect(labels).toHaveCount(2);

    for (let index = 0; index < 2; index += 1) {
      const label = labels.nth(index);
      const styles = await label.evaluate((el) => {
        const style = getComputedStyle(el);
        return {
          fontSize: parseFloat(style.fontSize),
          lineHeight: parseFloat(style.lineHeight),
        };
      });
      const box = await label.boundingBox();
      expect(box, 'la opción tiene caja').not.toBeNull();

      // Legibilidad (umbral M4): texto del selector ≥ 14 px, altura de línea ≥ 20 px.
      expect(styles.fontSize, `texto de la opción ${index} ≥ 14 px`).toBeGreaterThanOrEqual(14);
      expect(
        styles.lineHeight,
        `altura de línea de la opción ${index} ≥ 20 px`,
      ).toBeGreaterThanOrEqual(20);

      // Objetivo táctil ≥ 44 px (FR-022, heredado de F3).
      expect(box!.height, `altura táctil de la opción ${index} ≥ 44 px`).toBeGreaterThanOrEqual(44);
      expect(box!.width, `anchura táctil de la opción ${index} ≥ 44 px`).toBeGreaterThanOrEqual(44);
    }

    const groupBox = await group.boundingBox();
    expect(groupBox, 'el grupo tiene caja').not.toBeNull();
    // Huella compacta: la pastilla entera mide menos del doble del objetivo, sin huecos.
    expect(groupBox!.height, 'la pastilla del selector compacta ≤ 56 px').toBeLessThanOrEqual(56);
  });

  test('el botón del menú sigue alcanzable al final de la página (móvil) y los enlaces de la barra a la vista (escritorio)', async ({
    page,
  }) => {
    // Móvil: cabecera fija con el hamburguesa visibles y operable al fondo.
    await page.setViewportSize({ width: 320, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    const header = page.locator('header');
    const headerStyle = await header.evaluate((el) => getComputedStyle(el).position);
    expect(headerStyle, 'la cabecera es fija en móvil').toMatch(/fixed|sticky/);

    await page.mouse.wheel(0, 60000);
    await page.waitForFunction(
      () =>
        Math.abs(window.scrollY + window.innerHeight - document.documentElement.scrollHeight) < 5,
    );

    const burger = page.getByRole('button', { name: 'Abrir menú de secciones' });
    await expect(burger, 'hamburguesa visible al fondo de la página').toBeVisible();
    const burgerBox = await burger.boundingBox();
    expect(burgerBox!.y, 'hamburguesa dentro de la ventana').toBeGreaterThanOrEqual(0);
    await burger.click();
    await expect(page.getByRole('button', { name: 'Cerrar menú de secciones' })).toHaveAttribute(
      'aria-expanded',
      'true',
    );
    await page.keyboard.press('Escape');
    // Escritorio: la barra con enlaces sigue a la vista al fondo de la página.
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.goto('/');
    await page.evaluate(() =>
      window.scrollTo({ top: Number.MAX_SAFE_INTEGER, behavior: 'instant' }),
    );
    await page.waitForTimeout(100);
    const nav = page.getByRole('navigation', { name: 'Secciones de la portada' });
    await expect(nav.getByRole('link', { name: 'Contacto' })).toBeVisible();
  });
});
