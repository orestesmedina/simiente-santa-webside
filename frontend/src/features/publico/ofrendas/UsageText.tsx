import { SectionHeading } from '../../../components/publico/SectionHeading';
import { useLanguage } from '../i18n/useLanguage';

export interface UsageTextProps {
  /** Texto de uso ya resuelto a `lang` por el servidor (FR-008/FR-021). */
  text: string;
}

/**
 * Bloque «¿En qué usamos las donaciones?» de `/ofrendas` (⟲ D-3, ux.md §4.1):
 * rótulo de **interfaz fija** (nivel menor, `h3`) y el contenido del equipo
 * debajo, en columna de lectura con los saltos preservados.
 *
 * **No se monta si el texto no viene** (en borrador o sin escribir): sin hueco,
 * sin marcador (US3 esc. 7). Esa decisión la toma la página.
 */
export function UsageText({ text }: UsageTextProps) {
  const { t } = useLanguage();

  return (
    <section aria-labelledby="ofrendas-usage-title" className="flex flex-col gap-block">
      <SectionHeading id="ofrendas-usage-title" level={3}>
        {t('offerings.usage.heading')}
      </SectionHeading>
      <p className="max-w-[65ch] whitespace-pre-line text-lg leading-relaxed text-ink">{text}</p>
    </section>
  );
}
