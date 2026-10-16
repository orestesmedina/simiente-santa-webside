import { useEffect, useMemo, useState } from 'react';
import { z } from 'zod';
import type {
  DonacionTextoAdmin,
  DonacionTextoInput,
  DonacionTextKey,
} from '../../../api/ofrendas';
import { Button } from '../../../components/Button';
import { Dialog } from '../../../components/Dialog';
import { Notice } from '../../../components/Notice';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { TextArea } from '../../informacion/components/TextArea';
import { detailMessage, generalErrorMessage } from '../errors';
import { usePutDonacionTexto } from '../hooks/usePutDonacionTexto';
import {
  CANCEL,
  ENGLISH_HELP,
  REQUIRED_TEXT,
  SAVE,
  SAVED,
  SAVED_PUBLISHED,
  SUMMARY_TITLE,
  TAB_EN,
  TAB_ES,
  TEXT_FORM_TITLE_INTRO,
  TEXT_FORM_TITLE_USAGE,
  TEXT_HELP,
  TEXT_MAX,
  counter,
  spanishRequiredFor,
  textLimitMessage,
  textTooLong,
} from '../messages';

const LANGUAGE_TABS: TabItem[] = [
  { id: 'es', label: TAB_ES },
  { id: 'en', label: TAB_EN },
];

interface TextoFields {
  textEs: string;
  textEn: string;
}

const textoFieldsSchema = z.object({
  textEs: z.string().trim().min(1, REQUIRED_TEXT).max(TEXT_MAX, textTooLong(TEXT_MAX)),
  textEn: z.string().trim().max(TEXT_MAX, textTooLong(TEXT_MAX)).optional().default(''),
});

/** Orden de foco del primer campo inválido (`id` del control y su pestaña). */
const FOCUS_ORDER: Array<{ field: keyof TextoFields; id: string; tab: 'es' | 'en' }> = [
  { field: 'textEs', id: 'texto-es', tab: 'es' },
  { field: 'textEn', id: 'texto-en', tab: 'en' },
];

/** Campos del servidor que se pintan junto a su control (para los `details`). */
const SERVER_FIELDS = ['textEs', 'textEn'] as const;

export interface TextoFormProps {
  /** Clave del texto: `intro` (introductorio) o `usage` (de uso). */
  textKey: DonacionTextKey;
  /** Texto guardado; `null` si nunca se escribió (⟲ D-8). */
  texto: DonacionTextoAdmin | null;
  onDone: (feedback: { message: string }) => void;
  onCancel: () => void;
}

function issuesToErrors(error: z.ZodError): Record<string, string> {
  const mapped: Record<string, string> = {};
  for (const issue of error.issues) {
    const key = issue.path.join('.');
    if (!(key in mapped)) {
      mapped[key] = issue.message;
    }
  }
  return mapped;
}

/**
 * Edición del **texto introductorio** o del **texto de uso** de la sección
 * (ux.md §4.5, US2/US3/FR-021): un solo componente parametrizado por `textKey`,
 * en `Dialog`, con pestañas **Español** (obligatorio) / **English (opcional)**, un
 * `TextArea` grande con contador vivo (`{n}/2000`) que conserva saltos, tildes,
 * `ñ`, `¿¡` y emojis, y ayuda de escritura.
 *
 * **El estado del texto no se cambia aquí**: se cambia con `PublishControls` en
 * su tarjeta (Q10) — «editar» y «publicar» son dos decisiones separadas, como
 * para un método (⟲ D-11). El `PUT` crea o reemplaza el singleton y **preserva**
 * el estado actual; no hay operación de borrado (P7-17).
 */
export function TextoForm({ textKey, texto, onDone, onCancel }: TextoFormProps) {
  const putTexto = usePutDonacionTexto();
  const [langTab, setLangTab] = useState<'es' | 'en'>('es');
  const [fields, setFields] = useState<TextoFields>({
    textEs: texto?.textEs ?? '',
    textEn: texto?.textEn ?? '',
  });
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({});
  const [pendingFocus, setPendingFocus] = useState<{ id: string; nonce: number }>();

  const title = textKey === 'intro' ? TEXT_FORM_TITLE_INTRO : TEXT_FORM_TITLE_USAGE;
  const count = fields.textEs.length;
  const overLimit = count > TEXT_MAX;

  useEffect(() => {
    if (pendingFocus) {
      document.getElementById(pendingFocus.id)?.focus();
    }
  }, [pendingFocus]);

  const serverErrors = useMemo(() => {
    const mapped: Record<string, string> = {};
    const error = putTexto.error;
    if (error) {
      for (const field of SERVER_FIELDS) {
        const message = detailMessage(error, field);
        if (message) {
          mapped[field] = message;
        }
      }
    }
    return mapped;
  }, [putTexto.error]);

  const errors = { ...clientErrors, ...serverErrors };

  const update = (patch: Partial<TextoFields>) =>
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
    const parsed = textoFieldsSchema.safeParse(fields);
    const nextErrors: Record<string, string> = parsed.success ? {} : issuesToErrors(parsed.error);
    if (fields.textEn.trim() !== '' && fields.textEs.trim() === '') {
      nextErrors.textEn = spanishRequiredFor('el texto');
    }

    if (Object.keys(nextErrors).length > 0) {
      setClientErrors(nextErrors);
      focusFirstInvalid(nextErrors);
      return;
    }

    setClientErrors({});
    const input: DonacionTextoInput = {
      textEs: fields.textEs.trim(),
      textEn: fields.textEn.trim() === '' ? null : fields.textEn.trim(),
      // El estado NO se cambia en el formulario: se conserva el actual (Q10).
      publicationState: texto?.publicationState ?? 'draft',
    };
    putTexto.mutate(
      { textKey, input },
      {
        onSuccess: () =>
          onDone({ message: texto?.publicationState === 'published' ? SAVED_PUBLISHED : SAVED }),
      },
    );
  };

  const summaryMessages = Object.values(errors).filter((message): message is string =>
    Boolean(message),
  );
  const showGeneralError = putTexto.error && Object.keys(serverErrors).length === 0;

  return (
    <Dialog title={title} size="wide" onClose={onCancel}>
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

        {showGeneralError && <Notice variant="error">{generalErrorMessage(putTexto.error)}</Notice>}

        <Tabs
          tabs={LANGUAGE_TABS}
          active={langTab}
          onChange={(id) => setLangTab(id as 'es' | 'en')}
          label="Idioma del contenido"
        />

        {langTab === 'es' ? (
          <>
            <TextArea
              id="texto-es"
              label={`Texto (${TAB_ES})`}
              required
              value={fields.textEs}
              help={TEXT_HELP}
              error={errors.textEs}
              onChange={(event) => update({ textEs: event.target.value })}
            />
            <p
              className={overLimit ? 'text-sm font-medium text-red-700' : 'text-sm text-slate-600'}
            >
              {overLimit ? textLimitMessage(count) : counter(count, TEXT_MAX)}
            </p>
          </>
        ) : (
          <>
            <p className="text-sm text-slate-600">{ENGLISH_HELP}</p>
            <TextArea
              id="texto-en"
              label={`Texto (English)`}
              value={fields.textEn}
              help={TEXT_HELP}
              error={errors.textEn}
              onChange={(event) => update({ textEn: event.target.value })}
            />
          </>
        )}

        <div className="flex flex-wrap justify-end gap-3">
          <Button
            type="button"
            variant="secondary"
            onClick={onCancel}
            disabled={putTexto.isPending}
          >
            {CANCEL}
          </Button>
          <Button type="submit" loading={putTexto.isPending} disabled={overLimit}>
            {SAVE}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
