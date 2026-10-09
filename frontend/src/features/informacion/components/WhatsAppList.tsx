import { useState } from 'react';
import type { WhatsappChannelAdmin } from '../../../api/portada';
import { Button } from '../../../components/Button';
import { ConfirmDialog } from '../../../components/ConfirmDialog';
import { EmptyState } from '../../../components/EmptyState';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { StatusPill } from '../../../components/StatusPill';
import { Table, type Column } from '../../../components/Table';
import { useDeleteWhatsappChannel } from '../hooks/useDeleteWhatsappChannel';
import { useUpdateWhatsappChannel } from '../hooks/useUpdateWhatsappChannel';
import { useWhatsappChannels } from '../hooks/useWhatsappChannels';
import {
  ADD_CHANNEL,
  DELETE_ACTION,
  DELETE_CONFIRM_ACTION,
  DELETE_CONFIRM_TITLE,
  DELETED,
  DRAFT_NOTICE,
  EDIT_ACTION,
  EMPTY_WHATSAPP,
  PUBLISHED,
  UNPUBLISHED,
  WHATSAPP_DIRECT_LABEL,
  WHATSAPP_GROUP_LABEL,
  deleteConfirmDescription,
} from '../messages';
import { PublishControls } from './PublishControls';
import { WhatsAppForm } from './WhatsAppForm';

interface Feedback {
  variant: NoticeVariant;
  message: string;
  draft?: boolean;
}

/**
 * Listado de canales de WhatsApp (ux.md §4.7, FR-005/FR-013): nombre, tipo,
 * destino legible, estado y acciones por canal; el alta/edición abre
 * `WhatsAppForm` en un `Dialog`.
 */
export function WhatsAppList() {
  const { items } = useWhatsappChannels();
  const updateChannel = useUpdateWhatsappChannel();
  const deleteChannel = useDeleteWhatsappChannel();

  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<WhatsappChannelAdmin | null>(null);
  const [deleting, setDeleting] = useState<WhatsappChannelAdmin | null>(null);
  const [feedback, setFeedback] = useState<Feedback | null>(null);

  const handleDone = (result: { message: string; draft?: boolean }) => {
    setCreating(false);
    setEditing(null);
    setFeedback({ variant: 'success', ...result });
  };

  const columns: Column<WhatsappChannelAdmin>[] = [
    { key: 'name', header: 'Nombre', render: (channel) => channel.nameEs },
    {
      key: 'kind',
      header: 'Tipo',
      render: (channel) =>
        channel.kind === 'direct' ? WHATSAPP_DIRECT_LABEL : WHATSAPP_GROUP_LABEL,
    },
    { key: 'destination', header: 'Destino', render: (channel) => channel.destination },
    {
      key: 'state',
      header: 'Estado',
      render: (channel) => <StatusPill value={channel.publicationState} />,
    },
    {
      key: 'actions',
      header: 'Acciones',
      render: (channel) => (
        <div className="flex flex-wrap items-center gap-2">
          <Button size="sm" variant="secondary" onClick={() => setEditing(channel)}>
            {EDIT_ACTION}
          </Button>
          <PublishControls
            name={channel.nameEs}
            state={channel.publicationState}
            loading={updateChannel.isPending}
            onPublish={() =>
              updateChannel.mutate(
                { id: channel.id, input: { publicationState: 'published' } },
                { onSuccess: () => setFeedback({ variant: 'success', message: PUBLISHED }) },
              )
            }
            onUnpublish={() =>
              updateChannel.mutate(
                { id: channel.id, input: { publicationState: 'draft' } },
                { onSuccess: () => setFeedback({ variant: 'success', message: UNPUBLISHED }) },
              )
            }
          />
          <Button size="sm" variant="danger" onClick={() => setDeleting(channel)}>
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
          message={EMPTY_WHATSAPP}
          action={<Button onClick={() => setCreating(true)}>{ADD_CHANNEL}</Button>}
        />
      ) : (
        <>
          <div className="flex justify-end">
            <Button onClick={() => setCreating(true)}>{ADD_CHANNEL}</Button>
          </div>
          <Table
            columns={columns}
            rows={items}
            rowKey={(channel) => channel.id}
            caption="Canales de WhatsApp"
          />
        </>
      )}

      {creating && <WhatsAppForm onDone={handleDone} onCancel={() => setCreating(false)} />}
      {editing && (
        <WhatsAppForm channel={editing} onDone={handleDone} onCancel={() => setEditing(null)} />
      )}
      {deleting && (
        <ConfirmDialog
          title={DELETE_CONFIRM_TITLE}
          description={deleteConfirmDescription(deleting.nameEs)}
          confirmText={DELETE_CONFIRM_ACTION}
          danger
          loading={deleteChannel.isPending}
          onConfirm={() =>
            deleteChannel.mutate(deleting.id, {
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
