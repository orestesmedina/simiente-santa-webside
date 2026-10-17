import { expect, test, type Locator, type Page } from '@playwright/test';
import { createAdminContext } from './publicSeed';

/**
 * QA ⟲ (2026-10-17): criterio de la 2.ª iteración del bug del date picker del
 * panel que la spec `qa-bug-color-scheme-panel` no cubre: **un solo icono por
 * campo** — al añadir el botón propio, el indicador nativo
 * (`::-webkit-calendar-picker-indicator`) debe quedar oculto para que el
 * selector no se muestre «dos veces».
 *
 * Cubre los consumidores heredados de `Field`: `DatesEditor` (fórm. «Agregar
 * evento») y `DateRangeFilter` (filtros de Auditoría). `ActivityForm` usa el
 * mismo componente y se ejercita en `agenda-panel`.
 *
 * Nota de verificación (2026-10-17, QA): el detalle del indicador es un
 * pseudo-elemento interno del motor, y en la imagen de Playwright resultan
 * insuficientes de dos maneras:
 *   1. `getComputedStyle(el, '::-webkit-calendar-picker-indicator')` ni
 *      siquiera refleja estilos de autor inyectados con `!important`.
 *   2. CDP `CSS.getMatchedStylesForNode` no lista esas reglas para el pseudo.
 * Por eso la aserción es **visual y comparativa**, aunque el detector del
 * motor (lo que verifica es que, con la clase en el DOM, el indicador desaparece):
 * se captura el **interior** del campo —sin el borde, donde se pinta el
 * indicador—, se retira la clase **solo en la prueba** (el DOM se restaura) y
 * se re-captura: si ambas imágenes son idénticas, es que la clase ya no oculta
 * nada (regresión).
 *
 * ⟲ (2026-10-17, 3.ª iteración): la captura del campo completo resultó
 * inestable en ejecuciones encadenadas: el antialiasing de la esquina
 * redondeada alterna 1 nivel de color entre repintados del compositor (misma
 * `boundingBox`, fuentes `loaded`). Por eso se compara solo el interior, que
 * conserva la señal del indicador y descarta ese ruido.
 */
