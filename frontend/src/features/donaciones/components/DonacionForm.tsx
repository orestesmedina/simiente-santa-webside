import { useEffect, useMemo, useState } from 'react';
import type {
  DonacionAdmin,
  DonacionInput,
  DonacionPatch,
  MetodoKind,
  Moneda,
} from '../../../api/ofrendas';
import { Button } from '../../../components/Button';
import { Dialog } from '../../../components/Dialog';
import { Field } from '../../../components/Field';
import { Notice } from '../../../components/Notice';
import { Select } from '../../../components/Select';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { useCreateDonacion } from '../hooks/useCreateDonacion';
import { useUpdateDonacion } from '../hooks/useUpdateDonacion';
import { detailMessage, generalErrorMessage, isConflict } from '../errors';
import {
  ADD_ACCOUNT,
  ADD_SINPE,
  CANCEL,
  CREATED_ACCOUNT,
  CREATED_SINPE,
  CURRENCY_HELP,
  DRAFT_NOTICE_ACCOUNT,
  DRAFT_NOTICE_SINPE,
  DUPLICATE,
  EDIT_ACTION,
  ENGLISH_HELP,
  FIELD_CURRENCY,
  FIELD_HOLDER,
  FIELD_INSTITUTION,
  FIELD_NUMBER_IBAN,
  FIELD_NUMBER_SINPE,
  FIELD_ORDER,
  FIELD_TYPE,
  HOLDER_HELP,
  INSTITUTION_HELP,
  NUMBER_IBAN_HELP,
  NUMBER_SINPE_HELP,
  ORDER_HELP,
  SAVE,
  SAVED,
  SAVED_PUBLISHED,
  SUMMARY_TITLE,
  TAB_EN,
  TAB_ES,
  TYPE_FIXED_HELP,
  TYPE_IBAN,
  TYPE_SINPE,
} from '../messages';
import {
  CURRENCY_OPTIONS,
  EMPTY_DONACION_FIELDS,
  FOCUS_ORDER,
  crossFieldErrors,
  donacionSchema,
  issuesToErrors,
  nullable,
  type DonacionFields,
} from '../validation';

const LANGUAGE_TABS: TabItem[] = [
  { id: 'es', label: TAB_ES },
  { id: 'en', label: TAB_EN },
];

/** Campos del servidor que se pintan junto a su control (para los `details`). */
const SERVER_FIELDS = [
  'institution',
  'holderEs',
  'holderEn',
  'number',
  'currency',
  'sortOrder',
] as const;

export interface DonacionFormProps {
  /** Tipo fijo del método (fijado por el bloque de origen al crear, D-9). */
  kind: MetodoKind;
  /** Método a editar; sin él, alta. */
  donacion?: DonacionAdmin;
  onDone: (feedback: { message: string; draftNotice?: string }) => void;
  onCancel: () => void;
}

/** Moneda válida del formulario o `null` (sin elegir). */
function currencyValue(fields: DonacionFields): Moneda | null {
  return fields.currency === 'CRC' || fields.currency === 'USD' ? fields.currency : null;
}

function fieldsFromDonacion(donacion?: DonacionAdmin): DonacionFields {
  return {
    ...EMPTY_DONACION_FIELDS,
    institution: donacion?.institution ?? '',
    holderEs: donacion?.holderEs ?? '',
    holderEn: donacion?.holderEn ?? '',
    number: donacion?.number ?? '',
    currency: donacion?.currency ?? '',
    sortOrder: String(donacion?.sortOrder ?? 0),
  };
}

/**
 * Alta/edición de un método de donación (ux.md §4.4, US2/FR-010/FR-014/FR-015/
 * FR-016) en `Dialog`: pestañas **Español** / **English (opcional)** que
 * contienen **solo el titular** (el `institution` es texto **no traducible** y el
 * `number`/`kind`/`currency` nunca aparecen en las pestañas — SC-006), el **tipo**
 * como etiqueta fija (D-9), la **Moneda** obligatoria y **solo en IBAN** (Q2), el
 * **Orden** numérico y la validación de número espejo de la del servidor (que es
 * la autoridad).
 *
 * Los errores van junto al campo (los `details` del servidor se traducen por
 * campo) con resumen arriba y foco al primero —cambiando de pestaña si hace
 * falta— (FR-014); un **duplicado** (`409`) se avisa sin perder lo escrito y sin
 * guardar; un error del sistema tampoco pierde lo escrito.
 */
