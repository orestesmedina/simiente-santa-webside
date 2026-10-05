import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { PasswordField, type PasswordPolicyCheck } from './PasswordField';

const policy: PasswordPolicyCheck[] = [
  { id: 'min', label: 'Al menos 8 caracteres', met: true },
  { id: 'mayus', label: 'Una mayúscula', met: false },
];

describe('PasswordField', () => {
  it('muestra y oculta la contraseña con un botón accesible', async () => {
    render(<PasswordField label="Contraseña" autoComplete="new-password" />);
    const campo = screen.getByLabelText('Contraseña');
    expect(campo).toHaveAttribute('type', 'password');

    const toggle = screen.getByRole('button', { name: 'Mostrar contraseña' });
    expect(toggle).toHaveAttribute('aria-pressed', 'false');
    await userEvent.click(toggle);

    expect(campo).toHaveAttribute('type', 'text');
    expect(screen.getByRole('button', { name: 'Ocultar contraseña' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
  });

  it('renderiza la checklist en vivo de la política', () => {
    render(<PasswordField label="Contraseña nueva" policy={policy} />);
    const lista = screen.getByRole('list', { name: 'Requisitos de la contraseña' });
    expect(lista).toHaveTextContent('Al menos 8 caracteres');
    expect(lista).toHaveTextContent('Una mayúscula');
  });

  it('anuncia el error asociado al campo', () => {
    render(<PasswordField label="Contraseña" error="No cumple la política" />);
    expect(screen.getByRole('alert')).toHaveTextContent('No cumple la política');
  });
});
