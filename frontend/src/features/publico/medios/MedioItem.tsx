import type { MedioPublico } from '../../../api/medios';
import { Card } from '../../../components/publico/Card';
import { CardImage } from '../../../components/publico/CardImage';
import { useLanguage } from '../i18n/useLanguage';
import { CaratulaNeutra } from './CaratulaNeutra';
import { announceTitle } from './anuncio';
import { MedioLinks } from './MedioLinks';

export interface MedioItemProps {
  /** Ítem del extracto, ya resuelto a `lang` por el servidor (FR-025/FR-026). */
  item: MedioPublico;
}

/**
 * Ficha **compacta** del bloque «Medios» de la portada (US6, ux.md §4.9): la
 * variante pequeña de `MedioCard` con los enlaces conservados —carátula 16:9 (o
 * marcador neutro), nombre (h3, máximo 2 líneas visibles y el texto completo en
 * el DOM), descripción corta (2 líneas visibles, completa en el DOM; **omitida
 * si no existe**) y las pastillas de enlace con el **mismo anuncio completo** que
 * en la sección—.
 *
 * **La fecha se omite** (⟲ D-32): el orden ya comunica la recencia y el espacio
 * de la ficha es corto; la fecha vive en `/medios`.
 */
export function MedioItem({ item }: MedioItemProps) {
  const { t } = useLanguage();

  return (
    <Card as="li" padded={false} className="flex flex-col overflow-hidden">
      <CardImage
        image={
          item.coverUrl
            ? { url: item.coverUrl, alt: announceTitle(t, 'a11y.media.image', item.title) }
            : undefined
        }
        fallback={<CaratulaNeutra />}
      />

      <div className="flex flex-1 flex-col gap-2 p-fluid">
        <h3 className="line-clamp-2 font-sans text-lg font-semibold leading-snug text-ink">
          {item.title}
        </h3>

        {item.description && <p className="line-clamp-2 text-muted">{item.description}</p>}

        <div className="mt-auto pt-1">
          <MedioLinks title={item.title} videoUrl={item.videoUrl} audioUrl={item.audioUrl} />
        </div>
      </div>
    </Card>
  );
}
