import { expect, test, type Browser, type Locator, type Page } from '@playwright/test';
import { ADMIN, apiSend, loginViaUi, uniqueSuffix } from './helpers';
import {
  fixtureByKey,
  resetMedios,
  seedMedios,
  seedMediosPanelUsers,
  type MediosPanelUsers,
  type MediosSeed,
} from './mediosSeed';

/**
 * E2E del panel de medios (US2/US3, quickstart §2/§5/§6/§7/§8/§12) sobre datos
 * **sembrados por SQL** (RG9-7): el listado muestra el estado, los formatos, el
 * origen y las marcas de cada ficha con sus acciones; **ocultar** la saca del
 * público y **mostrar** la devuelve a su lugar por fecha (FR-009/FR-015); la
 * **descripción** se ajusta (es/en) y se publica, y el inglés sin su español no
 * guarda nada (FR-020/FR-023); el par marcado se **une** con vista previa y
 * **doble confirmación** (la absorbida desaparece del listado, FR-014/FR-019) o se
 * resuelve como **«no son el mismo sermón»** (US3 esc. 2); los **dos ejes** de
 * marcas conviven y se resuelven por separado, y una marca resuelta queda
 * «Resuelta» (P9-19/FR-017); el **estado de la precedencia A3** se presenta con
 * su marca y **«Mostrar de nuevo»** funciona (con su confirmación suave); sin el
 * permiso **`medios`** la API forzada responde `403` y la interfaz lo dice sin
 * detalles internos (SC-009); y **todo** queda en la auditoría de F2 con los
 * códigos `media.*` y su **etiqueta legible** (SC-010).
 *
 * **Esta spec nunca pulsa «Buscar episodios nuevos»** (RG9-7): la búsqueda
 * depende de la red real y dispararía la comprobación de disponibilidad contra
 * datos sembrados. La importación la cubren las pruebas de Go con dobles
 * (T920/T921). Como red de seguridad, la petición de búsqueda está **interceptada
 * y abortada**: si algún test la pulsara, el fallo sería visible (y la última
 * prueba lo comprueba explícitamente).
 *
 * Cómo se ejecutan (la base, el rate-limit de login y el navegador son
 * compartidos por toda la suite):
 *
 *   # Tanda 6 — medios (F9): las cuatro specs comparten la siembra/limpieza del
 *   # módulo (y la carátula subida), así que van juntas y con un solo worker.
 *   cd frontend && npx playwright test --config e2e/playwright.config.ts medios- portada-medios --workers=1
 *
 * La suite completa se reparte en **seis tandas** porque `POST
 * /api/v1/auth/login` limita a 20 intentos por minuto y por IP (ver
 * `medios-publica.spec.ts`). Esta spec hace **4 inicios de sesión** (el
 * administrador por la interfaz y las tres cuentas de permisos cruzados), muy por
 * debajo del tope de 20/min.
 */
test.describe.configure({ mode: 'serial' });

/* ── Textos del panel (messages.ts, FR-024: el panel va en español) ────── */

const TITLE = 'Medios';
const SEARCH_ZONE = 'Desde YouTube y el podcast';
const SEARCH_ACTION = 'Buscar episodios nuevos';
const SEARCH_SUPPORT =
  'Es seguro pulsarlo todas las veces que quieras: solo agrega lo que falte, nunca borra y nunca cambia lo que ya decidiste.';
const FILTER_HELP =
  'Las marcas las pone el sistema cuando algo necesita tu revisión; se resuelven en cada fila.';
const TABLE = 'Episodios de medios importados';
const NO_PERMISSION = 'No tienes acceso a esta sección';

const panelNav = (page: Page) => page.getByRole('navigation', { name: 'Navegación del panel' });
const quickAccess = (page: Page) => page.getByRole('navigation', { name: 'Accesos rápidos' });

/** Fila del listado (tabla en escritorio) de la ficha indicada. */
function row(page: Page, title: string): Locator {
  return page.getByRole('table', { name: TABLE }).getByRole('row').filter({ hasText: title });
}

/** Todas las filas del catálogo (sin la cabecera). */
function rows(page: Page): Locator {
  return page
    .getByRole('table', { name: TABLE })
    .getByRole('row')
    .filter({ has: page.locator('td') });
}

/**
 * Comprueba que la fila **no tiene ninguna marca pendiente**: el explicador de
 * una línea solo se pinta con la marca **abierta** (`MarcaDeRevision`); una marca
 * ya resuelta se conserva como registro («Resuelta» + cuál era) sin explicador.
 */
