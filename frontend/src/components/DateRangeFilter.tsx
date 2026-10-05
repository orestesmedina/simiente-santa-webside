import { Button } from './Button';
import { Field } from './Field';

export interface DateRangeFilterProps {
  from: string;
  to: string;
  onFromChange: (value: string) => void;
  onToChange: (value: string) => void;
  /** Aplica el rango (botón "Filtrar"). */
  onApply: () => void;
  /** Vacía el rango (botón "Quitar filtros"). */
  onClear: () => void;
  applyLabel?: string;
  clearLabel?: string;
}

/**
 * Filtro de rango de fechas "Desde"/"Hasta" con sus botones "Filtrar" y
 * "Quitar filtros" (ux.md §3.10). Envía con `Enter` a través del formulario.
 */
export function DateRangeFilter({
  from,
  to,
  onFromChange,
  onToChange,
  onApply,
  onClear,
  applyLabel = 'Filtrar',
  clearLabel = 'Quitar filtros',
}: DateRangeFilterProps) {
  return (
    <form
      className="flex flex-wrap items-end gap-3"
      onSubmit={(event) => {
        event.preventDefault();
        onApply();
      }}
    >
      <Field
        label="Desde"
        type="date"
        value={from}
        onChange={(event) => onFromChange(event.target.value)}
      />
      <Field
        label="Hasta"
        type="date"
        value={to}
        onChange={(event) => onToChange(event.target.value)}
      />
      <Button type="submit">{applyLabel}</Button>
      <Button type="button" variant="secondary" onClick={onClear}>
        {clearLabel}
      </Button>
    </form>
  );
}
