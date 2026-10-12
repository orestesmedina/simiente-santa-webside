import type { Language } from '../i18n/messages';

/**
 * Presentación de las fechas y horas de la agenda (FR-009/R4-3/⟲ D-22).
 *
 * **Sin conversión de zona horaria**: el dato del contrato (`Date` =
 * `YYYY-MM-DD`, `ClockTime` = `HH:MM`) es el del lugar y **nunca** se
 * recalcula. La fecha se construye desde sus partes en UTC —nunca con
 * `new Date('YYYY-MM-DD')` seguido de `Intl` sobre la zona del navegador, que
 * lo interpreta como UTC y desplaza un día en zonas UTC−x (incluida
 * `America/Costa_Rica`, UTC−6)— y se formatea con `timeZone: 'UTC'` para que el
 * día real no cambie.
 *
 * Las horas se localizan con `formatTimeRange` de `../schedule` (única
 * implementación de a.m./p.m. del sitio, patrón de F3).
 */

const ISO_DATE_PATTERN = /^(\d{4})-(\d{2})-(\d{2})$/;

interface DateParts {
  year: number;
  month: number;
  day: number;
}

/** Partes de una fecha `YYYY-MM-DD` real; `undefined` si no cumple el contrato. */
function parseDate(value: string): DateParts | undefined {
  const match = ISO_DATE_PATTERN.exec(value);
  if (!match) {
    return undefined;
  }
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const probe = new Date(Date.UTC(year, month - 1, day));
  const real =
    probe.getUTCFullYear() === year &&
    probe.getUTCMonth() === month - 1 &&
    probe.getUTCDate() === day;
  return real ? { year, month, day } : undefined;
}

/**
 * Fecha escrita en palabras, localizada por idioma: «sábado 17 de octubre» /
 * «Saturday, October 17». Un valor que no es una fecha real del contrato se
 * devuelve tal cual (la vista no se rompe; el servidor es la autoridad).
 */
export function formatLongDate(value: string, lang: Language): string {
  const parts = parseDate(value);
  if (!parts) {
    return value;
  }
  const date = new Date(Date.UTC(parts.year, parts.month - 1, parts.day));
  const formatter = new Intl.DateTimeFormat(lang, {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    timeZone: 'UTC',
  });

  const pieces: Partial<Record<'weekday' | 'day' | 'month', string>> = {};
  for (const part of formatter.formatToParts(date)) {
    if (part.type === 'weekday' || part.type === 'day' || part.type === 'month') {
      pieces[part.type] = part.value;
    }
  }
  const { weekday, day, month } = pieces;
  if (!weekday || !day || !month) {
    return formatter.format(date);
  }
  return lang === 'en' ? `${weekday}, ${month} ${day}` : `${weekday} ${day} de ${month}`;
}

/**
 * Valor del atributo `datetime` de `<time>` (FR-009): el día y, si existe, la
 * hora de inicio en 24 h. Un elemento sin horario lleva solo el día (⟲ D-12).
 */
export function datetimeValue(date: string, startTime?: string | null): string {
  return startTime ? `${date}T${startTime}` : date;
}
