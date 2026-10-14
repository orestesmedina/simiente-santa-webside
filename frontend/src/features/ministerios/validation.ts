import { z } from 'zod';
import {
  DESCRIPTION_MAX,
  EMAIL_INVALID,
  LEADER_NAME_MAX,
  LINK_INVALID,
  LINK_UNSAFE,
  NAME_MAX,
  PHONE_INVALID,
  PLACE_MAX,
  REQUIRED_DESCRIPTION,
  REQUIRED_LEADER_NAME_FOR_CONTACT,
  REQUIRED_NAME,
  SCHEDULE_MAX,
  WHATSAPP_URL_MAX,
  spanishRequiredFor,
  textTooLong,
} from './messages';

/**
 * Validación de cliente del módulo «Ministerios» (espejo de FR-014/FR-015; la
 * autoridad es el servidor, cuyos `details` se muestran junto a cada campo).
 * Zod cubre las reglas por campo y este módulo añade las **reglas por campo
 * entre idiomas** (Q6) y las del bloque Encargado (Q8) que Zod no expresa.
 */

/* ── Tipos y estado del formulario ───────────────────────────────────── */

/** Campos lingüísticos y no lingüísticos del formulario (texto en crudo). */
export interface MinistryFields {
  nameEs: string;
  nameEn: string;
  descriptionEs: string;
  descriptionEn: string;
  scheduleEs: string;
  scheduleEn: string;
  placeEs: string;
  placeEn: string;
  leaderNameEs: string;
  leaderNameEn: string;
  whatsappUrl: string;
}

export const EMPTY_MINISTRY_FIELDS: MinistryFields = {
  nameEs: '',
  nameEn: '',
  descriptionEs: '',
  descriptionEn: '',
  scheduleEs: '',
  scheduleEn: '',
  placeEs: '',
  placeEn: '',
  leaderNameEs: '',
  leaderNameEn: '',
  whatsappUrl: '',
};

/** Único medio de contacto del encargado (Q8: nunca los dos). */
export type ContactKind = 'none' | 'phone' | 'email';

/** Orden de foco del primer campo inválido (`id` del control y su pestaña). */
export const FOCUS_ORDER: Array<{ field: string; id: string; tab: 'es' | 'en' }> = [
  { field: 'nameEs', id: 'ministry-name-es', tab: 'es' },
  { field: 'descriptionEs', id: 'ministry-description-es', tab: 'es' },
  { field: 'scheduleEs', id: 'ministry-schedule-es', tab: 'es' },
  { field: 'placeEs', id: 'ministry-place-es', tab: 'es' },
  { field: 'leaderNameEs', id: 'ministry-leader-name-es', tab: 'es' },
  { field: 'whatsappUrl', id: 'ministry-whatsapp-url', tab: 'es' },
  { field: 'nameEn', id: 'ministry-name-en', tab: 'en' },
  { field: 'descriptionEn', id: 'ministry-description-en', tab: 'en' },
  { field: 'scheduleEn', id: 'ministry-schedule-en', tab: 'en' },
  { field: 'placeEn', id: 'ministry-place-en', tab: 'en' },
  { field: 'leaderNameEn', id: 'ministry-leader-name-en', tab: 'en' },
];

/* ── Reglas compartidas ──────────────────────────────────────────────── */

/** `""`/espacios → `null` (analyze I6: el servidor guarda `NULL`, no vacíos). */
export function nullable(value: string): string | null {
  const trimmed = value.trim();
  return trimmed === '' ? null : trimmed;
}

/**
 * Enlace `https` con host no vacío (espejo del tag `url` del servidor). Se
 * acepta **cualquier dominio** (Q3) y el valor nunca se transforma (FR-015).
 */
export function isSecureUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === 'https:' && url.host !== '';
  } catch {
    return false;
  }
}

/** ¿Tiene esquema (`https://`, `javascript:`, `mailto:`…)? */
const HAS_SCHEME = /^[a-z][a-z0-9+.-]*:/i;

/**
 * Mensaje del enlace (FR-015): sin esquema o sin host → «no parece una dirección
 * web»; con esquema distinto de `https` (incluidos `javascript:`, `data:`,
 * `vbscript:` y `http://`) → «no es válido o no es seguro».
 */
