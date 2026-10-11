import { expect, test, type Locator, type Page } from '@playwright/test';
import { createAdminContext, seedPublicPortada, type PublicSeed } from './publicSeed';

/**
 * E2E del movimiento reducido (F10 · T1033 · US3; SC-005). Con
 * `emulateMedia({ reducedMotion: 'reduce' })`: **0** revelados ni transiciones de
 * entrada —todo visible al instante, sin delays—, desplazamiento de anclas
 * instantáneo, cabecera sin transición, blobs sin animar y el menú móvil sin
 * animación; y el **100 % del contenido y las funciones** siguen disponibles.
 * Se contrasta con el modo normal (T1032), donde sí hay revelado.
 *
 * Ejecutar con `--workers=1` (singletons compartidos, rate-limit por IP).
 */

const SECTIONS_NAV = 'Secciones de la portada';

function allZero(value: string): boolean {
  return value
    .split(',')
    .map((part) => parseFloat(part))
    .every((number) => number === 0);
}

function pendingBlocks(page: Page): Locator {
  return page.locator(
    '.reveal-block[data-reveal="left"], .reveal-block[data-reveal="right"], .reveal-block[data-reveal="up"]',
  );
}

test.describe('Movimiento reducido (F10 · T1033)', () => {
  test.setTimeout(180_000);

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

  test('0 revelados: contenido completo al instante, sin delays ni transiciones', async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    // Ningún bloque queda pendiente (en el modo normal sí habría, T1032).
    await expect(pendingBlocks(page)).toHaveCount(0);

    const blocks = page.locator('.reveal-block');
    const count = await blocks.count();
    expect(count).toBeGreaterThan(0);
    for (let index = 0; index < count; index += 1) {
      const styles = await blocks.nth(index).evaluate((el) => {
        const style = getComputedStyle(el);
        return {
          opacity: style.opacity,
          transform: style.transform,
          delay: style.transitionDelay,
          duration: style.transitionDuration,
          reveal: el.getAttribute('data-reveal'),
        };
      });
      expect(styles.reveal).toBe('in');
      expect(styles.opacity).toBe('1');
      expect(styles.transform).toBe('none');
      expect(allZero(styles.delay), 'sin delays').toBe(true);
      expect(allZero(styles.duration), 'sin transiciones de entrada').toBe(true);
    }

    // Salto de anclas instantáneo y cabecera sin transición.
    expect(
      await page.evaluate(() => getComputedStyle(document.documentElement).scrollBehavior),
    ).toBe('auto');
    const headerDuration = await page
      .locator('header')
      .evaluate((el) => getComputedStyle(el).transitionDuration);
    expect(allZero(headerDuration), 'cabecera sin transición').toBe(true);

    // Blobs sin animar.
    const blob = page.locator('.public-blob').first();
    if ((await blob.count()) > 0) {
      expect(await blob.evaluate((el) => getComputedStyle(el).animationName)).toBe('none');
    }
  });

  test('el 100 % del contenido y las funciones siguen disponibles (menú sin animación)', async ({
    page,
  }) => {
    await page.setViewportSize({ width: 320, height: 900 });
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    // Secciones y navegación intactas.
    const nav = page.getByRole('navigation', { name: SECTIONS_NAV });
    await expect(nav).toBeVisible();
    await expect(page.locator('#horario')).toBeVisible();
    await expect(page.locator('#contacto')).toBeVisible();

    // Menú hamburguesa: abre y cierra al instante (sin opacidad/transform pendiente).
    const burger = page.getByRole('button', { name: 'Abrir menú de secciones' });
    await burger.click();
    const list = nav.locator('ul.public-nav-list');
    await expect(list).toBeVisible();
    const panel = await list.evaluate((el) => {
      const style = getComputedStyle(el);
      return { opacity: style.opacity, transform: style.transform };
    });
    expect(panel.opacity, 'el panel aparece al instante').toBe('1');
    await page.keyboard.press('Escape');
    await expect(burger).toHaveAttribute('aria-expanded', 'false');

    // Mapa del sitio y control de tema operables.
    const sitemap = page.getByRole('navigation', { name: 'Mapa del sitio' });
    await expect(sitemap.getByRole('link', { name: 'Inicio' })).toBeVisible();
    await page.getByRole('button', { name: 'Cambiar a tema oscuro' }).click();
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
  });

  test('activar «reducir movimiento» con el sitio abierto surte efecto sin recargar', async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.emulateMedia({ reducedMotion: 'no-preference' });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    // En modo normal hay contenido pendiente (contraste con T1032).
    const block = page.locator('#redes').locator('..');
    await expect(block, 'la sección de abajo arranca pendiente').toHaveAttribute(
      'data-reveal',
      /left|right|up/,
    );
    const before = await block.evaluate((el) => getComputedStyle(el).opacity);
    expect(parseFloat(before)).toBeLessThan(1);

    // Sin recargar: el contenido aparece completo de inmediato.
    await page.evaluate(() => {
      (window as unknown as { __noReload?: boolean }).__noReload = true;
    });
    await page.emulateMedia({ reducedMotion: 'reduce' });

    await expect(block).toHaveAttribute('data-reveal', 'in');
    await expect.poll(() => block.evaluate((el) => getComputedStyle(el).opacity)).toBe('1');
    expect(
      await page.evaluate(() => (window as unknown as { __noReload?: boolean }).__noReload),
      'el cambio no recarga la página',
    ).toBe(true);
    await expect(pendingBlocks(page)).toHaveCount(0);
  });
});
