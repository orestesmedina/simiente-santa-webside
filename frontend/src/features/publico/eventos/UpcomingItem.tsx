import { Link } from 'react-router-dom';
import type { ItemProximoPortada } from '../../../api/agenda';
import { Card } from '../../../components/publico/Card';
import { CardImage } from '../../../components/publico/CardImage';
import { useLanguage } from '../i18n/useLanguage';
import { DateBadge } from './DateBadge';

export interface UpcomingItemProps {
  /** Ítem del extracto, ya resuelto por el servidor a `lang` (FR-008). */
  item: ItemProximoPortada;
}

/**
 * Tarjeta **compacta** del extracto «Lo que viene» de la portada (ux.md §4.10,
 * ⟲ D-25/FR-025/FR-026): imagen 16:9 opcional, etiqueta de concepto
 * («Evento»/«Actividad», texto y no color), título que **es el enlace** al
 * bloque correspondiente de `/eventos`, fecha localizada con su `<time>` y el
 * lugar.
 *
 * Misma gramática que `EventCard`/`ActivityCard` pero sin chip de categoría, sin
 * frecuencia y sin detalle (el DTO del extracto no los trae: el detalle vive en
 * la sección). La tarjeta **no** es un enlace en bloque: solo el título lo es,
 * con el nombre accesible completo «Evento: {título}».
 */
export function UpcomingItem({ item }: UpcomingItemProps) {
  const { t } = useLanguage();
  const isEvent = item.kind === 'event';
  const badge = isEvent ? t('home.upcoming.badge.event') : t('home.upcoming.badge.activity');
  const to = isEvent ? '/eventos#eventos' : '/eventos#actividades';

  return (
    <Card as="li" padded={false} className="flex flex-col overflow-hidden">
      <CardImage image={item.image} />

      <div className="flex flex-1 flex-col gap-2 p-fluid">
        <span className="w-fit rounded-full border border-line px-3 py-1 font-sans text-sm text-muted">
          {badge}
        </span>

        <h3 className="font-sans text-[1.0625rem] font-semibold leading-snug text-ink">
          {/* El enlace es el **único** elemento clicable de la tarjeta y cumple
              el objetivo táctil ≥ 44 px (FR-023/FR-024): la caja es un `flex`
              de 44 px con el título —recortado a 2 líneas— centrado dentro. */}
          <Link
            to={to}
            aria-label={`${badge}: ${item.title}`}
            className="flex min-h-11 items-center underline-offset-4 hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus"
          >
            <span className="line-clamp-2">{item.title}</span>
          </Link>
        </h3>

        <DateBadge date={item.date} startTime={item.startTime} endTime={item.endTime} compact />

        <p className="flex items-center gap-1 text-muted">
          <span aria-hidden="true">📍</span>
          <span className="truncate" title={item.place}>
            {item.place}
          </span>
        </p>
      </div>
    </Card>
  );
}
