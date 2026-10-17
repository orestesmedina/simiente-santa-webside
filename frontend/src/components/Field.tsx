import { forwardRef, useId, useRef, type InputHTMLAttributes } from 'react';

export interface FieldProps extends InputHTMLAttributes<HTMLInputElement> {
  /** Etiqueta visible asociada al campo (nunca solo `placeholder`). */
  label: string;
  /** Mensaje de error; activa `aria-invalid` y `aria-describedby`. */
  error?: string;
  /** Texto de ayuda que se asocia al campo. */
  help?: string;
}

/** Tipos nativos con selector del navegador (`showPicker()`). */
const PICKER_LABELS: Record<string, string> = {
  date: 'Abrir calendario',
  time: 'Abrir selector de hora',
  'datetime-local': 'Abrir selector de fecha y hora',
  month: 'Abrir selector de mes',
  week: 'Abrir selector de semana',
};

function CalendarIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className="h-5 w-5"
    >
      <rect x="3" y="4.5" width="18" height="16" rx="2" />
      <path d="M3 9.5h18" />
      <path d="M8 3v3M16 3v3" />
    </svg>
  );
}

function ClockIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className="h-5 w-5"
    >
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7v5l3 2" />
    </svg>
  );
}

/**
 * Campo de formulario accesible: `<label>` explícito, ayuda y error asociados
 * por `aria-describedby` (ux.md §6). Acepta `ref` para integrarse con
 * React Hook Form.
 *
 * Los campos de fecha/hora (`date`/`time`, decisión D-4 de F4) añaden un botón
 * con icono **siempre visible** que abre el selector nativo con
 * `HTMLInputElement.showPicker()` (con `focus()` de respaldo). Se resuelve aquí
 * para que lo hereden todos los formularios del panel (`DatesEditor`,
 * `ActivityForm`, `DateRangeFilter`, horarios de servicios…), porque el icono
 * propio del input nativo no siempre es visible ni tiene un área de toque
 * suficiente en móvil. El input conserva su tipo nativo.
 */
export const Field = forwardRef<HTMLInputElement, FieldProps>(function Field(
  { label, error, help, id, required, disabled, type, className = '', ...rest },
  ref,
) {
  const generatedId = useId();
  const inputId = id ?? generatedId;
  const helpId = help ? `${inputId}-help` : undefined;
  const errorId = error ? `${inputId}-error` : undefined;
  const describedBy = [helpId, errorId].filter(Boolean).join(' ') || undefined;

  const inputRef = useRef<HTMLInputElement | null>(null);

  /** Combina el `ref` reenviado (React Hook Form) con el interno del botón. */
  const setInputRef = (node: HTMLInputElement | null) => {
    inputRef.current = node;
    if (typeof ref === 'function') {
      ref(node);
    } else if (ref) {
      ref.current = node;
    }
  };

  const pickerLabel = type ? PICKER_LABELS[type] : undefined;

  const openPicker = () => {
    const input = inputRef.current;
    if (!input) return;
    if (typeof input.showPicker === 'function') {
      try {
        input.showPicker();
        return;
      } catch {
        // Algunos navegadores lo bloquean (p. ej. sin gesto de usuario o en
        // iframes): se cae al foco del campo.
      }
    }
    input.focus();
  };

  const inputClasses = [
    'min-h-11 w-full rounded border px-3 py-2 text-slate-900',
    'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900',
    'disabled:cursor-not-allowed disabled:bg-slate-100',
    // Con botón propio, el indicador nativo (WebKit/Blink) sobra: dos iconos
    // juntos confunden sobre dónde pulsar. El resto del arreglo de
    // `color-scheme` sigue siendo necesario para el texto y el resto del campo.
    pickerLabel ? 'pr-11 [&::-webkit-calendar-picker-indicator]:hidden' : '',
    error ? 'border-red-700' : 'border-slate-300',
    className,
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={inputId} className="font-medium text-slate-900">
        {label}
        {required && (
          <span aria-hidden="true" className="text-red-700">
            {' '}
            *
          </span>
        )}
      </label>
      {/* Botón superpuesto al borde derecho del campo (el indicador nativo se
          oculta arriba): área de toque ≥44 px, siempre visible y con foco. */}
      <div className="relative">
        <input
          {...rest}
          id={inputId}
          ref={setInputRef}
          type={type}
          required={required}
          disabled={disabled}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
          className={inputClasses}
        />
        {pickerLabel && (
          <button
            type="button"
            aria-label={pickerLabel}
            aria-controls={inputId}
            disabled={disabled}
            onClick={openPicker}
            className={[
              'absolute inset-y-0 right-0 flex w-11 items-center justify-center rounded-r',
              'text-slate-600 hover:bg-slate-100 hover:text-slate-900',
              'focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-slate-900',
              'disabled:cursor-not-allowed disabled:text-slate-400 disabled:hover:bg-transparent',
            ].join(' ')}
          >
            {type === 'time' ? <ClockIcon /> : <CalendarIcon />}
          </button>
        )}
      </div>
      {help && (
        <p id={helpId} className="text-sm text-slate-600">
          {help}
        </p>
      )}
      {error && (
        <p id={errorId} role="alert" className="text-sm font-medium text-red-700">
          {error}
        </p>
      )}
    </div>
  );
});