async function sinMarcasPendientes(fila: Locator): Promise<void> {
  await expect(
    fila.getByText(
      /El sistema no pudo estar seguro|Se comprobó: el enlace de esta grabación|Salió de la ventana del/,
    ),
  ).toHaveCount(0);
}

/** Abre un contexto de visitante (sin sesión) y devuelve su página. */
async function visitorPage(browser: Browser): Promise<{ close: () => Promise<void>; page: Page }> {
  const context = await browser.newContext();
  const page = await context.newPage();
  return { close: () => context.close(), page };
}

/**
 * Nombres de los episodios publicables, en el orden que ve el visitante. La
 * sección entrega tandas de 12 con «Cargar más», así que se agotan las tandas
 * para poder comparar el orden **completo** entre estados.
 */
async function publicTitles(browser: Browser): Promise<string[]> {
  const visitor = await visitorPage(browser);
  try {
    await visitor.page.goto('/medios');
    const list = visitor.page.getByRole('list', { name: 'Lista de episodios de medios' });
    await expect(list).toBeVisible();
    for (let tanda = 0; tanda < 5; tanda += 1) {
      const mas = visitor.page.getByRole('button', { name: 'Cargar más' });
      if ((await mas.count()) === 0) {
        break;
      }
      const antes = await list.getByRole('listitem').count();
      await mas.click();
      await expect
        .poll(() => list.getByRole('listitem').count(), { message: 'la tanda se agregó' })
        .toBeGreaterThan(antes);
    }
    return await list.locator('h3').allInnerTexts();
  } finally {
    await visitor.close();
  }
}

/** HTML completo de la sección pública (para las aserciones de contenido). */
async function publicHtml(browser: Browser): Promise<string> {
  const visitor = await visitorPage(browser);
  try {
    await visitor.page.goto('/medios');
    await expect(
      visitor.page.getByRole('list', { name: 'Lista de episodios de medios' }),
    ).toBeVisible();
    return await visitor.page.evaluate(() => document.documentElement.outerHTML);
  } finally {
    await visitor.close();
  }
}

