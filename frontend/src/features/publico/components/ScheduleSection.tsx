import type { ScheduleItemPublic } from '../../../api/portada';
import { useLanguage } from '../i18n/useLanguage';
import { formatTimeRange, scheduleDayKey } from '../schedule';

export interface ScheduleSectionProps {
  items: ScheduleItemPublic[];
}

/**
 * Horario de servicios (ux.md §4.1): tabla ligera con `th` de ámbito en tableta
 * en adelante y la misma información como lista de tarjetas etiquetadas en
 * móvil. El día se localiza desde `dayOfWeek` (nunca texto libre) y la hora
 * muestra el rango solo cuando hay `endTime` (analyze C2).
 */
export function ScheduleSection({ items }: ScheduleSectionProps) {
  const { t } = useLanguage();

  const dayLabel = (dayOfWeek: number) => t(scheduleDayKey(dayOfWeek));

  return (
    <section id="horario" aria-labelledby="horario-title" className="py-12 sm:py-20">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <h2
          id="horario-title"
          className="font-display text-[clamp(1.75rem,4.5vw,2.5rem)] leading-tight text-navy"
        >
          {t('nav.sections.schedule')}
        </h2>

        <table className="mt-6 hidden w-full border-collapse text-left sm:table">
          <caption className="sr-only">{t('nav.sections.schedule')}</caption>
          <thead>
            <tr>
              <th scope="col" className="border-b border-navy-soft px-3 py-2 text-navy">
                {t('schedule.col.day')}
              </th>
              <th scope="col" className="border-b border-navy-soft px-3 py-2 text-navy">
                {t('schedule.col.time')}
              </th>
              <th scope="col" className="border-b border-navy-soft px-3 py-2 text-navy">
                {t('schedule.col.service')}
              </th>
              <th scope="col" className="border-b border-navy-soft px-3 py-2 text-navy">
                {t('schedule.col.place')}
              </th>
            </tr>
          </thead>
          <tbody>
            {items.map((item) => (
              <tr key={item.id} className="border-b border-navy-soft align-top">
                <td className="px-3 py-3 text-navy">{dayLabel(item.dayOfWeek)}</td>
                <td className="px-3 py-3 text-navy">
                  {formatTimeRange(item.startTime, item.endTime)}
                </td>
                <td className="px-3 py-3 font-medium text-navy">{item.name}</td>
                <td className="px-3 py-3 text-navy">{item.place}</td>
              </tr>
            ))}
          </tbody>
        </table>

        <ul className="mt-6 space-y-4 sm:hidden">
          {items.map((item) => (
            <li key={item.id} className="rounded-2xl border border-navy-soft bg-white p-4">
              <p className="font-sans text-xl font-semibold text-navy">{item.name}</p>
              <dl className="mt-2 space-y-1 text-navy">
                <div className="flex justify-between gap-3">
                  <dt className="font-medium">{t('schedule.col.day')}</dt>
                  <dd>{dayLabel(item.dayOfWeek)}</dd>
                </div>
                <div className="flex justify-between gap-3">
                  <dt className="font-medium">{t('schedule.col.time')}</dt>
                  <dd>{formatTimeRange(item.startTime, item.endTime)}</dd>
                </div>
                <div className="flex justify-between gap-3">
                  <dt className="font-medium">{t('schedule.col.place')}</dt>
                  <dd>{item.place}</dd>
                </div>
              </dl>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
