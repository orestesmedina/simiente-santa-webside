import { expect, test, type Browser, type BrowserContext, type Page } from '@playwright/test';
import { createAdminContext } from './publicSeed';
import { resetMedios, seedMedios, type MediosSeed } from './mediosSeed';

/**
 * E2E de accesibilidad por teclado de medios (T942, SC-008/FR-028…FR-030,
 * quickstart §13): el patrón de `ofrendas-accesibilidad.spec.ts` /
 * `ministerios-accesibilidad.spec.ts`, aplicado a la sección `/medios` y al
 * bloque de la portada.
 *
 * Cubre:
 *   · el primer `Tab` enfoca el enlace de salto al contenido;
 *   · la travesía con `Tab` alcanza navegación, fichas, **enlaces externos**
 *     (activables con `Enter`), «Cargar más», «Ver todo», selector de idioma y
 *     pie **con foco visible en cada paso**;
 *   · **sin trampas de foco**: el recorrido cierra volviendo al primer elemento y
 *     `Shift+Tab` sale de la lista de navegación;
 *   · cada **enlace externo** anuncia **servicio + salida del sitio + pestaña
 *     nueva** (texto visible y `aria-label`), en español y en inglés;
 *   · el selector de idioma se opera con las flechas del grupo de radios **sin
 *     perder el foco**;
 *   · todo se comprueba en **tema claro y oscuro** (F10), a 320 px (**0
 *     desplazamiento horizontal**), con el texto al **200 %** y con
 *     `prefers-reduced-motion` (el contenido queda íntegro al instante);
 *   · con **0 publicables**, la redirección de `/medios` a la portada no atrapa
 *     el foco: el visitante termina en una página navegable (P9-16).
 *
 * Queda MANUAL (no automatizable, se consigna en la revisión de QA): el lector de
 * pantalla real. Cómo se ejecutan (la base, el rate-limit de login y el navegador
 * son compartidos por toda la suite):
 *
 *   # Tanda 6 — medios (F9): las cuatro specs comparten la siembra/limpieza del
 *   # módulo (y la carátula subida), así que van juntas y con un solo worker.
 *   cd frontend && npx playwright test --config e2e/playwright.config.ts medios- portada-medios --workers=1
 *
 * (El reparto completo de la suite y el motivo de las tandas están en
 * `medios-publica.spec.ts`.)
 */
test.describe.configure({ mode: 'serial' });

/** Nombre accesible de la lista de fichas de la sección (i18n `es`). */
const LIST = 'Lista de episodios de medios';
/** Nombre accesible de la lista del bloque de la portada (i18n `es`). */
const BLOCK_LIST = 'Medios: mira los episodios recientes y entra a la sección completa';
const CTA = 'Ver todo';
/** Tope de pasos del recorrido por teclado (la sección añade fichas al ciclo). */
const MAX_TAB_STEPS = 200;

interface TabStep {
  label: string;
  indicator: boolean;
}

interface StepInfo {
  label: string | null;
  indicator: boolean;
}

/**
 * Recorre con `Tab` hasta cerrar el ciclo (vuelve al primer elemento). En cada
 * paso exige **foco visible** (contorno o sombra). Copia del helper de F3/F6/F7.
 */
async function tabCycle(page: Page, maxSteps = MAX_TAB_STEPS): Promise<TabStep[]> {
  const path: TabStep[] = [];
  for (let i = 0; i < maxSteps; i++) {
    await page.keyboard.press('Tab');
    const info = await page.evaluate<StepInfo>(() => {
      const el = document.activeElement;
      if (!el || el === document.body) return { label: null, indicator: false };
      const style = getComputedStyle(el);
      return {
        label: `${el.tagName.toLowerCase()} ${(
          (el as HTMLElement).innerText ||
          (el as HTMLInputElement).value ||
          ''
        )
          .trim()
          .slice(0, 40)}`,
        indicator:
          (style.outlineStyle !== 'none' && parseFloat(style.outlineWidth) > 0) ||
          style.boxShadow !== 'none',
      };
    });
    if (info.label === null) {
      if (path.length === 0 && i < 3) {
        continue; // la SPA aún no montó: reintenta el salto
      }
      break;
    }
    const label = info.label;
    expect(info.indicator, `foco visible en ${label} (paso ${i + 1} del recorrido)`).toBe(true);
    if (path.length > 0 && path[0].label === label) {
      path.push({ label, indicator: info.indicator });
      break; // el ciclo cerró: volvió al primer elemento
    }
    path.push({ label, indicator: info.indicator });
  }
  return path;
}

