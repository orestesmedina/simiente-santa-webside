import { act, render, renderHook, screen } from '@testing-library/react';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  installIntersectionObserver,
  IntersectionObserverStub,
  uninstallIntersectionObserver,
} from '../../../test/intersectionObserver';
import { installMatchMedia } from '../../../test/matchMedia';
import { Reveal } from './Reveal';
import { RevealGroup } from './RevealGroup';
import { REVEAL_FALLBACK_MS, resetRevealObserver } from './useReveal';
import { useScrolled } from './useScrolled';
import { REDUCED_MOTION_QUERY, usePrefersReducedMotion } from './usePrefersReducedMotion';

/** Hoja de estilos del sistema: se comprueban los tokens de movimiento. */
const css = readFileSync(join(process.cwd(), 'src', 'index.css'), 'utf8');

function tokenMs(name: string): number {
  return Number(new RegExp(`--${name}:\\s*([\\d.]+)ms`).exec(css)?.[1]);
}

function tokenPx(name: string): number {
  return Number(new RegExp(`--${name}:\\s*([\\d.]+)px`).exec(css)?.[1]);
}

afterEach(() => {
  installMatchMedia();
});

describe('usePrefersReducedMotion · preferencia en vivo (FR-006/SC-005)', () => {
  it('el valor inicial sigue la media query', () => {
    installMatchMedia({ [REDUCED_MOTION_QUERY]: true });

    const { result } = renderHook(() => usePrefersReducedMotion());

    expect(result.current).toBe(true);
  });

  it('sin la preferencia entra «con movimiento»', () => {
    installMatchMedia({ [REDUCED_MOTION_QUERY]: false });

    const { result } = renderHook(() => usePrefersReducedMotion());

    expect(result.current).toBe(false);
  });

  it('un cambio de la media se refleja sin recargar', () => {
    const media = installMatchMedia({ [REDUCED_MOTION_QUERY]: false });
    const { result } = renderHook(() => usePrefersReducedMotion());

    act(() => media.set(REDUCED_MOTION_QUERY, true));

    expect(result.current).toBe(true);
  });

  it('sin matchMedia degrada a «con movimiento» sin excepciones', () => {
    const original = window.matchMedia;
    Reflect.deleteProperty(window, 'matchMedia');
    try {
      const { result } = renderHook(() => usePrefersReducedMotion());
      expect(result.current).toBe(false);
    } finally {
      window.matchMedia = original;
    }
  });
});

describe('Reveal · revelado progresivo al scroll (FR-005/FR-009/SC-014)', () => {
  beforeEach(() => {
    installMatchMedia();
    resetRevealObserver();
    installIntersectionObserver();
  });

  afterEach(() => {
    uninstallIntersectionObserver();
    resetRevealObserver();
  });

  it('pasa de pendiente a revelado y los parámetros superan el suelo', () => {
    render(<Reveal from="up">contenido</Reveal>);
    const block = document.querySelector('.reveal-block') as HTMLElement;
    expect(block).toHaveAttribute('data-reveal', 'up');

    act(() => IntersectionObserverStub.instances[0].trigger(true));

    expect(block).toHaveAttribute('data-reveal', 'in');

    // Suelo de perceptibilidad (SC-014): duración en [400, 800] ms y distancia ≥ 16 px.
    expect(tokenMs('motion-duration-reveal')).toBeGreaterThanOrEqual(400);
    expect(tokenMs('motion-duration-reveal')).toBeLessThanOrEqual(800);
    expect(tokenPx('motion-distance')).toBeGreaterThanOrEqual(16);
    // La regla de revelado usa los tokens (sin valores sueltos por componente).
    expect(css).toContain('transition-duration: var(--motion-duration-reveal)');
    expect(css).toContain('translateY(var(--motion-distance))');
    expect(css).toContain('var(--motion-ease-out)');
  });

  it('se revela una sola vez (re-renderizar no re-anima)', () => {
    const { rerender } = render(<Reveal>uno</Reveal>);
    const block = document.querySelector('.reveal-block') as HTMLElement;
    act(() => IntersectionObserverStub.instances[0].trigger(true));
    expect(block).toHaveAttribute('data-reveal', 'in');

    rerender(<Reveal>dos</Reveal>);

    expect(document.querySelector('.reveal-block')).toBe(block);
    expect(block).toHaveAttribute('data-reveal', 'in');
    expect(IntersectionObserverStub.instances[0].targets.size).toBe(0);
  });

  it('con prefers-reduced-motion el contenido nace visible y no se observa', () => {
    installMatchMedia({ [REDUCED_MOTION_QUERY]: true });

    render(<Reveal>contenido</Reveal>);

    expect(document.querySelector('.reveal-block')).toHaveAttribute('data-reveal', 'in');
    expect(IntersectionObserverStub.instances).toHaveLength(0);
  });

  it('el foco revela al instante el elemento pendiente', () => {
    render(
      <Reveal>
        <a href="#destino">enlace</a>
      </Reveal>,
    );
    const block = document.querySelector('.reveal-block') as HTMLElement;
    expect(block).toHaveAttribute('data-reveal', 'up');

    act(() => {
      screen.getByRole('link', { name: 'enlace' }).focus();
    });

    expect(block).toHaveAttribute('data-reveal', 'in');
  });

  it('el salvavidas revela el contenido si el observer no dispara', () => {
    vi.useFakeTimers();
    try {
      render(<Reveal>contenido</Reveal>);
      const block = document.querySelector('.reveal-block') as HTMLElement;
      expect(block).toHaveAttribute('data-reveal', 'up');

      act(() => {
        vi.advanceTimersByTime(REVEAL_FALLBACK_MS);
      });

      expect(block).toHaveAttribute('data-reveal', 'in');
    } finally {
      vi.useRealTimers();
    }
  });

  it('un bloque fuera del viewport sigue pendiente tras el salvavidas y se revela al entrar (C1)', () => {
    vi.useFakeTimers();
    try {
      render(<Reveal from="up">contenido</Reveal>);
      const block = document.querySelector('.reveal-block') as HTMLElement;
      const observer = IntersectionObserverStub.instances[0];

      // El observer reporta el elemento, pero FUERA de pantalla: el salvavidas
      // debe cancelarse y el bloque quedarse pendiente (FR-005/SC-014).
      act(() => observer.trigger(false));
      expect(block).toHaveAttribute('data-reveal', 'up');

      act(() => {
        vi.advanceTimersByTime(REVEAL_FALLBACK_MS * 3);
      });
      expect(block, 'sigue pendiente aunque pase el salvavidas').toHaveAttribute(
        'data-reveal',
        'up',
      );

      // Al entrar en pantalla, se revela.
      act(() => observer.trigger(true));
      expect(block).toHaveAttribute('data-reveal', 'in');
    } finally {
      vi.useRealTimers();
    }
  });

  it('sin IntersectionObserver el contenido nace visible (JS degradado)', () => {
    uninstallIntersectionObserver();

    render(<Reveal>contenido</Reveal>);

    expect(document.querySelector('.reveal-block')).toHaveAttribute('data-reveal', 'in');
  });

  it('apagar prefers-reduced-motion en caliente no vuelve a ocultar el bloque (M2)', () => {
    const media = installMatchMedia({ [REDUCED_MOTION_QUERY]: true });

    render(<Reveal>contenido</Reveal>);
    const block = document.querySelector('.reveal-block') as HTMLElement;
    expect(block).toHaveAttribute('data-reveal', 'in');

    act(() => media.set(REDUCED_MOTION_QUERY, false));

    expect(block, 'sigue revelado tras apagar la preferencia').toHaveAttribute('data-reveal', 'in');
  });
});

