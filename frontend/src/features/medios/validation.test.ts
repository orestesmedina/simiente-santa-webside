import { describe, expect, it } from 'vitest';
import { DESCRIPTION_MAX, descriptionLimitMessage } from './messages';
import {
  descripcionFieldsFrom,
  descripcionFieldsSchema,
  issuesToErrors,
  overDescriptionLimit,
} from './validation';

/**
 * Validación de la descripción corta (FR-020/FR-023/FR-025): `""` es «sin
 * descripción» (no un error), el tope es de 300 caracteres por idioma y el
 * inglés **exige** su español —la regla que evita el `400` del servidor—.
 */
describe('validación de la descripción corta (FR-020/FR-023)', () => {
  it('el texto vacío es válido en los dos idiomas («sin descripción», precisión 3)', () => {
    const parsed = descripcionFieldsSchema.safeParse({ descriptionEs: '', descriptionEn: '' });

    expect(parsed.success).toBe(true);
    expect(
      descripcionFieldsSchema.safeParse({ descriptionEs: '   ', descriptionEn: '' }).success,
    ).toBe(true);
  });

  it('el tope es de 300 caracteres por idioma y el mensaje dice cuánto recortar', () => {
    const justo = 'a'.repeat(DESCRIPTION_MAX);
    expect(
      descripcionFieldsSchema.safeParse({ descriptionEs: justo, descriptionEn: '' }).success,
    ).toBe(true);

    const largo = 'a'.repeat(DESCRIPTION_MAX + 1);
    const parsed = descripcionFieldsSchema.safeParse({ descriptionEs: largo, descriptionEn: '' });

    expect(parsed.success).toBe(false);
    if (!parsed.success) {
      expect(issuesToErrors(parsed.error).descriptionEs).toBe(descriptionLimitMessage(301));
      expect(issuesToErrors(parsed.error).descriptionEs).toBe(
        '301 de 300 caracteres. Recorta 1 para poder guardar.',
      );
    }
    expect(overDescriptionLimit(largo)).toBe(true);
    expect(overDescriptionLimit(justo)).toBe(false);
  });

  it('el inglés sin español bloquea el guardado y dice qué falta (FR-023/FR-025)', () => {
    const parsed = descripcionFieldsSchema.safeParse({
      descriptionEs: '',
      descriptionEn: 'The mustard seed.',
    });

    expect(parsed.success).toBe(false);
    if (!parsed.success) {
      expect(issuesToErrors(parsed.error).descriptionEn).toBe(
        'El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de la descripción corta.',
      );
    }
  });

  it('con español presente el inglés es libre (y su tope también aplica)', () => {
    expect(
      descripcionFieldsSchema.safeParse({
        descriptionEs: 'La semilla de mostaza.',
        descriptionEn: 'The mustard seed.',
      }).success,
    ).toBe(true);
    expect(
      descripcionFieldsSchema.safeParse({
        descriptionEs: 'La semilla de mostaza.',
        descriptionEn: 'a'.repeat(DESCRIPTION_MAX + 5),
      }).success,
    ).toBe(false);
  });

  it('el estado inicial precarga lo vigente y trata el `null` como vacío', () => {
    expect(descripcionFieldsFrom('La higuera.', null)).toEqual({
      descriptionEs: 'La higuera.',
      descriptionEn: '',
    });
    expect(descripcionFieldsFrom(undefined, undefined)).toEqual({
      descriptionEs: '',
      descriptionEn: '',
    });
  });
});
