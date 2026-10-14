import { Field } from '../../../components/Field';
import { Select } from '../../../components/Select';
import { CONTACT_EMAIL, CONTACT_HELP, CONTACT_NONE, CONTACT_PHONE } from '../messages';
import type { ContactKind } from '../validation';

export interface ContactoEncargadoProps {
  /** Medio elegido: `none` (por defecto), `phone` o `email` (Q8). */
  kind: ContactKind;
  /** Valor del único campo de contacto. */
  value: string;
  error?: string;
  onKindChange: (kind: ContactKind) => void;
  onValueChange: (value: string) => void;
}

const KIND_OPTIONS = [
  { value: 'none', label: CONTACT_NONE },
  { value: 'phone', label: CONTACT_PHONE },
  { value: 'email', label: CONTACT_EMAIL },
];

/**
 * Medio de contacto del encargado (ux.md §4.5, Q8): un `Select` con «— Sin
 * contacto —» (por defecto) / «Celular» / «Correo» y **un único** campo
 * dinámico debajo que cambia de etiqueta y de validación.
 *
 * **Nunca se guardan los dos medios**: al cambiar el tipo, el valor anterior se
 * descarta (el cambio de tipo se hace en este componente para que la invariante
 * viva en un solo lugar). El contacto es opcional aunque haya encargado.
 */
export function ContactoEncargado({
  kind,
  value,
  error,
  onKindChange,
  onValueChange,
}: ContactoEncargadoProps) {
  const handleKindChange = (next: ContactKind) => {
    // El valor anterior se reemplaza por el nuevo campo vacío (Q8).
    onValueChange('');
    onKindChange(next);
  };

  return (
    <div className="flex flex-col gap-3">
      <Select
        id="ministry-contact-kind"
        label="Medio de contacto (opcional)"
        options={KIND_OPTIONS}
        value={kind}
        help={CONTACT_HELP}
        onChange={(event) => handleKindChange(event.target.value as ContactKind)}
      />

      {kind !== 'none' && (
        <Field
          id="ministry-contact-value"
          label={
            kind === 'phone' ? 'Número de celular (opcional)' : 'Correo electrónico (opcional)'
          }
          type={kind === 'email' ? 'email' : 'tel'}
          value={value}
          error={error}
          onChange={(event) => onValueChange(event.target.value)}
        />
      )}
    </div>
  );
}
