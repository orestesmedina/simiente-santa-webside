import { expect, test, type BrowserContext, type Page } from '@playwright/test';
import { apiSend } from './helpers';
import { createAdminContext } from './publicSeed';
import { resetMinisterios, seedMinisterios, type MinisterioSeed } from './ministeriosSeed';

/**
 * E2E del bloque «Ministerios» de la portada (US6, SC-016, quickstart §14): con
 * 8 ministerios publicados el bloque muestra **exactamente los 5 primeros** en el
 * orden manual (`sortOrder`, `id`), con su imagen, el chip «Ministerio» y el
 * horario/lugar cuando existen; con menos muestra **exactos** los que hay, **sin
 * huecos ni marcadores** (la tarjeta sin imagen no reserva espacio); con **0
 * publicados el bloque no existe** —ni título ni mensaje, a diferencia del bloque
 * «Lo que viene» de F4— y la portada queda coherente (SC-010/FR-028); el bloque
 * está **inmediatamente después del bloque de eventos y antes de «Quiénes
 * somos»** (FR-027/FR-029); el título-enlace de cada tarjeta y el botón «Ver más
 * ministerios» llevan a `/ministerios` en un solo paso; **0 borradores** (SC-002);
 * a 320 px no hay desplazamiento horizontal (SC-007) y el bloque se recorre con
 * teclado con foco visible (SC-008).
 *
 * Cómo se ejecutan (la base, el rate-limit de login y el navegador son
 * compartidos por toda la suite):
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
 * Nota: el filtro `portada-` de la tanda 3 (F3/F10) también incluye este archivo;
 * su tanda natural es la 4, pero si corre dentro de la 3 no rompe nada (se siembra
 * a sí mismo y usa el contexto de administrador compartido). `--workers=1` es
 * obligatorio en las cuatro (singletons compartidos y ministerios sembrados por
 * spec). En este entorno el navegador no arranca desde el host —faltan las
 * librerías del sistema (`libnspr4`)—, así que los comandos se ejecutan dentro de
 * la imagen oficial de Playwright, con red del host para llegar al stack de
 * `make up`:
 *
 *   docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
 *     -v "$PWD/..":/work -w /work/frontend mcr.microsoft.com/playwright:v1.63.0-jammy \
 *     npx playwright test --config e2e/playwright.config.ts ministerios- portada-ministerios --workers=1
 *
 * (El stack debe estar reconstruido con F6: `docker compose up -d --build`.)
 */
test.describe.configure({ mode: 'serial' });

const BLOCK = '#ministerios';
const BADGE = 'Ministerio';
const CTA = 'Ver más ministerios';

/** Tarjetas del bloque (una por ministerio del extracto). */
function blockCards(page: Page) {
  return page.locator(`${BLOCK} ul > li`);
}

/** `true` si el documento no tiene desplazamiento horizontal. */
async function hasNoHorizontalScroll(page: Page): Promise<boolean> {
  return page.evaluate(
    () => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1,
  );
}

/** Espera a que la portada esté pintada (hero + bloque de eventos de F4). */
async function waitForPortada(page: Page): Promise<void> {
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  await expect(page.locator('#proximamente')).toBeVisible();
}

/**
 * Comprueba el orden de aparición en el DOM de una secuencia de `id`s: cada uno
 * sigue al anterior (`compareDocumentPosition` + `DOCUMENT_POSITION_FOLLOWING`).
 */
async function domOrder(page: Page, ids: string[]): Promise<boolean[]> {
  return page.evaluate((list) => {
    const nodes = list.map((id) => document.getElementById(id));
    return nodes.map((node, index) => {
      if (index === 0) {
        return node !== null;
      }
      const previous = nodes[index - 1];
      if (!previous || !node) {
        return false;
      }
      return Boolean(previous.compareDocumentPosition(node) & Node.DOCUMENT_POSITION_FOLLOWING);
    });
  }, ids);
}

/** `true` si el elemento enfocado pinta un indicador de foco visible. */
async function hasFocusIndicator(page: Page): Promise<boolean> {
  return page.evaluate(() => {
    const el = document.activeElement;
    if (!el || el === document.body) {
      return false;
    }
    const style = getComputedStyle(el);
    return (
      (style.outlineStyle !== 'none' && parseFloat(style.outlineWidth) > 0) ||
      style.boxShadow !== 'none'
    );
  });
}

