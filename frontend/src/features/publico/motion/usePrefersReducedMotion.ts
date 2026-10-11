import { useEffect, useState } from 'react';

/** Media query de la preferencia de movimiento reducido (FR-006/SC-005). */
export const REDUCED_MOTION_QUERY = '(prefers-reduced-motion: reduce)';

/**
 * Lectura puntual de la preferencia, sin React. Degrada a «con movimiento»
 * (`false`) si el entorno no expone `matchMedia` (jsdom/entornos antiguos), de
 * modo que el contenido nunca se oculta por una capacidad ausente.
 */
export function prefersReducedMotion(): boolean {
  return window.matchMedia?.(REDUCED_MOTION_QUERY)?.matches ?? false;
}

/**
 * Segunda capa de `prefers-reduced-motion` (F10 · T1017/P10-5/R10-5): la primera
 * es el bloque CSS global de `index.css` (T1004). Expone la preferencia y la
 * respeta **en caliente** (sin recargar) mediante el evento `change` de
 * `matchMedia`; con ella activa, `Reveal`/`RevealGroup` no observan ni animan y
 * el menú móvil abre/cierra al instante.
 */
export function usePrefersReducedMotion(): boolean {
  const [reduced, setReduced] = useState(prefersReducedMotion);

  useEffect(() => {
    const media = window.matchMedia?.(REDUCED_MOTION_QUERY);
    if (!media?.addEventListener) {
      return;
    }
    const onChange = (event: MediaQueryListEvent) => setReduced(event.matches);
    media.addEventListener('change', onChange);
    return () => media.removeEventListener('change', onChange);
  }, []);

  return reduced;
}
