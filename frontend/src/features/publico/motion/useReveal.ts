/**
 * Observador compartido de revelado (F10 · T1018/P10-4/R10-5).
 *
 * Un **único** `IntersectionObserver` sirve a todos los `Reveal`: cuando un
 * elemento entra en pantalla se avisa **una sola vez** y se deja de observar.
 * Reglas progresivas: si el entorno no expone `IntersectionObserver`, el
 * elemento se revela de inmediato; un `focusin` sobre el elemento pendiente lo
 * revela al instante (el foco nunca queda sobre algo oculto) y un salvavidas de
 * 1 s revela todo lo pendiente si el observador no dispara.
 */

const REVEAL_THRESHOLD = 0.15;
const REVEAL_MARGIN = '0px 0px -10% 0px';

/** Salvavidas: si el observer no dispara, el contenido se revela igualmente. */
export const REVEAL_FALLBACK_MS = 1000;

type RevealCallback = () => void;

let observer: IntersectionObserver | null = null;
const callbacks = new Map<Element, RevealCallback>();

function getObserver(): IntersectionObserver | null {
  if (typeof IntersectionObserver === 'undefined') {
    return null;
  }
  if (!observer) {
    observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) {
            continue;
          }
          const callback = callbacks.get(entry.target);
          if (callback) {
            callback();
          }
        }
      },
      { threshold: REVEAL_THRESHOLD, rootMargin: REVEAL_MARGIN },
    );
  }
  return observer;
}

/**
 * Observa `element` y llama a `onReveal` una sola vez (al entrar en pantalla,
 * al recibir foco o tras el salvavidas). Devuelve la función de limpieza.
 */
export function observeReveal(element: Element, onReveal: RevealCallback): () => void {
  const io = getObserver();
  if (!io) {
    onReveal();
    return () => {};
  }

  // `will-change` temporal: se activa al empezar a observar y se retira al
  // terminar la transición de entrada (no se queda de forma permanente).
  const target = element as HTMLElement;
  target.style.willChange = 'opacity, transform';

  const clearWillChange = () => {
    target.style.willChange = '';
  };

  const reveal = () => {
    callbacks.delete(element);
    io.unobserve(element);
    element.removeEventListener('focusin', reveal);
    window.clearTimeout(timer);
    onReveal();
    element.addEventListener('transitionend', clearWillChange, { once: true });
  };

  callbacks.set(element, reveal);
  io.observe(element);
  element.addEventListener('focusin', reveal);
  const timer = window.setTimeout(reveal, REVEAL_FALLBACK_MS);

  return () => {
    callbacks.delete(element);
    io.unobserve(element);
    element.removeEventListener('focusin', reveal);
    element.removeEventListener('transitionend', clearWillChange);
    window.clearTimeout(timer);
    clearWillChange();
  };
}

/** Reinicia el observador compartido (pruebas). */
export function resetRevealObserver(): void {
  observer?.disconnect();
  observer = null;
  callbacks.clear();
}
