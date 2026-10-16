import { render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';
import type { MetodoDonacionPublico } from '../../../api/ofrendas';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { OfrendasItem } from './OfrendasItem';

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

describe('OfrendasItem · ficha compacta de la portada (US6, D-12/D-13/D-14)', () => {
  it('IBAN: chip de tipo, entidad (h3), número agrupado, moneda y titular', () => {
    render(<OfrendasItem item={iban} />, { wrapper });

    expect(screen.getByText('Cuenta IBAN')).toBeInTheDocument();
    expect(
      screen.getByRole('heading', { level: 3, name: 'Banco Nacional de Costa Rica' }),
    ).toBeInTheDocument();
    expect(screen.getByText('CR05 0152 0200 1234 5678 90')).toBeInTheDocument();
    expect(screen.getByText('colones')).toBeInTheDocument();
    expect(screen.getByText('Simiente Santa')).toBeInTheDocument();
  });

  it('SINPE: chip propio, número normalizado y SIN moneda (Q2)', () => {
    render(<OfrendasItem item={sinpe} />, { wrapper });

    expect(screen.getByText('SINPE Móvil')).toBeInTheDocument();
    expect(screen.getByText('+50688888888')).toBeInTheDocument();
    expect(screen.queryByText('colones')).toBeNull();
    expect(screen.queryByText('dólares')).toBeNull();
  });

  it('sin titular no aparece línea de titular (0 huecos)', () => {
    render(<OfrendasItem item={sinpe} />, { wrapper });

    expect(screen.queryByText('Simiente Santa')).toBeNull();
  });

  it('la ficha no es interactiva: sin enlace ni botón por ficha (D-14)', () => {
    render(<OfrendasItem item={iban} />, { wrapper });

    expect(screen.queryAllByRole('link')).toHaveLength(0);
    expect(screen.queryAllByRole('button')).toHaveLength(0);
  });
});
