import { useMemo, useState } from 'react';
import { z } from 'zod';
import type { IdentityAdmin, IdentityInput, PublicationState } from '../../../api/portada';
import { ApiError } from '../../../api/client';
import { Button } from '../../../components/Button';
import { Field } from '../../../components/Field';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { detailMessage, generalErrorMessage } from '../errors';
import { useIdentity } from '../hooks/useIdentity';
import { useUpdateIdentity } from '../hooks/useUpdateIdentity';
import { useUploadImage } from '../hooks/useUploadImage';
import {
  COVER_HELP,
  DRAFT_NOTICE,
  ENGLISH_HELP,
  LOGO_HELP,
  PUBLISHED,
  REQUIRED_ALT,
  REQUIRED_NAME,
  SAVED,
  SAVED_PUBLISHED,
  SYSTEM_ERROR,
  UNPUBLISHED,
} from '../messages';
import { ImageUploader } from './ImageUploader';
import { PublishControls } from './PublishControls';
import { TextArea } from './TextArea';

interface IdentityFields {
  nameEs: string;
  nameEn: string;
  taglineEs: string;
  taglineEn: string;
  missionEs: string;
  missionEn: string;
  visionEs: string;
  visionEn: string;
  logoFile: string | null;
  logoAltEs: string;
  logoAltEn: string;
  coverImageFile: string | null;
  coverImageAltEs: string;
  coverImageAltEn: string;
}

const LANGUAGE_TABS: TabItem[] = [
  { id: 'es', label: 'Español' },
  { id: 'en', label: 'English (opcional)' },
];

const identitySchema = z
  .object({
    nameEs: z.string().trim().min(1, REQUIRED_NAME),
    logoFile: z.string().nullable(),
    logoAltEs: z.string(),
    coverImageFile: z.string().nullable(),
    coverImageAltEs: z.string(),
  })
  .superRefine((values, ctx) => {
    if (values.logoFile && values.logoAltEs.trim() === '') {
      ctx.addIssue({ code: 'custom', path: ['logoAltEs'], message: REQUIRED_ALT });
    }
    if (values.coverImageFile && values.coverImageAltEs.trim() === '') {
      ctx.addIssue({ code: 'custom', path: ['coverImageAltEs'], message: REQUIRED_ALT });
    }
  });

const SERVER_FIELDS = [
  'nameEs',
  'nameEn',
  'taglineEs',
  'taglineEn',
  'missionEs',
  'missionEn',
  'visionEs',
  'visionEn',
  'logoAltEs',
  'logoAltEn',
  'coverImageAltEs',
  'coverImageAltEn',
] as const;

function fieldsFromIdentity(identity: IdentityAdmin | null): IdentityFields {
  return {
    nameEs: identity?.nameEs ?? '',
    nameEn: identity?.nameEn ?? '',
    taglineEs: identity?.taglineEs ?? '',
    taglineEn: identity?.taglineEn ?? '',
    missionEs: identity?.missionEs ?? '',
    missionEn: identity?.missionEn ?? '',
    visionEs: identity?.visionEs ?? '',
    visionEn: identity?.visionEn ?? '',
    logoFile: identity?.logoFile ?? null,
    logoAltEs: identity?.logoAltEs ?? '',
    logoAltEn: identity?.logoAltEn ?? '',
    coverImageFile: identity?.coverImageFile ?? null,
    coverImageAltEs: identity?.coverImageAltEs ?? '',
    coverImageAltEn: identity?.coverImageAltEn ?? '',
  };
}

function toNullable(value: string): string | null {
  const trimmed = value.trim();
  return trimmed === '' ? null : trimmed;
}

function buildInput(fields: IdentityFields, state: PublicationState): IdentityInput {
  return {
    nameEs: fields.nameEs.trim(),
    nameEn: toNullable(fields.nameEn),
    taglineEs: toNullable(fields.taglineEs),
    taglineEn: toNullable(fields.taglineEn),
    missionEs: toNullable(fields.missionEs),
    missionEn: toNullable(fields.missionEn),
    visionEs: toNullable(fields.visionEs),
    visionEn: toNullable(fields.visionEn),
    logoFile: fields.logoFile,
    logoAltEs: toNullable(fields.logoAltEs),
    logoAltEn: toNullable(fields.logoAltEn),
    coverImageFile: fields.coverImageFile,
    coverImageAltEs: toNullable(fields.coverImageAltEs),
    coverImageAltEn: toNullable(fields.coverImageAltEn),
    publicationState: state,
  };
}

