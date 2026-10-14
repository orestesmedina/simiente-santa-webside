import { expect, test, type Browser, type Locator, type Page } from '@playwright/test';
import { ADMIN, API_URL, apiSend, loginViaUi, uniqueSuffix } from './helpers';
import { FIXTURE_IMAGE, resetMinisterios } from './ministeriosSeed';

/**
 * E2E del panel de ministerios (US2/US3, quickstart §2–§6/§8/§12/§13): el equipo
 * **crea y publica un ministerio desde la interfaz en menos de 2 minutos**
 * (SC-004) y el público lo ve en la **primera carga** (SC-003); un alta con
 * encargado **nace con la ficha oculta** («Oculto por defecto», P6-3/D-13) y el
 * interruptor se enciende y se apaga **sin perder datos** (US2 esc. 10–11,
 * SC-014); un enlace no seguro (`javascript:`) y un contacto con **celular y
 * correo juntos** se rechazan sin guardar (SC-009/Q8); el **orden manual**
 * reordena el público (FR-004/FR-016); publicar, retirar y eliminar pasan por sus
 * controles y confirmaciones (SC-002/SC-003); **sin el permiso `ministerios`**
 * (incluida una cuenta con otros permisos de módulo) la API forzada responde
 * `403` y la interfaz lo dice sin detalles internos (SC-005); y **todo** queda
 * registrado en la auditoría de F2 con los códigos `ministry.*` y su **etiqueta
 * legible** (SC-011).
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
 * `--workers=1` es obligatorio en las cuatro (singletons compartidos y
 * ministerios sembrados por spec). Esta spec hace **3 inicios de sesión** (el
 * administrador por la interfaz y dos cuentas de permisos cruzados), muy por
 * debajo del tope de 20/min. En este entorno el navegador no arranca desde el
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

/** Nombres accesibles del sitio público (i18n `es`) y del panel. */
const SECTION = 'Ministerios de la iglesia';
const SWITCH_LABEL = 'Mostrar la ficha del encargado en el sitio público';
const HIDDEN_LEGEND =
  'La ficha del encargado quedará oculta: el público no verá ni su nombre ni su contacto, en ningún lugar del sitio. Tus datos quedan guardados aquí; puedes activar el interruptor otra vez cuando quieras publicarlos.';

const panelNav = (page: Page) => page.getByRole('navigation', { name: 'Navegación del panel' });
const quickAccess = (page: Page) => page.getByRole('navigation', { name: 'Accesos rápidos' });

/** Fila del listado (tabla en escritorio) del ministerio indicado. */
function row(page: Page, text: string) {
  return page.getByRole('table').getByRole('row').filter({ hasText: text });
}

/** Abre un contexto de visitante (sin sesión) y devuelve su página. */
async function visitorPage(browser: Browser): Promise<{ close: () => Promise<void>; page: Page }> {
  const context = await browser.newContext();
  const page = await context.newPage();
  return { close: () => context.close(), page };
}

/** Nombres de las fichas publicadas, tal como los ve el visitante. */
async function publicNames(browser: Browser): Promise<string[]> {
  const visitor = await visitorPage(browser);
  try {
    await visitor.page.goto('/ministerios');
    const grid = visitor.page.getByRole('list', { name: SECTION, exact: true });
    await expect(grid).toBeVisible();
    return await grid.locator('h3').allInnerTexts();
  } finally {
    await visitor.close();
  }
}

/** HTML completo de una superficie pública (para las aserciones de privacidad). */
async function publicHtml(browser: Browser, path: string): Promise<string> {
  const visitor = await visitorPage(browser);
  try {
    await visitor.page.goto(path);
    await expect(visitor.page.getByRole('heading', { level: 1 })).toBeVisible();
    return await visitor.page.evaluate(() => document.documentElement.outerHTML);
  } finally {
    await visitor.close();
  }
}

/** Rellena nombre y descripción en español (los dos campos obligatorios). */
async function fillCore(dialog: Locator, name: string, description: string): Promise<void> {
  await dialog.getByLabel('Nombre (Español)').fill(name);
  await dialog.getByLabel('Descripción (Español)').fill(description);
}

