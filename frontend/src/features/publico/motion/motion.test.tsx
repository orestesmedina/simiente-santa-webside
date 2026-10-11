import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { installMatchMedia } from '../../../test/matchMedia';
import { REDUCED_MOTION_QUERY, usePrefersReducedMotion } from './usePrefersReducedMotion';

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
