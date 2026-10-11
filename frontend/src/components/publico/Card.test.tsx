import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Card } from './Card';

describe('Card · superficie de contenido (FR-003/FR-024)', () => {
  it('usa los tokens semánticos de superficie y el padding del sistema', () => {
    const { container } = render(<Card>Contenido</Card>);

    const card = container.firstElementChild as HTMLElement;
    expect(card.className).toContain('bg-surface-raised');
    expect(card.className).toContain('border-line');
    expect(card.className).toContain('p-panel');
    expect(card.className).toContain('shadow-surface');
  });

  it('expone la variante glass con el vidrio (backdrop-blur)', () => {
    const { container } = render(<Card variant="glass">Contenido</Card>);

    expect((container.firstElementChild as HTMLElement).className).toContain('backdrop-blur');
  });

  it('sin interactive no añade microinteracción de elevación', () => {
    const { container } = render(<Card>Contenido</Card>);

    const card = container.firstElementChild as HTMLElement;
    expect(card.className).not.toContain('hover:-translate-y');
  });

  it('con interactive eleva solo con transform/borde (sin mover el maquetado)', () => {
    const { container } = render(<Card interactive>Contenido</Card>);

    const card = container.firstElementChild as HTMLElement;
    expect(card.className).toContain('hover:-translate-y');
    expect(card.className).toContain('--motion-duration-micro');
    expect(card.className).not.toContain('grayscale');
    expect(card.className).not.toContain('invert');
  });

  it('permite cambiar la etiqueta del contenedor (p. ej. item de lista)', () => {
    const { container } = render(
      <Card as="li" className="lista">
        Contenido
      </Card>,
    );

    expect(container.firstElementChild?.tagName).toBe('LI');
  });

  it('padded={false} quita el padding de panel (marco de imagen a sangre)', () => {
    const { container } = render(<Card padded={false}>Contenido</Card>);

    expect((container.firstElementChild as HTMLElement).className).not.toContain('p-panel');
  });
});
