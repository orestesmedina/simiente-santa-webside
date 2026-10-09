import type { ReactNode } from 'react';

export type EmptyStateKind = 'empty' | 'loading' | 'error';

export interface EmptyStateProps {
  message: string;
  kind?: EmptyStateKind;
  /** Acción opcional (p. ej. "Reintentar" o "Crear usuario"). */
  action?: ReactNode;
  className?: string;
}

/**
 * Cubre los estados de carga, vacío y error de las vistas con datos (ux.md §4).
 * `error` se anuncia como `role="alert"`; carga y vacío como `role="status"`.
 */
export function EmptyState({ message, kind = 'empty', action, className = '' }: EmptyStateProps) {
  const classes = [
    'flex flex-col items-center gap-3 rounded border border-dashed border-slate-300 bg-slate-50 px-4 py-8 text-center',
    className,
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <div
      role={kind === 'error' ? 'alert' : 'status'}
      aria-live={kind === 'error' ? 'assertive' : 'polite'}
      aria-busy={kind === 'loading' || undefined}
      className={classes}
    >
      {kind === 'loading' && (
        <span
          aria-hidden="true"
          className="h-6 w-6 animate-spin rounded-full border-2 border-slate-400 border-t-transparent"
        />
      )}
      <p className="text-slate-700">{message}</p>
      {action}
    </div>
  );
}
