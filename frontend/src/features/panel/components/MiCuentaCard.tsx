import { Link } from 'react-router-dom';
import type { SessionUser } from '../../../api/auth';
import { StatusPill } from '../../../components/StatusPill';

export interface MiCuentaCardProps {
  session: SessionUser;
}

/**
 * Tarjeta "Mi cuenta" del Inicio del panel (ux.md §3.3): datos de la sesión
 * (nombre, apellidos, correo y rol) y enlace al cambio de contraseña. Reutiliza
 * `StatusPill` para confirmar que la cuenta está activa (nunca solo color).
 */
export function MiCuentaCard({ session }: MiCuentaCardProps) {
  return (
    <section
      aria-labelledby="mi-cuenta-title"
      className="rounded border border-slate-200 p-4 md:p-6"
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h2 id="mi-cuenta-title" className="text-xl font-bold text-slate-900">
          Mi cuenta
        </h2>
        <StatusPill value="active" label="Cuenta activa" />
      </div>

      <dl className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div>
          <dt className="text-sm font-medium text-slate-700">Nombre</dt>
          <dd className="text-slate-900">{session.firstName}</dd>
        </div>
        <div>
          <dt className="text-sm font-medium text-slate-700">Apellidos</dt>
          <dd className="text-slate-900">{session.lastName}</dd>
        </div>
        <div>
          <dt className="text-sm font-medium text-slate-700">Correo</dt>
          <dd className="break-words text-slate-900">{session.email}</dd>
        </div>
        <div>
          <dt className="text-sm font-medium text-slate-700">Rol</dt>
          <dd className="text-slate-900">{session.roleName}</dd>
        </div>
      </dl>

      <p className="mt-4">
        <Link
          to="/cambiar-contrasena"
          className="inline-flex min-h-11 items-center rounded font-medium text-slate-900 underline underline-offset-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900"
        >
          Cambiar contraseña
        </Link>
      </p>
    </section>
  );
}
