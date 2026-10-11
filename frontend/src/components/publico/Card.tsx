import type { ReactNode } from 'react';

/** Superficies del sistema público (FR-003): sólida o «panel de vidrio». */
export type CardVariant = 'solid' | 'glass';

export interface CardProps {
  variant?: CardVariant;
  /** Añade elevación en hover **solo** cuando la tarjeta es interactiva. */
  interactive?: boolean;
  /** `false` quita el padding de panel (p. ej. un marco de imagen a sangre). */
  padded?: boolean;
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
const BASE = 'rounded-[26px] border border-line shadow-surface';

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
  padded = true,
  as: Tag = 'div',
  className = '',
  children,
}: CardProps) {
  const classes = [
    BASE,
    padded ? 'p-panel' : '',
    VARIANTS[variant],
    interactive ? INTERACTIVE : '',
    className,
  ]
    .filter(Boolean)
    .join(' ');

  return <Tag className={classes}>{children}</Tag>;
}
