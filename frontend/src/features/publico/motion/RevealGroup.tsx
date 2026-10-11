import { Children, cloneElement, isValidElement, type ReactNode } from 'react';
import type { RevealProps } from './Reveal';

export interface RevealGroupProps {
  className?: string;
  children: ReactNode;
}

/** Tope de saltos de escalonado (D5/D21): 3 × 80 ms = 240 ms máximo. */
export const MAX_REVEAL_INDEX = 3;

/**
 * Entrada escalonada acotada (F10 · T1019/FR-005/SC-014): reparte
 * `--reveal-index` (0…3) entre sus hijos `Reveal` para que el retardo de
 * `transition-delay` crezca con el token `--motion-stagger-step` (80 ms) y se
 * **detenga** en 3 saltos (240 ms), de modo que la secuencia total de un bloque
 * nunca supere el techo de 800 ms. Con movimiento reducido, el bloque global de
 * `index.css` anula los retardos (entrada instantánea).
 *
 * «Cuando convenga»: grupos de 2–4 elementos (tarjetas de una fila, CTAs,
 * chips), no listas enteras.
 */
export function RevealGroup({ className = '', children }: RevealGroupProps) {
  const items = Children.toArray(children);

  return (
    <div className={['reveal-group', className].filter(Boolean).join(' ')}>
      {items.map((child, index) => {
        if (!isValidElement<Pick<RevealProps, 'revealIndex'>>(child)) {
          return child;
        }
        return cloneElement(child, {
          revealIndex: Math.min(index, MAX_REVEAL_INDEX),
        });
      })}
    </div>
  );
}
