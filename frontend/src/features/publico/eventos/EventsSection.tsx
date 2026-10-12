import type { EventoPublico } from '../../../api/agenda';
import { SectionHeading } from '../../../components/publico/SectionHeading';
import { Reveal } from '../motion/Reveal';
import { RevealGroup } from '../motion/RevealGroup';
import { useLanguage } from '../i18n/useLanguage';
import { EventCard } from './EventCard';

export interface EventsSectionProps {
  /** Eventos próximos ya ordenados por el servidor (FR-003). */
  events: EventoPublico[];
}

/**
 * Bloque «Próximos eventos» (ux.md §4.1, FR-021): encabezado único del sistema
 * (kicker 01) y grilla de tarjetas. El bloque **solo se dibuja si tiene
 * elementos** (⟲ D-5/SC-012): nunca queda un rótulo sin contenido.
 */
export function EventsSection({ events }: EventsSectionProps) {
  const { t } = useLanguage();

  return (
    <section id="eventos" aria-labelledby="eventos-title" className="py-section">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <SectionHeading id="eventos-title" kicker="01">
          {t('events.block.scheduled')}
        </SectionHeading>

        <RevealGroup className="mt-block grid gap-block sm:grid-cols-2">
          {events.map((event) => (
            <Reveal key={event.id} className="h-full">
              <EventCard event={event} />
            </Reveal>
          ))}
        </RevealGroup>
      </div>
    </section>
  );
}
