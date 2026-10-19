import { mediaUrl } from '../../../api/client';
import type {
  MarcaDisponibilidad,
  MedioAdmin,
  OrigenMedio,
  Visibilidad,
} from '../../../api/medios';
import { Button } from '../../../components/Button';
import { EmptyState } from '../../../components/EmptyState';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { StatusPill } from '../../../components/StatusPill';
import { Table, type Column } from '../../../components/Table';
import { formatDate } from '../../../lib/format';
import { formatLongDate } from '../../publico/eventos/formatoFechas';
import { episodeDay } from '../../publico/medios/fecha';
import { useOcultarMostrarEpisodio } from '../hooks/useOcultarMostrarEpisodio';
import { useResolverMarca } from '../hooks/useResolverMarca';
import {
  ACTION_EDIT_DESCRIPTION,
  ACTION_HIDE,
  ACTION_REVIEW_DUPLICATE,
  ACTION_SHOW,
  COL_ACTIONS,
  COL_FORMATS,
  COL_MARK,
  COL_STATE,
  COL_TITLE,
  EMPTY_HELP,
  EMPTY_TITLE,
  FILTER_EMPTY,
  FORMAT_AUDIO,
  FORMAT_VIDEO,
  HIDDEN_NOTICE,
  LOADING,
  LOAD_ERROR,
  MARK_CLOSED,
  MARK_RESOLVED,
  RESOLVE_KEEP_HIDDEN,
  RESOLVE_KEEP_VISIBLE,
  RESOLVE_SHOW_AGAIN,
  RETRY,
  SHOWN_NOTICE,
  STATE_HIDDEN,
  STATE_VISIBLE,
  SYSTEM_ERROR,
  importedOn,
  listTruncated,
  markLabel,
} from '../messages';
import {
  enlaceDisponible,
  marcaDisponibilidadPendiente,
  marcasPendientes,
  marcasResueltas,
  parDuplicadoAbierto,
  proveedorDelEpisodio,
} from '../marcas';
import { FuentesList } from './FuentesList';
import { MarcaDeRevision } from './MarcaDeRevision';
import { OrigenBadge } from './OrigenBadge';

/** Chips de formato de la fila (⟲ D-15: el formato se ve, no se deduce). */
const FORMATS_BY_ORIGIN: Record<OrigenMedio, string[]> = {
  video: [FORMAT_VIDEO],
  audio: [FORMAT_AUDIO],
  both: [FORMAT_VIDEO, FORMAT_AUDIO],
};

export interface EpisodiosListProps {
  /** Catálogo tal como lo entregó el servidor (mismo orden que ve el visitante). */
  items: MedioAdmin[];
  /**
   * Total de fichas del catálogo (`total` del contrato): si supera lo listado,
   * la lista **avisa** de que hay más (B3) en vez de truncar en silencio.
   */
  total?: number;
  /** Primera carga sin datos en caché. */
  loading: boolean;
  /** La consulta falló (con o sin caché previa). */
  error: boolean;
  /** Hay listado en caché (el error se avisa en discreto, sin vaciar la lista). */
  hasData: boolean;
  /** Vuelve a pedir el listado (acción de «Reintentar»). */
  onRetry: () => void;
  /** `true` con un filtro de marcas activo: su vacío es un logro, no un hueco (§4.8). */
  filtered: boolean;
  /** Búsqueda en curso: **todas** las acciones de fila quedan deshabilitadas (⟲ D-16). */
  busy: boolean;
  /** Deja el aviso de la operación en la página (patrón F3/F7). */
  onFeedback: (variant: NoticeVariant, message: string) => void;
  /** Abre el diálogo de descripción (T937); sin él la acción no se pinta. */
  onEditDescription?: (item: MedioAdmin) => void;
  /** Abre el diálogo de unión (T938); sin él la acción no se pinta. */
  onReviewDuplicate?: (item: MedioAdmin) => void;
  /** Abre la confirmación de «Mostrar de nuevo» (T938); sin él la acción no se pinta. */
  onShowAgain?: (item: MedioAdmin) => void;
}

/** Miniatura de 48 px de la fila; marcador neutro si la ficha no tiene carátula. */
function Miniatura({ item }: { item: MedioAdmin }) {
  const src = mediaUrl(item.coverUrl);
  if (!src) {
    return (
      <span
        data-testid="miniatura-neutra"
        aria-hidden="true"
        className="flex h-12 w-12 shrink-0 items-center justify-center rounded border border-slate-200 bg-slate-100 text-slate-500"
      >
        ♪
      </span>
    );
  }
  return (
    <img
      src={src}
      alt=""
      className="h-12 w-12 shrink-0 rounded border border-slate-200 object-cover"
    />
  );
}

