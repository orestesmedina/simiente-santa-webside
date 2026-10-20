import { Link } from 'react-router-dom';
import type { ItemMinisterioPortada } from '../../../api/ministerios';
import { Card } from '../../../components/publico/Card';
import { CardImage } from '../../../components/publico/CardImage';
import { useLanguage } from '../i18n/useLanguage';

export interface MinistryItemProps {
  /** Ítem del extracto, ya resuelto a `lang` por el servidor (FR-008). */
  item: ItemMinisterioPortada;
}

/**
 * Tarjeta **compacta** del bloque «Ministerios» de la portada (US6, ux.md §4.9,
 * D-11/D-12): imagen 16:9 opcional, chip neutro «Ministerio» (texto, nunca
 * color), **título que es el enlace** a `/ministerios` con el nombre accesible
 * «Ministerio: {nombre}» —recortado a 2 líneas y con caja de 44 px— y las líneas
 * de horario y lugar **solo si existen** (texto truncado con el valor completo
 * en `title`).
 *
 * **Sin encargado y sin botón de unirse** (D-12): el extracto invita a la
 * sección, no ejecuta la acción; el CTA del bloque es el único. El DTO del
 * bloque ni siquiera tiene campos de encargado (privacidad estructural, FR-026).
 * La tarjeta **no** es un enlace en bloque: solo el título lo es.
 */
export function MinistryItem({ item }: MinistryItemProps) {
  const { t } = useLanguage();
  const badge = t('home.ministries.badge');

  return (
    <Card as="li" padded={false} className="flex flex-col overflow-hidden">
      <CardImage image={item.image} />

      <div className="flex flex-1 flex-col gap-2 p-fluid">
        <span className="w-fit rounded-full border border-line px-3 py-1 font-sans text-sm text-muted">
          {badge}
        </span>

        <h3 className="font-sans text-lg font-semibold leading-snug text-ink">
          {/* El enlace es el **único** elemento clicable de la tarjeta y cumple el
              objetivo táctil ≥ 44 px (FR-023): la caja es un `flex` de 44 px con
              el título —recortado a 2 líneas— centrado dentro. */}
          <Link
            to="/ministerios"
            aria-label={`${badge}: ${item.name}`}
            className="flex min-h-11 items-center underline-offset-4 hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus"
          >
            <span className="line-clamp-2">{item.name}</span>
          </Link>
        </h3>

        {item.schedule && (
          <p className="flex items-center gap-1 text-muted">
            <span aria-hidden="true">🕒</span>
            <span className="truncate" title={item.schedule}>
              {item.schedule}
            </span>
          </p>
        )}

        {item.place && (
          <p className="flex items-center gap-1 text-muted">
            <span aria-hidden="true">📍</span>
            <span className="truncate" title={item.place}>
              {item.place}
            </span>
          </p>
        )}
      </div>
    </Card>
  );
}
