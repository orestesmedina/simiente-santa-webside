import type { ReactNode } from 'react';
import { Link, useRoutes, type RouteObject } from 'react-router-dom';
import { StatusPage } from '../features/status/pages/StatusPage';
import { ADMIN_USERS_ROLES } from '../lib/permissions';
import { RequireAuth, RequirePasswordChange, RequirePermission } from './guards';
import { AppLayout, PanelLayout } from './layout';

/**
 * Pantallas provisionales de F2. T245–T250 las sustituyen por las páginas de
 * `features/`; existen aquí para que las 7 rutas de P18 y sus guards sean
 * navegables y verificables desde ya (T244).
 */
function PendingPage({ title, children }: { title: string; children?: ReactNode }) {
  return (
    <section className="space-y-4">
      <h1 className="text-2xl font-bold">{title}</h1>
      {children}
    </section>
  );
}

function LoginPending() {
  return <PendingPage title="Entrar al panel" />;
}

function ChangePasswordPending() {
  return <PendingPage title="Cambiar contraseña" />;
}

function ForbiddenPending() {
  return (
    <PendingPage title="No tienes acceso a esta sección">
      <p className="text-slate-700">
        Si necesitas entrar aquí, pide a un administrador de la iglesia que actualice tu rol.
      </p>
      <Link
        to="/panel"
        className="font-medium text-slate-900 underline underline-offset-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900"
      >
        Volver al inicio del panel
      </Link>
    </PendingPage>
  );
}

function PanelHomePending() {
  return <PendingPage title="Inicio" />;
}

function UsersPending() {
  return <PendingPage title="Usuarios" />;
}

function RolesPending() {
  return <PendingPage title="Roles" />;
}

function AuditPending() {
  return <PendingPage title="Auditoría" />;
}

function NotFoundPage() {
  return (
    <main className="mx-auto w-full max-w-3xl px-4 py-8">
      <h1 className="text-2xl font-bold">Página no encontrada</h1>
      <p className="mt-2 text-slate-700">
        La página que buscas no existe.{' '}
        <Link to="/panel" className="underline underline-offset-4">
          Volver al panel
        </Link>
        .
      </p>
    </main>
  );
}

const appRoutes: RouteObject[] = [
  {
    path: '/',
    element: <AppLayout />,
    children: [{ index: true, element: <StatusPage /> }],
  },
  { path: '/login', element: <LoginPending /> },
  { path: '/sin-permiso', element: <ForbiddenPending /> },
  {
    element: <RequireAuth />,
    children: [
      { path: '/cambiar-contrasena', element: <ChangePasswordPending /> },
      {
        element: <RequirePasswordChange />,
        children: [
          {
            path: '/panel',
            element: <PanelLayout />,
            children: [
              { index: true, element: <PanelHomePending /> },
              {
                element: <RequirePermission code={ADMIN_USERS_ROLES} />,
                children: [
                  { path: 'usuarios', element: <UsersPending /> },
                  { path: 'roles', element: <RolesPending /> },
                  { path: 'auditoria', element: <AuditPending /> },
                ],
              },
            ],
          },
        ],
      },
    ],
  },
  { path: '*', element: <NotFoundPage /> },
];

export function AppRoutes() {
  return useRoutes(appRoutes);
}
