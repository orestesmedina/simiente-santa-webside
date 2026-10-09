import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { ConfirmDialog } from './ConfirmDialog';

describe('ConfirmDialog', () => {
  it('muestra el texto y confirma con el botón indicado', async () => {
    const onConfirm = vi.fn();
    render(
      <ConfirmDialog
        title="Desactivar cuenta"
        description="La persona perderá el acceso de inmediato; sus datos se conservan."
        confirmText="Desactivar"
        danger
        onConfirm={onConfirm}
        onClose={() => {}}
      />,
    );

    expect(screen.getByRole('dialog', { name: 'Desactivar cuenta' })).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Desactivar' }));
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });

  it('cancelar invoca onClose sin confirmar', async () => {
    const onConfirm = vi.fn();
    const onClose = vi.fn();
    render(
      <ConfirmDialog
        title="Eliminar rol"
        description="Se borrará el rol."
        confirmText="Eliminar"
        onConfirm={onConfirm}
        onClose={onClose}
      />,
    );

    await userEvent.click(screen.getByRole('button', { name: 'Cancelar' }));
    expect(onClose).toHaveBeenCalledTimes(1);
    expect(onConfirm).not.toHaveBeenCalled();
  });
});