/**
 * Listado del catálogo de medios (ux.md §4.3/§5.b, FR-019): `Table` en escritorio
 * y tarjetas etiquetadas en móvil, con **una sola lista** —visibles, ocultos y
 * marcados: el equipo siempre ve lo que el público ve hoy— en el **orden del
 * visitante** (FR-006: sin orden manual, ⟲ D-11).
 *
 * Por fila: carátula miniatura (o marcador neutro), **nombre** (no editable) con
 * su **fecha** de publicación, los **chips de formato** («Video»/«Audio», ⟲ D-15),
 * el **origen** con la **fecha de importación** y el detalle de sus fuentes
 * (FR-019/⟲ A2), el estado (`StatusPill` «Visible»/«Oculto»), las **marcas
 * pendientes** con su explicador —puede haber **dos**, una por eje (P9-19); una
 * marca **ya resuelta** se conserva como registro («Resuelta» + cuál era), que
 * no se reabre por la misma causa— y las acciones: «Editar descripción» ·
 * «Ocultar»/«Mostrar» · «Revisar duplicado» (solo con el par **abierto**) · las
 * resoluciones de la marca de disponibilidad.
 *
 * Reglas de las acciones:
 *  · **Ocultar es directo** (sin confirmación: es reversible con «Mostrar») y su
 *    aviso dice el efecto real; «Mostrar» devuelve el episodio a su lugar por
 *    fecha (⟲ D-8).
 *  · Con una **marca de disponibilidad pendiente**, las acciones de esa marca
 *    cubren las dos direcciones —«Mostrar de nuevo»/«Mantenerse oculto» y
 *    «Mantener visible»/«Ocultar»—, así que el botón genérico no se duplica: cada
 *    resolución viaja con **su `mark`** y cierra solo su eje (FR-017/P9-19).
 *  · Durante una búsqueda (`busy`) **todas** quedan deshabilitadas: nada se
 *    cambia a medias (FR-023/⟲ D-16).
 *
 * Los diálogos los aportan T937 (descripción) y T938 (unión y «Mostrar de
 * nuevo»): la lista solo deja los ganchos, de modo que nunca se pinta un botón
 * sin salida.
 */
