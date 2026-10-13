import { useEffect, useMemo, useState } from 'react';
import type { ActividadAdmin, ActividadInput, CategoriaAdmin } from '../../../api/agenda';
import { ApiError } from '../../../api/client';
import { Button } from '../../../components/Button';
import { Dialog } from '../../../components/Dialog';
import { Field } from '../../../components/Field';
import { ImageUploader } from '../../../components/ImageUploader';
import { Notice } from '../../../components/Notice';
import { Select } from '../../../components/Select';
import { useCreateActivity, useUpdateActivity } from '../hooks/useActivityMutations';
import { useUploadAgendaImage } from '../hooks/useUploadAgendaImage';
import { detailMessage, generalErrorMessage, isConflict } from '../errors';
import {
  ACTIVITY_CREATED,
  ADD_ACTIVITY,
  IMAGE_HELP,
  REQUIRED_ACTIVITY_DATE,
  REQUIRED_IMAGE_ALT,
  SAVE,
  SAVED,
  SAVED_PUBLISHED,
  SYSTEM_ERROR,
  duplicateMessage,
} from '../messages';
import {
  issuesToErrors,
  itemFieldsSchema,
  nullable,
  validateOccurrence,
  type ItemFields,
  type OccurrenceErrors,
} from '../validation';
import { ItemFormFields } from './ItemFormFields';

export interface ActivityFormProps {
  /** Actividad a editar; sin ella, alta. */
  activity?: ActividadAdmin;
  categories: CategoriaAdmin[];
  onDone: (feedback: { message: string; draft?: boolean }) => void;
  onCancel: () => void;
}

const FOCUS_ORDER: Array<{ field: string; id: string; tab: 'es' | 'en' }> = [
  { field: 'titleEs', id: 'agenda-title-es', tab: 'es' },
  { field: 'date', id: 'agenda-activity-date', tab: 'es' },
  { field: 'placeEs', id: 'agenda-place-es', tab: 'es' },
  { field: 'mapUrl', id: 'agenda-map-url', tab: 'es' },
  { field: 'detailEs', id: 'agenda-detail-es', tab: 'es' },
  { field: 'titleEn', id: 'agenda-title-en', tab: 'en' },
  { field: 'placeEn', id: 'agenda-place-en', tab: 'en' },
  { field: 'detailEn', id: 'agenda-detail-en', tab: 'en' },
];

const SERVER_FIELDS = [
  ...FOCUS_ORDER.map((entry) => entry.field),
  'startTime',
  'endTime',
  'organizerEs',
  'organizerEn',
  'categoryId',
  'imageFile',
  'imageAltEs',
  'imageAltEn',
] as const;

function fieldsFromActivity(activity?: ActividadAdmin): ItemFields {
  return {
    titleEs: activity?.titleEs ?? '',
    titleEn: activity?.titleEn ?? '',
    detailEs: activity?.detailEs ?? '',
    detailEn: activity?.detailEn ?? '',
    placeEs: activity?.placeEs ?? '',
    placeEn: activity?.placeEn ?? '',
    mapUrl: activity?.mapUrl ?? '',
    organizerEs: activity?.organizerEs ?? '',
    organizerEn: activity?.organizerEn ?? '',
  };
}

/**
 * Alta/edición de una actividad puntual (ux.md §4.5, FR-011/FR-015/FR-019):
 * `Dialog` con las pestañas de idioma y los campos comunes (`ItemFormFields`)
 * más **una sola fecha** con hora de inicio y hora de fin opcionales —sin
 * periodicidad ni texto de frecuencia (Decisión 3)—. Las reglas de horario son
 * las mismas que en los eventos (`endTime` exige `startTime` y es posterior).
 */
