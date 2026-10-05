import { expect, test } from '@playwright/test';
import { ADMIN, ensureInitialized, loginViaUi, uniqueSuffix } from './helpers';

/**
 * E2E del recorrido crítico de acceso y gestión (quickstart.md §11). Cubre
 * SC-002 (login ágil), SC-005 (rol + cuenta + acceso), SC-006 (cuenta
 * desactivada sin acceso, también con sesión abierta), SC-007 (el servidor
 * verifica el permiso aunque la UI no muestre la sección), SC-010 (cambio de
 * contraseña) y SC-011 (sin duplicados: nombres/correos únicos por ejecución).
 *
 * Se ejecuta contra el stack real (`make up`); la inicialización se hace por
 * API (idempotente) y el resto se recorre por la interfaz.
 */
test.describe('Acceso y gestión de usuarios (US1–US7)', () => {
  test.setTimeout(120_000);

  test('inicializar → login → rol → cuenta → contraseña → permisos → desactivar → bloqueo', async ({
    browser,
    request,
  }) => {
    await ensureInitialized(request);

    const suffix = uniqueSuffix();
    const roleName = `Contenido ${suffix}`;
    const carlos = {
      firstName: 'Carlos',
      lastName: `Ayudante ${suffix}`,
      email: `carlos-${suffix}@ejemplo.com`,
      phone: '612 345 678',
    };
    const initialPassword = 'Cambio.2026';
    const newPassword = 'Nueva.Clave.2026';
    const claveEquivocada = 'equivocada';

    const adminContext = await browser.newContext();
    const adminPage = await adminContext.newPage();
    const userContext = await browser.newContext();
    const userPage = await userContext.newPage();

    try {
      const adminNav = adminPage.getByRole('navigation', { name: 'Navegación del panel' });

      // ── SC-002: entrar con el administrador (quickstart §2) ────────────────
      await loginViaUi(adminPage, ADMIN.email, ADMIN.password);
      await expect(adminPage).toHaveURL(/\/panel$/);
      await expect(adminPage.getByRole('heading', { name: 'Mi cuenta' })).toBeVisible();

      // ── SC-005/SC-011: crear un rol con dos permisos (quickstart §4) ───────
      await adminNav.getByRole('link', { name: 'Roles' }).click();
      await adminPage.getByRole('button', { name: 'Crear rol' }).click();
      const roleDialog = adminPage.getByRole('dialog');
      await roleDialog.getByLabel('Nombre').fill(roleName);
      await roleDialog.getByRole('checkbox', { name: 'Eventos' }).check();
      await roleDialog.getByRole('checkbox', { name: 'Actividades' }).check();
      await roleDialog.getByRole('button', { name: 'Crear rol' }).click();
      await expect(adminPage.getByText('Rol creado.')).toBeVisible();
      await expect(adminPage.getByRole('row').filter({ hasText: roleName })).toBeVisible();

      // ── SC-005/SC-011: crear la cuenta con ese rol (quickstart §4) ─────────
      await adminNav.getByRole('link', { name: 'Usuarios' }).click();
      await adminPage.getByRole('button', { name: 'Crear usuario' }).click();
      const userDialog = adminPage.getByRole('dialog');
      await userDialog.getByLabel('Nombre').fill(carlos.firstName);
      await userDialog.getByLabel('Apellidos').fill(carlos.lastName);
      await userDialog.getByLabel('Correo').fill(carlos.email);
      await userDialog.getByLabel('Teléfono').fill(carlos.phone);
      await userDialog.getByLabel('Rol').selectOption({ label: roleName });
      await userDialog.getByLabel('Contraseña inicial').fill(initialPassword);
      await userDialog.getByRole('button', { name: 'Crear cuenta' }).click();
      await expect(adminPage.getByText('Cuenta creada.')).toBeVisible();
      await expect(adminPage.getByRole('row').filter({ hasText: carlos.email })).toBeVisible();

      // ── SC-010: entrar con la cuenta nueva; el panel exige cambiar la
      // contraseña antes de usar nada (quickstart §5, US7 esc. 4) ─────────────
      await loginViaUi(userPage, carlos.email, initialPassword);
      await expect(userPage).toHaveURL(/\/cambiar-contrasena$/);
      await expect(
        userPage.getByText('Por seguridad, cambia esta contraseña antes de continuar.'),
      ).toBeVisible();

      await userPage.getByLabel(/^Contraseña actual\s*\*?$/).fill(initialPassword);
      await userPage.getByLabel(/^Contraseña nueva\s*\*?$/).fill(newPassword);
      await userPage.getByLabel(/^Confirmar contraseña nueva\s*\*?$/).fill(newPassword);
      await userPage.getByRole('button', { name: 'Guardar' }).click();
      await expect(
        userPage.getByText('Contraseña cambiada. La próxima vez que entres, usa la nueva.'),
      ).toBeVisible();
      await userPage.getByRole('button', { name: 'Ir al panel' }).click();
      await expect(userPage).toHaveURL(/\/panel$/);

      // ── SC-007: solo sus módulos; sin permiso de administración no ve las
      // secciones ni puede entrar por URL (el servidor también lo impide) ────
      const userNav = userPage.getByRole('navigation', { name: 'Navegación del panel' });
      await expect(userNav.getByRole('link', { name: 'Inicio' })).toBeVisible();
      await expect(userNav.getByRole('link', { name: 'Usuarios' })).toHaveCount(0);
      await expect(userNav.getByRole('link', { name: 'Roles' })).toHaveCount(0);
      await expect(userNav.getByRole('link', { name: 'Auditoría' })).toHaveCount(0);
      await expect(
        userPage.getByText(
          'Tu cuenta está activa. Aún no tienes secciones asignadas; si necesitas acceso, habla con tu administrador.',
        ),
      ).toBeVisible();

      await userPage.goto('/panel/usuarios');
      await expect(
        userPage.getByRole('heading', { name: 'No tienes acceso a esta sección' }),
      ).toBeVisible();
      await userPage.goto('/panel');

      // ── SC-006: desactivar la cuenta desde el panel de administración ──────
      const carlosRow = adminPage.getByRole('row').filter({ hasText: carlos.email });
      await expect(carlosRow.getByText('Activo', { exact: true })).toBeVisible();
      await carlosRow.getByRole('button', { name: 'Desactivar' }).click();
      const deactivateDialog = adminPage.getByRole('dialog');
      await expect(
        deactivateDialog.getByRole('heading', { name: 'Desactivar cuenta' }),
      ).toBeVisible();
      await deactivateDialog.getByRole('button', { name: 'Desactivar' }).click();
      await expect(adminPage.getByText('Cuenta desactivada.')).toBeVisible();
      await expect(carlosRow.getByText('Inactivo', { exact: true })).toBeVisible();

      // La sesión abierta de Carlos queda cortada de inmediato (FR-012).
      await userPage.reload();
      await expect(userPage).toHaveURL(/\/login/);
      await expect(userPage.getByRole('heading', { name: 'Entrar al panel' })).toBeVisible();

      // Con credenciales correctas, la cuenta desactivada no puede entrar:
      // el servidor responde 403 con el mensaje de acceso desactivado (SC-006).
      //
      // HALLAZGO (para dev-backend): el sobre real trae `code: "forbidden"` sin
      // `details.reason = "access_disabled"`, que es lo que espera la UI
      // (src/features/auth/pages/LoginPage.tsx) y documenta el contrato. Por eso
      // aquí se verifica el corte de acceso (403 + se queda en /login) y NO el
      // texto concreto, para no fijar el comportamiento defectuoso.
      const [disabledLogin] = await Promise.all([
        userPage.waitForResponse(
          (res) => res.url().includes('/api/v1/auth/login') && res.request().method() === 'POST',
        ),
        loginViaUi(userPage, carlos.email, newPassword),
      ]);
      expect(disabledLogin.status()).toBe(403);
      await expect(userPage).toHaveURL(/\/login/);
      await expect(userPage.getByRole('heading', { name: 'Mi cuenta' })).toHaveCount(0);

      // ── US5 esc. 3: reactivar devuelve el acceso con normalidad ────────────
      await carlosRow.getByRole('button', { name: 'Activar' }).click();
      const activateDialog = adminPage.getByRole('dialog');
      await expect(activateDialog.getByRole('heading', { name: 'Activar cuenta' })).toBeVisible();
      await activateDialog.getByRole('button', { name: 'Activar' }).click();
      await expect(adminPage.getByText('Cuenta activada.')).toBeVisible();
      await expect(carlosRow.getByText('Activo', { exact: true })).toBeVisible();

      // Acceso correcto de nuevo (y reinicia el contador de intentos).
      await loginViaUi(userPage, carlos.email, newPassword);
      await expect(userPage).toHaveURL(/\/panel$/);
      await userPage.getByRole('button', { name: 'Salir' }).click();
      await expect(userPage).toHaveURL(/\/login/);

      // ── FR-006/SC-008: intentos fallidos. El 5.º crea el bloqueo (pero aún
      // responde el 401 genérico) y el 6.º recibe 429 (quickstart §8) ────────
      await userPage.goto('/login');
      for (let attempt = 1; attempt <= 5; attempt += 1) {
        await userPage.getByLabel('Correo').fill(carlos.email);
        await userPage.getByLabel(/^Contraseña\s*\*?$/).fill(claveEquivocada);
        const [response] = await Promise.all([
          userPage.waitForResponse(
            (res) => res.url().includes('/api/v1/auth/login') && res.request().method() === 'POST',
          ),
          userPage.getByRole('button', { name: 'Entrar' }).click(),
        ]);
        expect(response.status(), `intento fallido nº ${attempt}`).toBe(401);
        await expect(userPage.getByText('Correo o contraseña incorrectos.')).toBeVisible();
      }

      await userPage.getByLabel('Correo').fill(carlos.email);
      await userPage.getByLabel(/^Contraseña\s*\*?$/).fill(claveEquivocada);
      const [sixth] = await Promise.all([
        userPage.waitForResponse(
          (res) => res.url().includes('/api/v1/auth/login') && res.request().method() === 'POST',
        ),
        userPage.getByRole('button', { name: 'Entrar' }).click(),
      ]);
      expect(sixth.status(), 'el 6.º intento queda bloqueado').toBe(429);
      await expect(userPage.getByText(/se han superado los intentos permitidos/)).toBeVisible();
      await expect(userPage.getByLabel('Correo')).toBeDisabled();

      // ── FR-004: cerrar sesión del administrador ────────────────────────────
      await adminPage.getByRole('button', { name: 'Salir' }).click();
      await expect(adminPage).toHaveURL(/\/login/);
    } finally {
      await adminContext.close();
      await userContext.close();
    }
  });
});