export function EpisodiosList({
  items,
  total,
  loading,
  error,
  hasData,
  onRetry,
  filtered,
  busy,
  onFeedback,
  onEditDescription,
  onReviewDuplicate,
  onShowAgain,
}: EpisodiosListProps) {
  const ocultarMostrar = useOcultarMostrarEpisodio();
  const resolver = useResolverMarca();

  const rowBusy = busy || ocultarMostrar.isPending || resolver.isPending;

  const cambiarVisibilidad = (item: MedioAdmin) => {
    const visibility: Visibilidad = item.visibility === 'visible' ? 'hidden' : 'visible';
    ocultarMostrar.mutate(
      { id: item.id, visibility },
      {
        onSuccess: () =>
          onFeedback('success', visibility === 'hidden' ? HIDDEN_NOTICE : SHOWN_NOTICE),
        onError: () => onFeedback('error', SYSTEM_ERROR),
      },
    );
  };

  const resolverMarca = (item: MedioAdmin, mark: MarcaDisponibilidad, visibility: Visibilidad) => {
    resolver.mutate(
      { id: item.id, input: { mark, visibility } },
      {
        onSuccess: () => onFeedback('success', MARK_CLOSED),
        onError: () => onFeedback('error', SYSTEM_ERROR),
      },
    );
  };

  const columns: Column<MedioAdmin>[] = [
    {
      key: 'title',
      header: COL_TITLE,
      render: (item) => (
        <div className="flex items-start gap-3">
          <Miniatura item={item} />
          <div className="flex min-w-0 flex-col gap-1">
            <span className="font-medium text-slate-900">{item.title}</span>
            <span className="text-xs text-slate-600">
              {formatLongDate(episodeDay(item.publishedAt), 'es')}
            </span>
            <span className="flex flex-wrap items-center gap-2">
              <OrigenBadge origin={item.origin} />
              <span className="text-xs text-slate-600">
                {importedOn(formatDate(item.createdAt))}
              </span>
            </span>
            <FuentesList sources={item.sources} />
          </div>
        </div>
      ),
    },
    {
      key: 'formats',
      header: COL_FORMATS,
      render: (item) => (
        <span className="flex flex-wrap gap-1">
          {FORMATS_BY_ORIGIN[item.origin].map((format) => (
            <span
              key={format}
              className="inline-flex items-center rounded border border-slate-300 px-2 py-0.5 text-xs font-medium text-slate-700"
            >
              {format}
            </span>
          ))}
        </span>
      ),
    },
    {
      key: 'state',
      header: COL_STATE,
      render: (item) => (
        <StatusPill
          value={item.visibility === 'visible' ? 'active' : 'inactive'}
          label={item.visibility === 'visible' ? STATE_VISIBLE : STATE_HIDDEN}
        />
      ),
    },
    {
      key: 'mark',
      header: COL_MARK,
      render: (item) => {
        const marks = marcasPendientes(item);
        if (marks.length === 0) {
          // Una marca **resuelta** se conserva como registro («Resuelta» + cuál
          // era): dice que el equipo ya pasó por ahí y que no se reabre por la
          // misma causa (FR-017/US3 esc. 6).
          const resueltas = marcasResueltas(item);
          if (resueltas.length === 0) {
            return null;
          }
          return (
            <div className="flex flex-col gap-1">
              <div>
                <StatusPill value="draft" label={MARK_RESOLVED} />
              </div>
              <p className="text-xs text-slate-600">{resueltas.map(markLabel).join(' · ')}</p>
            </div>
          );
        }
        return (
          <div className="flex flex-col gap-2">
            {marks.map((mark) => (
              <MarcaDeRevision
                key={mark}
                mark={mark}
                provider={proveedorDelEpisodio(item)}
                linkAvailable={enlaceDisponible(item)}
              />
            ))}
          </div>
        );
      },
    },
    {
      key: 'actions',
      header: COL_ACTIONS,
      render: (item) => {
        const disponibilidad = marcaDisponibilidadPendiente(item);
        return (
          <div className="flex flex-wrap items-center gap-2">
            {onEditDescription && (
              <Button
                size="sm"
                variant="secondary"
                aria-haspopup="dialog"
                disabled={rowBusy}
                onClick={() => onEditDescription(item)}
              >
                {ACTION_EDIT_DESCRIPTION}
              </Button>
            )}

            {onReviewDuplicate && parDuplicadoAbierto(item) && (
              <Button
                size="sm"
                variant="secondary"
                aria-haspopup="dialog"
                disabled={rowBusy}
                onClick={() => onReviewDuplicate(item)}
              >
                {ACTION_REVIEW_DUPLICATE}
              </Button>
            )}

            {disponibilidad === 'unavailable' && (
              <>
                {onShowAgain && (
                  <Button
                    size="sm"
                    variant="secondary"
                    aria-haspopup="dialog"
                    disabled={rowBusy}
                    onClick={() => onShowAgain(item)}
                  >
                    {RESOLVE_SHOW_AGAIN}
                  </Button>
                )}
                <Button
                  size="sm"
                  variant="secondary"
                  disabled={rowBusy}
                  onClick={() => resolverMarca(item, 'unavailable', 'hidden')}
                >
                  {RESOLVE_KEEP_HIDDEN}
                </Button>
              </>
            )}

            {disponibilidad === 'missing_from_feed' && (
              <>
                <Button
                  size="sm"
                  variant="secondary"
                  disabled={rowBusy}
                  onClick={() => resolverMarca(item, 'missing_from_feed', 'visible')}
                >
                  {RESOLVE_KEEP_VISIBLE}
                </Button>
                <Button
                  size="sm"
                  variant="secondary"
                  disabled={rowBusy}
                  onClick={() => resolverMarca(item, 'missing_from_feed', 'hidden')}
                >
                  {ACTION_HIDE}
                </Button>
              </>
            )}

            {!disponibilidad && (
              <Button
                size="sm"
                variant="secondary"
                disabled={rowBusy}
                onClick={() => cambiarVisibilidad(item)}
              >
                {item.visibility === 'visible' ? ACTION_HIDE : ACTION_SHOW}
              </Button>
            )}
          </div>
        );
      },
    },
  ];

  if (loading) {
    return <EmptyState kind="loading" message={LOADING} />;
  }

  if (error && !hasData) {
    return (
      <EmptyState
        kind="error"
        message={LOAD_ERROR}
        action={<Button onClick={onRetry}>{RETRY}</Button>}
      />
    );
  }

  if (items.length === 0) {
    // Con filtro activo el vacío es un logro del equipo (ux.md §4.8); sin filtro
    // remite a la zona de importación, que ya tiene el único botón (⟲ D-33).
    return <EmptyState message={filtered ? FILTER_EMPTY : `${EMPTY_TITLE} ${EMPTY_HELP}`} />;
  }

  const truncated = typeof total === 'number' && total > items.length;

  return (
    <div className="space-y-3">
      {error && hasData && <Notice variant="warning">{LOAD_ERROR}</Notice>}
      {truncated && <Notice variant="warning">{listTruncated(total, items.length)}</Notice>}
      <Table
        columns={columns}
        rows={items}
        rowKey={(item) => item.id}
        caption="Episodios de medios importados"
      />
    </div>
  );
}
