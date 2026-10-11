import { useState } from 'react';
import { z } from 'zod';
import type { ScheduleItemAdmin } from '../../../api/portada';
import { Button } from '../../../components/Button';
import { Dialog } from '../../../components/Dialog';
import { Field } from '../../../components/Field';
import { Notice } from '../../../components/Notice';
import { Select } from '../../../components/Select';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { detailMessage, generalErrorMessage } from '../errors';
import { useCreateService } from '../hooks/useCreateService';
import { useUpdateService } from '../hooks/useUpdateService';
import {
  DAY_OF_WEEK_LABELS,
  END_TIME_INVALID,
  ENGLISH_HELP,
  REQUIRED_PLACE,
  REQUIRED_SERVICE_NAME,
  REQUIRED_START_TIME,
  SAVED,
  SAVED_PUBLISHED,
  SERVICE_CREATED,
  START_TIME_INVALID,
} from '../messages';
import { TextArea } from './TextArea';

export interface ServiceFormProps {
  service?: ScheduleItemAdmin;
  onDone: (feedback: { message: string; draft?: boolean }) => void;
  onCancel: () => void;
}

const TIME_PATTERN = /^([01]\d|2[0-3]):[0-5]\d$/;

const LANGUAGE_TABS: TabItem[] = [
  { id: 'es', label: 'Español' },
  { id: 'en', label: 'English (opcional)' },
];

const DAY_OPTIONS = DAY_OF_WEEK_LABELS.map((label, value) => ({ value: String(value), label }));

interface ServiceFields {
  dayOfWeek: string;
  startTime: string;
  endTime: string;
  nameEs: string;
  nameEn: string;
  descriptionEs: string;
  descriptionEn: string;
  placeEs: string;
  placeEn: string;
}

function fieldsFromService(service?: ScheduleItemAdmin): ServiceFields {
  return {
    dayOfWeek: String(service?.dayOfWeek ?? 0),
    startTime: service?.startTime ?? '',
    endTime: service?.endTime ?? '',
    nameEs: service?.nameEs ?? '',
    nameEn: service?.nameEn ?? '',
    descriptionEs: service?.descriptionEs ?? '',
    descriptionEn: service?.descriptionEn ?? '',
    placeEs: service?.placeEs ?? '',
    placeEn: service?.placeEn ?? '',
  };
}

const serviceSchema = z
  .object({
    dayOfWeek: z.string(),
    startTime: z.string().regex(TIME_PATTERN, REQUIRED_START_TIME),
    endTime: z.string(),
    nameEs: z.string().trim().min(1, REQUIRED_SERVICE_NAME),
    placeEs: z.string().trim().min(1, REQUIRED_PLACE),
  })
  .superRefine((values, ctx) => {
    if (values.endTime.trim() === '') {
      return;
    }
    if (!TIME_PATTERN.test(values.endTime)) {
      ctx.addIssue({ code: 'custom', path: ['endTime'], message: START_TIME_INVALID });
    } else if (values.endTime <= values.startTime) {
      ctx.addIssue({ code: 'custom', path: ['endTime'], message: END_TIME_INVALID });
    }
  });

function nullable(value: string): string | null {
  const trimmed = value.trim();
  return trimmed === '' ? null : trimmed;
}

/**
 * Alta/edición de un servicio del horario (ux.md §4.6, analyze C2): día en
 * `Select` con las 7 opciones (nunca texto libre), hora de inicio «HH:MM»
 * obligatoria y hora de fin **opcional** posterior a la de inicio; nombre,
 * descripción y lugar con pestaña «English (opcional)».
 */
