import { expect, test, type Page } from '@playwright/test';
import { join } from 'node:path';
import { ADMIN, API_URL, apiSend, ensureInitialized, loginViaUi, uniqueSuffix } from './helpers';

/**
 * E2E del panel de la portada (T338, US2 · US3 panel · FR-012) sobre
 * `quickstart.md` §2–§6, §8 y §9. Cubre SC-003/SC-004 (publicar y ver el
 * cambio desde la primera carga), SC-005/SC-011 (permiso verificado en el
 * servidor) y SC-013 (todo queda auditado en F2).
 *
 * Se ejecuta contra el stack real (`make up` + `make db-migrate`). La cuenta
 * `ADMIN` de F2 (rol «Administrador», con todos los permisos) siembra un rol y
 * dos cuentas por API; luego se recorre la interfaz. La identidad de la iglesia
 * es un singleton compartido: **ejecutar con `--workers=1`** (`make e2e`) para
 * que esta spec no compita con `portada-publica.spec.ts`.
 */

/** Imagen real del repositorio, usada para la subida del logotipo (R3-8). */
const FIXTURE_IMAGE = join(process.cwd(), '..', 'resources', 'simiente.jpeg');

interface RoleCreated {
  id: string;
  name: string;
}

interface UserCreated {
  id: string;
  email: string;
}

interface AuditItem {
  action: string;
  actorName: string | null;
  actorEmail: string | null;
  targetLabel: string | null;
  result: string;
}

interface AuditList {
  items: AuditItem[];
  total: number;
}

interface Account {
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  initialPassword: string;
  newPassword: string;
}

/** Singletons del panel presentes en `GET /api/v1/admin/portada`. */
interface AdminState {
  identity: {
    nameEs: string;
    nameEn: string | null;
    taglineEs: string | null;
    taglineEn: string | null;
    missionEs: string | null;
    missionEn: string | null;
    visionEs: string | null;
    visionEn: string | null;
    logoFile: string | null;
    logoAltEs: string | null;
    logoAltEn: string | null;
    coverImageFile: string | null;
    coverImageAltEs: string | null;
    coverImageAltEn: string | null;
  } | null;
  about: { textEs: string; textEn: string | null } | null;
  contact: {
    addressEs: string;
    addressEn: string | null;
    email: string;
    phone: string;
  } | null;
}

/**
 * Deja identidad, «quiénes somos» y contacto en **borrador** (la identidad es
 * un singleton compartido con `portada-publica.spec.ts`). La prueba asume que
 * sus secciones parten sin publicar —el flujo guarda y luego publica—, así que
 * se restablece esa precondición sin importar el orden ni las corridas previas.
 * Si un singleton no existe todavía, no hay nada que restablecer.
 */
async function resetSingletonsToDraft(page: Page): Promise<void> {
  const state = (await (await apiSend(page, 'GET', '/api/v1/admin/portada')).json()) as AdminState;

  if (state.identity) {
    const { identity } = state;
    expect(
      (
        await apiSend(page, 'PUT', '/api/v1/admin/portada/identidad', {
          nameEs: identity.nameEs,
          nameEn: identity.nameEn,
          taglineEs: identity.taglineEs,
          taglineEn: identity.taglineEn,
          missionEs: identity.missionEs,
          missionEn: identity.missionEn,
          visionEs: identity.visionEs,
          visionEn: identity.visionEn,
          logoFile: identity.logoFile,
          logoAltEs: identity.logoAltEs,
          logoAltEn: identity.logoAltEn,
          coverImageFile: identity.coverImageFile,
          coverImageAltEs: identity.coverImageAltEs,
          coverImageAltEn: identity.coverImageAltEn,
          publicationState: 'draft',
        })
      ).ok(),
    ).toBe(true);
  }

  if (state.about) {
    expect(
      (
        await apiSend(page, 'PUT', '/api/v1/admin/portada/quienes-somos', {
          textEs: state.about.textEs,
          textEn: state.about.textEn,
          publicationState: 'draft',
        })
      ).ok(),
    ).toBe(true);
  }

  if (state.contact) {
    expect(
      (
        await apiSend(page, 'PUT', '/api/v1/admin/portada/contacto', {
          addressEs: state.contact.addressEs,
          addressEn: state.contact.addressEn,
          email: state.contact.email,
          phone: state.contact.phone,
          publicationState: 'draft',
        })
      ).ok(),
    ).toBe(true);
  }
}

