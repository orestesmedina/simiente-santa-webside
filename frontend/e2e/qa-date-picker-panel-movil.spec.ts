import { expect, test, type Locator, type Page } from '@playwright/test';
import { createAdminContext } from './publicSeed';

/**
 * QA (2026-10-17, 3.ª iteración del bug del date picker del panel): criterios
 * que `qa-date-picker-panel-alineacion` no mide y que la fila de fechas debe
 * cumplir a raíz del paso a rejilla de **tres pistas con `subgrid`**:
 *
 *   1. **Sin hueco reservado visible sin error:** la 3.ª pista (mensaje) es
 *      `auto`; sin error no debe aumentar el alto del campo (sin mensaje el
 *      input cierra el campo). Con error, el mensaje ocupa **su** pista bajo el
 *      control (`role="alert"` + `aria-invalid`, ya exigido en la prueba de
 *      alineación).
 *   2. **Móvil (<sm): la fila apila sin desbordes y «Quitar» queda accesible:**
 *      en móvil el `subgrid` (prefijado `sm:`) no aplica: la fila es una
 *      rejilla de una columna con `gap-y-3` (12 px entre campos); fecha, horas
 *      y «Quitar» apilan en vertical, el documento no gana scroll horizontal
 *      (con y sin error de validación) y el botón «Quitar la fecha N» es
 *      visible, con objetivo táctil razonable y funciona al pulsarlo.
 *
 * Ejecutar contra el stack levantado (`make up`) y el frontend reconstruido
 * (`docker compose up -d --build frontend`), en la imagen de Playwright:
 *
 *   docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
 *     -v "$PWD":/work -w /work/frontend mcr.microsoft.com/playwright:v1.63.0-jammy \
 *     npx playwright test --config e2e/playwright.config.ts \
 *     qa-date-picker-panel-movil --workers=1
 */
