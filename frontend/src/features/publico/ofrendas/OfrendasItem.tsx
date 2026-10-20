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
 * Tarjeta **rectangular tipo *credit card*** del bloque «Ofrendas y diezmos» de
 * la portada (rediseño 2026-10-19): chip de tipo —necesario porque en la portada
 * no hay grupos rotulados—, entidad (`h3`, **no** enlace), número en
 * presentación de lectura **sin elipsis nunca** (`variant="destacado"`), moneda
 * **solo en IBAN** y titular **solo si existe**.
 *
 * El alto se resuelve con `min-height` (44 en móvil, 52 desde `sm` y 72 desde
 * `md`, ~proporción de tarjeta bancaria con dos columnas) y **no** con un
 * `aspect-ratio` estricto: una cuenta muy larga debe poder crecer y envolver el
 * número (`overflow-wrap: anywhere` en `NumeroDonacion`) antes que recortarse.
 * El acento turquesa es decorativo (`aria-hidden`); todo el texto usa tokens
 * sólidos y mantiene AA en ambos temas.
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
    <Card
      as="li"
      padded={false}
      className="flex min-h-44 flex-col gap-2 p-5 sm:min-h-52 sm:p-6 md:min-h-72"
    >
      <span aria-hidden="true" className="h-1 w-10 rounded-full bg-accent" />

      <span className="w-fit rounded-full border border-line px-3 py-1 font-sans text-sm text-muted">
        {badge}
      </span>

      <h3 className="font-sans text-lg font-semibold leading-snug text-ink">{item.institution}</h3>

      <div className="mt-auto flex flex-col gap-1 pt-2">
        <NumeroDonacion kind={item.kind} number={item.number} variant="destacado" />

        {currencyKey && <p className="text-sm text-muted">{t(currencyKey)}</p>}

        {item.holder && (
          <p className="truncate text-sm text-muted" title={item.holder}>
            {item.holder}
          </p>
        )}
      </div>
    </Card>
  );
}
