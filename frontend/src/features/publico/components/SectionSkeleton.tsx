import { useLanguage } from '../i18n/useLanguage';

export interface SectionSkeletonProps {
  /** Número de líneas grises de la sección (además del título). */
  lines?: number;
}

/**
 * Esqueleto de sección del estado de carga de la portada (ux.md §4.1): títulos
 * y líneas `navy-soft` estáticos (sin animación, respeto de
 * `prefers-reduced-motion`) con un anuncio `role="status"` cortés.
 */
export function SectionSkeleton({ lines = 3 }: SectionSkeletonProps) {
  const { t } = useLanguage();

  return (
    <div
      role="status"
      aria-busy="true"
      className="mx-auto w-full max-w-6xl space-y-3 px-4 py-12 sm:px-6"
    >
      <span className="sr-only">{t('status.loading')}</span>
      <div aria-hidden="true" className="h-8 w-1/3 rounded bg-navy-soft" />
      {Array.from({ length: lines }, (_, index) => (
        <div key={index} aria-hidden="true" className="h-4 w-full rounded bg-navy-soft" />
      ))}
    </div>
  );
}
