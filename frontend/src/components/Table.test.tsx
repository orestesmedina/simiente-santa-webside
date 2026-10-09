import { render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Table, type Column } from './Table';

interface Fila {
  id: string;
  nombre: string;
  rol: string;
}

const columns: Column<Fila>[] = [
  { key: 'nombre', header: 'Nombre', render: (fila) => fila.nombre },
  { key: 'rol', header: 'Rol', render: (fila) => fila.rol },
];

const rows: Fila[] = [{ id: '1', nombre: 'Ana', rol: 'Administración' }];

describe('Table', () => {
  it('renderiza una tabla semántica con cabeceras de columna', () => {
    render(<Table columns={columns} rows={rows} rowKey={(fila) => fila.id} caption="Usuarios" />);
    const tabla = screen.getByRole('table', { name: 'Usuarios' });
    expect(within(tabla).getByRole('columnheader', { name: 'Nombre' })).toBeInTheDocument();
    expect(within(tabla).getByRole('cell', { name: 'Ana' })).toBeInTheDocument();
  });

  it('compone la misma información como tarjetas para móvil', () => {
    render(<Table columns={columns} rows={rows} rowKey={(fila) => fila.id} />);
    expect(screen.getByRole('list')).toBeInTheDocument();
    expect(screen.getAllByText('Ana')).toHaveLength(2);
  });
});
