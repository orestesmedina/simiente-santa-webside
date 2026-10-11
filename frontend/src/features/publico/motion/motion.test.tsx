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
import { REVEAL_FALLBACK_MS, resetRevealObserver } from './useReveal';
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

  it('sin IntersectionObserver el contenido nace visible (JS degradado)', () => {
    uninstallIntersectionObserver();

    render(<Reveal>contenido</Reveal>);

    expect(document.querySelector('.reveal-block')).toHaveAttribute('data-reveal', 'in');
  });
});
