import { Outlet } from 'react-router-dom';
import { useSiteLinks } from '../hooks/useSiteLinks';
import { usePublicHomeData } from '../hooks/usePublicHomeData';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { useLanguage } from '../i18n/useLanguage';
import { ThemeProvider } from '../theme/ThemeProvider';
import { PublicFooter } from './PublicFooter';
import { SiteHeader } from './SiteHeader';

/**
 * *Chrome* del sitio público (ux.md §4.1, D-5): enlace de salto, cabecera fija
 * (`SiteHeader`: marca, navegación de anclas a las secciones con contenido
 * publicado y selector de idioma) y pie con la frase de marca. La navegación se
 * alimenta de la fuente única `useSiteLinks` (FR-012/FR-027).
 *
 * Provee el `LanguageProvider` y el `ThemeProvider` de todo el sitio público
 * (T328/T1014); el panel no cambia de idioma ni de tema (FR-016/P10-11).
 */
export function PublicLayout() {
  return (
    <LanguageProvider>
      <ThemeProvider>
        <PublicChrome />
      </ThemeProvider>
    </LanguageProvider>
  );
}

function PublicChrome() {
  const { t } = useLanguage();
  const { data } = usePublicHomeData();
  const links = useSiteLinks();

  return (
    <div className="flex min-h-screen flex-col bg-white font-sans text-navy">
      <a
        href="#contenido"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:inline-flex focus:min-h-11 focus:items-center focus:rounded focus:bg-white focus:px-4 focus:py-2 focus:font-medium focus:text-navy focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-navy"
      >
        {t('a11y.skip')}
      </a>

      <SiteHeader identity={data?.identity} links={links} />

      <main id="contenido" className="flex-1">
        <Outlet />
      </main>

      <PublicFooter
        identity={data?.identity}
        links={links}
        contact={data?.contact}
        socials={data?.socials}
        whatsapp={data?.whatsapp}
      />
    </div>
  );
}
