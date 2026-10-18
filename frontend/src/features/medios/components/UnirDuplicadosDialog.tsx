import { useState } from 'react';
import type { MedioAdmin } from '../../../api/medios';
import { mediaUrl } from '../../../api/client';
import { Button } from '../../../components/Button';
import { ConfirmDialog } from '../../../components/ConfirmDialog';
import { Dialog } from '../../../components/Dialog';
import { Notice } from '../../../components/Notice';
import { StatusPill } from '../../../components/StatusPill';
import { formatLongDate } from '../../publico/eventos/formatoFechas';
import { episodeDay } from '../../publico/medios/fecha';
import { useResolverMarca } from '../hooks/useResolverMarca';
import { useUnirDuplicados } from '../hooks/useUnirDuplicados';
import { parParaUnion } from '../marcas';
import {
  CANCEL,
  MERGE_ACTION,
  MERGE_CONFIRM,
  MERGE_CONFIRM_ACTION,
  MERGE_CONFIRM_TITLE,
  MERGE_DIALOG_TITLE,
  MERGE_PREVIEW_NOTE,
  MERGE_PREVIEW_TITLE,
  MERGED_NOTICE,
  NOT_SAME_ACTION,
  NOT_SAME_NOTICE,
  STATE_HIDDEN,
  STATE_VISIBLE,
  SYSTEM_ERROR,
} from '../messages';
import { FuentesList } from './FuentesList';
import { OrigenBadge } from './OrigenBadge';

export interface UnirDuplicadosDialogProps {
  /** Ficha cuya fila abrió el diálogo (una de las dos del par). */
  item: MedioAdmin;
  /** La otra ficha del par marcado «posible duplicado» (FR-014). */
  pareja: MedioAdmin;
  onDone: (feedback: { message: string }) => void;
  onCancel: () => void;
}

/** Columna de una candidata: carátula, nombre, fecha, origen, enlaces y estado. */
function Candidata({ etiqueta, item }: { etiqueta: string; item: MedioAdmin }) {
  const src = mediaUrl(item.coverUrl);

  return (
    <article className="flex flex-col gap-2 rounded border border-slate-200 p-3">
      <p className="text-xs font-semibold uppercase tracking-wide text-slate-600">{etiqueta}</p>
      {src ? (
        <img src={src} alt="" className="aspect-video w-full rounded object-cover" />
      ) : (
        <span
          data-testid="candidata-sin-caratula"
          aria-hidden="true"
          className="flex aspect-video w-full items-center justify-center rounded bg-slate-100 text-slate-500"
        >
          ♪
        </span>
      )}
      <p className="font-medium text-slate-900">{item.title}</p>
      <p className="text-xs text-slate-600">{formatLongDate(episodeDay(item.publishedAt), 'es')}</p>
      <div>
        <OrigenBadge origin={item.origin} />
      </div>
      <FuentesList sources={item.sources} />
      <div>
        <StatusPill
          value={item.visibility === 'visible' ? 'active' : 'inactive'}
          label={item.visibility === 'visible' ? STATE_VISIBLE : STATE_HIDDEN}
        />
      </div>
    </article>
  );
}

/**
 * Dialog «Revisar duplicado» (ux.md §4.6, US3 esc. 1, FR-013/FR-014): **decisión
 * antes que sorpresa**.
 *
 * 1. Las **dos candidatas lado a lado**, cada una con carátula, nombre completo,
 *    fecha, enlaces, origen y su `StatusPill` de visibilidad —así se ve en
 *    pantalla la regla de FR-014: la primera importada está visible y la
 *    adicional, oculta—.
 * 2. La zona **«Así quedaría la ficha si las unes»**: carátula, **nombre de la
 *    ficha de YouTube** con la frase declaratoria «Al unir, el nombre publicado
 *    es el de YouTube.» (⟲ D-25/precisión 5), fecha y **los dos enlaces juntos**.
 * 3. Tres acciones: **«Unir»** → `ConfirmDialog` **danger** «… Esto no se puede
 *    deshacer.» («Unir definitivamente» / «Cancelar», ⟲ D-26: es la única
 *    operación irreversible del módulo); **«No son el mismo sermón»** (sin
 *    confirmación extra, ⟲ D-27: no destruye nada y es reversible) → resuelve la
 *    marca `possible_duplicate` con `visibility: visible` en **las dos** fichas
 *    del par (el servidor cierra el eje de duplicados del par en **una** sola
 *    operación, P9-19, y es idempotente: llamarla dos veces sería trabajo de
 *    más); y **«Cancelar»**, que cierra sin cambios y devuelve el foco a la fila
 *    que lo abrió.
 *
 * La unión manda la ficha del **video** como base (`parParaUnion`): es la que da
 * nombre, carátula y fecha al resultado y la que adopta el enlace que le faltaba.
 * **Sin un par abierto no existe ni el botón ni este diálogo** (⟲ D-28: unir solo
 * opera sobre pares que el sistema marcó).
 */
