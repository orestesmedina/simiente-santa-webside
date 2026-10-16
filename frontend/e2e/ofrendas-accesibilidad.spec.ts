import { expect, test, type Browser, type BrowserContext, type Page } from '@playwright/test';
import { createAdminContext } from './publicSeed';
import { resetOfrendas, seedOfrendas, type OfrendasSeed } from './ofrendasSeed';

/**
 * E2E de accesibilidad por teclado de ofrendas (T735, SC-008/FR-025,
 * quickstart §11): el patrón del precedente de F6
 * (`ministerios-accesibilidad.spec.ts`), aplicado a la sección `/ofrendas` y al
 * bloque de la portada.
 *
 * Cubre:
 *   · el primer `Tab` enfoca el enlace de salto al contenido;
 *   · la travesía con `Tab` alcanza navegación, textos, fichas, el
 *     `CopyAccountButton` (activable con `Enter`, con la confirmación anunciada
 *     por `role="status"`), el selector de idioma y el pie **con foco visible en
 *     cada paso**, y el bloque de la portada con su botón «Ver formas de
 *     ofrendar»;
 *   · **sin trampas de foco**: el ciclo cierra volviendo al primer elemento y
 *     `Shift+Tab` retrocede fuera de la lista de navegación;
 *   · el selector de idioma se opera con las flechas del grupo de radios **sin
 *     perder el foco**;
 *   · todo se comprueba en **tema claro y oscuro** (F10) y a 320 px (**0
 *     desplazamiento horizontal**, SC-007);
 *   · con **0 publicados**, la redirección de `/ofrendas` a la portada no atrapa
 *     el foco: el visitante termina en una página navegable (P7-21/RG7-12);
 *   · 0 elementos `<img>` en F7 (Q6) como aserción extra.
 *
 * Queda MANUAL (no automatizable, se consigna en la revisión de QA): el lector de
 * pantalla real. Cómo se ejecutan (la base, el rate-limit de login y el navegador
 * son compartidos por toda la suite):
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

const CTA = 'Ver formas de ofrendar';
/** Tope de pasos del recorrido por teclado (la sección añade fichas al ciclo). */
const MAX_TAB_STEPS = 120;

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
 * paso exige **foco visible** (contorno o sombra). Copia del helper de F3/F6.
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
 * luminancias relativas). Copia adaptada del helper de F3/F6.
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

/** Contexto de visitante en el tema indicado, con portapapeles para la copia. */
async function visitor(
  browser: Browser,
  theme: 'light' | 'dark',
): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext({ colorScheme: theme });
  await context.grantPermissions(['clipboard-read', 'clipboard-write']);
  return { context, page: await context.newPage() };
}

