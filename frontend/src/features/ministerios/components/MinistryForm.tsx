import { useEffect, useMemo, useState } from 'react';
import type { MinisterioAdmin, MinisterioInput } from '../../../api/ministerios';
import { ApiError } from '../../../api/client';
import { Button } from '../../../components/Button';
import { Dialog } from '../../../components/Dialog';
import { Field } from '../../../components/Field';
import { ImageUploader } from '../../../components/ImageUploader';
import { Notice } from '../../../components/Notice';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { TextArea } from '../../informacion/components/TextArea';
import { generalErrorMessage, detailMessage, isConflict } from '../errors';
import { useCreateMinistry } from '../hooks/useCreateMinistry';
import { useUpdateMinistry } from '../hooks/useUpdateMinistry';
import { useUploadMinistryImage } from '../hooks/useUploadMinistryImage';
import {
  ADD,
  CREATED,
  DESCRIPTION_MAX,
  DUPLICATE,
  ENGLISH_HELP,
  IMAGE_HELP,
  LEADER_NAME_MAX,
  LINK_HELP,
  NAME_MAX,
  PLACE_HELP,
  REQUIRED_IMAGE_ALT,
  SAVE,
  SAVED,
  SAVED_LEADER_HIDDEN,
  SAVED_LEADER_VISIBLE,
  SAVED_PUBLISHED,
  SCHEDULE_HELP,
  SCHEDULE_MAX,
  SORT_HELP,
  SYSTEM_ERROR,
  WHATSAPP_URL_MAX,
  counter,
  descriptionLimitMessage,
} from '../messages';
import {
  EMPTY_MINISTRY_FIELDS,
  FOCUS_ORDER,
  crossFieldErrors,
  issuesToErrors,
  ministryFieldsSchema,
  nullable,
  type ContactKind,
  type MinistryFields,
} from '../validation';
import { ContactoEncargado } from './ContactoEncargado';
import { ShowContactSwitch } from './ShowContactSwitch';

const LANGUAGE_TABS: TabItem[] = [
  { id: 'es', label: 'Español' },
  { id: 'en', label: 'English (opcional)' },
];

export interface MinistryFormProps {
  /** Ministerio a editar; sin él, alta. */
  ministry?: MinisterioAdmin;
  onDone: (feedback: { message: string; draft?: boolean; sharedLink?: boolean }) => void;
  onCancel: () => void;
}

/** Campos del servidor que se pintan junto a su control (para los `details`). */
const SERVER_FIELDS = [
  ...FOCUS_ORDER.map((entry) => entry.field),
  'leaderPhone',
  'leaderEmail',
  'imageFile',
  'imageAltEs',
  'imageAltEn',
  'contactValue',
] as const;

function fieldsFromMinistry(ministry?: MinisterioAdmin): MinistryFields {
  return {
    ...EMPTY_MINISTRY_FIELDS,
    nameEs: ministry?.nameEs ?? '',
    nameEn: ministry?.nameEn ?? '',
    descriptionEs: ministry?.descriptionEs ?? '',
    descriptionEn: ministry?.descriptionEn ?? '',
    scheduleEs: ministry?.scheduleEs ?? '',
    scheduleEn: ministry?.scheduleEn ?? '',
    placeEs: ministry?.placeEs ?? '',
    placeEn: ministry?.placeEn ?? '',
    leaderNameEs: ministry?.leaderNameEs ?? '',
    leaderNameEn: ministry?.leaderNameEn ?? '',
    whatsappUrl: ministry?.whatsappUrl ?? '',
  };
}

function contactFromMinistry(ministry?: MinisterioAdmin): {
  kind: ContactKind;
  value: string;
} {
  if (ministry?.leaderPhone) {
    return { kind: 'phone', value: ministry.leaderPhone };
  }
  if (ministry?.leaderEmail) {
    return { kind: 'email', value: ministry.leaderEmail };
  }
  return { kind: 'none', value: '' };
}

