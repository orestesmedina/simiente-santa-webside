import { Link } from 'react-router-dom';
import type { SessionUser } from '../../../api/auth';
import { ADMIN_USERS_ROLES, PORTADA, hasPermission } from '../../../lib/permissions';

interface QuickAccess {
  to: string;
  label: string;
  code: string;
}

const SECTIONS: QuickAccess[] = [
  { to: '/panel/informacion', label: 'Portada e información general', code: PORTADA },
  { to: '/panel/usuarios', label: 'Usuarios', code: ADMIN_USERS_ROLES },
  { to: '/panel/roles', label: 'Roles', code: ADMIN_USERS_ROLES },
  { to: '/panel/auditoria', label: 'Auditoría', code: ADMIN_USERS_ROLES },
];

export interface AccesosRapidosProps {
  session: SessionUser;
}

/**
 * Accesos rápidos del Inicio (ux.md §3.3): **solo** las secciones autorizadas
 * por permiso (`hasPermission`, FR-016). Sin ninguna autorizada no pinta nada;
 * el Inicio muestra entonces el aviso de "cuenta sin permisos de módulo".
 */
export function AccesosRapidos({ session }: AccesosRapidosProps) {
  const authorized = SECTIONS.filter((section) => hasPermission(session, section.code));
  if (authorized.length === 0) {
    return null;
  }

  return (
    <section
      aria-labelledby="accesos-rapidos-title"
      className="rounded border border-slate-200 p-4"
    >
      <h2 id="accesos-rapidos-title" className="text-xl font-bold text-slate-900">
        Gestión
      </h2>
      <nav aria-label="Accesos rápidos" className="mt-3">
        <ul className="flex flex-wrap gap-3">
          {authorized.map((section) => (
            <li key={section.to}>
              <Link
                to={section.to}
                className="inline-flex min-h-11 items-center rounded border border-slate-300 bg-white px-4 py-2 font-medium text-slate-900 hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900"
              >
                {section.label}
              </Link>
            </li>
          ))}
        </ul>
      </nav>
    </section>
  );
}
