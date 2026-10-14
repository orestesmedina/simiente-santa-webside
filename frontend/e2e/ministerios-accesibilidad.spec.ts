import { expect, test, type Browser, type BrowserContext, type Page } from '@playwright/test';
import { createAdminContext } from './publicSeed';
import {
  fixtureByKey,
  resetMinisterios,
  seedMinisterios,
  type MinisterioSeed,
} from './ministeriosSeed';

/**
 * E2E de accesibilidad por teclado de ministerios (T635, SC-008/FR-023,
 * quickstart §11): el patrón del precedente de F3
 * (`portada-accesibilidad.spec.ts`), aplicado a la sección `/ministerios` y al
 * bloque de la portada.
 *
 * Cubre:
 *   · el primer `Tab` enfoca el enlace de salto al contenido;
 *   · la travesía con `Tab` alcanza navegación, fichas, `WhatsappJoinButton`
 *     (activable con `Enter`), selector de idioma y pie **con foco visible en
 *     cada paso**, y el bloque de la portada con su botón «Ver más ministerios»;
 *   · **sin trampas de foco**: el ciclo cierra volviendo al primer elemento y
 *     `Shift+Tab` retrocede fuera de la lista de navegación;
 *   · el selector de idioma se opera con las flechas del grupo de radios **sin
 *     perder el foco**;
 *   · todo se comprueba en **tema claro y oscuro** (F10), a 320 px y con el
 *     texto al **200 %**;
 *   · con **0 publicados**, la redirección de `/ministerios` a la portada no
 *     atrapa el foco: el visitante termina en una página navegable
 *     (P6-21/RG6-13).
 *
 * Queda MANUAL (no automatizable, se consigna en la revisión de QA): el lector de
 * pantalla real. Cómo se ejecutan (la base, el rate-limit de login y el navegador
 * son compartidos por toda la suite):
 *
 *   # Tanda 4 — ministerios (F6): las cuatro specs comparten la siembra/limpieza
 *   # de ministerios, así que van juntas y con un solo worker.
 *   npx playwright test --config e2e/playwright.config.ts ministerios- portada-ministerios --workers=1
 *
 * La suite completa se reparte en **cuatro tandas** porque `POST /api/v1/auth/login`
 * limita a 20 intentos por minuto y por IP: con todo en una sola corrida, los
 * últimos specs reciben `429` y no llegan a entrar.
 *
 *   # Tanda 1 — F2 (acceso y auditoría: la que más inicios de sesión hace)
 *   npx playwright test --config e2e/playwright.config.ts acceso auditoria --workers=1
 *   # Tanda 2 — F4 (agenda)
 *   npx playwright test --config e2e/playwright.config.ts agenda- --workers=1
 *   # Tanda 3 — F3 y F10 (portada, sistema de diseño, tema y estado)
 *   npx playwright test --config e2e/playwright.config.ts portada- diseno- movimiento navegacion- tema- qa- status --workers=1
 *   # Tanda 4 — F6 (ministerios y su bloque de portada)
 *   npx playwright test --config e2e/playwright.config.ts ministerios- portada-ministerios --workers=1
 *
 * `--workers=1` es obligatorio en las cuatro (singletons compartidos y
 * ministerios sembrados por spec). En este entorno el navegador no arranca desde
 * el host —faltan las librerías del sistema (`libnspr4`)—, así que los comandos
 * se ejecutan dentro de la imagen oficial de Playwright, con red del host para
 * llegar al stack de `make up`:
 *
 *   docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
 *     -v "$PWD/..":/work -w /work/frontend mcr.microsoft.com/playwright:v1.63.0-jammy \
 *     npx playwright test --config e2e/playwright.config.ts ministerios- portada-ministerios --workers=1
 *
 * (El stack debe estar reconstruido con F6: `docker compose up -d --build`.)
 */
test.describe.configure({ mode: 'serial' });

/** Nombre accesible de la lista de fichas (i18n `es`). */
const SECTION = 'Ministerios de la iglesia';
/** Tope de pasos del recorrido por teclado (la sección añade fichas al ciclo). */
const MAX_TAB_STEPS = 45;

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
 * paso exige **foco visible** (contorno o sombra). Copia del helper de F3.
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
 * luminancias relativas). Copia adaptada del helper de F3.
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

