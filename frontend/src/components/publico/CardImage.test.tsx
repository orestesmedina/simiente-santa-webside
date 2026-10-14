import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import { CardImage } from './CardImage';

/**
 * Pieza compartida de imagen de tarjeta (promoción T623, FR-020): 16:9 a sangre,
 * sin hueco cuando no hay imagen y con retirada por `onError` sin romper el
 * layout. La usan las fichas de agenda de F4 y las de ministerios de F6.
 */

describe('CardImage (FR-020)', () => {
  it('pinta la imagen con su `alt` y resuelve la URL relativa contra la API', () => {
    render(<CardImage image={{ url: '/api/v1/media/img_x.png', alt: 'Equipo de alabanza' }} />);

    const img = screen.getByRole('img', { name: 'Equipo de alabanza' });
    expect(img).toHaveAttribute('src', `${API_BASE_URL}/api/v1/media/img_x.png`);
    expect(img.className).toContain('aspect-video');
  });

  it('sin imagen no deja hueco ni reserva', () => {
    const { container } = render(<CardImage />);
    expect(container.querySelector('img')).toBeNull();
  });

  it('si la imagen falla, un `onError` la retira sin romper el layout', () => {
    render(<CardImage image={{ url: '/api/v1/media/rota.png', alt: 'Imagen rota' }} />);

    fireEvent.error(screen.getByRole('img', { name: 'Imagen rota' }));

    expect(screen.queryByRole('img')).toBeNull();
  });
});
