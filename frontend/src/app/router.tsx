import type { ReactNode } from 'react';
import { Link, useRoutes, type RouteObject } from 'react-router-dom';
import { ChangePasswordPage } from '../features/auth/pages/ChangePasswordPage';
import { LoginPage } from '../features/auth/pages/LoginPage';
import { AuditPage } from '../features/auditoria/pages/AuditPage';
import { ActivitiesPage } from '../features/eventos/pages/ActivitiesPage';
import { CategoriesPage } from '../features/eventos/pages/CategoriesPage';
import { EventsPage as PanelEventsPage } from '../features/eventos/pages/EventsPage';
import { DonacionesPage } from '../features/donaciones/pages/DonacionesPage';
import { InformationPage } from '../features/informacion/pages/InformationPage';
import { InicioPage } from '../features/panel/pages/InicioPage';
import { MinistriesPage } from '../features/ministerios/pages/MinistriesPage';
import { MediosPage as PanelMediosPage } from '../features/medios/pages/MediosPage';
import { PublicLayout } from '../features/publico/components/PublicLayout';
import { EventsPage as PublicEventsPage } from '../features/publico/eventos/EventsPage';
import { MinisteriosPage as PublicMinisteriosPage } from '../features/publico/ministerios/MinisteriosPage';
import { MediosPage as PublicMediosPage } from '../features/publico/medios/MediosPage';
import { OfrendasPage as PublicOfrendasPage } from '../features/publico/ofrendas/OfrendasPage';
import { HomePage } from '../features/publico/pages/HomePage';
import { RolesPage } from '../features/roles/pages/RolesPage';
import { StatusPage } from '../features/status/pages/StatusPage';
import { UsersPage } from '../features/usuarios/pages/UsersPage';
import {
  ACTIVIDADES,
  ADMIN_USERS_ROLES,
  CATEGORIAS,
  DONACIONES,
  EVENTOS,
  MEDIOS,
  MINISTERIOS,
  PORTADA,
} from '../lib/permissions';
import { RequireAuth, RequirePasswordChange, RequirePermission } from './guards';
import { AppLayout, PanelLayout, PanelSurfaceLayout } from './layout';

/**
 * Pantalla provisional de `/sin-permiso` (T244). Las páginas de panel (`/panel`,
 * `/panel/roles`, `/panel/auditoria`) ya son las reales de `features/` (T249,
 * T250 y T254); el acceso y el cambio de contraseña se implementaron en T245/T246.
 */
function PendingPage({ title, children }: { title: string; children?: ReactNode }) {
  return (
    <section className="space-y-4">
      <h1 className="text-2xl font-bold text-panel-ink">{title}</h1>
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
      <h1 className="text-2xl font-bold text-panel-ink">Página no encontrada</h1>
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
    element: <PublicLayout />,
    children: [
      { index: true, element: <HomePage /> },
      // La sección de agenda es una **ruta real** (⟲ D-1), no un ancla de la portada.
      { path: 'eventos', element: <PublicEventsPage /> },
      // «Ministerios» es una **ruta real** (⟲ D-1 de F6) y una sola entrada de
      // navegación (T628). Sin publicados redirige a la portada desde la propia
      // página (D-14/P6-21), no desde aquí.
      { path: 'ministerios', element: <PublicMinisteriosPage /> },
      // «Ofrendas y diezmos» es una **ruta real** (⟲ D-1 de F7) y una sola
      // entrada de navegación (T727). Sin nada publicado redirige a la portada
      // desde la propia página (Q9/P7-21), no desde aquí.
      { path: 'ofrendas', element: <PublicOfrendasPage /> },
      // «Medios» es una **ruta real** (⟲ D-1 de F9) y una sola entrada de
      // navegación (T934). Sin nada publicable redirige a la portada desde la
      // propia página (Q10/P9-16), no desde aquí.
      { path: 'medios', element: <PublicMediosPage /> },
    ],
  },
  {
    path: '/health',
    element: <AppLayout />,
    children: [{ index: true, element: <StatusPage /> }],
  },
  // Pantallas del panel **sin su chrome** (acceso, «sin permiso», cambio de
  // contraseña y 404): comparten la superficie clara del panel para que el
  // lienzo del navegador (oscuro con el tema del dispositivo) no se vea por
  // detrás (bug del fondo negro, F4).
  {
    element: <PanelSurfaceLayout />,
    children: [
      { path: '/login', element: <LoginPage /> },
      { path: '/sin-permiso', element: <ForbiddenPending /> },
    ],
  },
  {
    element: <RequireAuth />,
    children: [
      {
        element: <PanelSurfaceLayout />,
        children: [{ path: '/cambiar-contrasena', element: <ChangePasswordPage /> }],
      },
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
                children: [{ path: 'informacion', element: <InformationPage /> }],
              },
              {
                element: <RequirePermission code={EVENTOS} />,
                children: [{ path: 'eventos', element: <PanelEventsPage /> }],
              },
              {
                element: <RequirePermission code={ACTIVIDADES} />,
                children: [{ path: 'actividades', element: <ActivitiesPage /> }],
              },
              {
                // El catálogo se autoriza con su permiso propio (⟲ D-6): ni
                // «eventos» ni «actividades» dan acceso a él (FR-012).
                element: <RequirePermission code={CATEGORIAS} />,
                children: [{ path: 'categorias', element: <CategoriesPage /> }],
              },
              {
                // F6 usa el permiso **ya sembrado** por F2 (P6-6): no se crea
                // ningún permiso nuevo y el catálogo sigue en 10.
                element: <RequirePermission code={MINISTERIOS} />,
                children: [{ path: 'ministerios', element: <MinistriesPage /> }],
              },
              {
                // F7 también usa el permiso **ya sembrado** por F2 (P7-6/Q8):
                // sin permiso nuevo y sin renombrarlo.
                element: <RequirePermission code={DONACIONES} />,
                children: [{ path: 'donaciones', element: <DonacionesPage /> }],
              },
              {
                // F9 usa el permiso **ya sembrado** por F2 (P9-15/FR-021): sin
                // permiso nuevo, sin renombrarlo (Q13) y con el catálogo en 10.
                element: <RequirePermission code={MEDIOS} />,
                children: [{ path: 'medios', element: <PanelMediosPage /> }],
              },
            ],
          },
        ],
      },
    ],
  },
  {
    path: '*',
    element: (
      <PanelSurfaceLayout>
        <NotFoundPage />
      </PanelSurfaceLayout>
    ),
  },
];

export function AppRoutes() {
  return useRoutes(appRoutes);
}
