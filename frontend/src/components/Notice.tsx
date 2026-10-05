import type { ReactNode } from 'react';

export type NoticeVariant = 'success' | 'error' | 'info' | 'warning';

export interface NoticeProps {
  variant?: NoticeVariant;
  children: ReactNode;
  /** Marca el aviso como temporal (p. ej. una confirmación que se auto-oculta). */
  temporal?: boolean;
  className?: string;
}

const VARIANT_CLASSES: Record<NoticeVariant, string> = {
  success: 'border-green-700 bg-green-50 text-green-900',
  error: 'border-red-700 bg-red-50 text-red-900',
  info: 'border-slate-400 bg-slate-50 text-slate-900',
  warning: 'border-amber-600 bg-amber-50 text-amber-900',
};

/**
 * Aviso de sistema con `aria-live`: `role="alert"` (assertive) para error y
 * advertencia; `role="status"` (polite) para éxito e información. Nunca
 * comunica solo con color (ux.md §6).
 */
export function Notice({
  variant = 'info',
  children,
  temporal = false,
  className = '',
}: NoticeProps) {
  const assertive = variant === 'error' || variant === 'warning';
  const classes = ['rounded border px-4 py-3', VARIANT_CLASSES[variant], className]
    .filter(Boolean)
    .join(' ');

  return (
    <div
      role={assertive ? 'alert' : 'status'}
      aria-live={assertive ? 'assertive' : 'polite'}
      data-temporal={temporal || undefined}
      className={classes}
    >
      {children}
    </div>
  );
}
