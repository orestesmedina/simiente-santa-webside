import { expect, test, type Browser, type Locator, type Page } from '@playwright/test';
import { ADMIN, API_URL, apiSend, loginViaUi, uniqueSuffix } from './helpers';
import {
  ensureDonacionesPermission,
  resetOfrendas,
  seedOfrendasPanelUsers,
  type OfrendasPanelUsers,
} from './ofrendasSeed';

/**
 * E2E del panel de donaciones (US2/US3, quickstart §2–§8/§12/§13): el equipo
 * **crea y publica un método desde la interfaz en menos de 2 minutos** (SC-004) y
 * el público lo ve en la **primera carga** (SC-003); un IBAN escrito con espacios
 * se guarda **normalizado** (FR-015); un IBAN mal formado y un SINPE no celular
 * (línea fija `2xxx`, 7/9 dígitos) se **rechazan sin guardar** (SC-009); la
 * **moneda es obligatoria en IBAN y no existe en SINPE** (Q2); un método **sin
 * titular** es válido y un `holderEn` sin `holderEs` avisa de qué falta; el
 * **orden** dentro del grupo reordena el público (FR-004/SC-003); los **dos
 * textos** se editan y publican **por separado** (Q10); retirar y eliminar pasan
 * por sus confirmaciones (SC-002/SC-011); **sin el permiso `donaciones`**
 * (incluida una cuenta con otros permisos de módulo) la API forzada responde
 * `403` y la interfaz lo dice sin detalles internos (SC-005); y **todo** queda
 * registrado en la auditoría de F2 con los códigos `donation.*` y su **etiqueta
 * legible** (SC-011).
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
 * sembrado por spec). Esta spec hace **4 inicios de sesión** (el administrador por
 * la interfaz y las tres cuentas de permisos cruzados), muy por debajo del tope de
 * 20/min. En este entorno el navegador no arranca desde el host —faltan las
 * librerías del sistema (`libnspr4`)—, así que los comandos se ejecutan dentro de
 * la imagen oficial de Playwright, con red del host para llegar al stack de
 * `make up`:
 *
 *   docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
 *     -v "$PWD/..":/work -w /work/frontend mcr.microsoft.com/playwright:v1.63.0-jammy \
 *     npx playwright test --config e2e/playwright.config.ts ofrendas- portada-ofrendas --workers=1
 *
 * (El stack debe estar reconstruido con F7: `docker compose up -d --build` y
 * `make db-migrate`.)
 */
test.describe.configure({ mode: 'serial' });

/* ── Textos del panel (messages.ts, FR-018: el panel va en español) ────── */

const IBAN_TABLE = 'Cuentas IBAN';
const SINPE_TABLE = 'Números de SINPE Móvil';
const IBAN_INVALID =
  'Ese número no parece una cuenta IBAN válida. Revisa la cantidad de letras y dígitos: vale tanto el IBAN de Costa Rica como uno internacional.';
const SINPE_INVALID =
  'Ese número no parece un celular de Costa Rica: son 8 dígitos, con o sin el prefijo +506.';
const REQUIRED_CURRENCY = 'Elige la moneda de la cuenta: colones o dólares.';
const DUPLICATE =
  'Ya existe esa misma cuenta o número (mismo tipo, entidad y número). Edita el que ya tienes o cambia algún dato.';
const PUBLISHED = 'Publicado en "Ofrendas y diezmos".';
const DELETED = 'Eliminado.';
const SAVED = 'Cambios guardados.';

const panelNav = (page: Page) => page.getByRole('navigation', { name: 'Navegación del panel' });
const quickAccess = (page: Page) => page.getByRole('navigation', { name: 'Accesos rápidos' });

/** Tabla de un grupo del listado por su `<caption>` (ux.md §4.3). */
const table = (page: Page, name: string) => page.getByRole('table', { name });

/** Fila del listado (tabla en escritorio) del método indicado. */
function row(page: Page, tableName: string, text: string): Locator {
  return table(page, tableName).getByRole('row').filter({ hasText: text });
}