export function ActivityForm({ activity, categories, onDone, onCancel }: ActivityFormProps) {
  const createActivity = useCreateActivity();
  const updateActivity = useUpdateActivity();
  const uploadImage = useUploadAgendaImage('actividades');
  const isEdit = Boolean(activity);

  const [langTab, setLangTab] = useState<'es' | 'en'>('es');
  const [fields, setFields] = useState<ItemFields>(() => fieldsFromActivity(activity));
  const [date, setDate] = useState(activity?.date ?? '');
  const [startTime, setStartTime] = useState(activity?.startTime ?? '');
  const [endTime, setEndTime] = useState(activity?.endTime ?? '');
  const [categoryId, setCategoryId] = useState(() => activity?.categoryId ?? '');
  const [imageFile, setImageFile] = useState<string | null>(activity?.imageFile ?? null);
  const [imageAltEs, setImageAltEs] = useState(activity?.imageAltEs ?? '');
  const [imageAltEn, setImageAltEn] = useState(activity?.imageAltEn ?? '');
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({});
  const [occurrenceErrors, setOccurrenceErrors] = useState<OccurrenceErrors>({});
  const [uploadError, setUploadError] = useState<string>();
  const [pendingFocus, setPendingFocus] = useState<{ id: string; nonce: number }>();

  const mutation = isEdit ? updateActivity : createActivity;
  const busy = createActivity.isPending || updateActivity.isPending || uploadImage.isPending;

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

  const categoryOptions = [
    { value: '', label: '— Sin categoría —' },
    ...categories.map((category) => ({ value: category.id, label: category.nameEs })),
  ];

  const update = (patch: Partial<ItemFields>) => setFields((current) => ({ ...current, ...patch }));

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
    const parsedFields = itemFieldsSchema.safeParse(fields);
    const nextErrors: Record<string, string> = parsedFields.success
      ? {}
      : issuesToErrors(parsedFields.error);

    if (imageFile && imageAltEs.trim() === '') {
      nextErrors.imageAltEs = REQUIRED_IMAGE_ALT;
    }

    const occurrence = validateOccurrence(date, startTime, endTime, REQUIRED_ACTIVITY_DATE);
    for (const [field, message] of Object.entries(occurrence)) {
      if (message) {
        nextErrors[field] = message;
      }
    }

    if (Object.keys(nextErrors).length > 0) {
      setClientErrors(nextErrors);
      setOccurrenceErrors(occurrence);
      const target = FOCUS_ORDER.find((entry) => nextErrors[entry.field]);
      if (target) {
        setLangTab(target.tab);
        setPendingFocus({ id: target.id, nonce: Date.now() });
      }
      return;
    }

    setClientErrors({});
    setOccurrenceErrors({});

    const payload: ActividadInput = {
      titleEs: fields.titleEs.trim(),
      titleEn: nullable(fields.titleEn),
      detailEs: fields.detailEs.trim(),
      detailEn: nullable(fields.detailEn),
      placeEs: fields.placeEs.trim(),
      placeEn: nullable(fields.placeEn),
      mapUrl: nullable(fields.mapUrl),
      organizerEs: nullable(fields.organizerEs),
      organizerEn: nullable(fields.organizerEn),
      categoryId: categoryId === '' ? null : categoryId,
      imageFile,
      imageAltEs: imageFile ? nullable(imageAltEs) : null,
      imageAltEn: imageFile ? nullable(imageAltEn) : null,
      date,
      ...(startTime ? { startTime } : {}),
      ...(endTime ? { endTime } : {}),
    };

    if (activity) {
      const wasPublished = activity.publicationState === 'published';
      updateActivity.mutate(
        { id: activity.id, input: payload },
        { onSuccess: () => onDone({ message: wasPublished ? SAVED_PUBLISHED : SAVED }) },
      );
      return;
    }
    createActivity.mutate(payload, {
      onSuccess: () => onDone({ message: ACTIVITY_CREATED, draft: true }),
    });
  };

  const summaryMessages = [...Object.values(errors), ...Object.values(occurrenceErrors)].filter(
    (message): message is string => Boolean(message),
  );
  const conflict = isConflict(mutation.error);
  const showGeneralError = mutation.error && Object.keys(serverErrors).length === 0 && !conflict;

  return (
    <Dialog title={isEdit ? 'Editar actividad' : ADD_ACTIVITY} size="wide" onClose={onCancel}>
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

        {conflict && <Notice variant="error">{duplicateMessage('actividad')}</Notice>}
        {showGeneralError && <Notice variant="error">{generalErrorMessage(mutation.error)}</Notice>}
        {uploadError && <Notice variant="error">{uploadError}</Notice>}

        <ItemFormFields
          fields={fields}
          onChange={update}
          errors={errors}
          langTab={langTab}
          onLangTabChange={setLangTab}
        >
          <fieldset className="space-y-3 rounded border border-slate-200 p-4">
            <legend className="px-1 font-medium text-slate-900">Fecha de la actividad</legend>
            <Field
              id="agenda-activity-date"
              label="Fecha"
              type="date"
              required
              value={date}
              error={occurrenceErrors.date}
              onChange={(inputEvent) => setDate(inputEvent.target.value)}
            />
            <Field
              label="Hora de inicio (opcional)"
              type="time"
              value={startTime}
              error={occurrenceErrors.startTime}
              onChange={(inputEvent) => setStartTime(inputEvent.target.value)}
            />
            <Field
              label="Hora de fin (opcional)"
              type="time"
              value={endTime}
              error={occurrenceErrors.endTime}
              onChange={(inputEvent) => setEndTime(inputEvent.target.value)}
            />
          </fieldset>

          <Select
            id="agenda-category"
            label="Categoría (opcional)"
            options={categoryOptions}
            value={categoryId}
            error={errors.categoryId}
            onChange={(selectEvent) => setCategoryId(selectEvent.target.value)}
          />

          <ImageUploader
            label="Imagen de la actividad"
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
        </ItemFormFields>

        <div className="flex flex-wrap justify-end gap-3">
          <Button type="button" variant="secondary" onClick={onCancel} disabled={busy}>
            Cancelar
          </Button>
          <Button type="submit" loading={busy}>
            {isEdit ? SAVE : ADD_ACTIVITY}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
