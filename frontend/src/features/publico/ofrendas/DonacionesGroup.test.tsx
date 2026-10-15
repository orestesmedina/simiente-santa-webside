import { render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';
import type { MetodoDonacionPublico } from '../../../api/ofrendas';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { DonacionesGroup } from './DonacionesGroup';

function wrapper({ children }: { children: ReactNode }) {
  return <LanguageProvider>{children}</LanguageProvider>;
}

const item: MetodoDonacionPublico = {
  id: '11111111-1111-1111-1111-111111111111',
  kind: 'iban',
  number: 'CR05015202001234567890',
  institution: 'Banco Nacional de Costa Rica',
  currency: 'CRC',
};

describe('DonacionesGroup (FR-004/SC-010, D-5)', () => {
  it('sin fichas en el grupo NO se dibuja el rótulo (cero rótulo huérfano)', () => {
    const { container } = render(
      <DonacionesGroup
        id="ofrendas-sinpe"
        kicker="02"
        titleKey="offerings.sinpe.heading"
        items={[]}
      />,
      { wrapper },
    );

    expect(container).toBeEmptyDOMElement();
  });

  it('con fichas dibuja el rótulo numerado y las tarjetas del grupo', () => {
    render(
      <DonacionesGroup
        id="ofrendas-iban"
        kicker="01"
        titleKey="offerings.iban.heading"
        items={[item]}
      />,
      { wrapper },
    );

    expect(screen.getByRole('heading', { level: 2, name: 'Cuentas IBAN' })).toBeInTheDocument();
    expect(screen.getByText('01')).toBeInTheDocument();
    expect(
      screen.getByRole('heading', { level: 3, name: 'Banco Nacional de Costa Rica' }),
    ).toBeInTheDocument();
  });
});
