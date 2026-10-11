import type { AboutPublic } from '../../../api/portada';
import { Card } from '../../../components/publico/Card';
import { SectionHeading } from '../../../components/publico/SectionHeading';
import { useLanguage } from '../i18n/useLanguage';

export interface WhoWeAreSectionProps {
  about: AboutPublic;
}

/**
 * «Quiénes somos» con el sistema visual de la propuesta B (F10 · T1022): panel
 * de vidrio (`Card glass`) con el encabezado único (`SectionHeading`, kicker 01)
 * y el texto con los saltos de línea simples preservados. Usa los tokens
 * semánticos de tema (ambos temas correctos) y `--space-section` para la
 * separación vertical (nunca padding propio mágico).
 */
export function WhoWeAreSection({ about }: WhoWeAreSectionProps) {
  const { t } = useLanguage();

  return (
    <section id="quienes-somos" aria-labelledby="quienes-somos-title" className="py-section">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <Card variant="glass" className="flex flex-col gap-block">
          <SectionHeading id="quienes-somos-title" kicker="01">
            {t('nav.sections.who')}
          </SectionHeading>
          <p className="max-w-3xl whitespace-pre-line text-lg leading-relaxed text-ink">
            {about.text}
          </p>
        </Card>
      </div>
    </section>
  );
}
