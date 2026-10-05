import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Notice } from './Notice';

describe('Notice', () => {
  it('anuncia los errores como alert', () => {
    render(<Notice variant="error">No se pudo guardar.</Notice>);
    expect(screen.getByRole('alert')).toHaveTextContent('No se pudo guardar.');
  });

  it('anuncia los avisos y éxitos como status', () => {
    render(<Notice variant="success">Cuenta creada.</Notice>);
    expect(screen.getByRole('status')).toHaveTextContent('Cuenta creada.');
  });

  it('las advertencias son alert y la información status', () => {
    const { rerender } = render(<Notice variant="warning">Cuidado.</Notice>);
    expect(screen.getByRole('alert')).toBeInTheDocument();
    rerender(<Notice variant="info">Nota.</Notice>);
    expect(screen.getByRole('status')).toBeInTheDocument();
  });
});