/**
 * Contraste WCAG de todo el texto visible de `main` y del pie (fórmula de las
 * luminancias relativas). Copia adaptada del helper de F3/F6/F7.
 */
async function contrastFailures(page: Page): Promise<string[]> {
  return page.evaluate(() => {
    const channels = (color: string): [number, number, number, number] => {
      const m = color.match(/rgba?\(\s*([\d.]+)[,\s]+([\d.]+)[,\s]+([\d.]+)(?:[,\s/]+([\d.]+))?\)/);
      return m
        ? [Number(m[1]), Number(m[2]), Number(m[3]), m[4] === undefined ? 1 : Number(m[4])]
        : [0, 0, 0, 1];
    };
    const lin = (chunk: number): number => {
      const v = chunk / 255;
      return v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4;
    };
    const luminance = (rgb: number[]): number => {
      const [r, g, b] = rgb.map(lin);
      return 0.2126 * r + 0.7152 * g + 0.0722 * b;
    };
    const opaqueBg = (el: Element): [number, number, number] => {
      let node: Element | null = el;
      while (node && node !== document.documentElement) {
        const [r, g, b, alpha] = channels(getComputedStyle(node).backgroundColor);
        if (alpha >= 1) return [r, g, b];
        node = node.parentElement;
      }
      return [255, 255, 255];
    };
    const failures: string[] = [];
    const selector = 'main :is(h1,h2,h3,p,li,a,button,span), footer :is(a,p,span)';
    for (const el of [...document.querySelectorAll(selector)]) {
      if (!el.textContent || !el.textContent.trim()) continue;
      const box = el.getBoundingClientRect();
      if (box.width === 0 || box.height === 0) continue;
      const style = getComputedStyle(el);
      const fg = channels(style.color);
      if (fg[3] < 1) continue;
      const bg = opaqueBg(el);
      const ratio =
        (Math.max(luminance(fg), luminance(bg)) + 0.05) /
        (Math.min(luminance(fg), luminance(bg)) + 0.05);
      const fontSize = parseFloat(style.fontSize);
      const bold = parseInt(style.fontWeight, 10) >= 700;
      const large = fontSize >= 24 || (fontSize >= 18.66 && bold);
      if (ratio < (large ? 3 : 4.5)) {
        failures.push(
          `${el.tagName.toLowerCase()}[${(el as HTMLElement).innerText.slice(
            0,
            30,
          )}] ${ratio.toFixed(2)}:1 (${style.fontSize})`,
        );
      }
    }
    return failures;
  });
}

/** Contexto de visitante en el tema indicado, con los servicios externos servidos en local. */
async function visitor(
  browser: Browser,
  theme: 'light' | 'dark',
): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext({ colorScheme: theme });
  // Los enlaces externos no se visitan de verdad: se responde con una página
  // mínima para comprobar **la pestaña nueva** sin salir a internet.
  await context.route(/youtube\.com|spotify\.com/, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'text/html',
      body: '<html><body>externo</body></html>',
    }),
  );
  return { context, page: await context.newPage() };
}

