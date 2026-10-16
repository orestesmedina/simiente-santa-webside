import { useState } from 'react';
import type { DonacionTextoAdmin, DonacionTextKey } from '../../../api/ofrendas';
import { Button } from '../../../components/Button';
import { EmptyState } from '../../../components/EmptyState';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { PublishControls } from '../../../components/PublishControls';
import { StatusPill } from '../../../components/StatusPill';
import { DonacionesListsByType } from '../components/DonacionesListsByType';
import { TextoForm } from '../components/TextoForm';
import { isForbidden } from '../errors';
import { useDonacionTextos } from '../hooks/useDonacionTextos';
import { useDonaciones } from '../hooks/useDonaciones';
import { usePutDonacionTexto } from '../hooks/usePutDonacionTexto';
import {
  EDIT_ACTION,
  HELP,
  LOADING_TEXTS,
  LOAD_ERROR_TEXTS,
  NO_PERMISSION_ERROR,
  NOTHING_PUBLISHED,
  PUBLISHED,
  RETRY,
  RETIRE_ACTION,
  RETIRE_CANCEL,
  RETIRE_CONFIRM_TITLE,
  TEXT_INTRO_TITLE,
  TEXT_NOT_WRITTEN,
  TEXT_USAGE_TITLE,
  TEXTS_ZONE_TITLE,
  TITLE,
  UNPUBLISHED,
  retireConfirmDescription,
} from '../messages';

interface TextFeedback {
  variant: NoticeVariant;
  message: string;
}

interface TextCardProps {
  title: string;
  texto: DonacionTextoAdmin | null;
  busy: boolean;
  onEdit: () => void;
  onPublish: () => void;
  onUnpublish: () => void;
}

/**
 * Tarjeta de un texto de la sección (ux.md §4.3/§4.5): título, fragmento del
 * contenido guardado y `StatusPill` con el estado **propio** del texto; si nunca
 * se ha escrito muestra «Todavía sin escribir.» —un estado válido, no un error—
 * con su botón «Editar» (⟲ D-8). El estado se cambia desde aquí con
 * `PublishControls`, nunca dentro del formulario (Q10/⟲ D-11).
 */
function TextCard({ title, texto, busy, onEdit, onPublish, onUnpublish }: TextCardProps) {
  return (
    <article className="flex flex-col gap-2 rounded border border-slate-200 p-4">
      <h3 className="font-semibold text-slate-900">{title}</h3>
      {texto ? (
        <>
          <div>
            <StatusPill value={texto.publicationState} />
          </div>
          <p className="line-clamp-3 whitespace-pre-line text-slate-700">{texto.textEs}</p>
        </>
      ) : (
        <p className="text-slate-700">{TEXT_NOT_WRITTEN}</p>
      )}
      <div className="flex flex-wrap items-center gap-2">
        <Button size="sm" variant="secondary" aria-haspopup="dialog" onClick={onEdit}>
          {EDIT_ACTION}
        </Button>
        {texto && (
          <PublishControls
            name={title}
            state={texto.publicationState}
            loading={busy}
            retireLabel={RETIRE_ACTION}
            confirmTitle={RETIRE_CONFIRM_TITLE}
            confirmDescription={retireConfirmDescription}
            cancelText={RETIRE_CANCEL}
            onPublish={onPublish}
            onUnpublish={onUnpublish}
          />
        )}
      </div>
    </article>
  );
}

/**
 * Gestión de donaciones del panel (`/panel/donaciones`, ux.md §4.3/§5, US2/US3):
 * título, banda de ayuda, **zona «Textos de la sección»** —dos tarjetas con el
 * estado propio de cada texto (Q10) y su edición en `TextoForm`— y el **listado
 * de métodos agrupado por tipo** (`DonacionesListsByType`).
 *
 * Avisa además cuando **nada está publicado** (ni métodos ni textos): la sección
 * y su bloque de portada desaparecen y `/ofrendas` redirige a la portada
 * (FR-005/FR-022/Q9). La autorización es del servidor (`AdminChain("donaciones")`):
 * el permiso `donaciones` protege la ruta en la interfaz y, si el servidor
 * rechaza en caliente, se avisa sin detalles internos (FR-011).
 */
