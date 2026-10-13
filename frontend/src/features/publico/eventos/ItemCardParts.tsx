import { useState } from 'react';
import { mediaUrl } from '../../../api/client';
import type { ImagenPublica } from '../../../api/agenda';
import { formatTimeRange } from '../schedule';
import { useLanguage } from '../i18n/useLanguage';
import { datetimeValue, formatLongDate } from './formatoFechas';
import { MapLink } from './MapLink';

/**
 * Piezas compartidas por `EventCard` y `ActivityCard` (ux.md §4.1): la imagen
 * a sangre del marco, la línea de lugar con su enlace de mapa y la fecha
 * compacta de la lista de fechas restantes. Sin lógica de negocio: solo
 * presentación sobre datos ya resueltos por el servidor.
 */

export interface CardImageProps {
  image?: ImagenPublica;
}

/**
 * Imagen opcional del elemento (FR-020): a sangre en el marco superior, 16:9 y
 * recorte al centro (⟲ D-9). Sin imagen la tarjeta **no muestra hueco** y, si
 * la imagen falla, un `onError` la retira sin transformar el layout (la tarjeta
 * conserva título, fecha, lugar y detalle).
 */
export function CardImage({ image }: CardImageProps) {
  const [failed, setFailed] = useState(false);
  const src = mediaUrl(image?.url);
  if (!src || failed) {
    return null;
  }

  return (
    <img
      src={src}
      alt={image?.alt ?? ''}
      onError={() => setFailed(true)}
      className="aspect-video w-full object-cover"
    />
  );
}

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
