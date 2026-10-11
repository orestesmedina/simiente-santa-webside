import { expect, test } from '@playwright/test';
import { createAdminContext, seedPublicPortada, type PublicSeed } from './publicSeed';

/**
 * E2E del tema claro/oscuro (F10 · T1031 · US5; SC-012). Cubre los tres caminos:
 * sistema → interruptor manual → persistencia; el arranque **sin parpadeo**
 * (anti-FOUC: el `data-theme` ya es el correcto al terminar de parsear el HTML);
 * el cambio en caliente del sistema solo sin elección manual; el logotipo con sus
 * colores originales en oscuro y las mismas funciones en ambos temas.
 *
 * Ejecutar con `--workers=1` (singletons compartidos, rate-limit por IP).
 */

const DARK_SURFACE = 'rgb(11, 20, 36)';
const LIGHT_SURFACE = 'rgb(26, 43, 74)';
const LOGO_PILL = 'rgb(245, 242, 236)';

test.describe('Tema claro/oscuro (F10 · T1031)', () => {
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

  test('sistema en oscuro sin elección: oscuro desde el primer pintado (sin flash)', async ({
    browser,
  }) => {
    const context = await browser.newContext({ colorScheme: 'dark' });
    try {
      const page = await context.newPage();
      await page.goto('/', { waitUntil: 'domcontentloaded' });

      // El script anti-FOUC ya dejó el tema correcto antes de que React monte.
      expect(await page.evaluate(() => document.documentElement.dataset.theme)).toBe('dark');
      expect(await page.evaluate(() => document.documentElement.style.colorScheme)).toBe('dark');
      expect(await page.evaluate(() => window.localStorage.getItem('ss.theme'))).toBeNull();

      await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();
      const surface = await page
        .locator('#root > div')
        .first()
        .evaluate((el) => getComputedStyle(el).backgroundColor);
      expect(surface, 'superficie oscura (navy profundo)').toBe(DARK_SURFACE);
    } finally {
      await context.close();
    }
  });

  test('elección manual: cambio inmediato sin recargar, persistente tras recargar y en visita posterior', async ({
    browser,
  }) => {
    const context = await browser.newContext({ colorScheme: 'light' });
    try {
      const page = await context.newPage();
      await page.goto('/');
      await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();
      await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
      expect(
        await page
          .locator('#root > div')
          .first()
          .evaluate((el) => getComputedStyle(el).backgroundColor),
      ).toBe(LIGHT_SURFACE);

      await page.evaluate(() => {
        (window as unknown as { __noReload?: boolean }).__noReload = true;
      });
      await page.getByRole('button', { name: 'Cambiar a tema oscuro' }).click();

      await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
      expect(
        await page.evaluate(() => (window as unknown as { __noReload?: boolean }).__noReload),
        'el cambio no recarga la página',
      ).toBe(true);
      expect(await page.evaluate(() => window.localStorage.getItem('ss.theme'))).toBe('dark');
      await expect(page.getByRole('button', { name: 'Cambiar a tema claro' })).toHaveAttribute(
        'aria-pressed',
        'true',
      );

      // Persiste al recargar.
      await page.reload();
      await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');

      // Y en una visita posterior (mismo almacenamiento, contexto nuevo).
      const storage = await context.storageState();
      const revisit = await browser.newContext({ storageState: storage, colorScheme: 'light' });
      try {
        const revisitPage = await revisit.newPage();
        await revisitPage.goto('/');
        await expect(revisitPage.locator('html')).toHaveAttribute('data-theme', 'dark');
      } finally {
        await revisit.close();
      }
    } finally {
      await context.close();
    }
  });

  test('el sistema cambia en caliente solo sin elección manual', async ({ browser }) => {
    const context = await browser.newContext({ colorScheme: 'dark' });
    try {
      const page = await context.newPage();
      await page.goto('/');
      await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');

      // Sin elección manual: sigue al sistema.
      await page.emulateMedia({ colorScheme: 'light' });
      await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');

      // Con elección manual, el sistema no la pisa.
      await page.getByRole('button', { name: 'Cambiar a tema oscuro' }).click();
      await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
      await page.emulateMedia({ colorScheme: 'light' });
      await page.waitForTimeout(200);
      await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
    } finally {
      await context.close();
    }
  });

  test('el logotipo conserva sus colores originales en ambos temas', async ({ browser }) => {
    for (const colorScheme of ['light', 'dark'] as const) {
      const context = await browser.newContext({ colorScheme });
      try {
        const page = await context.newPage();
        await page.goto('/');
        await expect(
          page.getByRole('heading', { level: 1, name: seed.identityName }),
        ).toBeVisible();

        const logo = page.getByRole('banner').getByRole('img', { name: seed.logoAlt });
        await expect(logo).toBeVisible();
        const styles = await logo.evaluate((el) => {
          const style = getComputedStyle(el);
          return { filter: style.filter, background: style.backgroundColor };
        });
        expect(styles.filter, `logo sin filtros en ${colorScheme}`).toBe('none');
        // El logotipo nunca se recolorea: se apoya en su pastilla crema.
        expect(styles.background, `pastilla del logo en ${colorScheme}`).toBe(LOGO_PILL);
      } finally {
        await context.close();
      }
    }
  });

  test('el control de tema es operable con teclado y las funciones son idénticas en ambos temas', async ({
    browser,
  }) => {
    const context = await browser.newContext({ colorScheme: 'light' });
    try {
      const page = await context.newPage();
      await page.setViewportSize({ width: 320, height: 900 });
      await page.goto('/');
      await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

      const toggle = page.locator('header button[aria-pressed]');
      await expect(toggle).toHaveAttribute('aria-pressed', 'false');
      await toggle.focus();
      await page.keyboard.press('Enter');

      await expect(toggle).toHaveAttribute('aria-pressed', 'true');
      await expect(toggle, 'el cambio de tema no pierde el foco').toBeFocused();

      // Mismas funciones: navegación, menú hamburguesa y mapa del sitio.
      const nav = page.getByRole('navigation', { name: 'Secciones de la portada' });
      await expect(nav).toBeVisible();
      await page.getByRole('button', { name: 'Abrir menú de secciones' }).click();
      await expect(nav.getByRole('link', { name: 'Contacto' })).toBeVisible();
      await page.keyboard.press('Escape');
      await expect(page.getByRole('navigation', { name: 'Mapa del sitio' })).toBeVisible();
      await expect(
        page
          .getByRole('navigation', { name: 'Mapa del sitio' })
          .getByRole('link', { name: 'Inicio' }),
      ).toBeVisible();
    } finally {
      await context.close();
    }
  });
});
