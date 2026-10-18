import type { FuenteAdmin } from '../../../api/medios';
import { providerLabel } from '../messages';

export interface FuentesListProps {
  /** Procedencia de cada enlace de la ficha (R9-11), tal como la trae el servidor. */
  sources: FuenteAdmin[];
}

/**
 * Detalle de las fuentes de una ficha en la lista del panel (FR-019, ⟲ A2 de la
 * fase 5): por cada enlace importado, el **proveedor** (dicho como lo llama la
 * iglesia: «YouTube» / «el podcast»), el **título del feed** de donde vino (que
 * al unir conserva el del podcast como dato de origen) y el **enlace** —en una
 * pestaña nueva y sin acceso al opener—.
 *
 * Es lo que el equipo revisa antes de unir dos entradas: si los dos títulos del
 * feed y las dos fechas se parecen, la unión tiene sentido.
 */
export function FuentesList({ sources }: FuentesListProps) {
  if (sources.length === 0) {
    return null;
  }

  return (
    <ul className="mt-1 space-y-1 text-xs text-slate-600">
      {sources.map((source) => (
        <li key={`${source.provider}-${source.externalId}`}>
          <span className="font-medium">{providerLabel(source.provider)}</span>
          {': '}
          <span>{source.feedTitle}</span>
          {' · '}
          <a
            href={source.url}
            target="_blank"
            rel="noopener noreferrer"
            className="break-all underline underline-offset-2 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900"
          >
            {source.url}
          </a>
        </li>
      ))}
    </ul>
  );
}
