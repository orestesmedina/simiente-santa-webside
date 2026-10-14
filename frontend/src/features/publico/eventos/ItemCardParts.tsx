import { formatTimeRange } from '../schedule';
import { useLanguage } from '../i18n/useLanguage';
import { datetimeValue, formatLongDate } from './formatoFechas';
import { MapLink } from './MapLink';

/**
 * Piezas compartidas por `EventCard` y `ActivityCard` (ux.md §4.1): la línea de
 * lugar con su enlace de mapa y la fecha compacta de la lista de fechas
 * restantes. Sin lógica de negocio: solo presentación sobre datos ya resueltos
 * por el servidor. La imagen de la tarjeta (`CardImage`) se promovió al
 * inventario único (`src/components/publico/CardImage.tsx`, T623) para que F6 la
 * reutilice en fichas y extracto.
 */

export interface PlaceLineProps {
  place: string;
  mapUrl?: string | null;
  /** Título del elemento, para el nombre accesible del enlace de mapa. */
  title: string;
}

/** Lugar (obligatorio) y, si el equipo cargó enlace, «Ver ubicación» (FR-018). */
export function PlaceLine({ place, mapUrl, title }: PlaceLineProps) {
  const { t } = useLanguage();

  return (
    <div className="flex flex-col gap-2">
      <p className="text-ink">
        <span className="font-medium text-muted">{t('events.location.label')}: </span>
        {place}
      </p>
      {mapUrl && <MapLink href={mapUrl} title={title} />}
    </div>
  );
}

export interface CompactTimeProps {
  date: string;
  startTime?: string | null;
  endTime?: string | null;
}

/** Fecha compacta de la lista «Más fechas» (⟲ D-11), con su `<time datetime>`. */
export function CompactTime({ date, startTime, endTime }: CompactTimeProps) {
  const { t, lang } = useLanguage();

  return (
    <time dateTime={datetimeValue(date, startTime)} className="text-ink">
      {formatLongDate(date, lang)}
      {startTime ? ` · ${formatTimeRange(startTime, endTime, t)}` : ''}
    </time>
  );
}