function imageUrl(file: string | null): string | undefined {
  return file ? `/api/v1/media/${file}` : undefined;
}

interface Feedback {
  variant: NoticeVariant;
  message: string;
}

/**
 * Formulario de identidad (ux.md §4.4, FR-002/FR-011/FR-015/FR-019): nombre
 * oficial (es obligatorio, en opcional), lema/misión/visión con pestañas de
 * idioma, logotipo e imagen de portada con previsualización y `alt` (es
 * obligatorio, en opcional) y publicación **por sección** con `PublishControls`
 * (FR-013/D-2). Validación de cliente con Zod (espejo de FR-015); la autoridad
 * es el servidor, cuyos `details` se muestran junto a cada campo.
 */
export function IdentityForm() {
  const { identity } = useIdentity();
  const updateIdentity = useUpdateIdentity();
  const uploadImage = useUploadImage();

  const [langTab, setLangTab] = useState<'es' | 'en'>('es');
  const [fields, setFields] = useState<IdentityFields>(() => fieldsFromIdentity(identity));
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({});
  const [feedback, setFeedback] = useState<Feedback | null>(null);
  const [draftNotice, setDraftNotice] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);

  // Reajuste de estado al cambiar la identidad del servidor (carga inicial,
  // guardado y publicación/retiro) sin un efecto que re-renderice en cascada.
  const [syncedRevision, setSyncedRevision] = useState(identity?.updatedAt ?? null);
  const revision = identity?.updatedAt ?? null;
  if (revision !== syncedRevision) {
    setSyncedRevision(revision);
    setFields(fieldsFromIdentity(identity));
  }

  const serverErrors = useMemo(() => {
    const mapped: Record<string, string> = {};
    if (updateIdentity.error) {
      for (const field of SERVER_FIELDS) {
        const message = detailMessage(updateIdentity.error, field);
        if (message) {
          mapped[field] = message;
        }
      }
    }
    return mapped;
  }, [updateIdentity.error]);

  const errors = { ...clientErrors, ...serverErrors };

  const update = (patch: Partial<IdentityFields>) => {
    setFields((current) => ({ ...current, ...patch }));
  };

  const handleUpload = (kind: 'logo' | 'cover', file: File) => {
    setUploadError(null);
    uploadImage.mutate(file, {
      onSuccess: (result) =>
        update(
          kind === 'logo' ? { logoFile: result.fileName } : { coverImageFile: result.fileName },
        ),
      onError: (error: ApiError) => {
        const message =
          detailMessage(error, 'file') ??
          detailMessage(error, 'fileSize') ??
          (error.message || SYSTEM_ERROR);
        setUploadError(message);
      },
    });
  };

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setFeedback(null);
    setDraftNotice(false);
    const parsed = identitySchema.safeParse(fields);
    if (!parsed.success) {
      const mapped: Record<string, string> = {};
      for (const issue of parsed.error.issues) {
        const key = issue.path.join('.');
        if (!(key in mapped)) {
          mapped[key] = issue.message;
        }
      }
      setClientErrors(mapped);
      return;
    }
    setClientErrors({});
    const wasPublished = identity?.publicationState === 'published';
    const isCreate = identity === null;
    updateIdentity.mutate(buildInput(fields, identity?.publicationState ?? 'draft'), {
      onSuccess: () => {
        setFeedback({ variant: 'success', message: wasPublished ? SAVED_PUBLISHED : SAVED });
        setDraftNotice(isCreate);
      },
    });
  };

  const handlePublish = () => {
    if (!identity) {
      return;
    }
    setFeedback(null);
    setDraftNotice(false);
    updateIdentity.mutate(buildInput(fieldsFromIdentity(identity), 'published'), {
      onSuccess: () => setFeedback({ variant: 'success', message: PUBLISHED }),
    });
  };

  const handleUnpublish = () => {
    if (!identity) {
      return;
    }
    setFeedback(null);
    setDraftNotice(false);
    updateIdentity.mutate(buildInput(fieldsFromIdentity(identity), 'draft'), {
      onSuccess: () => setFeedback({ variant: 'success', message: UNPUBLISHED }),
    });
  };

  const busy = updateIdentity.isPending;

  return (
    <form onSubmit={handleSubmit} noValidate className="space-y-4">
      {feedback && <Notice variant={feedback.variant}>{feedback.message}</Notice>}
      {draftNotice && <Notice variant="info">{DRAFT_NOTICE}</Notice>}
      {updateIdentity.error && Object.keys(serverErrors).length === 0 && (
        <Notice variant="error">{generalErrorMessage(updateIdentity.error)}</Notice>
      )}
      {uploadError && <Notice variant="error">{uploadError}</Notice>}

      <Tabs
        tabs={LANGUAGE_TABS}
        active={langTab}
        onChange={(id) => setLangTab(id as 'es' | 'en')}
        label="Idioma del contenido"
      />

      {langTab === 'es' ? (
        <>
          <Field
            label="Nombre oficial"
            required
            value={fields.nameEs}
            error={errors.nameEs}
            onChange={(event) => update({ nameEs: event.target.value })}
          />
          <Field
            label="Lema (Español)"
            value={fields.taglineEs}
            onChange={(event) => update({ taglineEs: event.target.value })}
          />
          <TextArea
            label="Misión (Español)"
            value={fields.missionEs}
            onChange={(event) => update({ missionEs: event.target.value })}
          />
          <TextArea
            label="Visión (Español)"
            value={fields.visionEs}
            onChange={(event) => update({ visionEs: event.target.value })}
          />
        </>
      ) : (
        <>
          <p className="text-sm text-slate-600">{ENGLISH_HELP}</p>
          <Field
            label="Nombre oficial (English)"
            value={fields.nameEn}
            onChange={(event) => update({ nameEn: event.target.value })}
          />
          <Field
            label="Lema (English)"
            value={fields.taglineEn}
            onChange={(event) => update({ taglineEn: event.target.value })}
          />
          <TextArea
            label="Misión (English)"
            value={fields.missionEn}
            onChange={(event) => update({ missionEn: event.target.value })}
          />
          <TextArea
            label="Visión (English)"
            value={fields.visionEn}
            onChange={(event) => update({ visionEn: event.target.value })}
          />
        </>
      )}

      <ImageUploader
        label="Logotipo"
        help={LOGO_HELP}
        currentUrl={imageUrl(fields.logoFile)}
        altEs={fields.logoAltEs}
        onAltEsChange={(value) => update({ logoAltEs: value })}
        altEn={fields.logoAltEn}
        onAltEnChange={(value) => update({ logoAltEn: value })}
        onFileSelected={(file) => handleUpload('logo', file)}
        uploading={uploadImage.isPending}
        altError={errors.logoAltEs}
      />
      <ImageUploader
        label="Imagen de portada"
        help={COVER_HELP}
        currentUrl={imageUrl(fields.coverImageFile)}
        altEs={fields.coverImageAltEs}
        onAltEsChange={(value) => update({ coverImageAltEs: value })}
        altEn={fields.coverImageAltEn}
        onAltEnChange={(value) => update({ coverImageAltEn: value })}
        onFileSelected={(file) => handleUpload('cover', file)}
        uploading={uploadImage.isPending}
        altError={errors.coverImageAltEs}
      />

      <div className="flex flex-wrap items-center justify-end gap-3">
        <Button type="submit" loading={busy}>
          Guardar
        </Button>
        {identity && (
          <PublishControls
            name={identity.nameEs}
            state={identity.publicationState}
            loading={busy}
            onPublish={handlePublish}
            onUnpublish={handleUnpublish}
          />
        )}
      </div>
    </form>
  );
}
