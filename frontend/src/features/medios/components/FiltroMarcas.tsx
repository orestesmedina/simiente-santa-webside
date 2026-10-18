import type { MarcaRevision } from '../../../api/medios';
import { FILTER_ALL, FILTER_HELP, FILTER_PENDING, markLabel } from '../messages';

/** Filtro de vista del listado: todo el catálogo o una clase de marca pendiente. */
export type FiltroValue = 'todo' | MarcaRevision;

export interface FiltroMarcasProps {
  /** Filtro activo. */
  value: FiltroValue;
  /** Pendientes por clase de marca (sobre el catálogo completo, no sobre el filtrado). */
  counts: Record<MarcaRevision, number>;
  onChange: (value: FiltroValue) => void;
}

/** Orden de presentación de las clases (el de la tabla de ux.md §4.7). */
const MARKS: readonly MarcaRevision[] = ['possible_duplicate', 'unavailable', 'missing_from_feed'];

/** Clases base de las pastillas: palabra siempre, estado por `aria-pressed` y color. */
function pillClasses(active: boolean): string {
  return [
    'inline-flex min-h-11 items-center rounded-full border px-4 font-medium',
    'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900',
    active
      ? 'border-slate-900 bg-slate-900 text-white'
      : 'border-slate-300 bg-white text-slate-900 hover:bg-slate-100',
  ].join(' ');
}

/**
 * Filtro de marcas del listado (ux.md §4.3, ⟲ D-10): **filtro de vista**, no de
 * datos, que lleva al equipo directo a lo que necesita revisión.
 *
 * La fila se rotula con «Solo lo pendiente de revisión» y ofrece «Todo» más
 * **una pastilla por clase de marca con pendientes**, con su **conteo**
 * («Posible duplicado (2)», «No disponible (1)», «Ya no aparece en el feed
 * (3)»). Cada pastilla filtra por el parámetro **real** del contrato
 * (`reviewMark` = la marca de esa clase **pendiente**, semántica de T925): con
 * los dos ejes (P9-19), «pendiente» es un par de duplicados abierto o una marca
 * de disponibilidad sin resolver. Las clases sin pendientes no se pintan (0
 * pastillas de ruido) y el grupo entero desaparece cuando no hay ninguna.
 *
 * Es provisional (ux.md D-10): retirarlo no toca la API.
 */
export function FiltroMarcas({ value, counts, onChange }: FiltroMarcasProps) {
  const pendientes = MARKS.filter((mark) => counts[mark] > 0);
  if (pendientes.length === 0) {
    return null;
  }

  return (
    <section aria-labelledby="medios-filtro-title" className="space-y-2">
      <h2 id="medios-filtro-title" className="text-lg font-semibold text-slate-900">
        {FILTER_PENDING}
      </h2>
      <p className="text-sm text-slate-600">{FILTER_HELP}</p>
      <div role="group" aria-label={FILTER_PENDING} className="flex flex-wrap gap-2">
        <button
          type="button"
          aria-pressed={value === 'todo'}
          onClick={() => onChange('todo')}
          className={pillClasses(value === 'todo')}
        >
          {FILTER_ALL}
        </button>
        {pendientes.map((mark) => (
          <button
            key={mark}
            type="button"
            aria-pressed={value === mark}
            onClick={() => onChange(mark)}
            className={pillClasses(value === mark)}
          >
            {`${markLabel(mark)} (${counts[mark]})`}
          </button>
        ))}
      </div>
    </section>
  );
}
