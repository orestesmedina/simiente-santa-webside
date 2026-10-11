import { expect, test, type Page } from '@playwright/test';
import { apiSend } from './helpers';
import { createAdminContext, seedPublicPortada, type PublicSeed } from './publicSeed';
import { waitForScrollEnd } from './publicUi';

/**
 * E2E de la navegación móvil (menú hamburguesa accesible) y del mapa del sitio
 * del pie (F10 · T1030 · US2; SC-003/SC-004/SC-009/SC-015). Verifica el patrón
 * *disclosure* de APG (sin diálogo ni trampa de foco), el colapso explícito de
 * la única lista, el bloqueo del scroll de fondo, el cruce de breakpoint y el
 * mapa del sitio sin enlaces rotos ni nombres duplicados.
 *
 * Ejecutar con `--workers=1` (singletons compartidos, rate-limit por IP).
 */

const SECTIONS_NAV = 'Secciones de la portada';
const SITEMAP = 'Mapa del sitio';

function revealBlocks(page: Page) {
  return page.locator('.reveal-block');
}

test.describe('Navegación móvil y mapa del sitio (F10 · T1030)', () => {
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

  test('menú hamburguesa accesible a 320 px: ratón, teclado, foco y colapso', async ({ page }) => {
    await page.setViewportSize({ width: 320, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    const nav = page.getByRole('navigation', { name: SECTIONS_NAV });
    const burger = page.getByRole('button', { name: 'Abrir menú de secciones' });
    const list = nav.locator('ul.public-nav-list');
    const firstLink = nav.getByRole('link', { name: 'Quiénes somos' });

    // Objetivo ≥ 44 px, estado anunciado y control del panel.
    const burgerBox = await burger.boundingBox();
    expect(burgerBox?.height ?? 0).toBeGreaterThanOrEqual(44);
    expect(burgerBox?.width ?? 0).toBeGreaterThanOrEqual(44);
    await expect(burger).toHaveAttribute('aria-expanded', 'false');
    await expect(burger).toHaveAttribute('aria-controls', (await list.getAttribute('id')) ?? '');

    // Cerrado: los enlaces NO son alcanzables (display:none) pero siguen en el DOM.
    await expect(list).toBeHidden();
    await expect(firstLink).toBeHidden();
    await expect(nav.locator('ul')).toHaveCount(1);

    // Abrir con ratón: el foco entra en el primer enlace; sin diálogo ni trampa.
    await burger.click();
    await expect(page.getByRole('button', { name: 'Cerrar menú de secciones' })).toHaveAttribute(
      'aria-expanded',
      'true',
    );
    await expect(list).toBeVisible();
    await expect(firstLink).toBeVisible();
    await expect(firstLink).toBeFocused();
    await expect(page.locator('[role="dialog"]')).toHaveCount(0);
    await expect(page.locator('[aria-modal]')).toHaveCount(0);

    // Una sola lista en el DOM: cada sección resuelve a una única coincidencia
    // dentro del landmark nombrado (D15/C4), también con el menú abierto.
    for (const name of ['Quiénes somos', 'Horario de servicios', 'Contacto']) {
      await expect(nav.getByRole('link', { name, exact: true })).toHaveCount(1);
    }

    // Sin trampa de foco: Tab puede salir del panel hacia la barra.
    await page.keyboard.press('Shift+Tab');
    const focused = await page.evaluate(() => {
      const el = document.activeElement as HTMLElement | null;
      return el?.closest('ul.public-nav-list') === null;
    });
    expect(focused, 'Shift+Tab sale del panel (sin trampa de foco)').toBe(true);

    // Scroll de fondo bloqueado con el panel abierto.
    await expect(page.locator('html')).toHaveAttribute('data-nav-open', 'true');
    expect(await page.evaluate(() => getComputedStyle(document.documentElement).overflow)).toBe(
      'hidden',
    );
    const scrollBefore = await page.evaluate(() => window.scrollY);
    await page.mouse.wheel(0, 600);
    await page.waitForTimeout(150);
    expect(await page.evaluate(() => window.scrollY)).toBe(scrollBefore);

    // Escape cierra y devuelve el foco al botón.
    await page.keyboard.press('Escape');
    const reopen = page.getByRole('button', { name: 'Abrir menú de secciones' });
    await expect(reopen).toHaveAttribute('aria-expanded', 'false');
    await expect(reopen).toBeFocused();
    await expect(page.locator('html')).not.toHaveAttribute('data-nav-open', 'true');

    // Abrir con teclado y cerrar con el scrim (también devuelve el foco).
    await reopen.focus();
    await page.keyboard.press('Enter');
    await expect(page.getByRole('button', { name: 'Cerrar menú de secciones' })).toHaveAttribute(
      'aria-expanded',
      'true',
    );
    await page.getByTestId('mobile-menu-scrim').click({ position: { x: 300, y: 800 } });
    await expect(reopen).toHaveAttribute('aria-expanded', 'false');
    await expect(reopen).toBeFocused();

    // Activar una sección cierra el panel y el destino queda visible.
    await reopen.click();
    await nav.getByRole('link', { name: 'Horario de servicios' }).click();
    await expect(reopen).toHaveAttribute('aria-expanded', 'false');
    const heading = page.getByRole('heading', { level: 2, name: 'Horario de servicios' });
    await expect(heading).toBeVisible();
    await waitForScrollEnd(page);
    const headerHeight = (await page.locator('header').boundingBox())?.height ?? 0;
    expect((await heading.boundingBox())?.y ?? 0).toBeGreaterThanOrEqual(headerHeight - 4);

    // Una sola lista en el DOM: cada sección resuelve a una única coincidencia.
    // Sin desplazamiento horizontal a 320 px.
    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
    );
    expect(overflow).toBeLessThanOrEqual(1);
  });

  test('al ensanchar a ≥ 861 px el panel se cierra y reaparecen las anclas', async ({ page }) => {
    await page.setViewportSize({ width: 320, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    const burger = page.getByRole('button', { name: 'Abrir menú de secciones' });
    await burger.click();
    await expect(page.getByRole('button', { name: 'Cerrar menú de secciones' })).toBeVisible();

    await page.setViewportSize({ width: 1000, height: 900 });

    // A ≥ 861 px el botón hamburguesa desaparece y reaparecen las anclas.
    await expect(page.getByRole('button', { name: 'Abrir menú de secciones' })).toBeHidden();
    const nav = page.getByRole('navigation', { name: SECTIONS_NAV });
    await expect(nav.getByRole('link', { name: 'Contacto' })).toBeVisible();
    await expect(page.locator('html')).not.toHaveAttribute('data-nav-open', 'true');
  });

  test('la cabecera gana data-scrolled al bajar en ambos modos', async ({ page }) => {
    for (const width of [320, 1280] as const) {
      await page.setViewportSize({ width, height: 900 });
      await page.goto('/');
      await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();
      await expect(page.locator('header')).toHaveAttribute('data-scrolled', 'false');
      await page.mouse.wheel(0, 900);
      await expect(page.locator('header'), `cabecera desplazada a ${width}px`).toHaveAttribute(
        'data-scrolled',
        'true',
      );
    }
  });

  test('mapa del sitio del pie: secciones, columnas/apilado, nombres únicos y sin rotos', async ({
    browser,
    page,
  }) => {
    // Escritorio: columnas.
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    const sitemap = page.getByRole('navigation', { name: SITEMAP });
    await expect(sitemap).toBeVisible();
    await expect(sitemap.getByRole('link', { name: 'Inicio', exact: true })).toBeVisible();
    for (const name of [
      'Quiénes somos',
      'Horario de servicios',
      'WhatsApp',
      'Contacto',
      'Redes sociales',
    ]) {
      await expect(sitemap.getByRole('link', { name, exact: true })).toBeVisible();
    }
    // Landmark con nombre propio, distinto del de la cabecera.
    await expect(page.getByRole('navigation', { name: SECTIONS_NAV })).toBeVisible();

    const columns = await sitemap.evaluate(
      (el) => getComputedStyle(el).gridTemplateColumns.split(' ').length,
    );
    expect(columns, 'tres columnas en escritorio').toBe(3);

    // Rótulos de grupo sin rol heading.
    for (const label of ['Explorar', 'Contacto', 'Síguenos']) {
      await expect(sitemap.getByRole('paragraph').filter({ hasText: label })).toHaveCount(1);
      await expect(sitemap.getByRole('heading', { name: label })).toHaveCount(0);
    }

    // Nombres accesibles únicos y sin colisión con los de las secciones.
    const names = await sitemap
      .getByRole('link')
      .evaluateAll((els) =>
        els.map((el) => el.getAttribute('aria-label') ?? (el.textContent ?? '').trim()),
      );
    expect(new Set(names).size).toBe(names.length);
    expect(names).toContain('Contacto: Correo');
    expect(names).toContain('Contacto: Teléfono');
    expect(names).toContain('Contacto: WhatsApp');
    expect(names.some((name) => /Ver en /.test(name))).toBe(false);
    expect(names.some((name) => /Escribir por WhatsApp|Entrar al grupo/.test(name))).toBe(false);
    expect(names).not.toContain(seed.contactEmail);
    expect(names).not.toContain(seed.contactPhoneNormalized);

    // Un enlace del pie lleva a su sección (título visible, barra no lo tapa).
    await sitemap.getByRole('link', { name: 'Contacto', exact: true }).click();
    const heading = page.getByRole('heading', { level: 2, name: 'Contacto' });
    await expect(heading).toBeVisible();
    await waitForScrollEnd(page);
    const headerHeight = (await page.locator('header').boundingBox())?.height ?? 0;
    expect((await heading.boundingBox())?.y ?? 0).toBeGreaterThanOrEqual(headerHeight - 4);

    // Móvil: apilado (una sola columna).
    await page.setViewportSize({ width: 320, height: 900 });
    await page.goto('/');
    const stackedColumns = await page
      .getByRole('navigation', { name: SITEMAP })
      .evaluate((el) => getComputedStyle(el).gridTemplateColumns.split(' ').length);
    expect(stackedColumns, 'apilado en móvil').toBe(1);

    // Sin secciones en borrador: quitar WhatsApp deja el enlace fuera (nav y pie).
    const admin = await createAdminContext(browser);
    try {
      const adminPage = await admin.newPage();
      const state = (await (await apiSend(adminPage, 'GET', '/api/v1/admin/portada')).json()) as {
        whatsapp: { items: { id: string }[] };
      };
      const channel = state.whatsapp.items[0];
      await apiSend(adminPage, 'PATCH', `/api/v1/admin/portada/whatsapp/${channel.id}`, {
        publicationState: 'draft',
      });
      // También el horario: el CTA «Ver horario» del hero debe desaparecer (I1).
      const scheduleState = (await (
        await apiSend(adminPage, 'GET', '/api/v1/admin/portada')
      ).json()) as { schedule: { items: { id: string }[] } };
      const service = scheduleState.schedule.items[0];
      await apiSend(adminPage, 'PATCH', `/api/v1/admin/portada/horario/${service.id}`, {
        publicationState: 'draft',
      });

      await page.reload();
      await expect(page.locator('#whatsapp')).toHaveCount(0);
      await expect(page.locator('#horario')).toHaveCount(0);
      // 0 enlaces rotos: ni el CTA del hero ni la nav ni el pie apuntan a lo ausente.
      await expect(page.locator('a[href="#horario"]')).toHaveCount(0);
      await expect(page.locator('a[href="#whatsapp"]')).toHaveCount(0);
      await expect(
        page
          .getByRole('navigation', { name: SECTIONS_NAV })
          .getByRole('link', { name: 'WhatsApp', exact: true }),
      ).toHaveCount(0);
      await expect(
        page
          .getByRole('navigation', { name: SITEMAP })
          .getByRole('link', { name: 'WhatsApp', exact: true }),
      ).toHaveCount(0);

      await apiSend(adminPage, 'PATCH', `/api/v1/admin/portada/whatsapp/${channel.id}`, {
        publicationState: 'published',
      });
      await apiSend(adminPage, 'PATCH', `/api/v1/admin/portada/horario/${service.id}`, {
        publicationState: 'published',
      });
    } finally {
      await admin.close();
    }
  });

  test('el contenido se revela al bajar también en móvil', async ({ page }) => {
    await page.setViewportSize({ width: 320, height: 900 });
    await page.goto('/');
    await expect(page.getByRole('heading', { level: 1, name: seed.identityName })).toBeVisible();

    const blocks = revealBlocks(page);
    expect(await blocks.count()).toBeGreaterThan(0);
    const last = blocks.last();
    await last.scrollIntoViewIfNeeded();
    await expect(last).toHaveAttribute('data-reveal', 'in');
  });
});
