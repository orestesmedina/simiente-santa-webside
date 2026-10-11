import type { Page } from '@playwright/test';

/**
 * Utilidades de UI para los e2e de F10 (no es un `*.spec.ts`).
 */

/**
 * Espera a que el desplazamiento se detenga. Necesario tras un salto de ancla
 * con `scroll-behavior: smooth`: medir la posición del destino antes de que la
 * animación termine daría un valor transitorio.
 */
export async function waitForScrollEnd(page: Page): Promise<void> {
  await page.waitForFunction(
    () =>
      new Promise<boolean>((resolve) => {
        const started = performance.now();
        let last = window.scrollY;
        let stable = 0;
        const check = () => {
          if (Math.abs(window.scrollY - last) < 1) {
            stable += 1;
          } else {
            stable = 0;
            last = window.scrollY;
          }
          // Estable durante ~300 ms y con al menos ~500 ms transcurridos: así no
          // se confunde el instante previo al arranque del desplazamiento suave.
          if (stable >= 3 && performance.now() - started > 500) {
            resolve(true);
          } else {
            setTimeout(check, 100);
          }
        };
        setTimeout(check, 100);
      }),
    undefined,
    { timeout: 8000 },
  );
}
