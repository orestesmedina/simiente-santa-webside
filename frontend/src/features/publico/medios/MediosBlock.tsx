import { Card } from '../../../components/publico/Card';
import { CardCarousel } from '../components/CardCarousel';
import { PublicButton } from '../../../components/publico/PublicButton';
import { SectionHeading } from '../../../components/publico/SectionHeading';
import { Notice } from '../../../components/Notice';
import { useLanguage } from '../i18n/useLanguage';
import { MedioItem } from './MedioItem';
import { usePortadaMedios } from './hooks/usePortadaMedios';

/**
 * Rejilla del **esqueleto** (1/2/3 columnas, igual que el carrusel real): al
 * llegar el dato no hay brinco de layout y nunca quedan 5 columnas.
 */
const SKELETON_GRID = 'grid grid-cols-1 gap-block md:grid-cols-2 lg:grid-cols-3';

/**
 * Esqueleto propio del bloque (ux.md §4.9): mismo encabezado y una fila de
 * fichas fantasma **con área de carátula 16:9** y líneas de texto, con las
 * medidas reales, de modo que al llegar el dato no haya brinco de layout. El
 * bloque **puede desaparecer** si el resultado es 0 publicables.
 */
function MediosSkeleton() {
  const { t } = useLanguage();

  return (
    <div role="status" aria-busy="true" className="flex flex-col gap-block">
      <span className="sr-only">{t('media.loading')}</span>
      <ul aria-hidden="true" className={SKELETON_GRID}>
        {[0, 1, 2, 3, 4].map((index) => (
          <li
            key={index}
            className="flex flex-col overflow-hidden rounded-[26px] border border-line bg-surface-raised"
          >
            <div className="aspect-video w-full motion-safe:animate-pulse bg-surface" />
            <div className="flex flex-col gap-2 p-fluid">
              <div className="h-5 w-full rounded motion-safe:animate-pulse bg-surface" />
              <div className="h-4 w-2/3 rounded motion-safe:animate-pulse bg-surface" />
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}

/**
 * Bloque «Medios» de la portada (`/`, US6/FR-027/SC-015, ux.md §4.9): hasta **5**
 * fichas publicables —tope firme resuelto en el **servidor**, sin parámetro de
 * límite; el cliente no recorta— en el orden de la sección, en un carrusel de
 * fichas compactas (1/2/3 visibles según el ancho) con el botón «Ver todo» hacia
 * `/medios`.
 *
 * **Sin publicables el bloque devuelve `null`: no se dibuja** —sin título, sin
 * mensaje, sin hueco— (FR-027; **diferencia deliberada** con `ProximamenteBlock`
 * de F4, que sí muestra su estado vacío). Igual con error **sin** caché (no deja
 * un título huérfano sin contenido). Con ítems en caché se conserva lo último
 * válido y se avisa en discreto.
 *
 * Se alimenta de `usePortadaMedios` (`GET /api/v1/portada/medios`); **no** es
 * fuente de la presencia en la navegación (P9-16: esa sale del documento de la
 * sección, `usePublicMedios`). Es el **único** contenido de F9 en la portada de
 * F3 y va inmediatamente después de `<OfrendasBlock />` (Q12/FR-027).
 */
export function MediosBlock() {
  const { t } = useLanguage();
  const query = usePortadaMedios();

  const items = query.data?.items ?? [];
  const hasItems = items.length > 0;

  if (query.isPending) {
    return (
      <section id="medios" aria-labelledby="medios-title" className="py-section">
        <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
          <Card variant="glass" className="flex flex-col gap-block">
            <MediosSkeleton />
          </Card>
        </div>
      </section>
    );
  }

  // Error sin caché: el bloque no se dibuja (ni título huérfano ni aviso de error).
  if (!query.data) {
    return null;
  }

  // Sin publicables: el bloque no existe (sin mensaje, a diferencia de F4).
  if (!hasItems) {
    return null;
  }

  return (
    <section id="medios" aria-labelledby="medios-title" className="py-section">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <Card variant="glass" className="flex flex-col gap-block">
          <div className="flex flex-col gap-2">
            <SectionHeading id="medios-title">{t('home.media.title')}</SectionHeading>
            <p className="text-muted">{t('home.media.subtitle')}</p>
          </div>

          {query.isError && <Notice variant="warning">{t('media.error.partial')}</Notice>}

          <CardCarousel label={t('a11y.home.media.heading')}>
            {items.map((item) => (
              <MedioItem key={item.id} item={item} />
            ))}
          </CardCarousel>

          <div className="flex justify-end max-md:justify-stretch">
            <PublicButton href="/medios" variant="secondary">
              {t('home.media.seeAll')}
            </PublicButton>
          </div>
        </Card>
      </div>
    </section>
  );
}
