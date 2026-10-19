import type { MarcaDisponibilidad, MedioAdmin } from '../../../api/medios';
import { ConfirmDialog } from '../../../components/ConfirmDialog';
import {
  RESOLVE_SHOW_AGAIN,
  SHOW_AGAIN_ACTION,
  SHOW_AGAIN_CANCEL,
  showAgainConfirm,
} from '../messages';

export interface ResolverMarcaDialogProps {
  /** Ficha cuya marca se va a resolver. */
  item: MedioAdmin;
  /** Marca **pendiente** que se resuelve (cada eje viaja con la suya, P9-19). */
  mark: MarcaDisponibilidad;
  loading: boolean;
  onConfirm: () => void;
  onClose: () => void;
}

/**
 * Confirmación **suave** de la única resolución de marca que puede volver a
 * exponer al público un enlace posiblemente defectuoso (ux.md §4.7, ⟲ D-28):
 * **«Mostrar de nuevo»** sobre una ficha con la marca «No disponible».
 *
 * Dice la consecuencia semanal completa —«Si el enlace de “{nombre}” sigue sin
 * funcionar, volverá a quedar oculto la próxima vez que se revise.»— y ofrece
 * «Mostrar» / «Dejarlo como está». Es coherente con la precedencia aprobada en
 * la fase 5 (A3): una indisponibilidad **confirmada** posterior vuelve a ocultar
 * la ficha y a marcarla, aunque el equipo la hubiera mostrado.
 *
 * Las resoluciones **inocuas** («Mantener oculto», «Mantener visible») son
 * directas desde la fila y no pasan por aquí (⟲ D-29).
 */
export function ResolverMarcaDialog({
  item,
  mark,
  loading,
  onConfirm,
  onClose,
}: ResolverMarcaDialogProps) {
  return (
    // La marca que se resuelve queda declarada en el DOM: cada eje se resuelve
    // por separado (P9-19) y así la interfaz —y sus pruebas— lo pueden leer.
    <div data-mark={mark}>
      <ConfirmDialog
        title={RESOLVE_SHOW_AGAIN}
        description={showAgainConfirm(item.title)}
        confirmText={SHOW_AGAIN_ACTION}
        cancelText={SHOW_AGAIN_CANCEL}
        loading={loading}
        onConfirm={onConfirm}
        onClose={onClose}
      />
    </div>
  );
}
