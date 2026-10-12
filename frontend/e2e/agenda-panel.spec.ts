import { expect, test, type Browser, type BrowserContext, type Page } from '@playwright/test';
import { apiSend, loginViaUi, uniqueSuffix } from './helpers';
import { createAdminContext } from './publicSeed';
import { resetAgenda, shiftDays, costaRicaToday } from './agendaSeed';

/**
 * E2e del panel de agenda (US2/US3/US4, quickstart §2–§7/§11/§12): el equipo
 * crea y publica un evento con fechas y una actividad desde la interfaz y el
 * público lo ve **en la primera carga** (SC-003/SC-004), el estado real
 * (borrador/publicado/pasado) se ve en el listado, el catálogo de categorías se
 * gestiona y una categoría **en uso no se puede borrar** (`409`, ⟲ D-16), los
 * tres permisos de F4 son independientes en servidor (SC-005/SC-011) y todo
 * queda registrado en la auditoría de F2 **con etiquetas legibles** (SC-013).
 *
 * Cómo se ejecuta (ver `agenda-publica.spec.ts`): los tres specs de la agenda
 * comparten estado, así que van en la misma tanda con `--workers=1`; aquí los
 * tests van en `serial` porque el segundo reutiliza el catálogo del primero.
 *
 *   cd frontend && npx playwright test --config e2e/playwright.config.ts --workers=1
 */

test.describe.configure({ mode: 'serial' });

const panelNav = (page: Page) => page.getByRole('navigation', { name: 'Navegación del panel' });
const quickAccess = (page: Page) => page.getByRole('navigation', { name: 'Accesos rápidos' });

/** Espera a que el listado del panel tenga la fila del elemento indicado. */
function row(page: Page, text: string) {
  return page.getByRole('table').getByRole('row').filter({ hasText: text });
}

/** Abre un contexto de visitante (sin sesión) y devuelve su página. */
async function visitorPage(browser: Browser): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext();
  return { context, page: await context.newPage() };
}

/** Rellena los campos obligatorios y comunes de un elemento de agenda. */
async function fillCommonFields(
  dialog: ReturnType<Page['getByRole']>,
  values: { title: string; place: string; detail: string },
) {
  await dialog.getByLabel(/^Título \(Español\)/).fill(values.title);
  await dialog.getByLabel(/^Lugar \(Español\)/).fill(values.place);
  await dialog.getByLabel(/^Detalle \(Español\)/).fill(values.detail);
}

/**
 * Busca un texto en el historial de acciones administrativas recorriendo su
 * paginación (la base compartida acumula registros de corridas anteriores y las
 * acciones de F4 pueden quedar fuera de la primera página).
 */
