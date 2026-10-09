import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { Pagination } from './Pagination';

describe('Pagination', () => {
  it('anuncia el resumen y deshabilita Anterior en la primera página', () => {
    render(
      <Pagination
        page={1}
        totalPages={3}
        onChange={() => {}}
        summaryText="Mostrando 1–20 de 45 registros"
      />,
    );
    expect(screen.getByRole('status')).toHaveTextContent('Mostrando 1–20 de 45 registros');
    expect(screen.getByRole('button', { name: 'Anterior' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Siguiente' })).toBeEnabled();
  });

  it('deshabilita Siguiente en la última página y permite retroceder', async () => {
    const onChange = vi.fn();
    render(
      <Pagination
        page={2}
        totalPages={2}
        onChange={onChange}
        summaryText="Mostrando 21–40 de 40 registros"
      />,
    );
    expect(screen.getByRole('button', { name: 'Siguiente' })).toBeDisabled();
    await userEvent.click(screen.getByRole('button', { name: 'Anterior' }));
    expect(onChange).toHaveBeenCalledWith(1);
  });
});
