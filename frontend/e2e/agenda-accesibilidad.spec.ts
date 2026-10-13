import { expect, test, type Page } from '@playwright/test';
import { createAdminContext } from './publicSeed';
import { resetAgenda, seedAgenda, type AgendaSeed } from './agendaSeed';

/**
 * E2e de accesibilidad por teclado de la sección pública `/eventos` (T442,
 * SC-008/FR-024, quickstart §10.3): el patrón del precedente de F3
 * (`portada-accesibilidad.spec.ts`).
 *
 * Cubre:
 *   · el primer `Tab` enfoca el enlace de salto al contenido;
 *   · la travesía con `Tab` alcanza navegación, tarjetas de cada bloque,
 *     «Ver ubicación», selector de idioma y pie **con foco visible en cada paso**;
 *   · activación con `Enter` (enlace de salto) y con `Espacio` (radios);
 *   · **sin trampas de foco**: `Shift+Tab` retrocede y el ciclo vuelve al primer
 *     elemento;
 *   · el selector de idioma se opera con las flechas del grupo de radios sin
 *     perder el foco;
 *   · todo se comprueba en **tema claro y oscuro** (F10);
 *   · con la sección **vacía**, el aviso de `HiddenNotice` sigue siendo
 *     alcanzable por teclado y **sin redirigir** (⟲ D-13).
 *
 * Cómo se ejecutan (la base, el rate-limit de login y el navegador son
 * compartidos por toda la suite):
 *
 *   # Tanda 2 — agenda (F4): los tres specs comparten la siembra/limpieza de la
 *   # agenda, así que van juntos y con un solo worker.
 *   npx playwright test --config e2e/playwright.config.ts agenda- --workers=1
 *
 * La suite completa se reparte en **tres tandas** porque `POST /api/v1/auth/login`
 * limita a 20 intentos por minuto y por IP: con todo en una sola corrida, los
 * últimos specs reciben `429` y no llegan a entrar.
 *
 *   # Tanda 1 — F2 (acceso y auditoría: la que más inicios de sesión hace)
 *   npx playwright test --config e2e/playwright.config.ts acceso auditoria --workers=1
 *   # Tanda 2 — F4 (agenda)
 *   npx playwright test --config e2e/playwright.config.ts agenda- --workers=1
 *   # Tanda 3 — F3 y F10 (portada, sistema de diseño, tema y estado)
 *   npx playwright test --config e2e/playwright.config.ts portada- diseno- movimiento navegacion- tema- qa- status --workers=1
 *
 * `--workers=1` es obligatorio en las tres (singletons compartidos y agenda
 * sembrada por spec). En este entorno el navegador no arranca desde el host
 * —faltan las librerías del sistema (`libnspr4`)—, así que los comandos se
 * ejecutan dentro de la imagen oficial de Playwright, con red del host para
 * llegar al stack de `make up`:
 *
 *   docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
 *     -v "$PWD/..":/work -w /work/frontend mcr.microsoft.com/playwright:v1.63.0-jammy \
 *     npx playwright test --config e2e/playwright.config.ts agenda- --workers=1
 *
 * (El stack debe estar reconstruido con F4: `docker compose up -d --build`.)
 */

/**
 * Enfoca el enlace de salto con el teclado. Reintenta el `Tab` mientras el foco
 * siga en el cuerpo: la SPA puede montar (y re-renderizar) justo después de la
 * primera pulsación y el foco se perdería.
 */
async function focusSkipLink(page: Page): Promise<void> {
  const skip = page.getByRole('link', { name: 'Saltar al contenido principal' });
  for (let attempt = 0; attempt < 4; attempt += 1) {
    await page.keyboard.press('Tab');
    if (await skip.evaluate((element) => element === document.activeElement)) {
      return;
    }
    if (!(await page.evaluate(() => document.activeElement === document.body))) {
      break; // el foco está en otro control: no es cuestión de reintentar
    }
  }
}

interface TabStep {
  label: string;
  indicator: boolean;
}

interface StepInfo {
  label: string | null;
  indicator: boolean;
}

/**
 * Recorre con `Tab` hasta cerrar el ciclo (vuelve al primer elemento) exigiendo
 * foco **visible** en cada paso (mismo criterio que el e2e de F3).
 */
