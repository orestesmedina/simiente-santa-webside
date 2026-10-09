import type { ButtonHTMLAttributes, ReactNode } from 'react';

export type ButtonVariant = 'primary' | 'secondary' | 'accent' | 'danger' | 'link';
export type ButtonSize = 'md' | 'sm';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  /** Muestra un indicador y deshabilita el botón mientras hay una operación. */
  loading?: boolean;
  children: ReactNode;
}

const VARIANT_CLASSES: Record<ButtonVariant, string> = {
  primary: 'border border-transparent bg-slate-900 text-white hover:bg-slate-700',
  secondary: 'border border-slate-300 bg-white text-slate-900 hover:bg-slate-100',
  accent: 'border border-transparent bg-teal text-navy hover:bg-teal-strong hover:text-white',
  danger: 'border border-transparent bg-red-700 text-white hover:bg-red-800',
  link: 'border border-transparent bg-transparent px-2 text-slate-900 underline underline-offset-4 hover:text-slate-700',
};

const SIZE_CLASSES: Record<ButtonSize, string> = {
  md: 'min-h-11 px-4 py-2',
  sm: 'min-h-11 px-3 py-2 text-sm',
};

/**
 * Botón del panel. Todas las variantes cumplen el objetivo táctil mínimo de
 * 44 px (ux.md §0), tienen foco visible y la variante `link` conserva semántica
 * de botón real (nunca un `a` que parece botón).
 */
export function Button({
  variant = 'primary',
  size = 'md',
  loading = false,
  disabled,
  type = 'button',
  className = '',
  children,
  ...rest
}: ButtonProps) {
  const classes = [
    'inline-flex items-center justify-center gap-2 rounded font-medium',
    'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900',
    'disabled:cursor-not-allowed disabled:opacity-60',
    VARIANT_CLASSES[variant],
    SIZE_CLASSES[size],
    className,
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <button
      type={type}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      className={classes}
      {...rest}
    >
      {loading && (
        <span
          aria-hidden="true"
          className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
        />
      )}
      {children}
    </button>
  );
}
