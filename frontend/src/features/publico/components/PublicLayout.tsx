import { useEffect, useRef } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import type { PortadaPublica } from '../../../api/portada';
import { useSiteLinks } from '../hooks/useSiteLinks';
import { usePublicHomeData } from '../hooks/usePublicHomeData';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { useLanguage } from '../i18n/useLanguage';
import { ThemeProvider } from '../theme/ThemeProvider';
import { GradientBlobs } from './GradientBlobs';
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

/**
 * Lleva la vista a la sección del hash de la URL (`/#whatsapp`) cuando se llega
 * desde **otra ruta** (QA Bug 1: `/eventos` → «WhatsApp»). El navegador resuelve
 * el salto al cargar el documento, pero en una SPA las secciones aparecen
 * **después** de resolver la consulta pública, así que sin este efecto el
 * visitante aterriza arriba del todo.
 *
 * Se ejecuta **una sola vez por carga de documento** y solo en la portada, y
 * espera a que la respuesta pública esté resuelta (las secciones ya montadas)
 * antes de dar el ancla por resuelta: un ancla pulsada dentro de la propia
 * portada la resuelve el navegador (navegación en el mismo documento) y no se
 * repite el desplazamiento suave. `scrollIntoView` sin opciones respeta el
 * `scroll-behavior` y el `scroll-padding-top` del sistema (F10), de modo que la
 * sección no queda tapada por la cabecera fija y `prefers-reduced-motion` sigue
 * mandando.
 */
function useHashScrollOnArrival(isHome: boolean, hash: string, data: PortadaPublica | undefined) {
  const handled = useRef(false);

  useEffect(() => {
    if (handled.current || !isHome || !data) {
      return; // la portada aún no está renderizada: se reintenta cuando llegue
    }
    if (hash === '') {
      handled.current = true; // carga sin ancla: no hay nada que resolver
      return;
    }
    const target = document.getElementById(hash.slice(1));
    if (!target) {
      handled.current = true; // ancla desconocida: no se insiste
      return;
    }
    handled.current = true;
    target.scrollIntoView();
  }, [isHome, hash, data]);
}

function PublicChrome() {
  const { t } = useLanguage();
  const { data } = usePublicHomeData();
  const links = useSiteLinks();
  const { pathname, hash } = useLocation();
  useHashScrollOnArrival(pathname === '/', hash, data);

  return (
    <div className="relative isolate flex min-h-screen flex-col bg-surface font-sans text-ink">
      <GradientBlobs />

      <a
        href="#contenido"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:inline-flex focus:min-h-11 focus:items-center focus:rounded focus:bg-logo-surface focus:px-4 focus:py-2 focus:font-medium focus:text-navy focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-accent"
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
