import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { BrandLogo, DEFAULT_BRAND_LOGO_SRC } from './BrandLogo';

describe('BrandLogo', () => {
  it('usa el recurso de marca del repositorio con su alt y dimensiones explícitas', () => {
    render(<BrandLogo />);
    const logo = screen.getByRole('img', { name: /logotipo de la iglesia/i });
    expect(logo).toHaveAttribute('src', DEFAULT_BRAND_LOGO_SRC);
    expect(logo).toHaveAttribute('width', '44');
    expect(logo).toHaveAttribute('height', '44');
  });

  it('conserva la proporción del logo (object-contain) sin efectos de marca', () => {
    render(<BrandLogo />);
    const logo = screen.getByRole('img');
    expect(logo.className).toContain('object-contain');
    // Reglas del Manual (R3-12): sin deformar, sin girar, sin filtros, sin sombras.
    expect(logo.className).not.toMatch(/filter|rotate|transform|shadow|mix-blend/);
  });

  it('permite usar el logotipo publicado con su propio alt y tamaño', () => {
    render(<BrandLogo src="/api/v1/media/img_abc.jpg" alt="Nuestro logo" size={64} />);
    const logo = screen.getByRole('img', { name: 'Nuestro logo' });
    expect(logo).toHaveAttribute('src', '/api/v1/media/img_abc.jpg');
    expect(logo).toHaveAttribute('width', '64');
    expect(logo).toHaveAttribute('height', '64');
  });
});
