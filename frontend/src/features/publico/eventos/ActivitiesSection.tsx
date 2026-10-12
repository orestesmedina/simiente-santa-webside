import type { ActividadPublica } from '../../../api/agenda';
import { SectionHeading } from '../../../components/publico/SectionHeading';
import { Reveal } from '../motion/Reveal';
import { RevealGroup } from '../motion/RevealGroup';
import { useLanguage } from '../i18n/useLanguage';
import { ActivityCard } from './ActivityCard';

export interface ActivitiesSectionProps {
  /** Actividades próximas ya ordenadas por el servidor (FR-003). */
  activities: ActividadPublica[];
}

/**
 * Bloque «Próximas actividades» (ux.md §4.1, FR-021): misma maqueta que el
 * bloque de eventos con el kicker 02. Dos conceptos, cero confusión: lo
 * periódico y lo puntual viven en bloques separados y rotulados (Q5).
 */
export function ActivitiesSection({ activities }: ActivitiesSectionProps) {
  const { t } = useLanguage();

  return (
    <section id="actividades" aria-labelledby="actividades-title" className="py-section">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <SectionHeading id="actividades-title" kicker="02">
          {t('events.block.activities')}
        </SectionHeading>

        <RevealGroup className="mt-block grid gap-block sm:grid-cols-2">
          {activities.map((activity) => (
            <Reveal key={activity.id} className="h-full">
              <ActivityCard activity={activity} />
            </Reveal>
          ))}
        </RevealGroup>
      </div>
    </section>
  );
}
