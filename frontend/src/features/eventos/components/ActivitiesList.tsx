import { useState } from 'react';
import type { ActividadAdmin, CategoriaAdmin } from '../../../api/agenda';
import { Button } from '../../../components/Button';
import { ConfirmDialog } from '../../../components/ConfirmDialog';
import { EmptyState } from '../../../components/EmptyState';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { PublishControls } from '../../../components/PublishControls';
import { Table, type Column } from '../../../components/Table';
import { formatLongDate } from '../../publico/eventos/formatoFechas';
import { useDeleteActivity, useUpdateActivity } from '../hooks/useActivityMutations';
import { ActivityForm } from './ActivityForm';
import { EstadoBadge } from './EstadoBadge';
import {
  ADD_ACTIVITY,
  ALL_PAST_NOTICE,
  DELETED,
  DELETE_ACTION,
  DELETE_CANCEL,
  DELETE_CONFIRM_ACTION,
  DELETE_CONFIRM_TITLE,
  DRAFT_NOTICE,
  EDIT_ACTION,
  NO_NEXT_DATE,
  NOTHING_PUBLISHED,
  PUBLISHED,
  RETIRE_ACTION,
  RETIRE_CANCEL,
  RETIRE_CONFIRM_TITLE,
  UNPUBLISHED,
  deleteConfirmDescription,
  emptyModuleMessage,
  retireConfirmDescription,
} from '../messages';

interface Feedback {
  variant: NoticeVariant;
  message: string;
  draft?: boolean;
}

export interface ActivitiesListProps {
  activities: ActividadAdmin[];
  categories: CategoriaAdmin[];
}

/** Fecha (solo una) escrita, y horario «8:00 – 11:00» o «—» si no tiene. */
function scheduleLabel(activity: ActividadAdmin): string {
  if (!activity.startTime) {
    return NO_NEXT_DATE;
  }
  return activity.endTime ? `${activity.startTime} – ${activity.endTime}` : activity.startTime;
}

/**
 * Listado del panel de actividades (ux.md §4.5, US3/US4): tabla con **una sola
 * fecha**, su horario, el estado y las mismas acciones por elemento que los
 * eventos. Los textos del módulo dicen «Actividad» (nunca «it» mezclado).
 */
export function ActivitiesList({ activities, categories }: ActivitiesListProps) {
  const updateActivity = useUpdateActivity();
  const deleteActivity = useDeleteActivity();

  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<ActividadAdmin | null>(null);
  const [deleting, setDeleting] = useState<ActividadAdmin | null>(null);
  const [feedback, setFeedback] = useState<Feedback | null>(null);

  const handleDone = (result: { message: string; draft?: boolean }) => {
    setCreating(false);
    setEditing(null);
    setFeedback({ variant: 'success', ...result });
  };

  const nonePublished =
    activities.length > 0 && activities.every((item) => item.publicationState !== 'published');
  const allPast = activities.length > 0 && activities.every((item) => item.isPast);

  const columns: Column<ActividadAdmin>[] = [
    { key: 'title', header: 'Título', render: (item) => item.titleEs },
    {
      key: 'date',
      header: 'Fecha',
      render: (item) => formatLongDate(item.date, 'es'),
    },
    { key: 'schedule', header: 'Horario', render: scheduleLabel },
    {
      key: 'state',
      header: 'Estado',
      render: (item) => <EstadoBadge state={item.publicationState} isPast={item.isPast} />,
    },
    {
      key: 'actions',
      header: 'Acciones',
      render: (item) => (
        <div className="flex flex-wrap items-center gap-2">
          <Button size="sm" variant="secondary" onClick={() => setEditing(item)}>
            {EDIT_ACTION}
          </Button>
          <PublishControls
            name={item.titleEs}
            state={item.publicationState}
            loading={updateActivity.isPending}
            retireLabel={RETIRE_ACTION}
            confirmTitle={RETIRE_CONFIRM_TITLE}
            confirmDescription={retireConfirmDescription}
            cancelText={RETIRE_CANCEL}
            onPublish={() =>
              updateActivity.mutate(
                { id: item.id, input: { publicationState: 'published' } },
                { onSuccess: () => setFeedback({ variant: 'success', message: PUBLISHED }) },
              )
            }
            onUnpublish={() =>
              updateActivity.mutate(
                { id: item.id, input: { publicationState: 'draft' } },
                { onSuccess: () => setFeedback({ variant: 'success', message: UNPUBLISHED }) },
              )
            }
          />
          <Button size="sm" variant="danger" onClick={() => setDeleting(item)}>
            {DELETE_ACTION}
          </Button>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-3">
      {feedback && <Notice variant={feedback.variant}>{feedback.message}</Notice>}
      {feedback?.draft && <Notice variant="info">{DRAFT_NOTICE}</Notice>}
      {nonePublished && <Notice variant="warning">{NOTHING_PUBLISHED}</Notice>}
      {allPast && <Notice variant="warning">{ALL_PAST_NOTICE}</Notice>}

      {activities.length === 0 ? (
        <EmptyState
          message={emptyModuleMessage('actividades')}
          action={<Button onClick={() => setCreating(true)}>{ADD_ACTIVITY}</Button>}
        />
      ) : (
        <>
          <div className="flex justify-end">
            <Button onClick={() => setCreating(true)}>{ADD_ACTIVITY}</Button>
          </div>
          <Table
            columns={columns}
            rows={activities}
            rowKey={(item) => item.id}
            caption="Actividades de la iglesia"
          />
        </>
      )}

      {creating && (
        <ActivityForm
          categories={categories}
          onDone={handleDone}
          onCancel={() => setCreating(false)}
        />
      )}
      {editing && (
        <ActivityForm
          activity={editing}
          categories={categories}
          onDone={handleDone}
          onCancel={() => setEditing(null)}
        />
      )}
      {deleting && (
        <ConfirmDialog
          title={DELETE_CONFIRM_TITLE}
          description={deleteConfirmDescription(deleting.titleEs)}
          confirmText={DELETE_CONFIRM_ACTION}
          cancelText={DELETE_CANCEL}
          danger
          loading={deleteActivity.isPending}
          onConfirm={() =>
            deleteActivity.mutate(deleting.id, {
              onSuccess: () => {
                setDeleting(null);
                setFeedback({ variant: 'success', message: DELETED });
              },
            })
          }
          onClose={() => setDeleting(null)}
        />
      )}
    </div>
  );
}
