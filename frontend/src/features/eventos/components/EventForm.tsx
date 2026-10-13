import { useEffect, useMemo, useState } from 'react';
import type { CategoriaAdmin, EventoAdmin, EventoInput, FechaInput } from '../../../api/agenda';
import { ApiError } from '../../../api/client';
import { Button } from '../../../components/Button';
import { Dialog } from '../../../components/Dialog';
import { Field } from '../../../components/Field';
import { ImageUploader } from '../../../components/ImageUploader';
import { Notice } from '../../../components/Notice';
import { Select } from '../../../components/Select';
import { useCreateEvent, useUpdateEvent } from '../hooks/useEventMutations';
import { useUploadAgendaImage } from '../hooks/useUploadAgendaImage';
import { detailMessage, generalErrorMessage, isConflict } from '../errors';
import {
  ADD_EVENT,
  EVENT_CREATED,
  FREQUENCY_MAX,
  IMAGE_HELP,
  REQUIRED_EVENT_DATES,
  REQUIRED_IMAGE_ALT,
  SAVE,
  SAVED,
  SAVED_PUBLISHED,
  SYSTEM_ERROR,
  counter,
  duplicateMessage,
  textTooLong,
} from '../messages';
import {
  EMPTY_DATE_ROW,
  isDateRowsValid,
  issuesToErrors,
  itemFieldsSchema,
  nullable,
  validateDateRows,
  type DateRow,
  type DateRowsValidation,
  type ItemFields,
} from '../validation';
import { DatesEditor } from './DatesEditor';
import { ItemFormFields } from './ItemFormFields';

export interface EventFormProps {
  /** Evento a editar; sin él, alta. */
  event?: EventoAdmin;
  /** Catálogo de categorías para el `Select` (lo consulta la página). */
  categories: CategoriaAdmin[];
  onDone: (feedback: { message: string; draft?: boolean }) => void;
  onCancel: () => void;
}

interface FrequencyFields {
  frequencyEs: string;
  frequencyEn: string;
}

/** Campos con `id` propio, en el orden en que se enfoca el primero inválido. */
const FOCUS_ORDER: Array<{ field: string; id: string; tab: 'es' | 'en' }> = [
  { field: 'titleEs', id: 'agenda-title-es', tab: 'es' },
  { field: 'frequencyEs', id: 'agenda-frequency-es', tab: 'es' },
  { field: 'placeEs', id: 'agenda-place-es', tab: 'es' },
  { field: 'mapUrl', id: 'agenda-map-url', tab: 'es' },
  { field: 'detailEs', id: 'agenda-detail-es', tab: 'es' },
  { field: 'titleEn', id: 'agenda-title-en', tab: 'en' },
  { field: 'frequencyEn', id: 'agenda-frequency-en', tab: 'en' },
  { field: 'placeEn', id: 'agenda-place-en', tab: 'en' },
  { field: 'detailEn', id: 'agenda-detail-en', tab: 'en' },
];

const SERVER_FIELDS = [
  ...FOCUS_ORDER.map((entry) => entry.field),
  'organizerEs',
  'organizerEn',
  'categoryId',
  'imageFile',
  'imageAltEs',
  'imageAltEn',
] as const;

function fieldsFromEvent(event?: EventoAdmin): ItemFields {
  return {
    titleEs: event?.titleEs ?? '',
    titleEn: event?.titleEn ?? '',
    detailEs: event?.detailEs ?? '',
    detailEn: event?.detailEn ?? '',
    placeEs: event?.placeEs ?? '',
    placeEn: event?.placeEn ?? '',
    mapUrl: event?.mapUrl ?? '',
    organizerEs: event?.organizerEs ?? '',
    organizerEn: event?.organizerEn ?? '',
  };
}

function frequencyFromEvent(event?: EventoAdmin): FrequencyFields {
  return {
    frequencyEs: event?.frequencyEs ?? '',
    frequencyEn: event?.frequencyEn ?? '',
  };
}

