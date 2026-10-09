import { forwardRef, useId, useState, type InputHTMLAttributes } from 'react';

/** Un requisito de la política de contraseñas (FR-010) con su estado en vivo. */
export interface PasswordPolicyCheck {
  id: string;
  label: string;
  met: boolean;
}

export interface PasswordFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label: string;
  error?: string;
  help?: string;
  /** Checklist en vivo de la política; se muestra bajo el campo. */
  policy?: PasswordPolicyCheck[];
}

/**
 * Campo de contraseña con botón mostrar/ocultar (`aria-pressed`, accesible por
 * teclado) y checklist en vivo de la política FR-010. Nunca persiste el valor:
 * el estado vive solo en el formulario. `autocomplete` lo fija quien lo usa
 * (`current-password` o `new-password`).
 */
export const PasswordField = forwardRef<HTMLInputElement, PasswordFieldProps>(
  function PasswordField(
    { label, error, help, policy, id, required, className = '', ...rest },
    ref,
  ) {
    const [visible, setVisible] = useState(false);
    const generatedId = useId();
    const inputId = id ?? generatedId;
    const helpId = help ? `${inputId}-help` : undefined;
    const errorId = error ? `${inputId}-error` : undefined;
    const policyId = policy && policy.length > 0 ? `${inputId}-policy` : undefined;
    const describedBy = [helpId, errorId, policyId].filter(Boolean).join(' ') || undefined;

    const inputClasses = [
      'min-h-12 w-full rounded border py-2 pl-3 pr-28 text-slate-900',
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
        <div className="relative">
          <input
            {...rest}
            id={inputId}
            ref={ref}
            type={visible ? 'text' : 'password'}
            required={required}
            aria-invalid={error ? true : undefined}
            aria-describedby={describedBy}
            className={inputClasses}
          />
          <button
            type="button"
            aria-pressed={visible}
            aria-label={visible ? 'Ocultar contraseña' : 'Mostrar contraseña'}
            onClick={() => setVisible((current) => !current)}
            className="absolute inset-y-0.5 right-0.5 rounded border border-slate-300 bg-white px-3 text-sm font-medium text-slate-900 hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900"
          >
            {visible ? 'Ocultar' : 'Mostrar'}
          </button>
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
        {policy && policy.length > 0 && (
          <ul
            id={policyId}
            aria-label="Requisitos de la contraseña"
            aria-live="polite"
            className="mt-1 space-y-1 text-sm"
          >
            {policy.map((item) => (
              <li key={item.id} className={item.met ? 'text-green-800' : 'text-slate-600'}>
                <span aria-hidden="true">{item.met ? '✔' : '○'}</span> {item.label}
                <span className="sr-only">: {item.met ? 'cumplido' : 'pendiente'}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
    );
  },
);
