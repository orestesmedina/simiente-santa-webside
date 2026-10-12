import { z } from 'zod';
import {
  DATE_INVALID,
  DETAIL_MAX,
  DUPLICATE_DATE,
  END_TIME_INVALID,
  END_WITHOUT_START,
  FREQUENCY_MAX,
  MAP_URL_INVALID,
  MAP_URL_MAX,
  MAX_DATES,
  ORGANIZER_MAX,
  PLACE_MAX,
  REQUIRED_DETAIL,
  REQUIRED_PLACE,
  REQUIRED_TITLE,
  TIME_INVALID,
  TITLE_MAX,
  TOO_MANY_DATES,
  textTooLong,
} from './messages';

/**
 * Validación de cliente del panel (espejo de FR-015; la autoridad es el
 * servidor, cuyos `details` se muestran junto a cada campo). Nada de librerías
 * de fechas: el contrato fija `YYYY-MM-DD` y `HH:MM` y aquí solo se comprueba
 * su forma y coherencia (⟲ D-4: inputs nativos).
 */

export const DATE_PATTERN = /^(\d{4})-(\d{2})-(\d{2})$/;
export const TIME_PATTERN = /^([01]\d|2[0-3]):[0-5]\d$/;

/** Fecha de calendario real (`2026-02-30` no existe). */
export function isRealDate(value: string): boolean {
  const match = DATE_PATTERN.exec(value);
  if (!match) {
    return false;
  }
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const probe = new Date(Date.UTC(year, month - 1, day));
  return (
    probe.getUTCFullYear() === year &&
    probe.getUTCMonth() === month - 1 &&
    probe.getUTCDate() === day
  );
}

/** Enlace `https` con host no vacío (espejo de la etiqueta `url` del servidor). */
export function isWebUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === 'https:' && url.host !== '';
  } catch {
    return false;
  }
}

/** `""`/espacios → `null` (analyze I6: el servidor guarda `NULL`, no vacíos). */
export function nullable(value: string): string | null {
  const trimmed = value.trim();
  return trimmed === '' ? null : trimmed;
}

/* ── Campos comunes a eventos y actividades ──────────────────────────── */

export interface ItemFields {
  titleEs: string;
  titleEn: string;
  detailEs: string;
  detailEn: string;
  placeEs: string;
  placeEn: string;
  mapUrl: string;
  organizerEs: string;
  organizerEn: string;
}

export const EMPTY_ITEM_FIELDS: ItemFields = {
  titleEs: '',
  titleEn: '',
  detailEs: '',
  detailEn: '',
  placeEs: '',
  placeEn: '',
  mapUrl: '',
  organizerEs: '',
  organizerEn: '',
};

const optionalText = (max: number) =>
  z.string().trim().max(max, textTooLong(max)).optional().default('');

export const itemFieldsSchema = z.object({
  titleEs: z.string().trim().min(1, REQUIRED_TITLE).max(TITLE_MAX, textTooLong(TITLE_MAX)),
  titleEn: optionalText(TITLE_MAX),
  detailEs: z.string().trim().min(1, REQUIRED_DETAIL).max(DETAIL_MAX, textTooLong(DETAIL_MAX)),
  detailEn: optionalText(DETAIL_MAX),
  placeEs: z.string().trim().min(1, REQUIRED_PLACE).max(PLACE_MAX, textTooLong(PLACE_MAX)),
  placeEn: optionalText(PLACE_MAX),
  mapUrl: z
    .string()
    .trim()
    .max(MAP_URL_MAX, textTooLong(MAP_URL_MAX))
    .refine((value) => value === '' || isWebUrl(value), MAP_URL_INVALID),
  organizerEs: optionalText(ORGANIZER_MAX),
  organizerEn: optionalText(ORGANIZER_MAX),
});

/** Frecuencia del evento (solo español obligatorio si existe; sin recurrencias, Q1). */
export const frequencySchema = z
  .object({
    frequencyEs: optionalText(FREQUENCY_MAX),
    frequencyEn: optionalText(FREQUENCY_MAX),
  })
  .partial();

/** Convierte los `issues` de Zod en el mapa `campo → mensaje` que usan los formularios. */
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

