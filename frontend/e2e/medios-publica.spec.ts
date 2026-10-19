import { expect, test, type Browser, type BrowserContext, type Page } from '@playwright/test';
import { createAdminContext } from './publicSeed';
import { resetMedios, seedMedios, type MediosSeed } from './mediosSeed';

/**
 * E2E de la sección pública `/medios` (US1/US4/US5, quickstart §1/§4/§6/§7/§9/§11):
 * el visitante **sin sesión** ve las fichas publicables —carátula **servida por
 * el sitio**, nombre del feed, fecha y descripción corta— del más reciente al más
 * antiguo; las fichas con dos enlaces y con uno solo se componen **sin líneas
 * vacías** (y sin descripción, **sin esa línea**); los episodios **ocultos, en un
 * par abierto o absorbidos no aparecen por ninguna vía** (SC-002) mientras que
 * «ya no aparece en el feed» **sí** se publica (FR-017); **0 peticiones a
 * terceros** y 0 `iframe`/`src` de YouTube, Spotify ni Anchor (SC-005); cada
 * enlace externo **anuncia el servicio, la salida del sitio y la pestaña nueva**
 * y **abre en pestaña nueva** (SC-006); con **0 publicables** la entrada de
 * navegación desaparece y `/medios` **redirige a la portada sin destello**
 * (SC-016); en inglés el contenido sin traducción cae al español y el nombre, la
 * fecha y los enlaces son **idénticos** (SC-011); a 320/768/1280 px no hay
 * desplazamiento horizontal (SC-017) y la sección solo muestra prédicas y
 * podcasts (SC-013).
 *
 * Cómo se ejecutan (la base, el rate-limit de login y el navegador son
 * compartidos por toda la suite):
 *
 *   # Tanda 6 — medios (F9): las cuatro specs comparten la siembra/limpieza del
 *   # módulo (y la carátula subida), así que van juntas y con un solo worker.
 *   cd frontend && npx playwright test --config e2e/playwright.config.ts medios- portada-medios --workers=1
 *
 * La suite completa se reparte en **seis tandas** porque `POST
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
 *   # Tanda 6 — F9 (medios)
 *   npx playwright test --config e2e/playwright.config.ts medios- portada-medios --workers=1
 *
 * `--workers=1` es obligatorio (contenido sembrado por spec y **una sola**
 * carátula subida por corrida). La siembra es **por SQL** en el contenedor `db`
 * (patrón RG9-7, ver `mediosSeed.ts`), así que estos e2e necesitan `docker
 * compose` disponible en la máquina que los ejecuta.
 */
test.describe.configure({ mode: 'serial' });

/** Nombres accesibles de la sección (`es`). */
const H1 = 'Medios';
const SUBTITLE =
  'Las prédicas y los podcasts de nuestra iglesia; se abren en YouTube y Spotify, en una pestaña nueva.';
const LIST = 'Lista de episodios de medios';
const CARGAR_MAS = 'Cargar más';
const NAV = 'Secciones de la portada';
const SITE_MAP = 'Mapa del sitio';
const NEW_TAB = '(pestaña nueva)';

/** El mismo rótulo en inglés (i18n `en`). */
const H1_EN = 'Media';
const LIST_EN = 'List of media episodes';

/** Dominios de terceros que la sección **nunca** debe pedir (SC-005/FR-004). */
const TERCEROS = [
  'youtube.com',
  'youtu.be',
  'ytimg.com',
  'spotify.com',
  'scdn.co',
  'anchor.fm',
  'cloudfront.net',
];

/**
 * Fecha larga tal como la pinta la vista: el **día de la iglesia**
 * (`America/Costa_Rica`, como resuelve `episodeDay`) escrito con la regla de
 * `formatLongDate` de F4 («lunes 5 de octubre» / «Monday, October 5»). Se
 * reproduce la composición del componente (`weekday day de month`) porque el
 * `Intl` del navegador añadiría la coma que la vista no usa.
 */
