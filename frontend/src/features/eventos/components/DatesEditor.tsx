import { Button } from '../../../components/Button';
import { Field } from '../../../components/Field';
import { Notice } from '../../../components/Notice';
import { DATES_HELP } from '../messages';
import { EMPTY_DATE_ROW, type DateRow, type DateRowErrors } from '../validation';

export interface DatesEditorProps {
  rows: DateRow[];
  onChange: (rows: DateRow[]) => void;
  /** Errores por índice de fila (cliente o servidor). */
  errors?: Record<number, DateRowErrors>;
  /** Aviso del bloque completo (sin fechas, tope o fecha repetida). */
  generalError?: string;
}

/**
 * Editor de la lista de fechas del evento (ux.md §4.4, FR-004/FR-015/⟲ D-4):
 * filas de `[fecha] [hora inicio opcional] [hora fin opcional]` con inputs
 * **nativos** (`date`/`time`, 0 errores de parseo y compatibles con móviles),
 * «Agregar fecha» y quitar por fila. Cada fila se valida por separado (fecha
 * real, `HH:MM`, **la hora de fin exige la de inicio** y es posterior —
 * decisión firme de la puerta 2) y el bloque exige al menos una fila.
 */
export function DatesEditor({ rows, onChange, errors = {}, generalError }: DatesEditorProps) {
  const update = (index: number, patch: Partial<DateRow>) =>
    onChange(rows.map((row, position) => (position === index ? { ...row, ...patch } : row)));

  const remove = (index: number) => onChange(rows.filter((_, position) => position !== index));

  const add = () => onChange([...rows, { ...EMPTY_DATE_ROW }]);

  return (
    <fieldset className="space-y-3 rounded border border-slate-200 p-4">
      <legend className="px-1 font-medium text-slate-900">Fechas del evento</legend>
      <p className="text-sm text-slate-600">{DATES_HELP}</p>

      {generalError && <Notice variant="error">{generalError}</Notice>}

      {rows.map((row, index) => (
        <div
          key={index}
          className="grid gap-y-3 border-b border-slate-200 pb-3 last:border-b-0 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)_auto] sm:grid-rows-[auto_auto_auto] sm:gap-x-3 sm:gap-y-1"
        >
          {/* `minmax(0,1fr)` (QA Bug 3): las columnas pueden encogerse por
              debajo del ancho mínimo de los inputs nativos, así que la fila
              nunca desborda el ancho del diálogo.

              Alineación con y sin error (QA ⟲, 3.ª y 4.ª iteración): la fila es
              una rejilla de **tres pistas** (etiqueta / control / mensaje) y
              cada `Field` se integra como `subgrid` (`subgridRow`). Así la
              etiqueta («Hora de inicio (opcional)» se parte en dos líneas),
              el control y el botón con icono caen siempre en la misma pista que
              los de los campos vecinos: los tres iconos comparten altura aunque
              las etiquetas ocupen distinto o aunque un mensaje de error (que se
              pinta en la tercera pista) aparezca solo bajo un campo. Alinear
              por arriba o por abajo (`sm:items-end`) movía los controles en
              cuanto una celda crecía con el mensaje. En escritorio
              `sm:gap-y-1` conserva el ritmo vertical compacto de 4 px entre
              pistas; en móvil la fila apila en una columna con `gap-y-3`
              (12 px) entre campos, porque el `subgrid` (prefijado con `sm:`)
              no aplica y los campos deben conservar su separación de siempre. */}
          <Field
            subgridRow
            id={`agenda-date-${index}`}
            label={`Fecha ${index + 1}`}
            type="date"
            required
            value={row.date}
            error={errors[index]?.date}
            onChange={(event) => update(index, { date: event.target.value })}
          />
          <Field
            subgridRow
            label="Hora de inicio (opcional)"
            type="time"
            value={row.startTime}
            error={errors[index]?.startTime}
            onChange={(event) => update(index, { startTime: event.target.value })}
          />
          <Field
            subgridRow
            label="Hora de fin (opcional)"
            type="time"
            value={row.endTime}
            error={errors[index]?.endTime}
            onChange={(event) => update(index, { endTime: event.target.value })}
          />
          <Button
            variant="secondary"
            size="sm"
            // Pista del control (2.ª fila y 4.ª columna) para alinearse con los
            // inputs. La posición debe ser explícita en **ambos** ejes: un
            // elemento con fila fija se coloca antes que los automáticos y
            // ocuparía la primera columna, desplazando los campos.
            className="sm:col-start-4 sm:row-start-2 sm:self-end"
            aria-label={`Quitar la fecha ${index + 1}`}
            onClick={() => remove(index)}
          >
            Quitar
          </Button>
        </div>
      ))}

      <Button variant="secondary" onClick={add}>
        Agregar fecha
      </Button>
    </fieldset>
  );
}
