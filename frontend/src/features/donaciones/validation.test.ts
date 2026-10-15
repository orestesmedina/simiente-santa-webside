import { describe, expect, it } from 'vitest';
import {
  CURRENCY_OPTIONS,
  crossFieldErrors,
  donacionSchema,
  EMPTY_DONACION_FIELDS,
  esNumeroValido,
  issuesToErrors,
  normalizarIban,
  normalizarSinpe,
  quitarSeparadores,
} from './validation';

describe('normalización de IBAN (espejo de FR-015)', () => {
  it('acepta un IBAN costarricense y lo normaliza a mayúsculas sin separadores', () => {
    expect(normalizarIban('CR05 0152 0200 1234 5678 90')).toBe('CR05015202001234567890');
    expect(normalizarIban('cr05-0152-0200-1234-5678-90')).toBe('CR05015202001234567890');
  });

  it('acepta un IBAN internacional con su longitud exacta', () => {
    // Alemania: DE + 20 = 22 caracteres.
    expect(normalizarIban('DE89 3704 0044 0532 0130 00')).toBe('DE89370400440532013000');
  });

  it('rechaza longitud incorrecta y estructura inválida', () => {
    expect(normalizarIban('CR0501520200123456789')).toBeNull(); // 21, falta un dígito
    expect(normalizarIban('XX05015202001234567890')).toBeNull(); // país desconocido
    expect(normalizarIban('0005015202001234567890')).toBeNull(); // sin letras de país
    expect(normalizarIban('')).toBeNull();
  });

  it('quitarSeparadores respeta otros caracteres', () => {
    expect(quitarSeparadores('a b\tc\u00a0d-e–f—g')).toBe('abcdefg');
  });
});

describe('normalización de SINPE (espejo de FR-016)', () => {
  it('acepta el número con o sin +506 y con separadores', () => {
    expect(normalizarSinpe('88887777')).toBe('+50688887777');
    expect(normalizarSinpe('+506 8888 7777')).toBe('+50688887777');
    expect(normalizarSinpe('50688887777')).toBe('+50688887777');
    expect(normalizarSinpe('+506-8888-7777')).toBe('+50688887777');
  });

  it('rechaza longitudes, prefijos y primeros dígitos no válidos', () => {
    expect(normalizarSinpe('8888777')).toBeNull(); // 7 dígitos
    expect(normalizarSinpe('888877778')).toBeNull(); // 9 dígitos
    expect(normalizarSinpe('5068888777')).toBeNull(); // 506 + 7, no 11
    expect(normalizarSinpe('58887777')).toBeNull(); // primer dígito 5
    expect(normalizarSinpe('8888777a')).toBeNull(); // con letra
  });

  it('esNumeroValido elige la regla por tipo', () => {
    expect(esNumeroValido('iban', 'CR05015202001234567890')).toBe(true);
    expect(esNumeroValido('iban', '+50688887777')).toBe(false);
    expect(esNumeroValido('sinpe', '+50688887777')).toBe(true);
    expect(esNumeroValido('sinpe', 'CR05015202001234567890')).toBe(false);
  });
});

describe('donacionSchema (espejo de FR-014)', () => {
  const base = { ...EMPTY_DONACION_FIELDS };

  it('exige entidad y número', () => {
    const result = donacionSchema('iban').safeParse(base);
    expect(result.success).toBe(false);
    if (!result.success) {
      const errors = issuesToErrors(result.error);
      expect(errors.institution).toBe('Escribe el nombre del banco.');
      expect(errors.number).toBe(
        'Escribe el número de la cuenta (o el número de celular para SINPE Móvil).',
      );
    }
  });

  it('en IBAN la moneda es obligatoria; en SINPE no existe el campo (Q2)', () => {
    const iban = donacionSchema('iban').safeParse({
      ...base,
      institution: 'Banco Nacional',
      number: 'CR05015202001234567890',
    });
    expect(iban.success).toBe(false);
    if (!iban.success) {
      expect(issuesToErrors(iban.error).currency).toBe(
        'Elige la moneda de la cuenta: colones o dólares.',
      );
    }

    const sinpe = donacionSchema('sinpe').safeParse({
      ...base,
      institution: 'Banco Promerica',
      number: '88887777',
    });
    expect(sinpe.success).toBe(true);
  });

  it('usa el mensaje de número según el tipo', () => {
    const iban = donacionSchema('iban').safeParse({
      ...base,
      institution: 'Banco',
      number: '88887777',
      currency: 'CRC',
    });
    expect(iban.success).toBe(false);
    if (!iban.success) {
      expect(issuesToErrors(iban.error).number).toMatch(/IBAN válida/);
    }

    const sinpe = donacionSchema('sinpe').safeParse({
      ...base,
      institution: 'Banco',
      number: '12345',
    });
    expect(sinpe.success).toBe(false);
    if (!sinpe.success) {
      expect(issuesToErrors(sinpe.error).number).toMatch(/celular de Costa Rica/);
    }
  });

  it('la primera opción de moneda no vale como valor elegido', () => {
    expect(CURRENCY_OPTIONS[0]).toMatchObject({ value: '', disabled: true });
  });
});

describe('crossFieldErrors (Decisión 8)', () => {
  it('el titular en inglés sin español bloquea indicando qué falta', () => {
    const errors = crossFieldErrors({ ...EMPTY_DONACION_FIELDS, holderEn: 'Simiente Santa' });
    expect(errors.holderEn).toBe(
      'El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de el titular.',
    );
  });

  it('un titular vacío no obliga a nada', () => {
    expect(crossFieldErrors(EMPTY_DONACION_FIELDS)).toEqual({});
  });
});
