import { z } from 'zod';
import { DESCRIPTION_MAX, descriptionLimitMessage, SPANISH_REQUIRED } from './messages';

/**
 * Validación de cliente de la descripción corta (espejo de las reglas del
 * servidor, FR-020/FR-023/FR-025; la autoridad sigue siendo el servidor, cuyos
 * `details` se muestran junto a cada campo).
 *
 * Reglas:
 *  · **`""` significa «sin descripción»**, no un error: la ficha omite la línea
 *    (decisión del 2026-10-18 / precisión 3 de la puerta 2). No hay mínimo.
 *  · Hasta **300** caracteres por idioma (`DESCRIPTION_MAX`, default revisable
 *    #1); por encima, el mensaje del catálogo dice el límite y **cuánto recortar**.
 *  · **El inglés exige su español** (FR-023/FR-025): rellenar «English» sin la
 *    versión española bloquea el guardado y dice qué falta. Es la regla que, si
 *    se saltara, haría responder `400` al servidor al vaciar el español con el
 *    inglés puesto.
 *
 * **No valida datos del feed** (nombre, fecha, enlaces, carátula): esos nunca
 * pasan por un formulario (FR-019).
 */

/** Campos del formulario de descripción (texto en crudo, tal como se escribe). */
export interface DescripcionFields {
  descriptionEs: string;
  descriptionEn: string;
}

/** Estado inicial del formulario: precargado con lo vigente o vacío. */
export function descripcionFieldsFrom(
  descriptionEs?: string | null,
  descriptionEn?: string | null,
): DescripcionFields {
  return {
    descriptionEs: descriptionEs ?? '',
    descriptionEn: descriptionEn ?? '',
  };
}

/**
 * Cuenta **runas** (puntos de código), como el servidor (`len([]rune(…))`) y la
 * base de datos (`char_length`), y no unidades UTF-16: un emoji —que en UTF-16
 * ocupa un par sustituto— cuenta **1**, no 2. Con `value.length` un texto de 300
 * caracteres reales con emojis se rechazaba antes de tiempo y el mensaje de
 * «cuánto recortar» contaba de más (B1 de la revisión).
 */
export function descriptionLength(value: string): number {
  return Array.from(value).length;
}

/** Un texto por idioma: sin mínimo y con el tope del contrato. */
function descriptionText(): z.ZodType<string> {
  return z.string().superRefine((value, ctx) => {
    const count = descriptionLength(value);
    if (count > DESCRIPTION_MAX) {
      ctx.addIssue({ code: 'custom', message: descriptionLimitMessage(count) });
    }
  });
}

export const descripcionFieldsSchema = z
  .object({
    descriptionEs: descriptionText(),
    descriptionEn: descriptionText(),
  })
  .superRefine((values, ctx) => {
    if (values.descriptionEn.trim() !== '' && values.descriptionEs.trim() === '') {
      ctx.addIssue({ code: 'custom', path: ['descriptionEn'], message: SPANISH_REQUIRED });
    }
  });

/** Errores por campo a partir de un `ZodError` (el primero de cada campo). */
export function issuesToErrors(error: z.ZodError): Record<string, string> {
  const mapped: Record<string, string> = {};
  for (const issue of error.issues) {
    const key = issue.path.join('.');
    if (key !== '' && !(key in mapped)) {
      mapped[key] = issue.message;
    }
  }
  return mapped;
}

/** `true` si el texto supera el tope (contador en error y «Guardar» deshabilitado). */
export function overDescriptionLimit(value: string): boolean {
  return descriptionLength(value) > DESCRIPTION_MAX;
}
