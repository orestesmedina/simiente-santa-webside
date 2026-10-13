import { useState } from 'react';
import type { MinisterioAdmin } from '../../../api/ministerios';
import { Button } from '../../../components/Button';
import { ConfirmDialog } from '../../../components/ConfirmDialog';
import { EmptyState } from '../../../components/EmptyState';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { PublishControls } from '../../../components/PublishControls';
import { StatusPill } from '../../../components/StatusPill';
import { Table, type Column } from '../../../components/Table';
import { useDeleteMinistry } from '../hooks/useDeleteMinistry';
import { useUpdateMinistry } from '../hooks/useUpdateMinistry';
import { MinistryForm } from './MinistryForm';
import {
  ADD,
  DELETED,
  DELETE_ACTION,
  DELETE_CANCEL,
  DELETE_CONFIRM_ACTION,
  DELETE_CONFIRM_TITLE,
  DRAFT_NOTICE,
  EDIT_ACTION,
  EMPTY_DASH,
  HIDDEN_LEADER_LABEL,
  HIDDEN_LEADER_LEGEND,
  NOTHING_PUBLISHED,
  PUBLISHED,
  RETIRE_ACTION,
  RETIRE_CANCEL,
  RETIRE_CONFIRM_TITLE,
  SHARED_LINK_NOTICE,
  UNPUBLISHED,
  deleteConfirmDescription,
  emptyModuleMessage,
  retireConfirmDescription,
} from '../messages';

interface Feedback {
  variant: NoticeVariant;
  message: string;
  draft?: boolean;
  sharedLink?: boolean;
}

export interface MinistriesListProps {
  ministries: MinisterioAdmin[];
}

/** «Todos los viernes, 7:00 p. m. · Templo central», o «—» si no hay ninguno. */
function schedulePlaceLabel(ministry: MinisterioAdmin): string {
  const parts = [ministry.scheduleEs?.trim(), ministry.placeEs?.trim()].filter(
    (value): value is string => Boolean(value),
  );
  return parts.length > 0 ? parts.join(' · ') : EMPTY_DASH;
}

/**
 * Listado del módulo (ux.md §4.3/§5.a, US2/US3): tabla en escritorio y tarjetas
 * en móvil (`Table`) con **Orden** (el valor numérico, ⟲ D-6), **Nombre**,
 * **Horario / Lugar** («—» si no hay), **Encargado** (el nombre guardado —el
 * panel SIEMPRE ve el dato, dato guardado ≠ dato publicado— con la etiqueta
 * **«Ficha oculta»** cuando el interruptor está apagado), **Estado**
 * (`StatusPill` borrador/publicado, **sin** marca «Pasado»: en F6 nada vence) y
 * **Acciones** (Editar · Publicar/Retirar de la sección · Eliminar).
 *
 * Publicar es directo; retirar pide una confirmación suave y eliminar una
 * explícita de no recuperable (FR-017). El aviso de éxito dice en claro qué pasó.
 * El orden lo resuelve el servidor (`sortOrder`, `id`): el mismo que ve el
 * visitante (US2 esc. 9).
 */
