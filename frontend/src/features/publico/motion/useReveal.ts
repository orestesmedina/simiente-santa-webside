/**
 * Observador compartido de revelado (F10 · T1018/P10-4/R10-5).
 *
 * Un **único** `IntersectionObserver` sirve a todos los `Reveal`: cuando un
 * elemento entra en pantalla se avisa **una sola vez** y se deja de observar.
 * Reglas progresivas: si el entorno no expone `IntersectionObserver`, el
 * elemento se revela de inmediato; un `focusin` sobre el elemento pendiente lo
 * revela al instante (el foco nunca queda sobre algo oculto).
 *
 * El **salvavidas** de 1 s cubre solo el caso de un observador que *nunca*
 * reporta: se cancela en cuanto el observer entrega una entrada para el elemento
 * —**intersecte o no**—, de modo que un bloque fuera del viewport **permanece
 * pendiente** hasta entrar en pantalla (el revelado al scroll es el núcleo de
 * FR-005/SC-014).
 */

const REVEAL_THRESHOLD = 0.15;
const REVEAL_MARGIN = '0px 0px -10% 0px';

/** Salvavidas: si el observer nunca reporta, el contenido se revela igualmente. */
export const REVEAL_FALLBACK_MS = 1000;

type RevealCallback = () => void;

interface RevealHandlers {
  /** El observer ha reportado el elemento (intersecte o no): el salvavidas sobra. */
  seen: () => void;
  /** El elemento ha entrado en pantalla: revelar. */
  reveal: RevealCallback;
}

let observer: IntersectionObserver | null = null;
const handlers = new Map<Element, RevealHandlers>();

function getObserver(): IntersectionObserver | null {
  if (typeof IntersectionObserver === 'undefined') {
    return null;
  }
  if (!observer) {
    observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          const handler = handlers.get(entry.target);
          if (!handler) {
            continue;
          }
          // Cualquier entrada (aunque el elemento esté fuera de pantalla)
          // demuestra que el observador funciona: se cancela el salvavidas.
          handler.seen();
          if (entry.isIntersecting) {
            handler.reveal();
          }
        }
      },
      { threshold: REVEAL_THRESHOLD, rootMargin: REVEAL_MARGIN },
    );
  }
  return observer;
}

/**
 * Observa `element` y llama a `onReveal` una sola vez (al entrar en pantalla, al
 * recibir foco o, si el observador nunca reporta, tras el salvavidas). Devuelve
 * la función de limpieza.
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
    handlers.delete(element);
    io.unobserve(element);
    element.removeEventListener('focusin', reveal);
    window.clearTimeout(timer);
    onReveal();
    element.addEventListener('transitionend', clearWillChange, { once: true });
  };

  const seen = () => {
    window.clearTimeout(timer);
  };

  handlers.set(element, { seen, reveal });
  io.observe(element);
  element.addEventListener('focusin', reveal);
  const timer = window.setTimeout(reveal, REVEAL_FALLBACK_MS);

  return () => {
    handlers.delete(element);
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
  handlers.clear();
}