describe('RevealGroup · entrada escalonada acotada (FR-005/SC-014)', () => {
  beforeEach(() => {
    installMatchMedia();
    resetRevealObserver();
    installIntersectionObserver();
  });

  afterEach(() => {
    uninstallIntersectionObserver();
    resetRevealObserver();
  });

  it('reparte índices 0,1,2,3 con tope: más de 4 hijos no amplía la secuencia', () => {
    render(
      <RevealGroup>
        <Reveal>a</Reveal>
        <Reveal>b</Reveal>
        <Reveal>c</Reveal>
        <Reveal>d</Reveal>
        <Reveal>e</Reveal>
        <Reveal>f</Reveal>
      </RevealGroup>,
    );

    const blocks = [...document.querySelectorAll('.reveal-block')] as HTMLElement[];
    expect(blocks.map((block) => block.style.getPropertyValue('--reveal-index'))).toEqual([
      '0',
      '1',
      '2',
      '3',
      '3',
      '3',
    ]);
  });

  it('el retardo se acota al token de escalonado (tope ≤ 240 ms)', () => {
    expect(tokenMs('motion-stagger-step')).toBeGreaterThan(0);
    expect(tokenMs('motion-stagger-step') * 3).toBeLessThanOrEqual(240);
    expect(css).toContain('var(--reveal-index, 0) * var(--motion-stagger-step)');
  });

  it('con movimiento reducido el CSS global anula los retardos', () => {
    expect(css).toContain('@media (prefers-reduced-motion: reduce)');
    expect(css).toContain('transition-delay: 0s !important');
  });
});

describe('useScrolled · cabecera al bajar (FR-014/FR-005)', () => {
  function Harness() {
    const { sentinelRef, scrolled } = useScrolled();
    return (
      <>
        <div ref={sentinelRef} data-testid="sentinel" />
        <header data-scrolled={scrolled ? 'true' : 'false'} />
      </>
    );
  }

  beforeEach(() => {
    installMatchMedia();
    installIntersectionObserver();
  });

  afterEach(() => {
    uninstallIntersectionObserver();
  });

  it('activa data-scrolled al bajar y lo quita al volver, sin listeners de scroll', () => {
    const windowSpy = vi.spyOn(window, 'addEventListener');
    const documentSpy = vi.spyOn(document, 'addEventListener');

    render(<Harness />);
    expect(screen.getByRole('banner')).toHaveAttribute('data-scrolled', 'false');

    act(() => IntersectionObserverStub.instances[0].trigger(false));
    expect(screen.getByRole('banner')).toHaveAttribute('data-scrolled', 'true');

    act(() => IntersectionObserverStub.instances[0].trigger(true));
    expect(screen.getByRole('banner')).toHaveAttribute('data-scrolled', 'false');

    const scrollListeners = [...windowSpy.mock.calls, ...documentSpy.mock.calls].filter(
      ([type]) => type === 'scroll',
    );
    expect(scrollListeners).toHaveLength(0);
    windowSpy.mockRestore();
    documentSpy.mockRestore();
  });
});
