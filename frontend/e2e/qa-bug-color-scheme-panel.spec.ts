import { expect, test } from '@playwright/test';
import { createAdminContext } from './publicSeed';

/**
 * QA · Regresión del bug «indicador invisible de los nativos de fecha/hora en
 * el panel» (⟲ QA, no modifica código de producción):
 *
 * F10 fija `color-scheme` globalmente en `<html>` (`index.html:32` según el
 * tema del sitio público, y también `ThemeProvider`). El panel es **siempre
 * claro** (FR-016/P10-11) pero **no resetea `color-scheme`**: con el tema
 * público en oscuro, `getComputedStyle` resuelve `dark` en los contenedores e
 * inputs del panel y el navegador pinta el indicador nativo del selector
 * (icono de calendario/reloj de `<input type="date">`/`type="time"`) en color
 * claro sobre el panel blanco → invisible, aunque el text siga legible por
 * llevar color explícito. Consecuencia práctica: la persona teclea la fecha a
 * ciegas y el icono de apertura del calendario «no está».
 *
 * Arreglo previsto (*lo hace `dev-frontend`, no QA*): forzar
 * `color-scheme: light` en los contenedores de superficie clara del panel
 * (`AppLayout`, `PanelLayout`, `PanelSurfaceLayout` en
 * `frontend/src/app/layout.tsx`). Esta prueba exige que:
 *
 *   1. El tema oscuro siga aplicado en `<html>` (`data-theme="dark"` y
 *      `color-scheme: dark` en el `style` del documento): el sitio público no
 *      debe verse afectado por el arreglo.
 *   2. Los campos de la fila de fechas del formulario «Agregar evento» sigan
 *      siendo nativos (`type="date"` y `type="time"`): la solución **no** es
 *      degradarlos a inputs de texto.
 *   3. El `color-scheme` **resuelto** en la superficie del panel (`main`, dentro
 *      de `PanelLayout`) y en los propios inputs sea `light`. Hoy falla en
 *      (3): resuelve `dark` (heredado de `<html>`). Tras el arreglo, debe
 *      pasar.
 *
 * Verificación del indicador del pseudo-elemento: se evaluó
 * `getComputedStyle(input, '::-webkit-calendar-picker-indicator')` y resulta
 * **no fiable** para detectar el bug (la `opacity`/`display` no cambian con el
 * `color-scheme`; lo cambia el color del icono pintado por el navegador, que
 * este computed style expone de forma estable en distintas imágenes). Por eso
 * la aserción principal es el `color-scheme` resuelto, que sí es estable y es
 * exactamente la propiedad que causa el bug.
 *
 * Ejecutar contra el stack levantado (`make up`); el host WSL carece de
 * `libnspr4`, así que se usa la imagen de Playwright. Con command el **fallo
 * esperado hoy** es: `color-scheme resuelto en la superficie del panel`
 * `Received: "dark"` en lugar de `"light"`:
 *
 *   docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
 *     -v "$PWD/..":/work -w /work/frontend mcr.microsoft.com/playwright:v1.63.0-jammy \
 *     npx playwright test --config e2e/playwright.config.ts qa-bug-color-scheme-panel --workers=1
 */

test.describe('Bug: color-scheme oscuro filtrado al panel (nativos de fecha/hora)', () => {
  test.setTimeout(120_000);

  test('tema oscuro del dispositivo → la superficie del panel y los nativos de fecha/hora resuelven color-scheme: light', async ({
    browser,
  }) => {
    // Sesión de administrador cacheada (un solo login por proceso, patrón F10)
    // con el tema del **dispositivo** emulado en oscuro y sin elección guardada
    // en `ss.theme`: el anti-FOUC de `index.html:32` aplicará
    // `colorScheme: 'dark'` en `<html>`, tal como lo ve una visita de noche.
    const context = await createAdminContext(browser);
    try {
      const page = await context.newPage();
      await page.emulateMedia({ colorScheme: 'dark' });
      await page.goto('/panel');

      // El tema oscuro sí quedó aplicado en el documento (el arreglo no debe
      // quitarlo: solo el panel deja de heredarlo).
      expect(await page.evaluate(() => document.documentElement.dataset.theme)).toBe('dark');
      expect(await page.evaluate(() => document.documentElement.style.colorScheme)).toBe('dark');

      // Hasta el formulario de creación de evento (/panel/eventos).
      await page
        .getByRole('navigation', { name: 'Navegación del panel' })
        .getByRole('link', { name: 'Gestión de eventos' })
        .click();
      await expect(page.getByRole('heading', { name: 'Gestión de eventos' })).toBeVisible();
      await page.getByRole('button', { name: 'Agregar evento' }).click();

      const dialog = page.getByRole('dialog');
      await expect(dialog).toBeVisible();

      // Los selectores nativos deben conservar su tipo: el arreglo de la
      // superficie no debería introducir la regresión contraria (degradarlos
      // la solución «fácil» de degradarlos a input de texto pierde el picker
      // en móvil, que era una decisión firme de F4/ux.md §4.4).
      const fecha = dialog.getByLabel(/^Fecha 1/);
      const inicio = dialog.getByLabel(/^Hora de inicio/);
      const fin = dialog.getByLabel(/^Hora de fin/);

      await expect(fecha, 'la fila de fechas existe con su input nativo').toBeVisible();
      await expect(inicio).toBeVisible();
      await expect(fin).toBeVisible();

      expect(await fecha.evaluate((el) => (el as HTMLInputElement).type)).toBe('date');
      expect(await inicio.evaluate((el) => (el as HTMLInputElement).type)).toBe('time');
      expect(await fin.evaluate((el) => (el as HTMLInputElement).type)).toBe('time');

      await expect(
        dialog.locator('input[type="date"]'),
        'al menos un selector nativo de fecha en el diálogo',
      ).not.toHaveCount(0);
      await expect(dialog.locator('input[type="time"]')).not.toHaveCount(0);

      // Núcleo del bug: el color-scheme RESUELTO. En la superficie del panel
      // (`main`, dentro de `PanelLayout`) y en los propios inputs debe ser
      // `light`: con `dark`, el navegador pinta el indicador nativo del
      // selector en color claro sobre el panel blanco (icono invisible).
      const esquemaSuperficie = await page
        .locator('main')
        .first()
        .evaluate((el) => getComputedStyle(el).colorScheme);
      expect(esquemaSuperficie, 'color-scheme resuelto en la superficie del panel').toBe('light');

      const esquemaFecha = await fecha.evaluate((el) => getComputedStyle(el).colorScheme);
      const esquemaInicio = await inicio.evaluate((el) => getComputedStyle(el).colorScheme);
      const esquemaFin = await fin.evaluate((el) => getComputedStyle(el).colorScheme);
      expect(esquemaFecha, 'color-scheme del input de fecha').toBe('light');
      expect(esquemaInicio, 'color-scheme del input de hora de inicio').toBe('light');
      expect(esquemaFin, 'color-scheme del input de hora de fin').toBe('light');
    } finally {
      await context.close();
    }
  });
});
