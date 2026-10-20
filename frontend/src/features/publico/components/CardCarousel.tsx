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
 * Clases de los controles: **superpuestos a los costados del rail** (`absolute`,
 * centrados en vertical y asomando sobre el borde), 48 px, vidrio del sistema y
 * objetivo táctil ≥ 44 px. En móvil (`< 768 px`) se ocultan: allí manda el swipe
 * y el degradado de borde + la barra de progreso comunican que hay más.
 *
 * El estado deshabilitado **desaparece** (opacidad 0 y sin puntero) en vez de
 * quedar gris; el atributo `disabled` sigue retirándolo del orden de tabulación.
 */
const CONTROL = [
  'absolute top-1/2 z-10 hidden h-12 w-12 min-h-11 min-w-11 -translate-y-1/2',
  'items-center justify-center rounded-full',
  'border border-line bg-surface/85 text-ink shadow-surface backdrop-blur-md',
  'transition duration-[var(--motion-duration-micro)] ease-[var(--motion-ease-standard)]',
  'hover:border-accent hover:bg-accent hover:text-navy',
  'focus-visible:border-accent focus-visible:bg-accent focus-visible:text-navy',
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus',
  'disabled:pointer-events-none disabled:opacity-0 md:inline-flex',
].join(' ');

const TRACK = 'public-carousel rounded-2xl';

const PROGRESS = 'relative h-1 w-full overflow-hidden rounded-full bg-line';
const THUMB = 'absolute inset-y-0 min-w-6 rounded-full bg-accent';

const clamp01 = (value: number) => Math.min(1, Math.max(0, value));

/** Icono chevron en SVG (nunca caracteres `←`/`→`): hereda el color del botón. */
function Chevron({ direction }: { direction: 'left' | 'right' }) {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      className="h-5 w-5"
    >
      <path d={direction === 'left' ? 'M15 18l-6-6 6-6' : 'M9 18l6-6-6-6'} />
    </svg>
  );
}

/**
 * Carrusel horizontal de tarjetas de la portada (2026-10-19): **rail moderno,
 * flechas laterales, avance de una tarjeta, degradados y progreso**.
 *
 * Muestra **1** tarjeta en móvil, **2** en tableta (≥ 768 px) y **3** en
 * escritorio (≥ 1024 px) —los anchos y el *snap* viven en `.public-carousel`,
 * `index.css`—. El desplazamiento nativo (swipe/trackpad) funciona sin JS; las
 * flechas avanzan **una tarjeta** (ancho del primer `<li>` + `column-gap`, con
 * `clientWidth` como respaldo si el layout no está disponible) y usan `smooth`
 * salvo con `prefers-reduced-motion`.
 *
 * Las flechas se superponen a los costados, centradas en vertical, y solo
 * existen en tableta/escritorio; en móvil el swipe, el degradado del borde y la
 * barra de progreso indican que hay más. La barra refleja la fracción visible
 * (`clientWidth / scrollWidth`) y avanza con `scrollLeft`; se escribe
 * **directamente por `ref`** en el thumb (cero re-render durante el scroll) y
 * solo se alimenta del estado `atStart`/`atEnd`/`hasOverflow`. Se recalcula al
 * hacer scroll (listener pasivo), al redimensionar y cuando cambian los hijos;
 * sin desborde queda oculta. Los degradados de borde (`mask-image`) aparecen
 * solo en el lado que tiene contenido (`data-at-start`/`data-at-end`).
 *
 * El anillo de foco vive en el contenedor `.carousel-rail` (no en la `<ul>`,
 * cuya máscara lo recortaría): `.carousel-rail:has(> .public-carousel:focus-visible)`.
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
  const thumbRef = useRef<HTMLSpanElement | null>(null);
  const listId = useId();
  const [atStart, setAtStart] = useState(true);
  const [atEnd, setAtEnd] = useState(true);
  const [hasOverflow, setHasOverflow] = useState(false);

  const updateBounds = useCallback(() => {
    const track = trackRef.current;
    if (!track) {
      return;
    }
    const max = track.scrollWidth - track.clientWidth;
    const overflow = max > 1;
    setAtStart(track.scrollLeft <= 1);
    // Sin desborde (una sola página) ambos controles quedan deshabilitados.
    setAtEnd(!overflow || track.scrollLeft >= max - 1);
    setHasOverflow(overflow);

    // Progreso escrito por `ref`: `style` directo, sin pasar por el estado y
    // sin re-renderizar en cada evento de scroll.
    const thumb = thumbRef.current;
    if (thumb) {
      const fraction = track.scrollWidth > 0 ? clamp01(track.clientWidth / track.scrollWidth) : 1;
      const offset = overflow ? clamp01(track.scrollLeft / max) : 0;
      thumb.style.width = `${fraction * 100}%`;
      thumb.style.left = `${offset * (1 - fraction) * 100}%`;
    }
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
    // Avance de **una tarjeta**: ancho del primer `<li>` + separación del rail.
    // En jsdom (sin layout) el ancho es 0 y se recurre al ancho visible.
    const first = track.querySelector('li');
    const cardWidth = first?.getBoundingClientRect().width ?? 0;
    const gap = Number.parseFloat(getComputedStyle(track).columnGap) || 0;
    const step = cardWidth > 0 ? cardWidth + gap : track.clientWidth;

    track.scrollBy({
      left: direction * step,
      behavior: reduced ? 'auto' : 'smooth',
    });
  };

  return (
    <div className={['flex flex-col gap-2', className ?? ''].filter(Boolean).join(' ')}>
      <div className="carousel-rail relative">
        <ul
          id={listId}
          ref={trackRef}
          tabIndex={0}
          aria-label={label}
          data-at-start={atStart ? 'true' : 'false'}
          data-at-end={atEnd ? 'true' : 'false'}
          className={TRACK}
        >
          {children}
        </ul>

        <button
          type="button"
          aria-label={t('carousel.prev')}
          aria-controls={listId}
          onClick={() => move(-1)}
          disabled={atStart}
          className={`${CONTROL} left-0 -translate-x-1/2`}
        >
          <Chevron direction="left" />
        </button>
        <button
          type="button"
          aria-label={t('carousel.next')}
          aria-controls={listId}
          onClick={() => move(1)}
          disabled={atEnd}
          className={`${CONTROL} right-0 translate-x-1/2`}
        >
          <Chevron direction="right" />
        </button>
      </div>

      <div
        data-testid="carousel-progress"
        aria-hidden="true"
        className={hasOverflow ? PROGRESS : `${PROGRESS} hidden`}
      >
        <span ref={thumbRef} data-testid="carousel-progress-thumb" className={THUMB} />
      </div>
    </div>
  );
}
