import { render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';
import type { MetodoDonacionPublico } from '../../../api/ofrendas';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { DonacionCard } from './DonacionCard';

function wrapper({ children }: { children: ReactNode }) {
  return <LanguageProvider>{children}</LanguageProvider>;
}

const iban: MetodoDonacionPublico = {
  id: '11111111-1111-1111-1111-111111111111',
  kind: 'iban',
  number: 'CR05015202001234567890',
  institution: 'Banco Nacional de Costa Rica',
  currency: 'CRC',
  holder: 'Simiente Santa',
};

const sinpe: MetodoDonacionPublico = {
  id: '22222222-2222-2222-2222-222222222222',
  kind: 'sinpe',
  number: '+50688888888',
  institution: 'Banco Promerica',
};

describe('DonacionCard (FR-003/FR-004/SC-013)', () => {
  it('ficha IBAN completa: entidad (h3), número en bloques, copiar, moneda y titular', () => {
    render(<DonacionCard item={iban} />, { wrapper });

    expect(
      screen.getByRole('heading', { level: 3, name: 'Banco Nacional de Costa Rica' }),
    ).toBeInTheDocument();
    expect(screen.getByText('CR05 0152 0200 1234 5678 90')).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: 'Copiar número de Banco Nacional de Costa Rica' }),
    ).toBeInTheDocument();
    expect(screen.getByText('Moneda: colones')).toBeInTheDocument();
    expect(screen.getByText('Titular: Simiente Santa')).toBeInTheDocument();
    // F7 no lleva imágenes nunca (Q6/FR-020).
    expect(screen.queryAllByRole('img')).toHaveLength(0);
  });

  it('ficha SINPE: número normalizado y SIN línea de moneda (Q2)', () => {
    render(<DonacionCard item={sinpe} />, { wrapper });

    expect(screen.getByText('+50688888888')).toBeInTheDocument();
    expect(screen.queryByText(/Moneda/)).toBeNull();
  });

  it('sin titular no aparece la línea (ni vacía ni con marcador, US1 esc. 5)', () => {
    render(<DonacionCard item={sinpe} />, { wrapper });

    expect(screen.queryByText(/Titular/)).toBeNull();
  });

  it('la tarjeta no es un enlace en bloque: el único interactivo es el botón de copia', () => {
    render(<DonacionCard item={iban} />, { wrapper });

    expect(screen.queryAllByRole('link')).toHaveLength(0);
    expect(screen.getAllByRole('button')).toHaveLength(1);
  });
});
