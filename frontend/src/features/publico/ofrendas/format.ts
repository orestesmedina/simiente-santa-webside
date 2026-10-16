import type { MetodoKind } from '../../../api/ofrendas';

/**
 * Presentación de los números de donación en el sitio público (FR-003/FR-015,
 * SC-009). Es **solo presentación**: el valor exacto y copiable es siempre el
 * `number` normalizado que publica la API y nunca se reconstruye desde aquí
 * (RG7-7: `quitar los separadores devuelve el número`).
 */

/**
 * IBAN agrupado en bloques de 4 separados por espacios para lectura y
 * transcripción (`CR05 0152 0200 1234 5678 90`). No recorta ni transforma
 * caracteres: quitar los espacios devuelve exactamente el valor guardado.
 */
export function agruparIban(number: string): string {
  return number.replace(/(.{4})/g, '$1 ').trim();
}

/**
 * Presentación de lectura del número según su tipo: el IBAN se agrupa en bloques
 * de 4 y el SINPE se muestra en su **formato normalizado** (`+506` + 8 dígitos),
 * tal como lo publica el servidor (P7-4/alineación de `tasks.md` §Alineación 9:
 * la agrupación del SINPE de `ux.md` D-4 sigue pendiente de confirmación humana).
 */
export function presentarNumero(kind: MetodoKind, number: string): string {
  return kind === 'iban' ? agruparIban(number) : number;
}
