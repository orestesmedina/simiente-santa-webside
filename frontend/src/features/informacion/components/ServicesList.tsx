import { useState } from 'react';
import type { ScheduleItemAdmin } from '../../../api/portada';
import { Button } from '../../../components/Button';
import { ConfirmDialog } from '../../../components/ConfirmDialog';
import { EmptyState } from '../../../components/EmptyState';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { StatusPill } from '../../../components/StatusPill';
import { Table, type Column } from '../../../components/Table';
import { useDeleteService } from '../hooks/useDeleteService';
import { useServices } from '../hooks/useServices';
import { useUpdateService } from '../hooks/useUpdateService';
import {
  ADD_SERVICE,
  DAY_OF_WEEK_LABELS,
  DELETE_ACTION,
  DELETE_CONFIRM_ACTION,
  DELETE_CONFIRM_TITLE,
  DELETED,
  DRAFT_NOTICE,
  EDIT_ACTION,
  EMPTY_SERVICES,
  PUBLISHED,
  UNPUBLISHED,
  deleteConfirmDescription,
} from '../messages';
import { PublishControls } from './PublishControls';
import { ServiceForm } from './ServiceForm';

interface Feedback {
  variant: NoticeVariant;
  message: string;
  draft?: boolean;
}

/**
 * Listado del horario (ux.md §4.6, FR-004/FR-013): tabla con día, hora, nombre y
 * lugar, píldora de estado y acciones Editar / Publicar / Retirar / Eliminar por
 * servicio; el alta y la edición abren `ServiceForm` en un `Dialog`.
 */
export function ServicesList() {
  const { items } = useServices();
  const updateService = useUpdateService();
  const deleteService = useDeleteService();

  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<ScheduleItemAdmin | null>(null);
  const [deleting, setDeleting] = useState<ScheduleItemAdmin | null>(null);
  const [feedback, setFeedback] = useState<Feedback | null>(null);

  const handleDone = (result: { message: string; draft?: boolean }) => {
    setCreating(false);
    setEditing(null);
    setFeedback({ variant: 'success', ...result });
  };

  const columns: Column<ScheduleItemAdmin>[] = [
    { key: 'day', header: 'Día', render: (service) => DAY_OF_WEEK_LABELS[service.dayOfWeek] },
    {
      key: 'time',
      header: 'Hora',
      render: (service) =>
        service.endTime ? `${service.startTime} – ${service.endTime}` : service.startTime,
    },
    { key: 'name', header: 'Servicio', render: (service) => service.nameEs },
    { key: 'place', header: 'Lugar', render: (service) => service.placeEs },
    {
      key: 'state',
      header: 'Estado',
      render: (service) => <StatusPill value={service.publicationState} />,
    },
    {
      key: 'actions',
      header: 'Acciones',
      render: (service) => (
        <div className="flex flex-wrap items-center gap-2">
          <Button size="sm" variant="secondary" onClick={() => setEditing(service)}>
            {EDIT_ACTION}
          </Button>
          <PublishControls
            name={service.nameEs}
            state={service.publicationState}
            loading={updateService.isPending}
            onPublish={() =>
              updateService.mutate(
                { id: service.id, input: { publicationState: 'published' } },
                { onSuccess: () => setFeedback({ variant: 'success', message: PUBLISHED }) },
              )
            }
            onUnpublish={() =>
              updateService.mutate(
                { id: service.id, input: { publicationState: 'draft' } },
                { onSuccess: () => setFeedback({ variant: 'success', message: UNPUBLISHED }) },
              )
            }
          />
          <Button size="sm" variant="danger" onClick={() => setDeleting(service)}>
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

      {items.length === 0 ? (
        <EmptyState
          message={EMPTY_SERVICES}
          action={<Button onClick={() => setCreating(true)}>{ADD_SERVICE}</Button>}
        />
      ) : (
        <>
          <div className="flex justify-end">
            <Button onClick={() => setCreating(true)}>{ADD_SERVICE}</Button>
          </div>
          <Table
            columns={columns}
            rows={items}
            rowKey={(service) => service.id}
            caption="Servicios del horario"
          />
        </>
      )}

      {creating && <ServiceForm onDone={handleDone} onCancel={() => setCreating(false)} />}
      {editing && (
        <ServiceForm service={editing} onDone={handleDone} onCancel={() => setEditing(null)} />
      )}
      {deleting && (
        <ConfirmDialog
          title={DELETE_CONFIRM_TITLE}
          description={deleteConfirmDescription(deleting.nameEs)}
          confirmText={DELETE_CONFIRM_ACTION}
          danger
          loading={deleteService.isPending}
          onConfirm={() =>
            deleteService.mutate(deleting.id, {
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
