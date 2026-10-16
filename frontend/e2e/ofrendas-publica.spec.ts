import { expect, test, type Browser, type BrowserContext, type Page } from '@playwright/test';
import { createAdminContext } from './publicSeed';
import {
  CANONICAL,
  fixtureByKey,
  resetOfrendas,
  seedOfrendas,
  type OfrendasSeed,
} from './ofrendasSeed';

/**
 * E2E de la sección pública `/ofrendas` (US1/US4/US5, quickstart
 * §1/§4/§5/§7/§9/§10/§11): el visitante **sin sesión** ve el texto introductorio
 * arriba, los dos grupos rotulados en el orden exacto de la sección («Cuentas
 * IBAN» y después «Números de SINPE Móvil») y el texto de uso abajo; los números
 * se muestran **íntegros** (IBAN agrupado en bloques de 4 en pantalla, SINPE en
 * su formato normalizado `+506…`) y el botón de copiar lleva al portapapeles el
 * número **normalizado** exacto (quitar los separadores de la presentación
 * devuelve ese valor, SC-009); la entidad, la **moneda solo en IBAN** y el
 * **titular solo cuando existe** se componen sin líneas vacías (SC-010); los
 * borradores —de método **y de texto**— no aparecen por ninguna vía (SC-002); con
 * **nada publicado** la visita directa a `/ofrendas` **redirige a la portada**
 * sin destello y sin dejar rastro en el historial (`replace`, RG7-12/SC-010); con
 * **solo textos publicados** la sección **sí** existe y el bloque de la portada
 * **no** (P7-13); en inglés el contenido sin traducción cae al español y el
 * número, el tipo, la entidad y la moneda son **idénticos** (SC-006); y a
 * 320/768/1280 px no hay desplazamiento horizontal con el IBAN internacional
 * largo legible (SC-007).
 *
 * Cómo se ejecutan (la base, el rate-limit de login y el navegador son
 * compartidos por toda la suite):
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

/** Nombres accesibles de la sección (`es`). */
const H1 = 'Ofrendas y diezmos';
const IBAN_GROUP = 'Cuentas IBAN';
const SINPE_GROUP = 'Números de SINPE Móvil';
const USAGE_HEADING = '¿En qué usamos las donaciones?';
const NAV = 'Secciones de la portada';
const SITE_MAP = 'Mapa del sitio';

/** El mismo rótulo en inglés (i18n `en`). */
const IBAN_GROUP_EN = 'Bank accounts (IBAN)';
const H1_EN = 'Offerings & tithes';

/** Agrupado de lectura del IBAN en bloques de 4 (mismo invariante de RG7-7). */
function agruparIban(number: string): string {
  return number.replace(/(.{4})/g, '$1 ').trim();
}

/** Abre un contexto de visitante (sin sesión) y devuelve su página. */
async function visitorContext(browser: Browser): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext();
  return { context, page: await context.newPage() };
}

/** Ficha de una tarjeta concreta dentro de un grupo rotulado. */
function card(page: Page, group: string, institution: string) {
  return page
    .getByRole('region', { name: group })
    .getByRole('listitem')
    .filter({ hasText: institution });
}

/** `true` si la página no tiene desplazamiento horizontal. */
async function hasNoHorizontalScroll(page: Page): Promise<boolean> {
  return page.evaluate(
    () =>
      document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1 &&
      document.body.scrollWidth <= document.documentElement.clientWidth + 1,
  );
}

