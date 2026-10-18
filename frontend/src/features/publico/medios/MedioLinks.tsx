import { PublicButton } from '../../../components/publico/PublicButton';
import { useLanguage } from '../i18n/useLanguage';
import { announceTitle } from './anuncio';

export interface MedioLinksProps {
  /** Nombre del episodio (tal como lo trae el feed, FR-003): va en el anuncio. */
  title: string;
  /** Enlace al video en YouTube; ausente si el sermón no tiene video. */
  videoUrl?: string;
  /** Enlace al audio (Spotify for Creators/Anchor); ausente si no tiene audio. */
  audioUrl?: string;
}

/**
 * Fila de enlaces externos de una ficha de medios (ux.md §4.1, ⟲ D-5/D-7): los
 * **únicos** elementos interactivos de la ficha. Cada botón lleva
 *  · el **servicio con palabra visible** («Ver en YouTube» / «Escuchar en
 *    Spotify»),
 *  · el **indicador visible de salida** (icono ↗ + microtexto «(pestaña
 *    nueva)»),
 *  · el **nombre accesible completo** en `aria-label` (servicio + pestaña nueva
 *    + que se sale del sitio de la iglesia), y
 *  · `target="_blank" rel="noopener noreferrer"`, que resuelve `PublicButton`
 *    con `external`.
 *
 * Un episodio con un solo formato muestra **solo su botón**: sin líneas vacías
 * ni marcadores «sin datos» (SC-010). En móvil los botones se apilan a todo el
 * ancho (objetivo ≥ 44 px); en pantallas mayores van en fila con `wrap`.
 */
export function MedioLinks({ title, videoUrl, audioUrl }: MedioLinksProps) {
  const { t } = useLanguage();

  return (
    <div className="flex flex-wrap gap-2 max-md:flex-col">
      {videoUrl && (
        <PublicButton
          external
          href={videoUrl}
          ariaLabel={announceTitle(t, 'a11y.media.link.youtube', title)}
          className="max-md:w-full"
        >
          <span>{t('media.link.youtube')}</span>
          <span aria-hidden="true">↗</span>
          <span aria-hidden="true" className="text-sm font-normal">
            {t('a11y.media.card.newTab')}
          </span>
        </PublicButton>
      )}
      {audioUrl && (
        <PublicButton
          external
          href={audioUrl}
          ariaLabel={announceTitle(t, 'a11y.media.link.spotify', title)}
          className="max-md:w-full"
        >
          <span>{t('media.link.spotify')}</span>
          <span aria-hidden="true">↗</span>
          <span aria-hidden="true" className="text-sm font-normal">
            {t('a11y.media.card.newTab')}
          </span>
        </PublicButton>
      )}
    </div>
  );
}