test.describe('Bug date-picker panel: un solo icono por campo (indicador nativo oculto)', () => {
  test.setTimeout(120_000);

  const CLASE_INDICADOR = '[&::-webkit-calendar-picker-indicator]:hidden';

  /** Recorte interior del campo (sin borde) donde se pinta el indicador. */
  const RECORTE_BORDE = 6;

  /**
   * Captura del **interior** del campo, donde vive el indicador nativo.
   *
   * Se excluyen 6 px de borde por cada lado a propósito: el antialiasing del
   * borde redondeado alterna 1 nivel de color entre repintados del compositor
   * (se comprobó que la `boundingBox` y las fuentes no cambian), y esa
   * diferencia ajena al indicador convertía la comparación en inestable. El
   * indicador nativo se pinta dentro del campo (a la derecha; con el icono
   * propio, en el hueco que deja `pr-11`), así que el recorte conserva la señal
   * que se quiere medir y descarta el ruido del borde.
   */
  const capturaIndicador = async (page: Page, campo: Locator): Promise<Buffer> => {
    // `page.screenshot({ clip })` recorta sobre el viewport (no desplaza solo):
    // el campo tiene que estar a la vista (el diálogo del evento desplaza).
    await campo.scrollIntoViewIfNeeded();
    const caja = await campo.boundingBox();
    if (!caja) {
      throw new Error('el campo no tiene caja para la captura');
    }
    return page.screenshot({
      clip: {
        x: caja.x + RECORTE_BORDE,
        y: caja.y + RECORTE_BORDE,
        width: caja.width - RECORTE_BORDE * 2,
        height: caja.height - RECORTE_BORDE * 2,
      },
    });
  };

  /** ¿Hace efecto la clase que oculta el indicador en este campo? */
  const claseOcultaEfectivamente = async (page: Page, selectorCampo: string): Promise<boolean> => {
    const campo = page.locator(selectorCampo);
    // Interior del campo con la clase (una sola toma: sin borde es estable).
    const fotoA = await capturaIndicador(page, campo);
    // Sin la clase: cambia el render (aparece el indicador nativo).
    await campo.evaluate((el, clase) => {
      (el as HTMLElement).classList.remove(clase);
    }, CLASE_INDICADOR);
    const fotoB = await capturaIndicador(page, campo);
    // La clase se repone para no dejar el DOM tocado por la prueba.
    await campo.evaluate((el, clase) => {
      (el as HTMLElement).classList.add(clase);
    }, CLASE_INDICADOR);
    return !fotoA.equals(fotoB);
  };

  test('un botón por input y el indicador nativo oculto en «Agregar evento» y en el filtro de auditoría', async ({
    browser,
  }) => {
    const context = await createAdminContext(browser);
    try {
      const page = await context.newPage();
      await page.goto('/panel');

      // ── DatesEditor: formulario «Agregar evento» ─────────────────────────
      await page
        .getByRole('navigation', { name: 'Navegación del panel' })
        .getByRole('link', { name: 'Gestión de eventos' })
        .click();
      await page.getByRole('button', { name: 'Agregar evento' }).click();
      const dialog = page.getByRole('dialog');
      await expect(dialog).toBeVisible();

      const fechas = dialog.locator('input[type="date"]');
      const horas = dialog.locator('input[type="time"]');
      expect(await fechas.count()).toBeGreaterThan(0);
      expect(await horas.count()).toBeGreaterThan(0);

      // Un único botón por input, apuntando al input correcto (aria-controls → id).
      const botones = dialog.getByRole('button', { name: /Abrir (calendario|selector de hora)/ });
      expect(await botones.count()).toBe((await fechas.count()) + (await horas.count()));
      for (const entrada of [...(await fechas.all()), ...(await horas.all())]) {
        const id = await entrada.evaluate((el) => (el as HTMLInputElement).id);
        await expect(dialog.locator(`button[aria-controls="${id}"]`)).toHaveCount(1);
      }

      // Núcleo del criterio: en cada campo, la clase oculta el indicador nativo.
      for (const entrada of [...(await fechas.all()), ...(await horas.all())]) {
        const id = await entrada.evaluate((el) => (el as HTMLInputElement).id);
        expect(
          await claseOcultaEfectivamente(page, `input#${id}`),
          `indicador nativo oculto (un solo icono) en #${id}`,
        ).toBe(true);
      }
      await page.keyboard.press('Escape');

      // ── DateRangeFilter: página de auditoría ────────────────────────────
      await page
        .getByRole('navigation', { name: 'Navegación del panel' })
        .getByRole('link', { name: 'Auditoría' })
        .click();
      await expect(page.getByRole('heading', { name: 'Auditoría' })).toBeVisible();

      // Antes de comparar, la fila de filtros debe dejar de recolocarse: las
      // fuentes (`font-display: swap` cambia el ancho del selector «Cuenta») y
      // los datos del historial (su tabla cambia el alto del documento y
      // aparece la barra de desplazamiento) llegan de forma asíncrona. Si la
      // fila se moviera entre la captura con clase y sin clase, la diferencia
      // de imágenes no sería atribuible al indicador.
      await page.evaluate(async () => {
        await document.fonts.ready;
      });
      await expect(page.getByText('Cargando registros…')).toHaveCount(0);
      await expect(page.getByText('Cargando cuentas…')).toHaveCount(0);

      const filtroFechas = page.locator('input[type="date"]');
      expect(await filtroFechas.count()).toBe(2);
      expect(await page.getByRole('button', { name: 'Abrir calendario' }).count()).toBe(2);
      for (const entrada of await filtroFechas.all()) {
        const id = await entrada.evaluate((el) => (el as HTMLInputElement).id);
        expect(
          await claseOcultaEfectivamente(page, `input#${id}`),
          `indicador nativo oculto (un solo icono) en #${id}`,
        ).toBe(true);
      }
    } finally {
      await context.close();
    }
  });
});
