import { useRoutes, type RouteObject } from 'react-router-dom';
import { StatusPage } from '../features/status/pages/StatusPage';
import { AppLayout } from './layout';

const appRoutes: RouteObject[] = [
  {
    path: '/',
    element: <AppLayout />,
    children: [{ index: true, element: <StatusPage /> }],
  },
];

export function AppRoutes() {
  return useRoutes(appRoutes);
}
