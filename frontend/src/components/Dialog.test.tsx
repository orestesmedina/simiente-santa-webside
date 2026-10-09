import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { Dialog } from './Dialog';

describe('Dialog', () => {
  it('expone el rol dialog con título y descripción asociados', () => {
    render(
      <Dialog title="Crear usuario" description="Completa los datos" onClose={() => {}}>
        <p>contenido</p>
      </Dialog>,
    );

    const dialogo = screen.getByRole('dialog', { name: 'Crear usuario' });
    expect(dialogo).toHaveAttribute('aria-modal', 'true');
    expect(dialogo).toHaveAccessibleDescription('Completa los datos');
  });

  it('cierra con la tecla Escape', async () => {
    const onClose = vi.fn();
    render(
      <Dialog title="Editar" onClose={onClose}>
        <button type="button">Guardar</button>
      </Dialog>,
    );

    await userEvent.keyboard('{Escape}');
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('el botón Cerrar invoca onClose', async () => {
    const onClose = vi.fn();
    render(
      <Dialog title="Editar" onClose={onClose}>
        <button type="button">Guardar</button>
      </Dialog>,
    );

    await userEvent.click(screen.getByRole('button', { name: 'Cerrar' }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
