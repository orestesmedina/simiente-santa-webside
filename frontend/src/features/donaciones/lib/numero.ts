import type { MetodoKind } from '../../../api/ofrendas';

/**
 * Presentación de los números en el **panel** (ux.md §4.3, ⟲ D-9). Misma
 * responsabilidad que `features/publico/ofrendas/format.ts` pero con los dos
 * tipos: el IBAN se agrupa en bloques de 4 para que el equipo lo lea y lo
 * transcriba sin error, y el SINPE se muestra tal como lo publica el servidor
 * (`+506` + 8 dígitos).
 *
 * Es **solo presentación**: el valor exacto y copiable es siempre el `number`
 * normalizado que devuelve la API (RG7-7: quitar los separadores de la
 * presentación devuelve el número). Aquí no se normaliza ni se reconstruye nada.
 */

/**
 * IBAN agrupado en bloques de 4 separados por espacios (`CR05 0152 0200 1234
 * 5678 90`). No recorta ni transforma caracteres: quitar los espacios devuelve
 * exactamente el valor guardado.
 */
export function agruparIban(number: string): string {
  return number.replace(/(.{4})/g, '$1 ').trim();
}

/** Presentación de lectura según el tipo: IBAN agrupado, SINPE normalizado. */
export function presentarNumero(kind: MetodoKind, number: string): string {
  return kind === 'iban' ? agruparIban(number) : number;
}
