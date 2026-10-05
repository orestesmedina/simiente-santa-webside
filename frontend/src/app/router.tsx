import type { ReactNode } from 'react';
import { Link, useRoutes, type RouteObject } from 'react-router-dom';
import { ChangePasswordPage } from '../features/auth/pages/ChangePasswordPage';
import { LoginPage } from '../features/auth/pages/LoginPage';
import { StatusPage } from '../features/status/pages/StatusPage';
import { UsersPage } from '../features/usuarios/pages/UsersPage';
import { ADMIN_USERS_ROLES } from '../lib/permissions';
import { RequireAuth, RequirePasswordChange, RequirePermission } from './guards';
import { AppLayout, PanelLayout } from './layout';

/**
 * Pantallas provisionales de F2 que T247–T250 y T254 sustituyen por las páginas
 * de `features/`; existen aquí para que las 7 rutas de P18 y sus guards sigan
 * navegables hasta que se implementen (T244). El acceso y el cambio de
 * contraseña ya son las páginas reales de `features/auth` (T245/T246).
 */
function PendingPage({ title, children }: { title: string; children?: ReactNode }) {
  return (
    <section className="space-y-4">
      <h1 className="text-2xl font-bold">{title}</h1>
      {children}
    </section>
  );
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
  { path: '/login', element: <LoginPage /> },
  { path: '/sin-permiso', element: <ForbiddenPending /> },
  {
    element: <RequireAuth />,
    children: [
      { path: '/cambiar-contrasena', element: <ChangePasswordPage /> },
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
                  { path: 'usuarios', element: <UsersPage /> },
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
