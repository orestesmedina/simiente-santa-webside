import { expect, test, type Browser, type BrowserContext, type Page } from '@playwright/test';
import { API_URL } from './helpers';
import { createAdminContext } from './publicSeed';
import {
  fixtureByKey,
  resetMinisterios,
  seedMinisterios,
  type MinisterioSeed,
} from './ministeriosSeed';

/**
 * E2E de la sección pública `/ministerios` (US1/US4/US5, quickstart §1/§4/§5/§7/
 * §8/§9/§11): el visitante **sin sesión** ve los ministerios publicados en el
 * orden manual (`sortOrder`, `id`) con su descripción, horario/lugar, imagen y
 * enlace de unión cuando existen; los **borradores no aparecen por ninguna vía**
 * (SC-002, ni sus imágenes); el botón de unirse existe **solo** si hay enlace
 * (SC-009/SC-017) y la línea del encargado **solo** si el servidor publicó su
 * ficha (FR-026); con la ficha oculta y datos reales hay **0 apariciones** del
 * nombre y del contacto en la **respuesta de red** (todas las peticiones, incluida
 * la API forzada sin sesión) y en el **HTML** (SC-014/RG6-1); un alta hecha por
 * la API **sin** `leaderVisible` nace oculta («Oculto por defecto», P6-3); el
 * inglés respeta el fallback `en → es` (SC-006); no hay desplazamiento horizontal
 * a 320/768/1280 px (SC-007); y **sin publicados la visita directa a
 * `/ministerios` redirige a la portada** —`replace`, sin destello de la grilla
 * vacía, del H1 ni de ningún aviso, y con la entrada de navegación oculta—
 * (SC-010/P6-21/RG6-13).
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
 * `--workers=1` es obligatorio en las cuatro (singletons compartidos y ministerios
 * sembrados por spec). Nota: el filtro `portada-` de la tanda 3 también incluye
 * `portada-ministerios.spec.ts`; su tanda natural es la 4, pero ejecutarla dentro
 * de la 3 no rompe nada (se siembra a sí misma y usa el contexto de
 * administrador compartido). En este entorno el navegador no arranca desde el
 * host —faltan las librerías del sistema (`libnspr4`)—, así que los comandos se
 * ejecutan dentro de la imagen oficial de Playwright, con red del host para
 * llegar al stack de `make up`:
 *
 *   docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
 *     -v "$PWD/..":/work -w /work/frontend mcr.microsoft.com/playwright:v1.63.0-jammy \
 *     npx playwright test --config e2e/playwright.config.ts ministerios- portada-ministerios --workers=1
 *
 * (El stack debe estar reconstruido con F6: `docker compose up -d --build`.)
 */
test.describe.configure({ mode: 'serial' });

/** Nombre accesible de la sección y de la lista de fichas (i18n `es`). */
const SECTION = 'Ministerios de la iglesia';
/** El mismo nombre en inglés (i18n `en`). */
const SECTION_EN = 'Church ministries';

/** Cadenas que **jamás** deben salir del servidor con la ficha del encargado oculta. */
function hiddenStrings(seed: MinisterioSeed): string[] {
  const sonido = fixtureByKey(seed, 'sonido');
  const misiones = fixtureByKey(seed, 'misiones');
  return [
    sonido.leaderNameEs ?? '',
    sonido.leaderPhone ?? '',
    (sonido.leaderPhone ?? '').replace(/[^\d]/g, ''),
    misiones.leaderNameEs ?? '',
    misiones.leaderPhone ?? '',
    (misiones.leaderPhone ?? '').replace(/[^\d]/g, ''),
  ].filter((value) => value !== '');
}

/** Ficha de una tarjeta concreta del listado de `/ministerios`. */
function card(page: Page, name: string, section: string = SECTION) {
  return page
    .getByRole('list', { name: section, exact: true })
    .getByRole('listitem')
    .filter({ hasText: name });
}