async function auditHas(page: Page, text: string): Promise<boolean> {
  const table = page.getByRole('table');
  const pagination = page.getByRole('navigation', { name: 'Paginación' });
  const previous = pagination.getByRole('button', { name: 'Anterior' });

  // Vuelve a la primera página antes de buscar (la búsqueda anterior pudo avanzar).
  while (await previous.isEnabled()) {
    await previous.click();
    await page.waitForTimeout(150);
  }

  for (let attempt = 0; attempt < 12; attempt += 1) {
    if ((await table.getByText(text).count()) > 0) {
      return true;
    }
    const next = pagination.getByRole('button', { name: 'Siguiente' });
    if (!(await next.isEnabled())) {
      return false;
    }
    await next.click();
    await page.waitForTimeout(150);
  }
  return false;
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

test.describe('Panel de agenda — gestión, publicación y permisos (US2, US3, US4)', () => {
  test.setTimeout(300_000);

  let adminContext: BrowserContext;
  let adminPage: Page;
  let suffix: string;
  let eventTitle: string;
  let eventTitleEdited: string;
  let eventPlace: string;
  let eventDetail: string;
  let eventFrequency: string;
  let eventDate: string;
  let activityTitle: string;
  let activityPlace: string;
  let activityDetail: string;
  let activityDate: string;
  let categoryName: string;
  let categoryNameEn: string;
  let categoryNameEdited: string;

  test.beforeAll(async ({ browser }) => {
    suffix = uniqueSuffix();
    eventTitle = `Vigilia del panel ${suffix}`;
    eventTitleEdited = `Vigilia del panel (editada) ${suffix}`;
    eventPlace = `Templo central ${suffix}`;
    eventDetail = `Noche de oración ${suffix}.`;
    eventFrequency = `Todos los viernes ${suffix}`;
    eventDate = shiftDays(costaRicaToday(), 7);
    activityTitle = `Repartir comida ${suffix}`;
    activityPlace = `Comedor comunal ${suffix}`;
    activityDetail = `Acción social ${suffix}.`;
    activityDate = shiftDays(costaRicaToday(), 8);
    categoryName = `Ayuda social ${suffix}`;
    categoryNameEn = `Social aid ${suffix}`;
    categoryNameEdited = `Ayuda social renombrada ${suffix}`;

    adminContext = await createAdminContext(browser);
    adminPage = await adminContext.newPage();
    await resetAgenda(adminPage);
  });

  test.afterAll(async () => {
    await adminContext.close();
  });

  test('crear y publicar un evento y una actividad; el público lo ve en la primera carga (SC-003/SC-004)', async ({
    browser,
  }) => {
    const page = adminPage;
    const started = Date.now();

    // ── Inicio del panel: los tres accesos de agenda, filtrados por permiso ─
    await page.goto('/panel');
    const cards = quickAccess(page);
    await expect(cards.getByRole('link', { name: 'Gestión de eventos' })).toHaveAttribute(
      'href',
      '/panel/eventos',
    );
    await expect(cards.getByRole('link', { name: 'Gestión de actividades' })).toHaveAttribute(
      'href',
      '/panel/actividades',
    );
    await expect(cards.getByRole('link', { name: 'Categorías' })).toHaveAttribute(
      'href',
      '/panel/categorias',
    );
    await expect(panelNav(page).getByRole('link', { name: 'Gestión de eventos' })).toBeVisible();

    // ── US2: alta de un evento con fechas desde la interfaz (FR-010/FR-015) ─
    await panelNav(page).getByRole('link', { name: 'Gestión de eventos' }).click();
    await expect(page.getByRole('heading', { name: 'Gestión de eventos' })).toBeVisible();
    await page.getByRole('button', { name: 'Agregar evento' }).click();

    const eventDialog = page.getByRole('dialog');
    await fillCommonFields(eventDialog, {
      title: eventTitle,
      place: eventPlace,
      detail: eventDetail,
    });
    await eventDialog.getByLabel(/^Frecuencia \(opcional, Español\)/).fill(eventFrequency);
    await eventDialog.getByLabel(/^Fecha 1/).fill(eventDate);
    await eventDialog.getByLabel(/^Hora de inicio/).fill('19:00');
    await eventDialog.getByLabel(/^Hora de fin/).fill('21:00');
    await eventDialog.getByLabel(/^Categoría/).selectOption({ label: 'Noche de oración' });
    await eventDialog.getByRole('button', { name: 'Agregar evento' }).click();

    // Nace en borrador: el público no lo ve todavía (FR-013).
    await expect(page.getByText('Evento creado.')).toBeVisible();
    await expect(
      page.getByText('Quedó en borrador: no lo verá el público hasta que lo publiques.'),
    ).toBeVisible();
    await expect(row(page, eventTitle).getByText('Borrador')).toBeVisible();

    const draftVisitor = await visitorPage(browser);
    try {
      await draftVisitor.page.goto('/eventos');
      await expect(draftVisitor.page.getByText(eventTitle)).toHaveCount(0);
    } finally {
      await draftVisitor.context.close();
    }

    // ── US4: publicar por elemento y verlo en la primera carga (SC-003) ─────
    await row(page, eventTitle).getByRole('button', { name: 'Publicar' }).click();
    await expect(page.getByText('Publicado en "Eventos y actividades".')).toBeVisible();
    await expect(row(page, eventTitle).getByText('Publicado')).toBeVisible();

    const publishedVisitor = await visitorPage(browser);
    try {
      await publishedVisitor.page.goto('/eventos');
      await expect(
        publishedVisitor.page.getByRole('heading', { level: 3, name: eventTitle }),
      ).toBeVisible();
      await expect(publishedVisitor.page.getByText(eventFrequency)).toBeVisible();
      await expect(publishedVisitor.page.getByText(eventPlace)).toBeVisible();
    } finally {
      await publishedVisitor.context.close();
    }

    // ── FR-014: editar un evento publicado → «ya visibles en la sección» ────
    await row(page, eventTitle).getByRole('button', { name: 'Editar' }).click();
    const editDialog = page.getByRole('dialog');
    await editDialog.getByLabel(/^Título \(Español\)/).fill(eventTitleEdited);
    await editDialog.getByRole('button', { name: 'Guardar cambios' }).click();
    await expect(page.getByText('Cambios guardados: ya visibles en la sección.')).toBeVisible();

    const editedVisitor = await visitorPage(browser);
    try {
      await editedVisitor.page.goto('/eventos');
      await expect(
        editedVisitor.page.getByRole('heading', { level: 3, name: eventTitleEdited }),
      ).toBeVisible();
      await expect(editedVisitor.page.getByText(eventTitle)).toHaveCount(0);
    } finally {
      await editedVisitor.context.close();
    }

    // ── US4: retirar de la sección (con confirmación) ──────────────────────
    await row(page, eventTitleEdited)
      .getByRole('button', { name: 'Retirar de la sección' })
      .click();
    const retireDialog = page.getByRole('dialog');
    await expect(retireDialog.getByText(/dejará de ser visible al público/)).toBeVisible();
    await retireDialog.getByRole('button', { name: 'Retirar' }).click();
    await expect(page.getByText(/Retirado de la sección/)).toBeVisible();

    const retiredVisitor = await visitorPage(browser);
    try {
      await retiredVisitor.page.goto('/eventos');
      await expect(retiredVisitor.page.getByText(eventTitleEdited)).toHaveCount(0);
    } finally {
      await retiredVisitor.context.close();
    }

    // ── SC-004: publicar de nuevo y crear+publicar una actividad ───────────
    await row(page, eventTitleEdited).getByRole('button', { name: 'Publicar' }).click();
    await expect(page.getByText('Publicado en "Eventos y actividades".')).toBeVisible();

    await panelNav(page).getByRole('link', { name: 'Gestión de actividades' }).click();
    await expect(page.getByRole('heading', { name: 'Gestión de actividades' })).toBeVisible();
    await page.getByRole('button', { name: 'Agregar actividad' }).click();

    const activityDialog = page.getByRole('dialog');
    await fillCommonFields(activityDialog, {
      title: activityTitle,
      place: activityPlace,
      detail: activityDetail,
    });
    await activityDialog.getByLabel(/^Fecha\s/).fill(activityDate);
    await activityDialog.getByLabel(/^Hora de inicio/).fill('08:00');
    await activityDialog.getByLabel(/^Hora de fin/).fill('11:00');
    await activityDialog.getByRole('button', { name: 'Agregar actividad' }).click();
    await expect(page.getByText('Actividad creada.')).toBeVisible();

    await row(page, activityTitle).getByRole('button', { name: 'Publicar' }).click();
    await expect(page.getByText('Publicado en "Eventos y actividades".')).toBeVisible();

    const elapsed = (Date.now() - started) / 1000;
    expect(
      elapsed,
      'SC-004: crear y publicar un evento y una actividad en menos de 2 minutos',
    ).toBeLessThan(120);

    const bothVisitor = await visitorPage(browser);
    try {
      await bothVisitor.page.goto('/eventos');
      await expect(
        bothVisitor.page.getByRole('heading', { level: 3, name: eventTitleEdited }),
      ).toBeVisible();
      await expect(
        bothVisitor.page.getByRole('heading', { level: 3, name: activityTitle }),
      ).toBeVisible();
    } finally {
      await bothVisitor.context.close();
    }

    // ── US2: eliminar con confirmación explícita (sin papelera, FR-015) ────
    await panelNav(page).getByRole('link', { name: 'Gestión de eventos' }).click();
    await row(page, eventTitleEdited).getByRole('button', { name: 'Eliminar' }).click();
    const deleteDialog = page.getByRole('dialog');
    await expect(
      deleteDialog.getByText(
        `"${eventTitleEdited}" se eliminará para siempre. Esto no se puede deshacer.`,
      ),
    ).toBeVisible();
    await deleteDialog.getByRole('button', { name: 'Eliminar' }).click();
    await expect(page.getByText('Eliminado.')).toBeVisible();
    await expect(row(page, eventTitleEdited)).toHaveCount(0);
  });

  test('catálogo de categorías: alta, renombrado y `409` en uso que solo deja cerrar (Q6/Q7, ⟲ D-16)', async () => {
    const page = adminPage;

    await panelNav(page).getByRole('link', { name: 'Categorías' }).click();
    await expect(page.getByRole('heading', { name: 'Categorías de agenda' })).toBeVisible();

    // El catálogo inicial sembrado por la migración `000007` está completo.
    for (const seeded of [
      'Acción social',
      'Predicación',
      'Adoración',
      'Entrevista',
      'Noche de oración',
      'Otro',
    ]) {
      await expect(row(page, seeded).first()).toBeVisible();
    }

    // ── Alta (nombre es obligatorio, en opcional) ──────────────────────────
    await page.getByRole('button', { name: 'Agregar categoría' }).click();
    const createDialog = page.getByRole('dialog');
    await createDialog.getByLabel(/^Nombre \(Español\)/).fill(categoryName);
    await createDialog.getByLabel(/^Nombre \(English/).fill(categoryNameEn);
    await createDialog.getByRole('button', { name: 'Agregar categoría' }).click();
    await expect(page.getByText('Categoría creada.')).toBeVisible();
    await expect(row(page, categoryName).getByText(categoryNameEn)).toBeVisible();

    // ── Renombrado: el cambio se ve de inmediato en los elementos ──────────
    await row(page, categoryName).getByRole('button', { name: 'Editar' }).click();
    const editDialog = page.getByRole('dialog');
    await editDialog.getByLabel(/^Nombre \(Español\)/).fill(categoryNameEdited);
    await editDialog.getByRole('button', { name: 'Guardar cambios' }).click();
    await expect(page.getByText('Categoría guardada.')).toBeVisible();
    await expect(row(page, categoryNameEdited)).toBeVisible();

    // ── `409` firme: en uso NO se borra y no se ofrece desligarla ──────────
    const categoryId = await (async () => {
      const body = (await (await apiSend(page, 'GET', '/api/v1/admin/categorias')).json()) as {
        items: { id: string; nameEs: string }[];
      };
      const found = body.items.find((item) => item.nameEs === categoryNameEdited);
      if (!found) throw new Error('No se encontró la categoría recién creada');
      return found.id;
    })();

    const linkedEvent = `Con categoría ${suffix}`;
    const created = await apiSend(page, 'POST', '/api/v1/admin/eventos', {
      titleEs: linkedEvent,
      detailEs: 'Elemento que usa la categoría.',
      placeEs: `Salón ${suffix}`,
      categoryId,
      dates: [{ date: shiftDays(costaRicaToday(), 9) }],
    });
    expect(created.ok(), `POST evento con categoría → ${created.status()}`).toBe(true);

    await row(page, categoryNameEdited).getByRole('button', { name: 'Eliminar' }).click();
    const inUseDialog = page.getByRole('dialog');
    await inUseDialog.getByRole('button', { name: 'Eliminar' }).click();
    await expect(inUseDialog.getByText('No se puede eliminar')).toBeVisible();
    await expect(
      inUseDialog.getByText('Está en uso por 1 elemento: quítala primero de los elementos.'),
    ).toBeVisible();
    // Su única acción es cerrar: nunca ofrece desligarla ni continuar.
    expect(await inUseDialog.getByRole('button', { name: 'Eliminar' }).count()).toBe(0);
    await inUseDialog.getByRole('button', { name: 'Cancelar' }).click();
    await expect(row(page, categoryNameEdited)).toBeVisible();

    // ── Quitada de los elementos, el borrado procede ───────────────────────
    const events = (await (
      await apiSend(page, 'GET', '/api/v1/admin/eventos?limit=100')
    ).json()) as {
      items: { id: string; titleEs: string }[];
    };
    const linked = events.items.find((item) => item.titleEs === linkedEvent);
    if (!linked) throw new Error('No se encontró el evento que usa la categoría');
    expect((await apiSend(page, 'DELETE', `/api/v1/admin/eventos/${linked.id}`)).status()).toBe(
      204,
    );

    await row(page, categoryNameEdited).getByRole('button', { name: 'Eliminar' }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Eliminar' }).click();
    await expect(page.getByText('Eliminado.')).toBeVisible();
    await expect(row(page, categoryNameEdited)).toHaveCount(0);
  });

  test('permisos cruzados de los tres módulos y auditoría con etiquetas legibles (SC-005/SC-011/SC-013)', async ({
    browser,
  }) => {
    const page = adminPage;

    // ── Cuentas con un solo permiso de agenda (creadas por API) ─────────────
    const accounts = [
      {
        code: 'eventos',
        label: 'Solo eventos',
        allowed: '/panel/eventos',
        denied: '/panel/actividades',
      },
      {
        code: 'actividades',
        label: 'Solo actividades',
        allowed: '/panel/actividades',
        denied: '/panel/eventos',
      },
      {
        code: 'categorias',
        label: 'Solo categorías',
        allowed: '/panel/categorias',
        denied: '/panel/eventos',
      },
    ] as const;

    for (const [index, account] of accounts.entries()) {
      const roleResponse = await apiSend(page, 'POST', '/api/v1/admin/roles', {
        name: `${account.label} ${suffix}`,
        permissions: [account.code],
      });
      expect(roleResponse.ok(), `POST rol ${account.code} → ${roleResponse.status()}`).toBe(true);
      const roleId = ((await roleResponse.json()) as { id: string }).id;

      const email = `agenda-${account.code}-${suffix}@ejemplo.com`;
      const userResponse = await apiSend(page, 'POST', '/api/v1/admin/usuarios', {
        firstName: 'Cuenta',
        lastName: `${account.label} ${suffix}`,
        email,
        phone: `6120${index}0${index}0`,
        roleId,
        password: 'Inicial.2026',
      });
      expect(userResponse.ok(), `POST usuario ${account.code} → ${userResponse.status()}`).toBe(
        true,
      );

      const context = await browser.newContext();
      const accountPage = await context.newPage();
      try {
        await loginNewAccount(accountPage, email, 'Inicial.2026', 'Definitiva.2026');

        // Solo su módulo en la navegación del panel (FR-012).
        await expect(
          panelNav(accountPage).getByRole('link', { name: 'Gestión de eventos' }),
        ).toHaveCount(account.code === 'eventos' ? 1 : 0);
        await expect(
          panelNav(accountPage).getByRole('link', { name: 'Gestión de actividades' }),
        ).toHaveCount(account.code === 'actividades' ? 1 : 0);
        await expect(panelNav(accountPage).getByRole('link', { name: 'Categorías' })).toHaveCount(
          account.code === 'categorias' ? 1 : 0,
        );

        // Su módulo abre; el que no le toca cae en «sin permiso» (FR-012/SC-005).
        await accountPage.goto(account.allowed);
        await expect(
          accountPage.getByRole('heading', { name: 'No tienes acceso a esta sección' }),
        ).toHaveCount(0);
        await accountPage.goto(account.denied);
        await expect(
          accountPage.getByRole('heading', { name: 'No tienes acceso a esta sección' }),
        ).toBeVisible();

        // Y el servidor es la única autoridad (SC-011): la API forzada responde 403.
        for (const path of [
          '/api/v1/admin/eventos',
          '/api/v1/admin/actividades',
          '/api/v1/admin/categorias',
        ]) {
          const own = path.endsWith(`/${account.code}`);
          const response = await apiSend(accountPage, 'GET', `${path}?limit=1`);
          expect(
            response.status(),
            `GET ${path} con la cuenta «${account.code}» (${own ? 'propio' : 'ajeno'})`,
          ).toBe(own ? 200 : 403);
        }
      } finally {
        await context.close();
      }
    }

    // ── SC-013: la auditoría de F2 registra F4 con etiquetas legibles ──────
    await page.goto('/panel/auditoria');
    await page.getByRole('button', { name: 'Acciones administrativas' }).click();
    const actions = page.getByRole('table');
    await expect(actions.getByRole('columnheader', { name: 'Acción' })).toBeVisible();

    // Los códigos crudos no se muestran nunca; sí sus etiquetas humanas y el
    // elemento concreto sobre el que se actuó (ux.md §4.9).
    await expect(actions.getByText('event.create')).toHaveCount(0);
    await expect(actions.getByText('activity.create')).toHaveCount(0);
    for (const text of [
      'Creó un evento',
      'Publicó un evento',
      'Creó una actividad',
      'Creó una categoría',
      eventTitleEdited,
      activityTitle,
    ]) {
      expect(await auditHas(page, text), `la auditoría registra «${text}»`).toBe(true);
    }
  });
});
