import type { SiteLink } from '../hooks/useSiteLinks';
import { useLanguage } from '../i18n/useLanguage';
import { MobileMenu } from './MobileMenu';
import { useMobileMenu } from './useMobileMenu';

export interface PublicNavProps {
  links: SiteLink[];
}

/**
 * Navegación de secciones del sitio público (T1012/T1013/FR-011/FR-012): **la
 * única lista de enlaces** dentro del `<nav aria-label>` de la cabecera
 * (D15/C4).
 *
 * En escritorio (≥ 861 px) los enlaces se reparten en filas `flex-wrap`
 * compactas — **sin** `overflow-x-auto` (fin del «scroll incómodo») — y en móvil
 * (≤ 860 px) el botón hamburguesa de `MobileMenu` controla **esa misma lista**
 * como panel (no hay una segunda lista): cerrada queda con `display:none` (fuera
 * del orden de tabulación y del árbol de accesibilidad, D16/I2) pero **sigue en
 * el DOM** (sin atributos `hidden`/`aria-hidden`/`inert`), de modo que jsdom la
 * sigue viendo y las pruebas congeladas de F3 pasan sin cambios.
 *
 * La clase `public-nav-list` vive en `index.css` (determinista, sin depender del
 * orden de utilidades de Tailwind) y el fallback sin JS la muestra siempre como
 * filas compactas (`html:not(.js)`).
 */
export function PublicNav({ links }: PublicNavProps) {
  const { t } = useLanguage();
  const menu = useMobileMenu();
  const { open, panelId, panelRef, close } = menu;

  return (
    <nav
      aria-label={t('nav.label')}
      className="order-2 min-w-0 flex-none min-[861px]:order-none min-[861px]:flex-1"
    >
      <MobileMenu menu={menu} />

      <ul id={panelId} ref={panelRef} data-open={open} className="public-nav-list">
        {links.map((link) => (
          <li key={link.id}>
            <a
              href={link.href}
              onClick={() => close(false)}
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
