import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { NumeroDisplay } from './NumeroDisplay';

const IBAN = 'CR05015202001234567890';
const SINPE = '+50688887777';

function stubClipboard(writeText: (value: string) => Promise<void>) {
  Object.defineProperty(navigator, 'clipboard', {
    value: { writeText },
    configurable: true,
    writable: true,
  });
}

describe('NumeroDisplay (panel, FR-015/SC-009, ⟲ D-9)', () => {
  beforeEach(() => stubClipboard(() => Promise.resolve()));
  afterEach(() => vi.restoreAllMocks());

  it('presenta el IBAN agrupado en bloques de 4 pero COPIA el número exacto', async () => {
    const writeText = vi.fn(() => Promise.resolve());
    stubClipboard(writeText);

    render(<NumeroDisplay kind="iban" number={IBAN} />);

    // Presentación agrupada para leer...
    expect(screen.getByText(/CR05 0152 0200 1234 5678 90/)).toBeInTheDocument();
    // ...y el número íntegro disponible para lectores de pantalla.
    expect(screen.getByText(IBAN)).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Copiar número' }));

    expect(writeText).toHaveBeenCalledWith(IBAN);
    expect(await screen.findByRole('status')).toHaveTextContent('Número copiado.');
  });

  it('el SINPE se muestra normalizado y se copia tal cual', async () => {
    const writeText = vi.fn(() => Promise.resolve());
    stubClipboard(writeText);

    render(<NumeroDisplay kind="sinpe" number={SINPE} />);

    expect(screen.getByText(SINPE)).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Copiar número' }));
    expect(writeText).toHaveBeenCalledWith(SINPE);
  });
});
