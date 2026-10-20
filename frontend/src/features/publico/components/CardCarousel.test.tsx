import { fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { REDUCED_MOTION_QUERY } from '../motion/usePrefersReducedMotion';
import { installMatchMedia } from '../../../test/matchMedia';
import { CardCarousel } from './CardCarousel';

function Tarjetas() {
  return (
    <>
      <li>Uno</li>
      <li>Dos</li>
      <li>Tres</li>
    </>
  );
}

function renderCarousel() {
  return render(
    <LanguageProvider>
      <CardCarousel label="Tarjetas de prueba">
        <Tarjetas />
      </CardCarousel>
    </LanguageProvider>,
  );
}

/**
 * jsdom no calcula layout: fija las medidas del track y dispara `scroll` para
 * que el carrusel recalcule sus extremos, igual que haría el navegador.
 */
function medir(
  track: HTMLElement,
  medidas: { scrollWidth: number; clientWidth: number; scrollLeft?: number },
) {
  Object.defineProperty(track, 'scrollWidth', {
    configurable: true,
    value: medidas.scrollWidth,
  });
  Object.defineProperty(track, 'clientWidth', {
    configurable: true,
    value: medidas.clientWidth,
  });
  Object.defineProperty(track, 'scrollLeft', {
    configurable: true,
    writable: true,
    value: medidas.scrollLeft ?? 0,
  });
  fireEvent.scroll(track);
}

afterEach(() => {
  installMatchMedia();
});

describe('CardCarousel · carrusel accesible de tarjetas (2026-10-19)', () => {
  it('conserva el nombre accesible de la lista y renderiza los hijos como <li>', () => {
    renderCarousel();

    const lista = screen.getByRole('list', { name: 'Tarjetas de prueba' });
    expect(lista.tagName).toBe('UL');
    expect(lista).toHaveClass('public-carousel');
    expect(lista).toHaveAttribute('tabindex', '0');
    const items = within(lista).getAllByRole('listitem');
    expect(items).toHaveLength(3);
    for (const item of items) {
      expect(item.tagName).toBe('LI');
    }
  });

  it('los controles son circulares, ≥ 44 px, con nombre traducido y aria-controls a la lista', () => {
    renderCarousel();

    const lista = screen.getByRole('list', { name: 'Tarjetas de prueba' });
    const anterior = screen.getByRole('button', { name: 'Anterior' });
    const siguiente = screen.getByRole('button', { name: 'Siguiente' });

    for (const boton of [anterior, siguiente]) {
      expect(boton).toHaveAttribute('aria-controls', lista.id);
      expect(boton).toHaveAttribute('type', 'button');
      expect(boton.className).toContain('min-h-11');
      expect(boton.className).toContain('min-w-11');
      expect(boton.className).toContain('rounded-full');
    }
  });

  it('se deshabilita en los extremos y desplaza una página completa', async () => {
    const user = userEvent.setup();
    renderCarousel();

    const lista = screen.getByRole('list', { name: 'Tarjetas de prueba' });
    const scrollBy = vi.fn();
    Object.defineProperty(lista, 'scrollBy', { configurable: true, value: scrollBy });
    const anterior = screen.getByRole('button', { name: 'Anterior' });
    const siguiente = screen.getByRole('button', { name: 'Siguiente' });

    // Al inicio solo se puede avanzar.
    medir(lista, { scrollWidth: 300, clientWidth: 100, scrollLeft: 0 });
    expect(anterior).toBeDisabled();
    expect(siguiente).toBeEnabled();

    await user.click(siguiente);
    expect(scrollBy).toHaveBeenCalledWith({ left: 100, behavior: 'smooth' });

    // Al final solo se puede retroceder.
    medir(lista, { scrollWidth: 300, clientWidth: 100, scrollLeft: 200 });
    expect(anterior).toBeEnabled();
    expect(siguiente).toBeDisabled();

    await user.click(anterior);
    expect(scrollBy).toHaveBeenLastCalledWith({ left: -100, behavior: 'smooth' });
  });

  it('sin desborde (una sola página) deshabilita ambos controles', () => {
    renderCarousel();

    const lista = screen.getByRole('list', { name: 'Tarjetas de prueba' });
    medir(lista, { scrollWidth: 100, clientWidth: 100, scrollLeft: 0 });

    expect(screen.getByRole('button', { name: 'Anterior' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Siguiente' })).toBeDisabled();
  });

  it('con prefers-reduced-motion desplaza sin animación', async () => {
    installMatchMedia({ [REDUCED_MOTION_QUERY]: true });
    const user = userEvent.setup();
    renderCarousel();

    const lista = screen.getByRole('list', { name: 'Tarjetas de prueba' });
    const scrollBy = vi.fn();
    Object.defineProperty(lista, 'scrollBy', { configurable: true, value: scrollBy });
    medir(lista, { scrollWidth: 300, clientWidth: 100, scrollLeft: 0 });

    await user.click(screen.getByRole('button', { name: 'Siguiente' }));

    expect(scrollBy).toHaveBeenCalledWith({ left: 100, behavior: 'auto' });
  });
});
