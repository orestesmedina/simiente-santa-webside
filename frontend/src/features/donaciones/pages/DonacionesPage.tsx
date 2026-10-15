import type { DonacionTextoAdmin } from '../../../api/ofrendas';
import { Button } from '../../../components/Button';
import { EmptyState } from '../../../components/EmptyState';
import { Notice } from '../../../components/Notice';
import { StatusPill } from '../../../components/StatusPill';
import { DonacionesListsByType } from '../components/DonacionesListsByType';
import { isForbidden } from '../errors';
import { useDonacionTextos } from '../hooks/useDonacionTextos';
import { useDonaciones } from '../hooks/useDonaciones';
import {
  EDIT_ACTION,
  HELP,
  LOADING_TEXTS,
  LOAD_ERROR_TEXTS,
  NO_PERMISSION_ERROR,
  NOTHING_PUBLISHED,
  RETRY,
  TEXT_INTRO_TITLE,
  TEXT_NOT_WRITTEN,
  TEXT_USAGE_TITLE,
  TEXTS_ZONE_TITLE,
  TITLE,
} from '../messages';

/**
 * Tarjeta de un texto de la sección (ux.md §4.3/§4.5): título, fragmento del
 * contenido guardado y `StatusPill` con el estado **propio** del texto; si nunca
 * se ha escrito muestra «Todavía sin escribir.» —un estado válido, no un error—
 * con su botón «Editar» (⟲ D-8). El estado se cambia fuera del formulario, desde
 * esta tarjeta (Q10).
 *
 * El botón «Editar» abre `TextoForm` en un `Dialog` (T731).
 */
function TextCard({ title, texto }: { title: string; texto: DonacionTextoAdmin | null }) {
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
      <div>
        <Button size="sm" variant="secondary" aria-haspopup="dialog">
          {EDIT_ACTION}
        </Button>
      </div>
    </article>
  );
}

/**
 * Gestión de donaciones del panel (`/panel/donaciones`, ux.md §4.3/§5, US2/US3):
 * título, banda de ayuda, **zona «Textos de la sección»** —dos tarjetas con el
 * estado propio de cada texto (Q10)— y el **listado de métodos agrupado por
 * tipo** (`DonacionesListsByType`).
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
                <div className="grid gap-4 sm:grid-cols-2">
                  <TextCard title={TEXT_INTRO_TITLE} texto={intro} />
                  <TextCard title={TEXT_USAGE_TITLE} texto={usage} />
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
    </section>
  );
}
