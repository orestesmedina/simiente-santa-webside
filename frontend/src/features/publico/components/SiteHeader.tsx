import type { ReactNode, RefObject } from 'react';
import { useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';
import { mediaUrl } from '../../../api/client';
import type { IdentityPublic } from '../../../api/portada';
import type { SiteLink } from '../hooks/useSiteLinks';
import { LanguageSwitcher } from '../i18n/LanguageSwitcher';
import { useLanguage } from '../i18n/useLanguage';
import { useScrolled } from '../motion/useScrolled';
import { ThemeToggle } from '../theme/ThemeToggle';
import { BrandLogo } from './BrandLogo';
import { PublicNav } from './PublicNav';

export interface SiteHeaderProps {
  /** Identidad publicada (marca y logotipo); si no hay, respaldo de marca. */
  identity?: IdentityPublic;
  /** Enlaces de sección (fuente única `useSiteLinks`). */
  links: SiteLink[];
  /** Slot del control de tema de la cabecera (lo monta T1028). */
  actions?: ReactNode;
}

/**
 * Mide la altura real de la cabecera y la publica como `--header-offset` en
 * `<html>` (T1014/FR-014/RG10-3): el `scroll-padding-top` de `index.css` la usa
 * para que un ancla nunca quede tapada por la barra fija. Se re-mide con
 * `ResizeObserver` (sin listener de `scroll`); el panel del menú móvil es un
 * overlay `position: fixed` y **no** altera la altura de la cabecera, de modo
 * que el valor es estable con el menú abierto o cerrado.
 */
function useHeaderOffset(ref: RefObject<HTMLElement | null>) {
  useEffect(() => {
    const element = ref.current;
    if (!element) {
      return;
    }
    const apply = () => {
      const height = element.getBoundingClientRect().height;
      document.documentElement.style.setProperty('--header-offset', `${height}px`);
    };
    apply();

    if (typeof ResizeObserver === 'undefined') {
      return () => document.documentElement.style.removeProperty('--header-offset');
    }
    const observer = new ResizeObserver(apply);
    observer.observe(element);
    return () => {
      observer.disconnect();
      document.documentElement.style.removeProperty('--header-offset');
    };
  }, [ref]);
}

/**
 * Cabecera fija del sitio público (T1011/FR-010/FR-014/FR-025): fila de marca
 * (logotipo con sus colores originales sobre `--logo-surface` + nombre completo
 * de la identidad como **enlace a `/`**), `<nav aria-label>` de secciones y
 * clúster de controles (selector de idioma + slot de tema).
 *
 * El `<nav>` **permanece visible en todos los anchos** (D14): solo su lista de
 * anclas colapsa en móvil (lo resuelve `PublicNav`, T1012/T1013). A < 480 px la
 * palabra de la marca se oculta **solo visualmente** con `sr-only` (nunca
 * `display:none`): el enlace conserva «Simiente» en su nombre accesible y en su
 * `innerText` (D20/C3, `tabCycle` congelado a 320 px).
 */
export function SiteHeader({ identity, links, actions }: SiteHeaderProps) {
  const { t } = useLanguage();
  const brandName = identity?.name ?? t('brand.name');
  const headerRef = useRef<HTMLElement | null>(null);
  const { sentinelRef, scrolled } = useScrolled();
  useHeaderOffset(headerRef);

  return (
    <>
      {/* Sentinela al principio de la página: al salir del viewport, la
          cabecera pasa a `data-scrolled` (superficie de vidrio, sin cambio de
          geometría). Observada con IntersectionObserver, sin listener de scroll. */}
      <div ref={sentinelRef} aria-hidden="true" className="h-px w-full" />

      <header
        ref={headerRef}
        data-scrolled={scrolled ? 'true' : 'false'}
        className="public-header sticky top-0 z-50 border-b border-line bg-surface text-ink"
      >
        <div className="mx-auto flex w-full max-w-6xl flex-wrap items-center gap-x-4 gap-y-2 px-4 py-3 sm:px-6">
          <Link
            to="/"
            className="flex min-h-11 items-center gap-2 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus"
          >
            <BrandLogo
              src={mediaUrl(identity?.logoUrl)}
              alt={identity?.logoAlt ?? t('a11y.logo')}
              size={40}
              className="rounded-full bg-logo-surface p-0.5"
            />
            <span className="font-display text-2xl max-[480px]:sr-only">{brandName}</span>
          </Link>

          <PublicNav links={links} />

          <div className="ml-auto flex items-center gap-2">
            <LanguageSwitcher className="text-ink" />
            {actions ?? <ThemeToggle />}
          </div>
        </div>
      </header>
    </>
  );
}
