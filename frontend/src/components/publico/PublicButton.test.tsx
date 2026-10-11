import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { PublicButton } from './PublicButton';

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
});
