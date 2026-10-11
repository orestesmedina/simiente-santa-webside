import type { ReactNode } from 'react';

/** Superficies del sistema público (FR-003): sólida o «panel de vidrio». */
export type CardVariant = 'solid' | 'glass';

export interface CardProps {
  variant?: CardVariant;
  /** Añade elevación en hover **solo** cuando la tarjeta es interactiva. */
  interactive?: boolean;
  as?: 'div' | 'section' | 'article' | 'li';
  className?: string;
  children: ReactNode;
}

/**
 * Superficie de contenido del sistema público (T1010): radio, borde, sombra y
 * fondo por tokens semánticos de tema (ambos temas correctos por construcción) y
 * padding generoso con `--pad-panel`. Es la **única** superficie de contenido del
 * sitio público (SC-011) y la heredan las secciones de F4–F9.
 *
 * La microinteracción de elevación (solo `transform`/borde, nunca layout) se
 * aplica **únicamente** si la tarjeta es interactiva; nunca sobre imágenes de
 * marca (FR-002).
 */
const BASE = 'rounded-[26px] border border-line p-panel shadow-surface';

const VARIANTS: Record<CardVariant, string> = {
  solid: 'bg-surface-raised',
  glass: 'bg-surface-raised backdrop-blur-md',
};

const INTERACTIVE =
  'transition duration-[var(--motion-duration-micro)] ease-[var(--motion-ease-standard)] ' +
  'hover:-translate-y-1.5 hover:border-accent';

export function Card({
  variant = 'solid',
  interactive = false,
  as: Tag = 'div',
  className = '',
  children,
}: CardProps) {
  const classes = [BASE, VARIANTS[variant], interactive ? INTERACTIVE : '', className]
    .filter(Boolean)
    .join(' ');

  return <Tag className={classes}>{children}</Tag>;
}