test.describe('Accesibilidad por teclado de ofrendas (SC-008)', () => {
  test.setTimeout(300_000);

  let adminContext: BrowserContext;
  let adminPage: Page;
  let seed: OfrendasSeed;

  test.beforeAll(async ({ browser }) => {
    adminContext = await createAdminContext(browser);
    adminPage = await adminContext.newPage();
    seed = await seedOfrendas(adminPage);
  });

  test.afterAll(async () => {
    if (adminPage) {
      await resetOfrendas(adminPage);
    }
    await adminContext?.close();
  });

  test('la sección /ofrendas se recorre con teclado, con foco visible, en claro y oscuro', async ({
    browser,
  }) => {
    for (const theme of ['light', 'dark'] as const) {
      const { context, page } = await visitor(browser, theme);
      try {
        await page.setViewportSize({ width: 1280, height: 900 });
        await page.goto('/ofrendas');
        await expect(
          page.getByRole('heading', { level: 1, name: 'Ofrendas y diezmos' }),
        ).toBeVisible();
        await expect(page.locator('html')).toHaveAttribute('data-theme', theme);

        // ── Recorrido completo por teclado (con foco visible en cada paso) ──
        const path = await tabCycle(page);
        expect(path[0]?.label, 'el primer Tab enfoca el enlace de salto').toContain(
          'Saltar al contenido principal',
        );
        const joined = path.map((step) => step.label).join(' | ');
        for (const mustHave of ['Ofrendas y diezmos', 'Copiar número', 'input es']) {
          expect(joined, `la travesía por teclado alcanza «${mustHave}»`).toContain(mustHave);
        }
        // El recorrido termina solo (no se agota el tope): no hay trampa de foco.
        expect(path.length, 'el recorrido por teclado termina (sin trampa de foco)').toBeLessThan(
          MAX_TAB_STEPS,
        );

        // ── El botón de copia se activa con `Enter` y anuncia el resultado ──
        const copy = page.getByRole('button', { name: /^Copiar número de/ }).first();
        await copy.focus();
        await copy.press('Enter');
        const card = page
          .getByRole('region', { name: 'Cuentas IBAN' })
          .getByRole('listitem')
          .filter({ hasText: seed.published[0].institution });
        await expect(card.getByRole('status')).toHaveText('Número copiado.');

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
        await expect(
          page.getByRole('heading', { level: 1, name: 'Offerings & tithes' }),
        ).toBeVisible();
        expect(
          await page.evaluate(() => {
            const el = document.activeElement as HTMLInputElement | null;
            return el?.tagName === 'INPUT' && el.value === 'en';
          }),
          'el foco sigue en el selector de idioma tras cambiar con las flechas',
        ).toBe(true);
        await page.keyboard.press('ArrowLeft'); // en → es
        await expect(page.locator('html')).toHaveAttribute('lang', 'es');
        await expect(
          page.getByRole('heading', { level: 1, name: 'Ofrendas y diezmos' }),
        ).toBeVisible();

        // ── 0 elementos `<img>` en F7 (Q6/FR-020) ──────────────────────────
        await expect(page.locator('main img')).toHaveCount(0);

        // ── Contraste de `main` y del pie en este tema (SC-008) ────────────
        expect(await contrastFailures(page), `contraste en tema ${theme}`).toEqual([]);

        // ── 320 px sin desplazamiento horizontal (SC-007) ──────────────────
        await page.setViewportSize({ width: 320, height: 900 });
        await page.goto('/ofrendas');
        await expect(page.getByRole('region', { name: 'Cuentas IBAN' })).toBeVisible();
        expect(
          await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth),
          'sin scroll horizontal a 320 px',
        ).toBeLessThanOrEqual(1);
      } finally {
        await context.close();
      }
    }
  });

  test('el bloque de la portada se recorre con teclado y su CTA es alcanzable (SC-008)', async ({
    browser,
  }) => {
    for (const theme of ['light', 'dark'] as const) {
      const { context, page } = await visitor(browser, theme);
      try {
        await page.setViewportSize({ width: 1280, height: 900 });
        await page.goto('/');
        await expect(page.locator('#ofrendas')).toBeVisible();
        await expect(page.locator('html')).toHaveAttribute('data-theme', theme);

        // El bloque presenta sus datos como lista rotulada para lectores (las
        // fichas no son clicables: el único interactivo es el botón, D-14).
        await expect(
          page.getByRole('list', {
            name: 'Ofrendas y diezmos: mira el resumen y entra a la sección completa',
          }),
        ).toBeVisible();
        await expect(page.locator('#ofrendas li a, #ofrendas li button')).toHaveCount(0);

        const path = await tabCycle(page);
        expect(path[0]?.label, 'el primer Tab enfoca el enlace de salto').toContain(
          'Saltar al contenido principal',
        );
        const joined = path.map((step) => step.label).join(' | ');
        // El botón del bloque es alcanzable y se recorre como los demás enlaces.
        expect(joined, `la travesía por teclado alcanza «${CTA}»`).toContain(CTA);
        expect(path.length, 'el recorrido por teclado termina (sin trampa de foco)').toBeLessThan(
          MAX_TAB_STEPS,
        );

        // Foco visible en el CTA del bloque.
        const cta = page.locator('#ofrendas').getByRole('link', { name: CTA });
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

        // 0 imágenes en el bloque (F7 no lleva ninguna, Q6).
        await expect(page.locator('#ofrendas img')).toHaveCount(0);

        expect(await contrastFailures(page), `contraste del bloque en tema ${theme}`).toEqual([]);
      } finally {
        await context.close();
      }
    }
  });

  test('sin nada publicado, la redirección deja el foco en una portada navegable (P7-21)', async ({
    browser,
  }) => {
    await resetOfrendas(adminPage);

    const { context, page } = await visitor(browser, 'light');
    try {
      await page.goto('/ofrendas');
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
});
