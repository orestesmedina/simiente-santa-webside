import type { ReactNode } from 'react';

export interface SectionHeadingProps {
  /** Kicker/número de sección (01–05) en Bebas Neue, opcional. */
  kicker?: string;
  /** Texto del título (cada sección resuelve su propia clave i18n). */
  children: ReactNode;
  /** `id` del encabezado (para `aria-labelledby` de la sección). */
  id?: string;
  /** Nivel del encabezado; por defecto `h2` bajo el `h1` del hero. */
  level?: 2 | 3;
  className?: string;
}

/**
 * Patrón **único** de título de sección (FR-003/FR-004): kicker/número en Bebas
 * Neue, título en Bebas y acento turquesa animable (`scaleX`) que respeta
 * `prefers-reduced-motion` por CSS (T1004). Todas las secciones de la portada y
 * las de F4–F9 usan este componente para su encabezado (0 jerarquías distintas).
 */
export function SectionHeading({
  kicker,
  children,
  id,
  level = 2,
  className = '',
}: SectionHeadingProps) {
  const Heading = (level === 3 ? 'h3' : 'h2') as 'h2' | 'h3';

  return (
    <div className={['flex flex-col gap-2', className].filter(Boolean).join(' ')}>
      {kicker && <span className="font-display text-lg tracking-widest text-accent">{kicker}</span>}
      <Heading id={id} className="font-display text-[clamp(1.75rem,4.5vw,2.5rem)] leading-tight">
        {children}
      </Heading>
      <span aria-hidden="true" className="section-heading-accent" />
    </div>
  );
}
