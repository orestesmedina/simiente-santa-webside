import type { MetodoDonacionPublico } from '../../../api/ofrendas';
import { Card } from '../../../components/publico/Card';
import { useLanguage } from '../i18n/useLanguage';
import { CopyAccountButton } from './CopyAccountButton';
import { NumeroDonacion } from './NumeroDonacion';

export interface DonacionCardProps {
  /** Método publicado, ya resuelto a `lang` por el servidor (FR-008). */
  item: MetodoDonacionPublico;
}

/**
 * Palabra de **interfaz** de la moneda; el dato (`CRC`/`USD`) no se traduce
 * (FR-007/SC-006): los dos idiomas muestran la misma moneda y solo cambia la
 * etiqueta.
 */
const CURRENCY_KEY = {
  CRC: 'offerings.currency.colones',
  USD: 'offerings.currency.dolares',
} as const;

/**
 * Ficha pública completa de un método de donación (FR-003/FR-004, ux.md §4.1):
 * `Card solid` del sistema de F10 **sin zona de imagen** (Q6/FR-020), con la
 * entidad como `h3`, el `NumeroDonacion` protagonista, el botón de copia, la
 * moneda —**solo IBAN**— y el titular **solo si existe**: sin dato no se dibuja
 * la línea (ni vacía ni con marcador, US1 esc. 5).
 *
 * El tipo no repite chip (lo nombra el grupo en el que vive) y la tarjeta **no
 * es un enlace en bloque**: el único interactivo es el botón de copia.
 */
export function DonacionCard({ item }: DonacionCardProps) {
  const { t } = useLanguage();
  const currencyKey = item.currency ? CURRENCY_KEY[item.currency] : undefined;

  return (
    <Card as="li" className="flex flex-col gap-3">
      <h3 className="font-sans text-[1.25rem] font-semibold leading-snug text-ink">
        {item.institution}
      </h3>

      <NumeroDonacion kind={item.kind} number={item.number} />

      <CopyAccountButton number={item.number} institution={item.institution} />

      {currencyKey && (
        <p className="text-muted">
          {t('offerings.currency.label')}: {t(currencyKey)}
        </p>
      )}

      {item.holder && (
        <p className="flex items-start gap-1 text-muted">
          <span aria-hidden="true">👤</span>
          <span className="[text-wrap:balance]">
            {t('offerings.holder.label')}: {item.holder}
          </span>
        </p>
      )}
    </Card>
  );
}
