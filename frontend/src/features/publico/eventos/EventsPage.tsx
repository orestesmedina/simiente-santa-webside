import { Card } from '../../../components/publico/Card';
import { Notice } from '../../../components/Notice';
import { PublicButton } from '../../../components/publico/PublicButton';
import { SectionSkeleton } from '../components/SectionSkeleton';
import { useLanguage } from '../i18n/useLanguage';
import { Reveal } from '../motion/Reveal';
import { ActivitiesSection } from './ActivitiesSection';
import { EventsSection } from './EventsSection';
import { HiddenNotice } from './HiddenNotice';
import { useUpcomingEvents } from './hooks/useUpcomingEvents';

/**
 * Página pública `/eventos` (ux.md §4.1, US1/US5/US6): H1 «Eventos y
 * actividades» y **dos bloques separados por concepto** (FR-021/Q5) compuestos
 * con el sistema de F10 —`Card`, `SectionHeading`, `Reveal`— y alimentados por
 * la consulta pública compartida `useUpcomingEvents` (la misma que decide el
 * ítem de navegación, ⟲ D-23).
 *
 * Solo se dibujan los bloques **con** elementos (SC-012) y, sin nada publicado
 * y próximo, la página muestra `HiddenNotice` —el aviso de interfaz— **sin
 * redirigir** a la portada (decisión firme de la puerta 2: aquí no hay
 * `useNavigate` ni `<Navigate>`). Estados de carga (esqueletos con las medidas
 * reales) y de error (con última versión válida o con «Volver a intentar»).
 */
export function EventsPage() {
  const { t } = useLanguage();
  const query = useUpcomingEvents();

  const events = query.data?.events ?? [];
  const activities = query.data?.activities ?? [];
  const hasContent = events.length > 0 || activities.length > 0;
  const withoutContent = query.isSuccess && !hasContent;

  return (
    <section aria-label={t('a11y.events.heading')} className="bg-surface text-ink">
      <div className="mx-auto w-full max-w-6xl px-4 py-section-tight sm:px-6">
        <h1 className="font-display text-[clamp(2.5rem,7vw,4.75rem)] leading-none">
          {t('events.page.title')}
        </h1>
        {hasContent && (
          <p className="mt-4 max-w-3xl text-lg leading-relaxed text-muted">
            {t('events.page.subtitle')}
          </p>
        )}
      </div>

      {query.isPending && (
        <>
          <SectionSkeleton lines={4} />
          <SectionSkeleton lines={3} />
        </>
      )}

      {query.isError && !query.data && (
        <div className="mx-auto w-full max-w-3xl px-4 pb-section sm:px-6">
          <Card variant="glass" className="flex flex-col items-start gap-block">
            <p className="text-lg leading-relaxed text-ink">{t('events.error.load')}</p>
            <PublicButton variant="primary" onClick={() => void query.refetch()}>
              {t('events.error.retry')}
            </PublicButton>
          </Card>
        </div>
      )}

      {query.isError && query.data && (
        <div className="mx-auto w-full max-w-6xl px-4 pb-4 sm:px-6">
          <Notice variant="warning">{t('events.error.partial')}</Notice>
        </div>
      )}

      {withoutContent && <HiddenNotice />}

      {events.length > 0 && (
        <Reveal from="left">
          <EventsSection events={events} />
        </Reveal>
      )}
      {activities.length > 0 && (
        <Reveal from="right">
          <ActivitiesSection activities={activities} />
        </Reveal>
      )}
    </section>
  );
}
