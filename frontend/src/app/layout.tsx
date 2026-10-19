import { useState, type ReactNode } from 'react';
import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom';
import { Button } from '../components/Button';
import { SessionWarning } from '../features/auth/components/SessionWarning';
import { useLogout } from '../features/auth/hooks/useLogout';
import { useSessionQuery } from '../features/auth/hooks/useSession';
import {
  ACTIVIDADES,
  ADMIN_USERS_ROLES,
  CATEGORIAS,
  DONACIONES,
  EVENTOS,
  MEDIOS,
  MINISTERIOS,
  PORTADA,
  hasPermission,
} from '../lib/permissions';

/**
 * Layout mínimo y semántico de la aplicación pública: encabezado con navegación
 * principal y región `main` donde renderiza la ruta activa. El foco visible
 * se deja al estilo por defecto del navegador (accesibilidad de la skill).
 *
 * La superficie es siempre clara (`bg-white`), así que fija `color-scheme: light`
 * (`[color-scheme:light]`) para que los controles nativos —p. ej. el indicador
 * de `<input type="date">`/`type="time"`— no hereden el `dark` global de
 * `<html>` (F10) y se pinten en claro sobre blanco (invisibles).
 */
export function AppLayout() {
  return (
    <div className="flex min-h-screen flex-col bg-white text-slate-900 [color-scheme:light]">
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

/**
 * Superficie clara de las pantallas del panel que viven **fuera** de
 * `PanelLayout` (acceso, cambio de contraseña, «sin permiso» y el 404): el panel
 * es siempre claro y no cambia de tema (FR-016/P10-11), así que estas rutas
 * fuerzan la **misma** superficie que `AppLayout`/`PanelLayout`
 * (`bg-white text-slate-900`).
 *
 * Sin ella, el lienzo del navegador se ve por detrás: el script anti-FOUC de
 * `index.html` fija `color-scheme: dark` en `<html>` cuando el tema es oscuro y
 * el navegador pinta el fondo del documento en negro, dejando la pantalla
 * ilegible (texto oscuro sobre negro). No se resuelve con un fondo global en
 * `body`: el sitio público necesita su lienzo oscuro (su layout pinta encima).
 *
 * Además fija `color-scheme: light` (`[color-scheme:light]`) en la superficie:
 * sin ese reset el `dark` de `<html>` se hereda y el navegador pinta el
 * indicador nativo de `<input type="date">`/`type="time"` en color claro sobre
 * el panel blanco (invisible). El tema del sitio público no se toca: el reset
 * solo aplica al subárbol del panel.
 */
export function PanelSurfaceLayout({ children }: { children?: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col bg-white text-panel-ink [color-scheme:light]">
      {children ?? <Outlet />}
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
  const logoutMutation = useLogout();

  const handleLogout = () => {
    logoutMutation.mutate(undefined, {
      onSuccess: () => navigate('/login', { replace: true }),
    });
  };

  const items: NavItem[] = [{ to: '/panel', label: 'Inicio', end: true }];
  if (hasPermission(session, PORTADA)) {
    items.push({ to: '/panel/informacion', label: 'Portada e información general' });
  }
  if (hasPermission(session, EVENTOS)) {
    items.push({ to: '/panel/eventos', label: 'Gestión de eventos' });
  }
  if (hasPermission(session, ACTIVIDADES)) {
    items.push({ to: '/panel/actividades', label: 'Gestión de actividades' });
  }
  if (hasPermission(session, CATEGORIAS)) {
    items.push({ to: '/panel/categorias', label: 'Categorías' });
  }
  if (hasPermission(session, MINISTERIOS)) {
    items.push({ to: '/panel/ministerios', label: 'Gestión de ministerios' });
  }
  if (hasPermission(session, DONACIONES)) {
    items.push({ to: '/panel/donaciones', label: 'Gestión de donaciones' });
  }
  if (hasPermission(session, MEDIOS)) {
    items.push({ to: '/panel/medios', label: 'Gestión de medios' });
  }
  if (hasPermission(session, ADMIN_USERS_ROLES)) {
    items.push(
      { to: '/panel/usuarios', label: 'Usuarios' },
      { to: '/panel/roles', label: 'Roles' },
      { to: '/panel/auditoria', label: 'Auditoría' },
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-white text-slate-900 [color-scheme:light]">
      <SessionWarning />
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
        <Button variant="secondary" loading={logoutMutation.isPending} onClick={handleLogout}>
          Salir
        </Button>
      </footer>
    </div>
  );
}
