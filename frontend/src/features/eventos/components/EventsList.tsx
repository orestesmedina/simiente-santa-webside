import { useState } from 'react';
import type { CategoriaAdmin, EventoAdmin } from '../../../api/agenda';
import { Button } from '../../../components/Button';
import { ConfirmDialog } from '../../../components/ConfirmDialog';
import { EmptyState } from '../../../components/EmptyState';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { PublishControls } from '../../../components/PublishControls';
import { Table, type Column } from '../../../components/Table';
import { formatLongDate } from '../../publico/eventos/formatoFechas';
import { useDeleteEvent, useUpdateEvent } from '../hooks/useEventMutations';
import { EstadoBadge } from './EstadoBadge';
import { EventForm } from './EventForm';
import {
  ADD_EVENT,
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
  futureDatesLabel,
  retireConfirmDescription,
} from '../messages';

interface Feedback {
  variant: NoticeVariant;
  message: string;
  draft?: boolean;
}

export interface EventsListProps {
  events: EventoAdmin[];
  categories: CategoriaAdmin[];
}

/** Fecha de la próxima ocurrencia, escrita y con la hora corta del contrato. */
function nextDateLabel(event: EventoAdmin): string {
  if (!event.nextDate) {
    return NO_NEXT_DATE;
  }
  const next = event.dates.find((date) => date.date === event.nextDate);
  const written = formatLongDate(event.nextDate, 'es');
  return next?.startTime ? `${written} · ${next.startTime}` : written;
}

/** «Todos los viernes · 3 fechas futuras» (ux.md §4.3). */
function frequencyLabel(event: EventoAdmin): string {
  const future = event.dates.filter((date) => !date.isPast).length;
  const frequency = event.frequencyEs?.trim();
  const parts: string[] = [];
  if (frequency) {
    parts.push(frequency);
  }
  if (future > 0) {
    parts.push(futureDatesLabel(future));
  }
  return parts.join(' · ') || NO_NEXT_DATE;
}

/**
 * Listado del panel de eventos (ux.md §4.3, US2/US4): tabla en escritorio y
 * tarjetas en móvil con próxima fecha, frecuencia y fechas futuras, estado
 * (`StatusPill` borrador/publicado) con la marca **«Pasado»** cuando
 * corresponde, y acciones Editar · Publicar/Retirar de la sección · Eliminar.
 * El alta y la edición abren `EventForm` en un `Dialog`; el borrado pide
 * confirmación explícita (no hay papelera, FR-015).
 */
export function EventsList({ events, categories }: EventsListProps) {
  const updateEvent = useUpdateEvent();
  const deleteEvent = useDeleteEvent();

  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<EventoAdmin | null>(null);
  const [deleting, setDeleting] = useState<EventoAdmin | null>(null);
  const [feedback, setFeedback] = useState<Feedback | null>(null);

  const handleDone = (result: { message: string; draft?: boolean }) => {
    setCreating(false);
    setEditing(null);
    setFeedback({ variant: 'success', ...result });
  };

  const nonePublished =
    events.length > 0 && events.every((event) => event.publicationState !== 'published');
  const allPast = events.length > 0 && events.every((event) => event.isPast);

  const columns: Column<EventoAdmin>[] = [
    { key: 'title', header: 'Título', render: (event) => event.titleEs },
    { key: 'nextDate', header: 'Próxima fecha', render: nextDateLabel },
    {
      key: 'frequency',
      header: 'Frecuencia / fechas futuras',
      render: frequencyLabel,
    },
    {
      key: 'state',
      header: 'Estado',
      render: (event) => <EstadoBadge state={event.publicationState} isPast={event.isPast} />,
    },
    {
      key: 'actions',
      header: 'Acciones',
      render: (event) => (
        <div className="flex flex-wrap items-center gap-2">
          <Button size="sm" variant="secondary" onClick={() => setEditing(event)}>
            {EDIT_ACTION}
          </Button>
          <PublishControls
            name={event.titleEs}
            state={event.publicationState}
            loading={updateEvent.isPending}
            retireLabel={RETIRE_ACTION}
            confirmTitle={RETIRE_CONFIRM_TITLE}
            confirmDescription={retireConfirmDescription}
            cancelText={RETIRE_CANCEL}
            onPublish={() =>
              updateEvent.mutate(
                { id: event.id, input: { publicationState: 'published' } },
                { onSuccess: () => setFeedback({ variant: 'success', message: PUBLISHED }) },
              )
            }
            onUnpublish={() =>
              updateEvent.mutate(
                { id: event.id, input: { publicationState: 'draft' } },
                { onSuccess: () => setFeedback({ variant: 'success', message: UNPUBLISHED }) },
              )
            }
          />
          <Button size="sm" variant="danger" onClick={() => setDeleting(event)}>
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

      {events.length === 0 ? (
        <EmptyState
          message={emptyModuleMessage('eventos')}
          action={<Button onClick={() => setCreating(true)}>{ADD_EVENT}</Button>}
        />
      ) : (
        <>
          <div className="flex justify-end">
            <Button onClick={() => setCreating(true)}>{ADD_EVENT}</Button>
          </div>
          <Table
            columns={columns}
            rows={events}
            rowKey={(event) => event.id}
            caption="Eventos de la iglesia"
          />
        </>
      )}

      {creating && (
        <EventForm
          categories={categories}
          onDone={handleDone}
          onCancel={() => setCreating(false)}
        />
      )}
      {editing && (
        <EventForm
          event={editing}
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
          loading={deleteEvent.isPending}
          onConfirm={() =>
            deleteEvent.mutate(deleting.id, {
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