function fechaLarga(iso: string, lang: 'es' | 'en'): string {
  const day = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'America/Costa_Rica',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(new Date(iso));
  const [year, month, date] = day.split('-').map(Number);
  const pieces: Partial<Record<'weekday' | 'day' | 'month', string>> = {};
  for (const part of new Intl.DateTimeFormat(lang, {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    timeZone: 'UTC',
  }).formatToParts(new Date(Date.UTC(year, month - 1, date)))) {
    if (part.type === 'weekday' || part.type === 'day' || part.type === 'month') {
      pieces[part.type] = part.value;
    }
  }
  const { weekday, day: dayNumber, month: monthName } = pieces;
  return lang === 'en'
    ? `${weekday}, ${monthName} ${dayNumber}`
    : `${weekday} ${dayNumber} de ${monthName}`;
}

/** Abre un contexto de visitante (sin sesión) y devuelve su página. */
async function visitorContext(browser: Browser): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext();
  return { context, page: await context.newPage() };
}

/** Ficha de la sección por el nombre del episodio (lista rotulada en el idioma activo). */
function card(page: Page, title: string, listName: string = LIST) {
  return page
    .getByRole('list', { name: listName })
    .getByRole('listitem')
    .filter({ hasText: title });
}

/** `true` si el documento no tiene desplazamiento horizontal. */
async function hasNoHorizontalScroll(page: Page): Promise<boolean> {
  return page.evaluate(
    () =>
      document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1 &&
      document.body.scrollWidth <= document.documentElement.clientWidth + 1,
  );
}

