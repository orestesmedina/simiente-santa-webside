import { expect, test, type Page } from '@playwright/test';
import { API_URL } from './helpers';
import { createAdminContext } from './publicSeed';
import { resetAgenda, seedAgenda, type AgendaSeed } from './agendaSeed';

/**
 * E2e de la sección pública `/eventos` (US1/US5/US6, quickstart §1/§8/§9/§10):
 * visitante sin cuenta que ve **lo que viene** —dos bloques, orden cronológico,
 * categoría, detalle e imagen—, que un elemento vencido o en borrador **no**
 * aparece por ninguna vía (SC-002/SC-009), que el idioma cambia interfaz y
 * formato de fecha sin alterar el dato (SC-006), que no hay desplazamiento
 * horizontal a 320/768/1280 px (SC-007) y que sin nada publicado y próximo la
 * sección **no se enlaza** y `/eventos` muestra el aviso **sin redirigir**
 * (SC-012/FR-013/⟲ D-13).
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

/** Mes en español de una fecha `YYYY-MM-DD`, para la aserción de la fecha escrita. */
function monthInSpanish(isoDate: string): string {
  const [year, month, day] = isoDate.split('-').map(Number);
  return new Intl.DateTimeFormat('es', { month: 'long', timeZone: 'UTC' }).format(
    new Date(Date.UTC(year, month - 1, day)),
  );
}

/** Mes en inglés de una fecha `YYYY-MM-DD` (formato del locale `en`). */
function monthInEnglish(isoDate: string): string {
  const [year, month, day] = isoDate.split('-').map(Number);
  return new Intl.DateTimeFormat('en', { month: 'long', timeZone: 'UTC' }).format(
    new Date(Date.UTC(year, month - 1, day)),
  );
}

/** Desplazamiento horizontal del documento (SC-007). */
async function horizontalOverflow(page: Page): Promise<number> {
  return page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
}

