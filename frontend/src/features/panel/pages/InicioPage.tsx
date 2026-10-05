import { Button } from '../../../components/Button';
import { EmptyState } from '../../../components/EmptyState';
import { Notice } from '../../../components/Notice';
import { useSessionQuery } from '../../auth/hooks/useSession';
import { ADMIN_USERS_ROLES, hasPermission } from '../../../lib/permissions';
import { AccesosRapidos } from '../components/AccesosRapidos';
import { MiCuentaCard } from '../components/MiCuentaCard';

/**
 * Texto del Edge Case de la spec/ux.md §3.3: una cuenta autenticada sin ningún
 * permiso de módulo entra al panel y solo ve "Mi cuenta"; esto es válido y no
 * rompe la página ni deja secciones vacías a la vista (F-02).
 */
export const NO_MODULES_MESSAGE =
  'Tu cuenta está activa. Aún no tienes secciones asignadas; si necesitas acceso, habla con tu administrador.';

/**
 * Inicio del panel (`/panel`, ux.md §3.3): la única pantalla que ve toda cuenta
 * autenticada. Muestra "Mi cuenta" y los accesos rápidos a las secciones
 * autorizadas; sin permisos de módulo muestra un aviso comprensible en lugar de
 * una sección vacía.
 */
export function InicioPage() {
  const sessionQuery = useSessionQuery();

  if (sessionQuery.isPending) {
    return (
      <section className="space-y-4">
        <h1 className="text-2xl font-bold text-slate-900">Inicio</h1>
        <EmptyState kind="loading" message="Cargando tu cuenta…" />
      </section>
    );
  }

  if (sessionQuery.isError || !sessionQuery.data) {
    return (
      <section className="space-y-4">
        <h1 className="text-2xl font-bold text-slate-900">Inicio</h1>
        <EmptyState
          kind="error"
          message="No se pudo cargar tu cuenta. Vuelve a intentarlo en unos minutos."
          action={<Button onClick={() => void sessionQuery.refetch()}>Reintentar</Button>}
        />
      </section>
    );
  }

  const { data: session } = sessionQuery;
  const canManage = hasPermission(session, ADMIN_USERS_ROLES);

  return (
    <section className="space-y-4">
      <h1 className="text-2xl font-bold text-slate-900">Inicio</h1>
      <MiCuentaCard session={session} />
      {canManage ? (
        <AccesosRapidos session={session} />
      ) : (
        <Notice variant="info">{NO_MODULES_MESSAGE}</Notice>
      )}
    </section>
  );
}
