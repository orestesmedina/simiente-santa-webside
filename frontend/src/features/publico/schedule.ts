import type { PublicMessageKey } from './i18n/messages';

/** Claves i18n de los siete días (`dayOfWeek` 0 = domingo … 6 = sábado). */
const DAY_KEYS: readonly PublicMessageKey[] = [
  'schedule.day.0',
  'schedule.day.1',
  'schedule.day.2',
  'schedule.day.3',
  'schedule.day.4',
  'schedule.day.5',
  'schedule.day.6',
];

/**
 * Clave i18n del día a partir del valor estructural `dayOfWeek` (analyze C2:
 * el día nunca es texto libre). Un valor fuera de 0–6 cae al domingo para no
 * romper la vista (el contrato solo admite 0–6).
 */
export function scheduleDayKey(dayOfWeek: number): PublicMessageKey {
  return DAY_KEYS[dayOfWeek] ?? DAY_KEYS[0];
}

/** Traduce una clave del catálogo de interfaz al idioma activo (i18n pública). */
export type Translate = (key: PublicMessageKey) => string;

/**
 * Presentación de una hora «HH:MM» (dato del contrato en 24 h) al formato
 * localizado a.m./p.m. que promete la UX (ux.md §4.6/D-3): «10:00 a. m.»,
 * «12:00 m.», «6:00 p. m.» en español; «10:00 AM», «12:00 PM», «6:00 PM» en
 * inglés. Un valor fuera del contrato «HH:MM» se muestra tal cual para no romper
 * la vista.
 */
export function formatClockTime(value: string, t: Translate): string {
  const match = /^(\d{1,2}):(\d{2})$/.exec(value);
  if (!match) {
    return value;
  }
  const hours = Number(match[1]);
  if (hours > 23) {
    return value;
  }
  const minutes = match[2];
  const displayHour = hours % 12 === 0 ? 12 : hours % 12;
  const period: PublicMessageKey =
    hours === 12 ? 'schedule.time.noon' : hours < 12 ? 'schedule.time.am' : 'schedule.time.pm';
  return `${displayHour}:${minutes} ${t(period)}`;
}

/**
 * Presentación del horario (analyze C2, ux.md §4.6/D-3): solo la hora de inicio,
 * o el rango «10:00 a. m. − 12:00 m.» cuando hay `endTime`. El dato sigue siendo
 * «HH:MM» de 24 h; aquí solo se localiza la presentación a.m./p.m.
 */
export function formatTimeRange(
  startTime: string,
  endTime: string | null | undefined,
  t: Translate,
): string {
  const start = formatClockTime(startTime, t);
  return endTime ? `${start} − ${formatClockTime(endTime, t)}` : start;
}
