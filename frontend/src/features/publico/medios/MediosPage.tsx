import { Navigate } from 'react-router-dom';
import { Card } from '../../../components/publico/Card';
import { PublicButton } from '../../../components/publico/PublicButton';
import { Notice } from '../../../components/Notice';
import { SectionSkeleton } from '../components/SectionSkeleton';
import { useLanguage } from '../i18n/useLanguage';
import { CargarMasButton } from './CargarMasButton';
import { MediosGrid } from './MediosGrid';
import { hasPublicMediosContent, usePublicMedios } from './hooks/usePublicMedios';

/**
 * Página pública `/medios` (US1/US4/US5, ux.md §4.1): H1 «Medios» + subtítulo
 * fijo que **anuncia la salida del sitio** (⟲ D-2) y **una sola lista** de
 * fichas —carátula, nombre, fecha, descripción corta y enlaces externos
 * anunciados— del más reciente al más antiguo, con «Cargar más» por tandas de
 * 12.
 *
 * El cliente **no** filtra, **no** reordena y **no** reimplementa el fallback de
 * idioma ni la regla de publicables: el servidor entrega solo lo publicable
 * (visibles, no absorbidas y sin par de duplicados pendiente, FR-009), ya
 * resuelto al idioma activo (FR-025) y en su orden (FR-006).
 *
 * **Redirección silenciosa (Q10/P9-16/SC-016)**: la decisión la toma esta página
 * con **su propia consulta** (`usePublicMedios`, compartida por `queryClient`
 * con la entrada de navegación de T934). Solo con la consulta en **éxito y 0
 * publicables** se redirige a la portada con `<Navigate to="/" replace />`: sin
 * H1, sin lista vacía, sin `HiddenNotice`, sin aviso y sin destello; con la
 * respuesta ya en caché la redirección ocurre en el **primer render** y, si aún
 * no llegó, se muestra **solo el esqueleto estándar** mientras se decide.
 *
 * **Ante error NO se redirige** (⟲ D-12): H1 + mensaje + «Volver a intentar»; con
 * contenido en caché se conserva lo último válido y se avisa en discreto. Es
 * **distinto a propósito** de `/eventos` de F4: no se unifica.
 *
 * **Cero recursos de terceros** (SC-005): ni reproductores ni peticiones a
 * YouTube/Spotify/Anchor; las carátulas son copias servidas por el propio sitio
 * y el único contacto con esos servicios es el `href` que el visitante activa.
 */
export function MediosPage() {
  const { t } = useLanguage();
  const query = usePublicMedios();

  const items = query.data?.items ?? [];
  const hasItems = items.length > 0;

  // Carga: solo el esqueleto estándar mientras se decide (nunca la lista vacía).
  if (query.isPending) {
    return <SectionSkeleton lines={4} />;
  }

  // Sección sin nada visible: redirección silenciosa a la portada.
  if (query.isSuccess && !hasPublicMediosContent(query.data)) {
    return <Navigate to="/" replace />;
  }

  return (
    <section className="bg-surface text-ink">
      <div className="mx-auto w-full max-w-6xl px-4 py-section-tight sm:px-6">
        <h1 className="font-display text-[clamp(2.5rem,7vw,4.75rem)] leading-none">
          {t('media.page.title')}
        </h1>
        {hasItems && (
          <p className="mt-4 max-w-3xl text-lg leading-relaxed text-muted">
            {t('media.page.subtitle')}
          </p>
        )}
      </div>

      {query.isError && !query.data && (
        <div className="mx-auto w-full max-w-3xl px-4 pb-section sm:px-6">
          <Card variant="glass" className="flex flex-col items-start gap-block">
            <p className="text-lg leading-relaxed text-ink">{t('media.error.load')}</p>
            <PublicButton variant="primary" onClick={() => void query.refetch()}>
              {t('media.error.retry')}
            </PublicButton>
          </Card>
        </div>
      )}

      {query.isError && query.data && (
        <div className="mx-auto w-full max-w-6xl px-4 pb-4 sm:px-6">
          <Notice variant="warning">{t('media.error.partial')}</Notice>
        </div>
      )}

      {hasItems && (
        <div className="mx-auto flex w-full max-w-6xl flex-col gap-block px-4 pb-section sm:px-6">
          <MediosGrid items={items} />
          {query.hasMore && <CargarMasButton onClick={query.loadMore} />}
        </div>
      )}
    </section>
  );
}
