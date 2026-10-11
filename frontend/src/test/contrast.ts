import { existsSync, readFileSync } from 'node:fs';
import { join } from 'node:path';

/**
 * Utilidades de contraste para las pruebas del sistema visual (F10).
 *
 * La hoja `src/index.css` se localiza de forma robusta (funciona tanto si vitest
 * se lanza desde `frontend/` como desde la raíz del repo) y los pares de color
 * se comprueban con la fórmula WCAG 2.1, de modo que un cambio de token que
 * rompa AA haga fallar la prueba en vez de pasar inadvertido (los fondos con
 * alfa no los resuelve la auditoría e2e congelada).
 */
export function readSystemCss(): string {
  const candidates = [
    join(process.cwd(), 'src', 'index.css'),
    join(process.cwd(), 'frontend', 'src', 'index.css'),
  ];
  const found = candidates.find((path) => existsSync(path));
  if (!found) {
    throw new Error(`No se encontró src/index.css (cwd=${process.cwd()})`);
  }
  return readFileSync(found, 'utf8');
}

/** Valores de un token CSS, en orden de aparición (p. ej. claro y luego oscuro). */
export function tokenValues(css: string, name: string): string[] {
  return [...css.matchAll(new RegExp(`--${name}:\\s*([^;]+);`, 'g'))].map((match) =>
    match[1].trim(),
  );
}

function luminance(hex: string): number {
  const value = hex.replace('#', '');
  const [r, g, b] = [0, 2, 4].map((index) => parseInt(value.slice(index, index + 2), 16) / 255);
  const lin = (channel: number) =>
    channel <= 0.03928 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4;
  return 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b);
}

/** Razón de contraste WCAG 2.1 entre dos colores hex sólidos. */
export function contrast(a: string, b: string): number {
  const [la, lb] = [luminance(a), luminance(b)];
  return (Math.max(la, lb) + 0.05) / (Math.min(la, lb) + 0.05);
}
