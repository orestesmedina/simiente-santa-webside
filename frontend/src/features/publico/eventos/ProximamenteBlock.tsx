import { Card } from '../../../components/publico/Card';
import { CardCarousel } from '../components/CardCarousel';
import { Notice } from '../../../components/Notice';
import { PublicButton } from '../../../components/publico/PublicButton';
import { SectionHeading } from '../../../components/publico/SectionHeading';
import { useLanguage } from '../i18n/useLanguage';
import { UpcomingItem } from './UpcomingItem';
import { usePortadaProximamente } from './hooks/usePortadaProximamente';

/**
 * Rejilla del **esqueleto** (1/2/3 columnas, igual que el carrusel real): al
 * llegar el dato no hay brinco de layout y nunca quedan 5 columnas.
 */
const SKELETON_GRID = 'grid grid-cols-1 gap-block md:grid-cols-2 lg:grid-cols-3';

/**
 * Esqueleto propio del bloque (ux.md §4.10, D-28): el mismo encabezado y una
 * fila de tarjetas fantasma **con las medidas reales** (imagen 16:9 y tres
 * líneas), de modo que al llegar el dato no hay brinco de layout. El bloque se
 * reserva al render —nunca desaparece, ni siquiera en el estado vacío—, así que
 * la espera se neutraliza con el esqueleto (patrón de F10).
 */
function ProximamenteSkeleton() {
  const { t } = useLanguage();

  return (
    <div role="status" aria-busy="true" className="flex flex-col gap-block">
      <span className="sr-only">{t('events.loading')}</span>
      <ul aria-hidden="true" className={SKELETON_GRID}>
        {[0, 1, 2, 3, 4].map((index) => (
          <li
            key={index}
            className="flex flex-col overflow-hidden rounded-[26px] border border-line bg-surface-raised"
          >
            <div className="aspect-video w-full motion-safe:animate-pulse bg-surface" />
            <div className="flex flex-col gap-2 p-fluid">
              <div className="h-4 w-1/2 rounded motion-safe:animate-pulse bg-surface" />
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
 * Bloque «Lo que viene» de la portada (`/`, US7/FR-025…FR-027, ux.md §4.10):
 * hasta **5** ítems próximos —eventos y actividades mezclados por cercanía, con
 * la ventana de 2 meses y su respaldo resueltos por el servidor (⟲ D-26/D-27)—
 * en un carrusel de tarjetas compactas (1/2/3 visibles según el ancho), con el
 * botón «Ver más eventos» hacia `/eventos`.
 *
 * Es la **única excepción** al «ocultar secciones vacías» de F3: sin ningún
 * ítem el bloque **se muestra** con su mensaje («Por el momento no hay eventos
 * ni actividades», `role="status"`) y **sin** botón (no enlaza a una sección
 * vacía, coherencia con SC-012). Con error y sin caché se pinta el mismo
 * mensaje con un aviso y «Volver a intentar»; con ítems en caché se conserva lo
 * último válido y se avisa en discreto (patrón `error.partial` de F3).
 *
 * La portada de F3 **no se rediseña**: es una inserción en su secuencia.
 */
export function ProximamenteBlock() {
  const { t } = useLanguage();
  const query = usePortadaProximamente();

  const items = query.data?.items ?? [];
  const hasItems = items.length > 0;
  const withoutData = query.isError && !query.data;

  return (
    <section id="proximamente" aria-labelledby="proximamente-title" className="py-section">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <Card variant="glass" className="flex flex-col gap-block">
          <div className="flex flex-col gap-2">
            <SectionHeading id="proximamente-title">{t('home.upcoming.title')}</SectionHeading>
            <p className="text-muted">{t('home.upcoming.subtitle')}</p>
          </div>

          {query.isPending && <ProximamenteSkeleton />}

          {query.isError && query.data && <Notice variant="warning">{t('error.partial')}</Notice>}

          {withoutData && (
            <div className="flex flex-col items-start gap-block">
              <Notice variant="error">{t('home.upcoming.error.load')}</Notice>
              <p role="status" className="text-lg leading-relaxed text-ink">
                {t('home.upcoming.empty')}
              </p>
              <PublicButton variant="secondary" onClick={() => void query.refetch()}>
                {t('home.upcoming.error.retry')}
              </PublicButton>
            </div>
          )}

          {hasItems && (
            <>
              <CardCarousel label={t('a11y.home.upcoming.heading')}>
                {items.map((item) => (
                  <UpcomingItem key={`${item.kind}-${item.id}`} item={item} />
                ))}
              </CardCarousel>
              <div className="flex justify-end max-md:justify-stretch">
                <PublicButton href="/eventos" variant="secondary">
                  {t('home.upcoming.seeAll')}
                </PublicButton>
              </div>
            </>
          )}

          {query.isSuccess && !hasItems && (
            <p role="status" className="text-lg leading-relaxed text-ink">
              {t('home.upcoming.empty')}
            </p>
          )}
        </Card>
      </div>
    </section>
  );
}
