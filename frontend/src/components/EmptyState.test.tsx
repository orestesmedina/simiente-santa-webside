import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { EmptyState } from './EmptyState';

describe('EmptyState', () => {
  it('el estado vacío se anuncia como status', () => {
    render(<EmptyState message="Aún no hay cuentas." />);
    expect(screen.getByRole('status')).toHaveTextContent('Aún no hay cuentas.');
  });

  it('el estado de error se anuncia como alert con acción', () => {
    render(
      <EmptyState
        kind="error"
        message="No se pudo cargar el listado."
        action={<button type="button">Reintentar</button>}
      />,
    );
    expect(screen.getByRole('alert')).toHaveTextContent('No se pudo cargar el listado.');
    expect(screen.getByRole('button', { name: 'Reintentar' })).toBeInTheDocument();
  });

  it('el estado de carga marca aria-busy', () => {
    render(<EmptyState kind="loading" message="Cargando usuarios…" />);
    expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
  });
});
