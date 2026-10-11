import type { SocialLinkPublic } from '../../../api/portada';
import { Card } from '../../../components/publico/Card';
import { SectionHeading } from '../../../components/publico/SectionHeading';
import { useLanguage } from '../i18n/useLanguage';
import { SOCIAL_NETWORK_LABELS } from '../social';
import { SocialIcon } from './SocialIcon';

export interface SocialSectionProps {
  items: SocialLinkPublic[];
}

/**
 * Redes sociales con el sistema visual de la propuesta B (F10 · T1024): chips
 * con elevación y relleno suave en hover (**solo `transform`/colores**) y el
 * nombre accesible de F3 intacto («Ver en {red}»). `SocialIcon` conserva su
 * contrato (icono decorativo con `currentColor`, dentro de la paleta).
 */
export function SocialSection({ items }: SocialSectionProps) {
  const { t } = useLanguage();

  return (
    <section id="redes" aria-labelledby="redes-title" className="py-section">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <Card variant="glass" className="flex flex-col gap-block">
          <SectionHeading id="redes-title" kicker="05">
            {t('nav.sections.social')}
          </SectionHeading>

          <ul className="flex flex-wrap gap-fluid">
            {items.map((link) => (
              <li key={link.id}>
                <a
                  href={link.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex min-h-11 items-center gap-3 rounded-full border border-line bg-surface-raised px-5 py-2 font-medium text-ink transition duration-[var(--motion-duration-micro)] ease-[var(--motion-ease-standard)] hover:-translate-y-1 hover:border-accent focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus"
                >
                  <SocialIcon network={link.network} />
                  <span>
                    {t('social.action', { network: SOCIAL_NETWORK_LABELS[link.network] })}
                  </span>
                </a>
              </li>
            ))}
          </ul>
        </Card>
      </div>
    </section>
  );
}