/** Fechas del evento en filas del editor; el alta arranca con una fila vacía. */
function rowsFromEvent(event?: EventoAdmin): DateRow[] {
  if (!event || event.dates.length === 0) {
    return [{ ...EMPTY_DATE_ROW }];
  }
  return event.dates.map((date) => ({
    date: date.date,
    startTime: date.startTime ?? '',
    endTime: date.endTime ?? '',
  }));
}

function buildDates(rows: DateRow[]): FechaInput[] {
  return rows.map((row) => ({
    date: row.date,
    ...(row.startTime ? { startTime: row.startTime } : {}),
    ...(row.endTime ? { endTime: row.endTime } : {}),
  }));
}

/**
 * Alta/edición de un evento del panel (ux.md §4.4, FR-010/FR-015/FR-018):
 * `Dialog` con pestañas de idioma, lista de fechas (`DatesEditor`), categoría,
 * imagen con `alt` obligatorio en español y las validaciones espejo de FR-015
 * (incluida la decisión firme: **la hora de fin exige la de inicio**). Guardar
 * sin tocar el idioma inglés es válido (Decisión 8) y, si el evento ya estaba
 * publicado, el aviso lo dice («se publica al guardar», FR-014).
 */
export function EventForm({ event, categories, onDone, onCancel }: EventFormProps) {
  const createEvent = useCreateEvent();
  const updateEvent = useUpdateEvent();
  const uploadImage = useUploadAgendaImage('eventos');
  const isEdit = Boolean(event);

  const [langTab, setLangTab] = useState<'es' | 'en'>('es');
  const [fields, setFields] = useState<ItemFields>(() => fieldsFromEvent(event));
  const [frequency, setFrequency] = useState<FrequencyFields>(() => frequencyFromEvent(event));
  const [rows, setRows] = useState<DateRow[]>(() => rowsFromEvent(event));
  const [categoryId, setCategoryId] = useState(() => event?.categoryId ?? '');
  const [imageFile, setImageFile] = useState<string | null>(event?.imageFile ?? null);
  const [imageAltEs, setImageAltEs] = useState(event?.imageAltEs ?? '');
  const [imageAltEn, setImageAltEn] = useState(event?.imageAltEn ?? '');
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({});
  const [dateValidation, setDateValidation] = useState<DateRowsValidation>({ rows: {} });
  const [uploadError, setUploadError] = useState<string>();
  const [pendingFocus, setPendingFocus] = useState<{ id: string; nonce: number }>();

  const mutation = isEdit ? updateEvent : createEvent;
  const busy = createEvent.isPending || updateEvent.isPending || uploadImage.isPending;

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
  const serverDatesError =
    detailMessage(mutation.error, 'dates') ?? detailMessage(mutation.error, 'endTime');

  const categoryOptions = [
    { value: '', label: '— Sin categoría —' },
    ...categories.map((category) => ({ value: category.id, label: category.nameEs })),
  ];

  const update = (patch: Partial<ItemFields>) => setFields((current) => ({ ...current, ...patch }));

  const focusFirstInvalid = (fieldErrors: Record<string, string>, datesInvalid: boolean) => {
    const target = FOCUS_ORDER.find((entry) => fieldErrors[entry.field]);
    if (target) {
      setLangTab(target.tab);
      setPendingFocus({ id: target.id, nonce: Date.now() });
      return;
    }
    if (datesInvalid) {
      setLangTab('es');
      setPendingFocus({ id: 'agenda-date-0', nonce: Date.now() });
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
    const parsedFields = itemFieldsSchema.safeParse(fields);
    const nextErrors: Record<string, string> = parsedFields.success
      ? {}
      : issuesToErrors(parsedFields.error);

    if (frequency.frequencyEs.trim().length > FREQUENCY_MAX) {
      nextErrors.frequencyEs = textTooLong(FREQUENCY_MAX);
    }
    if (frequency.frequencyEn.trim().length > FREQUENCY_MAX) {
      nextErrors.frequencyEn = textTooLong(FREQUENCY_MAX);
    }
    if (imageFile && imageAltEs.trim() === '') {
      nextErrors.imageAltEs = REQUIRED_IMAGE_ALT;
    }

    const validation = validateDateRows(rows, REQUIRED_EVENT_DATES);
    const datesInvalid = !isDateRowsValid(validation);

    if (Object.keys(nextErrors).length > 0 || datesInvalid) {
      setClientErrors(nextErrors);
      setDateValidation(validation);
      focusFirstInvalid(nextErrors, datesInvalid);
      return;
    }

    setClientErrors({});
    setDateValidation({ rows: {} });

    const payload: EventoInput = {
      titleEs: fields.titleEs.trim(),
      titleEn: nullable(fields.titleEn),
      frequencyEs: nullable(frequency.frequencyEs),
      frequencyEn: nullable(frequency.frequencyEn),
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
      dates: buildDates(rows),
    };

    if (event) {
      const wasPublished = event.publicationState === 'published';
      updateEvent.mutate(
        { id: event.id, input: payload },
        { onSuccess: () => onDone({ message: wasPublished ? SAVED_PUBLISHED : SAVED }) },
      );
      return;
    }
    createEvent.mutate(payload, {
      onSuccess: () => onDone({ message: EVENT_CREATED, draft: true }),
    });
  };

  const summaryMessages = [
    ...Object.values(errors),
    dateValidation.general,
    ...Object.values(dateValidation.rows).flatMap((rowErrors) => Object.values(rowErrors)),
  ].filter((message): message is string => Boolean(message));
  const conflict = isConflict(mutation.error);
  const showGeneralError =
    mutation.error && Object.keys(serverErrors).length === 0 && !conflict && !serverDatesError;

  return (
    <Dialog title={isEdit ? 'Editar evento' : ADD_EVENT} size="wide" onClose={onCancel}>
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

        {conflict && <Notice variant="error">{duplicateMessage('evento')}</Notice>}
        {showGeneralError && <Notice variant="error">{generalErrorMessage(mutation.error)}</Notice>}
        {uploadError && <Notice variant="error">{uploadError}</Notice>}

        <ItemFormFields
          fields={fields}
          onChange={update}
          errors={errors}
          langTab={langTab}
          onLangTabChange={setLangTab}
          spanishExtra={
            <Field
              id="agenda-frequency-es"
              label="Frecuencia (opcional, Español)"
              value={frequency.frequencyEs}
              help={counter(frequency.frequencyEs.length, FREQUENCY_MAX)}
              error={errors.frequencyEs}
              onChange={(inputEvent) =>
                setFrequency((current) => ({ ...current, frequencyEs: inputEvent.target.value }))
              }
            />
          }
          englishExtra={
            <Field
              id="agenda-frequency-en"
              label="Frecuencia (opcional, English)"
              value={frequency.frequencyEn}
              help={counter(frequency.frequencyEn.length, FREQUENCY_MAX)}
              error={errors.frequencyEn}
              onChange={(inputEvent) =>
                setFrequency((current) => ({ ...current, frequencyEn: inputEvent.target.value }))
              }
            />
          }
        >
          <DatesEditor
            rows={rows}
            onChange={setRows}
            errors={dateValidation.rows}
            generalError={
              dateValidation.general ?? serverDatesError ?? detailMessage(mutation.error, 'dates')
            }
          />

          <Select
            id="agenda-category"
            label="Categoría (opcional)"
            options={categoryOptions}
            value={categoryId}
            error={errors.categoryId}
            onChange={(selectEvent) => setCategoryId(selectEvent.target.value)}
          />

          <ImageUploader
            label="Imagen del evento"
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
            {isEdit ? SAVE : ADD_EVENT}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
