import { Navigate } from 'react-router-dom';
import { Card } from '../../../components/publico/Card';
import { PublicButton } from '../../../components/publico/PublicButton';
import { Notice } from '../../../components/Notice';
import { SectionSkeleton } from '../components/SectionSkeleton';
import { useLanguage } from '../i18n/useLanguage';
import { DonacionesGroup } from './DonacionesGroup';
import { IntroText } from './IntroText';
import { UsageText } from './UsageText';
import { hasPublicOfrendasContent, usePublicOfrendas } from './hooks/usePublicOfrendas';

/**
 * Página pública `/ofrendas` (US1/US4/US5, ux.md §4.1): H1 «Ofrendas y diezmos»
 * + subtítulo fijo, el texto introductorio, el grupo «01 Cuentas IBAN», el grupo
 * «02 Números de SINPE Móvil» (en ese orden, FR-004) y el texto de uso. El
 * cliente **no** filtra, **no** reordena y **no** reimplementa el fallback de
 * idioma: presenta tal como lo resolvió el servidor (FR-008/FR-012).
 *
 * **Redirección silenciosa (Q9/P7-21)**: la decisión la toma esta página con su
 * propia consulta `usePublicOfrendas`. Solo con la consulta en **éxito y sin
 * contenido publicado** (ni métodos **ni** textos, P7-13) se redirige a la
 * portada con `<Navigate to="/" replace />`: sin H1, sin grillas vacías, sin
 * `HiddenNotice`, sin aviso y sin parpadeo; con la respuesta ya en caché la
 * redirección ocurre en el **primer render** y, si aún no llegó, se muestra solo
 * el esqueleto estándar mientras se decide. Es **distinto a propósito** de
 * `/eventos` de F4: no se unifica.
 *
 * **Ante error NO se redirige** (patrón `MinisteriosPage`): H1 + mensaje +
 * «Volver a intentar»; con contenido en caché se conserva lo último válido y se
 * avisa en discreto.
 */
export function OfrendasPage() {
  const { t } = useLanguage();
  const query = usePublicOfrendas();

  const items = query.data?.items ?? [];
  const intro = query.data?.intro?.text;
  const usage = query.data?.usage?.text;
  const hasContent = hasPublicOfrendasContent(query.data);

  // Carga: solo el esqueleto estándar mientras se decide (nunca contenido vacío).
  if (query.isPending) {
    return <SectionSkeleton lines={4} />;
  }

  // Nada publicado (ni métodos ni textos): redirección silenciosa a la portada.
  if (query.isSuccess && !hasContent) {
    return <Navigate to="/" replace />;
  }

  const ibanItems = items.filter((item) => item.kind === 'iban');
  const sinpeItems = items.filter((item) => item.kind === 'sinpe');

  return (
    <section className="bg-surface text-ink">
      <div className="mx-auto w-full max-w-6xl px-4 py-section-tight sm:px-6">
        <h1 className="font-display text-[clamp(2.5rem,7vw,4.75rem)] leading-none">
          {t('offerings.page.title')}
        </h1>
        {hasContent && (
          <p className="mt-4 max-w-3xl text-lg leading-relaxed text-muted">
            {t('offerings.page.subtitle')}
          </p>
        )}
      </div>

      {query.isError && !query.data && (
        <div className="mx-auto w-full max-w-3xl px-4 pb-section sm:px-6">
          <Card variant="glass" className="flex flex-col items-start gap-block">
            <p className="text-lg leading-relaxed text-ink">{t('offerings.error.load')}</p>
            <PublicButton variant="primary" onClick={() => void query.refetch()}>
              {t('offerings.error.retry')}
            </PublicButton>
          </Card>
        </div>
      )}

      {query.isError && query.data && (
        <div className="mx-auto w-full max-w-6xl px-4 pb-4 sm:px-6">
          <Notice variant="warning">{t('offerings.error.partial')}</Notice>
        </div>
      )}

      {hasContent && (
        <div className="mx-auto flex w-full max-w-6xl flex-col gap-section px-4 pb-section sm:px-6">
          {intro && <IntroText text={intro} />}

          <DonacionesGroup
            id="ofrendas-iban"
            kicker="01"
            titleKey="offerings.iban.heading"
            items={ibanItems}
          />
          <DonacionesGroup
            id="ofrendas-sinpe"
            kicker="02"
            titleKey="offerings.sinpe.heading"
            items={sinpeItems}
          />

          {usage && <UsageText text={usage} />}
        </div>
      )}
    </section>
  );
}
