import { vi } from 'vitest';

type ChangeListener = (event: MediaQueryListEvent) => void;

interface QueryState {
  matches: boolean;
  listeners: Set<ChangeListener>;
}

export interface MatchMediaController {
  /** Cambia el resultado de una media query y notifica a sus oyentes. */
  set: (query: string, matches: boolean) => void;
  /** Nº de oyentes registrados en una media query (para aserciones). */
  listenerCount: (query: string) => number;
}

/**
 * Doble de `window.matchMedia` para las pruebas de tema y movimiento (F10).
 * jsdom no implementa `matchMedia` en vivo, así que este helper permite fijar
 * el resultado inicial de cada media query y disparar cambios en caliente.
 */
export function installMatchMedia(initial: Record<string, boolean> = {}): MatchMediaController {
  const queries = new Map<string, QueryState>();

  const stateOf = (query: string): QueryState => {
    const existing = queries.get(query);
    if (existing) {
      return existing;
    }
    const created: QueryState = { matches: initial[query] ?? false, listeners: new Set() };
    queries.set(query, created);
    return created;
  };

  window.matchMedia = vi.fn((query: string) => {
    const state = stateOf(query);
    const mediaQueryList = {
      media: query,
      get matches() {
        return state.matches;
      },
      onchange: null,
      addEventListener: (_type: string, listener: ChangeListener) => {
        state.listeners.add(listener);
      },
      removeEventListener: (_type: string, listener: ChangeListener) => {
        state.listeners.delete(listener);
      },
      addListener: (listener: ChangeListener) => {
        state.listeners.add(listener);
      },
      removeListener: (listener: ChangeListener) => {
        state.listeners.delete(listener);
      },
      dispatchEvent: () => true,
    };
    return mediaQueryList as unknown as MediaQueryList;
  }) as unknown as typeof window.matchMedia;

  return {
    set(query, matches) {
      const state = stateOf(query);
      state.matches = matches;
      for (const listener of [...state.listeners]) {
        listener({ matches, media: query } as MediaQueryListEvent);
      }
    },
    listenerCount(query) {
      return stateOf(query).listeners.size;
    },
  };
}