test.describe('Accesibilidad por teclado de medios (SC-008)', () => {
  test.setTimeout(300_000);

  let adminContext: BrowserContext;
  let adminPage: Page;
  let seed: MediosSeed;

  test.beforeAll(async ({ browser }) => {
    adminContext = await createAdminContext(browser);
    adminPage = await adminContext.newPage();
    // 14 publicables: la sección entra en «Cargar más» (tanda de 12) y el
    // recorrido por teclado lo incluye.
    seed = await seedMedios(adminPage, { publishedCount: 14 });
  });

  test.afterAll(async () => {
    resetMedios();
    await adminContext?.close();
  });

  test('la sección /medios se recorre con teclado, con foco visible, en claro y oscuro', async ({
    browser,
  }) => {
    for (const theme of ['light', 'dark'] as const) {
      const { context, page } = await visitor(browser, theme);
      try {
        await page.setViewportSize({ width: 1280, height: 900 });
        await page.goto('/medios');
        await expect(page.getByRole('heading', { level: 1, name: 'Medios' })).toBeVisible();
        await expect(page.locator('html')).toHaveAttribute('data-theme', theme);

        // ── Recorrido completo por teclado (con foco visible en cada paso) ──
        const path = await tabCycle(page);
        expect(path[0]?.label, 'el primer Tab enfoca el enlace de salto').toContain(
          'Saltar al contenido principal',
        );
        const joined = path.map((step) => step.label).join(' | ');
        for (const mustHave of [
          'Medios',
          'Ver en YouTube',
          'Escuchar en Spotify',
          'Cargar más',
          'input es',
        ]) {
          expect(joined, `la travesía por teclado alcanza «${mustHave}»`).toContain(mustHave);
        }
        // El recorrido termina solo (no se agota el tope): no hay trampa de foco.
        expect(path.length, 'el recorrido por teclado termina (sin trampa de foco)').toBeLessThan(
          MAX_TAB_STEPS,
        );

        // ── Cada enlace externo anuncia servicio, salida y pestaña nueva ────
        const list = page.getByRole('list', { name: LIST });
        const enlaces = list.getByRole('link');
        const total = await enlaces.count();
        expect(total, 'la sección ofrece enlaces externos').toBeGreaterThan(0);
        for (let index = 0; index < total; index += 1) {
          const link = enlaces.nth(index);
          const label = (await link.getAttribute('aria-label')) ?? '';
          const servicio = label.includes('YouTube') ? 'YouTube' : 'Spotify';
          expect(label, `el anuncio nombra el servicio (${servicio})`).toContain(servicio);
          expect(label, 'el anuncio dice que se abre en una pestaña nueva').toContain(
            'en una pestaña nueva',
          );
          expect(label, 'el anuncio dice que se sale del sitio').toContain('sales del sitio');
          await expect(link).toHaveAttribute('target', '_blank');
          await expect(link).toHaveAttribute('rel', 'noopener noreferrer');
          // El servicio también se dice con **palabra visible** dentro del enlace.
          await expect(link).toContainText(
            servicio === 'YouTube' ? 'Ver en YouTube' : 'Escuchar en Spotify',
          );
          await expect(link).toContainText('pestaña nueva');
        }

        // ── El enlace externo se activa con `Enter` (pestaña nueva) ─────────
        const primero = enlaces.first();
        await primero.focus();
        const [popup] = await Promise.all([context.waitForEvent('page'), primero.press('Enter')]);
        await popup.waitForLoadState('domcontentloaded');
        expect(popup.url()).toBe(await primero.getAttribute('href'));
        await popup.close();
        expect(page.url(), 'la pestaña del sitio sigue abierta').toContain('/medios');

        // ── Sin trampas: `Shift+Tab` retrocede fuera de la navegación ───────
        const navLink = page
          .getByRole('navigation', { name: 'Secciones de la portada' })
          .getByRole('link', { name: 'Quiénes somos' });
        await navLink.focus();
        await page.keyboard.press('Shift+Tab');
        expect(
          await page.evaluate(
            () =>
              document.activeElement?.closest('nav[aria-label="Secciones de la portada"]') === null,
          ),
          'Shift+Tab sale de la lista de navegación (sin trampa de foco)',
        ).toBe(true);

        // ── Selector de idioma con teclado, sin perder el foco ─────────────
        const header = page.getByRole('banner');
        await header.getByRole('radio', { name: 'Español' }).focus();
        await page.keyboard.press('ArrowRight'); // es → en
        await expect(page.locator('html')).toHaveAttribute('lang', 'en');
        await expect(page.getByRole('heading', { level: 1, name: 'Media' })).toBeVisible();
        expect(
          await page.evaluate(() => {
            const el = document.activeElement as HTMLInputElement | null;
            return el?.tagName === 'INPUT' && el.value === 'en';
          }),
          'el foco sigue en el selector de idioma tras cambiar con las flechas',
        ).toBe(true);
        // El anuncio de los enlaces se traduce con la interfaz.
        const enIngles = page
          .getByRole('list', { name: 'List of media episodes' })
          .getByRole('link')
          .first();
        const labelEn = (await enIngles.getAttribute('aria-label')) ?? '';
        expect(labelEn, 'el anuncio en inglés nombra el servicio').toMatch(/YouTube|Spotify/);
        expect(labelEn).toContain('in a new tab');
        expect(labelEn).toContain("leaves the church's website");

        await page.keyboard.press('ArrowLeft'); // en → es
        await expect(page.locator('html')).toHaveAttribute('lang', 'es');
        await expect(page.getByRole('heading', { level: 1, name: 'Medios' })).toBeVisible();

        // ── Contraste de `main` y del pie en este tema (SC-008) ────────────
        expect(await contrastFailures(page), `contraste en tema ${theme}`).toEqual([]);

        // ── 320 px sin desplazamiento horizontal (SC-017) ──────────────────
        await page.setViewportSize({ width: 320, height: 900 });
        await page.goto('/medios');
        await expect(page.getByRole('list', { name: LIST })).toBeVisible();
        expect(
          await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth),
          'sin scroll horizontal a 320 px',
        ).toBeLessThanOrEqual(1);
      } finally {
        await context.close();
      }
    }
  });

  test('el bloque de la portada se recorre con teclado y su «Ver todo» es alcanzable', async ({
    browser,
  }) => {
    for (const theme of ['light', 'dark'] as const) {
      const { context, page } = await visitor(browser, theme);
      try {
        await page.setViewportSize({ width: 1280, height: 900 });
        await page.goto('/');
        await expect(page.locator('#medios')).toBeVisible();
        await expect(page.locator('html')).toHaveAttribute('data-theme', theme);

        // El bloque presenta sus fichas como lista rotulada para lectores.
        await expect(page.getByRole('list', { name: BLOCK_LIST })).toBeVisible();

        const path = await tabCycle(page);
        expect(path[0]?.label, 'el primer Tab enfoca el enlace de salto').toContain(
          'Saltar al contenido principal',
        );
        const joined = path.map((step) => step.label).join(' | ');
        expect(joined, `la travesía por teclado alcanza «${CTA}»`).toContain(CTA);
        expect(joined, 'alcanza los enlaces externos del bloque').toContain('Ver en YouTube');
        expect(path.length, 'el recorrido por teclado termina (sin trampa de foco)').toBeLessThan(
          MAX_TAB_STEPS,
        );

        // Foco visible en el CTA del bloque.
        const cta = page.locator('#medios').getByRole('link', { name: CTA });
        await cta.focus();
        expect(
          await page.evaluate(() => {
            const el = document.activeElement;
            if (!el || el === document.body) return false;
            const style = getComputedStyle(el);
            return (
              (style.outlineStyle !== 'none' && parseFloat(style.outlineWidth) > 0) ||
              style.boxShadow !== 'none'
            );
          }),
          'foco visible en el CTA del bloque',
        ).toBe(true);

        // Los enlaces del bloque anuncian igual que en la sección (FR-005).
        const enlacesBloque = page.getByRole('list', { name: BLOCK_LIST }).getByRole('link');
        const total = await enlacesBloque.count();
        expect(total, 'el bloque conserva los enlaces por ficha').toBeGreaterThan(0);
        for (let index = 0; index < total; index += 1) {
          const label = (await enlacesBloque.nth(index).getAttribute('aria-label')) ?? '';
          expect(label).toMatch(/YouTube|Spotify/);
          expect(label).toContain('en una pestaña nueva');
          expect(label).toContain('sales del sitio');
        }

        expect(await contrastFailures(page), `contraste del bloque en tema ${theme}`).toEqual([]);

        // 320 px sin desplazamiento horizontal (SC-017).
        await page.setViewportSize({ width: 320, height: 900 });
        await page.goto('/');
        await expect(page.locator('#medios')).toBeVisible();
        expect(
          await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth),
          'sin scroll horizontal a 320 px',
        ).toBeLessThanOrEqual(1);
      } finally {
        await context.close();
      }
    }
  });

  test('sin nada publicable, la redirección deja el foco en una portada navegable (P9-16)', async ({
    browser,
  }) => {
    seed = await seedMedios(adminPage, { publishedCount: 0 });

    const { context, page } = await visitor(browser, 'light');
    try {
      await page.goto('/medios');
      await expect(page).toHaveURL(/\/$/);
      await expect(page.locator('#proximamente')).toBeVisible();

      // El foco no queda atrapado en un callejón: la portada responde al teclado
      // desde el primer `Tab` y el ciclo de foco funciona con normalidad.
      const path = await tabCycle(page);
      expect(path[0]?.label, 'el primer Tab enfoca el enlace de salto').toContain(
        'Saltar al contenido principal',
      );
      expect(path.length, 'la portada es navegable de principio a fin').toBeGreaterThan(5);
      expect(path.length, 'sin trampas de foco tras la redirección').toBeLessThan(MAX_TAB_STEPS);

      // Ni diálogos ni contenedores modales que capturen el foco.
      await expect(page.locator('[role="dialog"]')).toHaveCount(0);
      await expect(page.locator('[aria-modal="true"]')).toHaveCount(0);
    } finally {
      await context.close();
    }
  });

  test('zoom 200 % y movimiento reducido: la sección y el bloque siguen completos (US5 esc. 4)', async ({
    browser,
  }) => {
    // La prueba anterior dejó la sección **sin publicables** (redirección): se
    // vuelve a sembrar el catálogo para comprobar la sección de verdad.
    seed = await seedMedios(adminPage, { publishedCount: 14 });

    // El zoom real escala los px CSS: se simula con un viewport de la mitad
    // (640 px) y `deviceScaleFactor: 2`, como en el precedente de F6.
    const zoomContext = await browser.newContext({
      viewport: { width: 640, height: 800 },
      deviceScaleFactor: 2,
    });
    const zoomPage = await zoomContext.newPage();
    try {
      const primera = seed.publicables[0];
      await zoomPage.goto('/medios');
      await expect(zoomPage.getByRole('heading', { level: 1, name: 'Medios' })).toBeVisible();
      await expect(zoomPage.getByRole('list', { name: LIST })).toBeVisible();
      const ficha = zoomPage
        .getByRole('list', { name: LIST })
        .getByRole('listitem')
        .filter({ hasText: primera.title });
      // La ficha no pierde nada al 200 %: nombre, descripción y enlaces.
      await expect(ficha.locator('h3')).toBeVisible();
      await expect(ficha.getByRole('link')).toHaveCount(primera.sources.length);
      await expect(
        zoomPage.getByRole('navigation', { name: 'Secciones de la portada' }),
      ).toBeVisible();

      // Y el bloque de la portada, igual.
      await zoomPage.goto('/');
      await expect(zoomPage.locator('#medios')).toBeVisible();
      await expect(zoomPage.getByRole('list', { name: BLOCK_LIST })).toBeVisible();
      await expect(zoomPage.locator('#medios').getByRole('link', { name: CTA })).toBeVisible();
    } finally {
      await zoomContext.close();
    }

    // ── `prefers-reduced-motion`: sin revelado pendiente, todo al instante ──
    const reducedContext = await browser.newContext({ colorScheme: 'light' });
    const reducedPage = await reducedContext.newPage();
    try {
      await reducedPage.emulateMedia({ reducedMotion: 'reduce' });
      await reducedPage.goto('/medios');
      await expect(reducedPage.getByRole('list', { name: LIST })).toBeVisible();
      const bloques = reducedPage.locator('.reveal-block');
      const total = await bloques.count();
      expect(total, 'la sección usa el revelado del sistema').toBeGreaterThan(0);
      for (let index = 0; index < total; index += 1) {
        const styles = await bloques.nth(index).evaluate((el) => {
          const style = getComputedStyle(el);
          return {
            reveal: el.getAttribute('data-reveal'),
            opacity: style.opacity,
            transform: style.transform,
            duration: style.transitionDuration,
          };
        });
        expect(styles.reveal, 'sin revelado pendiente').toBe('in');
        expect(styles.opacity).toBe('1');
        expect(styles.transform).toBe('none');
        expect(
          styles.duration
            .split(',')
            .map((part) => parseFloat(part))
            .every((number) => number === 0),
          'sin transición de entrada',
        ).toBe(true);
      }

      await reducedPage.goto('/');
      await expect(reducedPage.locator('#medios')).toBeVisible();
      await expect(reducedPage.getByRole('list', { name: BLOCK_LIST })).toBeVisible();
    } finally {
      await reducedContext.close();
    }
  });
});