test.describe('Bug date-picker panel: fila de fechas sin hueco reservado y móvil apilado', () => {
  // 180 s como las suites con seed (navegacion-movil): contra arranque frío del
  // stack no basta el timeout por defecto; las aserciones no cambian.
  test.setTimeout(180_000);

  /** Abre «Gestión de eventos» → «Agregar evento» y devuelve el diálogo. */
  const abrirAltaDeEvento = async (page: Page): Promise<Locator> => {
    await page.goto('/panel');
    // La app es CSR: `goto` resuelve con la página vacía; esperar a que React
    // monte `/panel` (heading del inicio) antes de leer la navegación, o el
    // `isVisible()` resuelve antes de tiempo y decide mal (⟲ diagnóstico).
    const inicio = page.getByRole('heading', { name: 'Inicio', exact: true });
    await expect(inicio).toBeVisible();
    // En móvil (<md) la navegación del panel está plegada: se abre el menú
    // antes de pulsar el enlace (mismo patrón `disclosure` del panel).
    const enlace = page
      .getByRole('navigation', { name: 'Navegación del panel' })
      .getByRole('link', { name: 'Gestión de eventos' });
    if (!(await enlace.isVisible())) {
      await page.getByRole('button', { name: 'Menú' }).click();
    }
    await expect(enlace).toBeVisible();
    await enlace.click();
    await expect(page.getByRole('heading', { name: 'Gestión de eventos' })).toBeVisible();
    await page.getByRole('button', { name: 'Agregar evento' }).click();
    const dialogo = page.getByRole('dialog');
    await expect(dialogo).toBeVisible();
    return dialogo;
  };

  /** Raíz del campo (div de rejilla/flex que contiene label + control). */
  const raizDelCampo = (campo: Locator): Locator => campo.locator('xpath=ancestor::div[2]');

  test('sin error el mensaje no reserva hueco: el input cierra el campo; con error el mensaje ocupa su pista', async ({
    browser,
  }) => {
    const context = await createAdminContext(browser);
    try {
      const page = await context.newPage();
      const dialogo = await abrirAltaDeEvento(page);

      // El hueco residual admitido es el `sm:gap-y-1` de 4 px que la subgrid
      // hereda **de forma uniforme** entre las pistas de la rejilla de la fila
      // (compartida por los tres campos: es el ritmo vertical del campo, no un
      // hueco del mensaje, cuya pista `auto` mide 0 sin error).
      const HUECO_SIN_ERROR = 6; // 4 px de `sm:gap-y-1` + 2 px de tolerancia.

      // ── Estado sin error: el hueco del mensaje no existe ──
      const campos = [
        dialogo.getByLabel(/^Fecha 1/),
        dialogo.getByLabel(/^Hora de inicio/),
        dialogo.getByLabel(/^Hora de fin/),
      ];
      for (const campo of campos) {
        await expect(campo).toBeVisible();
        const input = await campo.boundingBox();
        const raiz = await raizDelCampo(campo).boundingBox();
        expect(input, 'el input tiene caja').not.toBeNull();
        expect(raiz, 'la raíz del campo tiene caja').not.toBeNull();
        const hueco = raiz!.y + raiz!.height - (input!.y + input!.height);
        expect(
          hueco,
          `sin error, bajo el campo ${await campo.getAttribute('id')} solo queda el hueco compartido del ritmo`,
        ).toBeLessThanOrEqual(HUECO_SIN_ERROR);
      }

      // ── Estado con error: el mensaje vive bajo su control (pista propia) ──
      await dialogo.getByLabel(/^Hora de fin/).fill('10:00'); // fin sin inicio
      await dialogo.getByRole('button', { name: 'Agregar evento' }).click();

      const mensaje = dialogo
        .locator('p[id$="-error"][role="alert"]')
        .filter({ hasText: 'Escribe la hora de inicio para poder poner una hora de fin.' });
      await expect(mensaje).toBeVisible();

      const fin = dialogo.getByLabel(/^Hora de fin/);
      await expect(fin).toHaveAttribute('aria-invalid', 'true');
      const inputFin = await fin.boundingBox();
      const raizFin = await raizDelCampo(fin).boundingBox();
      const cajaMensaje = await mensaje.boundingBox();
      const huecoConError = raizFin!.y + raizFin!.height - (inputFin!.y + inputFin!.height);
      // El mensaje entra entre el input y el borde inferior de la raíz: crece
      // por encima del hueco compartido (4 px) y su pista contiene la caja.
      expect(huecoConError, 'con error, la raíz crece para alojar el mensaje').toBeGreaterThan(
        HUECO_SIN_ERROR,
      );
      expect(cajaMensaje, 'el mensaje tiene caja propia').not.toBeNull();
      expect(
        huecoConError - cajaMensaje!.height,
        'el mensaje encaja en la pista (solo queda el gap compartido)',
      ).toBeLessThanOrEqual(HUECO_SIN_ERROR);
    } finally {
      await context.close();
    }
  });

  test('en móvil (<sm) la fila apila sin scroll horizontal y «Quitar» es accesible y funciona', async ({
    browser,
  }) => {
    const context = await createAdminContext(browser);
    try {
      const page = await context.newPage();
      // 375×812: iPhone clásico, bien por debajo del corte `sm` (640 px).
      await page.setViewportSize({ width: 375, height: 812 });
      // Con el menú del panel plegado en móvil, el ayudante abre el disclosure
      // antes de navegar.
      const dialogo = await abrirAltaDeEvento(page);

      const sinDesborde = async (etiqueta: string) => {
        const exceso = await page.evaluate(
          () => document.documentElement.scrollWidth - window.innerWidth,
        );
        expect(exceso, `${etiqueta}: sin scroll horizontal a 375 px`).toBeLessThanOrEqual(1);
      };

      /** Ritmo vertical entre raíces de campos consecutivos de la fila. */
      const huecoEntreCampos = async (etiqueta: string) => {
        // ⟲ revisión (I1): al apilar en una columna, los campos conservan la
        // separación de 12 px (`gap-y-3`); con `gap-y-1` quedaban a 4 px.
        const pares: Array<[string, RegExp, RegExp]> = [
          ['Fecha 1 → Hora de inicio', /^Fecha 1/, /^Hora de inicio/],
          ['Hora de inicio → Hora de fin', /^Hora de inicio/, /^Hora de fin/],
        ];
        for (const [par, arriba, abajo] of pares) {
          const a = await raizDelCampo(dialogo.getByLabel(arriba)).boundingBox();
          const b = await raizDelCampo(dialogo.getByLabel(abajo)).boundingBox();
          expect(a, `raíz superior del par ${par} tiene caja`).not.toBeNull();
          expect(b, `raíz inferior del par ${par} tiene caja`).not.toBeNull();
          const hueco = b!.y - (a!.y + a!.height);
          expect(
            hueco,
            `${etiqueta}: ${par} conserva el ritmo móvil de 12 px`,
          ).toBeGreaterThanOrEqual(11);
          expect(hueco, `${etiqueta}: ${par} no supera 13 px`).toBeLessThanOrEqual(13);
        }
      };

      // Sin error: los tres campos apilan en vertical (cada control encima del
      // siguiente) y ninguno sale del ancho de la pantalla.
      const nombres = [/^Fecha 1/, /^Hora de inicio/, /^Hora de fin/];
      let yPrevio = -Infinity;
      for (const nombre of nombres) {
        const campo = dialogo.getByLabel(nombre);
        await expect(campo).toBeVisible();
        const caja = await campo.boundingBox();
        expect(caja).not.toBeNull();
        expect(caja!.x, `el campo ${nombre} no sale por la derecha`).toBeGreaterThanOrEqual(0);
        expect(
          caja!.x + caja!.width,
          `el campo ${nombre} no excede el viewport`,
        ).toBeLessThanOrEqual(376);
        expect(caja!.y, 'los campos apilan hacia abajo').toBeGreaterThan(yPrevio);
        yPrevio = caja!.y;
      }
      await sinDesborde('sin error');
      await huecoEntreCampos('sin error');

      // «Quitar» de la fila 1: visible, objetivo táctil razonable y accesible
      // por teclado (foco con teclado + activación). Con una sola fila el
      // botón queda tras los campos; con dos se comprueba que quita la fila.
      const quitar1 = dialogo.getByRole('button', { name: 'Quitar la fecha 1' });
      await expect(quitar1).toBeVisible();
      const cajaQuitar = await quitar1.boundingBox();
      expect(cajaQuitar).not.toBeNull();
      expect(
        Math.min(cajaQuitar!.width, cajaQuitar!.height),
        'objetivo táctil de «Quitar» ≥ 24 px (WCAG 2.5.8)',
      ).toBeGreaterThanOrEqual(24);

      // Focusable: el foco llega a «Quitar» por teclado.
      await quitar1.focus();
      await expect(quitar1).toBeFocused();

      // Funciona: con una segunda fila, «Quitar la fecha 2» deja solo la 1.
      await dialogo.getByRole('button', { name: 'Agregar fecha' }).click();
      await expect(dialogo.getByLabel(/^Fecha 2/)).toBeVisible();
      await dialogo.getByRole('button', { name: 'Quitar la fecha 2' }).click();
      await expect(dialogo.getByLabel(/^Fecha 2/)).toHaveCount(0);
      await sinDesborde('tras quitar');

      // Con error de validación (mismo estado que mide la alineación) la
      // columna tampoco gana scroll horizontal.
      await dialogo.getByLabel(/^Hora de fin/).fill('10:00');
      await dialogo.getByRole('button', { name: 'Agregar evento' }).click();
      await expect(
        dialogo.locator('p[id$="-error"][role="alert"]').filter({
          hasText: 'Escribe la hora de inicio para poder poner una hora de fin.',
        }),
      ).toBeVisible();
      await sinDesborde('con error de validación');
      await huecoEntreCampos('con error de validación');
    } finally {
      await context.close();
    }
  });
});
