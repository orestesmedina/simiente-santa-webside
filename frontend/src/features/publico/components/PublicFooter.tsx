import { mediaUrl } from '../../../api/client';
import type {
  ContactPublic,
  IdentityPublic,
  SocialLinkPublic,
  WhatsappChannelPublic,
} from '../../../api/portada';
import type { SiteLink } from '../hooks/useSiteLinks';
import { LanguageSwitcher } from '../i18n/LanguageSwitcher';
import { useLanguage } from '../i18n/useLanguage';
import { BrandLogo } from './BrandLogo';
import { FooterSiteMap } from './FooterSiteMap';

export interface PublicFooterProps {
  /** Identidad publicada; si no hay, se usa el respaldo de marca (R3-18). */
  identity?: IdentityPublic;
  /** Secciones publicadas (misma fuente única que la cabecera). */
  links: SiteLink[];
  /** Datos de contacto (grupo Contacto del mapa). */
  contact?: ContactPublic;
  /** Redes (grupo Síguenos del mapa). */
  socials?: SocialLinkPublic[];
  /** Canales de WhatsApp (enlace «Contacto: WhatsApp» del mapa). */
  whatsapp?: WhatsappChannelPublic[];
}

/**
 * Pie del sitio público con la propuesta B «Marea Turquesa» (T1016/FR-027):
 * marca + logotipo (con sus colores originales sobre `--logo-surface`),
 * **mapa del sitio** (`FooterSiteMap`: Explorar / Contacto / Síguenos) y franja
 * inferior con el lema (`footer.welcome`) y el selector de idioma
 * (comportamiento F3 intacto, US4 esc. 5).
 *
 * El `<footer>`/landmark `contentinfo` de F3 se conserva tal cual (R10-9) y la
 * **marca del pie es texto, no enlace** (D20/C3): el enlace cuyo texto incluye
 * «Simiente» lo garantiza la cabecera (`SiteHeader`) y así el pie no añade
 * paradas al ciclo de tabulación congelado (I3/RG10-21).
 */
export function PublicFooter({ identity, links, contact, socials, whatsapp }: PublicFooterProps) {
  const { t } = useLanguage();
  const brandName = identity?.name ?? t('brand.name');

  return (
    <footer className="bg-surface text-ink">
      <div className="mx-auto grid w-full max-w-6xl gap-block px-4 py-section-tight sm:px-6 min-[861px]:grid-cols-[1.3fr_2fr]">
        <div className="flex items-center gap-3">
          <BrandLogo
            src={mediaUrl(identity?.logoUrl)}
            alt={identity?.logoAlt ?? t('a11y.logo')}
            size={40}
            className="rounded-full bg-logo-surface p-0.5"
          />
          <span className="font-display text-2xl">{brandName}</span>
        </div>

        <FooterSiteMap links={links} contact={contact} socials={socials} whatsapp={whatsapp} />
      </div>

      <div className="mx-auto flex w-full max-w-6xl flex-wrap items-center gap-4 border-t border-line px-4 py-4 sm:px-6">
        <p className="font-emotiva text-base italic text-muted">{t('footer.welcome')}</p>
        <LanguageSwitcher className="ml-auto text-ink" />
      </div>
    </footer>
  );
}
