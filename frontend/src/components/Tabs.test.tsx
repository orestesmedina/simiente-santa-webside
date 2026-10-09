import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { Tabs } from './Tabs';

const tabs = [
  { id: 'accesos', label: 'Historial de accesos' },
  { id: 'acciones', label: 'Acciones administrativas' },
];

describe('Tabs', () => {
  it('marca la pestaña activa con aria-current', () => {
    render(<Tabs tabs={tabs} active="accesos" onChange={() => {}} />);
    expect(screen.getByRole('button', { name: 'Historial de accesos' })).toHaveAttribute(
      'aria-current',
      'page',
    );
    expect(screen.getByRole('button', { name: 'Acciones administrativas' })).not.toHaveAttribute(
      'aria-current',
    );
  });

  it('notifica el cambio de pestaña', async () => {
    const onChange = vi.fn();
    render(<Tabs tabs={tabs} active="accesos" onChange={onChange} />);
    await userEvent.click(screen.getByRole('button', { name: 'Acciones administrativas' }));
    expect(onChange).toHaveBeenCalledWith('acciones');
  });
});