async function tabCycle(page: Page, maxSteps = 40): Promise<TabStep[]> {
  const path: TabStep[] = [];
  for (let index = 0; index < maxSteps; index += 1) {
    await page.keyboard.press('Tab');
    const info = await page.evaluate<StepInfo>(() => {
      const element = document.activeElement;
      if (!element || element === document.body) {
        return { label: null, indicator: false };
      }
      const style = getComputedStyle(element);
      const accessible =
        element.getAttribute('aria-label') ||
        (element as HTMLElement).innerText ||
        (element as HTMLInputElement).value ||
        '';
      return {
        label: `${element.tagName.toLowerCase()} ${accessible.trim().slice(0, 60)}`,
        indicator:
          (style.outlineStyle !== 'none' && parseFloat(style.outlineWidth) > 0) ||
          style.boxShadow !== 'none',
      };
    });

    if (info.label === null) {
      if (path.length === 0 && index < 3) {
        continue; // la SPA aún no montó: se reintenta el salto
      }
      break;
    }
    expect(info.indicator, `foco visible en «${info.label}» (paso ${index + 1})`).toBe(true);
    if (path.length > 0 && path[0].label === info.label) {
      path.push({ label: info.label, indicator: info.indicator });
      break; // el ciclo cerró: volvió al primer elemento
    }
    path.push({ label: info.label, indicator: info.indicator });
  }
  return path;
}

