import { useLanguage } from '../i18n/useLanguage';
import type { MobileMenuApi } from './useMobileMenu';

export interface MobileMenuProps {
  menu: MobileMenuApi;
}

/**
 * Control del menú móvil (T1013): botón hamburguesa (`<button>` real, 44 px,
 * nombre accesible localizado y dinámico, `aria-expanded`/`aria-controls`) e
 * icono ☰→✕ **decorativo**, más el scrim que cierra al tocar fuera. Va **dentro**
 * del `<nav>` (D14: el landmark permanece visible en todos los anchos; solo su
 * lista colapsa). No contiene el selector de idioma (D2).
 */
export function MobileMenu({ menu }: MobileMenuProps) {
  const { t } = useLanguage();
  const { open, panelId, buttonRef, toggle, close } = menu;

  return (
    <>
      <button
        ref={buttonRef}
        type="button"
        aria-expanded={open}
        aria-controls={panelId}
        aria-label={open ? t('nav.menu.close') : t('nav.menu.open')}
        onClick={toggle}
        className="inline-flex min-h-11 min-w-11 items-center justify-center rounded-xl border border-line bg-surface-raised text-ink transition duration-[var(--motion-duration-micro)] ease-[var(--motion-ease-standard)] hover:bg-accent hover:text-navy focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus min-[861px]:hidden"
      >
        <span aria-hidden="true" className="text-lg leading-none">
          {open ? '✕' : '☰'}
        </span>
      </button>

      {open && (
        <div
          aria-hidden="true"
          data-testid="mobile-menu-scrim"
          onClick={() => close()}
          className="fixed inset-0 z-30 bg-black/55 min-[861px]:hidden"
        />
      )}
    </>
  );
}
