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

  it('la variante accent usa el teal de marca con texto navy (CTAs de la portada)', () => {
    render(<Button variant="accent">Escribir por WhatsApp</Button>);
    const boton = screen.getByRole('button', { name: 'Escribir por WhatsApp' });
    expect(boton.className).toContain('bg-teal');
    expect(boton.className).toContain('text-navy');
  });

  it('mantiene intactas las variantes del panel de F2', () => {
    const { rerender } = render(<Button variant="primary">Guardar</Button>);
    expect(screen.getByRole('button', { name: 'Guardar' }).className).toContain('bg-slate-900');
    rerender(<Button variant="danger">Eliminar</Button>);
    expect(screen.getByRole('button', { name: 'Eliminar' }).className).toContain('bg-red-700');
  });
});