/** Inicia sesión con una cuenta nueva y resuelve el cambio obligatorio (F2). */
async function loginAndChangePassword(page: Page, account: Account): Promise<void> {
  await loginViaUi(page, account.email, account.initialPassword);
  await expect(page).toHaveURL(/\/cambiar-contrasena$/);
  await page.getByLabel(/^Contraseña actual\s*\*?$/).fill(account.initialPassword);
  await page.getByLabel(/^Contraseña nueva\s*\*?$/).fill(account.newPassword);
  await page.getByLabel(/^Confirmar contraseña nueva\s*\*?$/).fill(account.newPassword);
  await page.getByRole('button', { name: 'Guardar' }).click();
  await expect(
    page.getByText('Contraseña cambiada. La próxima vez que entres, usa la nueva.'),
  ).toBeVisible();
  await page.getByRole('button', { name: 'Ir al panel' }).click();
  await expect(page).toHaveURL(/\/panel$/);
}

test.describe('Panel de la portada (editar, publicar, permisos y auditoría)', () => {
  test.setTimeout(300_000);

  test('editor con permiso gestiona y publica; sin permiso queda denegado y auditado', async ({
    browser,
    request,
  }) => {
    await ensureInitialized(request);

    const suffix = uniqueSuffix();
    const editorRoleName = `Contenido ${suffix}`;
    const deniedRoleName = `Sin portada ${suffix}`;
    const editor: Account = {
      firstName: 'Elena',
      lastName: `Editora ${suffix}`,
      email: `editora-${suffix}@ejemplo.com`,
      phone: '+506 7000 5555',
      initialPassword: 'Inicial.2026',
      newPassword: 'Nueva.Clave.2026',
    };
    const denied: Account = {
      firstName: 'Sara',
      lastName: `SinPermiso ${suffix}`,
      email: `sinpermiso-${suffix}@ejemplo.com`,
      phone: '+506 7000 6666',
      initialPassword: 'Inicial.2026',
      newPassword: 'Nueva.Clave.2026',
    };
    const editorDisplayName = `${editor.firstName} ${editor.lastName}`;
    const deniedDisplayName = `${denied.firstName} ${denied.lastName}`;

    const identityName = `Iglesia Simiente ${suffix}`;
    const identityNameEn = `Simiente Church ${suffix}`;
    const mission = `Misión ${suffix}`;
    const aboutText = `Quiénes somos ${suffix}: una iglesia cercana.`;
    const address = `San José, ${suffix}`;
    const email = `hola-${suffix}@ejemplo.com`;
    const phone = '+506 7000 4444';
    const serviceName = `Culto ${suffix}`;
    const channelName = `Escríbenos ${suffix}`;
    const channelPhone = '+506 7000 3333';
    const instagramUrl = `https://www.instagram.com/simiente-${suffix}`;
    let logoFileName: string;

    const adminContext = await browser.newContext();
    const adminPage = await adminContext.newPage();

    try {
      // ── Cuentas de prueba (FR-012): una con permiso y otra sin él ─────────
      await loginViaUi(adminPage, ADMIN.email, ADMIN.password);
      await expect(adminPage).toHaveURL(/\/panel$/);

      // Precondición determinista: las secciones que esta prueba guarda y luego
      // publica parten en borrador (la base es compartida y no se limpia).
      await resetSingletonsToDraft(adminPage);

      const editorRoleResponse = await apiSend(adminPage, 'POST', '/api/v1/admin/roles', {
        name: editorRoleName,
        permissions: ['portada'],
      });
      expect(editorRoleResponse.status()).toBe(201);
      const editorRole = (await editorRoleResponse.json()) as RoleCreated;

      const editorResponse = await apiSend(adminPage, 'POST', '/api/v1/admin/usuarios', {
        firstName: editor.firstName,
        lastName: editor.lastName,
        email: editor.email,
        phone: editor.phone,
        roleId: editorRole.id,
        password: editor.initialPassword,
      });
      expect(editorResponse.status()).toBe(201);
      const editorUser = (await editorResponse.json()) as UserCreated;

      const deniedRoleResponse = await apiSend(adminPage, 'POST', '/api/v1/admin/roles', {
        name: deniedRoleName,
        permissions: ['eventos'],
      });
      expect(deniedRoleResponse.status()).toBe(201);
      const deniedRole = (await deniedRoleResponse.json()) as RoleCreated;

      const deniedResponse = await apiSend(adminPage, 'POST', '/api/v1/admin/usuarios', {
        firstName: denied.firstName,
        lastName: denied.lastName,
        email: denied.email,
        phone: denied.phone,
        roleId: deniedRole.id,
        password: denied.initialPassword,
      });
      expect(deniedResponse.status()).toBe(201);

      // ── El editor entra y ve el módulo (US2) ─────────────────────────────
      const editorContext = await browser.newContext();
      const editorPage = await editorContext.newPage();
      const publicContext = await browser.newContext();
      const publicPage = await publicContext.newPage();

      try {
        await loginAndChangePassword(editorPage, editor);
        const editorNav = editorPage.getByRole('navigation', { name: 'Navegación del panel' });
        await expect(
          editorNav.getByRole('link', { name: 'Portada e información general' }),
        ).toBeVisible();

        // El módulo precarga con el agregado `GET /api/v1/admin/portada` (T340).
        const [aggregate] = await Promise.all([
          editorPage.waitForResponse(
            (res) =>
              res.url().includes('/api/v1/admin/portada') && res.request().method() === 'GET',
          ),
          editorNav.getByRole('link', { name: 'Portada e información general' }).click(),
        ]);
        expect(aggregate.status()).toBe(200);
        const aggregateBody = (await aggregate.json()) as Record<string, unknown>;
        for (const key of ['identity', 'about', 'contact', 'schedule', 'whatsapp', 'socials']) {
          expect(aggregateBody, `el agregado trae ${key}`).toHaveProperty(key);
        }

        await expect(editorPage).toHaveURL(/\/panel\/informacion$/);
        await expect(
          editorPage.getByRole('heading', { name: 'Portada e información general' }),
        ).toBeVisible();
        await expect(
          editorPage.getByRole('navigation', { name: 'Secciones de la portada' }),
        ).toBeVisible();

        // ── Identidad (FR-002/FR-015/FR-019): error junto al campo, en y logo ─
        await expect(editorPage.getByRole('heading', { name: 'Identidad' })).toBeVisible();
        await editorPage.getByLabel(/^Nombre oficial\s*\*?$/).fill('');
        await editorPage.getByRole('button', { name: 'Guardar' }).click();
        await expect(editorPage.getByText('Escribe el nombre.')).toBeVisible();

        await editorPage.getByLabel(/^Nombre oficial\s*\*?$/).fill(identityName);
        await editorPage.getByRole('button', { name: 'English (opcional)' }).click();
        await editorPage.getByLabel(/^Nombre oficial \(English\)$/).fill(identityNameEn);
        await editorPage.getByRole('button', { name: 'Español', exact: true }).click();
        await editorPage.getByLabel(/^Misión \(Español\)$/).fill(mission);

        const logoGroup = editorPage.getByRole('group', { name: 'Logotipo' });
        const [uploadResponse] = await Promise.all([
          editorPage.waitForResponse(
            (res) =>
              res.url().includes('/api/v1/admin/portada/imagenes') &&
              res.request().method() === 'POST',
          ),
          logoGroup.getByLabel(/^Archivo$/).setInputFiles(FIXTURE_IMAGE),
        ]);
        expect(uploadResponse.status()).toBe(201);
        logoFileName = ((await uploadResponse.json()) as { fileName: string }).fileName;
        await logoGroup
          .getByLabel('Texto alternativo (Español)')
          .fill(`Logotipo de ${identityName}`);

        await editorPage.getByRole('button', { name: 'Guardar' }).click();
        await expect(editorPage.getByText('Cambios guardados.')).toBeVisible();
        await editorPage.getByRole('button', { name: 'Publicar' }).click();
        await expect(editorPage.getByText('Publicado en la portada.')).toBeVisible();
        await expect(editorPage.getByText('Publicado', { exact: true }).first()).toBeVisible();

        // ── Quiénes somos (FR-003): publicar por sección ──────────────────────
        await editorPage.getByRole('button', { name: 'Quiénes somos' }).click();
        await editorPage.getByLabel(/^Texto \(Español\)\s*\*?$/).fill(aboutText);
        await editorPage.getByRole('button', { name: 'Guardar' }).click();
        await expect(editorPage.getByText('Cambios guardados.')).toBeVisible();
        await editorPage.getByRole('button', { name: 'Publicar' }).click();
        await expect(editorPage.getByText('Publicado en la portada.')).toBeVisible();

        // ── Contacto (FR-007): publicar por sección ───────────────────────────
        await editorPage.getByRole('button', { name: 'Contacto' }).click();
        await editorPage.getByLabel(/^Dirección \(Español\)\s*\*?$/).fill(address);
        await editorPage.getByLabel(/^Correo\s*\*?$/).fill(email);
        await editorPage.getByLabel(/^Teléfono\s*\*?$/).fill(phone);
        await editorPage.getByRole('button', { name: 'Guardar' }).click();
        await expect(editorPage.getByText('Cambios guardados.')).toBeVisible();
        await editorPage.getByRole('button', { name: 'Publicar' }).click();
        await expect(editorPage.getByText('Publicado en la portada.')).toBeVisible();

        // ── Horario (FR-004/C2): alta en borrador y publicar por elemento ─────
        await editorPage.getByRole('button', { name: 'Horario de servicios' }).click();
        await editorPage.getByRole('button', { name: 'Agregar servicio' }).click();
        const serviceDialog = editorPage.getByRole('dialog');
        await serviceDialog.getByLabel('Día').selectOption('0');
        await serviceDialog.getByLabel('Hora de inicio').fill('10:00');
        await serviceDialog.getByLabel('Hora de fin (opcional)').fill('12:00');
        await serviceDialog.getByLabel(/^Nombre \(Español\)\s*\*?$/).fill(serviceName);
        await serviceDialog.getByLabel(/^Lugar \(Español\)\s*\*?$/).fill('Templo principal');
        await serviceDialog.getByRole('button', { name: 'Agregar servicio' }).click();
        await expect(editorPage.getByText('Servicio creado.')).toBeVisible();

        const scheduleTable = editorPage.getByRole('table', { name: 'Servicios del horario' });
        const serviceRow = scheduleTable.getByRole('row').filter({ hasText: serviceName });
        await expect(serviceRow.getByText('Borrador')).toBeVisible();
        await serviceRow.getByRole('button', { name: 'Publicar' }).click();
        await expect(editorPage.getByText('Publicado en la portada.')).toBeVisible();
        await expect(serviceRow.getByText('Publicado')).toBeVisible();

        // ── WhatsApp (FR-005): alta, publicar y duplicado exacto (409) ────────
        await editorPage.getByRole('button', { name: 'WhatsApp' }).click();
        await editorPage.getByRole('button', { name: 'Agregar canal' }).click();
        const channelDialog = editorPage.getByRole('dialog');
        await channelDialog.getByLabel(/^Nombre o propósito \(Español\)\s*\*?$/).fill(channelName);
        await channelDialog.getByLabel(/^Número\s*\*?$/).fill(channelPhone);
        await channelDialog.getByRole('button', { name: 'Agregar canal' }).click();
        await expect(editorPage.getByText('Canal de WhatsApp creado.')).toBeVisible();

        const channelTable = editorPage.getByRole('table', { name: 'Canales de WhatsApp' });
        const channelRow = channelTable.getByRole('row').filter({ hasText: channelName });
        await channelRow.getByRole('button', { name: 'Publicar' }).click();
        await expect(editorPage.getByText('Publicado en la portada.')).toBeVisible();

        // Duplicado exacto → el servidor responde 409 y el mensaje queda junto al campo.
        await editorPage.getByRole('button', { name: 'Agregar canal' }).click();
        const duplicateDialog = editorPage.getByRole('dialog');
        await duplicateDialog
          .getByLabel(/^Nombre o propósito \(Español\)\s*\*?$/)
          .fill(channelName);
        await duplicateDialog.getByLabel(/^Número\s*\*?$/).fill(channelPhone);
        await duplicateDialog.getByRole('button', { name: 'Agregar canal' }).click();
        await expect(
          duplicateDialog.getByText(
            'Ya existe un canal igual (mismo nombre y mismo destino). Edita el que ya tienes o cámbiale el nombre.',
          ),
        ).toBeVisible();
        await duplicateDialog.getByRole('button', { name: 'Cancelar' }).click();

        // ── Redes (FR-006): una por red y publicar por elemento ───────────────
        await editorPage.getByRole('button', { name: 'Redes sociales' }).click();
        const socialTable = editorPage.getByRole('table', { name: 'Redes sociales' });
        const instagramRow = socialTable.getByRole('row').filter({ hasText: 'Instagram' });
        if ((await instagramRow.getByRole('button', { name: 'Agregar enlace' }).count()) > 0) {
          await instagramRow.getByRole('button', { name: 'Agregar enlace' }).click();
        } else {
          await instagramRow.getByRole('button', { name: 'Editar' }).click();
        }
        const socialDialog = editorPage.getByRole('dialog');
        await socialDialog.getByLabel('Enlace de Instagram').fill(instagramUrl);
        await socialDialog.getByRole('button', { name: 'Guardar' }).click();
        await expect(editorPage.getByText('Enlace de Instagram guardado.')).toBeVisible();
        if ((await instagramRow.getByRole('button', { name: 'Publicar' }).count()) > 0) {
          await instagramRow.getByRole('button', { name: 'Publicar' }).click();
          await expect(editorPage.getByText('Publicado en la portada.')).toBeVisible();
        }

        // ── SC-003/SC-004: lo publicado se ve en la portada desde la 1.ª carga ─
        await publicPage.goto('/');
        await expect(
          publicPage.getByRole('heading', { level: 1, name: identityName }),
        ).toBeVisible();
        await expect(publicPage.locator('#horario').getByText(serviceName).first()).toBeVisible();
        await expect(
          publicPage.locator('#whatsapp li').filter({ hasText: channelName }),
        ).toBeVisible();

        // ── Retirar la identidad (por sección) y su imagen deja de servirse ───
        await editorPage.getByRole('button', { name: 'Identidad' }).click();
        await editorPage.getByRole('button', { name: 'Retirar de la portada' }).click();
        const unpublishDialog = editorPage.getByRole('dialog');
        await expect(
          unpublishDialog.getByRole('heading', { name: 'Retirar de la portada' }),
        ).toBeVisible();
        await unpublishDialog.getByRole('button', { name: 'Retirar' }).click();
        await expect(
          editorPage.getByText(
            'Retirado de la portada. Estará disponible como borrador para publicarlo cuando quieras.',
          ),
        ).toBeVisible();

        await publicPage.reload();
        await expect(publicPage.getByRole('heading', { level: 1, name: identityName })).toHaveCount(
          0,
        );
        expect((await request.get(`${API_URL}/api/v1/media/${logoFileName}`)).status()).toBe(404);

        await editorPage.getByRole('button', { name: 'Publicar' }).click();
        await expect(editorPage.getByText('Publicado en la portada.')).toBeVisible();
        await publicPage.reload();
        await expect(
          publicPage.getByRole('heading', { level: 1, name: identityName }),
        ).toBeVisible();
        expect((await request.get(`${API_URL}/api/v1/media/${logoFileName}`)).status()).toBe(200);

        // ── Retirar un elemento (por elemento) y volver a publicarlo ──────────
        await editorPage.getByRole('button', { name: 'Horario de servicios' }).click();
        const scheduleTable2 = editorPage.getByRole('table', { name: 'Servicios del horario' });
        const serviceRow2 = scheduleTable2.getByRole('row').filter({ hasText: serviceName });
        await serviceRow2.getByRole('button', { name: 'Retirar de la portada' }).click();
        const elementDialog = editorPage.getByRole('dialog');
        await elementDialog.getByRole('button', { name: 'Retirar' }).click();
        await expect(
          editorPage.getByText(
            'Retirado de la portada. Estará disponible como borrador para publicarlo cuando quieras.',
          ),
        ).toBeVisible();

        await publicPage.reload();
        await expect(publicPage.locator('#horario').getByText(serviceName)).toHaveCount(0);

        await serviceRow2.getByRole('button', { name: 'Publicar' }).click();
        await expect(editorPage.getByText('Publicado en la portada.')).toBeVisible();
        await publicPage.reload();
        await expect(publicPage.locator('#horario').getByText(serviceName).first()).toBeVisible();
      } finally {
        await editorContext.close();
        await publicContext.close();
      }

      // ── SC-005/SC-011: una cuenta sin permiso queda fuera por todas las vías ─
      const deniedContext = await browser.newContext();
      const deniedPage = await deniedContext.newPage();
      try {
        await loginAndChangePassword(deniedPage, denied);
        const deniedNav = deniedPage.getByRole('navigation', { name: 'Navegación del panel' });
        await expect(
          deniedNav.getByRole('link', { name: 'Portada e información general' }),
        ).toHaveCount(0);

        await deniedPage.goto('/panel/informacion');
        await expect(deniedPage).toHaveURL(/\/sin-permiso$/);
        await expect(
          deniedPage.getByRole('heading', { name: 'No tienes acceso a esta sección' }),
        ).toBeVisible();

        const forced = await apiSend(deniedPage, 'PUT', '/api/v1/admin/portada/contacto', {
          addressEs: 'x',
          email: 'a@b.co',
          phone: '8888888',
          publicationState: 'draft',
        });
        expect(forced.status()).toBe(403);
        const forcedBody = (await forced.json()) as { error?: { code?: string } };
        expect(forcedBody.error?.code).toBe('forbidden');
      } finally {
        await deniedContext.close();
      }

      // ── SC-013: todo queda registrado (FR-017) ────────────────────────────
      const auditResponse = await apiSend(
        adminPage,
        'GET',
        `/api/v1/admin/auditoria/acciones?userId=${editorUser.id}&limit=100`,
      );
      expect(auditResponse.ok()).toBe(true);
      const audit = (await auditResponse.json()) as AuditList;
      const actions = audit.items.map((item) => item.action);
      for (const expected of [
        'home.image.upload',
        'home.identity.update',
        'home.about.update',
        'home.contact.update',
        'home.schedule.create',
        'home.whatsapp.create',
        'home.publish',
        'home.unpublish',
      ]) {
        expect(actions, `la auditoría del editor registra ${expected}`).toContain(expected);
      }
      // La red social se **crea** si no existía o se **actualiza** si ya había un
      // enlace de esa red (FR-006: una por red); en ambos casos queda auditada.
      expect(
        actions.some(
          (action) => action === 'home.social.create' || action === 'home.social.update',
        ),
        'la auditoría del editor registra la gestión de la red social',
      ).toBe(true);
      // El filtro `userId` es por cuenta **involucrada** (F2/plan P22): incluye
      // la creación de la cuenta por la administradora (ella es el actor, la
      // editora el objetivo). Se comprueba el nombre en las acciones **hechas**
      // por la editora.
      expect(
        audit.items
          .filter((item) => item.actorEmail === editor.email)
          .every((item) => item.actorName === editorDisplayName),
      ).toBe(true);
      expect(audit.items.some((item) => (item.targetLabel ?? '').startsWith('Portada · '))).toBe(
        true,
      );

      // La denegación se registra con `result='denied'` (FR-017).
      const deniedAudit = (await (
        await apiSend(adminPage, 'GET', `/api/v1/admin/auditoria/acciones?limit=100`)
      ).json()) as AuditList;
      const denial = deniedAudit.items.find(
        (item) => item.actorEmail === denied.email && item.result === 'denied',
      );
      expect(denial, 'la denegación queda registrada').toBeDefined();
      expect(denial?.action).toBe('home.contact.update');

      // Y también se ve en la interfaz de auditoría de F2 (sin controles de edición).
      const adminNav = adminPage.getByRole('navigation', { name: 'Navegación del panel' });
      await adminNav.getByRole('link', { name: 'Auditoría' }).click();
      await adminPage.getByRole('button', { name: 'Acciones administrativas' }).click();
      const actionsTable = adminPage.getByRole('table', {
        name: 'Historial de acciones administrativas',
      });
      await adminPage.getByLabel('Cuenta').selectOption({ label: editorDisplayName });
      await adminPage.getByRole('button', { name: 'Filtrar' }).click();
      await expect(actionsTable.getByText('home.identity.update').first()).toBeVisible();
      // La subida del logotipo es la **primera** acción del editor; el historial
      // muestra 20 por página, así que puede quedar en la página siguiente.
      if ((await actionsTable.getByText('home.image.upload').count()) === 0) {
        const [pageResponse] = await Promise.all([
          adminPage.waitForResponse(
            (res) =>
              res.url().includes('/api/v1/admin/auditoria/acciones') &&
              res.request().method() === 'GET',
          ),
          adminPage.getByRole('button', { name: 'Siguiente' }).click(),
        ]);
        expect(pageResponse.ok()).toBe(true);
      }
      await expect(actionsTable.getByText('home.image.upload').first()).toBeVisible();
      await expect(actionsTable.getByText('Completada').first()).toBeVisible();
      await expect(actionsTable.getByText(/Portada · /).first()).toBeVisible();

      await adminPage.getByLabel('Cuenta').selectOption({ label: deniedDisplayName });
      await adminPage.getByRole('button', { name: 'Filtrar' }).click();
      await expect(actionsTable.getByText('home.contact.update').first()).toBeVisible();
      await expect(actionsTable.getByText('Denegada').first()).toBeVisible();
    } finally {
      await adminContext.close();
    }
  });
});
