import { useEffect, useRef, useState, type RefObject } from 'react';

export interface ScrolledState {
  /** Referencia a la sentinela situada al principio de la página. */
  sentinelRef: RefObject<HTMLDivElement | null>;
  /** `true` cuando la sentinela ha salido del viewport (el visitante bajó). */
  scrolled: boolean;
}

/**
 * Estado «cabecera desplazada» (F10 · T1020/P10-12/R10-13) mediante una
 * **sentinela observada** con `IntersectionObserver` — **sin** listener de
 * `scroll`: cuando la sentinela (al principio de la página) deja de ser visible,
 * el visitante ha bajado. La cabecera transiciona entonces **solo su
 * superficie** (`data-scrolled`: fondo de vidrio + borde turquesa + sombra), sin
 * cambiar su geometría (D8: la altura se mantiene, así que las anclas y el
 * `scroll-padding-top` siguen siendo estables). Con movimiento reducido el
 * cambio es instantáneo (lo anula el bloque global de `index.css`).
 */
export function useScrolled(): ScrolledState {
  const sentinelRef = useRef<HTMLDivElement | null>(null);
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const element = sentinelRef.current;
    if (!element || typeof IntersectionObserver === 'undefined') {
      return;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (entry) {
          setScrolled(!entry.isIntersecting);
        }
      },
      { threshold: 0 },
    );
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  return { sentinelRef, scrolled };
}
