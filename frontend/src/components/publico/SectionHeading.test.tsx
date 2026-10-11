import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { SectionHeading } from './SectionHeading';

describe('SectionHeading · patrón único de título (FR-003/FR-004)', () => {
  it('renderiza kicker y título con el nivel de encabezado correcto (h2)', () => {
    render(<SectionHeading kicker="01">Quiénes somos</SectionHeading>);

    expect(screen.getByText('01')).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2, name: 'Quiénes somos' })).toBeInTheDocument();
  });

  it('respeta el nivel de encabezado indicado (h3)', () => {
    render(<SectionHeading level={3}>Subtítulo</SectionHeading>);

    expect(screen.getByRole('heading', { level: 3, name: 'Subtítulo' })).toBeInTheDocument();
  });

  it('el acento turquesa es decorativo (aria-hidden) y usa la capa de movimiento', () => {
    const { container } = render(<SectionHeading>Contacto</SectionHeading>);

    const accent = container.querySelector('.section-heading-accent');
    expect(accent).not.toBeNull();
    expect(accent).toHaveAttribute('aria-hidden', 'true');
  });

  it('propaga el id al encabezado para el aria-labelledby de la sección', () => {
    render(<SectionHeading id="contacto-title">Contacto</SectionHeading>);

    expect(screen.getByRole('heading', { level: 2, name: 'Contacto' })).toHaveAttribute(
      'id',
      'contacto-title',
    );
  });

  it('usa utilidades semánticas (no colores explícitos de F3)', () => {
    render(<SectionHeading kicker="01">Contacto</SectionHeading>);

    const heading = screen.getByRole('heading', { level: 2 });
    expect(heading.className).toContain('font-display');
    expect(screen.getByText('01').className).toContain('text-accent');
  });
});
