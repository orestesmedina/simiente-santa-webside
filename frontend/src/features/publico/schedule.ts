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

/**
 * Presentación del horario: solo la hora de inicio, o el rango «10:00 – 12:00»
 * cuando hay `endTime` (analyze C2). Se conserva el formato de 24 h del dato.
 */
export function formatTimeRange(startTime: string, endTime?: string | null): string {
  return endTime ? `${startTime} – ${endTime}` : startTime;
}
