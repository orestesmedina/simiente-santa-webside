import { createRef } from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
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

/**
 * 2.ª iteración del bug «indicador invisible de los nativos de fecha/hora en el
 * panel»: además de pintar el selector nativo en claro (`color-scheme: light`),
 * cada campo de fecha/hora necesita un **icono visible** que diga dónde pulsar.
 * El botón vive en `Field`, así que lo heredan `DatesEditor`, `ActivityForm` y
 * `DateRangeFilter` sin tocar cada formulario.
 */
describe('Field con selector nativo de fecha/hora', () => {
  /** `showPicker()` no existe en jsdom: se simula/restaura a mano. */
  const definirShowPicker = (valor: (() => void) | undefined) => {
    if (valor) {
      Object.defineProperty(HTMLInputElement.prototype, 'showPicker', {
        configurable: true,
        writable: true,
        value: valor,
      });
    } else {
      delete (HTMLInputElement.prototype as { showPicker?: unknown }).showPicker;
    }
  };

  afterEach(() => {
    definirShowPicker(undefined);
    vi.restoreAllMocks();
  });

  it('añade un botón «Abrir calendario» visible a los campos de fecha', () => {
    render(<Field label="Fecha 1" type="date" />);

    const boton = screen.getByRole('button', { name: 'Abrir calendario' });
    expect(boton).toBeVisible();
    // No debe quedar dentro del `<label>` ni romper su asociación con el campo.
    expect(boton.closest('label')).toBeNull();
    // La etiqueta sigue asociada solo al campo.
    expect(screen.getByLabelText('Fecha 1').tagName).toBe('INPUT');
    // El icono es decorativo; el nombre accesible es del botón.
    expect(boton.querySelector('svg')).toHaveAttribute('aria-hidden', 'true');
  });

  it('el botón del campo de hora se llama «Abrir selector de hora»', () => {
    render(<Field label="Hora de inicio (opcional)" type="time" />);

    expect(screen.getByRole('button', { name: 'Abrir selector de hora' })).toBeVisible();
  });

  it('el botón apunta al input por aria-controls', () => {
    render(<Field label="Fecha" type="date" />);

    const input = screen.getByLabelText('Fecha');
    expect(screen.getByRole('button', { name: 'Abrir calendario' })).toHaveAttribute(
      'aria-controls',
      input.id,
    );
  });

  it('es type="button": pulsarlo no envía el formulario', async () => {
    const onSubmit = vi.fn((event: { preventDefault: () => void }) => event.preventDefault());
    const user = userEvent.setup();
    render(
      <form onSubmit={onSubmit}>
        <Field label="Fecha" type="date" />
      </form>,
    );

    const boton = screen.getByRole('button', { name: 'Abrir calendario' });
    expect(boton).toHaveAttribute('type', 'button');

    await user.click(boton);
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it('al pulsarlo abre el selector nativo con showPicker()', async () => {
    const showPicker = vi.fn();
    definirShowPicker(showPicker);
    const focus = vi.spyOn(HTMLInputElement.prototype, 'focus');
    const user = userEvent.setup();
    render(<Field label="Fecha" type="date" />);

    await user.click(screen.getByRole('button', { name: 'Abrir calendario' }));

    expect(showPicker).toHaveBeenCalledTimes(1);
    expect(focus).not.toHaveBeenCalled();
  });

  it('sin showPicker() hace focus() en el campo como respaldo', async () => {
    definirShowPicker(undefined);
    const focus = vi.spyOn(HTMLInputElement.prototype, 'focus');
    const user = userEvent.setup();
    render(<Field label="Hora de fin" type="time" />);

    await user.click(screen.getByRole('button', { name: 'Abrir selector de hora' }));

    expect(focus).toHaveBeenCalled();
    expect(focus.mock.instances[0]).toBe(screen.getByLabelText('Hora de fin'));
  });

  it('si showPicker() lanza (navegador que lo bloquea) cae a focus()', async () => {
    const showPicker = vi.fn(() => {
      throw new Error('NotAllowedError');
    });
    definirShowPicker(showPicker);
    const focus = vi.spyOn(HTMLInputElement.prototype, 'focus');
    const user = userEvent.setup();
    render(<Field label="Fecha de la actividad" type="date" />);

    await user.click(screen.getByRole('button', { name: 'Abrir calendario' }));

    expect(showPicker).toHaveBeenCalledTimes(1);
    expect(focus).toHaveBeenCalled();
  });

  it('los campos que no son de fecha/hora no añaden botón', () => {
    render(<Field label="Nombre" />);

    expect(screen.queryByRole('button')).toBeNull();
  });

  it('el botón se activa con teclado (Enter/Espacio) como cualquier botón', async () => {
    const showPicker = vi.fn();
    definirShowPicker(showPicker);
    const user = userEvent.setup();
    render(<Field label="Fecha" type="date" />);

    await user.tab(); // primer tab-stop: lleva al input
    await user.tab(); // segundo: el botón del selector
    expect(screen.getByRole('button', { name: 'Abrir calendario' })).toHaveFocus();

    await user.keyboard('{Enter}');
    expect(showPicker).toHaveBeenCalledTimes(1);
  });

  it('el input deshabilitado deshabilita también su botón', () => {
    render(<Field label="Fecha" type="date" disabled />);

    expect(screen.getByRole('button', { name: 'Abrir calendario' })).toBeDisabled();
  });

  it('mantiene el ref reenviado al input (los consumidores con React Hook Form no se rompen)', () => {
    const ref = createRef<HTMLInputElement>();
    render(<Field label="Fecha" type="date" ref={ref} />);

    expect(ref.current).toBe(screen.getByLabelText('Fecha'));
    expect(ref.current?.type).toBe('date');
  });
});
