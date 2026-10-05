import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom';
import { logout } from '../api/auth';
import { Button } from '../components/Button';
import { ADMIN_USERS_ROLES, hasPermission } from '../lib/permissions';
import { SESSION_QUERY_KEY, useSessionQuery } from './session';

/**
 * Layout mínimo y semántico de la aplicación pública: encabezado con navegación
 * principal y región `main` donde renderiza la ruta activa. El foco visible
 * se deja al estilo por defecto del navegador (accesibilidad de la skill).
 */
export function AppLayout() {
  return (
    <div className="flex min-h-screen flex-col bg-white text-slate-900">
      <header className="border-b border-slate-200">
        <nav aria-label="Navegación principal" className="mx-auto w-full max-w-3xl px-4 py-3">
          <Link
            to="/"
            className="rounded text-sm font-medium text-slate-700 underline-offset-4 hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900"
          >
            Inicio
          </Link>
        </nav>
      </header>
      <main className="mx-auto w-full max-w-3xl flex-1 px-4 py-8">
        <Outlet />
      </main>
    </div>
  );
}

interface NavItem {
  to: string;
  label: string;
  end?: boolean;
}

/**
 * Layout del panel autenticado (ux.md §3.2): barra superior con menú móvil,
 * navegación filtrada por permisos (Usuarios/Roles/Auditoría solo con
 * `admin_usuarios_roles`; una cuenta sin permisos de módulo ve solo Inicio),
 * identificación de la persona y "Salir".
 */
export function PanelLayout() {
  const [menuOpen, setMenuOpen] = useState(false);
  const { data: session } = useSessionQuery();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const logoutMutation = useMutation({
    mutationFn: logout,
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: SESSION_QUERY_KEY });
      navigate('/login', { replace: true });
    },
  });

  const items: NavItem[] = [{ to: '/panel', label: 'Inicio', end: true }];
  if (hasPermission(session, ADMIN_USERS_ROLES)) {
    items.push(
      { to: '/panel/usuarios', label: 'Usuarios' },
      { to: '/panel/roles', label: 'Roles' },
      { to: '/panel/auditoria', label: 'Auditoría' },
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-white text-slate-900">
      <header className="border-b border-slate-200">
        <div className="flex items-center justify-between px-4 py-3">
          <p className="font-semibold">Panel</p>
          <Button
            className="md:hidden"
            aria-expanded={menuOpen}
            aria-controls="panel-nav"
            onClick={() => setMenuOpen((open) => !open)}
          >
            Menú
          </Button>
        </div>
        <nav
          id="panel-nav"
          aria-label="Navegación del panel"
          className={`${menuOpen ? 'block' : 'hidden'} px-4 pb-3 md:block`}
        >
          <ul className="flex flex-col gap-1 md:flex-row md:gap-4">
            {items.map((item) => (
              <li key={item.to}>
                <NavLink
                  to={item.to}
                  end={item.end}
                  onClick={() => setMenuOpen(false)}
                  className={({ isActive }) =>
                    [
                      'inline-flex min-h-11 items-center rounded px-2 font-medium',
                      'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900',
                      isActive
                        ? 'text-slate-900 underline underline-offset-4'
                        : 'text-slate-600 hover:text-slate-900',
                    ].join(' ')
                  }
                >
                  {item.label}
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>
      </header>

      <main className="mx-auto w-full max-w-4xl flex-1 px-4 py-8">
        <Outlet />
      </main>

      <footer className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 px-4 py-3">
        <p className="text-sm text-slate-700">
          {session ? `${session.firstName} ${session.lastName} · ${session.roleName}` : null}
        </p>
        <Button
          variant="secondary"
          loading={logoutMutation.isPending}
          onClick={() => logoutMutation.mutate()}
        >
          Salir
        </Button>
      </footer>
    </div>
  );
}
