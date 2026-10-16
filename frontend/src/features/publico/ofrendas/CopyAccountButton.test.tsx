import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { CopyAccountButton } from './CopyAccountButton';

/** El botón resuelve sus textos con el diccionario activo (FR-009). */
function wrapper({ children }: { children: ReactNode }) {
  return <LanguageProvider>{children}</LanguageProvider>;
}

const NUMERO = 'CR05015202001234567890';

function stubClipboard(writeText: (value: string) => Promise<void>) {
  Object.defineProperty(navigator, 'clipboard', {
    value: { writeText },
    configurable: true,
    writable: true,
  });
}

describe('CopyAccountButton (FR-003/SC-009, D-5)', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => {
    window.localStorage.clear();
    vi.restoreAllMocks();
  });

  it('copia el número NORMALIZADO (no la presentación) y anuncia «Número copiado.»', async () => {
    const writeText = vi.fn(() => Promise.resolve());
    stubClipboard(writeText);

    render(<CopyAccountButton number={NUMERO} institution="Banco Nacional" />, { wrapper });

    await userEvent.click(screen.getByRole('button', { name: 'Copiar número de Banco Nacional' }));

    expect(writeText).toHaveBeenCalledWith(NUMERO);
    expect(await screen.findByRole('status')).toHaveTextContent('Número copiado.');
  });

  it('la ayuda de transcripción está siempre visible y el botón es el único interactivo', () => {
    stubClipboard(() => Promise.resolve());

    render(<CopyAccountButton number={NUMERO} institution="Banco Nacional" />, { wrapper });

    expect(
      screen.getByText('Puedes copiarlo en un toque o escribirlo tal cual, bloque por bloque.'),
    ).toBeInTheDocument();
    expect(screen.getAllByRole('button')).toHaveLength(1);
  });

  it('sin portapapeles no se queda en silencio: mantiene la vía de transcripción', async () => {
    Object.defineProperty(navigator, 'clipboard', { value: undefined, configurable: true });

    render(<CopyAccountButton number={NUMERO} institution="Banco Nacional" />, { wrapper });

    await userEvent.click(screen.getByRole('button', { name: 'Copiar número de Banco Nacional' }));

    expect(screen.queryByText('Número copiado.')).toBeNull();
    expect(screen.getByRole('status')).toHaveTextContent(
      'Puedes copiarlo en un toque o escribirlo tal cual, bloque por bloque.',
    );
  });
});
