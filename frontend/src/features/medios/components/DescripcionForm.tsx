import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { MedioAdmin, MedioPatch } from '../../../api/medios';
import { mediaUrl } from '../../../api/client';
import { Button } from '../../../components/Button';
import { Dialog } from '../../../components/Dialog';
import { Notice } from '../../../components/Notice';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { TextArea } from '../../informacion/components/TextArea';
import { formatLongDate } from '../../publico/eventos/formatoFechas';
import { episodeDay } from '../../publico/medios/fecha';
import { detailMessage, generalErrorMessage } from '../errors';
import { useActualizarDescripcion } from '../hooks/useActualizarDescripcion';
import {
  CANCEL,
  DESCRIPTION_CONTEXT,
  DESCRIPTION_FORM_TITLE,
  DESCRIPTION_HELP,
  DESCRIPTION_SAVED,
  DESCRIPTION_SAVED_PUBLISHED,
  ENGLISH_HELP,
  SAVE,
  SAVING,
  TAB_EN,
  TAB_ES,
  counter,
  descriptionLimitMessage,
} from '../messages';
import {
  descripcionFieldsFrom,
  descripcionFieldsSchema,
  issuesToErrors,
  overDescriptionLimit,
  type DescripcionFields,
} from '../validation';

const LANGUAGE_TABS: TabItem[] = [
  { id: 'es', label: TAB_ES },
  { id: 'en', label: TAB_EN },
];

/** Orden de foco del primer campo inválido (`id` del control y su pestaña). */
const FOCUS_ORDER: Array<{ field: keyof DescripcionFields; id: string; tab: 'es' | 'en' }> = [
  { field: 'descriptionEs', id: 'descripcion-es', tab: 'es' },
  { field: 'descriptionEn', id: 'descripcion-en', tab: 'en' },
];

/** Campos del servidor que se pintan junto a su control (para los `details`). */
const SERVER_FIELDS = ['descriptionEs', 'descriptionEn'] as const;

export interface DescripcionFormProps {
  /** Ficha cuyo texto se ajusta (nunca se edita ningún dato del feed, FR-019). */
  item: MedioAdmin;
  onDone: (feedback: { message: string }) => void;
  onCancel: () => void;
}

/**
 * Dialog «Editar descripción» del panel (ux.md §4.5, US3/FR-020/FR-023/FR-025):
 * la **única** edición de contenido que permite F9 —el nombre, la fecha, la
 * carátula y los enlaces vienen del feed y **no** se tocan (FR-019)—.
 *
 * · **Miniatura de contexto** arriba (carátula o marcador neutro + nombre +
 *   fecha, en solo lectura): los nombres y fechas contiguas se parecen y esto
 *   evita ajustar la ficha equivocada (⟲ D-23).
 * · **Pestañas** Español (obligatorio **si el campo se rellena**, FR-023) /
 *   English (opcional) con la ayuda de F3; el valor **vigente va precargado**
 *   —también el que generó el recorte automático del feed— y el texto del equipo
 *   manda y no se regenera (FR-015/SC-007).
 * · `TextArea` grande con **contador vivo** `{n}/{L}`; por encima del límite el
 *   contador pasa a error y «Guardar» queda deshabilitado con el aviso que dice
 *   cuánto recortar (nunca se guarda a medias, FR-023).
 * · **La visibilidad no se cambia aquí** (⟲ D-24): se cambia con
 *   «Ocultar»/«Mostrar» en la fila —«editar el texto» y «decir si se ve» son
 *   decisiones separadas—; el `PATCH` lleva **solo** las descripciones.
 * · **Guardar**: `Button` loading («Guardando…») y el modal **no se cierra hasta
 *   terminar**; `Esc` tampoco cierra mientras la operación está en vuelo (⟲ D-18:
 *   nada se pierde a medias). Al terminar, cierre + aviso de la página: «Cambios
 *   guardados: ya visibles en la sección.» si la ficha está visible, «Cambios
 *   guardados.» si está oculta (FR-015: se publica al guardar).
 *
 * Los errores van **junto al campo** (los `details` del servidor se traducen por
 * campo) con resumen arriba y el foco al primer campo inválido, cambiando de
 * pestaña si hace falta (FR-023).
 */
