import { Card } from '../../../components/publico/Card';
import { PublicButton } from '../../../components/publico/PublicButton';
import { SectionHeading } from '../../../components/publico/SectionHeading';
import { Notice } from '../../../components/Notice';
import { useLanguage } from '../i18n/useLanguage';
import { OfrendasItem } from './OfrendasItem';
import { usePortadaOfrendas } from './hooks/usePortadaOfrendas';

/** Rejilla de cuentas (2026-10-19): todas visibles, 1/2 columnas. Sin carrusel. */
const GRID = 'grid grid-cols-1 gap-block md:grid-cols-2';

/**
 * Esqueleto propio del bloque (ux.md §4.9): encabezado y una fila de fichas
 * fantasma **de texto** (más bajas que las de eventos/ministerios: el alto lo da
 * el número). El bloque **puede desaparecer** si el resultado es 0 publicados.
 */
function OfrendasSkeleton() {
  const { t } = useLanguage();

  return (
    <div role="status" aria-busy="true" className="flex flex-col gap-block">
      <span className="sr-only">{t('offerings.loading')}</span>
      <ul aria-hidden="true" className={GRID}>
        {[0, 1, 2, 3, 4].map((index) => (
          <li
            key={index}
            className="flex flex-col gap-2 rounded-[26px] border border-line bg-surface-raised p-fluid"
          >
            <div className="h-5 w-1/2 rounded motion-safe:animate-pulse bg-surface" />
            <div className="h-4 w-2/3 rounded motion-safe:animate-pulse bg-surface" />
            <div className="h-6 w-full rounded motion-safe:animate-pulse bg-surface" />
          </li>
        ))}
      </ul>
    </div>
  );
}

/**
 * Bloque «Ofrendas y diezmos» de la portada (`/`, US6/FR-022/SC-014, ux.md
 * §4.9): hasta **5** métodos publicados —tope firme resuelto en el servidor, sin
 * parámetro de límite— en el orden de la sección, **todas visibles** en una
 * rejilla de tarjetas rectangulares (1 columna en móvil, 2 desde `md`; sin
 * carrusel) con el botón «Ver formas de ofrendar» hacia `/ofrendas`.
 *
 * **Sin métodos publicados el bloque devuelve `null`: no se dibuja** —sin título,
 * sin mensaje, sin hueco— (FR-022: **diferencia deliberada** con
 * `ProximamenteBlock` de F4, que sí muestra su estado vacío). Igual con error
 * **sin** caché (no deja un título huérfano sin contenido). Con ítems en caché se
 * conserva lo último válido y se avisa en discreto (`error.partial` de F3).
 *
 * Se alimenta de `usePortadaOfrendas` (`GET /api/v1/portada/ofrendas`); **no** es
 * fuente de la presencia en la navegación (P7-8: esa sale del documento de la
 * sección, `usePublicOfrendas`; FR-022 solo mira métodos). La portada de F3 no se
 * rediseña: es una inserción tras `MinisteriosBlock`.
 */
export function OfrendasBlock() {
  const { t } = useLanguage();
  const query = usePortadaOfrendas();

  const items = query.data?.items ?? [];
  const hasItems = items.length > 0;

  if (query.isPending) {
    return (
      <section id="ofrendas" aria-labelledby="ofrendas-title" className="py-section">
        <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
          <Card variant="glass" className="flex flex-col gap-block">
            <OfrendasSkeleton />
          </Card>
        </div>
      </section>
    );
  }

  // Error sin caché: el bloque no se dibuja (ni título huérfano ni aviso de error).
  if (!query.data) {
    return null;
  }

  // Sin métodos publicados: el bloque no existe (sin mensaje, a diferencia de F4).
  if (!hasItems) {
    return null;
  }

  return (
    <section id="ofrendas" aria-labelledby="ofrendas-title" className="py-section">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <Card variant="glass" className="flex flex-col gap-block">
          <div className="flex flex-col gap-2">
            <SectionHeading id="ofrendas-title">{t('home.offerings.title')}</SectionHeading>
            <p className="text-muted">{t('home.offerings.subtitle')}</p>
          </div>

          {query.isError && <Notice variant="warning">{t('offerings.error.partial')}</Notice>}

          <ul aria-label={t('a11y.home.offerings.heading')} className={GRID}>
            {items.map((item) => (
              <OfrendasItem key={item.id} item={item} />
            ))}
          </ul>

          <div className="flex justify-end max-md:justify-stretch">
            <PublicButton href="/ofrendas" variant="secondary">
              {t('home.offerings.seeAll')}
            </PublicButton>
          </div>
        </Card>
      </div>
    </section>
  );
}
