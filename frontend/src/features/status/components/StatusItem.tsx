import type { VerificationValue } from '../estado';

interface StatusItemProps {
  nombre: string;
  valor: VerificationValue;
}

const VALUE_LABELS: Record<VerificationValue, string> = {
  'servidor-ok': 'en marcha',
  'bd-ok': 'conectada',
  'bd-error': 'no conectada',
  'sin-respuesta': 'sin respuesta',
  'no-comprobable': 'no se pudo comprobar',
};

/** Una fila del detalle de verificación: término (`<dt>`) + palabra de estado (`<dd>`). */
export function StatusItem({ nombre, valor }: StatusItemProps) {
  return (
    <div className="flex justify-between gap-4 py-1">
      <dt className="font-medium">{nombre}:</dt>
      <dd>{VALUE_LABELS[valor]}</dd>
    </div>
  );
}
