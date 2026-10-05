import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Select } from './Select';

const options = [
  { value: '', label: 'Elige un rol' },
  { value: 'r1', label: 'Administrador' },
  { value: 'r2', label: 'Editores' },
];

describe('Select', () => {
  it('asocia la etiqueta y lista las opciones', () => {
    render(<Select label="Rol" options={options} />);
    const select = screen.getByLabelText('Rol');
    expect(select).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'Administrador' })).toBeInTheDocument();
  });

  it('marca el error con aria-invalid y lo hace visible', () => {
    render(<Select label="Rol" options={options} error="Elige un rol de la lista." />);
    expect(screen.getByLabelText('Rol')).toHaveAttribute('aria-invalid', 'true');
    expect(screen.getByRole('alert')).toHaveTextContent('Elige un rol de la lista.');
  });
});