export function MinistriesList({ ministries }: MinistriesListProps) {
  const updateMinistry = useUpdateMinistry();
  const deleteMinistry = useDeleteMinistry();

  const [feedback, setFeedback] = useState<Feedback | null>(null);
  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<MinisterioAdmin | null>(null);
  const [deleting, setDeleting] = useState<MinisterioAdmin | null>(null);

  const nonePublished =
    ministries.length > 0 &&
    ministries.every((ministry) => ministry.publicationState !== 'published');

  /** Cierra el editor y deja el aviso de éxito (con sus matices) en el listado. */
  const handleDone = (result: { message: string; draft?: boolean; sharedLink?: boolean }) => {
    setCreating(false);
    setEditing(null);
    setFeedback({ variant: 'success', ...result });
  };

  const columns: Column<MinisterioAdmin>[] = [
    { key: 'sortOrder', header: 'Orden', render: (ministry) => String(ministry.sortOrder) },
    { key: 'name', header: 'Nombre', render: (ministry) => ministry.nameEs },
    {
      key: 'schedulePlace',
      header: 'Horario / Lugar',
      render: (ministry) => schedulePlaceLabel(ministry),
    },
    {
      key: 'leader',
      header: 'Encargado',
      render: (ministry) => (
        <span className="flex flex-wrap items-center gap-2">
          <span>{ministry.leaderNameEs?.trim() || EMPTY_DASH}</span>
          {ministry.leaderNameEs && !ministry.leaderVisible && (
            <span
              title={HIDDEN_LEADER_LEGEND}
              className="rounded-full border border-slate-400 px-2 py-0.5 text-xs font-medium text-slate-700"
            >
              {HIDDEN_LEADER_LABEL}
            </span>
          )}
        </span>
      ),
    },
    {
      key: 'state',
      header: 'Estado',
      render: (ministry) => <StatusPill value={ministry.publicationState} />,
    },
    {
      key: 'actions',
      header: 'Acciones',
      render: (ministry) => (
        <div className="flex flex-wrap items-center gap-2">
          <Button
            size="sm"
            variant="secondary"
            aria-haspopup="dialog"
            aria-expanded={editing?.id === ministry.id}
            onClick={() => setEditing(ministry)}
          >
            {EDIT_ACTION}
          </Button>
          <PublishControls
            name={ministry.nameEs}
            state={ministry.publicationState}
            loading={updateMinistry.isPending}
            retireLabel={RETIRE_ACTION}
            confirmTitle={RETIRE_CONFIRM_TITLE}
            confirmDescription={retireConfirmDescription}
            cancelText={RETIRE_CANCEL}
            onPublish={() =>
              updateMinistry.mutate(
                { id: ministry.id, input: { publicationState: 'published' } },
                { onSuccess: () => setFeedback({ variant: 'success', message: PUBLISHED }) },
              )
            }
            onUnpublish={() =>
              updateMinistry.mutate(
                { id: ministry.id, input: { publicationState: 'draft' } },
                { onSuccess: () => setFeedback({ variant: 'success', message: UNPUBLISHED }) },
              )
            }
          />
          <Button
            size="sm"
            variant="danger"
            aria-haspopup="dialog"
            aria-expanded={deleting?.id === ministry.id}
            onClick={() => setDeleting(ministry)}
          >
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
      {feedback?.sharedLink && <Notice variant="info">{SHARED_LINK_NOTICE}</Notice>}
      {nonePublished && <Notice variant="warning">{NOTHING_PUBLISHED}</Notice>}

      {ministries.length === 0 ? (
        <EmptyState
          message={emptyModuleMessage()}
          action={
            <Button
              aria-haspopup="dialog"
              aria-expanded={creating}
              onClick={() => setCreating(true)}
            >
              {ADD}
            </Button>
          }
        />
      ) : (
        <>
          <div className="flex justify-end">
            <Button
              aria-haspopup="dialog"
              aria-expanded={creating}
              onClick={() => setCreating(true)}
            >
              {ADD}
            </Button>
          </div>
          <Table
            columns={columns}
            rows={ministries}
            rowKey={(ministry) => ministry.id}
            caption="Ministerios de la iglesia"
          />
        </>
      )}

      {/* El alta y la edición montan `MinistryForm` en un `Dialog`. */}
      {creating && <MinistryForm onDone={handleDone} onCancel={() => setCreating(false)} />}
      {editing && (
        <MinistryForm ministry={editing} onDone={handleDone} onCancel={() => setEditing(null)} />
      )}
      {deleting && (
        <ConfirmDialog
          title={DELETE_CONFIRM_TITLE}
          description={deleteConfirmDescription(deleting.nameEs)}
          confirmText={DELETE_CONFIRM_ACTION}
          cancelText={DELETE_CANCEL}
          danger
          loading={deleteMinistry.isPending}
          onConfirm={() =>
            deleteMinistry.mutate(deleting.id, {
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
