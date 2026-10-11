import { expect, test, type Locator } from '@playwright/test';
import { ADMIN, ensureInitialized, loginViaUi } from './helpers';
import { seedPublicPortada, type PublicSeed } from './publicSeed';

/**
 * E2E de diseño, navegación fija y selector compacto (F10 · T1029 · US1/US2/US4;
 * SC-004/SC-007). Escritorio: barra fija con todas las anclas a la vista (sin
 * menú oculto), ancla que no tapa el título destino, 0 desplazamiento horizontal
 * de 320 a 1280 px, recorrido por teclado con foco visible, selector compacto con
 * nombres completos y persistencia, y cabecera completa a 320 px.
 *
 * Ejecutar con `--workers=1` (singletons compartidos, rate-limit por IP).
 */

interface Box {
  x: number;
  y: number;
  width: number;
  height: number;
}

function intersects(a: Box, b: Box): boolean {
  return a.x < b.x + b.width && b.x < a.x + a.width && a.y < b.y + b.height && b.y < a.y + a.height;
}

async function box(locator: Locator): Promise<Box> {
  const result = await locator.boundingBox();
  expect(result, 'el control tiene caja').not.toBeNull();
  return result as Box;
}

test.describe('Diseño, navegación fija y selector (F10 · T1029)', () => {
  test.setTimeout(180_000);

  let seed: PublicSeed;

  test.beforeAll(async ({ browser }) => {
    const context = await browser.newContext();
    try {
      await ensureInitialized(context.request);
      const page = await context.newPage();
      await loginViaUi(page, ADMIN.email, ADMIN.password);
      await expect(page).toHaveURL(/\/panel$/);
      seed = await seedPublicPortada(page);
    } finally {
      await context.close();
    }
  });

  test('escritorio: barra fija con anclas a la vista y el ancla no tapa el título', async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    const nav = page.getByRole('navigation', { name: 'Secciones de la portada' });
    for (const name of [
      'Quiénes somos',
      'Horario de servicios',
      'WhatsApp',
      'Contacto',
      'Redes sociales',
    ]) {
      await expect(nav.getByRole('link', { name }), `${name} a la vista`).toBeVisible();
    }
    // En escritorio no hay menú hamburguesa.
    await expect(page.getByRole('button', { name: 'Abrir menú de secciones' })).toBeHidden();

    // La barra permanece visible al bajar (fija).
    await page.mouse.wheel(0, 1500);
    await expect(nav.getByRole('link', { name: 'Contacto' })).toBeVisible();
    await expect(page.locator('header')).toBeVisible();

    // Clic en un ancla → el título destino queda visible y la barra no lo tapa.
    await nav.getByRole('link', { name: 'Horario de servicios' }).click();
    const heading = page.getByRole('heading', { level: 2, name: 'Horario de servicios' });
    await expect(heading).toBeVisible();
    const headerHeight = (await box(page.locator('header'))).height;
    const headingBox = await box(heading);
    expect(
      headingBox.y,
      'el título destino no queda tapado por la barra fija',
    ).toBeGreaterThanOrEqual(headerHeight - 4);
  });

  test('sin desplazamiento horizontal a 320/360/768/1024/1280 px', async ({ page }) => {
    for (const width of [320, 360, 768, 1024, 1280] as const) {
      await page.setViewportSize({ width, height: 900 });
      await page.goto('/');
      await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();
      const overflow = await page.evaluate(
        () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
      );
      expect(overflow, `sin scroll horizontal a ${width}px`).toBeLessThanOrEqual(1);
    }
  });

  test('teclado con foco visible y selector de idioma compacto que persiste', async ({
    browser,
  }) => {
    const context = await browser.newContext();
    try {
      const page = await context.newPage();
      await page.setViewportSize({ width: 1280, height: 900 });
      await page.goto('/');
      await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

      // El primer Tab enfoca el enlace de salto, con indicador de foco visible.
      await page.keyboard.press('Tab');
      const skip = page.getByRole('link', { name: 'Saltar al contenido principal' });
      await expect(skip).toBeFocused();
      const indicator = await skip.evaluate((el) => {
        const style = getComputedStyle(el);
        return (
          (style.outlineStyle !== 'none' && parseFloat(style.outlineWidth) > 0) ||
          style.boxShadow !== 'none'
        );
      });
      expect(indicator, 'foco visible en el enlace de salto').toBe(true);

      // Selector compacto: nombres completos, sin siglas ni menú.
      const header = page.locator('header');
      await expect(header.getByRole('radio', { name: 'Español' })).toBeChecked();
      await expect(header.getByRole('radio', { name: 'Inglés' })).toBeVisible();
      expect(await header.getByText('ES', { exact: true }).count()).toBe(0);

      // Cambio inmediato + persistencia en localStorage y tras recargar.
      await header.getByRole('radio', { name: 'Inglés' }).check();
      await expect(page.locator('html')).toHaveAttribute('lang', 'en');
      await expect(
        page.getByRole('navigation', { name: 'Page sections' }).getByRole('link', {
          name: 'Who we are',
        }),
      ).toBeVisible();
      expect(await page.evaluate(() => window.localStorage.getItem('ss.lang'))).toBe('en');

      await page.reload();
      await expect(page.locator('html')).toHaveAttribute('lang', 'en');
      await expect(header.getByRole('radio', { name: 'English' })).toBeChecked();

      // Re-visita con la preferencia guardada (contexto nuevo con el storage).
      const storage = await context.storageState();
      const revisit = await browser.newContext({ storageState: storage });
      try {
        const revisitPage = await revisit.newPage();
        await revisitPage.goto('/');
        await expect(revisitPage.locator('html')).toHaveAttribute('lang', 'en');
      } finally {
        await revisit.close();
      }
    } finally {
      await context.close();
    }
  });

  test('a 320 px marca, idioma, tema y hamburguesa conviven sin taparse', async ({ page }) => {
    await page.setViewportSize({ width: 320, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    const header = page.locator('header');
    const brand = header.getByRole('link', { name: /Simiente/ });
    const lang = header.getByRole('group', { name: 'Idioma' });
    const theme = header.getByRole('button', { name: /Cambiar a tema/ });
    const burger = header.getByRole('button', { name: 'Abrir menú de secciones' });

    for (const control of [brand, lang, theme, burger]) {
      await expect(control).toBeVisible();
    }
    // El enlace de marca conserva «Simiente» en su texto (D20/C3).
    await expect(brand).toContainText('Simiente');

    const boxes = await Promise.all([brand, lang, theme, burger].map((c) => box(c)));
    for (let i = 0; i < boxes.length; i += 1) {
      for (let j = i + 1; j < boxes.length; j += 1) {
        expect(
          intersects(boxes[i], boxes[j]),
          `los controles ${i} y ${j} no se solapan a 320 px`,
        ).toBe(false);
      }
    }
  });
});
