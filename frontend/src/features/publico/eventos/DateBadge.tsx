import { formatTimeRange } from '../schedule';
import { useLanguage } from '../i18n/useLanguage';
import { datetimeValue, formatLongDate } from './formatoFechas';

export interface DateBadgeProps {
  /** Fecha de calendario del contrato (`YYYY-MM-DD`). */
  date: string;
  /** Hora de inicio opcional (`HH:MM`); sin ella no se dibuja horario (⟲ D-12). */
  startTime?: string | null;
  /** Hora de fin opcional; solo se presenta si hay inicio. */
  endTime?: string | null;
  /** Rótulo opcional encima de la fecha (p. ej. «Próxima fecha»). */
  label?: string;
  /**
   * Variante **compacta** del extracto de la portada (ux.md §4.10): fecha y
   * horario en un cuerpo menor para que quepan en una línea en una tarjeta de
   * una quinta parte del ancho. Mismas reglas de formato y de `<time datetime>`.
   */
  compact?: boolean;
}

/**
 * Fecha y horario del elemento (ux.md §4.1, FR-009/FR-024): la fecha **escrita
 * en palabras** y localizada por idioma dentro de un `<time datetime>` que
 * transporta el dato real (`YYYY-MM-DD` + hora, sin conversión de zona), y el
 * horario al lado en Bebas como «momento de impacto» del sistema. Una fecha
 * sin hora no muestra bloque de horario (⟲ D-12).
 */
export function DateBadge({ date, startTime, endTime, label, compact = false }: DateBadgeProps) {
  const { t, lang } = useLanguage();
  const time = startTime ? formatTimeRange(startTime, endTime, t) : undefined;

  return (
    <p className="flex flex-col gap-1">
      {label && (
        <span className="font-sans text-xs font-semibold uppercase tracking-[0.16em] text-muted">
          {label}
        </span>
      )}
      <span className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
        <time
          dateTime={datetimeValue(date, startTime)}
          className={
            compact
              ? 'font-sans text-base font-semibold text-ink'
              : 'font-sans text-lg font-semibold text-ink'
          }
        >
          {formatLongDate(date, lang)}
        </time>
        {time && (
          <span
            className={
              compact
                ? 'font-display text-lg leading-none text-accent'
                : 'font-display text-2xl leading-none text-accent'
            }
            data-testid="hora"
          >
            {time}
          </span>
        )}
      </span>
    </p>
  );
}
