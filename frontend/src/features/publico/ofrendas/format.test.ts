import { describe, expect, it } from 'vitest';
import { agruparIban, presentarNumero } from './format';

/**
 * Presentación de los números públicos (FR-003/FR-015/SC-009): la agrupación del
 * IBAN es **solo presentación** —quitar los separadores devuelve el número
 * exacto— y el SINPE se muestra tal como lo publica el servidor (`+506` + 8
 * dígitos; P7-4/alineación de `tasks.md`), sin embellecerlo.
 */
describe('format de números de ofrendas (presentación, SC-009)', () => {
  it('agrupa el IBAN en bloques de 4 sin alterar ni un carácter', () => {
    const iban = 'CR05015202001234567890';

    const agrupado = agruparIban(iban);

    expect(agrupado).toBe('CR05 0152 0200 1234 5678 90');
    // Invariante RG7-7: quitar los separadores devuelve el número normalizado.
    expect(agrupado.replace(/\s/g, '')).toBe(iban);
  });

  it('agrupa también los IBAN internacionales largos sin recortarlos', () => {
    const iban = 'ES9121000418450200051332';

    expect(agruparIban(iban)).toBe('ES91 2100 0418 4502 0005 1332');
  });

  it('el SINPE se presenta en su formato normalizado (sin agrupar)', () => {
    expect(presentarNumero('sinpe', '+50688888888')).toBe('+50688888888');
    expect(presentarNumero('iban', 'CR05015202001234567890')).toBe('CR05 0152 0200 1234 5678 90');
  });
});
