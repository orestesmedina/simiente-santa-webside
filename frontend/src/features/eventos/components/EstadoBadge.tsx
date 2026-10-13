import type { PublicationState } from '../../../api/agenda';
import { StatusPill } from '../../../components/StatusPill';
import { PAST_LABEL, PAST_LEGEND } from '../messages';

export interface EstadoBadgeProps {
  state: PublicationState;
  /** `true` cuando el elemento no tiene ninguna fecha próxima (Q4). */
  isPast: boolean;
}

/**
 * Estado de publicación del elemento en el panel (ux.md §4.3, ⟲ D-8): la
 * píldora conserva sus dos valores (`Borrador`/`Publicado`) y, cuando un
 * elemento publicado ya no tiene fechas próximas, se añade una **etiqueta de
 * texto «Pasado»** con su leyenda —nunca un cuarto estado—: separa el estado de
 * publicación (lo que el equipo decidió) de la visibilidad real de hoy (lo que
 * el público ve). El estado no se toca al vencer (FR-006).
 */
export function EstadoBadge({ state, isPast }: EstadoBadgeProps) {
  return (
    <div className="flex flex-col gap-1">
      <span className="flex flex-wrap items-center gap-2">
        <StatusPill value={state} />
        {isPast && <span className="font-medium text-amber-800">{PAST_LABEL}</span>}
      </span>
      {isPast && <span className="text-sm text-slate-600">{PAST_LEGEND}</span>}
    </div>
  );
}
