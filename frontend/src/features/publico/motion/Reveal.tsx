import { useEffect, useRef, useState, type CSSProperties, type ReactNode } from 'react';
import { observeReveal } from './useReveal';
import { prefersReducedMotion, usePrefersReducedMotion } from './usePrefersReducedMotion';

/** Dirección de entrada del revelado (en móvil se usa `up` para todo). */
export type RevealFrom = 'up' | 'left' | 'right';

export interface RevealProps {
  as?: 'div' | 'section' | 'li' | 'span';
  from?: RevealFrom;
  /** Retardo explícito en ms (independiente del escalonado de `RevealGroup`). */
  delay?: number;
  /** Índice de escalonado inyectado por `RevealGroup` (0–3). */
  revealIndex?: number;
  className?: string;
  children: ReactNode;
}

/** Clase base del envoltorio de revelado (reglas en `index.css`). */
export const REVEAL_BASE_CLASS = 'reveal-block';

/**
 * Envoltorio de revelado al hacer scroll (F10 · T1018/FR-005/FR-009).
 *
 * **Progresivo por construcción**: el contenido nace visible; el estado
 * pendiente solo lo pinta el CSS bajo `html.js` (marcador de `index.html`) y se
 * anula con `prefers-reduced-motion` (doble capa: CSS global de T1004 + hook
 * `usePrefersReducedMotion`). Con JS caído, con movimiento reducido o si el
 * observador no dispara (salvavidas de 1 s) el contenido se ve igualmente; un
 * `focusin` sobre un elemento pendiente lo revela al instante. Solo anima
 * `opacity`/`transform` (compositor): nunca bloquea lectura ni acción.
 */
export function Reveal({
  as = 'div',
  from = 'up',
  delay,
  revealIndex,
  className = '',
  children,
}: RevealProps) {
  const prefersReduced = usePrefersReducedMotion();
  const ref = useRef<HTMLDivElement | null>(null);
  const supportsObserver = typeof IntersectionObserver !== 'undefined';
  // Si el bloque ya nace mostrado (movimiento reducido o sin `IntersectionObserver`)
  // queda marcado como revelado: apagar `prefers-reduced-motion` en caliente no
  // debe volver a ocultarlo ni re-animarlo (FR-008).
  const [observed, setObserved] = useState(
    () => prefersReducedMotion() || typeof IntersectionObserver === 'undefined',
  );

  const revealed = observed || prefersReduced || !supportsObserver;

  useEffect(() => {
    if (prefersReduced || !supportsObserver) {
      return;
    }
    const element = ref.current;
    if (!element) {
      return;
    }
    return observeReveal(element, () => setObserved(true));
  }, [prefersReduced, supportsObserver]);

  const style: Record<string, string> = {};
  if (delay !== undefined) {
    style['--reveal-delay'] = `${delay}ms`;
  }
  if (revealIndex !== undefined) {
    style['--reveal-index'] = String(revealIndex);
  }

  const Component = as as 'div';
  const classes = [REVEAL_BASE_CLASS, className].filter(Boolean).join(' ');

  return (
    <Component
      ref={ref}
      data-reveal={revealed ? 'in' : from}
      className={classes}
      style={style as CSSProperties}
    >
      {children}
    </Component>
  );
}
