import { expect, test, type Locator, type Page } from '@playwright/test';
import { createAdminContext } from './publicSeed';

/**
 * QA ⟲ (2026-10-17, 3.ª iteración): criterio que las specs anteriores del bug
 * del date picker del panel no cubren: **los botones con icono de una misma
 * fila de fecha/hora deben quedar a la misma altura**.
 *
 * En el formulario «Agregar evento», `DatesEditor` pinta una rejilla con la
 * fecha, la hora de inicio y la hora de fin. Las etiquetas no ocupan lo mismo
 * («Fecha 1» cabe en una línea; «Hora de inicio (opcional)» y «Hora de fin
 * (opcional)» se parten en dos al ancho del diálogo) y la fila alineaba por
 * arriba (`sm:items-start`): cada input —y su botón de icono— empezaba a
 * distinta altura, así que los iconos quedaban desalineados verticalmente.
 *
 * La prueba mide los `boundingBox()` reales de los tres botones de la primera
 * fila y exige que sus centros verticales coincidan con tolerancia de ±1 px.
 * Falla antes del arreglo (las etiquetas de dos líneas bajan su icono ~1
 * renglón) y pasa después.
 *
 * Nota de revisión de los otros consumidores de `Field`:
 *   - `DateRangeFilter` (filtros de Auditoría) ya alinea con `items-end` y sus
 *     etiquetas («Desde»/«Hasta») ocupan una línea; no sufre el desalineado.
 *   - `ServiceForm` (horarios de servicios) apila los campos en vertical
 *     (`space-y-4`), no comparte fila: no aplica.
 *
 * Ejecutar desde la raíz del repositorio contra el stack levantado (`make up`),
 * reconstruyendo antes el frontend (`docker compose up -d --build frontend`),
 * porque nginx sirve el bundle compilado. El host WSL carece de las librerías
 * del navegador, así que se usa la imagen de Playwright:
 *
 *   docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
 *     -v "$PWD":/work -w /work/frontend mcr.microsoft.com/playwright:v1.63.0-jammy \
 *     npx playwright test --config e2e/playwright.config.ts \
 *     qa-date-picker-panel-alineacion --workers=1
 */
test.describe('Bug date-picker panel: iconos alineados en la fila de fechas del evento', () => {
  test.setTimeout(120_000);

  const centroVertical = (caja: { y: number; height: number }): number => caja.y + caja.height / 2;

  /** Botón con icono que abre el selector nativo del input indicado. */
  const botonDelCampo = async (page: Page, dialogo: Locator, campo: Locator): Promise<Locator> => {
    const id = await campo.evaluate((el) => (el as HTMLInputElement).id);
    const boton = dialogo.locator(`button[aria-controls="${id}"]`);
    await expect(boton, `botón con icono del campo #${id}`).toBeVisible();
    return boton;
  };

  test('fecha, hora de inicio y hora de fin comparten el centro vertical de sus iconos', async ({
    browser,
  }) => {
    const context = await createAdminContext(browser);
    try {
      const page = await context.newPage();
      await page.goto('/panel');

      await page
        .getByRole('navigation', { name: 'Navegación del panel' })
        .getByRole('link', { name: 'Gestión de eventos' })
        .click();
      await expect(page.getByRole('heading', { name: 'Gestión de eventos' })).toBeVisible();
      await page.getByRole('button', { name: 'Agregar evento' }).click();

      const dialogo = page.getByRole('dialog');
      await expect(dialogo).toBeVisible();

      // Primera fila del editor de fechas: tres campos en la misma línea.
      const fecha = dialogo.getByLabel(/^Fecha 1/);
      const inicio = dialogo.getByLabel(/^Hora de inicio/);
      const fin = dialogo.getByLabel(/^Hora de fin/);
      await expect(fecha).toBeVisible();
      await expect(inicio).toBeVisible();
      await expect(fin).toBeVisible();

      const botones = [
        await botonDelCampo(page, dialogo, fecha),
        await botonDelCampo(page, dialogo, inicio),
        await botonDelCampo(page, dialogo, fin),
      ];

      const cajas = [];
      for (const boton of botones) {
        const caja = await boton.boundingBox();
        expect(caja, 'el botón con icono tiene caja').not.toBeNull();
        cajas.push(caja!);
      }

      const centros = cajas.map(centroVertical);
      const referencia = centros[0];
      const desfases = centros.map((centro) => centro - referencia);
      const mensaje =
        'centros verticales de los iconos (fecha, inicio, fin) relativos al primero: ' +
        `${desfases.map((valor) => valor.toFixed(1)).join(' / ')} px`;

      // Tolerancia ±1 px: los tres iconos comparten altura aunque las
      // etiquetas ocupen una o dos líneas (crecen hacia arriba, no hacia
      // abajo), y el botón «Quitar» se alinea con los inputs.
      for (const desfase of desfases) {
        expect(Math.abs(desfase), mensaje).toBeLessThanOrEqual(1);
      }
    } finally {
      await context.close();
    }
  });
});
