import { z } from 'zod';
import type { MetodoKind, Moneda } from '../../api/ofrendas';
import {
  CURRENCY_COLONES,
  CURRENCY_DOLARES,
  CURRENCY_PLACEHOLDER,
  HOLDER_MAX,
  IBAN_INVALID,
  INSTITUTION_MAX,
  NUMBER_INPUT_MAX,
  REQUIRED_CURRENCY,
  REQUIRED_INSTITUTION,
  REQUIRED_NUMBER,
  SINPE_INVALID,
  spanishRequiredFor,
  textTooLong,
} from './messages';

/**
 * Validación de cliente del módulo «Donaciones» (espejo de FR-014/FR-015/FR-016;
 * la autoridad es el servidor, cuyos `details` se muestran junto a cada campo).
 * El IBAN se valida por **estructura y longitud exacta por país** y el SINPE como
 * **celular costarricense de 8 dígitos con primer dígito 6/7/8**, igual que el
 * dominio (`internal/donaciones/numero.go`): el espejo avisa antes de enviar,
 * pero el valor guardado y normalizado lo produce el servidor.
 */

/* ── Campos y estado del formulario ──────────────────────────────────── */

export interface DonacionFields {
  institution: string;
  holderEs: string;
  holderEn: string;
  number: string;
  /** Moneda elegida en IBAN; `''` = sin elegir. En SINPE no aplica (Q2). */
  currency: '' | Moneda;
  /** Orden en crudo (texto del input numérico); default 0. */
  sortOrder: string;
}

export const EMPTY_DONACION_FIELDS: DonacionFields = {
  institution: '',
  holderEs: '',
  holderEn: '',
  number: '',
  currency: '',
  sortOrder: '0',
};

/** Opciones del `Select` de moneda (la primera no vale como valor, FR-014). */
export const CURRENCY_OPTIONS = [
  { value: '', label: CURRENCY_PLACEHOLDER, disabled: true },
  { value: 'CRC', label: CURRENCY_COLONES },
  { value: 'USD', label: CURRENCY_DOLARES },
];

/** Orden de foco del primer campo inválido (`id` del control y su pestaña). */
export const FOCUS_ORDER: Array<{ field: string; id: string; tab: 'es' | 'en' }> = [
  { field: 'institution', id: 'donacion-institution', tab: 'es' },
  { field: 'number', id: 'donacion-number', tab: 'es' },
  { field: 'currency', id: 'donacion-currency', tab: 'es' },
  { field: 'sortOrder', id: 'donacion-sort-order', tab: 'es' },
  { field: 'holderEs', id: 'donacion-holder-es', tab: 'es' },
  { field: 'holderEn', id: 'donacion-holder-en', tab: 'en' },
];

/* ── IBAN y SINPE (espejo puro del dominio) ──────────────────────────── */

const IBAN_STRUCTURE = /^[A-Z]{2}[0-9]{2}[A-Z0-9]+$/;

/** Registro IBAN embebido (longitud exacta por país); misma tabla del servidor. */
const IBAN_LENGTHS: Record<string, number> = {
  AD: 24,
  AE: 23,
  AL: 28,
  AT: 20,
  AZ: 28,
  BA: 20,
  BE: 16,
  BG: 22,
  BH: 22,
  BR: 29,
  CH: 21,
  CR: 22,
  CY: 28,
  CZ: 24,
  DE: 22,
  DK: 18,
  DO: 28,
  EE: 20,
  ES: 24,
  FI: 18,
  FO: 18,
  FR: 27,
  GB: 22,
  GE: 22,
  GI: 23,
  GL: 18,
  GR: 27,
  GT: 28,
  HR: 21,
  HU: 28,
  IE: 22,
  IL: 23,
  IS: 26,
  IT: 27,
  JO: 30,
  KZ: 20,
  KW: 30,
  XK: 20,
  LB: 28,
  LC: 32,
  LI: 21,
  LT: 20,
  LU: 20,
  LV: 21,
  MC: 27,
  MD: 24,
  ME: 22,
  MK: 19,
  MR: 27,
  MT: 31,
  MU: 30,
  NL: 18,
  NO: 15,
  PK: 24,
  PL: 28,
  PS: 29,
  PT: 25,
  QA: 29,
  RO: 24,
  RS: 22,
  SA: 24,
  SC: 31,
  SE: 24,
  SI: 19,
  SK: 24,
  SM: 27,
  ST: 25,
  TL: 23,
  TN: 24,
  TR: 26,
  UA: 29,
  VA: 22,
  VG: 24,
};

