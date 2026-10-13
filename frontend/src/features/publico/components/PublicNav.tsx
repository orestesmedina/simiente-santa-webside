import { useLocation } from 'react-router-dom';
import type { SiteLink } from '../hooks/useSiteLinks';
import { useLanguage } from '../i18n/useLanguage';
import { MobileMenu } from './MobileMenu';
import { useMobileMenu } from './useMobileMenu';

export interface PublicNavProps {
  links: SiteLink[];
}

/**
 * Ruta activa del router (QA Bug 2). `PublicNav` también se monta fuera de un
 * Router —las pruebas de presentación de F10 renderizan el componente suelto—,
 * así que la ausencia de contexto no es un error: sin ruta conocida, ninguna
 * opción de ruta se marca como actual.
 */
function useActivePath(): string | undefined {
  try {
    return useLocation().pathname;
  } catch {
    return undefined;
  }
}

/**
 * Clases base del enlace de la barra: objetivo táctil ≥ 44 px, foco visible con
 * el anillo del sistema y **estado `hover`** con el par acento/azul marino que
 * el propio sistema certifica AA en ambos temas (`hover:bg-accent` +
 * `hover:text-navy`, el mismo de la hamburguesa y del interruptor de tema).
 */
const LINK_BASE = [
  'inline-flex min-h-11 items-center whitespace-nowrap rounded px-2 font-medium',
  'transition-colors duration-[var(--motion-duration-micro)] ease-[var(--motion-ease-standard)]',
  'hover:bg-accent hover:text-navy',
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus',
].join(' ');

/**
 * Opción actual (`aria-current="page"`, QA Bug 2): subrayado con el acento
 * fuerte del tema —el turquesa que sí cumple AA como texto en claro y oscuro—
 * además del `aria-current` que lo anuncia a los lectores de pantalla.
 */
const LINK_CURRENT = [
  'text-accent-strong underline decoration-accent-strong decoration-2 underline-offset-4',
].join(' ');

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
  const activePath = useActivePath();

  /**
   * `true` si el enlace corresponde a la ruta actual. Los anclas de portada
   * (`/#ancla`) nunca son «la página actual»: su destino es la portada, no la
   * ruta en la que se está.
   */
  const isCurrent = (href: string): boolean => {
    if (!activePath) {
      return false;
    }
    const [path] = href.split('#');
    return path !== '/' && path === activePath;
  };

  return (
    <nav
      aria-label={t('nav.label')}
      className="order-2 min-w-0 flex-none min-[861px]:order-none min-[861px]:flex-1"
    >
      <MobileMenu menu={menu} />

      <ul id={panelId} ref={panelRef} data-open={open} className="public-nav-list">
        {links.map((link) => {
          const current = isCurrent(link.href);
          return (
            <li key={link.id}>
              <a
                href={link.href}
                onClick={() => close(false)}
                aria-current={current ? 'page' : undefined}
                className={current ? `${LINK_BASE} ${LINK_CURRENT}` : LINK_BASE}
              >
                {t(link.messageKey)}
              </a>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
