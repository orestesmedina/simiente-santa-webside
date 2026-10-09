import { forwardRef, useId, type TextareaHTMLAttributes } from 'react';

export interface TextAreaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label: string;
  error?: string;
  help?: string;
}

/**
 * Área de texto accesible del panel (misma apariencia que `Field`): etiqueta
 * visible, ayuda y error asociados por `aria-describedby`. La usan «quiénes
 * somos» y los textos largos de la identidad.
 */
export const TextArea = forwardRef<HTMLTextAreaElement, TextAreaProps>(function TextArea(
  { label, error, help, id, required, className = '', ...rest },
  ref,
) {
  const generatedId = useId();
  const fieldId = id ?? generatedId;
  const helpId = help ? `${fieldId}-help` : undefined;
  const errorId = error ? `${fieldId}-error` : undefined;
  const describedBy = [helpId, errorId].filter(Boolean).join(' ') || undefined;

  const classes = [
    'min-h-24 w-full rounded border px-3 py-2 text-slate-900',
    'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900',
    'disabled:cursor-not-allowed disabled:bg-slate-100',
    error ? 'border-red-700' : 'border-slate-300',
    className,
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={fieldId} className="font-medium text-slate-900">
        {label}
        {required && (
          <span aria-hidden="true" className="text-red-700">
            {' '}
            *
          </span>
        )}
      </label>
      <textarea
        {...rest}
        id={fieldId}
        ref={ref}
        required={required}
        aria-invalid={error ? true : undefined}
        aria-describedby={describedBy}
        className={classes}
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
