import type { OrigenMedio } from '../../../api/medios';
import { ORIGIN_AUDIO, ORIGIN_BOTH, ORIGIN_VIDEO } from '../messages';

export interface OrigenBadgeProps {
  /** Formato derivado de las fuentes de la ficha (FR-003/FR-019). */
  origin: OrigenMedio;
}

/** Etiqueta humana de cada origen (FR-019/⟲ A2: el origen se ve en la lista). */
const LABELS: Record<OrigenMedio, string> = {
  video: ORIGIN_VIDEO,
  audio: ORIGIN_AUDIO,
  both: ORIGIN_BOTH,
};

/**
 * Origen de un episodio en la lista del panel (FR-019, ⟲ A2 de la fase 5): de
 * qué feed vino —«De YouTube» / «Del podcast» / «De YouTube y del podcast»—,
 * **siempre con palabra** (nunca solo color, FR-030) y en una píldora neutra
 * que no compite con el `StatusPill` de visibilidad ni con las marcas.
 *
 * El detalle fino de cada enlace (proveedor, título del feed y URL) vive en
 * `FuentesList`, en la misma fila.
 */
export function OrigenBadge({ origin }: OrigenBadgeProps) {
  return (
    <span className="inline-flex items-center rounded-full border border-slate-300 bg-white px-2 py-0.5 text-xs font-medium text-slate-700">
      {LABELS[origin]}
    </span>
  );
}