test.describe('Sección pública de ofrendas (grupos, números exactos y redirección)', () => {
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
    // La base es compartida con F3/F4/F6/F10: se deja el módulo **vacío** para no
    // alterar sus portadas (el bloque de ofrendas no debe aparecer en sus
    // aserciones).
    if (adminPage) {
      await resetOfrendas(adminPage);
    }
    await adminContext?.close();
  });

  test('el visitante ve los dos grupos en orden, los textos y los números íntegros (SC-001/SC-013)', async ({
    browser,
  }) => {
    const { context, page } = await visitorContext(browser);
    try {
      await page.goto('/ofrendas');
      await expect(page.getByRole('heading', { level: 1, name: H1 })).toBeVisible();

      // ── Texto introductorio arriba y texto de uso abajo (FR-002/FR-021) ──
      const intro = seed.texts.intro!;
      await expect(page.getByText(intro.textEs)).toBeVisible();
      const usage = seed.texts.usage!;
      await expect(
        page.getByRole('region', { name: USAGE_HEADING }).getByText(usage.textEs),
      ).toBeVisible();

      // ── Dos grupos rotulados, IBAN primero y SINPE después (FR-004) ──────
      const iban = page.getByRole('region', { name: IBAN_GROUP });
      const sinpe = page.getByRole('region', { name: SINPE_GROUP });
      await expect(iban).toBeVisible();
      await expect(sinpe).toBeVisible();
      const ibanBeforeSinpe = await page.evaluate(() => {
        const headings = Array.from(document.querySelectorAll('h2')).map(
          (h) => h.textContent ?? '',
        );
        const i = headings.indexOf('Cuentas IBAN');
        const s = headings.indexOf('Números de SINPE Móvil');
        return i !== -1 && s !== -1 && i < s;
      });
      expect(ibanBeforeSinpe, 'el grupo IBAN va antes que el de SINPE').toBe(true);

      // ── Fichas publicadas exactas en cada grupo, sin borradores (SC-002) ──
      const ibanPublished = seed.published.filter((item) => item.kind === 'iban');
      const sinpePublished = seed.published.filter((item) => item.kind === 'sinpe');
      await expect(iban.getByRole('listitem')).toHaveCount(ibanPublished.length);
      await expect(sinpe.getByRole('listitem')).toHaveCount(sinpePublished.length);
      const html = await page.evaluate(() => document.documentElement.outerHTML);
      for (const draft of seed.drafts) {
        expect(html, `el borrador «${draft.key}» no debe aparecer`).not.toContain(
          draft.institution,
        );
        expect(html).not.toContain(draft.number);
      }

      // ── Número, moneda y titular de cada ficha (FR-003) ──────────────────
      const nacional = fixtureByKey(seed, 'iban-nacional');
      const nacionalCard = card(page, IBAN_GROUP, nacional.institution);
      await expect(nacionalCard.getByText(agruparIban(nacional.number))).toBeVisible();
      // El número íntegro y sin separadores está disponible para lectores.
      await expect(nacionalCard.getByText(nacional.number)).toHaveCount(1);
      await expect(nacionalCard).toContainText('Moneda: colones');
      await expect(nacionalCard).toContainText(`Titular: ${nacional.holderEs}`);

      // IBAN internacional de 24 caracteres: se muestra agrupado sin recortar.
      const internacional = fixtureByKey(seed, 'iban-internacional');
      expect(internacional.number).toHaveLength(24);
      const internacionalCard = card(page, IBAN_GROUP, internacional.institution);
      await expect(internacionalCard.getByText(agruparIban(internacional.number))).toBeVisible();
      await expect(internacionalCard).toContainText('Moneda: dólares');

      // ── Cuenta IBAN sin titular: sin línea de titular (US1 esc. 5) ───────
      const sinTitular = fixtureByKey(seed, 'iban-sin-titular');
      expect(sinTitular.number).toHaveLength(22);
      const sinTitularCard = card(page, IBAN_GROUP, sinTitular.institution);
      await expect(sinTitularCard).toContainText('Moneda: colones');
      await expect(sinTitularCard).not.toContainText('Titular');
      // Cero marcadores de relleno en la ficha.
      await expect(sinTitularCard.getByText('—', { exact: true })).toHaveCount(0);

      // ── SINPE: formato normalizado y **sin** moneda (Q2) ─────────────────
      const sinpeCard = card(
        page,
        SINPE_GROUP,
        fixtureByKey(seed, 'sinpe-con-titular').institution,
      );
      await expect(sinpeCard.getByText(CANONICAL.SINPE_CON_TITULAR)).toBeVisible();
      await expect(sinpeCard).not.toContainText('Moneda');
      const sinpeSinTitular = fixtureByKey(seed, 'sinpe-sin-titular');
      const sinpeSinTitularCard = card(page, SINPE_GROUP, sinpeSinTitular.institution);
      await expect(sinpeSinTitularCard.getByText(sinpeSinTitular.number)).toBeVisible();
      await expect(sinpeSinTitularCard).not.toContainText('Titular');
      await expect(sinpeSinTitularCard).not.toContainText('Moneda');

      // ── SC-013: el contenido de la sección no trae eventos, ministerios ni
      //    grupos (las demás secciones viven en sus rutas; la navegación y el
      //    pie son otra superficie y no cuentan aquí) ────────────────────────
      const mainHtml = await page.evaluate(() => document.querySelector('main')?.outerHTML ?? '');
      for (const forbidden of ['Ministerio', 'Lo que viene', 'Categoría', '/eventos']) {
        expect(mainHtml, `la sección no debe contener «${forbidden}»`).not.toContain(forbidden);
      }
    } finally {
      await context.close();
    }
  });

  test('el botón «Copiar número» copia el valor normalizado exacto (SC-009, RG7-7)', async ({
    browser,
  }) => {
    const context = await browser.newContext();
    await context.grantPermissions(['clipboard-read', 'clipboard-write']);
    const page = await context.newPage();
    try {
      const nacional = fixtureByKey(seed, 'iban-nacional');
      const sinpe = fixtureByKey(seed, 'sinpe-con-titular');

      await page.goto('/ofrendas');

      // ── IBAN: el portapapeles recibe el número normalizado (sin espacios) ─
      const ibanCard = card(page, IBAN_GROUP, nacional.institution);
      const ibanCopy = ibanCard.getByRole('button', {
        name: `Copiar número de ${nacional.institution}`,
      });
      await expect(ibanCopy).toBeVisible();
      await ibanCopy.click();
      await expect(ibanCard.getByRole('status')).toHaveText('Número copiado.');
      expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(nacional.number);

      // Quitar los separadores de la presentación devuelve el valor copiable.
      const presentado = await ibanCard.getByText(agruparIban(nacional.number)).innerText();
      expect(presentado.replace(/\s/g, '')).toBe(nacional.number);

      // ── SINPE: se copia también en su formato normalizado ────────────────
      const sinpeCard = card(page, SINPE_GROUP, sinpe.institution);
      await sinpeCard
        .getByRole('button', { name: `Copiar número de ${sinpe.institution}` })
        .click();
      await expect(sinpeCard.getByRole('status')).toHaveText('Número copiado.');
      expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(sinpe.number);
    } finally {
      await context.close();
    }
  });

  test('inglés con fallback `en → es` y datos idénticos entre idiomas (SC-006)', async ({
    browser,
  }) => {
    const { context, page } = await visitorContext(browser);
    try {
      const nacional = fixtureByKey(seed, 'iban-nacional');
      const dolares = fixtureByKey(seed, 'iban-dolares');
      const intro = seed.texts.intro!;
      const usage = seed.texts.usage!;

      await page.goto('/ofrendas');
      await expect(page.getByRole('region', { name: IBAN_GROUP })).toBeVisible();

      await page.getByRole('banner').getByRole('radio', { name: 'Inglés' }).check();
      await expect(page.getByRole('heading', { level: 1, name: H1_EN })).toBeVisible();

      // Interfaz y contenido con versión inglesa.
      await expect(page.getByRole('region', { name: IBAN_GROUP_EN })).toBeVisible();
      await expect(page.getByText(intro.textEn!)).toBeVisible();
      // Sin inglés: cae al español, idéntico diseño y sin marca de «sin traducir».
      await expect(
        page.getByRole('region', { name: 'How we use the offerings' }).getByText(usage.textEs),
      ).toBeVisible();
      // Titular solo en español: se muestra en español también en inglés.
      await expect(
        card(page, IBAN_GROUP_EN, dolares.institution).getByText(dolares.holderEs!),
      ).toBeVisible();

      // Número, tipo, entidad y moneda **idénticos** (FR-007/SC-006).
      await expect(
        card(page, IBAN_GROUP_EN, nacional.institution).getByText(agruparIban(nacional.number)),
      ).toBeVisible();

      await page.getByRole('banner').getByRole('radio', { name: 'Spanish' }).check();
      await expect(page.getByRole('heading', { level: 1, name: H1 })).toBeVisible();
      await expect(
        card(page, IBAN_GROUP, nacional.institution).getByText(agruparIban(nacional.number)),
      ).toBeVisible();
    } finally {
      await context.close();
    }
  });

  test('320/768/1280 px sin desplazamiento horizontal con el IBAN internacional (SC-007)', async ({
    browser,
  }) => {
    const { context, page } = await visitorContext(browser);
    try {
      const internacional = fixtureByKey(seed, 'iban-internacional');
      for (const width of [320, 768, 1280]) {
        await page.setViewportSize({ width, height: 900 });
        await page.goto('/ofrendas');
        await expect(page.getByRole('region', { name: IBAN_GROUP })).toBeVisible();
        expect(
          await hasNoHorizontalScroll(page),
          `sin desplazamiento horizontal a ${width} px`,
        ).toBe(true);

        // El número más largo no se recorta ni con elipsis ni desbordando.
        const numero = card(page, IBAN_GROUP, internacional.institution)
          .getByText(agruparIban(internacional.number))
          .first();
        const box = await numero.evaluate((el) => {
          const style = getComputedStyle(el);
          return {
            scrollWidth: el.scrollWidth,
            clientWidth: el.clientWidth,
            textOverflow: style.textOverflow,
          };
        });
        expect(box.textOverflow, `sin elipsis a ${width} px`).not.toBe('ellipsis');
        expect(box.scrollWidth, `el número cabe a ${width} px`).toBeLessThanOrEqual(
          box.clientWidth + 1,
        );
      }
    } finally {
      await context.close();
    }
  });

  test('un texto en borrador no deja hueco aunque haya métodos publicados (US3 esc. 7)', async ({
    browser,
  }) => {
    seed = await seedOfrendas(adminPage, { texts: 'usage-draft' });
    const { context, page } = await visitorContext(browser);
    try {
      await page.goto('/ofrendas');
      await expect(page.getByRole('heading', { level: 1, name: H1 })).toBeVisible();
      // El introductorio (publicado) sí; el de uso (borrador) no existe.
      await expect(page.getByText(seed.texts.intro!.textEs)).toBeVisible();
      await expect(page.getByRole('region', { name: USAGE_HEADING })).toHaveCount(0);
      // Los métodos siguen ahí: el grupo no desaparece.
      await expect(page.getByRole('region', { name: IBAN_GROUP })).toBeVisible();
    } finally {
      await context.close();
    }
  });

  test('solo textos publicados: la sección vive y el bloque de la portada no (P7-13)', async ({
    browser,
  }) => {
    seed = await seedOfrendas(adminPage, {
      publishedCount: 0,
      withDrafts: false,
      texts: 'published',
    });
    const { context, page } = await visitorContext(browser);
    try {
      await page.goto('/ofrendas');
      // La sección existe con solo textos: H1, intro y uso…
      await expect(page.getByRole('heading', { level: 1, name: H1 })).toBeVisible();
      await expect(page.getByText(seed.texts.intro!.textEs)).toBeVisible();
      await expect(page.getByRole('region', { name: USAGE_HEADING })).toBeVisible();
      // …y **ningún** grupo: sin métodos no hay rótulo huérfano (SC-010).
      await expect(page.getByRole('region', { name: IBAN_GROUP })).toHaveCount(0);
      await expect(page.getByRole('region', { name: SINPE_GROUP })).toHaveCount(0);

      // La entrada de navegación existe (la sección vive)…
      await expect(
        page.getByRole('navigation', { name: NAV }).getByRole('link', { name: H1 }),
      ).toBeVisible();
      // …pero el bloque de la portada no se dibuja (solo mira métodos, FR-022).
      await page.goto('/');
      await expect(page.locator('#proximamente')).toBeVisible();
      await expect(page.locator('#ofrendas')).toHaveCount(0);
    } finally {
      await context.close();
    }
  });

  test('sin nada publicado: redirección silenciosa a la portada y sin destello (SC-010, P7-21)', async ({
    browser,
  }) => {
    // Estado de partida con contenido publicado (control positivo del detector).
    seed = await seedOfrendas(adminPage);
    const { context, page } = await visitorContext(browser);
    try {
      // Observador instalado antes de navegar: registra si en algún instante se
      // pinta la sección, sus grupos o su H1 (destello).
      await page.addInitScript(() => {
        const store = window as unknown as { __ofrendasFlash?: string[] };
        store.__ofrendasFlash = [];
        const record = () => {
          const markers: string[] = [];
          for (const heading of Array.from(document.querySelectorAll('h1'))) {
            if ((heading.textContent ?? '').trim() === 'Ofrendas y diezmos') {
              markers.push('h1');
            }
          }
          for (const name of ['Cuentas IBAN', 'Números de SINPE Móvil']) {
            for (const section of Array.from(
              document.querySelectorAll('section[aria-labelledby]'),
            )) {
              const label = document.getElementById(section.getAttribute('aria-labelledby') ?? '');
              if ((label?.textContent ?? '').trim() === name) {
                markers.push(`grupo:${name}`);
              }
            }
          }
          if (document.querySelector('button[aria-label^="Copiar número de"]')) {
            markers.push('ficha');
          }
          for (const marker of markers) {
            if (!store.__ofrendasFlash!.includes(marker)) {
              store.__ofrendasFlash!.push(marker);
            }
          }
        };
        record();
        new MutationObserver(record).observe(document, {
          childList: true,
          subtree: true,
          characterData: true,
        });
      });

      const flash = () =>
        page.evaluate(
          () => (window as unknown as { __ofrendasFlash?: string[] }).__ofrendasFlash ?? [],
        );

      // Control positivo: con publicados la misma observación **sí** registra la
      // sección. Sin esto, la aserción de «sin destello» no probaría nada.
      await page.goto('/ofrendas');
      await expect(page.getByRole('region', { name: IBAN_GROUP })).toBeVisible();
      expect(await flash(), 'el detector de destello no es vacuo').toContain('h1');

      // Escenario 0 publicados (SC-010/P7-21).
      await resetOfrendas(adminPage);

      await page.goto('/');
      await expect(page.locator('#proximamente')).toBeVisible();
      const historyBase = await page.evaluate(() => history.length);

      await page.goto('/ofrendas');

      // URL final: la portada, con contenido y navegación operables.
      await expect(page).toHaveURL(/\/$/);
      await expect(page.locator('#proximamente')).toBeVisible();
      await expect(page.getByRole('navigation', { name: NAV })).toBeVisible();

      // Sin destello: nunca se pintó la sección, ni los grupos, ni el H1.
      expect(await flash(), 'la redirección no pinta la sección, los grupos ni el H1').toEqual([]);

      // `replace` (RG7-12): no añade entrada al historial y «Atrás» no vuelve a
      // `/ofrendas` (sin bucles).
      expect(
        await page.evaluate(() => history.length),
        'la redirección usa `replace` (no añade entrada al historial)',
      ).toBe(historyBase + 1);
      await page.goBack();
      await expect(page).toHaveURL(/\/$/);
      expect(page.url()).not.toContain('/ofrendas');

      // La entrada de navegación está oculta en barra y en el mapa del sitio.
      await expect(
        page.getByRole('navigation', { name: NAV }).getByRole('link', { name: H1 }),
      ).toHaveCount(0);
      await expect(
        page.getByRole('navigation', { name: SITE_MAP }).getByRole('link', { name: H1 }),
      ).toHaveCount(0);
      await expect(page.locator('a[href="/ofrendas"]')).toHaveCount(0);
    } finally {
      await context.close();
    }
  });
});