/* ── Lista de fechas del evento ──────────────────────────────────────── */

/** Fila del editor de fechas: los tres valores en crudo del input nativo. */
export interface DateRow {
  date: string;
  startTime: string;
  endTime: string;
}

export const EMPTY_DATE_ROW: DateRow = { date: '', startTime: '', endTime: '' };

export interface DateRowErrors {
  date?: string;
  startTime?: string;
  endTime?: string;
}

export interface DateRowsValidation {
  /** Errores por fila (índice de la fila → errores de sus campos). */
  rows: Record<number, DateRowErrors>;
  /** Aviso general del bloque de fechas (vacío, tope o fecha repetida). */
  general?: string;
}

const dateRowSchema = z.object({
  date: z.string().trim(),
  startTime: z.string().trim(),
  endTime: z.string().trim(),
});

/**
 * Valida la lista completa de fechas (FR-015/⟲ D-4 y decisión firme de la
 * puerta 2): al menos una fila, cada fecha real y sin repetir, `≤ 50` filas, y
 * por fila **la hora de fin exige la de inicio** y debe ser posterior a ella.
 * Devuelve los errores por fila y el aviso general; sin errores devuelve un
 * objeto vacío.
 */
export function validateDateRows(rows: DateRow[], requiredMessage: string): DateRowsValidation {
  const result: DateRowsValidation = { rows: {} };
  if (rows.length === 0) {
    return { rows: {}, general: requiredMessage };
  }
  if (rows.length > MAX_DATES) {
    result.general = TOO_MANY_DATES;
  }

  const seen = new Set<string>();

  rows.forEach((row, index) => {
    const parsed = dateRowSchema.safeParse(row);
    const errors: DateRowErrors = {};
    const values = parsed.success ? parsed.data : row;

    if (values.date === '') {
      errors.date = requiredMessage;
    } else if (!isRealDate(values.date)) {
      errors.date = DATE_INVALID;
    } else if (seen.has(values.date)) {
      errors.date = DUPLICATE_DATE;
    } else {
      seen.add(values.date);
    }

    if (values.startTime !== '' && !TIME_PATTERN.test(values.startTime)) {
      errors.startTime = TIME_INVALID;
    }
    if (values.endTime !== '') {
      if (values.startTime === '') {
        errors.endTime = END_WITHOUT_START;
      } else if (!TIME_PATTERN.test(values.endTime)) {
        errors.endTime = TIME_INVALID;
      } else if (values.endTime <= values.startTime) {
        errors.endTime = END_TIME_INVALID;
      }
    }

    if (Object.keys(errors).length > 0) {
      result.rows[index] = errors;
    }
  });

  return result;
}

/** `true` si la validación de la lista no encontró ningún problema. */
export function isDateRowsValid(validation: DateRowsValidation): boolean {
  return validation.general === undefined && Object.keys(validation.rows).length === 0;
}

/* ── Fecha única de una actividad ────────────────────────────────────── */

export interface OccurrenceErrors {
  date?: string;
  startTime?: string;
  endTime?: string;
}

/**
 * Valida la fecha única de una actividad (FR-005/FR-015/Decisión 3): fecha real
 * obligatoria, horas `HH:MM` opcionales y **la hora de fin exige la de inicio**
 * y debe ser posterior (misma decisión firme que en los eventos).
 */
export function validateOccurrence(
  date: string,
  startTime: string,
  endTime: string,
  requiredMessage: string,
): OccurrenceErrors {
  const errors: OccurrenceErrors = {};

  if (date.trim() === '') {
    errors.date = requiredMessage;
  } else if (!isRealDate(date)) {
    errors.date = DATE_INVALID;
  }

  if (startTime !== '' && !TIME_PATTERN.test(startTime)) {
    errors.startTime = TIME_INVALID;
  }
  if (endTime !== '') {
    if (startTime === '') {
      errors.endTime = END_WITHOUT_START;
    } else if (!TIME_PATTERN.test(endTime)) {
      errors.endTime = TIME_INVALID;
    } else if (endTime <= startTime) {
      errors.endTime = END_TIME_INVALID;
    }
  }

  return errors;
}