/**
 * Quita los separadores de lectura admitidos (espacios —incluido el no
 * separable—, tabuladores y guiones) y recorta los extremos, igual que el
 * dominio. No elimina ningún otro carácter.
 */
export function quitarSeparadores(raw: string): string {
  return raw.replace(/[ \t\u00a0\u2013\u2014-]/g, '').trim();
}

/** IBAN normalizado (mayúsculas, sin separadores) o `null` si no es válido. */
export function normalizarIban(raw: string): string | null {
  const normalizado = quitarSeparadores(raw).toUpperCase();
  if (!IBAN_STRUCTURE.test(normalizado)) {
    return null;
  }
  const longitud = IBAN_LENGTHS[normalizado.slice(0, 2)];
  if (!longitud || normalizado.length !== longitud) {
    return null;
  }
  return normalizado;
}

/** SINPE normalizado (`+506` + 8 dígitos) o `null` si no es un celular de CR. */
export function normalizarSinpe(raw: string): string | null {
  const normalizado = quitarSeparadores(raw);
  let digitos: string;
  if (normalizado.startsWith('+506')) {
    digitos = normalizado.slice(4);
  } else if (normalizado.startsWith('506') && normalizado.length === 11) {
    digitos = normalizado.slice(3);
  } else {
    digitos = normalizado;
  }
  if (digitos.length !== 8 || !/^[0-9]+$/.test(digitos)) {
    return null;
  }
  const inicio = digitos[0];
  if (inicio < '6' || inicio > '8') {
    return null;
  }
  return `+506${digitos}`;
}

/** ¿El número escrito es válido para el tipo indicado? (espejo del servidor). */
export function esNumeroValido(kind: MetodoKind, raw: string): boolean {
  return kind === 'iban' ? normalizarIban(raw) !== null : normalizarSinpe(raw) !== null;
}

/* ── Esquema Zod por campo (dependiente del tipo) ────────────────────── */

const optionalText = (max: number) =>
  z.string().trim().max(max, textTooLong(max)).optional().default('');

/**
 * Esquema del formulario para un tipo concreto: obligatorios (`institution` y
 * `number`), límites y la **moneda obligatoria solo en IBAN** (Q2/FR-014). El
 * número se valida con la regla de su tipo.
 */
export function donacionSchema(kind: MetodoKind) {
  return z.object({
    institution: z
      .string()
      .trim()
      .min(1, REQUIRED_INSTITUTION)
      .max(INSTITUTION_MAX, textTooLong(INSTITUTION_MAX)),
    holderEs: optionalText(HOLDER_MAX),
    holderEn: optionalText(HOLDER_MAX),
    number: z
      .string()
      .trim()
      .min(1, REQUIRED_NUMBER)
      .max(NUMBER_INPUT_MAX, textTooLong(NUMBER_INPUT_MAX))
      .refine(
        (value) => esNumeroValido(kind, value),
        kind === 'iban' ? IBAN_INVALID : SINPE_INVALID,
      ),
    currency: z
      .string()
      .refine((value) => kind !== 'iban' || value === 'CRC' || value === 'USD', REQUIRED_CURRENCY),
    sortOrder: z.string().optional().default('0'),
  });
}

/** Convierte los `issues` de Zod en el mapa `campo → mensaje` del formulario. */
export function issuesToErrors(error: z.ZodError): Record<string, string> {
  const mapped: Record<string, string> = {};
  for (const issue of error.issues) {
    const key = issue.path.join('.');
    if (!(key in mapped)) {
      mapped[key] = issue.message;
    }
  }
  return mapped;
}

/**
 * Regla por campo entre idiomas (Decisión 8): el titular en inglés exige su
 * versión en español; **un campo vacío no obliga a nada** (Q2, US2 esc. 8).
 */
export function crossFieldErrors(fields: DonacionFields): Record<string, string> {
  const errors: Record<string, string> = {};
  if (fields.holderEn.trim() !== '' && fields.holderEs.trim() === '') {
    errors.holderEn = spanishRequiredFor('el titular');
  }
  return errors;
}

/** `""`/espacios → `null` (el servidor guarda `NULL`, no vacíos). */
export function nullable(value: string): string | null {
  const trimmed = value.trim();
  return trimmed === '' ? null : trimmed;
}
