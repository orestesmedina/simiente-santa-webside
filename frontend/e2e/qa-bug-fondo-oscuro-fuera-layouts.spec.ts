import { expect, test, type Locator } from '@playwright/test';

/**
 * QA · Regresión del bug «fondo negro fuera de los layouts» (⟲ QA, no modifica
 * código de producción):
 *
 * Con el tema oscuro del dispositivo, el script anti-FOUC de `index.html`
 * (`frontend/index.html:32`) fija `documentElement.style.colorScheme = 'dark'`,
 * por lo que el **lienzo del navegador** se pinta de negro. Las pantallas que
 * viven **fuera** de `PublicLayout`/`PanelLayout` y no declaran fondo propio
 * (`/login`, `/cambiar-contrasena`, `/sin-permiso` y el 404) dejan ver ese lienzo:
 * fondo negro con texto oscuro → página ilegible. `PanelLayout` sí fuerza
 * `bg-white`, de ahí que el resto del panel no padezca el bug.
 *
 * Esta prueba exige que, con `prefers-color-scheme: dark` (y sin elección
 * guardada en `ss.theme`, tal como llega un visitante con dispositivo oscuro),
 * el contenedor principal de cada pantalla afectada pinte **fondo claro no
 * transparente** (`rgb(255, 255, 255)`, resolviendo la cadena de ancestros para
 * no atar el arreglo a un elemento concreto) y que el texto tenga color oscuro
 * (legible sobre claro). Tras el arreglo, debe seguir pasando.
 *
 * Ejecutar contra el stack levantado (`make up`); el host WSL carece de
 * `libnss3`, así que se usa la imagen de Playwright:
 *
 *   docker run --rm --network host --user "$(id -u):$(id -g)" -e HOME=/tmp \
 *     -v "$PWD/..":/work -w /work/frontend mcr.microsoft.com/playwright:v1.63.0-jammy \
 *     npx playwright test --config e2e/playwright.config.ts qa-bug-fondo-oscuro --workers=1
 */

const CLARO = 'rgb(255, 255, 255)';

/** Primer fondo no transparente de la cadena `elemento → <html>` (el pintado real). */
async function fondoPintado(locator: Locator): Promise<string> {
  return locator.evaluate((el) => {
    let node: Element | null = el;
    while (node) {
      const bg = getComputedStyle(node).backgroundColor;
      if (bg && !/^rgba\(\s*\d+,\s*\d+,\s*\d+,\s*0\s*\)$/.test(bg)) return bg;
      node = node.parentElement;
    }
    return 'rgba(0, 0, 0, 0)';
  });
}

/** Luminancia relativa (0=negro, 1=blanco) de un color `rgb(...)`. */
function luminancia(rgb: string): number {
  const match = /rgba?\(\s*(\d+),\s*(\d+),\s*(\d+)/.exec(rgb);
  if (!match) return NaN;
  const [, r, g, b] = match.map(Number) as unknown as number[];
  return (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255;
}

test.describe('Bug: pantallas fuera de los layouts con fondo negro en tema oscuro', () => {
  test.setTimeout(120_000);

  for (const ruta of ['/login', '/sin-permiso', '/esta-ruta-no-existe-404']) {
    test(`tema oscuro del dispositivo → «${ruta}» pinta fondo claro y texto oscuro`, async ({
      browser,
    }) => {
      // Dispositivo en oscuro, sin elección guardada: el anti-FOUC de
      // `index.html:32` aplicará `colorScheme: 'dark'` → lienzo negro.
      const context = await browser.newContext({ colorScheme: 'dark' });
      try {
        const page = await context.newPage();
        await page.goto(ruta);

        // El servidor responde y React pintó la pantalla (no un error de red).
        expect(await page.evaluate(() => document.documentElement.dataset.theme)).toBe('dark');
        expect(await page.evaluate(() => document.documentElement.style.colorScheme)).toBe('dark');

        let pantalla: string;
        let titulo: string;
        if (ruta === '/login') {
          pantalla = 'main';
          titulo = 'Entrar al panel';
        } else if (ruta === '/sin-permiso') {
          pantalla = 'section';
          titulo = 'No tienes acceso a esta sección';
        } else {
          pantalla = 'main';
          titulo = 'Página no encontrada';
        }

        const contenedor = page.locator(pantalla).first();
        await expect(contenedor.getByRole('heading', { name: titulo })).toBeVisible();

        // Fondo claro y NO transparente en la cadena pintada (el lienzo negro
        // del navegador no debe verse por ninguna parte).
        const fondo = await fondoPintado(contenedor);
        expect(fondo, `fondo pintado en ${ruta}`).toBe(CLARO);

        // Texto legible: el título usa color oscuro (legible sobre claro).
        const colorTitulo = await contenedor
          .getByRole('heading', { name: titulo })
          .evaluate((el) => getComputedStyle(el).color);
        expect(luminancia(colorTitulo), `color del título en ${ruta}`).toBeLessThan(0.5);
      } finally {
        await context.close();
      }
    });
  }
});