/** Contexto de visitante en el tema indicado (`colorScheme` sin elección guardada). */
async function visitor(
  browser: Browser,
  theme: 'light' | 'dark',
): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext({ colorScheme: theme });
  return { context, page: await context.newPage() };
}

test.describe('Accesibilidad por teclado de ministerios (SC-008)', () => {
  test.setTimeout(300_000);

  let adminContext: BrowserContext;
  let adminPage: Page;
  let seed: MinisterioSeed;

  test.beforeAll(async ({ browser }) => {
    adminContext = await createAdminContext(browser);
    adminPage = await adminContext.newPage();
    seed = await seedMinisterios(adminPage);
  });

  test.afterAll(async () => {
    if (adminPage) {
      await resetMinisterios(adminPage);
    }
    await adminContext?.close();
  });

  test('la sección /ministerios se recorre con teclado, con foco visible, en claro y oscuro', async ({
    browser,
  }) => {
    const alabanza = fixtureByKey(seed, 'alabanza');
    const jovenes = fixtureByKey(seed, 'jovenes');

    for (const theme of ['light', 'dark'] as const) {
      const { context, page } = await visitor(browser, theme);
      try {
        await page.setViewportSize({ width: 1280, height: 900 });
        await page.goto('/ministerios');
        await expect(page.getByRole('heading', { level: 1, name: 'Ministerios' })).toBeVisible();
        await expect(page.locator('html')).toHaveAttribute('data-theme', theme);

        // ── Recorrido completo por teclado (con foco visible en cada paso) ──
        const path = await tabCycle(page);
        expect(path[0]?.label, 'el primer Tab enfoca el enlace de salto').toContain(
          'Saltar al contenido principal',
        );
        const joined = path.map((step) => step.label).join(' | ');
        // Navegación, botón de unión de una ficha y selector de idioma, en un
        // orden comprensible (cabecera → contenido → pie).
        for (const mustHave of ['Ministerios', 'Unirse al grupo de WhatsApp', 'input es']) {
          expect(joined, `la travesía por teclado alcanza «${mustHave}»`).toContain(mustHave);
        }
        // El recorrido termina solo (no se agota el tope): no hay trampa de
        // foco que devuelva siempre al mismo elemento.
        expect(path.length, 'el recorrido por teclado termina (sin trampa de foco)').toBeLessThan(
          MAX_TAB_STEPS,
        );

        // ── El botón de unión se activa con `Enter` (FR-003/FR-023) ────────
        // La navegación externa se responde con un stub local para no depender
        // de la red y conservar la URL de destino.
        const stub = {
          status: 200,
          contentType: 'text/html',
          body: '<!doctype html><title>stub</title><p>stub</p>',
        };
        await context.route('https://chat.whatsapp.com/**', (route) => route.fulfill(stub));
        await context.route('https://wa.me/**', (route) => route.fulfill(stub));
        const join = page.getByRole('link', {
          name: `Unirse al grupo de WhatsApp de ${alabanza.nameEs}`,
        });
        await join.focus();
        const [popup] = await Promise.all([page.waitForEvent('popup'), join.press('Enter')]);
        await expect.poll(() => popup.url()).toBe(alabanza.whatsappUrl);
        await popup.close();
        await expect(
          page.getByRole('link', { name: `Unirse al grupo de WhatsApp de ${jovenes.nameEs}` }),
        ).toHaveCount(1);

        // ── Sin trampas: `Shift+Tab` retrocede fuera de la navegación ──────
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

        // …y desde la primera ficha del contenido también se puede retroceder.
        const firstJoin = page.getByRole('link', { name: /Unirse al grupo de WhatsApp/ }).first();
        await firstJoin.focus();
        await page.keyboard.press('Shift+Tab');
        expect(
          await page.evaluate(() => document.activeElement !== document.body),
          'Shift+Tab desde la ficha retrocede sin quedar atrapado',
        ).toBe(true);

        // ── Selector de idioma con teclado, sin perder el foco ─────────────
        const header = page.getByRole('banner');
        await header.getByRole('radio', { name: 'Español' }).focus();
        await page.keyboard.press('ArrowRight'); // es → en
        await expect(page.locator('html')).toHaveAttribute('lang', 'en');
        await expect(page.getByRole('heading', { level: 1, name: 'Ministries' })).toBeVisible();
        expect(
          await page.evaluate(() => {
            const el = document.activeElement as HTMLInputElement | null;
            return el?.tagName === 'INPUT' && el.value === 'en';
          }),
          'el foco sigue en el selector de idioma tras cambiar con las flechas',
        ).toBe(true);
        await page.keyboard.press('ArrowLeft'); // en → es
        await expect(page.locator('html')).toHaveAttribute('lang', 'es');
        await expect(page.getByRole('heading', { level: 1, name: 'Ministerios' })).toBeVisible();

        // ── Contraste de `main` y del pie en este tema (SC-008) ────────────
        expect(await contrastFailures(page), `contraste en tema ${theme}`).toEqual([]);

        // ── 320 px sin desplazamiento horizontal ───────────────────────────
        await page.setViewportSize({ width: 320, height: 900 });
        await page.goto('/ministerios');
        await expect(page.getByRole('list', { name: SECTION, exact: true })).toBeVisible();
        expect(
          await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth),
          'sin scroll horizontal a 320 px',
        ).toBeLessThanOrEqual(1);
      } finally {
        await context.close();
      }
    }

    // ── Texto al 200 % SIN pérdida de información (WCAG 1.4.4) ────────────
    // El zoom real escala los px CSS: se simula con un viewport de la mitad
    // (160 px) y `deviceScaleFactor: 2`, como en el precedente de F3.
    const zoomContext = await browser.newContext({
      viewport: { width: 160, height: 450 },
      deviceScaleFactor: 2,
    });
    const zoomPage = await zoomContext.newPage();
    try {
      await zoomPage.goto('/ministerios');
      for (const locator of [
        zoomPage.getByRole('heading', { level: 1, name: 'Ministerios' }),
        zoomPage.getByRole('list', { name: SECTION, exact: true }),
        zoomPage.getByText(alabanza.nameEs),
        zoomPage.getByText(alabanza.scheduleEs!),
        zoomPage.getByRole('link', {
          name: `Unirse al grupo de WhatsApp de ${alabanza.nameEs}`,
        }),
        zoomPage.getByRole('navigation', { name: 'Secciones de la portada' }),
      ]) {
        await expect(
          locator.first(),
          'con el texto al 200 % se conserva todo el contenido',
        ).toBeVisible();
      }
    } finally {
      await zoomContext.close();
    }
  });

  test('el bloque de la portada se recorre con teclado y su CTA es alcanzable (SC-008)', async ({
    browser,
  }) => {
    const first = seed.portada[0];
    const last = seed.portada[seed.portada.length - 1];

    for (const theme of ['light', 'dark'] as const) {
      const { context, page } = await visitor(browser, theme);
      try {
        await page.setViewportSize({ width: 1280, height: 900 });
        await page.goto('/');
        await expect(page.locator('#ministerios')).toBeVisible();
        await expect(page.locator('html')).toHaveAttribute('data-theme', theme);

        const path = await tabCycle(page);
        expect(path[0]?.label, 'el primer Tab enfoca el enlace de salto').toContain(
          'Saltar al contenido principal',
        );
        const joined = path.map((step) => step.label).join(' | ');
        // El helper recorta la etiqueta a 40 caracteres: se compara el prefijo.
        for (const mustHave of [
          first.nameEs.slice(0, 25),
          last.nameEs.slice(0, 25),
          'Ver más ministerios',
        ]) {
          expect(joined, `la travesía por teclado alcanza «${mustHave}»`).toContain(mustHave);
        }
        expect(path.length, 'el recorrido por teclado termina (sin trampa de foco)').toBeLessThan(
          MAX_TAB_STEPS,
        );

        expect(await contrastFailures(page), `contraste del bloque en tema ${theme}`).toEqual([]);
      } finally {
        await context.close();
      }
    }
  });

  test('sin publicados, la redirección deja el foco en una portada navegable (P6-21)', async ({
    browser,
  }) => {
    await resetMinisterios(adminPage);

    const { context, page } = await visitor(browser, 'light');
    try {
      await page.goto('/ministerios');
      await expect(page).toHaveURL(/\/$/);
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible();

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
