import type { SiteLink } from '../hooks/useSiteLinks';
import { useLanguage } from '../i18n/useLanguage';

export interface PublicNavProps {
  links: SiteLink[];
}

/**
 * Navegación de secciones del sitio público (T1012/FR-011/FR-012): **la única
 * lista de enlaces** dentro del `<nav aria-label>` de la cabecera (D15/C4).
 *
 * En escritorio (≥ 861 px) los enlaces se reparten en filas `flex-wrap`
 * compactas — **sin** `overflow-x-auto` (fin del «scroll incómodo») — y en móvil
 * la lista se colapsa con `display:none` bajo la media query (D16/I2): los
 * enlaces ocultos **no** son enfocables ni están en el árbol de accesibilidad,
 * pero **siguen en el DOM** (sin atributos `hidden`/`aria-hidden`/`inert`), de
 * modo que jsdom los sigue viendo y las pruebas congeladas de F3 pasan sin
 * cambios. El botón hamburguesa y el panel los añade `MobileMenu` (T1013) sobre
 * esta misma lista.
 *
 * La clase `public-nav-list` vive en `index.css` (determinista, sin depender del
 * orden de utilidades de Tailwind) y el fallback sin JS la muestra siempre como
 * filas compactas (`html:not(.js)`).
 */
export function PublicNav({ links }: PublicNavProps) {
  const { t } = useLanguage();

  return (
    <nav
      aria-label={t('nav.label')}
      className="order-2 min-w-0 flex-none min-[861px]:order-none min-[861px]:flex-1"
    >
      <ul className="public-nav-list">
        {links.map((link) => (
          <li key={link.id}>
            <a
              href={link.href}
              className="inline-flex min-h-11 items-center whitespace-nowrap font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus"
            >
              {t(link.messageKey)}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );
}
