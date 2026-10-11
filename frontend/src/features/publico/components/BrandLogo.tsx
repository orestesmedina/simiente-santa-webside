/** Logotipo de marca versionado en el repositorio (R3-12/R3-18). */
export const DEFAULT_BRAND_LOGO_SRC = '/brand/simiente-logo.jpeg';

/** Texto alternativo de respaldo de interfaz (ux.md §7.2, `a11y.logo`). */
export const DEFAULT_BRAND_LOGO_ALT = 'Logotipo de la iglesia Simiente Santa';

export interface BrandLogoProps {
  /**
   * URL del logotipo publicado por el panel. Si se omite se usa el recurso de
   * marca del repositorio (fallback estático del *chrome*, R3-18: nunca
   * contenido en borrador).
   */
  src?: string;
  /** Texto alternativo; si se omite se usa el respaldo de interfaz. */
  alt?: string;
  /** Lado en px de la caja cuadrada del logo (objetivo táctil/maquetado). */
  size?: number;
  className?: string;
}

/**
 * Muestra el logotipo de la iglesia.
 *
 * Reglas del Manual de Identidad (R3-12): el logo **nunca** se deforma
 * (`object-contain`), **nunca** cambia de color (sin `filter`/`mix-blend`),
 * **nunca** se gira (sin `rotate`/`transform`) y **no** lleva efectos (sin
 * sombras). Las dimensiones son explícitas para no desplazar el maquetado al
 * cargar (ux.md §8).
 */
export function BrandLogo({
  src = DEFAULT_BRAND_LOGO_SRC,
  alt = DEFAULT_BRAND_LOGO_ALT,
  size = 44,
  className = '',
}: BrandLogoProps) {
  const classes = ['object-contain', className].filter(Boolean).join(' ');

  return <img src={src} alt={alt} width={size} height={size} className={classes} />;
}
