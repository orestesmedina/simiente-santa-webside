import { render, screen } from '@testing-library/react';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { PublicButton } from './PublicButton';

/** Hoja de estilos del sistema: se comprueban los pares de contraste del hover. */
const css = readFileSync(join(process.cwd(), 'src', 'index.css'), 'utf8');

/** Valores de un token, en orden de aparición (claro y luego oscuro). */
function tokenValues(name: string): string[] {
  return [...css.matchAll(new RegExp(`--${name}:\\s*([^;]+);`, 'g'))].map((match) =>
    match[1].trim(),
  );
}

function luminance(hex: string): number {
  const value = hex.replace('#', '');
  const [r, g, b] = [0, 2, 4].map((index) => parseInt(value.slice(index, index + 2), 16) / 255);
  const lin = (channel: number) =>
    channel <= 0.03928 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4;
  return 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b);
}

/** Razón de contraste WCAG 2.1 entre dos colores hex. */
function contrast(a: string, b: string): number {
  const [la, lb] = [luminance(a), luminance(b)];
  return (Math.max(la, lb) + 0.05) / (Math.min(la, lb) + 0.05);
}

describe('PublicButton · primitiva de acción (FR-003/FR-008)', () => {
  it('renderiza un botón con objetivo táctil ≥ 44 px', () => {
    render(<PublicButton>Acción</PublicButton>);

    const button = screen.getByRole('button', { name: 'Acción' });
    expect(button.className).toContain('min-h-11');
    expect(button.className).toContain('items-center');
  });

  it('expone las tres variantes con clases distintas', () => {
    const { rerender } = render(<PublicButton variant="primary">A</PublicButton>);
    expect(screen.getByRole('button').className).toContain('bg-accent');

    rerender(<PublicButton variant="accent">A</PublicButton>);
    expect(screen.getByRole('button').className).toContain('bg-coral');

    rerender(<PublicButton variant="secondary">A</PublicButton>);
    expect(screen.getByRole('button').className).toContain('border-current');

    rerender(<PublicButton variant="whatsapp">A</PublicButton>);
    expect(screen.getByRole('button').className).toContain('bg-leaf');
  });

  it('renderiza un enlace cuando recibe href', () => {
    render(<PublicButton href="#horario">Ver horario</PublicButton>);

    const link = screen.getByRole('link', { name: 'Ver horario' });
    expect(link).toHaveAttribute('href', '#horario');
    expect(link).not.toHaveAttribute('target');
  });

  it('los enlaces externos abren fuera sin acceso al opener', () => {
    render(
      <PublicButton href="https://wa.me/50688888888" external>
        WhatsApp
      </PublicButton>,
    );

    const link = screen.getByRole('link', { name: 'WhatsApp' });
    expect(link).toHaveAttribute('target', '_blank');
    expect(link).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('el foco visible y la microinteracción usan los tokens de movimiento', () => {
    render(<PublicButton>Acción</PublicButton>);

    const className = screen.getByRole('button').className;
    expect(className).toContain('focus-visible:outline');
    expect(className).toContain('--motion-duration-micro');
    // Solo transform/colores: sin utilidades de tamaño/layout arbitrario.
    expect(className).not.toMatch(/\bw-\[|\bh-\[\d/);
  });

  it('no aplica filtros ni recoloreos (el logotipo queda fuera del sistema)', () => {
    render(<PublicButton>Acción</PublicButton>);

    const className = screen.getByRole('button').className;
    expect(className).not.toContain('grayscale');
    expect(className).not.toContain('invert');
    expect(className).not.toContain('sepia');
  });

  it('el hover del botón primario cumple AA en ambos temas (C2)', () => {
    render(<PublicButton variant="primary">Acción</PublicButton>);

    const className = screen.getByRole('button').className;
    expect(className).toContain('hover:bg-accent-hover');
    expect(className).toContain('hover:text-on-accent-hover');

    // El par de hover (fondo/texto) se define por tema y debe superar AA (4,5:1).
    const backgrounds = tokenValues('accent-hover');
    const texts = tokenValues('on-accent-hover');
    expect(backgrounds, 'un valor por tema (claro y oscuro)').toHaveLength(2);
    expect(texts, 'un valor por tema (claro y oscuro)').toHaveLength(2);
    expect(contrast(backgrounds[0], texts[0]), 'hover en tema claro').toBeGreaterThanOrEqual(4.5);
    expect(contrast(backgrounds[1], texts[1]), 'hover en tema oscuro').toBeGreaterThanOrEqual(4.5);
  });
});