export function DonacionesPage() {
  const donaciones = useDonaciones();
  const textos = useDonacionTextos();
  const putTexto = usePutDonacionTexto();

  const [editingText, setEditingText] = useState<DonacionTextKey | null>(null);
  const [textFeedback, setTextFeedback] = useState<TextFeedback | null>(null);

  const forbidden = donaciones.isError && isForbidden(donaciones.error);
  const hasAccounts = Boolean(donaciones.data);
  const hasTexts = Boolean(textos.data);

  const items = donaciones.data?.items ?? [];
  const intro = textos.data?.intro ?? null;
  const usage = textos.data?.usage ?? null;

  const hasContent = items.length > 0 || intro !== null || usage !== null;
  const anyPublished =
    items.some((item) => item.publicationState === 'published') ||
    intro?.publicationState === 'published' ||
    usage?.publicationState === 'published';
  const nothingPublished = hasContent && !anyPublished;

  /** Cambia el estado de un texto desde su tarjeta (Q10), conservando su cuerpo. */
  const changeTextState = (key: DonacionTextKey, texto: DonacionTextoAdmin, publish: boolean) => {
    putTexto.mutate(
      {
        textKey: key,
        input: {
          textEs: texto.textEs,
          textEn: texto.textEn ?? null,
          publicationState: publish ? 'published' : 'draft',
        },
      },
      {
        onSuccess: () =>
          setTextFeedback({ variant: 'success', message: publish ? PUBLISHED : UNPUBLISHED }),
      },
    );
  };

  const handleTextDone = (feedback: { message: string }) => {
    setEditingText(null);
    setTextFeedback({ variant: 'success', message: feedback.message });
  };

  return (
    <section className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">{TITLE}</h1>
        <p className="mt-2 text-slate-700">{HELP}</p>
      </div>

      {forbidden ? (
        <Notice variant="error">{NO_PERMISSION_ERROR}</Notice>
      ) : (
        <>
          <section aria-labelledby="donaciones-textos-title" className="space-y-3">
            <h2 id="donaciones-textos-title" className="text-xl font-semibold text-slate-900">
              {TEXTS_ZONE_TITLE}
            </h2>
            {textos.isPending ? (
              <EmptyState kind="loading" message={LOADING_TEXTS} />
            ) : textos.isError && !hasTexts ? (
              <EmptyState
                kind="error"
                message={LOAD_ERROR_TEXTS}
                action={<Button onClick={() => void textos.refetch()}>{RETRY}</Button>}
              />
            ) : (
              <>
                {textos.isError && <Notice variant="warning">{LOAD_ERROR_TEXTS}</Notice>}
                {textFeedback && (
                  <Notice variant={textFeedback.variant}>{textFeedback.message}</Notice>
                )}
                <div className="grid gap-4 sm:grid-cols-2">
                  <TextCard
                    title={TEXT_INTRO_TITLE}
                    texto={intro}
                    busy={putTexto.isPending}
                    onEdit={() => setEditingText('intro')}
                    onPublish={() => intro && changeTextState('intro', intro, true)}
                    onUnpublish={() => intro && changeTextState('intro', intro, false)}
                  />
                  <TextCard
                    title={TEXT_USAGE_TITLE}
                    texto={usage}
                    busy={putTexto.isPending}
                    onEdit={() => setEditingText('usage')}
                    onPublish={() => usage && changeTextState('usage', usage, true)}
                    onUnpublish={() => usage && changeTextState('usage', usage, false)}
                  />
                </div>
              </>
            )}
          </section>

          {nothingPublished && <Notice variant="warning">{NOTHING_PUBLISHED}</Notice>}

          <DonacionesListsByType
            donaciones={items}
            loading={donaciones.isPending}
            error={donaciones.isError && !hasAccounts}
            onRetry={() => void donaciones.refetch()}
          />
        </>
      )}

      {editingText && (
        <TextoForm
          textKey={editingText}
          texto={editingText === 'intro' ? intro : usage}
          onDone={handleTextDone}
          onCancel={() => setEditingText(null)}
        />
      )}
    </section>
  );
}
