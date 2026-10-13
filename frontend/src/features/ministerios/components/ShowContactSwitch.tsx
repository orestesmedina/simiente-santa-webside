import { Notice } from '../../../components/Notice';
import {
  LEGEND_LEADER_HIDDEN,
  LEGEND_LEADER_VISIBLE,
  LEGEND_NEEDS_NAME,
  SWITCH_HELP,
  SWITCH_LABEL,
} from '../messages';

export interface ShowContactSwitchProps {
  /** Estado del interruptor: **apagado** al crear (D-13, «oculto por defecto»). */
  checked: boolean;
  /** `true` si el ministerio tiene encargado con nombre (para las leyendas). */
  hasLeader: boolean;
  onChange: (checked: boolean) => void;
}

/**
 * Interruptor de visibilidad de la ficha del encargado (ux.md §4.5, FR-025/
 * FR-026, ⟲ D-8/D-13): `role="switch"` con `aria-checked`, etiqueta visible y
 * operable con teclado. **Nace apagado** al crear: guardar un encargado no lo
 * publica por sí solo; publicarlo es siempre un acto expreso del equipo.
 *
 * Explica en palabras del equipo qué se hará público: ayuda fija siempre visible,
 * `Notice` warning claramente visible cuando está apagado (el estado de fábrica),
 * leyenda verde solo al encender y aviso breve si se enciende sin nombre. Nunca
 * comunica solo con color (el estado también es texto).
 */
export function ShowContactSwitch({ checked, hasLeader, onChange }: ShowContactSwitchProps) {
  return (
    <div className="flex flex-col gap-2">
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className="inline-flex min-h-11 items-center gap-3 text-left focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900"
      >
        <span
          aria-hidden="true"
          className={`relative inline-block h-6 w-11 shrink-0 rounded-full transition-colors ${
            checked ? 'bg-green-700' : 'bg-slate-400'
          }`}
        >
          <span
            className={`absolute top-0.5 h-5 w-5 rounded-full bg-white transition-all ${
              checked ? 'left-[1.375rem]' : 'left-0.5'
            }`}
          />
        </span>
        <span className="font-medium text-slate-900">{SWITCH_LABEL}</span>
      </button>

      <p className="text-sm text-slate-600">{SWITCH_HELP}</p>

      {checked ? (
        hasLeader ? (
          <Notice variant="success">{LEGEND_LEADER_VISIBLE}</Notice>
        ) : (
          <p className="text-sm font-medium text-slate-700">{LEGEND_NEEDS_NAME}</p>
        )
      ) : (
        <Notice variant="warning">{LEGEND_LEADER_HIDDEN}</Notice>
      )}
    </div>
  );
}
