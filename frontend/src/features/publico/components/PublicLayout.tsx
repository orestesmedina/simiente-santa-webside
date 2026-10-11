import { Link, Outlet } from 'react-router-dom';
import { mediaUrl } from '../../../api/client';
import { usePublicHomeData } from '../hooks/usePublicHomeData';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { LanguageSwitcher } from '../i18n/LanguageSwitcher';
import { useLanguage } from '../i18n/useLanguage';
import { presentSections } from '../sections';
import { BrandLogo } from './BrandLogo';
import { PublicFooter } from './PublicFooter';

/**
 * *Chrome* del sitio público (ux.md §4.1, D-5): enlace de salto, cabecera
 * `navy` con logo + nombre (identidad publicada o respaldo de marca, R3-18),
 * navegación de anclas **solo** a las secciones con contenido publicado
 * (SC-012) y selector de idioma; pie con la frase de marca. Sin menú
 * hamburguesa: las anclas son visibles siempre.
 *
 * Provee el `LanguageProvider` de todo el sitio público (T328); el panel no
 * cambia de idioma (FR-016).
 */
export function PublicLayout() {
  return (
    <LanguageProvider>
      <PublicChrome />
    </LanguageProvider>
  );
}

function PublicChrome() {
  const { t } = useLanguage();
  const { data } = usePublicHomeData();
  const sections = presentSections(data);
  const brandName = data?.identity?.name ?? t('brand.name');

  return (
    <div className="flex min-h-screen flex-col bg-white font-sans text-navy">
      <a
        href="#contenido"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:inline-flex focus:min-h-11 focus:items-center focus:rounded focus:bg-white focus:px-4 focus:py-2 focus:font-medium focus:text-navy focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-navy"
      >
        {t('a11y.skip')}
      </a>

      <header className="bg-navy text-white">
        <div className="mx-auto flex w-full max-w-6xl flex-wrap items-center gap-x-6 gap-y-2 px-4 py-3 sm:px-6">
          <Link
            to="/"
            className="flex min-h-11 items-center gap-2 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
          >
            <BrandLogo
              src={mediaUrl(data?.identity?.logoUrl)}
              alt={data?.identity?.logoAlt ?? t('a11y.logo')}
              size={40}
            />
            <span className="font-display text-2xl">{brandName}</span>
          </Link>

          <nav aria-label={t('nav.label')} className="min-w-0 flex-1">
            <ul className="flex items-center gap-4 overflow-x-auto">
              {sections.map((section) => (
                <li key={section.id}>
                  <a
                    href={`#${section.anchor}`}
                    className="inline-flex min-h-11 items-center whitespace-nowrap focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
                  >
                    {t(section.messageKey)}
                  </a>
                </li>
              ))}
            </ul>
          </nav>

          <LanguageSwitcher className="text-white" />
        </div>
      </header>

      <main id="contenido" className="flex-1">
        <Outlet />
      </main>

      <PublicFooter identity={data?.identity} socials={data?.socials} />
    </div>
  );
}