export function DonacionForm({ kind, donacion, onDone, onCancel }: DonacionFormProps) {
  const createDonacion = useCreateDonacion();
  const updateDonacion = useUpdateDonacion();
  const isEdit = Boolean(donacion);

  const [langTab, setLangTab] = useState<'es' | 'en'>('es');
  const [fields, setFields] = useState<DonacionFields>(() => fieldsFromDonacion(donacion));
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({});
  const [pendingFocus, setPendingFocus] = useState<{ id: string; nonce: number }>();

  const mutation = isEdit ? updateDonacion : createDonacion;
  const busy = createDonacion.isPending || updateDonacion.isPending;

  const typeLabel = kind === 'iban' ? TYPE_IBAN : TYPE_SINPE;
  const addLabel = kind === 'iban' ? ADD_ACCOUNT : ADD_SINPE;

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

  const update = (patch: Partial<DonacionFields>) =>
    setFields((current) => ({ ...current, ...patch }));

  const focusFirstInvalid = (fieldErrors: Record<string, string>) => {
    const target = FOCUS_ORDER.find((entry) => fieldErrors[entry.field]);
    if (target) {
      setLangTab(target.tab);
      setPendingFocus({ id: target.id, nonce: Date.now() });
    }
  };

  const handleSubmit = (formEvent: React.FormEvent<HTMLFormElement>) => {
    formEvent.preventDefault();
    const parsed = donacionSchema(kind).safeParse(fields);
    const nextErrors: Record<string, string> = parsed.success ? {} : issuesToErrors(parsed.error);
    Object.assign(nextErrors, crossFieldErrors(fields));

    if (Object.keys(nextErrors).length > 0) {
      setClientErrors(nextErrors);
      focusFirstInvalid(nextErrors);
      return;
    }

    setClientErrors({});
    const order = Number.parseInt(fields.sortOrder.trim(), 10);
    const sortOrder = Number.isNaN(order) ? 0 : order;

    if (donacion) {
      const wasPublished = donacion.publicationState === 'published';
      const input: DonacionPatch = {
        institution: fields.institution.trim(),
        number: fields.number.trim(),
        holderEs: nullable(fields.holderEs),
        holderEn: nullable(fields.holderEn),
        sortOrder,
      };
      if (kind === 'iban') {
        input.currency = currencyValue(fields);
      }
      updateDonacion.mutate(
        { id: donacion.id, input },
        { onSuccess: () => onDone({ message: wasPublished ? SAVED_PUBLISHED : SAVED }) },
      );
      return;
    }

    const input: DonacionInput = {
      kind,
      institution: fields.institution.trim(),
      number: fields.number.trim(),
      holderEs: nullable(fields.holderEs),
      holderEn: nullable(fields.holderEn),
      sortOrder,
    };
    if (kind === 'iban') {
      input.currency = currencyValue(fields) ?? undefined;
    }
    createDonacion.mutate(input, {
      onSuccess: () =>
        onDone({
          message: kind === 'iban' ? CREATED_ACCOUNT : CREATED_SINPE,
          draftNotice: kind === 'iban' ? DRAFT_NOTICE_ACCOUNT : DRAFT_NOTICE_SINPE,
        }),
    });
  };

  const summaryMessages = Object.values(errors).filter((message): message is string =>
    Boolean(message),
  );
  const conflict = isConflict(mutation.error);
  const showGeneralError = mutation.error && Object.keys(serverErrors).length === 0 && !conflict;

  return (
    <Dialog
      title={isEdit ? `${EDIT_ACTION} ${typeLabel}` : addLabel}
      size="wide"
      onClose={onCancel}
    >
      <form onSubmit={handleSubmit} noValidate className="space-y-4">
        {summaryMessages.length > 0 && (
          <Notice variant="error">
            <p className="font-medium">{SUMMARY_TITLE}</p>
            <ul className="mt-1 list-disc pl-5">
              {summaryMessages.map((message) => (
                <li key={message}>{message}</li>
              ))}
            </ul>
          </Notice>
        )}

        {conflict && <Notice variant="error">{DUPLICATE}</Notice>}
        {showGeneralError && <Notice variant="error">{generalErrorMessage(mutation.error)}</Notice>}

        <div className="flex flex-col gap-1">
          <span className="font-medium text-slate-900">{FIELD_TYPE}</span>
          <p className="text-slate-700">{typeLabel}</p>
          <p className="text-sm text-slate-600">{TYPE_FIXED_HELP}</p>
        </div>

        <Tabs
          tabs={LANGUAGE_TABS}
          active={langTab}
          onChange={(id) => setLangTab(id as 'es' | 'en')}
          label="Idioma del contenido"
        />

        {langTab === 'es' ? (
          <Field
            id="donacion-holder-es"
            label={`${FIELD_HOLDER} (Español)`}
            value={fields.holderEs}
            help={HOLDER_HELP}
            error={errors.holderEs}
            onChange={(event) => update({ holderEs: event.target.value })}
          />
        ) : (
          <>
            <p className="text-sm text-slate-600">{ENGLISH_HELP}</p>
            <Field
              id="donacion-holder-en"
              label={`${FIELD_HOLDER} (English)`}
              value={fields.holderEn}
              error={errors.holderEn}
              onChange={(event) => update({ holderEn: event.target.value })}
            />
          </>
        )}

        <Field
          id="donacion-institution"
          label={FIELD_INSTITUTION}
          required
          value={fields.institution}
          help={INSTITUTION_HELP}
          error={errors.institution}
          onChange={(event) => update({ institution: event.target.value })}
        />

        <Field
          id="donacion-number"
          label={kind === 'iban' ? FIELD_NUMBER_IBAN : FIELD_NUMBER_SINPE}
          required
          value={fields.number}
          help={kind === 'iban' ? NUMBER_IBAN_HELP : NUMBER_SINPE_HELP}
          error={errors.number}
          onChange={(event) => update({ number: event.target.value })}
        />

        {kind === 'iban' && (
          <Select
            id="donacion-currency"
            label={FIELD_CURRENCY}
            required
            options={CURRENCY_OPTIONS}
            value={fields.currency}
            help={CURRENCY_HELP}
            error={errors.currency}
            onChange={(event) =>
              update({ currency: event.target.value as DonacionFields['currency'] })
            }
          />
        )}

        <Field
          id="donacion-sort-order"
          label={FIELD_ORDER}
          type="number"
          step={1}
          value={fields.sortOrder}
          help={ORDER_HELP}
          error={errors.sortOrder}
          onChange={(event) => update({ sortOrder: event.target.value })}
        />

        <div className="flex flex-wrap justify-end gap-3">
          <Button type="button" variant="secondary" onClick={onCancel} disabled={busy}>
            {CANCEL}
          </Button>
          <Button type="submit" loading={busy}>
            {isEdit ? SAVE : addLabel}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
