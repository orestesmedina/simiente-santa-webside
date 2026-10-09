import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { Button } from './Button';

describe('Button', () => {
  it('es un botón real con foco y objetivo táctil mínimo', () => {
    render(<Button>Guardar</Button>);
    const boton = screen.getByRole('button', { name: 'Guardar' });
    expect(boton).toHaveAttribute('type', 'button');
    expect(boton.className).toContain('min-h-11');
  });

  it('en modo loading se deshabilita y lo anuncia con aria-busy', () => {
    render(<Button loading>Guardando…</Button>);
    const boton = screen.getByRole('button', { name: 'Guardando…' });
    expect(boton).toBeDisabled();
    expect(boton).toHaveAttribute('aria-busy', 'true');
  });

  it('ejecuta la acción al pulsarlo', async () => {
    const onClick = vi.fn();
    render(<Button onClick={onClick}>Reintentar</Button>);
    await userEvent.click(screen.getByRole('button', { name: 'Reintentar' }));
    expect(onClick).toHaveBeenCalledTimes(1);
  });
});
