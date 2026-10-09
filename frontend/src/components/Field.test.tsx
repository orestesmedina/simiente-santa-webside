import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Field } from './Field';

describe('Field', () => {
  it('asocia la etiqueta con el campo por htmlFor/id', () => {
    render(<Field label="Correo" type="email" />);
    expect(screen.getByLabelText('Correo')).toBeInTheDocument();
  });

  it('asocia la ayuda y el error con aria-describedby y aria-invalid', () => {
    render(
      <Field label="Teléfono" error="Escribe un teléfono válido" help="Con prefijo si procede" />,
    );
    const campo = screen.getByLabelText('Teléfono');
    expect(campo).toHaveAttribute('aria-invalid', 'true');
    const descritos = campo.getAttribute('aria-describedby')?.split(' ') ?? [];
    expect(descritos).toHaveLength(2);
    descritos.forEach((id) => expect(document.getElementById(id)).not.toBeNull());
    expect(screen.getByRole('alert')).toHaveTextContent('Escribe un teléfono válido');
  });

  it('sin error no marca aria-invalid', () => {
    render(<Field label="Nombre" />);
    expect(screen.getByLabelText('Nombre')).not.toHaveAttribute('aria-invalid');
  });
});
