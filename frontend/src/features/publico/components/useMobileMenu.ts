import { useCallback, useEffect, useId, useRef, useState, type RefObject } from 'react';

/** Corte de dispositivo: móvil ≤ 860 px / escritorio ≥ 861 px (P10-13/D4). */
const DESKTOP_QUERY = '(min-width: 861px)';

export interface MobileMenuApi {
  /** Panel abierto. */
  open: boolean;
  /** `id` de la lista/panel (destino de `aria-controls`). */
  panelId: string;
  /** Referencia a la lista de enlaces (la única del `<nav>`). */
  panelRef: RefObject<HTMLUListElement | null>;
  /** Referencia al botón hamburguesa. */
  buttonRef: RefObject<HTMLButtonElement | null>;
  /** Abre/cierra el panel. */
  toggle: () => void;
  /** Cierra el panel; `restoreFocus` devuelve el foco al botón. */
  close: (restoreFocus?: boolean) => void;
}

/**
 * Lógica del menú hamburguesa accesible (T1013/D17/I1): **patrón *disclosure*
 * de las guías APG, sin diálogo modal**. El botón controla **la misma lista de
 * enlaces** del `<nav>` (D15/C4: nunca se monta una segunda lista); al abrir el
 * foco entra en el primer enlace y `Escape`/scrim/botón cierran devolviendo el
 * foco al botón. **No hay trampa de foco** (`Tab`/`Shift+Tab` pueden salir del
 * panel hacia la barra de controles y el resto de la página) y **no** se usa
 * `role="dialog"` ni `aria-modal`. El scroll de fondo se bloquea con
 * `html[data-nav-open='true']` y al cruzar a escritorio el panel se cierra solo.
 */
export function useMobileMenu(): MobileMenuApi {
  const [open, setOpen] = useState(false);
  const panelId = useId();
  const panelRef = useRef<HTMLUListElement | null>(null);
  const buttonRef = useRef<HTMLButtonElement | null>(null);

  const close = useCallback((restoreFocus = true) => {
    setOpen(false);
    if (restoreFocus) {
      buttonRef.current?.focus();
    }
  }, []);

  const toggle = useCallback(() => {
    setOpen((current) => !current);
  }, []);

  // Bloqueo del scroll de fondo mientras el panel está abierto.
  useEffect(() => {
    if (!open) {
      return;
    }
    document.documentElement.dataset.navOpen = 'true';
    return () => {
      delete document.documentElement.dataset.navOpen;
    };
  }, [open]);

  // Al abrir, el foco entra en el primer enlace de la lista (disclosure APG).
  useEffect(() => {
    if (!open) {
      return;
    }
    panelRef.current?.querySelector<HTMLElement>('a')?.focus();
  }, [open]);

  // `Escape` cierra y devuelve el foco al botón.
  useEffect(() => {
    if (!open) {
      return;
    }
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        close(true);
      }
    };
    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [open, close]);

  // Cruce de breakpoint: al pasar a escritorio el panel se cierra solo.
  useEffect(() => {
    const media = window.matchMedia?.(DESKTOP_QUERY);
    if (!media?.addEventListener) {
      return;
    }
    const onChange = (event: MediaQueryListEvent) => {
      if (event.matches) {
        setOpen(false);
      }
    };
    media.addEventListener('change', onChange);
    return () => media.removeEventListener('change', onChange);
  }, []);

  return { open, panelId, panelRef, buttonRef, toggle, close };
}
