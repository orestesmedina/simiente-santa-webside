import { useRoutes, type RouteObject } from 'react-router-dom';
import { AppLayout } from './layout';

/**
 * Destino provisional de la ruta inicial. En T024 se sustituye por la página
 * real de la feature `status` (`StatusPage`).
 */
function PendingStatusPage() {
  return <h1>Estado del sistema</h1>;
}

const appRoutes: RouteObject[] = [
  {
    path: '/',
    element: <AppLayout />,
    children: [{ index: true, element: <PendingStatusPage /> }],
  },
];

export function AppRoutes() {
  return useRoutes(appRoutes);
}
