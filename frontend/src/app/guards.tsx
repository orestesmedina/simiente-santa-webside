import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { Button } from '../components/Button';
import { EmptyState } from '../components/EmptyState';
import { hasPermission } from '../lib/permissions';
import { isUnauthenticated, useSessionQuery } from './session';

function SessionLoading() {
  return (
    <div className="mx-auto max-w-lg p-6">
      <EmptyState kind="loading" message="Comprobando tu acceso…" />
    </div>
  );
}

/**
 * Exige sesión iniciada (FR-001/SC-001). Sin sesión o con `401` redirige a
 * `/login` indicando el `destino` para volver tras entrar. Un fallo que no sea
 * `401` (p. ej. red) muestra un aviso con "Reintentar" en lugar de expulsar.
 */
export function RequireAuth() {
  const query = useSessionQuery();
  const location = useLocation();

  if (query.isPending) {
    return <SessionLoading />;
  }

  if (query.isError) {
    if (isUnauthenticated(query.error)) {
      const destino = `${location.pathname}${location.search}`;
      const params = new URLSearchParams({ motivo: 'sin-sesion', destino });
      return <Navigate to={`/login?${params.toString()}`} replace />;
    }
    return (
      <div className="mx-auto max-w-lg p-6">
        <EmptyState
          kind="error"
          message="No se pudo verificar tu acceso. Vuelve a intentarlo en unos minutos."
          action={<Button onClick={() => void query.refetch()}>Reintentar</Button>}
        />
      </div>
    );
  }

  return <Outlet />;
}

/**
 * Obliga a cambiar la contraseña antes de usar el panel (FR-010, US7 esc. 4):
 * con `mustChangePassword` redirige a `/cambiar-contrasena`.
 */
export function RequirePasswordChange() {
  const query = useSessionQuery();

  if (query.isSuccess && query.data.mustChangePassword) {
    return <Navigate to="/cambiar-contrasena" replace />;
  }

  return <Outlet />;
}

/**
 * Exige un permiso del catálogo (FR-016). Sin él redirige a `/sin-permiso`; no
 * adivina ni muestra la sección (la autoridad real es el servidor).
 */
export function RequirePermission({ code }: { code: string }) {
  const query = useSessionQuery();

  if (query.isSuccess && !hasPermission(query.data, code)) {
    return <Navigate to="/sin-permiso" replace />;
  }

  return <Outlet />;
}