export function ServiceForm({ service, onDone, onCancel }: ServiceFormProps) {
  const createService = useCreateService();
  const updateService = useUpdateService();
  const isEdit = Boolean(service);

  const [langTab, setLangTab] = useState<'es' | 'en'>('es');
  const [fields, setFields] = useState<ServiceFields>(() => fieldsFromService(service));
  const [errors, setErrors] = useState<Record<string, string>>({});

  const mutation = isEdit ? updateService : createService;
  const busy = createService.isPending || updateService.isPending;

  const update = (patch: Partial<ServiceFields>) =>
    setFields((current) => ({ ...current, ...patch }));

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const parsed = serviceSchema.safeParse(fields);
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
      dayOfWeek: Number(fields.dayOfWeek),
      startTime: fields.startTime,
      endTime: nullable(fields.endTime),
      nameEs: fields.nameEs.trim(),
      nameEn: nullable(fields.nameEn),
      descriptionEs: nullable(fields.descriptionEs),
      descriptionEn: nullable(fields.descriptionEn),
      placeEs: fields.placeEs.trim(),
      placeEn: nullable(fields.placeEn),
    };

    if (service) {
      const wasPublished = service.publicationState === 'published';
      updateService.mutate(
        { id: service.id, input: payload },
        { onSuccess: () => onDone({ message: wasPublished ? SAVED_PUBLISHED : SAVED }) },
      );
      return;
    }
    createService.mutate(payload, {
      onSuccess: () => onDone({ message: SERVICE_CREATED, draft: true }),
    });
  };

  const serverError = mutation.error;
  const serverFieldErrors: Record<string, string> = {};
  if (serverError) {
    for (const field of [
      'dayOfWeek',
      'startTime',
      'endTime',
      'nameEs',
      'nameEn',
      'descriptionEs',
      'descriptionEn',
      'placeEs',
      'placeEn',
    ]) {
      const message = detailMessage(serverError, field);
      if (message) {
        serverFieldErrors[field] = message;
      }
    }
  }
  const displayErrors = { ...errors, ...serverFieldErrors };
  const fieldError = (field: string) => displayErrors[field];
  const showGeneralError = serverError && Object.keys(serverFieldErrors).length === 0;

  return (
    <Dialog title={isEdit ? 'Editar servicio' : 'Agregar servicio'} onClose={onCancel}>
      <form onSubmit={handleSubmit} noValidate className="space-y-4">
        {showGeneralError && <Notice variant="error">{generalErrorMessage(serverError)}</Notice>}

        <Select
          label="Día"
          required
          options={DAY_OPTIONS}
          value={fields.dayOfWeek}
          error={fieldError('dayOfWeek')}
          onChange={(event) => update({ dayOfWeek: event.target.value })}
        />
        <Field
          label="Hora de inicio"
          type="time"
          required
          value={fields.startTime}
          error={fieldError('startTime')}
          onChange={(event) => update({ startTime: event.target.value })}
        />
        <Field
          label="Hora de fin (opcional)"
          type="time"
          value={fields.endTime}
          help="Déjala vacía si no hay una hora de fin; si la pones, debe ser posterior a la de inicio."
          error={fieldError('endTime')}
          onChange={(event) => update({ endTime: event.target.value })}
        />

        <Tabs
          tabs={LANGUAGE_TABS}
          active={langTab}
          onChange={(id) => setLangTab(id as 'es' | 'en')}
          label="Idioma del contenido"
        />

        {langTab === 'es' ? (
          <>
            <Field
              label="Nombre (Español)"
              required
              value={fields.nameEs}
              error={fieldError('nameEs')}
              onChange={(event) => update({ nameEs: event.target.value })}
            />
            <TextArea
              label="Descripción (Español)"
              value={fields.descriptionEs}
              error={fieldError('descriptionEs')}
              onChange={(event) => update({ descriptionEs: event.target.value })}
            />
            <Field
              label="Lugar (Español)"
              required
              value={fields.placeEs}
              error={fieldError('placeEs')}
              onChange={(event) => update({ placeEs: event.target.value })}
            />
          </>
        ) : (
          <>
            <p className="text-sm text-slate-600">{ENGLISH_HELP}</p>
            <Field
              label="Nombre (English)"
              value={fields.nameEn}
              error={fieldError('nameEn')}
              onChange={(event) => update({ nameEn: event.target.value })}
            />
            <TextArea
              label="Descripción (English)"
              value={fields.descriptionEn}
              error={fieldError('descriptionEn')}
              onChange={(event) => update({ descriptionEn: event.target.value })}
            />
            <Field
              label="Lugar (English)"
              value={fields.placeEn}
              error={fieldError('placeEn')}
              onChange={(event) => update({ placeEn: event.target.value })}
            />
          </>
        )}

        <div className="flex flex-wrap justify-end gap-3">
          <Button type="button" variant="secondary" onClick={onCancel} disabled={busy}>
            Cancelar
          </Button>
          <Button type="submit" loading={busy}>
            {isEdit ? 'Guardar cambios' : 'Agregar servicio'}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
