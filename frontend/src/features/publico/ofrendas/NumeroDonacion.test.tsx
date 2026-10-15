import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { NumeroDonacion } from './NumeroDonacion';

/**
 * `NumeroDonacion` (FR-003/FR-015/SC-009): el IBAN se lee agrupado en bloques de
 * 4 y el SINPE en su formato normalizado; en ambos casos el número **íntegro y
 * sin separadores** queda expuesto para el lector de pantalla, de modo que la
 * lectura corrida coincide exactamente con el valor copiable.
 */
describe('NumeroDonacion (FR-003/SC-009)', () => {
  it('muestra el IBAN agrupado en bloques de 4 y expone el número exacto sin separadores', () => {
    render(<NumeroDonacion kind="iban" number="CR05015202001234567890" />);

    // Representación de lectura (agrupada).
    expect(screen.getByText('CR05 0152 0200 1234 5678 90')).toBeInTheDocument();
    // Lectura corrida: el número completo, sin separadores de bloques.
    const exacto = screen.getByText('CR05015202001234567890');
    expect(exacto).toBeInTheDocument();
    expect(exacto).toHaveClass('sr-only');
  });

  it('muestra el SINPE normalizado (sin agrupar) y sin perder caracteres', () => {
    render(<NumeroDonacion kind="sinpe" number="+50688888888" />);

    expect(screen.getByText('+50688888888')).toBeInTheDocument();
  });

  it('nunca recorta el número: sin elipsis ni título truncado', () => {
    const { container } = render(<NumeroDonacion kind="iban" number="ES9121000418450200051332" />);

    const parrafo = container.querySelector('p');
    expect(parrafo?.className).toContain('overflow-wrap:anywhere');
    expect(parrafo?.className).not.toContain('truncate');
    expect(parrafo?.className).not.toContain('line-clamp');
  });
});
