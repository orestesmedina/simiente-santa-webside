import { forwardRef, useId, type InputHTMLAttributes } from 'react';

export interface FieldProps extends InputHTMLAttributes<HTMLInputElement> {
  /** Etiqueta visible asociada al campo (nunca solo `placeholder`). */
  label: string;
  /** Mensaje de error; activa `aria-invalid` y `aria-describedby`. */
  error?: string;
  /** Texto de ayuda que se asocia al campo. */
  help?: string;
}

/**
 * Campo de formulario accesible: `<label>` explícito, ayuda y error asociados
 * por `aria-describedby` (ux.md §6). Acepta `ref` para integrarse con
 * React Hook Form.
 */
export const Field = forwardRef<HTMLInputElement, FieldProps>(function Field(
  { label, error, help, id, required, className = '', ...rest },
  ref,
) {
  const generatedId = useId();
  const inputId = id ?? generatedId;
  const helpId = help ? `${inputId}-help` : undefined;
  const errorId = error ? `${inputId}-error` : undefined;
  const describedBy = [helpId, errorId].filter(Boolean).join(' ') || undefined;

  const inputClasses = [
    'min-h-11 w-full rounded border px-3 py-2 text-slate-900',
    'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900',
    'disabled:cursor-not-allowed disabled:bg-slate-100',
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
      <input
        {...rest}
        id={inputId}
        ref={ref}
        required={required}
        aria-invalid={error ? true : undefined}
        aria-describedby={describedBy}
        className={inputClasses}
      />
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
