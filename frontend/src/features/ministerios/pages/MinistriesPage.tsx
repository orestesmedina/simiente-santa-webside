import { Button } from '../../../components/Button';
import { EmptyState } from '../../../components/EmptyState';
import { Notice } from '../../../components/Notice';
import { MinistriesList } from '../components/MinistriesList';
import { isForbidden } from '../errors';
import { useMinistries } from '../hooks/useMinistries';
import { HELP, LOAD_ERROR, NO_PERMISSION_ERROR, RETRY, TITLE } from '../messages';

/**
 * Módulo «Gestión de ministerios» del panel (`/panel/ministerios`, ux.md
 * §4.3/§5.a, US2/US3): título, banda de ayuda y el listado con sus estados
 * —cargando, error (con «Reintentar»), sin permiso del servidor y éxito—.
 *
 * El listado completo (sin paginar, ux.md §4.3) llega ya ordenado por el
 * servidor en `sortOrder`, `id`: el mismo orden que consume el público. Con todo
 * en borrador el listado avisa además de que ni la sección ni el bloque de la
 * portada se ven y de que `/ministerios` redirige a la portada (FR-005/FR-028).
 *
 * La autorización es del servidor (`AdminChain("ministerios")`): el permiso
 * `ministerios` de F2 protege la ruta en la interfaz (`RequirePermission`) y, si
 * el servidor rechaza en caliente, se avisa sin detalles internos (FR-011).
 */
export function MinistriesPage() {
  const query = useMinistries();
  const hasData = Boolean(query.data);

  return (
    <section className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">{TITLE}</h1>
        <p className="mt-2 text-slate-700">{HELP}</p>
      </div>

      {query.isError && isForbidden(query.error) && (
        <Notice variant="error">{NO_PERMISSION_ERROR}</Notice>
      )}

      {query.isPending && <EmptyState kind="loading" message="Cargando ministerios…" />}

      {query.isError && !isForbidden(query.error) && !hasData && (
        <EmptyState
          kind="error"
          message={LOAD_ERROR}
          action={<Button onClick={() => void query.refetch()}>{RETRY}</Button>}
        />
      )}

      {/* Error con contenido en caché (ux.md §5.a): se conserva el último listado
          válido y el aviso explica que la actualización falló. */}
      {query.isError && !isForbidden(query.error) && hasData && (
        <Notice variant="warning">{LOAD_ERROR}</Notice>
      )}

      {query.data && <MinistriesList ministries={query.data.items} />}
    </section>
  );
}
