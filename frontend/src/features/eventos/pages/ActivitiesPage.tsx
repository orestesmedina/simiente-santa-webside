import { Button } from '../../../components/Button';
import { EmptyState } from '../../../components/EmptyState';
import { Notice } from '../../../components/Notice';
import { isForbidden } from '../errors';
import { useActivities } from '../hooks/useActivities';
import { useEventCategories } from '../hooks/useEventCategories';
import {
  ACTIVITIES_HELP,
  ACTIVITIES_LOAD_ERROR,
  ACTIVITIES_TITLE,
  NO_PERMISSION_ERROR,
  RETRY,
} from '../messages';
import { ActivitiesList } from '../components/ActivitiesList';

/**
 * Módulo «Gestión de actividades» del panel (`/panel/actividades`, ux.md §4.5/
 * §5.a): simétrico al de eventos, con el modelo de **fecha única** y sus
 * propios permisos (tener «eventos» no da acceso a este módulo, FR-012).
 */
export function ActivitiesPage() {
  const activitiesQuery = useActivities();
  const categoriesQuery = useEventCategories();

  return (
    <section className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">{ACTIVITIES_TITLE}</h1>
        <p className="mt-2 text-slate-700">{ACTIVITIES_HELP}</p>
      </div>

      {activitiesQuery.isError && isForbidden(activitiesQuery.error) && (
        <Notice variant="error">{NO_PERMISSION_ERROR}</Notice>
      )}

      {activitiesQuery.isPending && <EmptyState kind="loading" message="Cargando actividades…" />}

      {activitiesQuery.isError && !isForbidden(activitiesQuery.error) && (
        <EmptyState
          kind="error"
          message={ACTIVITIES_LOAD_ERROR}
          action={<Button onClick={() => void activitiesQuery.refetch()}>{RETRY}</Button>}
        />
      )}

      {activitiesQuery.isSuccess && (
        <ActivitiesList
          activities={activitiesQuery.data.items}
          categories={categoriesQuery.data?.items ?? []}
        />
      )}
    </section>
  );
}
