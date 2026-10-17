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
 * La prueba mide los `boundingBox()` reales de los tres botones con icono y del
 * botón «Quitar» de la primera fila, y exige que sus centros verticales
 * coincidan con tolerancia de ±1 px. Falla antes del arreglo (las etiquetas de
 * dos líneas bajan su icono ~1 renglón) y pasa después.
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
  const botonDelCampo = async (dialogo: Locator, campo: Locator): Promise<Locator> => {
    const id = await campo.evaluate((el) => (el as HTMLInputElement).id);
    const boton = dialogo.locator(`button[aria-controls="${id}"]`);
    await expect(boton, `botón con icono del campo #${id}`).toBeVisible();
    return boton;
  };

  /** Abre «Gestión de eventos» → «Agregar evento» y devuelve el diálogo. */
  const abrirAltaDeEvento = async (page: Page): Promise<Locator> => {
    await page.goto('/panel');
    await page
      .getByRole('navigation', { name: 'Navegación del panel' })
      .getByRole('link', { name: 'Gestión de eventos' })
      .click();
    await expect(page.getByRole('heading', { name: 'Gestión de eventos' })).toBeVisible();
    await page.getByRole('button', { name: 'Agregar evento' }).click();

    const dialogo = page.getByRole('dialog');
    await expect(dialogo).toBeVisible();
    return dialogo;
  };

  /** Centros verticales de los controles de la primera fila (iconos y «Quitar»). */
  const centrosDeLaFila = async (
    dialogo: Locator,
  ): Promise<{ desfases: number[]; mensaje: string }> => {
    const fecha = dialogo.getByLabel(/^Fecha 1/);
    const inicio = dialogo.getByLabel(/^Hora de inicio/);
    const fin = dialogo.getByLabel(/^Hora de fin/);
    await expect(fecha).toBeVisible();
    await expect(inicio).toBeVisible();
    await expect(fin).toBeVisible();

    // El botón «Quitar» también vive en la pista del control: debe compartir
    // centro vertical con los inputs y sus iconos, no solo estar visible.
    const quitar = dialogo.getByRole('button', { name: 'Quitar la fecha 1' });
    await expect(quitar).toBeVisible();

    const botones = [
      await botonDelCampo(dialogo, fecha),
      await botonDelCampo(dialogo, inicio),
      await botonDelCampo(dialogo, fin),
      quitar,
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
    return {
      desfases,
      mensaje:
        'centros verticales de los controles (fecha, inicio, fin, quitar) relativos al primero: ' +
        `${desfases.map((valor) => valor.toFixed(1)).join(' / ')} px`,
    };
  };

  /** Tolerancia ±1 px entre los centros de los controles de la fila. */
  const exigirAlineados = (desfases: number[], mensaje: string): void => {
    for (const desfase of desfases) {
      expect(Math.abs(desfase), mensaje).toBeLessThanOrEqual(1);
    }
  };

  test('los controles de la fila comparten centro vertical (iconos y «Quitar»)', async ({
    browser,
  }) => {
    const context = await createAdminContext(browser);
    try {
      const page = await context.newPage();
      const dialogo = await abrirAltaDeEvento(page);

      // Tolerancia ±1 px: los tres iconos y «Quitar» comparten altura aunque
      // las etiquetas ocupen una o dos líneas (crecen hacia arriba, no hacia
      // abajo).
      const { desfases, mensaje } = await centrosDeLaFila(dialogo);
      exigirAlineados(desfases, mensaje);
    } finally {
      await context.close();
    }
  });

  /**
   * ⟲ (2026-10-17, 3.ª iteración, corrección): `sm:items-end` alinea los
   * iconos cuando las celdas miden lo mismo, pero un **mensaje de validación**
   * bajo un campo hace crecer solo esa celda: la fila toma el alto de la más
   * alta y las demás —alineadas por abajo— bajan su input, así que el icono
   * del campo con error sube respecto a los otros (medido 0 / +44 / −40 px en
   * el diálogo real). El error de la fila lo introduce nuestro propio formulario
   * («Hora de fin» sin «Hora de inicio», validación real de FR-015), así que el
   * estado debe quedar estable igual que el normal.
   */
  test('los iconos siguen alineados cuando la fila muestra un error de validación', async ({
    browser,
  }) => {
    const context = await createAdminContext(browser);
    try {
      const page = await context.newPage();
      const dialogo = await abrirAltaDeEvento(page);

      // «Hora de fin» sin «Hora de inicio» es inválido (END_WITHOUT_START).
      const fin = dialogo.getByLabel(/^Hora de fin/);
      await fin.fill('10:00');

      // Enviar: la validación del cliente pinta el error bajo «Hora de fin»
      // (y el de fecha obligatoria bajo «Fecha 1»). Se usa el botón de envío
      // del propio diálogo, no el que abrió el formulario.
      await dialogo.getByRole('button', { name: 'Agregar evento' }).click();

      // El mensaje sigue mostrándose y anunciándose a lectores de pantalla (el
      // `role="alert"` del propio campo, no el resumen general del formulario).
      const errorFin = dialogo
        .locator('p[id$="-error"][role="alert"]')
        .filter({ hasText: 'Escribe la hora de inicio para poder poner una hora de fin.' });
      await expect(errorFin).toBeVisible();
      await expect(fin).toHaveAttribute('aria-invalid', 'true');

      // La medición no debe depender de fuentes ni de un repintado pendiente.
      await page.evaluate(async () => {
        await document.fonts.ready;
        await new Promise((resolve) => requestAnimationFrame(() => resolve(null)));
      });

      const { desfases, mensaje } = await centrosDeLaFila(dialogo);
      exigirAlineados(desfases, `con error de validación · ${mensaje}`);

      // El error no desaparece al medir: la fila queda en el estado con error.
      await expect(errorFin).toBeVisible();
    } finally {
      await context.close();
    }
  });
});
