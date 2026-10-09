import type { ReactNode } from 'react';
import { Link, useRoutes, type RouteObject } from 'react-router-dom';
import { ChangePasswordPage } from '../features/auth/pages/ChangePasswordPage';
import { LoginPage } from '../features/auth/pages/LoginPage';
import { AuditPage } from '../features/auditoria/pages/AuditPage';
import { InicioPage } from '../features/panel/pages/InicioPage';
import { RolesPage } from '../features/roles/pages/RolesPage';
import { StatusPage } from '../features/status/pages/StatusPage';
import { UsersPage } from '../features/usuarios/pages/UsersPage';
import { ADMIN_USERS_ROLES, PORTADA } from '../lib/permissions';
import { RequireAuth, RequirePasswordChange, RequirePermission } from './guards';
import { AppLayout, PanelLayout } from './layout';

/**
 * Pantalla provisional de `/sin-permiso` (T244). Las páginas de panel (`/panel`,
 * `/panel/roles`, `/panel/auditoria`) ya son las reales de `features/` (T249,
 * T250 y T254); el acceso y el cambio de contraseña se implementaron en T245/T246.
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

/**
 * Marcador provisional de la portada pública (T331). T332 sustituye esta ruta
 * por `PublicLayout` + `HomePage` con el contenido publicado. Hasta entonces la
 * ruta `/` queda montada y separada de la pantalla de estado.
 */
function PublicHomePlaceholder() {
  return (
    <main className="mx-auto w-full max-w-3xl px-4 py-8">
      <h1 className="text-2xl font-bold">Portada</h1>
    </main>
  );
}

/**
 * Marcador provisional del módulo del panel (T331). T333 lo sustituye por
 * `InformationPage` (pestañas de las 6 piezas). La ruta ya está protegida por
 * `RequirePermission code={PORTADA}`.
 */
function InformationPlaceholder() {
  return (
    <section className="space-y-4">
      <h1 className="text-2xl font-bold">Portada e información general</h1>
    </section>
  );
}

const appRoutes: RouteObject[] = [
  { path: '/', element: <PublicHomePlaceholder /> },
  {
    path: '/health',
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
              { index: true, element: <InicioPage /> },
              {
                element: <RequirePermission code={ADMIN_USERS_ROLES} />,
                children: [
                  { path: 'usuarios', element: <UsersPage /> },
                  { path: 'roles', element: <RolesPage /> },
                  { path: 'auditoria', element: <AuditPage /> },
                ],
              },
              {
                element: <RequirePermission code={PORTADA} />,
                children: [{ path: 'informacion', element: <InformationPlaceholder /> }],
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
