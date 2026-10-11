import { Link } from 'react-router-dom';
import type { ContactPublic, SocialLinkPublic, WhatsappChannelPublic } from '../../../api/portada';
import type { SiteLink } from '../hooks/useSiteLinks';
import { useLanguage } from '../i18n/useLanguage';
import { SOCIAL_NETWORK_LABELS } from '../social';

export interface FooterSiteMapProps {
  /** Secciones publicadas (misma fuente única que la cabecera). */
  links: SiteLink[];
  /** Datos de contacto del bloque 04. */
  contact?: ContactPublic;
  /** Redes del bloque 05. */
  socials?: SocialLinkPublic[];
  /** Canales de WhatsApp del bloque 03 (se usa el primero). */
  whatsapp?: WhatsappChannelPublic[];
}

/** Teléfono listo para `tel:` (sin espacios ni separadores). */
function phoneHref(phone: string): string {
  return `tel:${phone.replace(/[^\d+]/g, '')}`;
}

/** Estilo compartido de los enlaces del mapa (objetivo ≥ 44 px, foco visible). */
const LINK =
  'inline-flex min-h-11 items-center text-muted transition-colors hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus';

/**
 * Rótulo de grupo: **etiqueta no-heading** (`<p>`, D18/C1). Nunca `h2`/`h3` ni
 * ningún rol `heading`: `homePage.test.tsx` resuelve encabezados por nombre a
 * cualquier nivel y «Contacto» colisionaría con el `h2` de la sección 04.
 */
const GROUP_LABEL = 'font-sans text-xs font-semibold uppercase tracking-[0.16em] text-accent';

/**
 * Mapa del sitio del pie (T1015/P10-14/R10-15/FR-027): `<nav>` con **nombre
 * accesible propio** (distinto del de la cabecera) y tres grupos condicionales
 * —Explorar / Contacto / Síguenos—.
 *
 * Reglas inviolables contra las pruebas congeladas:
 *  · **Un grupo sin ítems no se renderiza** y, si ningún grupo tiene ítems, el
 *    mapa entero **no se dibuja** (D18/C1).
 *  · **Rótulos de grupo sin rol `heading`** (D18/C1).
 *  · **Nombres accesibles únicos** (D19/C2): los enlaces de Contacto llevan
 *    prefijo de contexto («Contacto: Correo»…), la **dirección es texto** y las
 *    redes usan etiquetas **planas** (`SOCIAL_NETWORK_LABELS`), nunca
 *    `social.action` («Ver en …»).
 *  · **Sin enlaces rotos** (FR-027): solo destinos existentes (secciones
 *    publicadas + la portada). Para F4–F9 basta añadir un grupo cuando la
 *    sección exista; el componente no inventa enlaces.
 */
export function FooterSiteMap({ links, contact, socials = [], whatsapp = [] }: FooterSiteMapProps) {
  const { t } = useLanguage();

  const contactPrefix = t('footer.sitemap.contact');
  const whatsappChannel = whatsapp[0];

  const hasExplore = links.length > 0;
  const hasContact = Boolean(contact) || Boolean(whatsappChannel);
  const hasSocial = socials.length > 0;

  if (!hasExplore && !hasContact && !hasSocial) {
    return null;
  }

  return (
    <nav aria-label={t('footer.sitemap.label')} className="grid gap-block min-[861px]:grid-cols-3">
      {hasExplore && (
        <div>
          <p className={GROUP_LABEL}>{t('footer.sitemap.explore')}</p>
          <ul className="mt-fluid space-y-1">
            <li>
              <Link to="/" className={LINK}>
                {t('footer.sitemap.home')}
              </Link>
            </li>
            {links.map((link) => (
              <li key={link.id}>
                <a href={link.href} className={LINK}>
                  {t(link.messageKey)}
                </a>
              </li>
            ))}
          </ul>
        </div>
      )}

      {hasContact && (
        <div>
          <p className={GROUP_LABEL}>{contactPrefix}</p>
          <ul className="mt-fluid space-y-1">
            {contact?.address && (
              <li>
                <span className="inline-flex min-h-11 items-center text-muted">
                  {t('contact.address')}: {contact.address}
                </span>
              </li>
            )}
            {contact?.email && (
              <li>
                <a
                  href={`mailto:${contact.email}`}
                  aria-label={`${contactPrefix}: ${t('contact.email')}`}
                  className={LINK}
                >
                  {t('contact.email')}
                </a>
              </li>
            )}
            {contact?.phone && (
              <li>
                <a
                  href={phoneHref(contact.phone)}
                  aria-label={`${contactPrefix}: ${t('contact.phone')}`}
                  className={LINK}
                >
                  {t('contact.phone')}
                </a>
              </li>
            )}
            {whatsappChannel && (
              <li>
                <a
                  href={whatsappChannel.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  aria-label={`${contactPrefix}: ${t('nav.sections.whatsapp')}`}
                  className={LINK}
                >
                  {t('nav.sections.whatsapp')}
                </a>
              </li>
            )}
          </ul>
        </div>
      )}

      {hasSocial && (
        <div>
          <p className={GROUP_LABEL}>{t('footer.sitemap.social')}</p>
          <ul className="mt-fluid space-y-1">
            {socials.map((link) => (
              <li key={link.id}>
                <a href={link.url} target="_blank" rel="noopener noreferrer" className={LINK}>
                  {SOCIAL_NETWORK_LABELS[link.network]}
                </a>
              </li>
            ))}
          </ul>
        </div>
      )}
    </nav>
  );
}
