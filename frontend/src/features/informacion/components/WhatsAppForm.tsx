import { useState } from 'react';
import { z } from 'zod';
import { ApiError } from '../../../api/client';
import type { WhatsappChannelAdmin } from '../../../api/portada';
import { Button } from '../../../components/Button';
import { Dialog } from '../../../components/Dialog';
import { Field } from '../../../components/Field';
import { Notice } from '../../../components/Notice';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { detailMessage, generalErrorMessage } from '../errors';
import { useCreateWhatsappChannel } from '../hooks/useCreateWhatsappChannel';
import { useUpdateWhatsappChannel } from '../hooks/useUpdateWhatsappChannel';
import {
  ENGLISH_HELP,
  PHONE_INVALID,
  REQUIRED_DESTINATION,
  REQUIRED_WHATSAPP_NAME,
  SAVED,
  SAVED_PUBLISHED,
  WHATSAPP_CREATED,
  WHATSAPP_DUPLICATE,
  WHATSAPP_GROUP_HELP,
  WHATSAPP_GROUP_INVALID,
  WHATSAPP_PHONE_HELP,
} from '../messages';
import { isWhatsappGroupUrl, isValidPhone } from '../validation';

export interface WhatsAppFormProps {
  channel?: WhatsappChannelAdmin;
  onDone: (feedback: { message: string; draft?: boolean }) => void;
  onCancel: () => void;
}

const LANGUAGE_TABS: TabItem[] = [
  { id: 'es', label: 'Español' },
  { id: 'en', label: 'English (opcional)' },
];

interface ChannelFields {
  nameEs: string;
  nameEn: string;
  kind: 'direct' | 'group';
  destination: string;
}

function fieldsFromChannel(channel?: WhatsappChannelAdmin): ChannelFields {
  return {
    nameEs: channel?.nameEs ?? '',
    nameEn: channel?.nameEn ?? '',
    kind: channel?.kind ?? 'direct',
    destination: channel?.destination ?? '',
  };
}

const channelSchema = z
  .object({
    nameEs: z.string().trim().min(1, REQUIRED_WHATSAPP_NAME),
    kind: z.enum(['direct', 'group']),
    destination: z.string().trim().min(1, REQUIRED_DESTINATION),
  })
  .superRefine((values, ctx) => {
    if (values.kind === 'direct' && !isValidPhone(values.destination)) {
      ctx.addIssue({ code: 'custom', path: ['destination'], message: PHONE_INVALID });
    }
    if (values.kind === 'group' && !isWhatsappGroupUrl(values.destination)) {
      ctx.addIssue({ code: 'custom', path: ['destination'], message: WHATSAPP_GROUP_INVALID });
    }
  });

/**
 * Alta/edición de un canal de WhatsApp (ux.md §4.7, FR-005): nombre/propósito,
 * tipo (mensaje directo o grupo) y destino único con ayuda contextual; el
 * teléfono se valida con el criterio de F2 y el grupo debe ser un enlace https
 * de `chat.whatsapp.com`/`wa.me`.
 */
export function WhatsAppForm({ channel, onDone, onCancel }: WhatsAppFormProps) {
  const createChannel = useCreateWhatsappChannel();
  const updateChannel = useUpdateWhatsappChannel();
  const isEdit = Boolean(channel);

  const [langTab, setLangTab] = useState<'es' | 'en'>('es');
  const [fields, setFields] = useState<ChannelFields>(() => fieldsFromChannel(channel));
  const [errors, setErrors] = useState<Record<string, string>>({});

  const busy = createChannel.isPending || updateChannel.isPending;
  const serverError = isEdit ? updateChannel.error : createChannel.error;

  const serverFieldErrors: Record<string, string> = {};
  if (serverError) {
    for (const field of ['nameEs', 'nameEn', 'kind', 'destination']) {
      const message = detailMessage(serverError, field);
      if (message) {
        serverFieldErrors[field] = message;
      }
    }
  }
  const displayErrors = { ...errors, ...serverFieldErrors };
  const isConflict = serverError instanceof ApiError && serverError.status === 409;
  const serverMessage = isConflict
    ? WHATSAPP_DUPLICATE
    : serverError && Object.keys(serverFieldErrors).length === 0
      ? generalErrorMessage(serverError)
      : undefined;

  const update = (patch: Partial<ChannelFields>) =>
    setFields((current) => ({ ...current, ...patch }));

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const parsed = channelSchema.safeParse(fields);
    if (!parsed.success) {
      const mapped: Record<string, string> = {};
      for (const issue of parsed.error.issues) {
        const key = issue.path.join('.');
        if (!(key in mapped)) {
          mapped[key] = issue.message;
        }
      }
      setErrors(mapped);
      return;
    }
    setErrors({});
    const payload = {
      nameEs: fields.nameEs.trim(),
      nameEn: fields.nameEn.trim() === '' ? null : fields.nameEn.trim(),
      kind: fields.kind,
      destination: fields.destination.trim(),
    };

    if (channel) {
      const wasPublished = channel.publicationState === 'published';
      updateChannel.mutate(
        { id: channel.id, input: payload },
        { onSuccess: () => onDone({ message: wasPublished ? SAVED_PUBLISHED : SAVED }) },
      );
      return;
    }
    createChannel.mutate(payload, {
      onSuccess: () => onDone({ message: WHATSAPP_CREATED, draft: true }),
    });
  };

  return (
    <Dialog title={isEdit ? 'Editar canal' : 'Agregar canal'} onClose={onCancel}>
      <form onSubmit={handleSubmit} noValidate className="space-y-4">
        {serverMessage && <Notice variant="error">{serverMessage}</Notice>}

        <Tabs
          tabs={LANGUAGE_TABS}
          active={langTab}
          onChange={(id) => setLangTab(id as 'es' | 'en')}
          label="Idioma del contenido"
        />

        {langTab === 'es' ? (
          <Field
            label="Nombre o propósito (Español)"
            required
            value={fields.nameEs}
            error={displayErrors.nameEs}
            onChange={(event) => update({ nameEs: event.target.value })}
          />
        ) : (
          <>
            <p className="text-sm text-slate-600">{ENGLISH_HELP}</p>
            <Field
              label="Nombre o propósito (English)"
              value={fields.nameEn}
              onChange={(event) => update({ nameEn: event.target.value })}
            />
          </>
        )}

        <fieldset className="space-y-2 rounded border border-slate-200 p-4">
          <legend className="px-1 font-medium text-slate-900">Tipo de canal</legend>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              name="whatsapp-kind"
              checked={fields.kind === 'direct'}
              onChange={() => update({ kind: 'direct', destination: '' })}
            />
            Número para mensaje directo
          </label>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              name="whatsapp-kind"
              checked={fields.kind === 'group'}
              onChange={() => update({ kind: 'group', destination: '' })}
            />
            Enlace de grupo
          </label>
        </fieldset>

        <Field
          label={fields.kind === 'direct' ? 'Número' : 'Enlace del grupo'}
          required
          value={fields.destination}
          help={fields.kind === 'direct' ? WHATSAPP_PHONE_HELP : WHATSAPP_GROUP_HELP}
          error={displayErrors.destination}
          onChange={(event) => update({ destination: event.target.value })}
        />

        <div className="flex flex-wrap justify-end gap-3">
          <Button type="button" variant="secondary" onClick={onCancel} disabled={busy}>
            Cancelar
          </Button>
          <Button type="submit" loading={busy}>
            {isEdit ? 'Guardar cambios' : 'Agregar canal'}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