/** Abre un contexto de visitante (sin sesión) y devuelve su página. */
async function visitorContext(browser: Browser): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext();
  return { context, page: await context.newPage() };
}

/** `true` si la página no tiene desplazamiento horizontal. */
async function hasNoHorizontalScroll(page: Page): Promise<boolean> {
  return page.evaluate(
    () =>
      document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1 &&
      document.body.scrollWidth <= document.documentElement.clientWidth + 1,
  );
}

test.describe('Sección pública de ministerios (visitante, privacidad y redirección)', () => {
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
    // La base es compartida con F3/F4/F10: se deja el módulo **vacío** para no
    // alterar sus portadas (el bloque de ministerios no debe aparecer en sus
    // aserciones).
    if (adminPage) {
      await resetMinisterios(adminPage);
    }
    await adminContext?.close();
  });

  test('el visitante ve los publicados en orden, con sus datos, y ningún borrador', async ({
    browser,
  }) => {
    const { context, page } = await visitorContext(browser);
    try {
      await page.goto('/ministerios');
      await expect(page.getByRole('heading', { level: 1, name: 'Ministerios' })).toBeVisible();

      const grid = page.getByRole('list', { name: SECTION, exact: true });
      const cards = grid.getByRole('listitem');

      // ── Orden manual (sortOrder, id) y solo publicados (SC-002) ───────────
      await expect(cards).toHaveCount(seed.published.length);
      expect(await cards.locator('h3').allInnerTexts()).toEqual(
        seed.published.map((fixture) => fixture.nameEs),
      );
      for (const draft of seed.drafts) {
        await expect(page.getByText(draft.nameEs)).toHaveCount(0);
      }

      // ── Ficha completa: imagen+alt, descripción, horario, lugar, encargado
      //    visible y botón de unión con el `href` guardado (FR-003/FR-015) ────
      const alabanza = fixtureByKey(seed, 'alabanza');
      const alabanzaCard = card(page, alabanza.nameEs);
      await expect(alabanzaCard).toHaveCount(1);
      await expect(alabanzaCard.getByRole('img')).toHaveAttribute('alt', alabanza.imageAltEs!);
      await expect(alabanzaCard).toContainText(alabanza.descriptionEs);
      await expect(alabanzaCard).toContainText(alabanza.scheduleEs!);
      await expect(alabanzaCard).toContainText(alabanza.placeEs!);
      await expect(alabanzaCard).toContainText(alabanza.leaderNameEs!);
      await expect(alabanzaCard).toContainText(alabanza.leaderPhone!);

      const join = alabanzaCard.getByRole('link', {
        name: `Unirse al grupo de WhatsApp de ${alabanza.nameEs}`,
      });
      await expect(join).toBeVisible();
      // El enlace se sirve **tal cual** se guardó (FR-015): sin transformar.
      await expect(join).toHaveAttribute('href', alabanza.whatsappUrl!);
      await expect(join).toHaveAttribute('target', '_blank');
      await expect(join).toHaveAttribute('rel', /noopener/);

      // ── Sin enlace no hay botón (SC-009) y sin encargado publicado no hay
      //    línea de encargado (FR-026/SC-017) ────────────────────────────────
      const sonido = fixtureByKey(seed, 'sonido');
      const sonidoCard = card(page, sonido.nameEs);
      await expect(sonidoCard).toHaveCount(1);
      await expect(sonidoCard.getByRole('link', { name: /Unirse al grupo/ })).toHaveCount(0);
      await expect(sonidoCard.getByText(sonido.leaderNameEs!)).toHaveCount(0);
      // Sin imagen no hay hueco ni marcador: la tarjeta no trae ningún `img`.
      await expect(sonidoCard.getByRole('img')).toHaveCount(0);
      await expect(sonidoCard.getByRole('link', { name: /Unirse al grupo/ })).toHaveCount(0);

      // ── Solo nombre y descripción: la ficha queda completa y coherente
      //    (SC-017) ───────────────────────────────────────────────────────────
      const matrimonios = fixtureByKey(seed, 'matrimonios');
      const matrimoniosCard = card(page, matrimonios.nameEs);
      await expect(matrimoniosCard).toContainText(matrimonios.descriptionEs);
      await expect(matrimoniosCard.getByRole('img')).toHaveCount(0);
      await expect(matrimoniosCard.getByRole('link')).toHaveCount(0);

      // ── Horario de texto libre largo, servido sin recalcular (SC-015) ─────
      const accionSocial = fixtureByKey(seed, 'accion-social');
      await expect(card(page, accionSocial.nameEs)).toContainText(accionSocial.scheduleEs!);

      // ── Encargado visible con **solo** correo (Q8) ────────────────────────
      const jovenes = fixtureByKey(seed, 'jovenes');
      const jovenesCard = card(page, jovenes.nameEs);
      await expect(jovenesCard).toContainText(jovenes.leaderNameEs!);
      await expect(jovenesCard).toContainText(jovenes.leaderEmail!);

      // ── Tildes, `ñ`, `¿¡` y emojis intactos (edge case) ───────────────────
      const ninos = fixtureByKey(seed, 'ninos');
      const ninosCard = card(page, ninos.nameEs);
      await expect(ninosCard).toContainText(ninos.descriptionEs);

      // ── SC-002 por la vía de las imágenes: la de un borrador no se sirve ──
      expect(seed.draftImageFile).not.toBe('');
      expect(
        (await page.request.get(`${API_URL}/api/v1/media/${seed.draftImageFile}`)).status(),
        'la imagen de un ministerio en borrador no se sirve',
      ).toBe(404);
      expect(
        (await page.request.get(`${API_URL}/api/v1/media/${alabanza.imageFile!}`)).status(),
        'la imagen de un ministerio publicado sí se sirve',
      ).toBe(200);
    } finally {
      await context.close();
    }
  });

  test('privacidad del encargado: 0 apariciones en la red y en el HTML (SC-014, P6-3)', async ({
    browser,
    request,
  }) => {
    const sonido = fixtureByKey(seed, 'sonido');
    const misiones = fixtureByKey(seed, 'misiones');

    // ── P6-3: el alta sin la clave `leaderVisible` nace **oculta** ─────────
    expect(sonido.leaderVisible).toBe(false);
    expect(misiones.leaderVisible, 'el alta sin `leaderVisible` nace oculta').toBe(false);
    expect(misiones.createdWithoutLeaderVisibleKey).toBe(true);
    // …y el panel **sí** conserva los datos (dato guardado ≠ dato publicado).
    const adminList = (await (
      await adminPage.request.get(`${API_URL}/api/v1/admin/ministerios?limit=100`)
    ).json()) as { items: { id: string; leaderNameEs: string | null; leaderVisible: boolean }[] };
    const stored = adminList.items.find((item) => item.id === misiones.id);
    expect(stored?.leaderNameEs).toBe(misiones.leaderNameEs);
    expect(stored?.leaderVisible).toBe(false);

    // ── Todas las respuestas de la API durante la visita ──────────────────
    const responses: Promise<string>[] = [];
    const { context, page } = await visitorContext(browser);
    try {
      page.on('response', (response) => {
        if (response.url().includes('/api/v1/')) {
          responses.push(response.text().catch(() => ''));
        }
      });

      await page.goto('/ministerios');
      await expect(page.getByRole('list', { name: SECTION, exact: true })).toBeVisible();

      // 0 líneas de encargado en las fichas cuyo encargado no está publicado.
      const cards = page.getByRole('list', { name: SECTION, exact: true }).getByRole('listitem');
      await expect(cards).toHaveCount(seed.published.length);
      for (const fixture of [sonido, misiones]) {
        await expect(card(page, fixture.nameEs).getByText('Encargado')).toHaveCount(0);
      }
      const htmlMinistries = await page.evaluate(() => document.documentElement.outerHTML);

      // La portada comparte el documento del bloque de ministerios: su HTML
      // tampoco puede llevar el dato oculto (FR-026).
      await page.goto('/');
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
      const htmlHome = await page.evaluate(() => document.documentElement.outerHTML);

      const network = (await Promise.all(responses)).join('\n');

      // ── API forzada **sin sesión** (contexto `request` sin cookies) ───────
      const forced: string[] = [];
      for (const path of [
        '/api/v1/ministerios?lang=es',
        '/api/v1/ministerios?lang=en',
        '/api/v1/portada/ministerios?lang=es',
        '/api/v1/portada/ministerios?lang=en',
      ]) {
        const response = await request.get(`${API_URL}${path}`);
        expect(response.status(), `GET ${path} sin sesión`).toBe(200);
        forced.push(await response.text());
      }
      // La API forzada sirve la sección (hay publicados), pero sin el encargado.
      expect(forced.join('\n')).toContain(sonido.nameEs);

      const forbidden = hiddenStrings(seed);
      expect(forbidden.length, 'hay datos del encargado oculto que vigilar').toBeGreaterThan(0);
      for (const value of forbidden) {
        expect(network, `el tráfico de red no debe contener «${value}»`).not.toContain(value);
        expect(htmlMinistries, `el HTML de /ministerios no debe contener «${value}»`).not.toContain(
          value,
        );
        expect(htmlHome, `el HTML de la portada no debe contener «${value}»`).not.toContain(value);
        expect(forced.join('\n'), `la API forzada no debe contener «${value}»`).not.toContain(
          value,
        );
      }

      // Y el borrador tampoco aparece por la vía del HTML.
      for (const draft of seed.drafts) {
        expect(htmlMinistries).not.toContain(draft.nameEs);
        expect(htmlHome).not.toContain(draft.nameEs);
      }
    } finally {
      await context.close();
    }
  });

  test('inglés con fallback `en → es` y vuelta a español (SC-006)', async ({ browser }) => {
    const { context, page } = await visitorContext(browser);
    try {
      await page.goto('/ministerios');
      // El selector de idioma vive en la cabecera (y también en el pie): se
      // ancla al `banner` para no caer en una violación de modo estricto.
      await page.getByRole('banner').getByRole('radio', { name: 'Inglés' }).check();

      await expect(page.getByRole('heading', { level: 1, name: 'Ministries' })).toBeVisible();

      const alabanza = fixtureByKey(seed, 'alabanza');
      const alabanzaCard = card(page, alabanza.nameEn!, SECTION_EN);
      await expect(alabanzaCard).toContainText(alabanza.descriptionEn!);
      await expect(
        alabanzaCard.getByRole('link', { name: `Join the WhatsApp group of ${alabanza.nameEn}` }),
      ).toBeVisible();

      // Sin traducción: el campo cae al español (nunca queda vacío).
      const sonido = fixtureByKey(seed, 'sonido');
      await expect(card(page, sonido.nameEs, SECTION_EN)).toHaveCount(1);
      await expect(card(page, sonido.nameEs, SECTION_EN)).toContainText(sonido.descriptionEs);

      await page.getByRole('banner').getByRole('radio', { name: 'Spanish' }).check();
      await expect(page.getByRole('heading', { level: 1, name: 'Ministerios' })).toBeVisible();
      await expect(card(page, alabanza.nameEs)).toHaveCount(1);
    } finally {
      await context.close();
    }
  });

  test('320/768/1280 px sin desplazamiento horizontal (SC-007)', async ({ browser }) => {
    const { context, page } = await visitorContext(browser);
    try {
      for (const width of [320, 768, 1280]) {
        await page.setViewportSize({ width, height: 900 });
        await page.goto('/ministerios');
        await expect(page.getByRole('heading', { level: 1, name: 'Ministerios' })).toBeVisible();
        await expect(page.getByRole('list', { name: SECTION, exact: true })).toBeVisible();
        expect(
          await hasNoHorizontalScroll(page),
          `sin desplazamiento horizontal a ${width} px`,
        ).toBe(true);
      }
    } finally {
      await context.close();
    }
  });

  test('sin publicados: redirección silenciosa a la portada y entrada oculta (SC-010)', async ({
    browser,
  }) => {
    const { context, page } = await visitorContext(browser);
    try {
      // Observador instalado **antes** de navegar: registra si en algún instante
      // se pinta la sección, la grilla o el H1 de ministerios (destello).
      await page.addInitScript(() => {
        const store = window as unknown as { __ministriesFlash?: string[] };
        store.__ministriesFlash = [];
        const record = () => {
          const markers: string[] = [];
          if (document.querySelector('section[aria-label="Ministerios de la iglesia"]')) {
            markers.push('seccion');
          }
          if (document.querySelector('ul[aria-label="Ministerios de la iglesia"]')) {
            markers.push('grilla');
          }
          for (const heading of Array.from(document.querySelectorAll('h1'))) {
            if ((heading.textContent ?? '').trim() === 'Ministerios') {
              markers.push('h1');
            }
          }
          for (const marker of markers) {
            if (!store.__ministriesFlash!.includes(marker)) {
              store.__ministriesFlash!.push(marker);
            }
          }
        };
        record();
        // El `initScript` corre **antes** de que exista `<html>`: se observa el
        // propio `document` (cubre todo el árbol) en vez de `documentElement`.
        new MutationObserver(record).observe(document, {
          childList: true,
          subtree: true,
          characterData: true,
        });
      });

      const flash = () =>
        page.evaluate(
          () => (window as unknown as { __ministriesFlash?: string[] }).__ministriesFlash ?? [],
        );

      // Control positivo: con publicados, la misma observación **sí** registra la
      // sección. Sin esto, la aserción de «sin destello» no probaría nada.
      await page.goto('/ministerios');
      await expect(page.getByRole('list', { name: SECTION, exact: true })).toBeVisible();
      expect(await flash(), 'el detector de destello no es vacuo').toContain('seccion');

      // Escenario 0 publicados (SC-010/P6-21).
      await resetMinisterios(adminPage);

      await page.goto('/');
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
      const historyBase = await page.evaluate(() => history.length);

      await page.goto('/ministerios');

      // URL final: la portada. La portada carga completa y navegable.
      await expect(page).toHaveURL(/\/$/);
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
      await expect(page.locator('#proximamente')).toBeVisible();
      await expect(page.getByRole('navigation', { name: 'Secciones de la portada' })).toBeVisible();

      // Sin destello: nunca se pintó la sección, ni la grilla vacía, ni el H1.
      expect(
        await flash(),
        'la redirección no pinta ni la sección, ni la grilla, ni el H1',
      ).toEqual([]);

      // `replace` (RG6-13): la redirección no añade una entrada al historial y
      // «Atrás» no devuelve a `/ministerios`.
      expect(
        await page.evaluate(() => history.length),
        'la redirección usa `replace` (no añade entrada al historial)',
      ).toBe(historyBase + 1);
      await page.goBack();
      await expect(page).toHaveURL(/\/$/);
      expect(page.url()).not.toContain('/ministerios');

      // La entrada de navegación sigue oculta en barra, menú y pie.
      await expect(
        page
          .getByRole('navigation', { name: 'Secciones de la portada' })
          .getByRole('link', { name: 'Ministerios' }),
      ).toHaveCount(0);
      await expect(
        page.getByRole('navigation', { name: 'Mapa del sitio' }).getByRole('link', {
          name: 'Ministerios',
        }),
      ).toHaveCount(0);
      await expect(page.locator('a[href="/ministerios"]')).toHaveCount(0);
    } finally {
      await context.close();
    }
  });
});
