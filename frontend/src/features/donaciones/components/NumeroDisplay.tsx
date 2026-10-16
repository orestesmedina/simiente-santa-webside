import { useEffect, useState } from 'react';
import type { MetodoKind } from '../../../api/ofrendas';
import { Button } from '../../../components/Button';
import { presentarNumero } from '../lib/numero';
import { COPIED, COPY_NUMBER } from '../messages';

export interface NumeroDisplayProps {
  /** Tipo del método (`iban`/`sinpe`): decide la presentación de lectura. */
  kind: MetodoKind;
  /** Número **exacto y normalizado** devuelto por la API (FR-015/SC-009). */
  number: string;
}

/** Tiempo que se muestra la confirmación de copiado antes de ocultarse. */
const FEEDBACK_MS = 2000;

/**
 * Presentación y copia del número en el listado del panel (ux.md §4.3, ⟲ D-9):
 * IBAN agrupado en bloques de 4 o SINPE normalizado, con el valor íntegro y sin
 * separadores disponible para lectores de pantalla, y un botón **Copiar número**
 * que lleva al portapapeles el `number` exacto que devuelve la API (nunca la
 * presentación agrupada, SC-009). La confirmación se anuncia con `role="status"`.
 */
export function NumeroDisplay({ kind, number }: NumeroDisplayProps) {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!copied) {
      return;
    }
    const timer = window.setTimeout(() => setCopied(false), FEEDBACK_MS);
    return () => window.clearTimeout(timer);
  }, [copied]);

  async function handleCopy() {
    try {
      if (!navigator.clipboard?.writeText) {
        throw new Error('portapapeles no disponible');
      }
      await navigator.clipboard.writeText(number);
      setCopied(true);
    } catch {
      setCopied(false);
    }
  }

  const presentado = presentarNumero(kind, number);

  return (
    <div className="flex flex-col items-start gap-1">
      <span className="font-medium tabular-nums text-slate-900 [overflow-wrap:anywhere]">
        {presentado === number ? (
          number
        ) : (
          <>
            <span aria-hidden="true">{presentado}</span>
            <span className="sr-only">{number}</span>
          </>
        )}
      </span>
      <div className="flex flex-wrap items-center gap-2">
        <Button size="sm" variant="secondary" onClick={() => void handleCopy()}>
          {COPY_NUMBER}
        </Button>
        {copied && (
          <span role="status" className="text-sm font-medium text-green-800">
            {COPIED}
          </span>
        )}
      </div>
    </div>
  );
}
