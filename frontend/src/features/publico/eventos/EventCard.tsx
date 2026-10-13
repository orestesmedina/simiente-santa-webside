import type { EventoPublico } from '../../../api/agenda';
import { Card } from '../../../components/publico/Card';
import { useLanguage } from '../i18n/useLanguage';
import { CategoryChip } from './CategoryChip';
import { DateBadge } from './DateBadge';
import { CardImage, CompactTime, PlaceLine } from './ItemCardParts';

export interface EventCardProps {
  /** Evento publicado con **al menos una fecha próxima** (FR-004). */
  event: EventoPublico;
}

/**
 * Tarjeta de un evento (ux.md §4.1, FR-004/FR-018/FR-020): imagen opcional,
 * chip de categoría, título, la **próxima fecha** destacada (el servidor la
 * entrega primero en `dates`), el texto de frecuencia si existe, las demás
 * fechas próximas (todas a la vista, ⟲ D-11), lugar con enlace de mapa y el
 * detalle completo (sin «Ver más», ⟲ D-3).
 *
 * El orden del DOM es el cronológico que el servidor promete (FR-003): la
 * tarjeta no reordena nada.
 */
export function EventCard({ event }: EventCardProps) {
  const { t } = useLanguage();
  const [nextDate, ...remainingDates] = event.dates;

  return (
    <Card as="article" padded={false} className="flex h-full flex-col overflow-hidden">
      <CardImage image={event.image} />

      <div className="flex flex-1 flex-col gap-3 p-panel">
        {event.category && <CategoryChip name={event.category.name} />}

        <h3 className="font-sans text-xl font-semibold text-ink">{event.title}</h3>

        {nextDate && (
          <DateBadge
            date={nextDate.date}
            startTime={nextDate.startTime}
            endTime={nextDate.endTime}
            label={t('events.date.next')}
          />
        )}

        {event.frequency && (
          <p className="text-balance font-sans italic text-muted">{event.frequency}</p>
        )}

        {remainingDates.length > 0 && (
          <div>
            <p className="font-sans text-xs font-semibold uppercase tracking-[0.16em] text-muted">
              {t('events.date.more')}
            </p>
            <ul className="mt-1 flex flex-col gap-1">
              {remainingDates.map((occurrence) => (
                <li key={occurrence.date}>
                  <CompactTime
                    date={occurrence.date}
                    startTime={occurrence.startTime}
                    endTime={occurrence.endTime}
                  />
                </li>
              ))}
            </ul>
          </div>
        )}

        <PlaceLine place={event.place} mapUrl={event.mapUrl} title={event.title} />

        <p className="whitespace-pre-line text-lg leading-relaxed text-ink">{event.detail}</p>

        {event.organizer && <p className="text-sm text-muted">{event.organizer}</p>}
      </div>
    </Card>
  );
}
