import { expect, test, type Page } from '@playwright/test';
import { ADMIN, API_URL, apiSend, ensureInitialized, loginViaUi, uniqueSuffix } from './helpers';

/**
 * E2E del recorrido de auditoría (quickstart.md §10). Cubre SC-012 (encontrar
 * los accesos y las acciones de una cuenta en menos de 1 minuto) y SC-013 (todo
 * queda registrado y **nada** es editable ni borrable).
 *
 * Genera actividad real —accesos exitosos y fallidos (incluido un correo
 * inexistente) y acciones de gestión— y después recorre la sección de registro:
 * los dos historiales con fecha/hora, resultado e IP; filtros por cuenta y
 * rango de fechas con paginación; el último acceso en la ficha (y "Nunca ha
 * entrado al panel." en una cuenta nueva); el intento sin cuenta asociada; y la
 * ausencia de controles de edición/borrado.
 */

interface UserListItem {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
}

interface RoleListItem {
  id: string;
  name: string;
}

function localDateInput(offsetDays = 0): string {
  const date = new Date();
  date.setDate(date.getDate() + offsetDays);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

async function findUser(page: Page, email: string): Promise<UserListItem> {
  const response = await apiSend(page, 'GET', '/api/v1/admin/usuarios?limit=100');
  const body = (await response.json()) as { items: UserListItem[] };
  const user = body.items.find((item) => item.email === email);
  if (!user) {
    throw new Error(`No se encontró la cuenta ${email} recién creada`);
  }
  return user;
}

async function findRole(page: Page, name: string): Promise<RoleListItem> {
  const response = await apiSend(page, 'GET', '/api/v1/admin/roles?limit=100');
  const body = (await response.json()) as { items: RoleListItem[] };
  const role = body.items.find((item) => item.name === name);
  if (!role) {
    throw new Error(`No se encontró el rol ${name} recién creado`);
  }
  return role;
}

test.describe('Auditoría: registro de accesos y acciones (US8)', () => {
  test.setTimeout(180_000);

  test('accesos y acciones quedan registrados y no son editables ni borrables', async ({
    browser,
    request,
  }) => {
    await ensureInitialized(request);

    const suffix = uniqueSuffix();
    const roleName = `Auditoría ${suffix}`;
    const throwawayRole = `Temporal ${suffix}`;
    const auditor = {
      firstName: 'Aurelio',
      lastName: `Auditor ${suffix}`,
      email: `auditor-${suffix}@ejemplo.com`,
      password: 'Inicial.2026',
      newPassword: 'Restablecida.2026',
    };
    const nueva = {
      firstName: 'Nieves',
      lastName: `SinUso ${suffix}`,
      email: `nueva-${suffix}@ejemplo.com`,
      password: 'Inicial.2026',
    };
    const nadieEmail = `nadie-${suffix}@ejemplo.com`;
    // La cuenta se edita por API más abajo; los JOIN de la auditoría muestran
    // siempre el nombre vigente.
    const auditorUpdatedFirstName = `${auditor.firstName} Editado`;
    const auditorDisplayName = `${auditorUpdatedFirstName} ${auditor.lastName}`;

    const context = await browser.newContext();
    const page = await context.newPage();

    try {
      const nav = page.getByRole('navigation', { name: 'Navegación del panel' });

      // ── Acceso exitoso del administrador (queda en el historial) ──────────
      await loginViaUi(page, ADMIN.email, ADMIN.password);
      await expect(page).toHaveURL(/\/panel$/);

      // ── Acciones de gestión desde la interfaz: crear rol y cuenta ─────────
      await nav.getByRole('link', { name: 'Roles' }).click();
      await page.getByRole('button', { name: 'Crear rol' }).click();
      const roleDialog = page.getByRole('dialog');
      await roleDialog.getByLabel('Nombre').fill(roleName);
      await roleDialog.getByRole('checkbox', { name: 'Eventos' }).check();
      await roleDialog.getByRole('button', { name: 'Crear rol' }).click();
      await expect(page.getByText('Rol creado.')).toBeVisible();

      await nav.getByRole('link', { name: 'Usuarios' }).click();
      await page.getByRole('button', { name: 'Crear usuario' }).click();
      const userDialog = page.getByRole('dialog');
      await userDialog.getByLabel('Nombre').fill(auditor.firstName);
      await userDialog.getByLabel('Apellidos').fill(auditor.lastName);
      await userDialog.getByLabel('Correo').fill(auditor.email);
      await userDialog.getByLabel('Teléfono').fill('612 000 111');
      await userDialog.getByLabel('Rol').selectOption({ label: roleName });
      await userDialog.getByLabel('Contraseña inicial').fill(auditor.password);
      await userDialog.getByRole('button', { name: 'Crear cuenta' }).click();
      await expect(page.getByText('Cuenta creada.')).toBeVisible();

      // Cuenta que nunca inicia sesión: su ficha dirá que no tiene accesos.
      await page.getByRole('button', { name: 'Crear usuario' }).click();
      const nuevaDialog = page.getByRole('dialog');
      await nuevaDialog.getByLabel('Nombre').fill(nueva.firstName);
      await nuevaDialog.getByLabel('Apellidos').fill(nueva.lastName);
      await nuevaDialog.getByLabel('Correo').fill(nueva.email);
      await nuevaDialog.getByLabel('Teléfono').fill('612 000 222');
      await nuevaDialog.getByLabel('Rol').selectOption({ label: roleName });
      await nuevaDialog.getByLabel('Contraseña inicial').fill(nueva.password);
      await nuevaDialog.getByRole('button', { name: 'Crear cuenta' }).click();
      await expect(page.getByText('Cuenta creada.')).toBeVisible();

      // ── Acciones de gestión por API (rápido y suficiente para el registro) ─
      const auditorUser = await findUser(page, auditor.email);
      const role = await findRole(page, roleName);

      // Varias ediciones del rol para que el historial de acciones tenga más de
      // una página (limit por defecto = 20) y la paginación sea real.
      for (let i = 0; i < 22; i += 1) {
        const permissions = i % 2 === 0 ? ['eventos', 'actividades'] : ['eventos'];
        const patch = await apiSend(page, 'PATCH', `/api/v1/admin/roles/${role.id}`, {
          permissions,
        });
        expect(patch.ok(), `PATCH rol nº ${i + 1}`).toBe(true);
      }

      // Editar, desactivar, reactivar y restablecer la contraseña de la cuenta.
      expect(
        (
          await apiSend(page, 'PATCH', `/api/v1/admin/usuarios/${auditorUser.id}`, {
            firstName: auditorUpdatedFirstName,
          })
        ).ok(),
      ).toBe(true);
      expect(
        (
          await apiSend(page, 'PATCH', `/api/v1/admin/usuarios/${auditorUser.id}`, {
            isActive: false,
          })
        ).ok(),
      ).toBe(true);
      expect(
        (
          await apiSend(page, 'PATCH', `/api/v1/admin/usuarios/${auditorUser.id}`, {
            isActive: true,
          })
        ).ok(),
      ).toBe(true);
      expect(
        (
          await apiSend(page, 'POST', `/api/v1/admin/usuarios/${auditorUser.id}/password`, {
            password: auditor.newPassword,
          })
        ).ok(),
      ).toBe(true);

      // Crear y eliminar un rol sin uso (acción completada).
      const created = await apiSend(page, 'POST', '/api/v1/admin/roles', {
        name: throwawayRole,
        permissions: ['noticias'],
      });
      expect(created.ok()).toBe(true);
      const throwaway = (await created.json()) as RoleListItem;
      expect((await apiSend(page, 'DELETE', `/api/v1/admin/roles/${throwaway.id}`)).ok()).toBe(
        true,
      );

      // Acción fallida: nombre de rol duplicado (409) → queda como "No completada".
      const duplicate = await apiSend(page, 'POST', '/api/v1/admin/roles', {
        name: roleName,
        permissions: ['eventos'],
      });
      expect(duplicate.status()).toBe(409);

      // ── Accesos: fallido con cuenta asociada y correo inexistente ─────────
      const failedAccount = await request.post(`${API_URL}/api/v1/auth/login`, {
        data: { email: auditor.email, password: 'Contraseña.Mala.2026' },
      });
      expect(failedAccount.status()).toBe(401);
      const failedUnknown = await request.post(`${API_URL}/api/v1/auth/login`, {
        data: { email: nadieEmail, password: 'Contraseña.Mala.2026' },
      });
      expect(failedUnknown.status()).toBe(401);

      // El intento sin cuenta no crea ninguna cuenta "fantasma" (US8 esc. 7).
      const usersResponse = await apiSend(page, 'GET', '/api/v1/admin/usuarios?limit=100');
      const usersBody = (await usersResponse.json()) as { items: UserListItem[] };
      expect(usersBody.items.some((item) => item.email === nadieEmail)).toBe(false);

      // ── Recorrido de la sección de registro ───────────────────────────────
      await nav.getByRole('link', { name: 'Auditoría' }).click();
      await expect(page.getByRole('heading', { name: 'Auditoría' })).toBeVisible();

      const accessTable = page.getByRole('table', { name: 'Historial de accesos al panel' });
      await expect(accessTable).toBeVisible();
      await expect(accessTable.getByRole('columnheader', { name: 'Fecha y hora' })).toBeVisible();
      await expect(accessTable.getByRole('columnheader', { name: 'Resultado' })).toBeVisible();
      await expect(accessTable.getByRole('columnheader', { name: 'IP de origen' })).toBeVisible();

      // Fecha/hora, resultado e IP visibles en cada fila.
      await expect(accessTable.getByText(/\d{2}\/\d{2}\/\d{4}, \d{2}:\d{2}/).first()).toBeVisible();
      await expect(accessTable.getByText(/\d{1,3}(\.\d{1,3}){3}/).first()).toBeVisible();
      await expect(accessTable.getByText('Exitoso').first()).toBeVisible();
      await expect(accessTable.getByText('Fallido').first()).toBeVisible();

      // Intento sin cuenta asociada, sin revelar ningún correo (F-01/FR-026).
      await expect(accessTable.getByText('Intento sin cuenta asociada').first()).toBeVisible();
      await expect(page.getByText(nadieEmail)).toHaveCount(0);

      // Filtro por cuenta: solo los registros de esa cuenta.
      await page.getByLabel('Cuenta').selectOption({ label: auditorDisplayName });
      await page.getByRole('button', { name: 'Filtrar' }).click();
      await expect(accessTable.getByText(auditorDisplayName).first()).toBeVisible();
      await expect(accessTable.getByText('Fallido').first()).toBeVisible();

      // Filtro por una cuenta sin accesos: estado vacío filtrado.
      await page
        .getByLabel('Cuenta')
        .selectOption({ label: `${nueva.firstName} ${nueva.lastName}` });
      await page.getByRole('button', { name: 'Filtrar' }).click();
      await expect(page.getByText(/no hay registros que coincidan con esos filtros/)).toBeVisible();
      await page.getByRole('button', { name: 'Quitar filtros' }).click();

      // Filtro por rango de fechas (con datos) y rango sin resultados.
      await page.getByLabel('Desde').fill(localDateInput(-1));
      await page.getByLabel('Hasta').fill(localDateInput(1));
      await page.getByRole('button', { name: 'Filtrar' }).click();
      await expect(accessTable.getByRole('row').nth(1)).toBeVisible();
      await expect(page.getByRole('navigation', { name: 'Paginación' })).toBeVisible();

      await page.getByLabel('Desde').fill('2099-01-01');
      await page.getByLabel('Hasta').fill('2099-01-02');
      await page.getByRole('button', { name: 'Filtrar' }).click();
      await expect(page.getByText(/no hay registros que coincidan con esos filtros/)).toBeVisible();
      await page.getByRole('button', { name: 'Quitar filtros' }).click();

      // ── Historial de acciones administrativas ─────────────────────────────
      await page.getByRole('button', { name: 'Acciones administrativas' }).click();
      const actionsTable = page.getByRole('table', {
        name: 'Historial de acciones administrativas',
      });
      await expect(actionsTable).toBeVisible();
      await expect(actionsTable.getByRole('columnheader', { name: 'Quién' })).toBeVisible();
      await expect(actionsTable.getByRole('columnheader', { name: 'Acción' })).toBeVisible();
      await expect(actionsTable.getByRole('columnheader', { name: 'Sobre' })).toBeVisible();
      await expect(actionsTable.getByRole('columnheader', { name: 'Resultado' })).toBeVisible();

      // Cada fila trae fecha/hora y resultado; las acciones recientes se ven.
      await expect(
        actionsTable.getByText(/\d{2}\/\d{2}\/\d{4}, \d{2}:\d{2}/).first(),
      ).toBeVisible();
      await expect(actionsTable.getByText('Completada').first()).toBeVisible();
      await expect(actionsTable.getByText('No completada').first()).toBeVisible();
      await expect(
        actionsTable.getByText('Restableció la contraseña', { exact: true }).first(),
      ).toBeVisible();
      await expect(
        actionsTable.getByText('Desactivó una cuenta', { exact: true }).first(),
      ).toBeVisible();
      await expect(
        actionsTable.getByText('Activó una cuenta', { exact: true }).first(),
      ).toBeVisible();
      await expect(actionsTable.getByText('Eliminó un rol', { exact: true }).first()).toBeVisible();

      const summary = page.getByRole('status').filter({ hasText: 'registros' });

      // Filtro por cuenta: lo que hizo y lo que se hizo sobre ella.
      await page.getByLabel('Cuenta').selectOption({ label: auditorDisplayName });
      await page.getByRole('button', { name: 'Filtrar' }).click();
      await expect(actionsTable.getByText(auditor.email).first()).toBeVisible();
      await expect(
        actionsTable.getByText('Restableció la contraseña', { exact: true }).first(),
      ).toBeVisible();
      await page.getByRole('button', { name: 'Quitar filtros' }).click();

      // La inicialización (FR-007) queda como `user.create` sin actor (la UI lo
      // muestra como "Sistema") y es, por definición, el registro MÁS ANTIGUO.
      // Se comprueba por API para no depender de la última página del historial.
      const newest = await apiSend(page, 'GET', '/api/v1/admin/auditoria/acciones?limit=1');
      const newestBody = (await newest.json()) as { total: number };
      const oldest = await apiSend(
        page,
        'GET',
        `/api/v1/admin/auditoria/acciones?limit=1&offset=${Math.max(0, newestBody.total - 1)}`,
      );
      const oldestBody = (await oldest.json()) as {
        items: { action: string; actorName: string | null; targetLabel: string | null }[];
      };
      expect(oldestBody.items[0]?.action).toBe('user.create');
      expect(oldestBody.items[0]?.actorName).toBeNull();
      expect(oldestBody.items[0]?.targetLabel).toBe(ADMIN.email);

      // Paginación que conserva los filtros aplicados.
      await page.getByLabel('Desde').fill(localDateInput(-1));
      await page.getByLabel('Hasta').fill(localDateInput(1));
      await page.getByRole('button', { name: 'Filtrar' }).click();
      await expect(summary).toHaveText(/Mostrando 1–\d+ de \d+ registros/);
      const nextPage = page.getByRole('button', { name: 'Siguiente' });
      if (await nextPage.isEnabled()) {
        await nextPage.click();
        await expect(summary).toHaveText(/Mostrando 21–\d+ de \d+ registros/);
        await expect(page.getByLabel('Desde')).toHaveValue(localDateInput(-1));
        await expect(page.getByLabel('Hasta')).toHaveValue(localDateInput(1));
      }
      await page.getByRole('button', { name: 'Quitar filtros' }).click();

      // ── FR-025/SC-013: no hay ningún control de edición ni borrado ────────
      await expect(page.getByRole('button', { name: /editar|eliminar|borrar/i })).toHaveCount(0);
      const deleteAudit = await apiSend(page, 'DELETE', '/api/v1/admin/auditoria/accesos');
      expect([404, 405]).toContain(deleteAudit.status());
      const patchAudit = await apiSend(page, 'PATCH', '/api/v1/admin/auditoria/acciones', {});
      expect([404, 405]).toContain(patchAudit.status());
      const postAudit = await apiSend(page, 'POST', '/api/v1/admin/auditoria/accesos', {});
      expect([404, 405]).toContain(postAudit.status());

      // ── Último acceso en la ficha (FR-021) ────────────────────────────────
      await nav.getByRole('link', { name: 'Usuarios' }).click();

      // Cuenta recién creada y nunca usada: sin accesos inventados.
      const nuevaRow = page.getByRole('row').filter({ hasText: nueva.email });
      await nuevaRow.getByRole('button', { name: 'Editar' }).click();
      const nuevaEditDialog = page.getByRole('dialog');
      await expect(nuevaEditDialog.getByText('Nunca ha entrado al panel.')).toBeVisible();
      await nuevaEditDialog.getByRole('button', { name: 'Cerrar' }).click();

      // Ana (la cuenta más antigua) sí tiene un acceso exitoso; al ordenarse el
      // listado por fecha de creación descendente, está en la última página.
      const usersSummary = page.getByRole('status').filter({ hasText: 'usuarios' });
      const usersSummaryText = (await usersSummary.textContent()) ?? '';
      const totalUsers = Number(/de (\d+) usuarios/.exec(usersSummaryText)?.[1] ?? '0');
      const totalUserPages = Math.max(1, Math.ceil(totalUsers / 20));
      for (let pageNumber = 1; pageNumber < totalUserPages; pageNumber += 1) {
        await page.getByRole('button', { name: 'Siguiente' }).click();
        await expect(usersSummary).toHaveText(new RegExp(`Mostrando ${pageNumber * 20 + 1}–`));
      }
      const anaRow = page.getByRole('row').filter({ hasText: ADMIN.email });
      await expect(anaRow).toBeVisible();
      await anaRow.getByRole('button', { name: 'Editar' }).click();
      const anaDialog = page.getByRole('dialog');
      const anaDetail = anaDialog.getByRole('region', { name: 'Detalle de la cuenta' });
      await expect(anaDetail.getByText('Último acceso')).toBeVisible();
      await expect(anaDetail.getByText('Nunca ha entrado al panel.')).toHaveCount(0);
      await expect(anaDetail.getByText(/\d{2}\/\d{2}\/\d{4}/)).toBeVisible();
      await anaDialog.getByRole('button', { name: 'Cerrar' }).click();

      await page.getByRole('button', { name: 'Salir' }).click();
      await expect(page).toHaveURL(/\/login/);
    } finally {
      await context.close();
    }
  });
});
