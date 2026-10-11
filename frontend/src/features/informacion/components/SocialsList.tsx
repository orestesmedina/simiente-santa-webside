import { useState } from 'react';
import { z } from 'zod';
import { ApiError } from '../../../api/client';
import type { SocialLinkAdmin, SocialNetwork } from '../../../api/portada';
import { Button } from '../../../components/Button';
import { ConfirmDialog } from '../../../components/ConfirmDialog';
import { Dialog } from '../../../components/Dialog';
import { Field } from '../../../components/Field';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { StatusPill } from '../../../components/StatusPill';
import { Table, type Column } from '../../../components/Table';
import { SOCIAL_NETWORK_LABELS, SOCIAL_NETWORKS } from '../../publico/social';
import { detailMessage, generalErrorMessage } from '../errors';
import { useCreateSocialLink } from '../hooks/useCreateSocialLink';
import { useDeleteSocial } from '../hooks/useDeleteSocial';
import { useSocials } from '../hooks/useSocials';
import { useUpdateSocial } from '../hooks/useUpdateSocial';
import {
  ADD_SOCIAL,
  DELETE_ACTION,
  DELETE_CONFIRM_ACTION,
  DELETE_CONFIRM_TITLE,
  DELETED,
  EDIT_ACTION,
  NO_SOCIAL_LINK,
  PUBLISHED,
  REQUIRED_URL,
  SOCIAL_URL_HELP,
  UNPUBLISHED,
  URL_INVALID,
  deleteConfirmDescription,
  socialDuplicateMessage,
  socialSavedMessage,
} from '../messages';
import { isHttpsUrl } from '../validation';
import { PublishControls } from './PublishControls';

interface SocialRow {
  network: SocialNetwork;
  link?: SocialLinkAdmin;
}

interface Feedback {
  variant: NoticeVariant;
  message: string;
}

const urlSchema = z.object({
  url: z.string().trim().min(1, REQUIRED_URL).refine(isHttpsUrl, URL_INVALID),
});

interface SocialFormProps {
  network: SocialNetwork;
  link?: SocialLinkAdmin;
  onDone: (message: string) => void;
  onCancel: () => void;
}

function SocialForm({ network, link, onDone, onCancel }: SocialFormProps) {
  const create = useCreateSocialLink();
  const update = useUpdateSocial();
  const [url, setUrl] = useState(link?.url ?? '');
  const [error, setError] = useState<string>();
  const label = SOCIAL_NETWORK_LABELS[network];
  const busy = create.isPending || update.isPending;
  const serverError = link ? update.error : create.error;

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const parsed = urlSchema.safeParse({ url });
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message);
      return;
    }
    setError(undefined);
    if (link) {
      update.mutate(
        { id: link.id, input: { url: url.trim() } },
        { onSuccess: () => onDone(socialSavedMessage(label)) },
      );
      return;
    }
    create.mutate(
      { network, url: url.trim() },
      { onSuccess: () => onDone(socialSavedMessage(label)) },
    );
  };

  let serverMessage: string | undefined;
  if (serverError) {
    serverMessage =
      serverError instanceof ApiError && serverError.status === 409
        ? socialDuplicateMessage(label)
        : (detailMessage(serverError, 'url') ??
          detailMessage(serverError, 'network') ??
          generalErrorMessage(serverError));
  }

  return (
    <Dialog
      title={link ? `Editar enlace de ${label}` : `Agregar enlace de ${label}`}
      onClose={onCancel}
    >
      <form onSubmit={handleSubmit} noValidate className="space-y-4">
        {serverMessage && <Notice variant="error">{serverMessage}</Notice>}
        <Field
          label={`Enlace de ${label}`}
          required
          value={url}
          help={SOCIAL_URL_HELP}
          error={error}
          onChange={(event) => setUrl(event.target.value)}
        />
        <div className="flex flex-wrap justify-end gap-3">
          <Button type="button" variant="secondary" onClick={onCancel} disabled={busy}>
            Cancelar
          </Button>
          <Button type="submit" loading={busy}>
            Guardar
          </Button>
        </div>
      </form>
    </Dialog>
  );
}

/**
 * Redes sociales (ux.md §4.8, FR-006/FR-013): el catálogo fijo se muestra
 * **completo** (una fila por red); las redes sin enlace indican «Sin enlace
 * todavía» y ofrecen agregarlo; cada enlace tiene su estado y se publica/retira
 * por elemento.
 */
export function SocialsList() {
  const { items } = useSocials();
  const updateSocial = useUpdateSocial();
  const deleteSocial = useDeleteSocial();

  const [form, setForm] = useState<{ network: SocialNetwork; link?: SocialLinkAdmin } | null>(null);
  const [deleting, setDeleting] = useState<SocialLinkAdmin | null>(null);
  const [feedback, setFeedback] = useState<Feedback | null>(null);

  const byNetwork = new Map(items.map((link) => [link.network, link]));
  const rows: SocialRow[] = SOCIAL_NETWORKS.map((network) => ({
    network,
    link: byNetwork.get(network),
  }));

  const columns: Column<SocialRow>[] = [
    { key: 'network', header: 'Red', render: (row) => SOCIAL_NETWORK_LABELS[row.network] },
    {
      key: 'url',
      header: 'Enlace',
      render: (row) => row.link?.url ?? NO_SOCIAL_LINK,
    },
    {
      key: 'state',
      header: 'Estado',
      render: (row) => (row.link ? <StatusPill value={row.link.publicationState} /> : '—'),
    },
    {
      key: 'actions',
      header: 'Acciones',
      render: (row) => {
        const link = row.link;
        if (!link) {
          return (
            <Button size="sm" variant="secondary" onClick={() => setForm({ network: row.network })}>
              {ADD_SOCIAL}
            </Button>
          );
        }
        return (
          <div className="flex flex-wrap items-center gap-2">
            <Button
              size="sm"
              variant="secondary"
              onClick={() => setForm({ network: row.network, link })}
            >
              {EDIT_ACTION}
            </Button>
            <PublishControls
              name={SOCIAL_NETWORK_LABELS[row.network]}
              state={link.publicationState}
              loading={updateSocial.isPending}
              onPublish={() =>
                updateSocial.mutate(
                  { id: link.id, input: { publicationState: 'published' } },
                  { onSuccess: () => setFeedback({ variant: 'success', message: PUBLISHED }) },
                )
              }
              onUnpublish={() =>
                updateSocial.mutate(
                  { id: link.id, input: { publicationState: 'draft' } },
                  { onSuccess: () => setFeedback({ variant: 'success', message: UNPUBLISHED }) },
                )
              }
            />
            <Button size="sm" variant="danger" onClick={() => setDeleting(link)}>
              {DELETE_ACTION}
            </Button>
          </div>
        );
      },
    },
  ];

  return (
    <div className="space-y-3">
      {feedback && <Notice variant={feedback.variant}>{feedback.message}</Notice>}

      <Table columns={columns} rows={rows} rowKey={(row) => row.network} caption="Redes sociales" />

      {form && (
        <SocialForm
          network={form.network}
          link={form.link}
          onDone={(message) => {
            setForm(null);
            setFeedback({ variant: 'success', message });
          }}
          onCancel={() => setForm(null)}
        />
      )}
      {deleting && (
        <ConfirmDialog
          title={DELETE_CONFIRM_TITLE}
          description={deleteConfirmDescription(SOCIAL_NETWORK_LABELS[deleting.network])}
          confirmText={DELETE_CONFIRM_ACTION}
          danger
          loading={deleteSocial.isPending}
          onConfirm={() =>
            deleteSocial.mutate(deleting.id, {
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
