import type { IdentityPublic, SocialLinkPublic } from '../../../api/portada';
import { LanguageSwitcher } from '../i18n/LanguageSwitcher';
import { useLanguage } from '../i18n/useLanguage';
import { SOCIAL_NETWORK_LABELS } from '../social';
import { BrandLogo } from './BrandLogo';
import { SocialIcon } from './SocialIcon';

export interface PublicFooterProps {
  /** Identidad publicada; si no hay, se usa el respaldo de marca (R3-18). */
  identity?: IdentityPublic;
  socials?: SocialLinkPublic[];
}

/**
 * Pie del sitio público (ux.md §4.1): logo y nombre (identidad publicada o
 * respaldo estático de marca), redes compactas, la frase de voz de marca
 * `footer.welcome` (**cadena de interfaz**, analyze I4: no es contenido
 * editable) y el selector de idioma.
 */
export function PublicFooter({ identity, socials }: PublicFooterProps) {
  const { t } = useLanguage();
  const brandName = identity?.name ?? t('brand.name');

  return (
    <footer className="bg-navy text-white">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-10 sm:px-6">
        <div className="flex items-center gap-3">
          <BrandLogo src={identity?.logoUrl} alt={identity?.logoAlt ?? t('a11y.logo')} size={40} />
          <span className="font-display text-2xl">{brandName}</span>
        </div>

        {socials && socials.length > 0 && (
          <ul className="flex flex-wrap gap-x-6 gap-y-2">
            {socials.map((link) => (
              <li key={link.id}>
                <a
                  href={link.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex min-h-11 items-center gap-2 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
                >
                  <SocialIcon network={link.network} />
                  <span>{SOCIAL_NETWORK_LABELS[link.network]}</span>
                </a>
              </li>
            ))}
          </ul>
        )}

        <p className="font-emotiva text-xl italic">{t('footer.welcome')}</p>

        <LanguageSwitcher className="text-white" />
      </div>
    </footer>
  );
}
