import { expect, test, type Locator, type Page } from '@playwright/test';
import { ADMIN, ensureInitialized, loginViaUi } from './helpers';
import { seedPublicPortada, type PublicSeed } from './publicSeed';

/**
 * E2E del sistema de movimiento en modo normal (F10 · T1032 · US3; SC-014).
 * Verifica que el revelado al scroll es **perceptible de verdad** (suelo: ≥ 400
 * ms y ≥ 16 px; techo: ≤ 800 ms), que el escalonado existe y está acotado, que la
 * cabecera transiciona su superficie al bajar sin cambiar de altura, que el
 * ancla desliza con el título visible, que la animación **no bloquea** y que
 * cambiar de idioma o tema **no re-anima** lo ya visto.
 *
 * Ejecutar con `--workers=1` (singletons compartidos, rate-limit por IP).
 */

const SECTIONS_NAV = 'Secciones de la portada';

function pendingBlocks(page: Page): Locator {
  return page.locator(
    '.reveal-block[data-reveal="left"], .reveal-block[data-reveal="right"], .reveal-block[data-reveal="up"]',
  );
}

test.describe('Movimiento perceptible y acotado (F10 · T1032)', () => {
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

  test('el revelado supera el suelo de perceptibilidad y respeta el techo', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    const block = pendingBlocks(page).last();
    await expect(block, 'hay una sección fuera de pantalla pendiente').toBeVisible();

    const pending = await block.evaluate((el) => {
      const style = getComputedStyle(el);
      return {
        opacity: parseFloat(style.opacity),
        transform: style.transform,
        duration: style.transitionDuration,
        delay: style.transitionDelay,
      };
    });

    // Estado pendiente: opacidad < 1 y desplazamiento ≥ 16 px.
    expect(pending.opacity).toBeLessThan(1);
    const matrix = /matrix\(([^)]+)\)/.exec(pending.transform);
    expect(matrix, 'el bloque pendiente está desplazado').not.toBeNull();
    const distance = Math.max(
      Math.abs(Number(matrix?.[1].split(',')[4])),
      Math.abs(Number(matrix?.[1].split(',')[5])),
    );
    expect(distance, 'distancia de entrada ≥ 16 px').toBeGreaterThanOrEqual(16);

    // Duración en el rango del sistema (≥ 400 ms y ≤ 800 ms — el mismo techo del e2e).
    const durationMs = parseFloat(pending.duration) * 1000;
    expect(durationMs, 'duración ≥ 400 ms').toBeGreaterThanOrEqual(400);
    expect(durationMs, 'duración ≤ 800 ms').toBeLessThanOrEqual(800);

    // Al entrar en pantalla termina en opacidad 1 y sin transformación.
    await block.scrollIntoViewIfNeeded();
    await expect(block).toHaveAttribute('data-reveal', 'in');
    await expect.poll(() => block.evaluate((el) => getComputedStyle(el).opacity)).toBe('1');
    await expect.poll(() => block.evaluate((el) => getComputedStyle(el).transform)).toBe('none');
  });

  test('el escalonado existe y está acotado a ≤ 240 ms', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    const delays = await page
      .locator('.reveal-block')
      .evaluateAll((els) => els.map((el) => getComputedStyle(el).transitionDelay));

    expect(delays.length).toBeGreaterThanOrEqual(3);
    const seconds = delays.map((value) => parseFloat(value));
    expect(Math.max(...seconds), 'tope de escalonado ≤ 240 ms').toBeLessThanOrEqual(0.24);
    expect(new Set(delays).size, 'hay más de un retardo (escalonado)').toBeGreaterThan(1);
  });

  test('la cabecera transiciona su superficie al bajar sin cambiar de altura', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    const header = page.locator('header');
    await expect(header).toHaveAttribute('data-scrolled', 'false');
    const before = await header.boundingBox();

    await page.mouse.wheel(0, 900);

    await expect(header).toHaveAttribute('data-scrolled', 'true');
    const after = await header.boundingBox();
    expect(
      Math.abs((after?.height ?? 0) - (before?.height ?? 0)),
      'sin cambio de geometría (misma altura)',
    ).toBeLessThanOrEqual(1);

    // Al volver arriba, regresa.
    await page.evaluate(() => window.scrollTo(0, 0));
    await expect(header).toHaveAttribute('data-scrolled', 'false');
  });

  test('el ancla desliza con el título visible y la animación no bloquea', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    const nav = page.getByRole('navigation', { name: SECTIONS_NAV });

    // Clic inmediato (mientras el resto se anima): responde al instante.
    await nav.getByRole('link', { name: 'Redes sociales' }).click();
    const heading = page.getByRole('heading', { level: 2, name: 'Redes sociales' });
    await expect(heading).toBeVisible();
    const headerHeight = (await page.locator('header').boundingBox())?.height ?? 0;
    expect((await heading.boundingBox())?.y ?? 0).toBeGreaterThanOrEqual(headerHeight - 4);

    // El foco sobre un bloque pendiente lo revela al instante (nunca oculta el foco).
    await page.evaluate(() => window.scrollTo(0, 0));
    const pending = pendingBlocks(page).last();
    await expect(pending).toBeVisible();
    const chip = pending.getByRole('link', { name: /Ver en Instagram/ });
    await expect(chip).toBeEnabled();
    await chip.focus();
    await expect(pending).toHaveAttribute('data-reveal', 'in');
  });

  test('cambiar de idioma o de tema no re-anima lo ya visto', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    const blocks = page.locator('.reveal-block');
    await expect(blocks.last()).toHaveAttribute('data-reveal', 'in');

    const revealed = await blocks.evaluateAll((els) =>
      els.map((el) => el.getAttribute('data-reveal')),
    );
    expect(revealed.every((value) => value === 'in')).toBe(true);

    await page.locator('header').getByRole('radio', { name: 'Inglés' }).check();
    await expect(page.getByRole('navigation', { name: 'Page sections' })).toBeVisible();
    expect(
      await page
        .locator('.reveal-block')
        .evaluateAll((els) => els.map((el) => el.getAttribute('data-reveal'))),
      'el idioma no re-anima las secciones ya reveladas',
    ).toEqual(revealed);

    await page.getByRole('button', { name: 'Cambiar a tema oscuro' }).click();
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
    expect(
      await page
        .locator('.reveal-block')
        .evaluateAll((els) => els.map((el) => el.getAttribute('data-reveal'))),
      'el tema no re-anima las secciones ya reveladas',
    ).toEqual(revealed);
  });
});