/** `true` si el bloque de la portada contiene el texto indicado. */
async function portadaContiene(browser: Browser, text: string): Promise<boolean> {
  const visitor = await visitorPage(browser);
  try {
    await visitor.page.goto('/');
    await expect(visitor.page.locator('#proximamente')).toBeVisible();
    return (await visitor.page.locator('#medios').count()) > 0
      ? (await visitor.page.locator('#medios').getByText(text).count()) > 0
      : false;
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

test.describe('Panel de medios — gestión, unión, marcas, permisos y auditoría', () => {
  test.setTimeout(600_000);

  let page: Page;
  let closeAdmin: () => Promise<void>;
  let seed: MediosSeed;
  let users: MediosPanelUsers;
  /** Peticiones a «Buscar episodios nuevos» que la spec NO debe hacer (RG9-7). */
  const busquedas: string[] = [];

  test.beforeAll(async ({ browser }) => {
    const context = await browser.newContext();
    page = await context.newPage();
    closeAdmin = () => context.close();

    // Login real por la interfaz con la cuenta que tiene el permiso `medios`.
    await loginViaUi(page, ADMIN.email, ADMIN.password);
    await expect(page).toHaveURL(/\/panel$/);

    // Red de seguridad de RG9-7: la búsqueda está interceptada y abortada.
    await page.route('**/api/v1/admin/medios/buscar-nuevos', async (route) => {
      busquedas.push(route.request().url());
      await route.abort();
    });

    seed = await seedMedios(page);
  });

  test.afterAll(async () => {
    // Se borra lo sembrado y se devuelven a `visible` las fichas reales que la
    // siembra ocultó temporalmente (ver `mediosSeed.ts`).
    resetMedios();
    await closeAdmin();
  });

  test('el listado muestra estado, formatos, origen, fuentes y marcas (FR-019/⟲ A2)', async () => {
    await page.goto('/panel/medios');
    await expect(page.getByRole('heading', { level: 1, name: TITLE })).toBeVisible();

    // ── La zona de búsqueda está disponible, sin pulsarla nunca (RG9-7) ──
    await expect(page.getByRole('heading', { name: SEARCH_ZONE })).toBeVisible();
    await expect(page.getByRole('button', { name: SEARCH_ACTION })).toBeEnabled();
    await expect(page.getByText(SEARCH_SUPPORT)).toBeVisible();
    expect(busquedas, 'la spec nunca pulsa «Buscar episodios nuevos»').toEqual([]);

    // ── Ficha con los dos enlaces: estado, formatos, origen y fuentes ────
    const dosEnlaces = fixtureByKey(seed, 'dos-enlaces');
    const ficha = row(page, dosEnlaces.title);
    await expect(ficha.getByText('Visible', { exact: true })).toBeVisible();
    await expect(ficha.getByText('Video', { exact: true })).toBeVisible();
    await expect(ficha.getByText('Audio', { exact: true })).toBeVisible();
    await expect(ficha.getByText('De YouTube y del podcast', { exact: true })).toBeVisible();
    await expect(ficha).toContainText(/Importado el \d{2}\/\d{2}\/\d{4}/);
    // Las dos procedencias con el título de **cada** feed (FR-019/A2). El
    // título del feed del video coincide con el nombre publicado de la ficha, así
    // que la fila de la fuente se localiza por su enlace.
    await expect(ficha.getByText('YouTube', { exact: true })).toBeVisible();
    await expect(ficha.getByText('el podcast', { exact: true })).toBeVisible();
    for (const source of dosEnlaces.sources) {
      const fuente = ficha.locator('li').filter({ hasText: source.url });
      await expect(fuente).toContainText(source.feedTitle);
      const enlace = ficha.locator(`a[href="${source.url}"]`);
      await expect(enlace).toHaveAttribute('target', '_blank');
      await expect(enlace).toHaveAttribute('rel', 'noopener noreferrer');
    }

    // ── Ficha de un solo formato y ficha oculta por el equipo ────────────
    const soloVideo = row(page, fixtureByKey(seed, 'solo-video').title);
    await expect(soloVideo.getByText('Video', { exact: true })).toBeVisible();
    await expect(soloVideo.getByText('Audio', { exact: true })).toHaveCount(0);
    const oculta = row(page, fixtureByKey(seed, 'oculta-equipo').title);
    await expect(oculta.getByText('Oculto', { exact: true })).toBeVisible();

    // ── Marcas con su explicador y sus acciones ──────────────────────────
    const parBase = row(page, fixtureByKey(seed, 'par-union-base').title);
    await expect(parBase.getByText('Posible duplicado', { exact: true })).toBeVisible();
    await expect(parBase).toContainText(
      'El sistema no pudo estar seguro de si estas dos entradas son el mismo sermón.',
    );
    await expect(parBase.getByRole('button', { name: 'Revisar duplicado' })).toBeVisible();

    const noDisponible = row(page, fixtureByKey(seed, 'marca-unavailable').title);
    await expect(noDisponible.getByText('No disponible', { exact: true })).toBeVisible();
    await expect(noDisponible).toContainText(
      'Se comprobó: el enlace de esta grabación ya no funciona (borrada o privada). El público no la ve.',
    );
    await expect(noDisponible.getByRole('button', { name: 'Mostrar de nuevo' })).toBeVisible();
    await expect(noDisponible.getByRole('button', { name: 'Mantener oculto' })).toBeVisible();
    // Con una marca de disponibilidad, la acción genérica no se duplica.
    await expect(noDisponible.getByRole('button', { name: 'Mostrar', exact: true })).toHaveCount(0);

    const vivo = row(page, fixtureByKey(seed, 'vivo-fuera').title);
    await expect(vivo.getByText('Ya no aparece en el feed', { exact: true })).toBeVisible();
    await expect(vivo).toContainText(
      'Su enlace sigue funcionando y sigue visible para el público.',
    );
    await expect(vivo.getByText('Visible', { exact: true })).toBeVisible();
    await expect(vivo.getByRole('button', { name: 'Mantener visible' })).toBeVisible();

    // La comprobación que no se completa lo dice, y la ficha sigue visible.
    const inconcluso = row(page, fixtureByKey(seed, 'inconcluso').title);
    await expect(inconcluso).toContainText(
      'No se pudo comprobar si su enlace sigue funcionando; mientras tanto, sigue visible para el público.',
    );

    // Una marca ya resuelta se conserva como registro y no ofrece resoluciones.
    const resuelta = row(page, fixtureByKey(seed, 'marca-resuelta').title);
    await expect(resuelta.getByText('Resuelta', { exact: true })).toBeVisible();
    await expect(resuelta.getByText('Ya no aparece en el feed', { exact: true })).toBeVisible();
    await expect(resuelta.getByRole('button', { name: 'Mantener visible' })).toHaveCount(0);
    await expect(resuelta.getByRole('button', { name: 'Mostrar de nuevo' })).toHaveCount(0);

    // Los **dos ejes** conviven en una misma ficha (P9-19).
    const mixta = row(page, fixtureByKey(seed, 'par-mixto-base').title);
    await expect(mixta.getByText('Posible duplicado', { exact: true })).toBeVisible();
    await expect(mixta.getByText('Ya no aparece en el feed', { exact: true })).toBeVisible();

    // ── Filtro de marcas con sus conteos (⟲ D-10) ───────────────────────
    const pendientes = [...seed.publicables, ...seed.ocultas];
    const conteos = {
      duplicate: pendientes.filter((item) => item.duplicateOfKey).length,
      unavailable: pendientes.filter((item) => item.availabilityMark === 'unavailable').length,
      missing: pendientes.filter(
        (item) => item.availabilityMark === 'missing_from_feed' && !item.availabilityResolved,
      ).length,
    };
    const filtro = page.getByRole('group', { name: 'Solo lo pendiente de revisión' });
    await expect(
      page.getByRole('heading', { name: 'Solo lo pendiente de revisión' }),
    ).toBeVisible();
    await expect(page.getByText(FILTER_HELP)).toBeVisible();
    await expect(filtro.getByRole('button', { name: 'Todo' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
    await expect(
      filtro.getByRole('button', { name: `Posible duplicado (${conteos.duplicate})` }),
    ).toBeVisible();
    await expect(
      filtro.getByRole('button', { name: `No disponible (${conteos.unavailable})` }),
    ).toBeVisible();
    await expect(
      filtro.getByRole('button', { name: `Ya no aparece en el feed (${conteos.missing})` }),
    ).toBeVisible();

    // ── El filtro es de vista: con «No disponible» queda solo esa ficha ──
    const total = await rows(page).count();
    await filtro.getByRole('button', { name: `No disponible (${conteos.unavailable})` }).click();
    await expect(rows(page)).toHaveCount(conteos.unavailable);
    await expect(row(page, fixtureByKey(seed, 'marca-unavailable').title)).toBeVisible();
    await filtro.getByRole('button', { name: 'Todo' }).click();
    await expect(rows(page)).toHaveCount(total);

    // La absorbida no aparece en ninguna lista (FR-019).
    await expect(page.getByText(fixtureByKey(seed, 'absorbida').title)).toHaveCount(0);
  });

  test('ocultar y mostrar: el público deja de verlo y vuelve a su lugar por fecha (FR-009/FR-015)', async ({
    browser,
  }) => {
    await page.goto('/panel/medios');
    const soloAudio = fixtureByKey(seed, 'solo-audio');
    const ordenInicial = await publicTitles(browser);
    expect(ordenInicial).toContain(soloAudio.title);
    expect(await portadaContiene(browser, soloAudio.title)).toBe(true);

    // ── Ocultar: directo, sin confirmación, con el aviso del efecto real ──
    await row(page, soloAudio.title).getByRole('button', { name: 'Ocultar', exact: true }).click();
    await expect(page.getByText('Quedó oculto: el público no lo ve.')).toBeVisible();
    await expect(row(page, soloAudio.title).getByText('Oculto', { exact: true })).toBeVisible();

    const ordenOculto = await publicTitles(browser);
    expect(ordenOculto, 'desaparece de la sección').not.toContain(soloAudio.title);
    expect(ordenOculto).toEqual(ordenInicial.filter((title) => title !== soloAudio.title));
    expect(await portadaContiene(browser, soloAudio.title), 'desaparece del bloque').toBe(false);

    // ── Mostrar: reaparece **en su lugar** del orden, no al final ────────
    await row(page, soloAudio.title).getByRole('button', { name: 'Mostrar', exact: true }).click();
    await expect(page.getByText('Ahora está visible en “Medios”.')).toBeVisible();
    const ordenMostrado = await publicTitles(browser);
    expect(ordenMostrado, 'vuelve a su lugar por fecha').toEqual(ordenInicial);
    expect(await portadaContiene(browser, soloAudio.title)).toBe(true);
  });

  test('editar la descripción (es/en): se publica, y el inglés sin español no guarda (FR-020/FR-023)', async ({
    browser,
  }) => {
    await page.goto('/panel/medios');
    const item = fixtureByKey(seed, 'solo-video');
    const textoEs = `¡Qué día! 🙌 Prédica ajustada por el equipo (${seed.suffix})`;
    const textoEn = `What a day! 🙌 Sermon edited by the team (${seed.suffix})`;

    const patches: string[] = [];
    page.on('request', (request) => {
      if (request.method() === 'PATCH' && request.url().includes('/api/v1/admin/medios/')) {
        patches.push(request.url());
      }
    });

    // ── Inglés sin su español: no permite guardar y no llama a la API ────
    await row(page, item.title).getByRole('button', { name: 'Editar descripción' }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByRole('heading', { name: 'Editar descripción' })).toBeVisible();
    await expect(dialog).toContainText('Estás ajustando la descripción de:');
    await expect(dialog.getByText(item.title)).toBeVisible();
    // El valor vigente va precargado.
    await expect(dialog.getByLabel('Español')).toHaveValue(item.descriptionEs ?? '');

    await dialog.getByRole('button', { name: 'English (opcional)' }).click();
    await expect(dialog.getByText(/Si lo dejas vacío, el público verá/)).toBeVisible();
    await dialog.getByLabel('English (opcional)').fill(textoEn);
    // El español se vacía en su propia pestaña (cada pestaña pinta su campo).
    await dialog.getByRole('button', { name: 'Español' }).click();
    await dialog.getByLabel('Español').fill('');
    await dialog.getByRole('button', { name: 'Guardar' }).click();
    await expect(dialog.getByText(/Falta la versión en español/).first()).toBeVisible();
    await expect(dialog).toBeVisible();
    expect(patches, 'nada se guarda a medias (FR-023)').toEqual([]);

    // ── Con las dos versiones: se guarda y se publica al momento ─────────
    // El intento fallido dejó el foco en el campo en inglés (es el que falla),
    // así que se vuelve a la pestaña del español para rellenarlo.
    await dialog.getByRole('button', { name: 'Español' }).click();
    await dialog.getByLabel('Español').fill(textoEs);
    await dialog.getByRole('button', { name: 'English (opcional)' }).click();
    await dialog.getByLabel('English (opcional)').fill(textoEn);
    await dialog.getByRole('button', { name: 'Guardar' }).click();
    await expect(page.getByText('Cambios guardados: ya visibles en la sección.')).toBeVisible();
    expect(patches, 'el guardado llegó a la API').toHaveLength(1);

    // ── El visitante ve el texto del equipo (es) y su inglés (en) ────────
    const htmlEs = await publicHtml(browser);
    expect(htmlEs).toContain(textoEs);

    const visitor = await visitorPage(browser);
    try {
      await visitor.page.goto('/medios');
      await expect(
        visitor.page.getByRole('list', { name: 'Lista de episodios de medios' }),
      ).toBeVisible();
      await visitor.page.getByRole('banner').getByRole('radio', { name: 'Inglés' }).check();
      await expect(
        visitor.page.getByRole('list', { name: 'List of media episodes' }).getByText(textoEn),
      ).toBeVisible();
    } finally {
      await visitor.close();
    }
  });

  test('unir el par marcado con vista previa y doble confirmación (US3 esc. 1)', async ({
    browser,
  }) => {
    await page.goto('/panel/medios');
    const base = fixtureByKey(seed, 'par-union-base');
    const otra = fixtureByKey(seed, 'par-union-otra');
    await expect(page.getByRole('table', { name: TABLE })).toBeVisible();
    await expect(row(page, base.title)).toBeVisible();
    const filasAntes = await rows(page).count();

    await row(page, base.title).getByRole('button', { name: 'Revisar duplicado' }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByRole('heading', { name: 'Revisar duplicado' })).toBeVisible();
    // Las dos candidatas lado a lado, con su visibilidad actual (FR-014): la
    // primera importada está visible y la adicional, oculta.
    await expect(dialog.getByText(base.title, { exact: true }).first()).toBeVisible();
    await expect(dialog.getByText(otra.title, { exact: true }).first()).toBeVisible();
    await expect(dialog.getByText('Visible', { exact: true })).toBeVisible();
    await expect(dialog.getByText('Oculto', { exact: true })).toBeVisible();
    // La vista previa del resultado, con la nota del nombre que manda y **los
    // dos enlaces juntos**.
    const previa = dialog.getByRole('region', { name: 'Así quedaría la ficha si las unes.' });
    await expect(
      previa.getByRole('heading', { name: 'Así quedaría la ficha si las unes.' }),
    ).toBeVisible();
    await expect(previa.getByText('Al unir, el nombre publicado es el de YouTube.')).toBeVisible();
    for (const source of [...base.sources, ...otra.sources]) {
      await expect(previa.locator(`a[href="${source.url}"]`)).toBeVisible();
    }

    // ── «Unir» exige la confirmación explícita (no se puede deshacer) ────
    await dialog.getByRole('button', { name: 'Unir', exact: true }).click();
    const confirm = page.getByRole('dialog').filter({ hasText: 'Esto no se puede deshacer.' });
    await expect(confirm.getByRole('heading', { name: 'Unir las dos entradas' })).toBeVisible();
    await confirm.getByRole('button', { name: 'Unir definitivamente' }).click();
    await expect(page.getByText('Unidas en una sola ficha, visible en “Medios”.')).toBeVisible();

    // ── Una sola ficha con los dos enlaces, sin marca; la absorbida se va ──
    const unida = row(page, base.title);
    await expect(unida.getByText('Video', { exact: true })).toBeVisible();
    await expect(unida.getByText('Audio', { exact: true })).toBeVisible();
    await expect(unida.getByText('De YouTube y del podcast', { exact: true })).toBeVisible();
    await expect(unida.getByText('Posible duplicado', { exact: true })).toHaveCount(0);
    // La absorbida desaparece del listado: queda con `mergedIntoId` y el listado
    // (que es el que pinta la tabla) ya no la trae.
    await expect(rows(page)).toHaveCount(filasAntes - 1);
    const listado = (await (
      await apiSend(page, 'GET', '/api/v1/admin/medios?limit=100')
    ).json()) as { items: { id: string }[] };
    expect(
      listado.items.map((item) => item.id),
      'la ficha absorbida ya no está en el listado',
    ).not.toContain(otra.id);
    expect(listado.items.map((item) => item.id)).toContain(base.id);

    // El público ve **una** ficha con los dos enlaces.
    const html = await publicHtml(browser);
    expect(html).toContain(base.title);
    expect(html).not.toContain(otra.title);
  });

  test('resolver «no son el mismo sermón»: ambas visibles y sin marca (US3 esc. 2)', async ({
    browser,
  }) => {
    await page.goto('/panel/medios');
    const base = fixtureByKey(seed, 'par-distintos-base');
    const otra = fixtureByKey(seed, 'par-distintos-otra');

    await row(page, base.title).getByRole('button', { name: 'Revisar duplicado' }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByRole('button', { name: 'No son el mismo sermón' }).click();
    await expect(
      page.getByText('Listo: quedaron las dos como episodios distintos, visibles sin marca.'),
    ).toBeVisible();

    // Ninguna de las dos conserva la marca **pendiente**: las dos están visibles
    // y la fila conserva el registro («Resuelta»), sin explicador ni botón de
    // revisión (el par ya no está abierto).
    for (const item of [base, otra]) {
      const fila = row(page, item.title);
      await expect(fila.getByText('Visible', { exact: true })).toBeVisible();
      await expect(fila.getByText('Resuelta', { exact: true })).toBeVisible();
      await expect(fila.getByRole('button', { name: 'Revisar duplicado' })).toHaveCount(0);
      await sinMarcasPendientes(fila);
    }

    // Y el público las ve a las dos (ninguna se repite).
    const orden = await publicTitles(browser);
    expect(orden).toContain(base.title);
    expect(orden).toContain(otra.title);
  });

  test('los dos ejes se resuelven por separado y la marca resuelta queda «Resuelta» (P9-19/FR-017)', async () => {
    await page.goto('/panel/medios');
    const mixta = fixtureByKey(seed, 'par-mixto-base');

    // Resolver el eje de **disponibilidad** («Mantener visible») deja intacto el
    // eje de duplicados: la fila sigue con «Posible duplicado».
    const fila = row(page, mixta.title);
    await expect(fila.getByText('Ya no aparece en el feed', { exact: true })).toBeVisible();
    await fila.getByRole('button', { name: 'Mantener visible' }).click();
    await expect(
      page.getByText('La marca se cerró y no se vuelve a abrir por esta misma causa.'),
    ).toBeVisible();
    await expect(fila.getByText('Ya no aparece en el feed', { exact: true })).toHaveCount(0);
    await expect(fila.getByText('Posible duplicado', { exact: true })).toBeVisible();

    // Y el eje de duplicados se resuelve después, por su cuenta: la fila queda
    // con los dos ejes cerrados y su registro («Resuelta» + las dos marcas).
    await fila.getByRole('button', { name: 'Revisar duplicado' }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'No son el mismo sermón' }).click();
    await expect(
      page.getByText('Listo: quedaron las dos como episodios distintos, visibles sin marca.'),
    ).toBeVisible();
    const cerrada = row(page, mixta.title);
    await expect(cerrada.getByText('Resuelta', { exact: true })).toBeVisible();
    await expect(cerrada.getByText('Posible duplicado · Ya no aparece en el feed')).toBeVisible();
    await sinMarcasPendientes(cerrada);
  });

  test('la precedencia A3: «Mostrar de nuevo» con su confirmación (FR-017/SC-007)', async ({
    browser,
  }) => {
    await page.goto('/panel/medios');
    const item = fixtureByKey(seed, 'marca-unavailable');
    expect(await publicTitles(browser)).not.toContain(item.title);

    await row(page, item.title).getByRole('button', { name: 'Mostrar de nuevo' }).click();
    const confirm = page.getByRole('dialog');
    await expect(confirm.getByRole('heading', { name: 'Mostrar de nuevo' })).toBeVisible();
    await expect(confirm).toContainText(
      `Si el enlace de “${item.title}” sigue sin funcionar, volverá a quedar oculto la próxima vez que se revise.`,
    );
    await expect(confirm.getByRole('button', { name: 'Dejarlo como está' })).toBeVisible();

    await confirm.getByRole('button', { name: 'Mostrar', exact: true }).click();
    await expect(
      page.getByText('La marca se cerró y no se vuelve a abrir por esta misma causa.'),
    ).toBeVisible();

    // La ficha queda visible, sin la marca abierta (con su registro «Resuelta»)
    // y el público la ve.
    const fila = row(page, item.title);
    await expect(fila.getByText('Visible', { exact: true })).toBeVisible();
    await expect(fila.getByText('Resuelta', { exact: true })).toBeVisible();
    await expect(fila.getByText('No disponible', { exact: true })).toBeVisible();
    await sinMarcasPendientes(fila);
    expect(await publicTitles(browser)).toContain(item.title);
  });

  test('permisos cruzados: sin el permiso `medios` no se entra ni se opera (SC-009)', async ({
    browser,
  }) => {
    users = await seedMediosPanelUsers(page, uniqueSuffix());
    const existente = (await (
      await apiSend(page, 'GET', '/api/v1/admin/medios?limit=1')
    ).json()) as { items: { id: string }[] };
    const someId = existente.items[0]?.id ?? '00000000-0000-0000-0000-000000000000';

    // ── Cuenta con **otros** permisos de módulo (no `medios`) ────────────
    const otherContext = await browser.newContext();
    const otherPage = await otherContext.newPage();
    try {
      await loginNewAccount(otherPage, users.otherModules.email, 'Inicial.2026', 'Definitiva.2026');
      await expect(
        panelNav(otherPage).getByRole('link', { name: 'Gestión de medios' }),
      ).toHaveCount(0);
      await expect(
        quickAccess(otherPage).getByRole('link', { name: 'Gestión de medios' }),
      ).toHaveCount(0);
      await expect(
        panelNav(otherPage).getByRole('link', { name: 'Gestión de eventos' }),
      ).toBeVisible();

      await otherPage.goto('/panel/medios');
      await expect(otherPage.getByText(NO_PERMISSION)).toBeVisible();

      // El servidor es la única autoridad (SC-009): 0 operaciones sin permiso.
      for (const [method, path, body] of [
        ['GET', '/api/v1/admin/medios?limit=1', undefined],
        ['POST', '/api/v1/admin/medios/buscar-nuevos', undefined],
        ['PATCH', `/api/v1/admin/medios/${someId}`, { visibility: 'hidden' }],
        ['POST', `/api/v1/admin/medios/${someId}/unir`, { otherId: someId }],
        [
          'POST',
          `/api/v1/admin/medios/${someId}/marcas/resolver`,
          { mark: 'unavailable', visibility: 'visible' },
        ],
      ] as const) {
        const response = await apiSend(otherPage, method, path, body);
        expect(response.status(), `${method} ${path} sin el permiso`).toBe(403);
      }

      // Y la denegación queda registrada (result='denied').
      const denied = (await (
        await apiSend(page, 'GET', '/api/v1/admin/auditoria/acciones?limit=100')
      ).json()) as { items: { action: string; result: string }[] };
      expect(
        denied.items.some((item) => item.result === 'denied' && item.action.startsWith('media.')),
        'la denegación queda registrada en la auditoría',
      ).toBe(true);
    } finally {
      await otherContext.close();
    }

    // ── Cuenta **sin** permisos de módulo ────────────────────────────────
    const noneContext = await browser.newContext();
    const nonePage = await noneContext.newPage();
    try {
      await loginNewAccount(nonePage, users.none.email, 'Inicial.2026', 'Definitiva.2026');
      await expect(
        quickAccess(nonePage).getByRole('link', { name: 'Gestión de medios' }),
      ).toHaveCount(0);
      await nonePage.goto('/panel/medios');
      await expect(nonePage.getByText(NO_PERMISSION)).toBeVisible();
      expect(
        (await apiSend(nonePage, 'GET', '/api/v1/admin/medios?limit=1')).status(),
        'GET medios sin ningún permiso',
      ).toBe(403);
    } finally {
      await noneContext.close();
    }

    // ── Cuenta con **solo** `medios`: su módulo sí, los ajenos no ────────
    const ownContext = await browser.newContext();
    const ownPage = await ownContext.newPage();
    try {
      await loginNewAccount(ownPage, users.medios.email, 'Inicial.2026', 'Definitiva.2026');
      await expect(
        panelNav(ownPage).getByRole('link', { name: 'Gestión de medios' }),
      ).toBeVisible();
      await expect(panelNav(ownPage).getByRole('link', { name: 'Gestión de eventos' })).toHaveCount(
        0,
      );
      await ownPage.goto('/panel/medios');
      await expect(ownPage.getByRole('heading', { level: 1, name: TITLE })).toBeVisible();
      for (const path of ['/api/v1/admin/eventos?limit=1', '/api/v1/admin/ministerios?limit=1']) {
        expect((await apiSend(ownPage, 'GET', path)).status(), `GET ${path} ajeno`).toBe(403);
      }
      // Su propio módulo responde 200 (y sigue sin pulsarse la búsqueda).
      expect((await apiSend(ownPage, 'GET', '/api/v1/admin/medios?limit=1')).status()).toBe(200);
    } finally {
      await ownContext.close();
    }
  });

  test('la auditoría registra las operaciones con etiquetas legibles (SC-010)', async () => {
    // RG9-7: la búsqueda no se pulsó en ningún momento de la spec.
    expect(busquedas, 'la spec nunca pulsa «Buscar episodios nuevos»').toEqual([]);

    await page.goto('/panel/medios');
    await panelNav(page).getByRole('link', { name: 'Auditoría' }).click();
    await expect(page.getByRole('heading', { name: 'Auditoría' })).toBeVisible();
    await page.getByRole('button', { name: 'Acciones administrativas' }).click();

    const actionsTable = page.getByRole('table', {
      name: 'Historial de acciones administrativas',
    });
    await expect(actionsTable).toBeVisible();
    const summary = page.getByRole('status').filter({ hasText: 'registros' });

    // Se recorren las páginas necesarias: la tanda deja operaciones más recientes
    // que las de este recorrido.
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

    // Las cinco operaciones que la spec sí ejecuta, con su etiqueta humana.
    // (`media.sync` —«Buscó episodios nuevos»— no se puede producir aquí: la
    // búsqueda no se pulsa por RG9-7; su etiqueta la fija la prueba unitaria de
    // `ACTION_LABELS` y su registro, las pruebas de Go de T920/T921.)
    for (const label of [
      'Modificó la descripción de un episodio',
      'Ocultó un episodio',
      'Mostró un episodio',
      'Unió dos entradas en una sola ficha',
      'Resolvió la marca de un episodio',
    ]) {
      expect(actions, `la auditoría muestra «${label}»`).toContain(label);
    }
    // Ningún código crudo a la vista.
    expect(actions).not.toContain('media.');
    // El objetivo se describe en palabras («Medios · <nombre>», targetKind `content`).
    expect(actions).toContain('Medios ·');
    expect(actions).toContain(`Medios · ${fixtureByKey(seed, 'solo-audio').title}`);
  });
});
