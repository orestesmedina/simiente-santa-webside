import type { MarcaRevision, Proveedor } from '../../../api/medios';
import { StatusPill, type StatusPillValue } from '../../../components/StatusPill';
import { markExplainer, markLabel } from '../messages';

export interface MarcaDeRevisionProps {
  /** Marca **pendiente** de resolver que tiene la ficha. */
  mark: MarcaRevision;
  /** Servicio del que salió el episodio (para el explicador de «ya no aparece»). */
  provider: Proveedor;
  /**
   * `true` si la comprobación de disponibilidad confirmó que el enlace sigue
   * funcionando (`availabilityResult === 'available'`): cambia la línea del
   * explicador de «ya no aparece en el feed».
   */
  linkAvailable?: boolean;
}

/**
 * Tono de cada marca (ux.md §1/§4.7): ámbar para lo que hay que revisar, rojo
 * para lo que está fuera del público y grafito para el aviso inocuo. El color
 * **siempre** acompaña a la palabra (FR-030: nunca solo color).
 */
const TONES: Record<MarcaRevision, StatusPillValue> = {
  possible_duplicate: 'not-completed',
  unavailable: 'failure',
  missing_from_feed: 'draft',
};

/**
 * Marca de revisión de una fila del panel (ux.md §4.7, FR-030): **píldora con
 * palabra** (`StatusPill` con la etiqueta del catálogo) más su **explicador de
 * una línea**, que dice qué significa y —en «ya no aparece en el feed»— **nombra
 * la incertidumbre** en vez de esconderla.
 *
 * Las marcas viven en **dos ejes independientes** (P9-19): una fila puede
 * mostrar **dos** píldoras a la vez (un par de duplicados abierto **y** una marca
 * de disponibilidad) y cada una se resuelve por su cuenta.
 *
 * Los iconos por marca (⟲ D-30) **no** se implementan: la píldora con palabra y
 * su explicador ya cumplen «no solo color».
 */
export function MarcaDeRevision({ mark, provider, linkAvailable }: MarcaDeRevisionProps) {
  return (
    <div className="flex flex-col gap-1">
      <div>
        <StatusPill value={TONES[mark]} label={markLabel(mark)} />
      </div>
      <p className="text-xs text-slate-600">{markExplainer(mark, provider, linkAvailable)}</p>
    </div>
  );
}
