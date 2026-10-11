import { Card } from '../../../components/publico/Card';
import { useLanguage } from '../i18n/useLanguage';

export interface SectionSkeletonProps {
  /** Número de líneas grises de la sección (además del título). */
  lines?: number;
}

/**
 * Esqueleto de sección del estado de carga de la portada con la propuesta B
 * (F10 · T1024/FR-009): paneles de vidrio **fantasma** con las mismas medidas y
 * separaciones que la sección destino (mismo `py-section` y mismo `Card glass`),
 * de modo que nada «brinque» al llegar el contenido. El pulso es sutil y
 * **solo si hay movimiento** (`motion-safe:`); el estado de carga nunca depende
 * de la animación (con movimiento reducido se ve el panel completo y estático).
 */
export function SectionSkeleton({ lines = 3 }: SectionSkeletonProps) {
  const { t } = useLanguage();

  return (
    <div role="status" aria-busy="true" className="py-section">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <Card variant="glass" className="flex flex-col gap-block">
          <span className="sr-only">{t('status.loading')}</span>
          <div
            aria-hidden="true"
            className="h-8 w-1/3 rounded bg-surface-raised motion-safe:animate-pulse"
          />
          {Array.from({ length: lines }, (_, index) => (
            <div
              key={index}
              aria-hidden="true"
              className="h-4 w-full rounded bg-surface-raised motion-safe:animate-pulse"
            />
          ))}
        </Card>
      </div>
    </div>
  );
}
