import { useCallback, useEffect, useId, useRef, useState, type ReactNode } from 'react';
import { useLanguage } from '../i18n/useLanguage';
import { usePrefersReducedMotion } from '../motion/usePrefersReducedMotion';

export interface CardCarouselProps {
  /** Nombre accesible de la lista (lo que anuncia el lector al enfocarla). */
  label: string;
  /** Tarjetas: cada hijo debe ser un `<li>` (la lista es la única del carrusel). */
  children: ReactNode;
  className?: string;
}

/**
 * Clases de los controles: circulares, objetivo táctil ≥ 44 px, foco visible del
 * sistema y estado deshabilitado perceptible (sin depender solo del color:
 * `disabled` retira el botón del orden de tabulación y atenúa su opacidad).
 */
const CONTROL = [
  'inline-flex min-h-11 min-w-11 items-center justify-center rounded-full',
  'border border-line bg-surface-raised text-ink',
  'transition duration-[var(--motion-duration-micro)] ease-[var(--motion-ease-standard)]',
  'hover:border-accent hover:text-accent',
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus',
  'disabled:cursor-not-allowed disabled:opacity-40',
].join(' ');

const TRACK = [
  'public-carousel rounded-2xl',
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus',
].join(' ');

/**
 * Carrusel horizontal de tarjetas de la portada (2026-10-19).
 *
 * Muestra **1** tarjeta en móvil, **2** en tableta (≥ 768 px) y **3** en
 * escritorio (≥ 1024 px) —los anchos y el *snap* viven en `.public-carousel`,
 * `index.css`—. El desplazamiento nativo (swipe/trackpad) funciona sin JS y las
 * flechas desplazan una «página» (`scrollBy` del ancho visible), con `smooth` o
 * `auto` según `prefers-reduced-motion`. Se deshabilitan en los extremos y se
 * recalculan al hacer scroll (listener pasivo), al redimensionar y cuando
 * cambian los hijos (nuevos datos).
 *
 * Es una **pieza de presentación de la portada** (vive en `features/publico/`
 * porque usa el i18n público): no conoce el negocio ni el contenido; solo
 * recibe el nombre accesible de la lista y las tarjetas `<li>`. La `<ul>` es
 * focusable (`tabIndex={0}`) para poder recorrerla con las flechas del teclado
 * cuando el navegador lo soporte, y conserva el nombre accesible recibido.
 */
export function CardCarousel({ label, children, className }: CardCarouselProps) {
  const { t } = useLanguage();
  const reduced = usePrefersReducedMotion();
  const trackRef = useRef<HTMLUListElement | null>(null);
  const listId = useId();
  const [atStart, setAtStart] = useState(true);
  const [atEnd, setAtEnd] = useState(true);

  const updateBounds = useCallback(() => {
    const track = trackRef.current;
    if (!track) {
      return;
    }
    const max = track.scrollWidth - track.clientWidth;
    setAtStart(track.scrollLeft <= 1);
    // Sin desborde (una sola página) ambos controles quedan deshabilitados.
    setAtEnd(max <= 1 || track.scrollLeft >= max - 1);
  }, []);

  useEffect(() => {
    const track = trackRef.current;
    if (!track) {
      return;
    }
    updateBounds();
    track.addEventListener('scroll', updateBounds, { passive: true });
    window.addEventListener('resize', updateBounds);
    return () => {
      track.removeEventListener('scroll', updateBounds);
      window.removeEventListener('resize', updateBounds);
    };
  }, [updateBounds, children]);

  const move = (direction: -1 | 1) => {
    const track = trackRef.current;
    if (!track) {
      return;
    }
    track.scrollBy({
      left: direction * track.clientWidth,
      behavior: reduced ? 'auto' : 'smooth',
    });
  };

  return (
    <div className={['flex flex-col gap-3', className ?? ''].filter(Boolean).join(' ')}>
      <div className="flex justify-end gap-2">
        <button
          type="button"
          aria-label={t('carousel.prev')}
          aria-controls={listId}
          onClick={() => move(-1)}
          disabled={atStart}
          className={CONTROL}
        >
          <span aria-hidden="true" className="text-lg leading-none">
            ←
          </span>
        </button>
        <button
          type="button"
          aria-label={t('carousel.next')}
          aria-controls={listId}
          onClick={() => move(1)}
          disabled={atEnd}
          className={CONTROL}
        >
          <span aria-hidden="true" className="text-lg leading-none">
            →
          </span>
        </button>
      </div>

      <ul id={listId} ref={trackRef} tabIndex={0} aria-label={label} className={TRACK}>
        {children}
      </ul>
    </div>
  );
}