export function linkError(value: string): string | undefined {
  const trimmed = value.trim();
  if (trimmed === '') {
    return undefined;
  }
  if (!HAS_SCHEME.test(trimmed)) {
    return LINK_INVALID;
  }
  if (!/^https:/i.test(trimmed)) {
    return LINK_UNSAFE;
  }
  return isSecureUrl(trimmed) ? undefined : LINK_INVALID;
}

/** Celular con los separadores habituales y ≥ 7 dígitos (espejo del servidor). */
export function isPhone(value: string): boolean {
  const trimmed = value.trim();
  if (!/^\+?[0-9() -]+$/.test(trimmed)) {
    return false;
  }
  return trimmed.replace(/\D/g, '').length >= 7;
}

/** Correo con forma de correo (espejo del tag `email` del servidor). */
export function isEmail(value: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(value.trim());
}

/* ── Esquema Zod por campo ───────────────────────────────────────────── */

const optionalText = (max: number) =>
  z.string().trim().max(max, textTooLong(max)).optional().default('');

export const ministryFieldsSchema = z.object({
  nameEs: z.string().trim().min(1, REQUIRED_NAME).max(NAME_MAX, textTooLong(NAME_MAX)),
  nameEn: optionalText(NAME_MAX),
  descriptionEs: z
    .string()
    .trim()
    .min(1, REQUIRED_DESCRIPTION)
    .max(DESCRIPTION_MAX, textTooLong(DESCRIPTION_MAX)),
  descriptionEn: optionalText(DESCRIPTION_MAX),
  scheduleEs: optionalText(SCHEDULE_MAX),
  scheduleEn: optionalText(SCHEDULE_MAX),
  placeEs: optionalText(PLACE_MAX),
  placeEn: optionalText(PLACE_MAX),
  leaderNameEs: optionalText(LEADER_NAME_MAX),
  leaderNameEn: optionalText(LEADER_NAME_MAX),
  whatsappUrl: z
    .string()
    .trim()
    .max(WHATSAPP_URL_MAX, textTooLong(WHATSAPP_URL_MAX))
    .refine((value) => linkError(value) === undefined, {
      message: LINK_UNSAFE,
    }),
});

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

/* ── Reglas por campo entre idiomas (Q6) y del encargado (Q8) ────────── */

/** Pares `*En`/`*Es` con la etiqueta que usa el mensaje de español faltante. */
const LANGUAGE_PAIRS: Array<{ en: keyof MinistryFields; es: keyof MinistryFields; label: string }> =
  [
    { en: 'nameEn', es: 'nameEs', label: 'el nombre' },
    { en: 'descriptionEn', es: 'descriptionEs', label: 'la descripción' },
    { en: 'scheduleEn', es: 'scheduleEs', label: 'el horario' },
    { en: 'placeEn', es: 'placeEs', label: 'el lugar' },
    { en: 'leaderNameEn', es: 'leaderNameEs', label: 'el nombre del encargado' },
  ];

export interface MinistryFormValues {
  fields: MinistryFields;
  contactKind: ContactKind;
  contactValue: string;
}

/**
 * Reglas de coherencia que Zod no expresa: **español por campo** (un campo
 * rellenado en inglés sin su versión española bloquea el guardado indicando
 * cuál falta; **un campo vacío no obliga a nada**, Q6) y **contacto** (un solo
 * medio, con formato válido, y siempre con nombre del encargado, Q8).
 */
export function crossFieldErrors(values: MinistryFormValues): Record<string, string> {
  const errors: Record<string, string> = {};
  const { fields, contactKind, contactValue } = values;

  for (const pair of LANGUAGE_PAIRS) {
    if (fields[pair.en].trim() !== '' && fields[pair.es].trim() === '') {
      errors[pair.en] = spanishRequiredFor(pair.label);
    }
  }

  const contact = contactValue.trim();
  if (contactKind !== 'none' && contact !== '') {
    if (fields.leaderNameEs.trim() === '') {
      errors.leaderNameEs = REQUIRED_LEADER_NAME_FOR_CONTACT;
    }
    if (contactKind === 'phone' && !isPhone(contact)) {
      errors.contactValue = PHONE_INVALID;
    }
    if (contactKind === 'email' && !isEmail(contact)) {
      errors.contactValue = EMAIL_INVALID;
    }
  }

  return errors;
}
