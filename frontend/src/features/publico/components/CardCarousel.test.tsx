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
 * que el carrusel recalcule sus extremos y su progreso, igual que el navegador.
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

/** Simula el ancho del primer `<li>` (en jsdom todos miden 0 por defecto). */
function medirPrimeraTarjeta(lista: HTMLElement, width: number) {
  const primera = within(lista).getAllByRole('listitem')[0];
  Object.defineProperty(primera, 'getBoundingClientRect', {
    configurable: true,
    value: () => ({ width, height: 0, top: 0, left: 0, right: width, bottom: 0, x: 0, y: 0 }),
  });
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
    // El anillo de foco va en el contenedor sin máscara, no en la lista.
    expect(lista.className).not.toContain('focus-visible:outline');
    expect(lista.parentElement).toHaveClass('carousel-rail');
    const items = within(lista).getAllByRole('listitem');
    expect(items).toHaveLength(3);
    for (const item of items) {
      expect(item.tagName).toBe('LI');
    }
  });

  it('los controles son laterales, circulares, ≥ 44 px, con nombre traducido y aria-controls a la lista', () => {
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
      // Rail moderno: superpuestos a los costados y ocultos en móvil.
      expect(boton.className).toContain('absolute');
      expect(boton.className).toContain('top-1/2');
      expect(boton.className).toContain('hidden');
      expect(boton.className).toContain('md:inline-flex');
      // El deshabilitado se desvanece, no se pinta gris.
      expect(boton.className).toContain('disabled:opacity-0');
      expect(boton.className).toContain('disabled:pointer-events-none');
    }

    expect(anterior.className).toContain('-translate-x-1/2');
    expect(siguiente.className).toContain('translate-x-1/2');
  });

  it('se deshabilita en los extremos, marca el degradado correcto y avanza una tarjeta', async () => {
    const user = userEvent.setup();
    renderCarousel();

    const lista = screen.getByRole('list', { name: 'Tarjetas de prueba' });
    const scrollBy = vi.fn();
    Object.defineProperty(lista, 'scrollBy', { configurable: true, value: scrollBy });
    const anterior = screen.getByRole('button', { name: 'Anterior' });
    const siguiente = screen.getByRole('button', { name: 'Siguiente' });

    // Valores asimétricos a propósito: 60 px de tarjeta + 8 px de gap = 68 px,
    // muy distinto de los 120 px visibles del track. Si `move()` volviera a
    // avanzar una página (`clientWidth`), esta aserción fallaría (120 ≠ 68).
    medirPrimeraTarjeta(lista, 60);
    lista.style.columnGap = '8px';

    // Al inicio solo se puede avanzar; el degradado aparece a la derecha.
    medir(lista, { scrollWidth: 360, clientWidth: 120, scrollLeft: 0 });
    expect(anterior).toBeDisabled();
    expect(siguiente).toBeEnabled();
    expect(lista).toHaveAttribute('data-at-start', 'true');
    expect(lista).toHaveAttribute('data-at-end', 'false');

    await user.click(siguiente);
    expect(scrollBy).toHaveBeenCalledWith({ left: 68, behavior: 'smooth' });

    // Al final solo se puede retroceder; el degradado aparece a la izquierda.
    medir(lista, { scrollWidth: 360, clientWidth: 120, scrollLeft: 240 });
    expect(anterior).toBeEnabled();
    expect(siguiente).toBeDisabled();
    expect(lista).toHaveAttribute('data-at-start', 'false');
    expect(lista).toHaveAttribute('data-at-end', 'true');

    await user.click(anterior);
    expect(scrollBy).toHaveBeenLastCalledWith({ left: -68, behavior: 'smooth' });
  });

  it('sin layout disponible cae al ancho visible del track', async () => {
    const user = userEvent.setup();
    renderCarousel();

    const lista = screen.getByRole('list', { name: 'Tarjetas de prueba' });
    const scrollBy = vi.fn();
    Object.defineProperty(lista, 'scrollBy', { configurable: true, value: scrollBy });
    medir(lista, { scrollWidth: 360, clientWidth: 120, scrollLeft: 0 });

    await user.click(screen.getByRole('button', { name: 'Siguiente' }));

    expect(scrollBy).toHaveBeenCalledWith({ left: 120, behavior: 'smooth' });
  });

  it('sin desborde (una sola página) deshabilita ambos controles y oculta el progreso', () => {
    renderCarousel();

    const lista = screen.getByRole('list', { name: 'Tarjetas de prueba' });
    medir(lista, { scrollWidth: 120, clientWidth: 120, scrollLeft: 0 });

    expect(screen.getByRole('button', { name: 'Anterior' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Siguiente' })).toBeDisabled();
    expect(lista).toHaveAttribute('data-at-start', 'true');
    expect(lista).toHaveAttribute('data-at-end', 'true');

    const progreso = screen.getByTestId('carousel-progress');
    expect(progreso).toHaveAttribute('aria-hidden', 'true');
    expect(progreso).toHaveClass('hidden');

    // Aun oculto, el thumb por `ref` queda al 100 % (sin desborde).
    const thumb = screen.getByTestId('carousel-progress-thumb');
    expect(parseFloat(thumb.style.width)).toBeCloseTo(100);
    expect(parseFloat(thumb.style.left)).toBeCloseTo(0);
  });

  it('el indicador de progreso es decorativo y refleja fracción visible y avance', () => {
    renderCarousel();

    const lista = screen.getByRole('list', { name: 'Tarjetas de prueba' });
    const progreso = screen.getByTestId('carousel-progress');
    expect(progreso).toHaveAttribute('aria-hidden', 'true');

    // 100 de 400 visibles: el thumb ocupa el 25 % y arranca a la izquierda.
    medir(lista, { scrollWidth: 400, clientWidth: 100, scrollLeft: 0 });
    expect(progreso).not.toHaveClass('hidden');
    const thumb = screen.getByTestId('carousel-progress-thumb');
    expect(parseFloat(thumb.style.width)).toBeCloseTo(25);
    expect(parseFloat(thumb.style.left)).toBeCloseTo(0);

    // A mitad del recorrido (150 de 300): se desplaza dentro del carril.
    medir(lista, { scrollWidth: 400, clientWidth: 100, scrollLeft: 150 });
    expect(parseFloat(thumb.style.width)).toBeCloseTo(25);
    expect(parseFloat(thumb.style.left)).toBeCloseTo(37.5);

    // Al final el thumb toca el borde derecho sin desbordar.
    medir(lista, { scrollWidth: 400, clientWidth: 100, scrollLeft: 300 });
    expect(parseFloat(thumb.style.width)).toBeCloseTo(25);
    expect(parseFloat(thumb.style.left)).toBeCloseTo(75);
  });

  it('con prefers-reduced-motion desplaza sin animación', async () => {
    installMatchMedia({ [REDUCED_MOTION_QUERY]: true });
    const user = userEvent.setup();
    renderCarousel();

    const lista = screen.getByRole('list', { name: 'Tarjetas de prueba' });
    const scrollBy = vi.fn();
    Object.defineProperty(lista, 'scrollBy', { configurable: true, value: scrollBy });
    // Misma tarjeta asimétrica: 60 px + 8 px de gap = 68 px (no los 120 visibles).
    medirPrimeraTarjeta(lista, 60);
    lista.style.columnGap = '8px';
    medir(lista, { scrollWidth: 360, clientWidth: 120, scrollLeft: 0 });

    await user.click(screen.getByRole('button', { name: 'Siguiente' }));

    expect(scrollBy).toHaveBeenCalledWith({ left: 68, behavior: 'auto' });
  });
});
