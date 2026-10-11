import type { ScheduleItemPublic } from '../../../api/portada';
import { Card } from '../../../components/publico/Card';
import { SectionHeading } from '../../../components/publico/SectionHeading';
import { useLanguage } from '../i18n/useLanguage';
import { formatTimeRange, scheduleDayKey } from '../schedule';

export interface ScheduleSectionProps {
  items: ScheduleItemPublic[];
}

/**
 * Horario de servicios con el sistema visual de la propuesta B (F10 · T1022):
 * panel de vidrio, hora en Bebas turquesa y filas legibles. **Conserva la
 * representación dual de F3** (contrato congelado, M2): la `table`
 * (`hidden … sm:table`) con `caption` `sr-only` cuyo nombre accesible sigue
 * siendo «Horario de servicios» (lo exige `portada-publica.spec.ts`) **y** la
 * lista de tarjetas `ul` (`sm:hidden`, consultada como `#horario li` por
 * `portada-accesibilidad.spec.ts` a 160 px). Solo cambia el vestido.
 */
export function ScheduleSection({ items }: ScheduleSectionProps) {
  const { t } = useLanguage();

  const dayLabel = (dayOfWeek: number) => t(scheduleDayKey(dayOfWeek));

  return (
    <section id="horario" aria-labelledby="horario-title" className="py-section">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <Card variant="glass" className="flex flex-col gap-block">
          <SectionHeading id="horario-title" kicker="02">
            {t('nav.sections.schedule')}
          </SectionHeading>

          <table className="hidden w-full border-collapse text-left sm:table">
            <caption className="sr-only">{t('nav.sections.schedule')}</caption>
            <thead>
              <tr>
                <th scope="col" className="border-b border-line px-3 py-2 text-accent">
                  {t('schedule.col.day')}
                </th>
                <th scope="col" className="border-b border-line px-3 py-2 text-accent">
                  {t('schedule.col.time')}
                </th>
                <th scope="col" className="border-b border-line px-3 py-2 text-accent">
                  {t('schedule.col.service')}
                </th>
                <th scope="col" className="border-b border-line px-3 py-2 text-accent">
                  {t('schedule.col.place')}
                </th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-b border-line align-top">
                  <td className="px-3 py-3 text-ink">{dayLabel(item.dayOfWeek)}</td>
                  <td className="px-3 py-3 font-display text-2xl leading-none text-accent">
                    {formatTimeRange(item.startTime, item.endTime, t)}
                  </td>
                  <td className="px-3 py-3 font-medium text-ink">{item.name}</td>
                  <td className="px-3 py-3 text-ink">{item.place}</td>
                </tr>
              ))}
            </tbody>
          </table>

          <ul className="flex flex-col gap-fluid sm:hidden">
            {items.map((item) => (
              <li key={item.id} className="rounded-2xl border border-line bg-surface-raised p-4">
                <p className="font-sans text-xl font-semibold text-ink">{item.name}</p>
                <p className="mt-1 font-display text-3xl leading-none text-accent">
                  {formatTimeRange(item.startTime, item.endTime, t)}
                </p>
                <dl className="mt-2 flex flex-col gap-1 text-ink">
                  <div className="flex justify-between gap-3">
                    <dt className="font-medium text-muted">{t('schedule.col.day')}</dt>
                    <dd>{dayLabel(item.dayOfWeek)}</dd>
                  </div>
                  <div className="flex justify-between gap-3">
                    <dt className="font-medium text-muted">{t('schedule.col.place')}</dt>
                    <dd>{item.place}</dd>
                  </div>
                </dl>
              </li>
            ))}
          </ul>
        </Card>
      </div>
    </section>
  );
}