/** Inicia sesión con una cuenta nueva y supera el cambio obligatorio. */
async function loginNewAccount(page: Page, email: string, initial: string, next: string) {
  await loginViaUi(page, email, initial);
  await expect(page).toHaveURL(/\/cambiar-contrasena$/);
  await page.getByLabel(/^Contraseña actual\s*\*?$/).fill(initial);
  await page.getByLabel(/^Contraseña nueva\s*\*?$/).fill(next);
  await page.getByLabel(/^Confirmar contraseña nueva\s*\*?$/).fill(next);
  await page.getByRole('button', { name: 'Guardar' }).click();
  await page.getByRole('button', { name: 'Ir al panel' }).click();
  await expect(page).toHaveURL(/\/panel$/);
}

test.describe('Panel de ministerios — gestión, publicación, permisos y auditoría', () => {
  test.setTimeout(420_000);

  let page: Page;
  let closeAdmin: () => Promise<void>;
  let suffix: string;
  let titleA: string;
  let titleB: string;
  let leaderName: string;
  let leaderPhone: string;

  test.beforeAll(async ({ browser }) => {
    suffix = uniqueSuffix();
    titleA = `Ministerio de prueba A ${suffix}`;
    titleB = `Ministerio de prueba B ${suffix}`;
    leaderName = `Encargado Oculto ${suffix}`;
    // Número **distinto** del contacto público de la iglesia que ya vive en el
    // pie de la portada (F3): la aserción de privacidad comprueba que este valor
    // no aparece en ninguna superficie pública.
    leaderPhone = '+506 7000 8642';

    const context = await browser.newContext();
    page = await context.newPage();
    closeAdmin = () => context.close();

    // Login real por la interfaz con la cuenta que **sí** tiene el permiso
    // `ministerios` (FR-011).
    await loginViaUi(page, ADMIN.email, ADMIN.password);
    await expect(page).toHaveURL(/\/panel$/);
    await resetMinisterios(page);
  });

  test.afterAll(async () => {
    // La base es compartida con F3/F4/F10: se deja el módulo vacío.
    await resetMinisterios(page);
    await closeAdmin();
  });

  test('crear y publicar en menos de 2 minutos; el público lo ve en la primera carga (SC-003/SC-004)', async ({
    browser,
  }) => {
    // ── El módulo aparece en navegación y accesos rápidos (FR-011) ────────
    await page.goto('/panel');
    await expect(
      quickAccess(page).getByRole('link', { name: 'Gestión de ministerios' }),
    ).toHaveAttribute('href', '/panel/ministerios');
    await panelNav(page).getByRole('link', { name: 'Gestión de ministerios' }).click();
    await expect(page.getByRole('heading', { name: 'Gestión de ministerios' })).toBeVisible();
    await expect(page.getByText('Todavía no hay ministerios')).toBeVisible();

    // ── SC-004: alta + publicación desde la interfaz, cronometrada ────────
    const started = Date.now();
    await page.getByRole('button', { name: 'Agregar ministerio' }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByRole('heading', { name: 'Agregar ministerio' })).toBeVisible();

    // Solo nombre y descripción (el resto es opcional, FR-014/Q1/Q2/Q6).
    await fillCore(dialog, titleA, `Descripción de ${titleA}.`);
    await dialog.getByRole('button', { name: 'Agregar ministerio' }).click();

    await expect(page.getByText('Ministerio creado.')).toBeVisible();
    await expect(page.getByText('Quedó en borrador')).toBeVisible();
    await expect(row(page, titleA).getByText('Borrador')).toBeVisible();

    // Publicar es una acción directa, sin confirmación (FR-013).
    await row(page, titleA).getByRole('button', { name: 'Publicar' }).click();
    await expect(page.getByText('Publicado en "Ministerios".')).toBeVisible();
    await expect(row(page, titleA).getByText('Publicado')).toBeVisible();

    // ── SC-003: el visitante lo ve en la **primera carga** ────────────────
    expect(await publicNames(browser)).toEqual([titleA]);

    const elapsed = Date.now() - started;
    expect(elapsed, `crear y publicar tardó ${elapsed} ms (tope 120 000)`).toBeLessThan(120_000);
  });

  test('el encargado nace oculto y el interruptor se enciende y se apaga sin perder datos (P6-3, SC-014)', async ({
    browser,
  }) => {
    // ── Alta con encargado **sin tocar** el interruptor ──────────────────
    await page.goto('/panel/ministerios');
    await page.getByRole('button', { name: 'Agregar ministerio' }).click();
    const dialog = page.getByRole('dialog');
    await fillCore(dialog, titleB, `Descripción de ${titleB}.`);
    await dialog.getByLabel('Nombre del encargado (Español)').fill(leaderName);
    await dialog.getByLabel('Medio de contacto (opcional)').selectOption({ label: 'Celular' });
    await dialog.getByLabel('Número de celular (opcional)').fill(leaderPhone);

    // «Oculto por defecto» (D-13/P6-3): el interruptor llega **apagado** y la
    // leyenda explica qué implica.
    const toggle = dialog.getByRole('switch', { name: SWITCH_LABEL });
    await expect(toggle).toHaveAttribute('aria-checked', 'false');
    await expect(dialog.getByText(HIDDEN_LEGEND)).toBeVisible();

    await dialog.getByRole('button', { name: 'Agregar ministerio' }).click();
    await expect(page.getByText('Ministerio creado.')).toBeVisible();

    // El panel **sí** conserva el dato (dato guardado ≠ dato publicado)…
    await expect(row(page, titleB).getByText(leaderName)).toBeVisible();
    await expect(row(page, titleB).getByText('Ficha oculta')).toBeVisible();

    // …y el servidor lo confirma por API.
    const stored = (await (
      await page.request.get(`${API_URL}/api/v1/admin/ministerios?limit=100`)
    ).json()) as {
      items: { nameEs: string; leaderVisible: boolean; leaderPhone: string | null }[];
    };
    const storedB = stored.items.find((item) => item.nameEs === titleB);
    expect(storedB?.leaderVisible).toBe(false);
    expect(storedB?.leaderPhone).toBe(leaderPhone);

    await row(page, titleB).getByRole('button', { name: 'Publicar' }).click();
    await expect(row(page, titleB).getByText('Publicado')).toBeVisible();

    // Publicado pero con la ficha oculta: el público no recibe nada del
    // encargado (SC-014), ni en el HTML de la sección ni en el de la portada.
    // (El orden entre dos ministerios con el mismo `sortOrder` lo decide el
    // `id` —FR-004—, así que aquí solo se comprueba la pertenencia.)
    expect((await publicNames(browser)).sort()).toEqual([titleA, titleB].sort());
    for (const path of ['/ministerios', '/']) {
      const html = await publicHtml(browser, path);
      expect(html, `el HTML de ${path} no debe contener el nombre del encargado`).not.toContain(
        leaderName,
      );
      expect(html, `el HTML de ${path} no debe contener el celular del encargado`).not.toContain(
        leaderPhone,
      );
      expect(html).not.toContain(leaderPhone.replace(/[^\d]/g, ''));
    }

    // ── Encender el interruptor **sin perder datos** (US2 esc. 10–11) ────
    await row(page, titleB).getByRole('button', { name: 'Editar' }).click();
    const editDialog = page.getByRole('dialog');
    await expect(editDialog.getByLabel('Nombre del encargado (Español)')).toHaveValue(leaderName);
    await expect(editDialog.getByLabel('Número de celular (opcional)')).toHaveValue(leaderPhone);
    await editDialog.getByRole('switch', { name: SWITCH_LABEL }).click();
    await expect(editDialog.getByRole('switch', { name: SWITCH_LABEL })).toHaveAttribute(
      'aria-checked',
      'true',
    );
    // La imagen se sube en el mismo guardado (FR-020 → `ministry.image.upload`).
    await editDialog.getByLabel('Archivo').setInputFiles(FIXTURE_IMAGE);
    await editDialog.getByLabel('Texto alternativo (Español)').fill(`Foto de ${titleB}`);
    await editDialog.getByRole('button', { name: 'Guardar cambios' }).click();
    await expect(
      page.getByText('Cambios guardados. La ficha del encargado está publicada.'),
    ).toBeVisible();

    // Ahora sí: la ficha publicada llega al visitante.
    expect(await publicHtml(browser, '/ministerios')).toContain(leaderName);

    // ── Apagarlo: el dato se conserva y deja de ser público ──────────────
    await row(page, titleB).getByRole('button', { name: 'Editar' }).click();
    const hideDialog = page.getByRole('dialog');
    await hideDialog.getByRole('switch', { name: SWITCH_LABEL }).click();
    await hideDialog.getByRole('button', { name: 'Guardar cambios' }).click();
    await expect(
      page.getByText(
        'Cambios guardados. La ficha del encargado quedó oculta: el público no verá su nombre ni su contacto.',
      ),
    ).toBeVisible();

    // Los datos siguen guardados y editables (no se borró nada)…
    await row(page, titleB).getByRole('button', { name: 'Editar' }).click();
    const reopen = page.getByRole('dialog');
    await expect(reopen.getByLabel('Nombre del encargado (Español)')).toHaveValue(leaderName);
    await expect(reopen.getByLabel('Número de celular (opcional)')).toHaveValue(leaderPhone);
    await expect(reopen.getByRole('switch', { name: SWITCH_LABEL })).toHaveAttribute(
      'aria-checked',
      'false',
    );
    await reopen.getByRole('button', { name: 'Cancelar' }).click();

    // …y el público ya no los ve.
    expect(await publicHtml(browser, '/ministerios')).not.toContain(leaderName);
  });

  test('el orden manual reordena el público y el panel (FR-004/FR-016/SC-003)', async ({
    browser,
  }) => {
    await page.goto('/panel/ministerios');
    // Con empate de Orden (0) el desempate es el `id` (FR-004), así que no se
    // asume un orden inicial: se fija uno explícito y se comprueba que manda.
    await row(page, titleA).getByRole('button', { name: 'Editar' }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('Orden (opcional)').fill('50');
    await dialog.getByRole('button', { name: 'Guardar cambios' }).click();
    await expect(page.getByText('Cambios guardados: ya visibles en la sección.')).toBeVisible();

    // El panel muestra el valor numérico y el mismo orden que el público.
    const rows = page.getByRole('table').getByRole('row');
    await expect(rows.nth(1)).toContainText(titleB);
    await expect(rows.nth(2)).toContainText(titleA);
    expect(await publicNames(browser)).toEqual([titleB, titleA]);
  });

  test('rechazos sin guardar: enlace no seguro y celular con correo (SC-009/Q8)', async () => {
    await page.goto('/panel/ministerios');

    // ── `javascript:` rechazado en la interfaz, sin guardar nada ─────────
    await row(page, titleA).getByRole('button', { name: 'Editar' }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('Enlace del grupo de WhatsApp (opcional)').fill('javascript:alert(1)');
    await dialog.getByRole('button', { name: 'Guardar cambios' }).click();
    // El mensaje aparece junto al campo (y en el resumen del diálogo).
    await expect(
      dialog
        .getByText(
          'Ese enlace no es válido o no es seguro. Debe empezar por https:// y ser una dirección normal.',
        )
        .first(),
    ).toBeVisible();
    // El diálogo sigue abierto (nada se guardó a medias).
    await expect(dialog).toBeVisible();
    const saved = (await (
      await page.request.get(`${API_URL}/api/v1/admin/ministerios?limit=100`)
    ).json()) as { items: { nameEs: string; whatsappUrl: string | null }[] };
    expect(saved.items.find((item) => item.nameEs === titleA)?.whatsappUrl ?? null).toBeNull();
    await dialog.getByRole('button', { name: 'Cancelar' }).click();

    // ── Y el servidor es la autoridad final (SC-009) ─────────────────────
    const unsafe = await apiSend(page, 'POST', '/api/v1/admin/ministerios', {
      nameEs: `Enlace inseguro ${suffix}`,
      descriptionEs: 'Intento con javascript:.',
      whatsappUrl: 'javascript:alert(1)',
    });
    expect(unsafe.status(), 'el servidor rechaza un enlace no seguro').toBe(400);

    // ── Celular y correo **juntos**: la interfaz solo ofrece uno (Q8), así que
    //    se fuerza por API para probar la regla del servidor ───────────────
    const both = await apiSend(page, 'POST', '/api/v1/admin/ministerios', {
      nameEs: `Contacto doble ${suffix}`,
      descriptionEs: 'Intento con celular y correo.',
      leaderNameEs: leaderName,
      leaderPhone,
      leaderEmail: `doble-${suffix}@ejemplo.com`,
    });
    expect(both.status(), 'el servidor rechaza celular y correo juntos').toBe(400);

    // Ninguna de las dos altas quedó guardada.
    const after = (await (
      await page.request.get(`${API_URL}/api/v1/admin/ministerios?limit=100`)
    ).json()) as { items: { nameEs: string }[] };
    expect(after.items.map((item) => item.nameEs)).not.toContain(`Enlace inseguro ${suffix}`);
    expect(after.items.map((item) => item.nameEs)).not.toContain(`Contacto doble ${suffix}`);
  });

  test('retirar y eliminar con confirmación; el público deja de verlo (SC-002/SC-003)', async ({
    browser,
  }) => {
    await page.goto('/panel/ministerios');

    // ── Retirar de la sección: confirmación suave (FR-012/FR-013) ────────
    await row(page, titleA).getByRole('button', { name: 'Retirar de la sección' }).click();
    const retireDialog = page.getByRole('dialog');
    await expect(
      retireDialog.getByRole('heading', { name: 'Retirar de la sección' }),
    ).toBeVisible();
    await retireDialog.getByRole('button', { name: 'Retirar' }).click();
    await expect(page.getByText('Retirado de la sección.')).toBeVisible();
    await expect(row(page, titleA).getByText('Borrador')).toBeVisible();

    // SC-002: el retirado ya no está en la sección pública.
    expect(await publicNames(browser)).toEqual([titleB]);

    // ── Eliminar: confirmación explícita de no recuperable (FR-017) ──────
    await row(page, titleA).getByRole('button', { name: 'Eliminar' }).click();
    const deleteDialog = page.getByRole('dialog');
    await expect(deleteDialog).toContainText('se eliminará para siempre');
    await deleteDialog.getByRole('button', { name: 'Eliminar' }).click();
    await expect(page.getByText('Eliminado.')).toBeVisible();
    await expect(row(page, titleA)).toHaveCount(0);
  });

  test('permisos cruzados: sin el permiso `ministerios` no se aplica nada (SC-005)', async ({
    browser,
  }) => {
    // ── Una cuenta con **otros** permisos de módulo (no `ministerios`) ───
    const otherRole = await apiSend(page, 'POST', '/api/v1/admin/roles', {
      name: `Otros módulos ${suffix}`,
      permissions: ['eventos', 'noticias'],
    });
    expect(otherRole.ok(), `POST rol → ${otherRole.status()}`).toBe(true);
    const otherRoleId = ((await otherRole.json()) as { id: string }).id;
    const otherEmail = `sin-ministerios-${suffix}@ejemplo.com`;
    expect(
      (
        await apiSend(page, 'POST', '/api/v1/admin/usuarios', {
          firstName: 'Sara',
          lastName: `SinMinisterios ${suffix}`,
          email: otherEmail,
          phone: '612 000 333',
          roleId: otherRoleId,
          password: 'Inicial.2026',
        })
      ).ok(),
    ).toBe(true);

    // ── Y una cuenta con **solo** `ministerios` (el cruce inverso) ────────
    const ownRole = await apiSend(page, 'POST', '/api/v1/admin/roles', {
      name: `Solo ministerios ${suffix}`,
      permissions: ['ministerios'],
    });
    expect(ownRole.ok()).toBe(true);
    const ownRoleId = ((await ownRole.json()) as { id: string }).id;
    const ownEmail = `solo-ministerios-${suffix}@ejemplo.com`;
    expect(
      (
        await apiSend(page, 'POST', '/api/v1/admin/usuarios', {
          firstName: 'Mateo',
          lastName: `SoloMinisterios ${suffix}`,
          email: ownEmail,
          phone: '612 000 444',
          roleId: ownRoleId,
          password: 'Inicial.2026',
        })
      ).ok(),
    ).toBe(true);

    const existing = (await (
      await page.request.get(`${API_URL}/api/v1/admin/ministerios?limit=100`)
    ).json()) as { items: { id: string }[] };
    const someId = existing.items[0]?.id ?? '00000000-0000-0000-0000-000000000000';

    // ── Cuenta sin `ministerios`: su módulo abre, el ajeno no ────────────
    const otherContext = await browser.newContext();
    const otherPage = await otherContext.newPage();
    try {
      await loginNewAccount(otherPage, otherEmail, 'Inicial.2026', 'Definitiva.2026');
      await expect(
        panelNav(otherPage).getByRole('link', { name: 'Gestión de ministerios' }),
      ).toHaveCount(0);
      await expect(
        quickAccess(otherPage).getByRole('link', { name: 'Gestión de ministerios' }),
      ).toHaveCount(0);
      await expect(
        panelNav(otherPage).getByRole('link', { name: 'Gestión de eventos' }),
      ).toBeVisible();

      await otherPage.goto('/panel/ministerios');
      await expect(
        otherPage.getByRole('heading', { name: 'No tienes acceso a esta sección' }),
      ).toBeVisible();

      // El servidor es la única autoridad (SC-005): 0 operaciones sin permiso.
      for (const [method, path, body] of [
        ['GET', '/api/v1/admin/ministerios?limit=1', undefined],
        ['POST', '/api/v1/admin/ministerios', { nameEs: 'Hack', descriptionEs: 'Intento' }],
        ['GET', `/api/v1/admin/ministerios/${someId}`, undefined],
        ['PATCH', `/api/v1/admin/ministerios/${someId}`, { nameEs: 'Hack' }],
        ['DELETE', `/api/v1/admin/ministerios/${someId}`, undefined],
      ] as const) {
        const response = await apiSend(otherPage, method, path, body);
        expect(response.status(), `${method} ${path} sin el permiso`).toBe(403);
      }

      // La subida de imágenes también exige el permiso.
      const csrf = (await otherPage.context().cookies(API_URL)).find(
        (cookie) => cookie.name === 'csrf_token',
      );
      const upload = await otherPage.request.post(`${API_URL}/api/v1/admin/ministerios/imagenes`, {
        headers: { 'X-CSRF-Token': csrf?.value ?? '' },
        multipart: {
          file: { name: 'x.jpg', mimeType: 'image/jpeg', buffer: Buffer.from([0xff, 0xd8, 0xff]) },
        },
      });
      expect(upload.status(), 'POST imagenes sin el permiso').toBe(403);

      // Y la auditoría registró los intentos denegados (result='denied').
      const denied = (await (
        await apiSend(page, 'GET', '/api/v1/admin/auditoria/acciones?limit=100')
      ).json()) as { items: { action: string; result: string }[] };
      expect(
        denied.items.some(
          (item) => item.result === 'denied' && item.action.startsWith('ministry.'),
        ),
        'la denegación queda registrada en la auditoría',
      ).toBe(true);
    } finally {
      await otherContext.close();
    }

    // ── Cuenta con **solo** `ministerios`: no entra en los otros módulos ──
    const ownContext = await browser.newContext();
    const ownPage = await ownContext.newPage();
    try {
      await loginNewAccount(ownPage, ownEmail, 'Inicial.2026', 'Definitiva.2026');
      await expect(
        panelNav(ownPage).getByRole('link', { name: 'Gestión de ministerios' }),
      ).toBeVisible();
      await expect(panelNav(ownPage).getByRole('link', { name: 'Gestión de eventos' })).toHaveCount(
        0,
      );

      await ownPage.goto('/panel/eventos');
      await expect(
        ownPage.getByRole('heading', { name: 'No tienes acceso a esta sección' }),
      ).toBeVisible();
      for (const path of ['/api/v1/admin/eventos?limit=1', '/api/v1/admin/portada']) {
        expect((await apiSend(ownPage, 'GET', path)).status(), `GET ${path} ajeno`).toBe(403);
      }
      // Su propio módulo, en cambio, responde 200.
      expect((await apiSend(ownPage, 'GET', '/api/v1/admin/ministerios?limit=1')).status()).toBe(
        200,
      );
    } finally {
      await ownContext.close();
    }
  });

  test('la auditoría registra las 8 operaciones con etiquetas legibles (SC-011)', async () => {
    await page.goto('/panel/ministerios');
    await panelNav(page).getByRole('link', { name: 'Auditoría' }).click();
    await expect(page.getByRole('heading', { name: 'Auditoría' })).toBeVisible();
    await page.getByRole('button', { name: 'Acciones administrativas' }).click();

    const table = page.getByRole('table', { name: 'Historial de acciones administrativas' });
    await expect(table).toBeVisible();
    const summary = page.getByRole('status').filter({ hasText: 'registros' });

    // Se recorren las páginas necesarias: otras specs de la tanda dejan
    // operaciones de ministerios más recientes que las de este recorrido.
    const chunks: string[] = [await table.innerText()];
    let previous = await summary.innerText();
    for (let index = 0; index < 3; index += 1) {
      const next = page.getByRole('button', { name: 'Siguiente' });
      if (!(await next.isEnabled())) {
        break;
      }
      await next.click();
      await expect(summary).not.toHaveText(previous);
      previous = await summary.innerText();
      chunks.push(await table.innerText());
    }
    const actions = chunks.join('\n');

    for (const label of [
      'Creó el ministerio',
      'Modificó el ministerio',
      'Eliminó el ministerio',
      'Publicó el ministerio',
      'Retiró el ministerio',
      'Publicó la ficha del encargado del ministerio',
      'Ocultó la ficha del encargado del ministerio',
      'Subió la imagen de un ministerio',
    ]) {
      expect(actions, `la auditoría muestra «${label}»`).toContain(label);
    }
    // Ningún código crudo a la vista.
    expect(actions).not.toContain('ministry.');
    // El objetivo se describe en palabras («Ministerios · …», targetKind `content`).
    expect(actions).toContain('Ministerios ·');
  });
});
