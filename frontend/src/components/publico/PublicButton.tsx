import type { ReactNode } from 'react';

/** Variantes de acción del sitio público (FR-003): un solo estilo por función. */
export type PublicButtonVariant = 'primary' | 'accent' | 'secondary';

export interface PublicButtonProps {
  variant?: PublicButtonVariant;
  /** Si se indica, el control se renderiza como enlace `<a>`. */
  href?: string;
  /** Abre el enlace en pestaña nueva sin acceso al opener (`target`/`rel`). */
  external?: boolean;
  onClick?: () => void;
  type?: 'button' | 'submit';
  className?: string;
  children: ReactNode;
}

/**
 * Clases base del botón de acción del sistema público (FR-003/FR-004/FR-008):
 * objetivo táctil ≥ 44 px (`min-h-11`), tipografía Poppins, foco visible con el
 * anillo turquesa y microinteracción de `--motion-duration-micro`. La
 * microinteracción mueve **solo** `transform`/`opacity`/colores (nunca
 * propiedades de layout): el maquetado no se desplaza.
 */
const BASE =
  'inline-flex min-h-11 items-center justify-center gap-2 rounded-full px-6 py-2 font-semibold ' +
  'transition duration-[var(--motion-duration-micro)] ease-[var(--motion-ease-standard)] ' +
  'hover:-translate-y-0.5 active:scale-[1.02] ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus';

const VARIANTS: Record<PublicButtonVariant, string> = {
  primary: 'bg-accent text-navy hover:bg-accent-strong hover:text-white',
  // Coral solo como detalle gráfico (nunca como texto de marca).
  accent: 'bg-coral text-navy',
  secondary: 'border-2 border-current bg-transparent text-ink hover:bg-surface-raised',
};

function classesFor(variant: PublicButtonVariant, className: string): string {
  return [BASE, VARIANTS[variant], className].filter(Boolean).join(' ');
}

/**
 * Botón/enlace de acción del sistema visual público (T1008): es el **único**
 * estilo de botón de acción del sitio (SC-011) y lo heredan las secciones de
 * F4–F9. Nunca aplica efectos sobre el logotipo (FR-002).
 */
export function PublicButton({
  variant = 'primary',
  href,
  external = false,
  onClick,
  type = 'button',
  className = '',
  children,
}: PublicButtonProps) {
  const classes = classesFor(variant, className);

  if (href) {
    return (
      <a
        href={href}
        className={classes}
        onClick={onClick}
        {...(external ? { target: '_blank', rel: 'noopener noreferrer' } : {})}
      >
        {children}
      </a>
    );
  }

  return (
    <button type={type} className={classes} onClick={onClick}>
      {children}
    </button>
  );
}