test.describe('Bloque «Ministerios» de la portada (US6, SC-016)', () => {
  test.setTimeout(300_000);

  let adminContext: BrowserContext;
  let adminPage: Page;
  let seed: MinisterioSeed;

  test.beforeAll(async ({ browser }) => {
    adminContext = await createAdminContext(browser);
    adminPage = await adminContext.newPage();
    // «Quiénes somos» publicado: hace falta para comprobar que el bloque de
    // ministerios queda **antes** de esa sección (FR-027).
    await apiSend(adminPage, 'PUT', '/api/v1/admin/portada/quienes-somos', {
      textEs: 'Somos la iglesia del barrio.',
      publicationState: 'published',
    });
    seed = await seedMinisterios(adminPage);
  });

  test.afterAll(async () => {
    if (adminPage) {
      await resetMinisterios(adminPage);
    }
    await adminContext?.close();
  });

  test('con 8 publicados muestra exactamente los 5 primeros, en orden y en su posición (SC-016)', async ({
    browser,
  }) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    try {
      await page.setViewportSize({ width: 1280, height: 900 });
      await page.goto('/');
      await waitForPortada(page);

      // ── Exactamente 5 (tope firme del servidor) ──────────────────────────
      await expect(page.locator(BLOCK)).toBeVisible();
      await expect(page.getByRole('heading', { level: 2, name: 'Ministerios' })).toBeVisible();
      const cards = blockCards(page);
      await expect(cards).toHaveCount(5);

      // Orden manual: los 5 primeros del listado publicado.
      expect(await cards.locator('h3').allInnerTexts()).toEqual(
        seed.portada.map((fixture) => fixture.nameEs),
      );
      // El sexto en adelante no aparece.
      for (const fixture of seed.published.slice(5)) {
        await expect(page.locator(BLOCK).getByText(fixture.nameEs)).toHaveCount(0);
      }
      // 0 borradores por ninguna vía (SC-002).
      const html = await page.evaluate(() => document.documentElement.outerHTML);
      for (const draft of seed.drafts) {
        expect(html).not.toContain(draft.nameEs);
      }

      // ── Chip «Ministerio», imagen con alt y horario/lugar si existen ─────
      for (const fixture of seed.portada) {
        const card = cards.filter({ hasText: fixture.nameEs });
        await expect(card.getByText(BADGE, { exact: true })).toBeVisible();
        if (fixture.imageFile) {
          await expect(card.getByRole('img')).toHaveAttribute('alt', fixture.imageAltEs!);
        } else {
          // Sin imagen no hay hueco reservado.
          await expect(card.getByRole('img')).toHaveCount(0);
          await expect(card.locator('.aspect-video')).toHaveCount(0);
        }
        if (fixture.scheduleEs) {
          await expect(card).toContainText(fixture.scheduleEs);
        }
        if (fixture.placeEs) {
          await expect(card).toContainText(fixture.placeEs);
        }
      }

      // ── Posición: después de «Lo que viene» y antes de «Quiénes somos» ───
      expect(
        await domOrder(page, ['proximamente', 'ministerios', 'quienes-somos']),
        'el bloque va inmediatamente después de eventos y antes de «quiénes somos»',
      ).toEqual([true, true, true]);

      // ── Un solo paso a `/ministerios` desde el título y desde el CTA ─────
      const first = seed.portada[0];
      const titleLink = page.locator(BLOCK).getByRole('link', {
        name: `${BADGE}: ${first.nameEs}`,
      });
      await expect(titleLink).toHaveAttribute('href', '/ministerios');
      const cta = page.locator(BLOCK).getByRole('link', { name: CTA });
      await expect(cta).toHaveAttribute('href', '/ministerios');

      await titleLink.click();
      await expect(page).toHaveURL(/\/ministerios$/);
      await expect(page.getByRole('heading', { level: 1, name: 'Ministerios' })).toBeVisible();
      await page.goBack();
      await waitForPortada(page);
      await page.locator(BLOCK).getByRole('link', { name: CTA }).click();
      await expect(page).toHaveURL(/\/ministerios$/);
    } finally {
      await context.close();
    }
  });

  test('con menos publicados muestra exactos, sin huecos ni marcadores (SC-016/SC-017)', async ({
    browser,
  }) => {
    seed = await seedMinisterios(adminPage, { publishedCount: 2, withDrafts: false });
    const withImage = seed.published[0];

    const context = await browser.newContext();
    const page = await context.newPage();
    try {
      await page.goto('/');
      await waitForPortada(page);

      const cards = blockCards(page);
      await expect(cards).toHaveCount(2);
      expect(await cards.locator('h3').allInnerTexts()).toEqual(
        seed.published.map((fixture) => fixture.nameEs),
      );

      // La tarjeta con imagen la muestra con su alt…
      await expect(cards.nth(0).getByRole('img')).toHaveAttribute('alt', withImage.imageAltEs!);
      // …y la que no la tiene **no reserva** espacio (sin `img` ni caja 16:9).
      await expect(cards.nth(1).getByRole('img')).toHaveCount(0);
      await expect(cards.nth(1).locator('.aspect-video')).toHaveCount(0);
      // Sin marcadores de relleno en ninguna tarjeta.
      for (const card of [cards.nth(0), cards.nth(1)]) {
        await expect(card.getByText('—', { exact: true })).toHaveCount(0);
      }

      // El CTA sigue presente y la posición no cambia.
      await expect(page.locator(BLOCK).getByRole('link', { name: CTA })).toBeVisible();
      expect(await domOrder(page, ['proximamente', 'ministerios', 'quienes-somos'])).toEqual([
        true,
        true,
        true,
      ]);
    } finally {
      await context.close();
    }
  });

  test('sin publicados el bloque no existe y la portada queda coherente (FR-028/SC-010)', async ({
    browser,
  }) => {
    await resetMinisterios(adminPage);

    const context = await browser.newContext();
    const page = await context.newPage();
    try {
      await page.goto('/');
      await waitForPortada(page);

      // Sin título, sin mensaje y sin hueco: el bloque **no se dibuja**.
      await expect(page.locator(BLOCK)).toHaveCount(0);
      await expect(page.getByRole('heading', { level: 2, name: 'Ministerios' })).toHaveCount(0);
      await expect(page.getByRole('link', { name: CTA })).toHaveCount(0);
      await expect(page.getByText('Las áreas donde puedes servir')).toHaveCount(0);
      // A diferencia de «Lo que viene» (F4), que sí mantiene su mensaje.
      await expect(page.locator('#proximamente')).toBeVisible();

      // La portada sigue coherente y navegable…
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
      await expect(page.locator('#quienes-somos')).toBeVisible();
      await expect(page.getByRole('navigation', { name: 'Secciones de la portada' })).toBeVisible();
      // …y ninguna superficie enlaza la sección (0 enlaces rotos, SC-010).
      await expect(page.locator('a[href="/ministerios"]')).toHaveCount(0);
      expect(await hasNoHorizontalScroll(page)).toBe(true);
    } finally {
      await context.close();
    }
  });

  test('320 px sin scroll horizontal y recorrido por teclado con foco visible (SC-007/SC-008)', async ({
    browser,
  }) => {
    seed = await seedMinisterios(adminPage);

    const context = await browser.newContext();
    const page = await context.newPage();
    try {
      await page.setViewportSize({ width: 320, height: 900 });
      await page.goto('/');
      await waitForPortada(page);
      await expect(page.locator(BLOCK)).toBeVisible();
      expect(await hasNoHorizontalScroll(page), 'sin desplazamiento horizontal a 320 px').toBe(
        true,
      );

      // El bloque se apila en una columna a 320 px.
      const columns = await page.evaluate(
        () =>
          getComputedStyle(document.querySelector('#ministerios ul') as HTMLElement)
            .gridTemplateColumns,
      );
      expect(columns.trim().split(/\s+/).length, 'una sola columna a 320 px').toBe(1);

      // Foco visible en el título-enlace y en el CTA del bloque.
      const titleLink = page
        .locator(BLOCK)
        .getByRole('link', { name: `${BADGE}: ${seed.portada[0].nameEs}` });
      await titleLink.focus();
      expect(await hasFocusIndicator(page), 'foco visible en el título-enlace').toBe(true);
      const cta = page.locator(BLOCK).getByRole('link', { name: CTA });
      await cta.focus();
      expect(await hasFocusIndicator(page), 'foco visible en el CTA').toBe(true);

      // Con `Tab` se pasa de un título al siguiente dentro del bloque.
      await titleLink.focus();
      await page.keyboard.press('Tab');
      const next = await page.evaluate(
        () => (document.activeElement as HTMLElement | null)?.innerText.trim() ?? '',
      );
      expect(next).toContain(seed.portada[1].nameEs.slice(0, 20));
    } finally {
      await context.close();
    }
  });
});
