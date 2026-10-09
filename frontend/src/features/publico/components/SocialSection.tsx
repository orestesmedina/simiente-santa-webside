import type { SocialLinkPublic } from '../../../api/portada';
import { useLanguage } from '../i18n/useLanguage';
import { SOCIAL_NETWORK_LABELS } from '../social';
import { SocialIcon } from './SocialIcon';

export interface SocialSectionProps {
  items: SocialLinkPublic[];
}

/**
 * Redes sociales (ux.md §4.1): una tarjeta por red publicada con icono **y**
 * nombre visible; el enlace se abre en una pestaña nueva sin acceso al opener.
 */
export function SocialSection({ items }: SocialSectionProps) {
  const { t } = useLanguage();

  return (
    <section id="redes" aria-labelledby="redes-title" className="py-12 sm:py-20">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <h2
          id="redes-title"
          className="font-display text-[clamp(1.75rem,4.5vw,2.5rem)] leading-tight text-navy"
        >
          {t('nav.sections.social')}
        </h2>
        <ul className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {items.map((link) => (
            <li key={link.id}>
              <a
                href={link.url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex min-h-12 items-center gap-3 rounded-2xl border border-navy-soft bg-white p-4 font-medium text-navy hover:border-teal focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-navy"
              >
                <SocialIcon network={link.network} />
                <span>{t('social.action', { network: SOCIAL_NETWORK_LABELS[link.network] })}</span>
              </a>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
