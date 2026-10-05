import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { DateRangeFilter } from './DateRangeFilter';

describe('DateRangeFilter', () => {
  it('expone los campos Desde y Hasta y aplica con el botón Filtrar', async () => {
    const onApply = vi.fn();
    const { rerender } = render(
      <DateRangeFilter
        from=""
        to=""
        onFromChange={() => {}}
        onToChange={() => {}}
        onApply={onApply}
        onClear={() => {}}
      />,
    );

    await userEvent.click(screen.getByRole('button', { name: 'Filtrar' }));
    expect(onApply).toHaveBeenCalledTimes(1);

    rerender(
      <DateRangeFilter
        from="2026-10-01"
        to=""
        onFromChange={() => {}}
        onToChange={() => {}}
        onApply={onApply}
        onClear={() => {}}
      />,
    );
    expect(screen.getByLabelText('Desde')).toHaveValue('2026-10-01');
  });

  it('quitar filtros invoca onClear', async () => {
    const onClear = vi.fn();
    render(
      <DateRangeFilter
        from="2026-10-01"
        to="2026-10-05"
        onFromChange={() => {}}
        onToChange={() => {}}
        onApply={() => {}}
        onClear={onClear}
      />,
    );
    await userEvent.click(screen.getByRole('button', { name: 'Quitar filtros' }));
    expect(onClear).toHaveBeenCalledTimes(1);
  });
});
