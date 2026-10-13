import type { ActividadPublica } from '../../../api/agenda';
import { Card } from '../../../components/publico/Card';
import { useLanguage } from '../i18n/useLanguage';
import { CategoryChip } from './CategoryChip';
import { DateBadge } from './DateBadge';
import { CardImage, PlaceLine } from './ItemCardParts';

export interface ActivityCardProps {
  /** Actividad publicada con **fecha única** (FR-005/Decisión 3). */
  activity: ActividadPublica;
}

/**
 * Tarjeta de una actividad puntual (ux.md §4.1, FR-005/FR-019/FR-020): misma
 * maqueta que `EventCard` con tres diferencias — **no hay frecuencia**, el
 * rótulo de fecha muestra **una sola fecha** y el horario puede faltar por
 * completo (⟲ D-12: no se dibuja «Hora: —»).
 */
export function ActivityCard({ activity }: ActivityCardProps) {
  const { t } = useLanguage();

  return (
    <Card as="article" padded={false} className="flex h-full flex-col overflow-hidden">
      <CardImage image={activity.image} />

      <div className="flex flex-1 flex-col gap-3 p-panel">
        {activity.category && <CategoryChip name={activity.category.name} />}

        <h3 className="font-sans text-xl font-semibold text-ink">{activity.title}</h3>

        <DateBadge
          date={activity.date}
          startTime={activity.startTime}
          endTime={activity.endTime}
          label={t('events.date.next')}
        />

        <PlaceLine place={activity.place} mapUrl={activity.mapUrl} title={activity.title} />

        <p className="whitespace-pre-line text-lg leading-relaxed text-ink">{activity.detail}</p>

        {activity.organizer && <p className="text-sm text-muted">{activity.organizer}</p>}
      </div>
    </Card>
  );
}
