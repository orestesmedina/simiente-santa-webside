import { formatLongDate } from '../eventos/formatoFechas';
import type { Language } from '../i18n/messages';

/**
 * Fecha de publicación de un episodio (FR-006/FR-026).
 *
 * El contrato entrega `publishedAt` como instante (RFC 3339) y el **día** que
 * vale para la iglesia es el de su zona (`America/Costa_Rica`, la misma que usa
 * la unión automática por «mismo día», FR-013): el feed real
 * `Fri, 03 Jul 2026 01:58:52 GMT` es el **2 de julio** aquí. Por eso el día se
 * resuelve en esa zona y no en UTC ni en la del navegador, y después se escribe
 * con `formatLongDate` de F4 (una sola presentación de fechas en el sitio).
 *
 * El **dato** es único entre idiomas (FR-026): cambia la palabra, nunca el día.
 */
export const CHURCH_TIME_ZONE = 'America/Costa_Rica';

/** Formateador reutilizable: solo extrae el día `YYYY-MM-DD` de la zona de la iglesia. */
const DAY_FORMATTER = new Intl.DateTimeFormat('en-CA', {
  timeZone: CHURCH_TIME_ZONE,
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
});

/**
 * Día calendario del episodio en la zona de la iglesia, en `YYYY-MM-DD` (el
 * formato que entiende `formatLongDate`). Un valor que no sea una fecha válida
 * se devuelve tal cual: la vista no se rompe y el servidor sigue siendo la
 * autoridad del dato.
 */
export function episodeDay(publishedAt: string): string {
  const instant = new Date(publishedAt);
  if (Number.isNaN(instant.getTime())) {
    return publishedAt;
  }
  return DAY_FORMATTER.format(instant);
}

/** Fecha del episodio escrita en palabras, localizada por el idioma activo. */
export function formatEpisodeDate(publishedAt: string, lang: Language): string {
  return formatLongDate(episodeDay(publishedAt), lang);
}