/** Tarjeta de un texto de la sección (introductorio o de uso). */
function textCard(page: Page, title: string): Locator {
  return page.getByRole('article').filter({ hasText: title });
}

/** Abre un contexto de visitante (sin sesión) y devuelve su página. */
async function visitorPage(browser: Browser): Promise<{ close: () => Promise<void>; page: Page }> {
  const context = await browser.newContext();
  const page = await context.newPage();
  return { close: () => context.close(), page };
}

/** HTML completo de una superficie pública (para las aserciones de contenido). */
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

/** Instituciones publicadas de un grupo, en el orden de la sección. */
async function publicInstitutions(browser: Browser, group: string): Promise<string[]> {
  const visitor = await visitorPage(browser);
  try {
    await visitor.page.goto('/ofrendas');
    const region = visitor.page.getByRole('region', { name: group });
    await expect(region).toBeVisible();
    return await region.locator('h3').allInnerTexts();
  } finally {
    await visitor.close();
  }
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

test.describe('Panel de donaciones — gestión, publicación, permisos y auditoría', () => {
  test.setTimeout(600_000);

  let page: Page;
  let closeAdmin: () => Promise<void>;
  let suffix: string;
  let institutionA: string;
  let institutionB: string;
  let institutionC: string;
  let users: OfrendasPanelUsers;

  /** IBAN costarricense válido de 22 caracteres, escrito con guiones y minúsculas. */
  const IBAN_A_INPUT = 'cr55-0152-0200-1234-5678-90';
  const IBAN_A = 'CR55015202001234567890';
  const IBAN_B_INPUT = 'CR66 0152 0200 1234 5678 91';

  test.beforeAll(async ({ browser }) => {
    suffix = uniqueSuffix();
    institutionA = `Banco de prueba A ${suffix}`;
    institutionB = `Banco de prueba B ${suffix}`;
    institutionC = `SINPE de prueba C ${suffix}`;

    const context = await browser.newContext();
    page = await context.newPage();
    closeAdmin = () => context.close();

    // Login real por la interfaz con la cuenta que tiene el permiso `donaciones`.
    await loginViaUi(page, ADMIN.email, ADMIN.password);
    await expect(page).toHaveURL(/\/panel$/);
    await ensureDonacionesPermission(page);
    await resetOfrendas(page);
  });

  test.afterAll(async () => {
    // La base es compartida con F3/F4/F6/F10: se deja el módulo vacío.
    if (page) {
      await resetOfrendas(page);
    }
    await closeAdmin();
  });

  test('crear y publicar un IBAN en menos de 2 minutos; el público lo ve en la primera carga (SC-003/SC-004)', async ({
    browser,
  }) => {
    // ── El módulo aparece en navegación y accesos rápidos (FR-011) ────────
    await page.goto('/panel');
    await expect(
      quickAccess(page).getByRole('link', { name: 'Gestión de donaciones' }),
    ).toHaveAttribute('href', '/panel/donaciones');
    await panelNav(page).getByRole('link', { name: 'Gestión de donaciones' }).click();
    await expect(page.getByRole('heading', { name: 'Gestión de donaciones' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Textos de la sección' })).toBeVisible();
    // Estados vacíos por bloque (ux.md §4.6).
    await expect(page.getByText('Todavía no hay cuentas IBAN')).toBeVisible();
    await expect(page.getByText('Todavía no hay números de SINPE Móvil')).toBeVisible();

    // ── SC-004: alta + publicación desde la interfaz, cronometrada ────────
    const started = Date.now();
    await page.getByRole('button', { name: 'Agregar cuenta' }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByRole('heading', { name: 'Agregar cuenta' })).toBeVisible();

    await dialog.getByLabel('Entidad financiera').fill(institutionA);
    // Se escribe con espacios (y guiones): el servidor lo normaliza (FR-015).
    await dialog.getByLabel('Número de cuenta (IBAN)').fill(IBAN_A_INPUT);
    await dialog.getByLabel('Moneda').selectOption({ label: 'Colones' });
    await dialog.getByLabel('Orden').fill('20');
    await dialog.getByRole('button', { name: 'Agregar cuenta' }).click();

    await expect(page.getByText('Cuenta creada.')).toBeVisible();
    await expect(page.getByText('Quedó en borrador')).toBeVisible();
    await expect(row(page, IBAN_TABLE, institutionA).getByText('Borrador')).toBeVisible();

    // Publicar es una acción directa, sin confirmación (FR-013).
    await row(page, IBAN_TABLE, institutionA).getByRole('button', { name: 'Publicar' }).click();
    await expect(page.getByText(PUBLISHED)).toBeVisible();
    await expect(row(page, IBAN_TABLE, institutionA).getByText('Publicado')).toBeVisible();

    // ── SC-003: el visitante lo ve en la **primera carga**, normalizado ───
    const html = await publicHtml(browser, '/ofrendas');
    expect(html).toContain(institutionA);
    expect(html).toContain(IBAN_A);
    expect(html).not.toContain(IBAN_A_INPUT);

    const elapsed = Date.now() - started;
    expect(elapsed, `crear y publicar tardó ${elapsed} ms (tope 120 000)`).toBeLessThan(120_000);
  });

  test('rechazos sin guardar: IBAN y SINPE mal formados, moneda obligatoria y ausente (SC-009/Q2)', async () => {
    await page.goto('/panel/donaciones');

    // ── IBAN mal formado (21 caracteres): rechazado sin guardar ──────────
    await page.getByRole('button', { name: 'Agregar cuenta' }).click();
    let dialog = page.getByRole('dialog');
    await dialog.getByLabel('Entidad financiera').fill(`Banco inválido ${suffix}`);
    await dialog.getByLabel('Número de cuenta (IBAN)').fill('CR05 0152 0200 1234 5678 9');
    await dialog.getByLabel('Moneda').selectOption({ label: 'Colones' });
    await dialog.getByRole('button', { name: 'Agregar cuenta' }).click();
    await expect(dialog.getByText(IBAN_INVALID).first()).toBeVisible();
    // Nada se guardó a medias: el diálogo sigue abierto.
    await expect(dialog).toBeVisible();
    await dialog.getByRole('button', { name: 'Cancelar' }).click();

    // ── Moneda obligatoria en IBAN: número válido y sin elegir moneda ────
    await page.getByRole('button', { name: 'Agregar cuenta' }).click();
    dialog = page.getByRole('dialog');
    await dialog.getByLabel('Entidad financiera').fill(`Banco sin moneda ${suffix}`);
    await dialog.getByLabel('Número de cuenta (IBAN)').fill('CR05 0152 0200 1234 5678 90');
    await dialog.getByRole('button', { name: 'Agregar cuenta' }).click();
    await expect(dialog.getByText(REQUIRED_CURRENCY).first()).toBeVisible();
    await dialog.getByRole('button', { name: 'Cancelar' }).click();

    // ── SINPE: el formulario **no ofrece** la moneda (Q2) ────────────────
    await page.getByRole('button', { name: 'Agregar número' }).click();
    dialog = page.getByRole('dialog');
    await expect(dialog.getByLabel('Moneda')).toHaveCount(0);
    await dialog.getByLabel('Entidad financiera').fill(`SINPE inválido ${suffix}`);

    // Línea fija `2xxx` rechazada (default revisable 6/7/8, RG7-3).
    await dialog.getByLabel('Número de celular (SINPE Móvil)').fill('22222222');
    await dialog.getByRole('button', { name: 'Agregar número' }).click();
    await expect(dialog.getByText(SINPE_INVALID).first()).toBeVisible();
    // 7 dígitos rechazados.
    await dialog.getByLabel('Número de celular (SINPE Móvil)').fill('8555444');
    await dialog.getByRole('button', { name: 'Agregar número' }).click();
    await expect(dialog.getByText(SINPE_INVALID).first()).toBeVisible();
    await dialog.getByRole('button', { name: 'Cancelar' }).click();

    // Ninguna de las altas inválidas quedó guardada.
    const list = (await (
      await page.request.get(`${API_URL}/api/v1/admin/donaciones?limit=100`)
    ).json()) as { items: { institution: string }[] };
    const institutions = list.items.map((item) => item.institution);
    expect(institutions).not.toContain(`Banco inválido ${suffix}`);
    expect(institutions).not.toContain(`Banco sin moneda ${suffix}`);
    expect(institutions).not.toContain(`SINPE inválido ${suffix}`);
  });

  test('un método sin titular es válido; `holderEn` sin `holderEs` avisa qué falta', async () => {
    await page.goto('/panel/donaciones');

    // ── Alta sin titular: válido (US2 esc. 8) ────────────────────────────
    await page.getByRole('button', { name: 'Agregar cuenta' }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('Entidad financiera').fill(institutionB);
    await dialog.getByLabel('Número de cuenta (IBAN)').fill(IBAN_B_INPUT);
    await dialog.getByLabel('Moneda').selectOption({ label: 'Dólares' });
    await dialog.getByLabel('Orden').fill('30');
    await dialog.getByRole('button', { name: 'Agregar cuenta' }).click();
    await expect(page.getByText('Cuenta creada.')).toBeVisible();
    await expect(row(page, IBAN_TABLE, institutionB).getByText('—')).toBeVisible();

    // ── Edición: el tipo es etiqueta fija, sin selector (D-9) ────────────
    await row(page, IBAN_TABLE, institutionB).getByRole('button', { name: 'Editar' }).click();
    const edit = page.getByRole('dialog');
    await expect(edit.getByRole('heading', { name: 'Editar Cuenta IBAN' })).toBeVisible();
    await expect(edit.getByRole('combobox', { name: 'Tipo' })).toHaveCount(0);
    await expect(edit.getByText('El tipo no se cambia aquí')).toBeVisible();

    // ── Titular en inglés sin su versión en español: error por campo ─────
    await edit.getByRole('button', { name: 'English (opcional)' }).click();
    await edit.getByLabel('Titular de la cuenta (English)').fill(`English holder ${suffix}`);
    await edit.getByRole('button', { name: 'Guardar cambios' }).click();
    await expect(edit.getByText(/Falta la versión en español/).first()).toBeVisible();
    await expect(edit).toBeVisible();
    await edit.getByRole('button', { name: 'Cancelar' }).click();

    // El método sigue sin titular (nada se guardó).
    const list = (await (
      await page.request.get(`${API_URL}/api/v1/admin/donaciones?limit=100`)
    ).json()) as { items: { institution: string; holderEs: string | null }[] };
    const stored = list.items.find((item) => item.institution === institutionB);
    expect(stored?.holderEs ?? null).toBeNull();
  });

  test('SINPE sin `+506` se guarda normalizado y un duplicado avisa sin guardar (FR-016/P7-10)', async () => {
    await page.goto('/panel/donaciones');

    // ── SINPE escrito sin el prefijo: el servidor lo normaliza (FR-016) ──
    await page.getByRole('button', { name: 'Agregar número' }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('Entidad financiera').fill(institutionC);
    await dialog.getByLabel('Número de celular (SINPE Móvil)').fill('8555 4444');
    await dialog.getByRole('button', { name: 'Agregar número' }).click();
    await expect(page.getByText('Número de SINPE creado.')).toBeVisible();
    await expect(row(page, SINPE_TABLE, institutionC)).toContainText('+50685554444');
    // El bloque SINPE **no** tiene columna Moneda (Q2).
    await expect(
      table(page, SINPE_TABLE).getByRole('columnheader', { name: 'Moneda' }),
    ).toHaveCount(0);

    // ── Duplicado exacto: mismo tipo + número normalizado → aviso claro ──
    await page.getByRole('button', { name: 'Agregar cuenta' }).click();
    const dup = page.getByRole('dialog');
    await dup.getByLabel('Entidad financiera').fill(`Copia de A ${suffix}`);
    // El mismo número, escrito de otra forma: normaliza al mismo valor (P7-10).
    await dup.getByLabel('Número de cuenta (IBAN)').fill('CR550152-0200-1234-5678-90');
    await dup.getByLabel('Moneda').selectOption({ label: 'Colones' });
    await dup.getByRole('button', { name: 'Agregar cuenta' }).click();
    await expect(dup.getByText(DUPLICATE)).toBeVisible();
    // Sin perder lo escrito y sin guardar.
    await expect(dup.getByLabel('Entidad financiera')).toHaveValue(`Copia de A ${suffix}`);
    await dup.getByRole('button', { name: 'Cancelar' }).click();

    const list = (await (
      await page.request.get(`${API_URL}/api/v1/admin/donaciones?limit=100`)
    ).json()) as { items: { institution: string }[] };
    expect(list.items.map((item) => item.institution)).not.toContain(`Copia de A ${suffix}`);
  });

  test('el orden dentro del grupo reordena el panel y el público (FR-004/SC-003)', async ({
    browser,
  }) => {
    await page.goto('/panel/donaciones');

    // B pasa a ir antes que A (orden menor), y se publica para verlo en público.
    await row(page, IBAN_TABLE, institutionB).getByRole('button', { name: 'Editar' }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('Orden').fill('10');
    await dialog.getByRole('button', { name: 'Guardar cambios' }).click();
    await expect(page.getByText(SAVED)).toBeVisible();

    await row(page, IBAN_TABLE, institutionB).getByRole('button', { name: 'Publicar' }).click();
    await expect(page.getByText(PUBLISHED)).toBeVisible();

    // El panel muestra el orden del visitante (iban por `sortOrder`).
    const rows = table(page, IBAN_TABLE).getByRole('row');
    await expect(rows.nth(1)).toContainText(institutionB);
    await expect(rows.nth(2)).toContainText(institutionA);

    // …y el público, exactamente el mismo orden (SC-003).
    expect(await publicInstitutions(browser, IBAN_TABLE)).toEqual([institutionB, institutionA]);
  });

  test('los dos textos se editan y publican por separado (Q10/US3)', async ({ browser }) => {
    await page.goto('/panel/donaciones');

    const introText = `Tu aporte sostiene la obra 🙏\n¡Gracias por donar! (${suffix})`;
    const usageText = `En obra social, templo y misiones 🎉 (${suffix})`;

    // ── Introductorio: guardar **no** lo publica (Q10) ───────────────────
    await textCard(page, 'Introductorio').getByRole('button', { name: 'Editar' }).click();
    let dialog = page.getByRole('dialog');
    await expect(dialog.getByRole('heading', { name: 'Texto introductorio' })).toBeVisible();
    await dialog.getByLabel('Texto (Español)').fill(introText);
    await dialog.getByRole('button', { name: 'Guardar cambios' }).click();
    await expect(page.getByText(SAVED)).toBeVisible();
    await expect(textCard(page, 'Introductorio').getByText('Borrador')).toBeVisible();

    // Publicar es una decisión aparte, desde su tarjeta.
    await textCard(page, 'Introductorio').getByRole('button', { name: 'Publicar' }).click();
    await expect(page.getByText(PUBLISHED)).toBeVisible();
    await expect(textCard(page, 'Introductorio').getByText('Publicado')).toBeVisible();

    // ── De uso: mismo formulario parametrizado, estado independiente ─────
    await textCard(page, '¿En qué usamos las donaciones?')
      .getByRole('button', { name: 'Editar' })
      .click();
    dialog = page.getByRole('dialog');
    await expect(
      dialog.getByRole('heading', { name: 'Texto de uso de las donaciones' }),
    ).toBeVisible();
    await dialog.getByLabel('Texto (Español)').fill(usageText);
    await dialog.getByRole('button', { name: 'Guardar cambios' }).click();
    await expect(page.getByText(SAVED)).toBeVisible();
    await textCard(page, '¿En qué usamos las donaciones?')
      .getByRole('button', { name: 'Publicar' })
      .click();
    await expect(page.getByText(PUBLISHED)).toBeVisible();

    // ── El público ve los dos textos (el introductorio ya estaba publicado
    //    mientras el de uso seguía en borrador) ────────────────────────────
    const html = await publicHtml(browser, '/ofrendas');
    expect(html).toContain(introText.split('\n')[0]);
    expect(html).toContain(usageText);
  });

  test('retirar y eliminar con confirmación; el público deja de verlos (SC-002/SC-011)', async ({
    browser,
  }) => {
    await page.goto('/panel/donaciones');

    // ── Retirar de la sección: confirmación suave (FR-012) ───────────────
    await row(page, IBAN_TABLE, institutionB)
      .getByRole('button', { name: 'Retirar de la sección' })
      .click();
    const retireDialog = page.getByRole('dialog');
    await expect(
      retireDialog.getByRole('heading', { name: 'Retirar de la sección' }),
    ).toBeVisible();
    await retireDialog.getByRole('button', { name: 'Retirar' }).click();
    await expect(page.getByText('Retirado de la sección')).toBeVisible();
    await expect(row(page, IBAN_TABLE, institutionB).getByText('Borrador')).toBeVisible();
    expect(await publicInstitutions(browser, IBAN_TABLE)).not.toContain(institutionB);

    // ── Eliminar: confirmación explícita de no recuperable (FR-017) ──────
    await row(page, IBAN_TABLE, institutionA).getByRole('button', { name: 'Eliminar' }).click();
    const deleteDialog = page.getByRole('dialog');
    await expect(deleteDialog).toContainText('se eliminará para siempre');
    await deleteDialog.getByRole('button', { name: 'Eliminar' }).click();
    await expect(page.getByText(DELETED)).toBeVisible();
    await expect(row(page, IBAN_TABLE, institutionA)).toHaveCount(0);

    const html = await publicHtml(browser, '/ofrendas');
    expect(html).not.toContain(institutionA);
    expect(html).not.toContain(IBAN_A);
  });

  test('permisos cruzados: sin el permiso `donaciones` no se entra ni se opera (SC-005)', async ({
    browser,
  }) => {
    users = await seedOfrendasPanelUsers(page, suffix);

    const existing = (await (
      await page.request.get(`${API_URL}/api/v1/admin/donaciones?limit=100`)
    ).json()) as { items: { id: string }[] };
    const someId = existing.items[0]?.id ?? '00000000-0000-0000-0000-000000000000';

    // ── Cuenta con **otros** permisos de módulo (no `donaciones`) ────────
    const otherContext = await browser.newContext();
    const otherPage = await otherContext.newPage();
    try {
      await loginNewAccount(otherPage, users.otherModules.email, 'Inicial.2026', 'Definitiva.2026');
      await expect(
        panelNav(otherPage).getByRole('link', { name: 'Gestión de donaciones' }),
      ).toHaveCount(0);
      await expect(
        quickAccess(otherPage).getByRole('link', { name: 'Gestión de donaciones' }),
      ).toHaveCount(0);
      await expect(
        panelNav(otherPage).getByRole('link', { name: 'Gestión de eventos' }),
      ).toBeVisible();

      await otherPage.goto('/panel/donaciones');
      await expect(otherPage.getByText('No tienes acceso a esta sección')).toBeVisible();

      // El servidor es la única autoridad (SC-005): 0 operaciones sin permiso.
      for (const [method, path, body] of [
        ['GET', '/api/v1/admin/donaciones?limit=1', undefined],
        [
          'POST',
          '/api/v1/admin/donaciones',
          { kind: 'iban', number: 'CR55015202001234567890', institution: 'Hack', currency: 'CRC' },
        ],
        ['GET', `/api/v1/admin/donaciones/${someId}`, undefined],
        ['PATCH', `/api/v1/admin/donaciones/${someId}`, { institution: 'Hack' }],
        ['DELETE', `/api/v1/admin/donaciones/${someId}`, undefined],
        ['GET', '/api/v1/admin/donaciones/textos', undefined],
        ['PUT', '/api/v1/admin/donaciones/textos/intro', { textEs: 'Hack' }],
      ] as const) {
        const response = await apiSend(otherPage, method, path, body);
        expect(response.status(), `${method} ${path} sin el permiso`).toBe(403);
      }

      // Y la auditoría registró los intentos denegados (result='denied').
      const denied = (await (
        await apiSend(page, 'GET', '/api/v1/admin/auditoria/acciones?limit=100')
      ).json()) as { items: { action: string; result: string }[] };
      expect(
        denied.items.some(
          (item) => item.result === 'denied' && item.action.startsWith('donation.'),
        ),
        'la denegación queda registrada en la auditoría',
      ).toBe(true);
    } finally {
      await otherContext.close();
    }

    // ── Cuenta **sin** permisos: ni accesos ni módulo ────────────────────
    const noneContext = await browser.newContext();
    const nonePage = await noneContext.newPage();
    try {
      await loginNewAccount(nonePage, users.none.email, 'Inicial.2026', 'Definitiva.2026');
      await expect(
        quickAccess(nonePage).getByRole('link', { name: 'Gestión de donaciones' }),
      ).toHaveCount(0);
      await nonePage.goto('/panel/donaciones');
      await expect(nonePage.getByText('No tienes acceso a esta sección')).toBeVisible();
      expect(
        (await apiSend(nonePage, 'GET', '/api/v1/admin/donaciones?limit=1')).status(),
        'GET donaciones sin ningún permiso',
      ).toBe(403);
    } finally {
      await noneContext.close();
    }

    // ── Cuenta con **solo** `donaciones`: su módulo sí, los ajenos no ────
    const ownContext = await browser.newContext();
    const ownPage = await ownContext.newPage();
    try {
      await loginNewAccount(ownPage, users.donaciones.email, 'Inicial.2026', 'Definitiva.2026');
      await expect(
        panelNav(ownPage).getByRole('link', { name: 'Gestión de donaciones' }),
      ).toBeVisible();
      await expect(panelNav(ownPage).getByRole('link', { name: 'Gestión de eventos' })).toHaveCount(
        0,
      );
      await ownPage.goto('/panel/eventos');
      await expect(ownPage.getByText('No tienes acceso a esta sección')).toBeVisible();
      for (const path of ['/api/v1/admin/eventos?limit=1', '/api/v1/admin/ministerios?limit=1']) {
        expect((await apiSend(ownPage, 'GET', path)).status(), `GET ${path} ajeno`).toBe(403);
      }
      // Su propio módulo, en cambio, responde 200.
      expect((await apiSend(ownPage, 'GET', '/api/v1/admin/donaciones?limit=1')).status()).toBe(
        200,
      );
    } finally {
      await ownContext.close();
    }
  });

  test('la auditoría registra las operaciones con etiquetas legibles (SC-011)', async () => {
    await page.goto('/panel/donaciones');
    await panelNav(page).getByRole('link', { name: 'Auditoría' }).click();
    await expect(page.getByRole('heading', { name: 'Auditoría' })).toBeVisible();
    await page.getByRole('button', { name: 'Acciones administrativas' }).click();

    const actionsTable = page.getByRole('table', {
      name: 'Historial de acciones administrativas',
    });
    await expect(actionsTable).toBeVisible();
    const summary = page.getByRole('status').filter({ hasText: 'registros' });

    // Se recorren las páginas necesarias: otras specs de la tanda dejan
    // operaciones más recientes que las de este recorrido.
    const chunks: string[] = [await actionsTable.innerText()];
    let previous = await summary.innerText();
    for (let index = 0; index < 3; index += 1) {
      const next = page.getByRole('button', { name: 'Siguiente' });
      if (!(await next.isEnabled())) {
        break;
      }
      await next.click();
      await expect(summary).not.toHaveText(previous);
      previous = await summary.innerText();
      chunks.push(await actionsTable.innerText());
    }
    const actions = chunks.join('\n');

    for (const label of [
      'Creó una cuenta o número de donaciones',
      'Modificó una cuenta o número de donaciones',
      'Eliminó una cuenta o número de donaciones',
      'Publicó una cuenta o número de donaciones',
      'Retiró una cuenta o número de donaciones',
      'Modificó el texto introductorio o el de uso',
      'Publicó el texto introductorio o el de uso',
    ]) {
      expect(actions, `la auditoría muestra «${label}»`).toContain(label);
    }
    // Ningún código crudo a la vista.
    expect(actions).not.toContain('donation.');
    // El objetivo se describe en palabras («Donaciones · …», targetKind `content`).
    expect(actions).toContain('Donaciones ·');
  });
});