export function UnirDuplicadosDialog({
  item,
  pareja,
  onDone,
  onCancel,
}: UnirDuplicadosDialogProps) {
  const unir = useUnirDuplicados();
  const resolver = useResolverMarca();
  const [confirmando, setConfirmando] = useState(false);
  const [error, setError] = useState<string>();

  const { base, otra } = parParaUnion(item, pareja);
  const enlacesDeLaUnion = [...base.sources, ...otra.sources];
  const srcBase = mediaUrl(base.coverUrl);
  const busy = unir.isPending || resolver.isPending;

  const confirmarUnion = () => {
    unir.mutate(
      { id: base.id, otherId: otra.id },
      {
        onSuccess: () => onDone({ message: MERGED_NOTICE }),
        onError: () => {
          setConfirmando(false);
          setError(SYSTEM_ERROR);
        },
      },
    );
  };

  const noSonElMismo = () => {
    resolver.mutate(
      { id: item.id, input: { mark: 'possible_duplicate', visibility: 'visible' } },
      {
        onSuccess: () => onDone({ message: NOT_SAME_NOTICE }),
        onError: () => setError(SYSTEM_ERROR),
      },
    );
  };

  return (
    <>
      <Dialog title={MERGE_DIALOG_TITLE} size="wide" onClose={onCancel}>
        {error && <Notice variant="error">{error}</Notice>}

        <div className="mt-2 grid gap-3 sm:grid-cols-2">
          <Candidata etiqueta="A" item={item} />
          <Candidata etiqueta="B" item={pareja} />
        </div>

        <section aria-labelledby="medios-union-previa" className="mt-4 space-y-2">
          <h3 id="medios-union-previa" className="font-semibold text-slate-900">
            {MERGE_PREVIEW_TITLE}
          </h3>
          <p className="text-sm text-slate-700">{MERGE_PREVIEW_NOTE}</p>
          <article className="flex flex-col gap-2 rounded border border-slate-300 bg-slate-50 p-3">
            {srcBase ? (
              <img src={srcBase} alt="" className="aspect-video w-full rounded object-cover" />
            ) : (
              <span
                data-testid="previa-sin-caratula"
                aria-hidden="true"
                className="flex aspect-video w-full items-center justify-center rounded bg-slate-100 text-slate-500"
              >
                ♪
              </span>
            )}
            <p className="font-medium text-slate-900">{base.title}</p>
            <p className="text-xs text-slate-600">
              {formatLongDate(episodeDay(base.publishedAt), 'es')}
            </p>
            <FuentesList sources={enlacesDeLaUnion} />
          </article>
        </section>

        <div className="mt-4 flex flex-wrap justify-end gap-3">
          <Button variant="secondary" onClick={onCancel} disabled={busy}>
            {CANCEL}
          </Button>
          <Button variant="secondary" onClick={noSonElMismo} disabled={busy}>
            {NOT_SAME_ACTION}
          </Button>
          <Button onClick={() => setConfirmando(true)} disabled={busy}>
            {MERGE_ACTION}
          </Button>
        </div>
      </Dialog>

      {confirmando && (
        <ConfirmDialog
          title={MERGE_CONFIRM_TITLE}
          description={MERGE_CONFIRM}
          confirmText={MERGE_CONFIRM_ACTION}
          danger
          loading={unir.isPending}
          onConfirm={confirmarUnion}
          onClose={() => setConfirmando(false)}
        />
      )}
    </>
  );
}
