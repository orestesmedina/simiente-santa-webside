import type { MetodoKind } from '../../../api/ofrendas';
import { presentarNumero } from './format';

/** Tamaño del número: protagonista en la sección, discreto en el extracto. */
export type NumeroDonacionVariant = 'destacado' | 'compacto';

export interface NumeroDonacionProps {
  /** Tipo del método (`iban`/`sinpe`): decide la presentación de lectura. */
  kind: MetodoKind;
  /** Número exacto y normalizado tal como lo publica la API (SC-009). */
  number: string;
  variant?: NumeroDonacionVariant;
}

/**
 * Presentación **exacta** del número de donación (FR-003/FR-015/SC-009, ux.md
 * §4.1/§6.3): IBAN agrupado en bloques de 4 o SINPE en su formato normalizado,
 * con `overflow-wrap: anywhere` para que se reacomode a 320 px **sin perder ni
 * un carácter** y **nunca** con elipsis.
 *
 * La representación agrupada es decorativa para el lector de pantalla
 * (`aria-hidden`); el valor íntegro y sin separadores viaja en un texto solo
 * para lectores, de modo que la lectura corrida coincide con lo que se copia.
 */
const SIZE: Record<NumeroDonacionVariant, string> = {
  destacado: 'text-[clamp(1.125rem,4.2vw,1.5rem)] font-semibold tracking-[0.08em]',
  compacto: 'text-[1.0625rem] font-semibold tracking-[0.04em]',
};

export function NumeroDonacion({ kind, number, variant = 'destacado' }: NumeroDonacionProps) {
  const presentado = presentarNumero(kind, number);

  return (
    <p className={`font-sans tabular-nums text-ink [overflow-wrap:anywhere] ${SIZE[variant]}`}>
      {/* Cuando la presentación difiere del valor exacto (IBAN agrupado), la
          versión agrupada es decorativa y el número íntegro y sin separadores
          viaja en texto solo para lectores: la lectura corrida coincide con lo
          que se copia. El SINPE ya se presenta normalizado, así que no se
          duplica. */}
      {presentado === number ? (
        number
      ) : (
        <>
          <span aria-hidden="true">{presentado}</span>
          <span className="sr-only">{number}</span>
        </>
      )}
    </p>
  );
}
