import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { StatusPill } from './StatusPill';

describe('StatusPill', () => {
  it('muestra texto explícito para el estado de la cuenta', () => {
    render(<StatusPill value="active" />);
    expect(screen.getByText('Activo')).toBeInTheDocument();
  });

  it('distingue inactivo con su propio texto', () => {
    render(<StatusPill value="inactive" />);
    expect(screen.getByText('Inactivo')).toBeInTheDocument();
  });

  it('cubre los resultados de auditoría', () => {
    const { rerender } = render(<StatusPill value="success" />);
    expect(screen.getByText('Exitoso')).toBeInTheDocument();
    rerender(<StatusPill value="failure" />);
    expect(screen.getByText('Fallido')).toBeInTheDocument();
    rerender(<StatusPill value="not-completed" />);
    expect(screen.getByText('No completada')).toBeInTheDocument();
  });

  it('permite sobrescribir la etiqueta', () => {
    render(<StatusPill value="success" label="Completada" />);
    expect(screen.getByText('Completada')).toBeInTheDocument();
  });

  it('muestra el estado de publicación con texto explícito (Borrador/Publicado)', () => {
    const { rerender } = render(<StatusPill value="draft" />);
    expect(screen.getByText('Borrador')).toBeInTheDocument();
    rerender(<StatusPill value="published" />);
    expect(screen.getByText('Publicado')).toBeInTheDocument();
  });

  it('distingue borrador (gris) de publicado (verde) y no solo por color', () => {
    const { rerender } = render(<StatusPill value="draft" />);
    expect(screen.getByText('Borrador').className).toContain('bg-slate-100');
    rerender(<StatusPill value="published" />);
    expect(screen.getByText('Publicado').className).toContain('bg-green-100');
  });
});