test.describe('Sección pública /eventos — lo que viene, idioma y aviso (US1, US5, US6)', () => {
  test.setTimeout(240_000);

  test('visitante: lo próximo, nada de borradores ni vencidos, inglés, tres anchos y aviso sin redirigir', async ({
    browser,
    request,
  }) => {
    // ── Siembra (con la sesión de administrador cacheada) ──────────────────
    const adminContext = await createAdminContext(browser);
    const adminPage = await adminContext.newPage();
    let seed: AgendaSeed;
    try {
      await resetAgenda(adminPage);
      seed = await seedAgenda(adminPage);
    } finally {
      await adminContext.close();
    }

    const eventsNav = (page: Page) =>
      page.getByRole('navigation', { name: 'Secciones de la portada' });
    const sitemap = (page: Page) => page.getByRole('navigation', { name: 'Mapa del sitio' });

    // ── 1. Visitante sin sesión: dos bloques con lo próximo (US1) ──────────
    const visitor = await browser.newContext();
    const page = await visitor.newPage();
    try {
      await page.goto('/eventos');

      // H1 + subtítulo + los dos bloques, en el orden de la página (FR-021/Q5).
      await expect(
        page.getByRole('heading', { level: 1, name: 'Eventos y actividades' }),
      ).toBeVisible();
      await expect(
        page.getByText(
          'Lo que viene: reuniones de la iglesia y acciones concretas en las que puedes participar.',
        ),
      ).toBeVisible();
      const headings = await page.evaluate(() =>
        [...document.querySelectorAll('h2')].map((heading) => heading.textContent?.trim()),
      );
      expect(headings).toEqual(['Próximos eventos', 'Próximas actividades']);
      expect(await page.locator('h1').count()).toBe(1);

      // El evento publicado y próximo: próxima fecha escrita con su `<time>` real.
      const eventCard = page
        .locator('article')
        .filter({ has: page.getByRole('heading', { level: 3, name: seed.eventTitle }) });
      await expect(eventCard).toBeVisible();
      const nextTime = eventCard.locator(`time[datetime="${seed.eventNextDate}T19:00"]`);
      await expect(nextTime).toBeVisible();
      await expect(nextTime).toHaveText(
        new RegExp(
          `${Number(seed.eventNextDate.split('-')[2])} de ${monthInSpanish(seed.eventNextDate)}`,
        ),
      );
      await expect(eventCard.getByText('7:00 p. m. − 9:00 p. m.')).toBeVisible();

      // Frecuencia, fechas restantes (⟲ D-11), lugar, mapa, categoría y detalle.
      await expect(eventCard.getByText(seed.eventFrequency)).toBeVisible();
      await expect(eventCard.getByText('Más fechas')).toBeVisible();
      await expect(
        eventCard.locator(`time[datetime="${seed.eventLaterDate}T19:00"]`),
      ).toBeVisible();
      await expect(eventCard.getByText(seed.eventPlace)).toBeVisible();
      const mapLink = eventCard.getByRole('link', {
        name: `Ver ubicación de "${seed.eventTitle}"`,
      });
      await expect(mapLink).toHaveAttribute('href', seed.eventMapUrl);
      await expect(mapLink).toHaveAttribute('target', '_blank');
      await expect(mapLink).toHaveAttribute('rel', /noopener/);
      await expect(eventCard.getByLabel(`Categoría: ${seed.categoryName}`)).toBeVisible();

      const detail = eventCard.getByText(/Trae tu Biblia/);
      await expect(detail).toBeVisible();
      expect(
        await detail.evaluate((element) => ({
          text: element.textContent,
          whiteSpace: getComputedStyle(element).whiteSpace,
        })),
        'el detalle conserva los saltos de línea (FR-018)',
      ).toEqual({ text: seed.eventDetail, whiteSpace: 'pre-line' });

      // Imagen con el `alt` del equipo (FR-020) servida por la API.
      const image = eventCard.getByRole('img', { name: seed.eventImageAlt });
      await expect(image).toBeVisible();
      await expect(image).toHaveAttribute(
        'src',
        new RegExp(`/api/v1/media/${seed.eventImageFile}$`),
      );
      expect((await request.get(`${API_URL}/api/v1/media/${seed.eventImageFile}`)).status()).toBe(
        200,
      );

      // La actividad publicada y próxima, con su fecha única y su horario.
      const activityCard = page
        .locator('article')
        .filter({ has: page.getByRole('heading', { level: 3, name: seed.activityTitle }) });
      await expect(activityCard).toBeVisible();
      await expect(
        activityCard.locator(`time[datetime="${seed.activityDate}T08:00"]`),
      ).toBeVisible();
      await expect(activityCard.getByText('8:00 a. m. − 11:00 a. m.')).toBeVisible();
      await expect(activityCard.getByText(seed.activityPlace)).toBeVisible();
      await expect(activityCard.getByRole('img', { name: seed.activityImageAlt })).toBeVisible();

      // SC-002/SC-009: nada vencido, nada en borrador, por ninguna vía.
      await expect(page.getByText(seed.pastEventTitle)).toHaveCount(0);
      await expect(page.getByText(seed.draftEventTitle)).toHaveCount(0);
      await expect(page.getByText(seed.todayFinishedTitle)).toHaveCount(0);
      await expect(page.getByText(seed.pastActivityTitle)).toHaveCount(0);
      expect(
        (await request.get(`${API_URL}/api/v1/media/${seed.draftImageFile}`)).status(),
        'la imagen de un borrador no se sirve (SC-002)',
      ).toBe(404);

      // FR-021/FR-022: el ítem de la sección está en la barra y en el pie.
      await expect(
        eventsNav(page).getByRole('link', { name: 'Eventos y actividades' }),
      ).toHaveAttribute('href', '/eventos');
      await expect(
        sitemap(page).getByRole('link', { name: 'Eventos y actividades' }),
      ).toHaveAttribute('href', '/eventos');
    } finally {
      await visitor.close();
    }

    // ── 2. US5: cambio a inglés (interfaz, contenido y fechas) ─────────────
    const englishContext = await browser.newContext();
    const englishPage = await englishContext.newPage();
    try {
      await englishPage.goto('/eventos');
      await englishPage.locator('header').getByRole('radio', { name: 'Inglés' }).check();

      await expect(englishPage.locator('html')).toHaveAttribute('lang', 'en');
      await expect(
        englishPage.getByRole('heading', { level: 2, name: 'Upcoming events' }),
      ).toBeVisible();
      await expect(
        englishPage.getByRole('heading', { level: 2, name: 'Upcoming activities' }),
      ).toBeVisible();
      await expect(
        englishPage.getByRole('navigation', { name: 'Page sections' }).getByRole('link', {
          name: 'Events & activities',
        }),
      ).toBeVisible();

      // Contenido con inglés → inglés; sin inglés → español (fallback del servidor).
      await expect(
        englishPage.getByRole('heading', { level: 3, name: seed.eventTitleEn }),
      ).toBeVisible();
      await expect(
        englishPage.getByRole('heading', { level: 3, name: seed.activityTitle }),
      ).toBeVisible();

      // La fecha se reescribe en inglés sin cambiar el día ni la hora reales.
      const englishNext = englishPage.locator(`time[datetime="${seed.eventNextDate}T19:00"]`);
      await expect(englishNext).toBeVisible();
      await expect(englishNext).toHaveText(
        new RegExp(
          `${monthInEnglish(seed.eventNextDate)} ${Number(seed.eventNextDate.split('-')[2])}`,
        ),
      );
      await expect(englishPage.getByText('7:00 PM − 9:00 PM')).toBeVisible();

      // Y vuelve al español sin perder nada.
      await englishPage.locator('header').getByRole('radio', { name: 'Spanish' }).check();
      await expect(englishPage.locator('html')).toHaveAttribute('lang', 'es');
      await expect(
        englishPage.getByRole('heading', { level: 1, name: 'Eventos y actividades' }),
      ).toBeVisible();
    } finally {
      await englishContext.close();
    }

    // ── 3. SC-007: tres anchos sin desplazamiento horizontal ───────────────
    const widthContext = await browser.newContext();
    const widthPage = await widthContext.newPage();
    try {
      for (const width of [320, 768, 1280]) {
        await widthPage.setViewportSize({ width, height: 900 });
        await widthPage.goto('/eventos');
        await expect(
          widthPage.getByRole('heading', { level: 1, name: 'Eventos y actividades' }),
        ).toBeVisible();
        await expect(
          widthPage.getByRole('heading', { level: 3, name: seed.eventTitle }),
        ).toBeVisible();
        expect(
          await horizontalOverflow(widthPage),
          `sin scroll horizontal a ${width} px`,
        ).toBeLessThanOrEqual(1);
      }
    } finally {
      await widthContext.close();
    }

    // ── 4. SC-012/FR-013/⟲ D-13: sin nada publicado y próximo ──────────────
    const cleanupContext = await createAdminContext(browser);
    const cleanupPage = await cleanupContext.newPage();
    try {
      await resetAgenda(cleanupPage);
    } finally {
      await cleanupContext.close();
    }

    const hiddenContext = await browser.newContext();
    const hiddenPage = await hiddenContext.newPage();
    try {
      await hiddenPage.goto('/eventos');

      // La página existe con su aviso y **no** redirige a la portada.
      await expect(hiddenPage.getByText('Por ahora no hay eventos ni actividades')).toBeVisible();
      await expect(hiddenPage).toHaveURL(/\/eventos$/);
      await expect(
        hiddenPage.getByRole('heading', { level: 1, name: 'Eventos y actividades' }),
      ).toBeVisible();
      await expect(hiddenPage.locator('h2')).toHaveCount(0);
      await expect(hiddenPage.getByRole('heading', { level: 3 })).toHaveCount(0);

      // Sin contenido, el ítem no existe en ninguna superficie (barra, menú
      // móvil —la misma lista— y mapa del sitio del pie): cero enlaces rotos.
      await expect(hiddenPage.getByRole('link', { name: 'Eventos y actividades' })).toHaveCount(0);

      // Y la portada sigue coherente: sin ítem de agenda y sin huérfanos.
      await hiddenPage.goto('/');
      await expect(
        hiddenPage.getByRole('navigation', { name: 'Secciones de la portada' }),
      ).toBeVisible();
      await expect(hiddenPage.getByRole('link', { name: 'Eventos y actividades' })).toHaveCount(0);
    } finally {
      await hiddenContext.close();
    }
  });
});
