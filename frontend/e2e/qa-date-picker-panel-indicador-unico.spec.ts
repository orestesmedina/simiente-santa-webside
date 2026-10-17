import { expect, test, type Page } from '@playwright/test';
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
 * se captura PNG del campo, se retira la clase **solo en la prueba** (el DOM
 * se restaura) y se re-captura: si ambas imágenes son idénticas, es que la
 * clase ya no oculta nada (regresión).
 */
test.describe('Bug date-picker panel: un solo icono por campo (indicador nativo oculto)', () => {
  test.setTimeout(120_000);

  const CLASE_INDICADOR = '[&::-webkit-calendar-picker-indicator]:hidden';

  /** ¿Hace efecto la clase que oculta el indicador en este campo? */
  const claseOcultaEfectivamente = async (page: Page, selectorCampo: string): Promise<boolean> => {
    const campo = page.locator(selectorCampo);
    // Estabilidad de la captura con la clase (dos tomas iguales).
    const fotoA1 = await campo.screenshot();
    const fotoA2 = await campo.screenshot();
    expect(fotoA1.equals(fotoA2), 'captura estable del campo').toBe(true);
    // Sin la clase: cambia el render (aparece el indicador nativo).
    await campo.evaluate((el, clase) => {
      (el as HTMLElement).classList.remove(clase);
    }, CLASE_INDICADOR);
    const fotoB = await campo.screenshot();
    // La clase se repone para no dejar el DOM tocado por la prueba.
    await campo.evaluate((el, clase) => {
      (el as HTMLElement).classList.add(clase);
    }, CLASE_INDICADOR);
    return !fotoA1.equals(fotoB);
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