test.describe('Sección pública de medios (fichas, privacidad, enlaces y redirección)', () => {
  test.setTimeout(300_000);

  let adminContext: BrowserContext;
  let adminPage: Page;
  let seed: MediosSeed;

  test.beforeAll(async ({ browser }) => {
    adminContext = await createAdminContext(browser);
    adminPage = await adminContext.newPage();
    seed = await seedMedios(adminPage);
  });

  test.afterAll(async () => {
    // Se borra lo sembrado y se devuelven a `visible` las fichas reales que la
    // siembra ocultó temporalmente (ver `mediosSeed.ts`).
    resetMedios();
    await adminContext?.close();
  });

  test('las fichas publicables con carátula, nombre, fecha y descripción, en orden (SC-001/SC-002/SC-013)', async ({
    browser,
  }) => {
    const { context, page } = await visitorContext(browser);
    try {
      await page.goto('/medios');
      await expect(page.getByRole('heading', { level: 1, name: H1 })).toBeVisible();
      await expect(page.getByText(SUBTITLE)).toBeVisible();

      // ── Una sola lista con los publicables, del más reciente al más antiguo ──
      const list = page.getByRole('list', { name: LIST });
      await expect(list.getByRole('listitem')).toHaveCount(seed.publicables.length);
      expect(await list.locator('h3').allInnerTexts()).toEqual(
        seed.publicables.map((item) => item.title),
      );

      for (const item of seed.publicables) {
        const ficha = card(page, item.title);
        const tieneVideo = item.sources.some((source) => source.provider === 'youtube');
        const tieneAudio = item.sources.some((source) => source.provider === 'podcast');

        // ── Nombre (h3, no enlace) y fecha ────────────────────────────────
        await expect(ficha.getByRole('heading', { level: 3, name: item.title })).toBeVisible();
        await expect(ficha.getByText(fechaLarga(item.publishedAt, 'es'))).toBeVisible();

        // ── Carátula servida por el sitio (o marcador neutro) ─────────────
        if (item.hasCover) {
          const cover = ficha.locator('img');
          await expect(cover).toHaveCount(1);
          await expect(cover).toHaveAttribute('alt', `Carátula del episodio ${item.title}`);
          const src = await cover.getAttribute('src');
          expect(src, 'la carátula se sirve desde el propio sitio').toContain('/api/v1/media/');
          expect(src, 'la carátula no viene de un tercero').not.toMatch(/youtube|spotify|anchor/);
        } else {
          await expect(ficha.getByTestId('caratula-neutra')).toBeVisible();
          // El marcador neutro lleva el logotipo de marca (decorativo, `alt=""`):
          // lo que no puede haber es una **carátula** del episodio.
          await expect(ficha.locator('img[alt^="Carátula del episodio"]')).toHaveCount(0);
        }

        // ── Descripción corta: solo la línea que existe (SC-010) ──────────
        // La ficha compone la fecha en un `p` y la descripción en otro: sin
        // descripción el segundo `p` **no existe** (nunca un hueco).
        if (item.descriptionEs) {
          await expect(ficha.locator('p')).toHaveCount(2);
          await expect(ficha).toContainText(item.descriptionEs);
        } else {
          await expect(ficha.locator('p')).toHaveCount(1);
        }

        // ── Enlaces: solo los formatos que la ficha tiene (sin líneas vacías) ──
        await expect(ficha.getByRole('link')).toHaveCount(Number(tieneVideo) + Number(tieneAudio));
        if (tieneVideo) {
          await expect(
            ficha.getByRole('link', { name: /^Abre YouTube en una pestaña nueva/ }),
          ).toBeVisible();
        } else {
          await expect(ficha.getByRole('link', { name: /^Abre YouTube/ })).toHaveCount(0);
        }
        if (tieneAudio) {
          await expect(
            ficha.getByRole('link', { name: /^Abre Spotify en una pestaña nueva/ }),
          ).toBeVisible();
        } else {
          await expect(ficha.getByRole('link', { name: /^Abre Spotify/ })).toHaveCount(0);
        }
        // Cero marcadores de relleno en la ficha.
        await expect(ficha.getByText('—', { exact: true })).toHaveCount(0);
      }

      // ── SC-002: lo no publicable no sale por ninguna vía ────────────────
      const html = await page.evaluate(() => document.documentElement.outerHTML);
      for (const oculta of seed.ocultas) {
        expect(html, `la ficha «${oculta.key}» no debe aparecer`).not.toContain(oculta.title);
      }

      // ── SC-013: solo prédicas y podcasts ────────────────────────────────
      const mainHtml = await page.evaluate(() => document.querySelector('main')?.outerHTML ?? '');
      for (const ajeno of ['Ministerio', 'Lo que viene', 'Categoría', '/eventos', '/ofrendas']) {
        expect(mainHtml, `la sección no debe contener «${ajeno}»`).not.toContain(ajeno);
      }
    } finally {
      await context.close();
    }
  });

  test('la entrada «Medios» está en la barra y en el mapa del sitio (FR-007/FR-008)', async ({
    browser,
  }) => {
    const { context, page } = await visitorContext(browser);
    try {
      await page.goto('/');
      await expect(
        page.getByRole('navigation', { name: NAV }).getByRole('link', { name: H1 }),
      ).toHaveAttribute('href', '/medios');
      await expect(
        page.getByRole('navigation', { name: SITE_MAP }).getByRole('link', { name: H1 }),
      ).toHaveAttribute('href', '/medios');

      // En un solo paso: el enlace de la barra lleva a la sección.
      await page.getByRole('navigation', { name: NAV }).getByRole('link', { name: H1 }).click();
      await expect(page).toHaveURL(/\/medios$/);
      await expect(page.getByRole('heading', { level: 1, name: H1 })).toBeVisible();
    } finally {
      await context.close();
    }
  });

  test('0 peticiones a terceros, 0 iframes y carátulas del propio sitio (SC-005)', async ({
    browser,
  }) => {
    const { context, page } = await visitorContext(browser);
    const terceros: string[] = [];
    page.on('request', (request) => {
      const url = request.url();
      if (TERCEROS.some((host) => url.includes(host))) {
        terceros.push(url);
      }
    });
    try {
      for (const path of ['/medios', '/']) {
        await page.goto(path);
        await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
      }

      expect(terceros, 'ninguna petición a YouTube, Spotify ni Anchor').toEqual([]);

      await page.goto('/medios');
      await expect(page.getByRole('list', { name: LIST })).toBeVisible();
      const dom = await page.evaluate(() => {
        const thirdPartyAttrs = ['src', 'data', 'poster'];
        const encontrados: string[] = [];
        for (const el of Array.from(document.querySelectorAll('*'))) {
          for (const attr of thirdPartyAttrs) {
            const value = el.getAttribute(attr);
            if (value && /youtube|ytimg|spotify|scdn\.co|anchor\.fm|cloudfront/i.test(value)) {
              encontrados.push(`${el.tagName.toLowerCase()}[${attr}=${value}]`);
            }
          }
        }
        return {
          iframes: document.querySelectorAll('iframe').length,
          encontrados,
          imagenes: Array.from(document.querySelectorAll('main img')).map(
            (img) => img.getAttribute('src') ?? '',
          ),
        };
      });

      expect(dom.iframes, '0 reproductores incrustados').toBe(0);
      expect(dom.encontrados, '0 atributos con recursos de terceros').toEqual([]);
      // Las carátulas son **copias servidas por el sitio** (`/api/v1/media/…`);
      // el marcador neutro usa el logotipo de marca del propio frontend. Ningún
      // `<img>` de la sección sale del sitio (SC-005).
      expect(
        dom.imagenes.filter((src) => src.includes('/api/v1/media/')).length,
        'las fichas con carátula usan una copia servida por el sitio',
      ).toBeGreaterThan(0);
      for (const src of dom.imagenes) {
        expect(src, `ningún recurso de la sección sale del sitio (${src})`).not.toMatch(
          /youtube|ytimg|spotify|anchor|cloudfront/i,
        );
      }
    } finally {
      await context.close();
    }
  });

  test('los enlaces externos se anuncian y abren en pestaña nueva (SC-006/FR-005)', async ({
    browser,
  }) => {
    const context = await browser.newContext();
    // Los servicios externos no se visitan de verdad: se responde con una página
    // mínima para poder comprobar **la pestaña nueva y su URL** sin depender de
    // la red ni cargar nada de terceros en el contexto del sitio.
    await context.route(/youtube\.com|spotify\.com/, (route) =>
      route.fulfill({
        status: 200,
        contentType: 'text/html',
        body: '<html><body>externo</body></html>',
      }),
    );
    const page = await context.newPage();
    try {
      const dosEnlaces = seed.porClave['dos-enlaces'];
      await page.goto('/medios');
      const ficha = card(page, dosEnlaces.title);

      // ── Anuncio visible: servicio + indicador de pestaña nueva ──────────
      await expect(ficha.getByText('Ver en YouTube')).toBeVisible();
      await expect(ficha.getByText('Escuchar en Spotify')).toBeVisible();
      await expect(ficha.getByText(NEW_TAB).first()).toBeVisible();

      // ── Anuncio completo en el nombre accesible ─────────────────────────
      const youtube = ficha.getByRole('link', { name: /^Abre YouTube en una pestaña nueva/ });
      const spotify = ficha.getByRole('link', { name: /^Abre Spotify en una pestaña nueva/ });
      const youtubeUrl = dosEnlaces.sources.find((source) => source.provider === 'youtube')!.url;
      const spotifyUrl = dosEnlaces.sources.find((source) => source.provider === 'podcast')!.url;

      for (const [link, url, servicio] of [
        [youtube, youtubeUrl, 'YouTube'],
        [spotify, spotifyUrl, 'Spotify'],
      ] as const) {
        await expect(link).toHaveAttribute('href', url);
        await expect(link).toHaveAttribute('target', '_blank');
        await expect(link).toHaveAttribute('rel', 'noopener noreferrer');
        const label = await link.getAttribute('aria-label');
        expect(label, `el anuncio nombra el servicio (${servicio})`).toContain(servicio);
        expect(label, 'el anuncio dice que se abre en una pestaña nueva').toContain(
          'en una pestaña nueva',
        );
        expect(label, 'el anuncio dice que se sale del sitio').toContain('sales del sitio');
      }

      // ── Se abre en pestaña nueva y la del sitio sigue abierta ───────────
      const [popup] = await Promise.all([context.waitForEvent('page'), youtube.click()]);
      await popup.waitForLoadState('domcontentloaded');
      expect(popup.url()).toBe(youtubeUrl);
      expect(page.url()).toContain('/medios');
      await popup.close();
    } finally {
      await context.close();
    }
  });

  test('inglés con fallback `en → es` y datos idénticos entre idiomas (SC-011)', async ({
    browser,
  }) => {
    const { context, page } = await visitorContext(browser);
    try {
      const dosEnlaces = seed.porClave['dos-enlaces'];
      const sinIngles = seed.porClave['sin-ingles'];

      await page.goto('/medios');
      await expect(page.getByRole('heading', { level: 1, name: H1 })).toBeVisible();

      const fechaEs = fechaLarga(dosEnlaces.publishedAt, 'es');
      const videoEs = await card(page, dosEnlaces.title)
        .getByRole('link', { name: /^Abre YouTube/ })
        .getAttribute('href');

      await page.getByRole('banner').getByRole('radio', { name: 'Inglés' }).check();
      await expect(page.getByRole('heading', { level: 1, name: H1_EN })).toBeVisible();
      // La lista se rotula en el idioma activo (i18n), como todo lo demás.
      await expect(page.getByRole('list', { name: LIST_EN })).toBeVisible();

      // Contenido con versión inglesa…
      await expect(card(page, dosEnlaces.title, LIST_EN)).toContainText(dosEnlaces.descriptionEn!);
      // …y sin ella: cae al español, sin marca de «sin traducir» ni hueco.
      await expect(card(page, sinIngles.title, LIST_EN)).toContainText(sinIngles.descriptionEs!);
      await expect(card(page, sinIngles.title, LIST_EN)).not.toContainText('—');

      // Nombre, fecha y enlaces idénticos (FR-026).
      await expect(
        card(page, dosEnlaces.title, LIST_EN).getByRole('heading', {
          level: 3,
          name: dosEnlaces.title,
        }),
      ).toBeVisible();
      expect(
        await card(page, dosEnlaces.title, LIST_EN)
          .getByRole('link', { name: /^Opens YouTube/ })
          .getAttribute('href'),
      ).toBe(videoEs);
      // La fecha se escribe en el idioma activo (dato único, palabra localizada).
      await expect(card(page, dosEnlaces.title, LIST_EN)).toContainText(
        fechaLarga(dosEnlaces.publishedAt, 'en'),
      );
      expect(fechaLarga(dosEnlaces.publishedAt, 'es')).not.toBe(
        fechaLarga(dosEnlaces.publishedAt, 'en'),
      );

      // Y vuelve a español sin alterar nada.
      await page.getByRole('banner').getByRole('radio', { name: 'Spanish' }).check();
      await expect(page.getByRole('heading', { level: 1, name: H1 })).toBeVisible();
      await expect(card(page, dosEnlaces.title)).toContainText(dosEnlaces.descriptionEs!);
      await expect(card(page, dosEnlaces.title)).toContainText(fechaEs);
    } finally {
      await context.close();
    }
  });

  test('320/768/1280 px sin desplazamiento horizontal (SC-017)', async ({ browser }) => {
    const { context, page } = await visitorContext(browser);
    try {
      for (const width of [320, 768, 1280]) {
        await page.setViewportSize({ width, height: 900 });
        await page.goto('/medios');
        await expect(page.getByRole('list', { name: LIST })).toBeVisible();
        expect(
          await hasNoHorizontalScroll(page),
          `sin desplazamiento horizontal a ${width} px`,
        ).toBe(true);
        // La ficha completa (carátula, nombre, descripción y enlaces) sigue ahí.
        const ficha = card(page, seed.porClave['dos-enlaces'].title);
        await expect(ficha.locator('h3')).toBeVisible();
        await expect(ficha.getByRole('link', { name: /^Abre YouTube/ })).toBeVisible();
        await expect(ficha.getByRole('link', { name: /^Abre Spotify/ })).toBeVisible();
      }
    } finally {
      await context.close();
    }
  });

  test('«Cargar más» acumula tandas y desaparece al agotarse (FR-006/Q11)', async ({ browser }) => {
    seed = await seedMedios(adminPage, { publishedCount: 14 });

    const { context, page } = await visitorContext(browser);
    try {
      await page.goto('/medios');
      const list = page.getByRole('list', { name: LIST });
      const mas = page.getByRole('button', { name: CARGAR_MAS });

      // Primera tanda: 12 de 14, con el botón al final.
      await expect(list.getByRole('listitem')).toHaveCount(12);
      await expect(mas).toBeVisible();
      expect(await list.locator('h3').allInnerTexts()).toEqual(
        seed.publicables.slice(0, 12).map((item) => item.title),
      );

      // Cada toque agrega la tanda siguiente **al final de la misma lista**.
      await mas.click();
      await expect(list.getByRole('listitem')).toHaveCount(14);
      expect(await list.locator('h3').allInnerTexts()).toEqual(
        seed.publicables.map((item) => item.title),
      );
      // Al agotarse, el botón desaparece sin mensaje.
      await expect(page.getByRole('button', { name: CARGAR_MAS })).toHaveCount(0);
    } finally {
      await context.close();
    }
  });

  test('sin nada publicable: la entrada desaparece y `/medios` redirige sin destello (SC-016)', async ({
    browser,
  }) => {
    // Control positivo: con publicables la misma observación **sí** registra la
    // sección. Sin esto, la aserción de «sin destello» no probaría nada.
    seed = await seedMedios(adminPage);
    const { context, page } = await visitorContext(browser);
    try {
      await page.addInitScript(() => {
        const store = window as unknown as { __mediosFlash?: string[] };
        store.__mediosFlash = [];
        const record = () => {
          const markers: string[] = [];
          for (const heading of Array.from(document.querySelectorAll('h1'))) {
            if ((heading.textContent ?? '').trim() === 'Medios') {
              markers.push('h1');
            }
          }
          if (document.querySelector('ul[aria-label="Lista de episodios de medios"]')) {
            markers.push('lista');
          }
          if (document.querySelector('a[aria-label^="Abre YouTube en una pestaña nueva"]')) {
            markers.push('ficha');
          }
          for (const marker of markers) {
            if (!store.__mediosFlash!.includes(marker)) {
              store.__mediosFlash!.push(marker);
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
          () => (window as unknown as { __mediosFlash?: string[] }).__mediosFlash ?? [],
        );

      await page.goto('/medios');
      await expect(page.getByRole('list', { name: LIST })).toBeVisible();
      expect(await flash(), 'el detector de destello no es vacuo').toContain('h1');

      // Escenario 0 publicables (SC-016): las fichas marcadas/ocultas siguen en
      // la base, pero **nada** es publicable.
      seed = await seedMedios(adminPage, { publishedCount: 0 });
      expect(seed.publicables).toHaveLength(0);

      await page.goto('/');
      await expect(page.locator('#proximamente')).toBeVisible();
      const historyBase = await page.evaluate(() => history.length);

      await page.goto('/medios');

      // URL final: la portada, con contenido y navegación operables.
      await expect(page).toHaveURL(/\/$/);
      await expect(page.locator('#proximamente')).toBeVisible();
      await expect(page.getByRole('navigation', { name: NAV })).toBeVisible();

      // Sin destello: nunca se pintó la sección, su lista ni sus fichas.
      expect(await flash(), 'la redirección no pinta la sección').toEqual([]);

      // `replace` (P9-16): no añade entrada al historial y «Atrás» no vuelve.
      expect(
        await page.evaluate(() => history.length),
        'la redirección usa `replace` (no añade entrada al historial)',
      ).toBe(historyBase + 1);
      await page.goBack();
      await expect(page).toHaveURL(/\/$/);
      expect(page.url()).not.toContain('/medios');

      // La entrada está oculta en barra, mapa del pie y en cualquier enlace.
      await expect(
        page.getByRole('navigation', { name: NAV }).getByRole('link', { name: H1 }),
      ).toHaveCount(0);
      await expect(
        page.getByRole('navigation', { name: SITE_MAP }).getByRole('link', { name: H1 }),
      ).toHaveCount(0);
      await expect(page.locator('a[href="/medios"]')).toHaveCount(0);

      // El bloque de la portada tampoco se dibuja (FR-027).
      await expect(page.locator('#medios')).toHaveCount(0);
    } finally {
      await context.close();
    }
  });
});
