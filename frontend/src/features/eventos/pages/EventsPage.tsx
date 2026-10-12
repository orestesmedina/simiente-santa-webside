import { Button } from '../../../components/Button';
import { EmptyState } from '../../../components/EmptyState';
import { Notice } from '../../../components/Notice';
import { isForbidden } from '../errors';
import { useEventCategories } from '../hooks/useEventCategories';
import { useEvents } from '../hooks/useEvents';
import {
  EVENTS_HELP,
  EVENTS_LOAD_ERROR,
  EVENTS_TITLE,
  NO_PERMISSION_ERROR,
  RETRY,
} from '../messages';
import { EventsList } from '../components/EventsList';

/**
 * Módulo «Gestión de eventos» del panel (`/panel/eventos`, ux.md §4.3/§5.a):
 * título, banda de ayuda y el listado con sus estados —cargando, error (con
 * reintento), sin permiso del servidor y éxito—. El catálogo de categorías se
 * consulta una sola vez aquí y alimenta el `Select` de cada formulario.
 */
export function EventsPage() {
  const eventsQuery = useEvents();
  const categoriesQuery = useEventCategories();

  return (
    <section className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">{EVENTS_TITLE}</h1>
        <p className="mt-2 text-slate-700">{EVENTS_HELP}</p>
      </div>

      {eventsQuery.isError && isForbidden(eventsQuery.error) && (
        <Notice variant="error">{NO_PERMISSION_ERROR}</Notice>
      )}

      {eventsQuery.isPending && <EmptyState kind="loading" message="Cargando eventos…" />}

      {eventsQuery.isError && !isForbidden(eventsQuery.error) && (
        <EmptyState
          kind="error"
          message={EVENTS_LOAD_ERROR}
          action={<Button onClick={() => void eventsQuery.refetch()}>{RETRY}</Button>}
        />
      )}

      {eventsQuery.isSuccess && (
        <EventsList
          events={eventsQuery.data.items}
          categories={categoriesQuery.data?.items ?? []}
        />
      )}
    </section>
  );
}
