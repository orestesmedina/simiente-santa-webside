import { render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { DatesEditor } from './DatesEditor';
import type { DateRow } from '../validation';

/**
 * ⟲ 4.ª iteración del bug del date picker del panel: `sm:items-end` alineaba los
 * iconos en el estado normal, pero al aparecer un mensaje de validación bajo un
 * campo esa celda crecía y su control se desplazaba (0 / +44 / −40 px medidos en
 * el diálogo real). La fila pasa a ser una rejilla de **tres pistas** (etiqueta,
 * control, mensaje) y cada campo se integra como `subgrid`, así que el control
 * cae siempre en la misma pista tenga o no mensaje.
 *
 * La alineación real la mide el e2e (`qa-date-picker-panel-alineacion`) porque
 * jsdom no calcula layout; aquí se fija la estructura que la hace posible.
 */
describe('DatesEditor · pistas de la fila (alineación con y sin error)', () => {
  const fila: DateRow[] = [{ date: '2026-10-17', startTime: '', endTime: '21:00' }];
  const raizDelCampo = (etiqueta: RegExp): HTMLElement => {
    const raiz = screen.getByLabelText(etiqueta).closest('div')?.parentElement;
    if (!raiz) throw new Error('el campo no tiene raíz');
    return raiz;
  };

  it('la fila declara tres pistas y los tres campos se integran como subgrid', () => {
    render(<DatesEditor rows={fila} onChange={() => {}} />);

    const campos = [
      raizDelCampo(/^Fecha 1/),
      raizDelCampo(/^Hora de inicio/),
      raizDelCampo(/^Hora de fin/),
    ];
    for (const campo of campos) {
      expect(campo).toHaveClass('sm:row-span-3', 'sm:grid-rows-subgrid');
    }
    // Los tres comparten el mismo contenedor de rejilla (una sola fila).
    const rejilla = campos[0].parentElement;
    expect(rejilla).toBe(campos[1].parentElement);
    expect(rejilla).toBe(campos[2].parentElement);
    expect(rejilla).toHaveClass('sm:grid-rows-[auto_auto_auto]');
  });

  it('con error solo en un campo, el mensaje vive en su propia pista (no mueve el control)', () => {
    render(
      <DatesEditor
        rows={[{ date: '2026-10-17', startTime: '', endTime: '21:00' }]}
        onChange={() => {}}
        errors={{ 0: { endTime: 'Escribe la hora de inicio para poder poner una hora de fin.' } }}
      />,
    );

    const fin = raizDelCampo(/^Hora de fin/);
    const mensaje = screen.getByRole('alert');
    expect(mensaje).toHaveTextContent(
      'Escribe la hora de inicio para poder poner una hora de fin.',
    );
    // El mensaje es el último hijo de la raíz del campo: ocupa la 3.ª pista,
    // después del control, sin empujar el input ni su botón con icono.
    expect(fin.lastElementChild).toBe(mensaje);
    expect(fin.querySelector('input[type="time"]')).not.toBeNull();
    expect(within(fin).getByRole('button', { name: 'Abrir selector de hora' })).toBeInTheDocument();
    // Los campos sin error conservan la misma raíz de subgrid (misma base).
    expect(raizDelCampo(/^Fecha 1/)).toHaveClass('sm:grid-rows-subgrid');
    expect(raizDelCampo(/^Hora de inicio/)).toHaveClass('sm:grid-rows-subgrid');
  });
});