export function DescripcionForm({ item, onDone, onCancel }: DescripcionFormProps) {
  const actualizar = useActualizarDescripcion();

  const [langTab, setLangTab] = useState<'es' | 'en'>('es');
  const [fields, setFields] = useState<DescripcionFields>(() =>
    descripcionFieldsFrom(item.descriptionEs, item.descriptionEn),
  );
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({});
  const [pendingFocus, setPendingFocus] = useState<{ id: string; nonce: number }>();

  // `Esc`/clic fuera no cierran mientras se guarda: el modal no se cierra hasta
  // terminar (⟲ D-18). El ref mantiene estable el `onClose` del `Dialog` para no
  // re-ejecutar su efecto (y no robar el foco) en cada pulsación de tecla.
  const savingRef = useRef(false);
  useEffect(() => {
    savingRef.current = actualizar.isPending;
  }, [actualizar.isPending]);
  const handleClose = useCallback(() => {
    if (!savingRef.current) {
      onCancel();
    }
  }, [onCancel]);

  useEffect(() => {
    if (pendingFocus) {
      document.getElementById(pendingFocus.id)?.focus();
    }
  }, [pendingFocus]);

  const serverErrors = useMemo(() => {
    const mapped: Record<string, string> = {};
    for (const field of SERVER_FIELDS) {
      const message = detailMessage(actualizar.error, field);
      if (message) {
        mapped[field] = message;
      }
    }
    return mapped;
  }, [actualizar.error]);

  const errors = { ...clientErrors, ...serverErrors };
  const overLimit =
    overDescriptionLimit(fields.descriptionEs) || overDescriptionLimit(fields.descriptionEn);

  const update = (patch: Partial<DescripcionFields>) =>
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
    const parsed = descripcionFieldsSchema.safeParse(fields);
    const nextErrors = parsed.success ? {} : issuesToErrors(parsed.error);

    if (Object.keys(nextErrors).length > 0) {
      setClientErrors(nextErrors);
      focusFirstInvalid(nextErrors);
      return;
    }

    setClientErrors({});
    // `""` (o solo espacios) significa «sin descripción» y así viaja al servidor:
    // la ficha omite la línea (decisión del 2026-10-18). La visibilidad NO se
    // toca (⟲ D-24).
    const input: MedioPatch = {
      descriptionEs: fields.descriptionEs.trim(),
      descriptionEn: fields.descriptionEn.trim(),
    };
    actualizar.mutate(
      { id: item.id, input },
      {
        onSuccess: () =>
          onDone({
            message:
              item.visibility === 'visible' ? DESCRIPTION_SAVED_PUBLISHED : DESCRIPTION_SAVED,
          }),
      },
    );
  };

  const summaryMessages = Object.values(errors).filter((message): message is string =>
    Boolean(message),
  );
  const showGeneralError = actualizar.error && Object.keys(serverErrors).length === 0;
  const current = langTab === 'es' ? fields.descriptionEs : fields.descriptionEn;
  const count = current.length;
  const overCurrent = overDescriptionLimit(current);

  return (
    <Dialog title={DESCRIPTION_FORM_TITLE} size="wide" onClose={handleClose}>
      <form onSubmit={handleSubmit} noValidate className="space-y-4">
        {/* Contexto: qué episodio se está ajustando (⟲ D-23). */}
        <div className="flex items-center gap-3 rounded border border-slate-200 p-3">
          {mediaUrl(item.coverUrl) ? (
            <img
              src={mediaUrl(item.coverUrl)}
              alt=""
              className="h-12 w-12 shrink-0 rounded border border-slate-200 object-cover"
            />
          ) : (
            <span
              data-testid="contexto-sin-caratula"
              aria-hidden="true"
              className="flex h-12 w-12 shrink-0 items-center justify-center rounded border border-slate-200 bg-slate-100 text-slate-500"
            >
              ♪
            </span>
          )}
          <div className="min-w-0">
            <p className="text-sm text-slate-600">{DESCRIPTION_CONTEXT}</p>
            <p className="font-medium text-slate-900">{item.title}</p>
            <p className="text-xs text-slate-600">
              {formatLongDate(episodeDay(item.publishedAt), 'es')}
            </p>
          </div>
        </div>

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

        {showGeneralError && (
          <Notice variant="error">{generalErrorMessage(actualizar.error)}</Notice>
        )}

        <Tabs
          tabs={LANGUAGE_TABS}
          active={langTab}
          onChange={(id) => setLangTab(id as 'es' | 'en')}
          label="Idioma del contenido"
        />

        {langTab === 'es' ? (
          <>
            <TextArea
              id="descripcion-es"
              label={TAB_ES}
              value={fields.descriptionEs}
              help={DESCRIPTION_HELP}
              error={errors.descriptionEs}
              onChange={(event) => update({ descriptionEs: event.target.value })}
            />
            <p
              aria-live="polite"
              className={
                overCurrent ? 'text-sm font-medium text-red-700' : 'text-sm text-slate-600'
              }
            >
              {overCurrent ? descriptionLimitMessage(count) : counter(count)}
            </p>
          </>
        ) : (
          <>
            <p className="text-sm text-slate-600">{ENGLISH_HELP}</p>
            <TextArea
              id="descripcion-en"
              label={TAB_EN}
              value={fields.descriptionEn}
              help={DESCRIPTION_HELP}
              error={errors.descriptionEn}
              onChange={(event) => update({ descriptionEn: event.target.value })}
            />
            <p
              aria-live="polite"
              className={
                overCurrent ? 'text-sm font-medium text-red-700' : 'text-sm text-slate-600'
              }
            >
              {overCurrent ? descriptionLimitMessage(count) : counter(count)}
            </p>
          </>
        )}

        <div className="flex flex-wrap justify-end gap-3">
          <Button
            type="button"
            variant="secondary"
            onClick={onCancel}
            disabled={actualizar.isPending}
          >
            {CANCEL}
          </Button>
          <Button
            type="submit"
            loading={actualizar.isPending}
            disabled={overLimit}
            aria-disabled={overLimit || undefined}
          >
            {actualizar.isPending ? SAVING : SAVE}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
