import type { MedioPublico } from '../../../api/medios';
import { Card } from '../../../components/publico/Card';
import { CardImage } from '../../../components/publico/CardImage';
import { useLanguage } from '../i18n/useLanguage';
import { CaratulaNeutra } from './CaratulaNeutra';
import { announceTitle } from './anuncio';
import { formatEpisodeDate } from './fecha';
import { MedioLinks } from './MedioLinks';

export interface MedioCardProps {
  /** Ficha **publicable**, ya resuelta a `lang` por el servidor (FR-025/FR-026). */
  item: MedioPublico;
}

/**
 * Ficha pública de un episodio (ux.md §4.1, FR-002…FR-006): `Card solid`
 * `as="li"` con carátula a sangre 16:9 —o el **marcador neutro** si el episodio
 * no tiene carátula o la copia falla (⟲ D-4)—, nombre (h3, **no** enlace: la
 * ficha no es clicable en bloque, ⟲ D-5), fecha de publicación y descripción
 * corta.
 *
 * **Cada elemento existe solo si tiene dato** (SC-010): sin descripción no hay
 * esa línea (precisión 3 de la puerta 2), y con un solo formato se pinta solo
 * ese enlace. La descripción se recorta **solo visualmente** (`line-clamp`),
 * con el texto completo en el DOM para lectores de pantalla (⟲ D-6); el
 * recorte real del dato lo fija el servidor (FR-020).
 */
export function MedioCard({ item }: MedioCardProps) {
  const { t, lang } = useLanguage();

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

      <div className="flex flex-1 flex-col gap-3 p-fluid">
        <h3 className="font-sans text-[1.25rem] font-semibold leading-snug text-ink">
          {item.title}
        </h3>

        <p className="text-sm text-muted">{formatEpisodeDate(item.publishedAt, lang)}</p>

        {item.description && (
          <p className="line-clamp-4 text-lg leading-relaxed text-ink">{item.description}</p>
        )}

        <div className="mt-auto">
          <MedioLinks title={item.title} videoUrl={item.videoUrl} audioUrl={item.audioUrl} />
        </div>
      </div>
    </Card>
  );
}