/**
 * Alta/edición de un ministerio (ux.md §4.4/§4.5, US2/FR-010/FR-014/FR-015):
 * `Dialog` con pestañas **Español** (obligatorio) / **English (opcional)** y la
 * regla **por campo** (un campo opcional rellenado en inglés sin su versión
 * española bloquea el guardado indicando qué falta; **un campo vacío no obliga a
 * nada**, Q6); horario y lugar como **textos libres** (FR-024: se muestran tal
 * como se escriban), bloque Encargado con `ContactoEncargado` (un solo medio,
 * Q8) y `ShowContactSwitch` —**apagado por defecto al crear** (D-13, «oculto por
 * defecto»)—, enlace `https` estricto, **Orden** numérico e imagen con `alt`
 * obligatorio en español.
 *
 * Los errores van junto al campo (los `details` del servidor se traducen por
 * campo) y el foco va al primer campo inválido, cambiando de pestaña si hace
 * falta (FR-014). Nada se guarda a medias y nada se pierde si el servidor
 * rechaza. `Esc` cierra el diálogo como en F3/F4 (mismo `Dialog`).
 */
export function MinistryForm({ ministry, onDone, onCancel }: MinistryFormProps) {
  const createMinistry = useCreateMinistry();
  const updateMinistry = useUpdateMinistry();
  const uploadImage = useUploadMinistryImage();
  const isEdit = Boolean(ministry);

  const [langTab, setLangTab] = useState<'es' | 'en'>('es');
  const [fields, setFields] = useState<MinistryFields>(() => fieldsFromMinistry(ministry));
  const initialContact = contactFromMinistry(ministry);
  const [contactKind, setContactKind] = useState<ContactKind>(initialContact.kind);
  const [contactValue, setContactValue] = useState(initialContact.value);
  // «Oculto por defecto» (D-13): al crear nace apagado; al editar manda la
  // última decisión guardada.
  const [leaderVisible, setLeaderVisible] = useState(ministry?.leaderVisible ?? false);
  const [sortOrder, setSortOrder] = useState(String(ministry?.sortOrder ?? 0));
  const [imageFile, setImageFile] = useState<string | null>(ministry?.imageFile ?? null);
  const [imageAltEs, setImageAltEs] = useState(ministry?.imageAltEs ?? '');
  const [imageAltEn, setImageAltEn] = useState(ministry?.imageAltEn ?? '');
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({});
  const [uploadError, setUploadError] = useState<string>();
  const [pendingFocus, setPendingFocus] = useState<{ id: string; nonce: number }>();

  const mutation = isEdit ? updateMinistry : createMinistry;
  const busy = createMinistry.isPending || updateMinistry.isPending || uploadImage.isPending;

  // El foco del primer campo inválido puede vivir en la otra pestaña de idioma:
  // se cambia de pestaña y se enfoca cuando el campo ya está en el DOM.
  useEffect(() => {
    if (pendingFocus) {
      document.getElementById(pendingFocus.id)?.focus();
    }
  }, [pendingFocus]);

  const serverErrors = useMemo(() => {
    const mapped: Record<string, string> = {};
    const error = mutation.error;
    if (error) {
      for (const field of SERVER_FIELDS) {
        const message = detailMessage(error, field);
        if (message) {
          mapped[field] = message;
        }
      }
    }
    return mapped;
  }, [mutation.error]);

  const errors = { ...clientErrors, ...serverErrors };

  const update = (patch: Partial<MinistryFields>) =>
    setFields((current) => ({ ...current, ...patch }));

  const hasLeader = fields.leaderNameEs.trim() !== '' || fields.leaderNameEn.trim() !== '';

  const focusFirstInvalid = (fieldErrors: Record<string, string>) => {
    const target = FOCUS_ORDER.find((entry) => fieldErrors[entry.field]);
    if (target) {
      setLangTab(target.tab);
      setPendingFocus({ id: target.id, nonce: Date.now() });
      return;
    }
    if (fieldErrors.leaderNameEs || fieldErrors.contactValue) {
      setPendingFocus({ id: 'ministry-leader-name-es', nonce: Date.now() });
    }
  };

  const handleUpload = (file: File) => {
    setUploadError(undefined);
    uploadImage.mutate(file, {
      onSuccess: (result) => setImageFile(result.fileName),
      onError: (error: ApiError) => {
        setUploadError(
          detailMessage(error, 'file') ??
            detailMessage(error, 'fileSize') ??
            (error.message || SYSTEM_ERROR),
        );
      },
    });
  };

  const handleSubmit = (formEvent: React.FormEvent<HTMLFormElement>) => {
    formEvent.preventDefault();
    const parsed = ministryFieldsSchema.safeParse(fields);
    const nextErrors: Record<string, string> = parsed.success ? {} : issuesToErrors(parsed.error);
    Object.assign(nextErrors, crossFieldErrors({ fields, contactKind, contactValue }));

    if (imageFile && imageAltEs.trim() === '') {
      nextErrors.imageAltEs = REQUIRED_IMAGE_ALT;
    }

    if (Object.keys(nextErrors).length > 0) {
      setClientErrors(nextErrors);
      focusFirstInvalid(nextErrors);
      return;
    }

    setClientErrors({});

    const order = Number.parseInt(sortOrder.trim(), 10);
    const payload: MinisterioInput = {
      nameEs: fields.nameEs.trim(),
      nameEn: nullable(fields.nameEn),
      descriptionEs: fields.descriptionEs.trim(),
      descriptionEn: nullable(fields.descriptionEn),
      scheduleEs: nullable(fields.scheduleEs),
      scheduleEn: nullable(fields.scheduleEn),
      placeEs: nullable(fields.placeEs),
      placeEn: nullable(fields.placeEn),
      leaderNameEs: nullable(fields.leaderNameEs),
      leaderNameEn: nullable(fields.leaderNameEn),
      leaderPhone: contactKind === 'phone' ? nullable(contactValue) : null,
      leaderEmail: contactKind === 'email' ? nullable(contactValue) : null,
      leaderVisible,
      // Sin enlace = `null` explícito, como el resto de opcionales (`nullable`):
      // el contrato admite `null` en `MinisterioInput`/`MinisterioPatch` (analyze
      // I1) y el `PATCH` lo interpreta como borrado. Nunca `""`: no cumple el
      // `pattern` del esquema. Quitar el enlace al editar, por tanto, lo borra.
      whatsappUrl: nullable(fields.whatsappUrl),
      sortOrder: Number.isNaN(order) ? 0 : order,
      imageFile,
      imageAltEs: imageFile ? nullable(imageAltEs) : null,
      imageAltEn: imageFile ? nullable(imageAltEn) : null,
    };

    if (ministry) {
      const wasPublished = ministry.publicationState === 'published';
      const visibilityChanged = ministry.leaderVisible !== leaderVisible;
      const message = visibilityChanged
        ? leaderVisible
          ? SAVED_LEADER_VISIBLE
          : SAVED_LEADER_HIDDEN
        : wasPublished
          ? SAVED_PUBLISHED
          : SAVED;
      updateMinistry.mutate(
        { id: ministry.id, input: payload },
        {
          onSuccess: (updated) =>
            onDone({ message, sharedLink: updated.whatsappUrlShared === true }),
        },
      );
      return;
    }

    createMinistry.mutate(payload, {
      onSuccess: (created) =>
        onDone({ message: CREATED, draft: true, sharedLink: created.whatsappUrlShared === true }),
    });
  };

  const summaryMessages = Object.values(errors).filter((message): message is string =>
    Boolean(message),
  );
  const conflict = isConflict(mutation.error);
  const showGeneralError = mutation.error && Object.keys(serverErrors).length === 0 && !conflict;

  return (
    <Dialog title={isEdit ? 'Editar ministerio' : ADD} size="wide" onClose={onCancel}>
      <form onSubmit={handleSubmit} noValidate className="space-y-4">
        {summaryMessages.length > 0 && (
          <Notice variant="error">
            <p className="font-medium">Revisa estos campos:</p>
            <ul className="mt-1 list-disc pl-5">
              {summaryMessages.map((message) => (
                <li key={message}>{message}</li>
              ))}
            </ul>
          </Notice>
        )}

        {conflict && <Notice variant="error">{DUPLICATE}</Notice>}
        {showGeneralError && <Notice variant="error">{generalErrorMessage(mutation.error)}</Notice>}
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
              id="ministry-name-es"
              label="Nombre (Español)"
              required
              value={fields.nameEs}
              help={counter(fields.nameEs.length, NAME_MAX)}
              error={errors.nameEs}
              onChange={(event) => update({ nameEs: event.target.value })}
            />
            <TextArea
              id="ministry-description-es"
              label="Descripción (Español)"
              required
              value={fields.descriptionEs}
              help={
                fields.descriptionEs.length > DESCRIPTION_MAX
                  ? descriptionLimitMessage(fields.descriptionEs.length)
                  : counter(fields.descriptionEs.length, DESCRIPTION_MAX)
              }
              error={errors.descriptionEs}
              onChange={(event) => update({ descriptionEs: event.target.value })}
            />
            <Field
              id="ministry-schedule-es"
              label="Horario (Español, opcional)"
              value={fields.scheduleEs}
              help={SCHEDULE_HELP}
              error={errors.scheduleEs}
              onChange={(event) => update({ scheduleEs: event.target.value })}
            />
            <Field
              id="ministry-place-es"
              label="Lugar (Español, opcional)"
              value={fields.placeEs}
              help={PLACE_HELP}
              error={errors.placeEs}
              onChange={(event) => update({ placeEs: event.target.value })}
            />
          </>
        ) : (
          <>
            <p className="text-sm text-slate-600">{ENGLISH_HELP}</p>
            <Field
              id="ministry-name-en"
              label="Nombre (English)"
              value={fields.nameEn}
              help={counter(fields.nameEn.length, NAME_MAX)}
              error={errors.nameEn}
              onChange={(event) => update({ nameEn: event.target.value })}
            />
            <TextArea
              id="ministry-description-en"
              label="Descripción (English)"
              value={fields.descriptionEn}
              help={counter(fields.descriptionEn.length, DESCRIPTION_MAX)}
              error={errors.descriptionEn}
              onChange={(event) => update({ descriptionEn: event.target.value })}
            />
            <Field
              id="ministry-schedule-en"
              label="Horario (English, opcional)"
              value={fields.scheduleEn}
              error={errors.scheduleEn}
              onChange={(event) => update({ scheduleEn: event.target.value })}
            />
            <Field
              id="ministry-place-en"
              label="Lugar (English, opcional)"
              value={fields.placeEn}
              error={errors.placeEn}
              onChange={(event) => update({ placeEn: event.target.value })}
            />
          </>
        )}

        <fieldset className="space-y-3 rounded border border-slate-200 p-4">
          <legend className="px-1 font-medium text-slate-900">Encargado (opcional)</legend>

          {langTab === 'es' ? (
            <Field
              id="ministry-leader-name-es"
              label="Nombre del encargado (Español)"
              value={fields.leaderNameEs}
              help={counter(fields.leaderNameEs.length, LEADER_NAME_MAX)}
              error={errors.leaderNameEs}
              onChange={(event) => update({ leaderNameEs: event.target.value })}
            />
          ) : (
            <Field
              id="ministry-leader-name-en"
              label="Nombre del encargado (English)"
              value={fields.leaderNameEn}
              help={counter(fields.leaderNameEn.length, LEADER_NAME_MAX)}
              error={errors.leaderNameEn}
              onChange={(event) => update({ leaderNameEn: event.target.value })}
            />
          )}

          <ContactoEncargado
            kind={contactKind}
            value={contactValue}
            error={errors.contactValue ?? errors.leaderPhone ?? errors.leaderEmail}
            onKindChange={setContactKind}
            onValueChange={setContactValue}
          />

          <ShowContactSwitch
            checked={leaderVisible}
            hasLeader={hasLeader}
            onChange={setLeaderVisible}
          />
        </fieldset>

        <Field
          id="ministry-whatsapp-url"
          label="Enlace del grupo de WhatsApp (opcional)"
          value={fields.whatsappUrl}
          help={LINK_HELP}
          error={errors.whatsappUrl}
          onChange={(event) => update({ whatsappUrl: event.target.value })}
        />

        <Field
          id="ministry-sort-order"
          label="Orden (opcional)"
          type="number"
          step={1}
          value={sortOrder}
          help={SORT_HELP}
          error={errors.sortOrder}
          onChange={(event) => setSortOrder(event.target.value)}
        />

        <ImageUploader
          label="Imagen del ministerio"
          help={IMAGE_HELP}
          currentUrl={imageFile ? `/api/v1/media/${imageFile}` : undefined}
          altEs={imageAltEs}
          onAltEsChange={setImageAltEs}
          altEn={imageAltEn}
          onAltEnChange={setImageAltEn}
          onFileSelected={handleUpload}
          uploading={uploadImage.isPending}
          error={errors.imageFile}
          altError={errors.imageAltEs}
        />

        <div className="flex flex-wrap justify-end gap-3">
          <Button type="button" variant="secondary" onClick={onCancel} disabled={busy}>
            Cancelar
          </Button>
          <Button type="submit" loading={busy}>
            {isEdit ? SAVE : ADD}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}

/** Reexporta el límite del enlace para que el formulario y las pruebas coincidan. */
export const WHATSAPP_URL_MAX_LENGTH = WHATSAPP_URL_MAX;
export const SCHEDULE_MAX_LENGTH = SCHEDULE_MAX;
