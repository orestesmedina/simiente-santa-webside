import { Navigate } from 'react-router-dom';
import { Card } from '../../../components/publico/Card';
import { PublicButton } from '../../../components/publico/PublicButton';
import { Notice } from '../../../components/Notice';
import { SectionSkeleton } from '../components/SectionSkeleton';
import { useLanguage } from '../i18n/useLanguage';
import { Reveal } from '../motion/Reveal';
import { RevealGroup } from '../motion/RevealGroup';
import { MinistryCard } from './MinistryCard';
import { usePublicMinistries } from './hooks/usePublicMinistries';

/** Una sola grilla de fichas: 1 columna en móvil, 2 desde tableta (ux.md §4.1). */
const GRID = 'grid grid-cols-1 gap-block md:grid-cols-2';

/**
 * Página pública `/ministerios` (US1/US4/US5, ux.md §4.1): H1 «Ministerios» +
 * subtítulo fijo (D-2) y **una grilla única** de fichas con `RevealGroup`
 * (stagger suave), en el orden manual que entrega el servidor (`sortOrder`,
 * `id`) — el cliente no filtra, no reordena y no reimplementa el fallback de
 * idioma ni la privacidad del encargado (FR-008/FR-026).
 *
 * **Redirección silenciosa (D-14/P6-21/RG6-13)**: la decisión la toma esta
 * página con **su propia consulta** (`usePublicMinistries`). Solo con la
 * consulta en **éxito y sin `items`** se redirige a la portada con
 * `<Navigate to="/" replace />`: sin H1, sin grilla vacía, sin `HiddenNotice`,
 * sin aviso y sin mensaje, y con `replace` para no dejar `/ministerios` en el
 * historial («Atrás» no vuelve a rebotar). Con la respuesta ya en caché la
 * redirección ocurre en el **primer render** (sin destello); si aún no llegó, se
 * muestra **solo el esqueleto estándar** mientras se decide.
 *
 * **Ante error NO se redirige** (patrón `EventsPage`): H1 + mensaje + «Volver a
 * intentar»; con contenido en caché se conserva lo último válido y se avisa en
 * discreto. Es **distinto a propósito** de `/eventos` de F4: no se unifica.
 */
export function MinisteriosPage() {
  const { t } = useLanguage();
  const query = usePublicMinistries();

  const items = query.data?.items ?? [];
  const hasItems = items.length > 0;

  // Carga: solo el esqueleto estándar mientras se decide (nunca la grilla vacía).
  if (query.isPending) {
    return <SectionSkeleton lines={4} />;
  }

  // Sección oculta: éxito y sin publicados → redirección silenciosa a la portada.
  if (query.isSuccess && !hasItems) {
    return <Navigate to="/" replace />;
  }

  return (
    <section aria-label={t('a11y.ministries.heading')} className="bg-surface text-ink">
      <div className="mx-auto w-full max-w-6xl px-4 py-section-tight sm:px-6">
        <h1 className="font-display text-[clamp(2.5rem,7vw,4.75rem)] leading-none">
          {t('ministries.page.title')}
        </h1>
        {hasItems && (
          <p className="mt-4 max-w-3xl text-lg leading-relaxed text-muted">
            {t('ministries.page.subtitle')}
          </p>
        )}
      </div>

      {query.isError && !query.data && (
        <div className="mx-auto w-full max-w-3xl px-4 pb-section sm:px-6">
          <Card variant="glass" className="flex flex-col items-start gap-block">
            <p className="text-lg leading-relaxed text-ink">{t('ministries.error.load')}</p>
            <PublicButton variant="primary" onClick={() => void query.refetch()}>
              {t('ministries.error.retry')}
            </PublicButton>
          </Card>
        </div>
      )}

      {query.isError && query.data && (
        <div className="mx-auto w-full max-w-6xl px-4 pb-4 sm:px-6">
          <Notice variant="warning">{t('ministries.error.partial')}</Notice>
        </div>
      )}

      {/* El recorte del desplazamiento lateral del revelado lo hace `RevealGroup`
          (`overflow-x: clip`, F10 T1018): sin él la grilla desbordaría el ancho
          del documento y la página ganaría scroll horizontal (FR-023/SC-007). */}
      {hasItems && (
        <RevealGroup>
          <Reveal from="left">
            <div className="mx-auto w-full max-w-6xl px-4 pb-section sm:px-6">
              <ul aria-label={t('a11y.ministries.heading')} className={GRID}>
                {items.map((ministry) => (
                  <MinistryCard key={ministry.id} ministry={ministry} />
                ))}
              </ul>
            </div>
          </Reveal>
        </RevealGroup>
      )}
    </section>
  );
}
