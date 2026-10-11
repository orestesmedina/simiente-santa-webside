export type StatusPillValue =
  | 'active'
  | 'inactive'
  | 'draft'
  | 'published'
  | 'success'
  | 'failure'
  | 'completed'
  | 'not-completed'
  | 'denied';

export interface StatusPillProps {
  value: StatusPillValue;
  /** Sobrescribe el texto por defecto (p. ej. una etiqueta del contrato). */
  label?: string;
}

const DEFAULT_LABELS: Record<StatusPillValue, string> = {
  active: 'Activo',
  inactive: 'Inactivo',
  draft: 'Borrador',
  published: 'Publicado',
  success: 'Exitoso',
  failure: 'Fallido',
  completed: 'Completada',
  'not-completed': 'No completada',
  denied: 'Denegada',
};

const TONE_CLASSES: Record<StatusPillValue, string> = {
  active: 'bg-green-100 text-green-900 border border-green-700',
  inactive: 'bg-slate-100 text-slate-900 border border-slate-400',
  draft: 'bg-slate-100 text-slate-900 border border-slate-400',
  published: 'bg-green-100 text-green-900 border border-green-700',
  success: 'bg-green-100 text-green-900 border border-green-700',
  failure: 'bg-red-100 text-red-900 border border-red-700',
  completed: 'bg-green-100 text-green-900 border border-green-700',
  'not-completed': 'bg-amber-100 text-amber-900 border border-amber-700',
  denied: 'bg-red-100 text-red-900 border border-red-700',
};

/**
 * Píldora de estado con color **y** texto explícito (ux.md §6: nunca solo
 * color). Se usa para el estado de la cuenta (`active`/`inactive`) y para el
 * resultado de los registros de auditoría.
 */
export function StatusPill({ value, label }: StatusPillProps) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-3 py-1 text-sm font-medium ${TONE_CLASSES[value]}`}
    >
      {label ?? DEFAULT_LABELS[value]}
    </span>
  );
}
