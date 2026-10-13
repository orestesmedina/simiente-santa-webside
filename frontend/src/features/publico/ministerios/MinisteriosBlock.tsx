import { Card } from '../../../components/publico/Card';
import { PublicButton } from '../../../components/publico/PublicButton';
import { SectionHeading } from '../../../components/publico/SectionHeading';
import { Notice } from '../../../components/Notice';
import { useLanguage } from '../i18n/useLanguage';
import { MinistryItem } from './MinistryItem';
import { usePortadaMinisterios } from './hooks/usePortadaMinisterios';

/** Misma rejilla que la fila real (D-28): 1 columna, 2 en tableta, 5 en escritorio. */
const GRID = 'grid grid-cols-1 gap-block md:grid-cols-2 lg:grid-cols-5';

/**
 * Esqueleto propio del bloque (ux.md §4.9, D-9): mismo encabezado y una fila de
 * tarjetas fantasma **con las medidas reales** (imagen 16:9 y tres líneas), de
 * modo que al llegar el dato no hay brinco de layout. El bloque **puede
 * desaparecer** si el resultado es 0 publicados (caso raro y aceptado).
 */
function MinisteriosSkeleton() {
  const { t } = useLanguage();

  return (
    <div role="status" aria-busy="true" className="flex flex-col gap-block">
      <span className="sr-only">{t('ministries.loading')}</span>
      <ul aria-hidden="true" className={GRID}>
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
 * Bloque «Ministerios» de la portada (`/`, US6/FR-027/FR-028, ux.md §4.9): hasta
 * **5** ministerios publicados —tope firme resuelto en el servidor, sin
 * parámetro de límite— en el orden manual del módulo (`sortOrder`, `id`), en una
 * fila de tarjetas compactas con el botón «Ver más ministerios» hacia
 * `/ministerios`.
 *
 * **Sin publicados el bloque devuelve `null`: no se dibuja** —sin título, sin
 * mensaje, sin hueco— (FR-028: regla general de «ocultar secciones vacías» de
 * F3; **diferencia deliberada** con `ProximamenteBlock` de F4, que sí muestra su
 * mensaje de estado vacío). Igual con error **sin** caché (D-10: no deja un
 * título huérfano sin contenido). Con ítems en caché se conserva lo último
 * válido y se avisa en discreto (`error.partial` de F3).
 *
 * La portada de F3 **no se rediseña**: es una inserción en su secuencia (tras el
 * bloque de eventos y antes de «quiénes somos»). Comparte la consulta
 * `usePortadaMinisterios` con `useSiteLinks` por `queryClient`, de modo que una
 * sola petición alimenta bloque, barra, menú móvil y pie.
 */
export function MinisteriosBlock() {
  const { t } = useLanguage();
  const query = usePortadaMinisterios();

  const items = query.data?.items ?? [];
  const hasItems = items.length > 0;

  if (query.isPending) {
    return (
      <section id="ministerios" aria-labelledby="ministerios-title" className="py-section">
        <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
          <Card variant="glass" className="flex flex-col gap-block">
            <MinisteriosSkeleton />
          </Card>
        </div>
      </section>
    );
  }

  // Error sin caché: el bloque no se dibuja (ni título huérfano ni aviso de error).
  if (!query.data) {
    return null;
  }

  // Sin publicados: el bloque no existe (sin mensaje, a diferencia de F4).
  if (!hasItems) {
    return null;
  }

  return (
    <section id="ministerios" aria-labelledby="ministerios-title" className="py-section">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <Card variant="glass" className="flex flex-col gap-block">
          <div className="flex flex-col gap-2">
            <SectionHeading id="ministerios-title">{t('home.ministries.title')}</SectionHeading>
            <p className="text-muted">{t('home.ministries.subtitle')}</p>
          </div>

          {query.isError && <Notice variant="warning">{t('error.partial')}</Notice>}

          <ul aria-label={t('a11y.home.ministries.heading')} className={GRID}>
            {items.map((item) => (
              <MinistryItem key={item.id} item={item} />
            ))}
          </ul>

          <div className="flex justify-end max-md:justify-stretch">
            <PublicButton href="/ministerios" variant="secondary">
              {t('home.ministries.seeAll')}
            </PublicButton>
          </div>
        </Card>
      </div>
    </section>
  );
}
