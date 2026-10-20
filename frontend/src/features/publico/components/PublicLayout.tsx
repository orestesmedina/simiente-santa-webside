import { useEffect, useRef } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import { useSiteLinks, type SiteLink } from '../hooks/useSiteLinks';
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
 * Lleva la vista a la sección del hash de la URL (`/#whatsapp` en la portada,
 * `/eventos#actividades` en la sección de agenda).
 *
 * El navegador resuelve el salto al cargar el documento, pero en una SPA las
 * secciones aparecen **después** de resolver su consulta pública, así que sin
 * este efecto el visitante aterriza arriba del todo. El efecto se reintenta
 * cuando llegan los datos públicos (depende de `links`, que cambia con ellos:
 * portada y agenda) y da cada ancla por resuelta una sola vez por
 * ruta+fragmento, de modo que un ancla pulsada dentro de la propia página no
 * repite el desplazamiento suave del navegador.
 *
 * `scrollIntoView` sin opciones respeta el `scroll-behavior` y el
 * `scroll-padding-top` del sistema (F10): la sección no queda tapada por la
 * cabecera fija y `prefers-reduced-motion` sigue mandando.
 */
function useHashScrollOnArrival(pathname: string, hash: string, links: SiteLink[]) {
  const handled = useRef('');

  useEffect(() => {
    if (hash === '') {
      return;
    }
    const key = `${pathname}${hash}`;
    if (handled.current === key) {
      return;
    }
    const target = document.getElementById(hash.slice(1));
    if (!target) {
      return; // aún no está renderizada: se reintenta cuando lleguen los datos
    }
    handled.current = key;
    target.scrollIntoView();
  }, [pathname, hash, links]);
}

function PublicChrome() {
  const { t } = useLanguage();
  const { data } = usePublicHomeData();
  const links = useSiteLinks();
  const { pathname, hash } = useLocation();
  useHashScrollOnArrival(pathname, hash, links);

  return (
    <div className="relative flex min-h-screen flex-col bg-surface font-sans text-ink">
      {/*
        El aislamiento (contexto de apilamiento) se limita a la capa decorativa,
        no a toda la página: si el contenedor exterior aislara, el scrim del menú
        móvil (portal en `document.body`, z-30) quedaría por encima de la cabecera
        (z-50) y del panel (z-40), que están dentro de ese contexto (bug 3). El
        `isolate` propio de esta capa mantiene los blobs `-z-10` sobre el fondo
        `bg-surface` y debajo del contenido.
      */}
      <div className="isolate">
        <GradientBlobs />
      </div>

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