test.describe('Accesibilidad por teclado de /eventos (SC-008, FR-024)', () => {
  test.setTimeout(240_000);

  test('teclado, foco visible, sin trampas y aviso alcanzable en claro y oscuro', async ({
    browser,
  }) => {
    // ── Siembra ────────────────────────────────────────────────────────────
    const adminContext = await createAdminContext(browser);
    const adminPage = await adminContext.newPage();
    let seed: AgendaSeed;
    try {
      await resetAgenda(adminPage);
      seed = await seedAgenda(adminPage);
    } finally {
      await adminContext.close();
    }

    const context = await browser.newContext({ viewport: { width: 1280, height: 900 } });
    const page = await context.newPage();
    try {
      await page.goto('/eventos');
      await expect(
        page.getByRole('heading', { level: 1, name: 'Eventos y actividades' }),
      ).toBeVisible();

      for (const theme of ['light', 'dark'] as const) {
        // Cada tema se comprueba sobre una **carga limpia**: el navegador
        // recuerda el último elemento enfocado y el «primer Tab» solo es el
        // enlace de salto cuando la travesía arranca desde el principio.
        await page.goto('/eventos');
        await expect(
          page.getByRole('heading', { level: 1, name: 'Eventos y actividades' }),
        ).toBeVisible();

        if (theme === 'dark') {
          // El interruptor de tema es un botón con nombre accesible dinámico y
          // estado anunciado; se activa con teclado y no pierde el foco.
          const toggle = page.getByRole('button', { name: 'Cambiar a tema oscuro' });
          await toggle.focus();
          await page.keyboard.press('Enter');
          await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
          await expect(page.getByRole('button', { name: 'Cambiar a tema claro' })).toBeFocused();
          await page.goto('/eventos'); // recarga limpia con el tema ya guardado
        }
        await expect(page.locator('html')).toHaveAttribute('data-theme', theme);

        // ── El primer Tab enfoca el enlace de salto al contenido ────────────
        await focusSkipLink(page);
        const skip = page.getByRole('link', { name: 'Saltar al contenido principal' });
        await expect(skip).toBeFocused();
        await expect(skip).toBeVisible();

        // ── Travesía completa con foco visible en cada paso ────────────────
        const path = await tabCycle(page);
        const labels = path.map((step) => step.label);
        const joined = labels.join(' | ');
        expect(labels[0], `el recorrido empieza en el logotipo (tema ${theme})`).toContain(
          'Simiente',
        );
        for (const expected of [
          'Eventos y actividades',
          `Ver ubicación de "${seed.eventTitle}"`,
          `Ver ubicación de "${seed.activityTitle}"`,
          // En un grupo de radios nativo el `Tab` solo entra por el marcado:
          // el resto se recorre con las flechas (se comprueba más abajo).
          'input es',
          'Cambiar a tema',
        ]) {
          expect(
            joined,
            `el recorrido por teclado alcanza «${expected}» (tema ${theme})`,
          ).toContain(expected);
        }
        // Sin trampas de foco: la travesía termina porque el foco sale del
        // documento (no se agota el tope de pasos ni se queda dando vueltas).
        expect(path.length, `la travesía no se agota en el tope (tema ${theme})`).toBeLessThan(40);
        expect(
          await page.evaluate(() => document.activeElement === document.body),
          `el foco sale del documento al final de la travesía (tema ${theme})`,
        ).toBe(true);

        // ── `Shift+Tab` retrocede: no hay trampa de foco ────────────────────
        await page.goto('/eventos');
        await focusSkipLink(page);
        await expect(skip).toBeFocused();
        await page.keyboard.press('Shift+Tab');
        const back = await page.evaluate(() => {
          const element = document.activeElement;
          return element
            ? `${element.tagName.toLowerCase()} ${(element as HTMLElement).innerText || ''}`
                .trim()
                .slice(0, 40)
            : null;
        });
        expect(back, 'Shift+Tab sale del primer elemento en lugar de quedarse atrapado').not.toBe(
          'a Saltar al contenido principal',
        );
      }

      // ── Activación con `Enter` del enlace de salto ────────────────────────
      await page.goto('/eventos');
      await expect(
        page.getByRole('heading', { level: 1, name: 'Eventos y actividades' }),
      ).toBeVisible();
      await page.keyboard.press('Tab');
      await page.keyboard.press('Enter');
      await expect(page).toHaveURL(/#contenido$/);
      await expect(page.locator('#contenido')).toBeVisible();

      // ── Selector de idioma con las flechas, sin perder el foco ────────────
      const spanish = page.locator('header').getByRole('radio', { name: 'Español' });
      await spanish.focus();
      await expect(spanish).toBeFocused();
      await page.keyboard.press('ArrowRight');
      await expect(page.locator('html')).toHaveAttribute('lang', 'en');
      await expect(page.locator('header').getByRole('radio', { name: 'English' })).toBeChecked();
      await expect(page.locator('header').getByRole('radio', { name: 'English' })).toBeFocused();
      await page.keyboard.press('ArrowLeft');
      await expect(page.locator('html')).toHaveAttribute('lang', 'es');
      await expect(page.locator('header').getByRole('radio', { name: 'Español' })).toBeChecked();
      await expect(spanish).toBeFocused();

      // ── `Espacio` activa los radios (grupo nativo) ───────────────────────
      await page.keyboard.press('Space');
      await expect(page.locator('header').getByRole('radio', { name: 'Español' })).toBeChecked();
      await expect(spanish).toBeFocused();

      // ── Menú móvil (320 px): la misma lista, alcanzable por teclado ───────
      await page.setViewportSize({ width: 320, height: 900 });
      await page.goto('/eventos');
      await expect(
        page.getByRole('heading', { level: 1, name: 'Eventos y actividades' }),
      ).toBeVisible();
      // El nombre accesible del control cambia al abrirse («Abrir»/«Cerrar»):
      // se ancla el locator al texto común para que siga resolviendo.
      const menuButton = page.getByRole('button', { name: /menú de secciones/ });
      await menuButton.focus();
      await page.keyboard.press('Enter');
      await expect(page.getByRole('button', { name: 'Cerrar menú de secciones' })).toHaveAttribute(
        'aria-expanded',
        'true',
      );
      const eventsLink = page
        .getByRole('navigation', { name: 'Secciones de la portada' })
        .getByRole('link', { name: 'Eventos y actividades' });
      let reachedInMenu = false;
      for (let step = 0; step < 10 && !reachedInMenu; step += 1) {
        await page.keyboard.press('Tab');
        reachedInMenu = await eventsLink.evaluate((element) => element === document.activeElement);
      }
      expect(reachedInMenu, 'el menú móvil lleva por teclado al ítem de la agenda').toBe(true);
      await expect(eventsLink).toBeFocused();
      // Y se sale del panel con `Shift+Tab` (sin trampa de foco): el foco
      // retrocede al ítem anterior de la misma lista, no se queda atrapado.
      await page.keyboard.press('Shift+Tab');
      const previous = await page.evaluate(() => document.activeElement?.textContent?.trim() ?? '');
      expect(previous, 'Shift+Tab retrocede dentro del menú móvil (sin trampa de foco)').not.toBe(
        'Eventos y actividades',
      );

      // ── Sección vacía: el aviso es alcanzable y no redirige (⟲ D-13) ─────
      // (contexto de administrador nuevo: el de la siembra ya se cerró; el
      // `storageState` está cacheado, así que no cuesta otro inicio de sesión).
      const cleanupContext = await createAdminContext(browser);
      const cleanupPage = await cleanupContext.newPage();
      try {
        await resetAgenda(cleanupPage);
      } finally {
        await cleanupContext.close();
      }
      await page.setViewportSize({ width: 1280, height: 900 });
      await page.goto('/eventos');
      await focusSkipLink(page);
      await expect(page.getByRole('link', { name: 'Saltar al contenido principal' })).toBeFocused();
      await page.keyboard.press('Enter');
      await expect(page).toHaveURL(/\/eventos#contenido$/);
      await expect(
        page.locator('#contenido').getByText('Por ahora no hay eventos ni actividades'),
      ).toBeVisible();
      // El recorrido por teclado sigue funcionando y el ítem no existe.
      const emptyPath = await tabCycle(page);
      expect(emptyPath.length).toBeGreaterThan(4);
      expect(emptyPath.length).toBeLessThan(40);
      await expect(page.getByRole('link', { name: 'Eventos y actividades' })).toHaveCount(0);
    } finally {
      await context.close();
    }
  });
});
