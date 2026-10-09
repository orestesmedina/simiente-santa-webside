import type { AboutPublic } from '../../../api/portada';
import { useLanguage } from '../i18n/useLanguage';

export interface WhoWeAreSectionProps {
  about: AboutPublic;
}

/**
 * «Quiénes somos» (ux.md §4.1): texto plano con los saltos de línea simples
 * preservados (`whitespace-pre-line`) en una columna de lectura cómoda.
 */
export function WhoWeAreSection({ about }: WhoWeAreSectionProps) {
  const { t } = useLanguage();

  return (
    <section
      id="quienes-somos"
      aria-labelledby="quienes-somos-title"
      className="bg-cream py-12 sm:py-20"
    >
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <h2
          id="quienes-somos-title"
          className="font-display text-[clamp(1.75rem,4.5vw,2.5rem)] leading-tight text-navy"
        >
          {t('nav.sections.who')}
        </h2>
        <p className="mt-4 max-w-3xl whitespace-pre-line text-lg leading-relaxed text-navy">
          {about.text}
        </p>
      </div>
    </section>
  );
}
