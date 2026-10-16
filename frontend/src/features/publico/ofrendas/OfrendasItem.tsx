import type { MetodoDonacionPublico } from '../../../api/ofrendas';
import { Card } from '../../../components/publico/Card';
import { useLanguage } from '../i18n/useLanguage';
import { NumeroDonacion } from './NumeroDonacion';

export interface OfrendasItemProps {
  /** Método publicado, ya resuelto a `lang` por el servidor (FR-008). */
  item: MetodoDonacionPublico;
}

/** Palabra de interfaz de la moneda (el dato `CRC`/`USD` no se traduce, FR-007). */
const CURRENCY_KEY = {
  CRC: 'offerings.currency.colones',
  USD: 'offerings.currency.dolares',
} as const;

/**
 * Ficha **compacta** de texto del bloque «Ofrendas y diezmos» de la portada (US6,
 * ux.md §4.9, D-12/D-13/D-14): chip de tipo —necesario porque en la portada no
 * hay grupos rotulados—, entidad (`h3`, **no** enlace), número en presentación
 * de lectura **sin elipsis nunca**, moneda **solo en IBAN** y titular **solo si
 * existe**.
 *
 * **Sin imágenes** (Q6/FR-020) y **sin botón de copiar ni enlace por ficha**
 * (D-14): el único paso del bloque es su botón hacia `/ofrendas`, donde vive la
 * copia exacta.
 */
export function OfrendasItem({ item }: OfrendasItemProps) {
  const { t } = useLanguage();
  const badge =
    item.kind === 'iban' ? t('home.offerings.badge.iban') : t('home.offerings.badge.sinpe');
  const currencyKey = item.currency ? CURRENCY_KEY[item.currency] : undefined;

  return (
    <Card as="li" className="flex flex-col gap-2">
      <span className="w-fit rounded-full border border-line px-3 py-1 font-sans text-sm text-muted">
        {badge}
      </span>

      <h3 className="line-clamp-2 font-sans text-[1.0625rem] font-semibold leading-snug text-ink">
        {item.institution}
      </h3>

      <NumeroDonacion kind={item.kind} number={item.number} variant="compacto" />

      {currencyKey && <p className="text-sm text-muted">{t(currencyKey)}</p>}

      {item.holder && (
        <p className="truncate text-sm text-muted" title={item.holder}>
          {item.holder}
        </p>
      )}
    </Card>
  );
}
