import type { ReactNode } from 'react';

export interface Column<T> {
  key: string;
  header: string;
  render?: (row: T) => ReactNode;
}

export interface TableProps<T> {
  columns: Column<T>[];
  rows: T[];
  rowKey: (row: T) => string;
  /** Descripción accesible de la tabla (`<caption>` visualmente oculta). */
  caption?: string;
}

/**
 * Tabla accesible: `<caption>`, cabeceras `<th scope="col">` y, en móvil
 * (ux.md §4.b), la misma información como lista de tarjetas etiquetadas. Ambas
 * representaciones comparten datos y columnas para no duplicar markup en las
 * features.
 */
export function Table<T>({ columns, rows, rowKey, caption }: TableProps<T>) {
  return (
    <>
      <table className="hidden w-full border-collapse text-left md:table">
        {caption && <caption className="sr-only">{caption}</caption>}
        <thead>
          <tr>
            {columns.map((column) => (
              <th
                key={column.key}
                scope="col"
                className="border-b border-slate-300 px-3 py-2 font-semibold text-slate-900"
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={rowKey(row)} className="border-b border-slate-200 align-top">
              {columns.map((column) => (
                <td key={column.key} className="px-3 py-3">
                  {column.render ? column.render(row) : null}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>

      <ul className="space-y-3 md:hidden">
        {rows.map((row) => (
          <li key={rowKey(row)} className="rounded border border-slate-200 p-4">
            {columns.map((column) => (
              <div
                key={column.key}
                className="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1 py-1"
              >
                <span className="font-medium text-slate-700">{column.header}</span>
                <span>{column.render ? column.render(row) : null}</span>
              </div>
            ))}
          </li>
        ))}
      </ul>
    </>
  );
}
