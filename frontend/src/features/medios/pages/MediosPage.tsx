import { useCallback, useState } from 'react';
import type { MarcaDisponibilidad, MedioAdmin, ResumenBusqueda } from '../../../api/medios';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { BuscarEpisodios } from '../components/BuscarEpisodios';
import { DescripcionForm } from '../components/DescripcionForm';
import { EpisodiosList } from '../components/EpisodiosList';
import { FiltroMarcas, type FiltroValue } from '../components/FiltroMarcas';
import { ResolverMarcaDialog } from '../components/ResolverMarcaDialog';
import { ResumenImportacion } from '../components/ResumenImportacion';
import { UnirDuplicadosDialog } from '../components/UnirDuplicadosDialog';
import { isForbidden } from '../errors';
import { useBuscarEpisodios } from '../hooks/useBuscarEpisodios';
import { useMedios } from '../hooks/useMedios';
import { useResolverMarca } from '../hooks/useResolverMarca';
import { contarPendientes, marcaDisponibilidadPendiente } from '../marcas';
import { HELP, MARK_CLOSED, NO_PERMISSION_ERROR, SYSTEM_ERROR, TITLE } from '../messages';

/**
 * Módulo «Gestión de medios» del panel (`/panel/medios`, ux.md §4.3/§5, US2/US3):
 * título + banda de ayuda, la **zona de búsqueda** arriba (independiente del
 * listado: está disponible desde el primer render), el **resumen** de la última
 * búsqueda —permanente hasta la próxima (⟲ D-22)—, el **filtro de marcas** y el
 * **listado** del catálogo con sus estados (cargando, vacío, error con
 * «Reintentar», sin permiso del servidor y éxito).
 *
 * El panel va **siempre en español** (FR-024) y **no** ofrece alta manual ni
 * edición de datos del feed ni borrado (FR-019): un episodio que no debe verse
 * **se oculta**.
 *
 * El filtro de marcas (⟲ D-10) es un **filtro de vista**: usa el parámetro real
 * `reviewMark` del contrato (la marca de esa clase **pendiente**) y los conteos
 * salen del catálogo **completo** —que se sigue pidiendo para poder contarlos—,
 * de modo que cambiar de pastilla nunca oculta lo que hay pendiente.
 *
 * La autorización es del servidor (`AdminChain("medios")`, FR-021): el permiso
 * `medios` protege la ruta en la interfaz (`RequirePermission`) y, si el
 * servidor rechaza en caliente, se avisa sin detalles internos.
 */
export function MediosPage() {
  const [filtro, setFiltro] = useState<FiltroValue>('todo');
  const [resumen, setResumen] = useState<ResumenBusqueda>();
  const [feedback, setFeedback] = useState<{ variant: NoticeVariant; message: string }>();
  const [editando, setEditando] = useState<MedioAdmin>();
  const [union, setUnion] = useState<{ item: MedioAdmin; pareja: MedioAdmin }>();
  const [mostrarDeNuevo, setMostrarDeNuevo] = useState<{
    item: MedioAdmin;
    mark: MarcaDisponibilidad;
  }>();

  const buscar = useBuscarEpisodios();
  const resolver = useResolverMarca();
  // Catálogo completo (para los conteos del filtro) y vista filtrada. Con «Todo»
  // las dos consultas comparten clave y TanStack Query resuelve una sola
  // petición; con un filtro activo, cada una pide lo suyo.
  const completo = useMedios();
  const conFiltro = useMedios(filtro === 'todo' ? {} : { reviewMark: filtro });
  const lista = filtro === 'todo' ? completo : conFiltro;

  const counts = contarPendientes(completo.data?.items ?? []);
  const forbidden = isForbidden(lista.error);

  // Estables: el `Dialog` del formulario re-ejecuta su efecto (y recoloca el
  // foco) si `onClose` cambia de identidad en cada render.
  const abrirDescripcion = useCallback((item: MedioAdmin) => setEditando(item), []);
  const cerrarDescripcion = useCallback(() => setEditando(undefined), []);

  /**
   * «Revisar duplicado» (⟲ D-28: solo con un par **abierto**): el diálogo
   * necesita **las dos** fichas del par, así que la pareja se busca en el
   * catálogo completo (que incluye las ocultas). Sin pareja no se abre nada.
   */
  const abrirUnion = useCallback(
    (item: MedioAdmin) => {
      const pareja = (completo.data?.items ?? []).find(
        (candidata) => candidata.id === item.duplicateOfId,
      );
      if (pareja) {
        setUnion({ item, pareja });
      }
    },
    [completo.data],
  );
  const cerrarUnion = useCallback(() => setUnion(undefined), []);

  /** «Mostrar de nuevo»: la única resolución con confirmación suave (⟲ D-28). */
  const abrirMostrarDeNuevo = useCallback((item: MedioAdmin) => {
    const mark = marcaDisponibilidadPendiente(item);
    if (mark) {
      setMostrarDeNuevo({ item, mark });
    }
  }, []);
  const cerrarMostrarDeNuevo = useCallback(() => setMostrarDeNuevo(undefined), []);

  const confirmarMostrarDeNuevo = () => {
    if (!mostrarDeNuevo) {
      return;
    }
    resolver.mutate(
      {
        id: mostrarDeNuevo.item.id,
        input: { mark: mostrarDeNuevo.mark, visibility: 'visible' },
      },
      {
        onSuccess: () => {
          setMostrarDeNuevo(undefined);
          setFeedback({ variant: 'success', message: MARK_CLOSED });
        },
        onError: () => {
          setMostrarDeNuevo(undefined);
          setFeedback({ variant: 'error', message: SYSTEM_ERROR });
        },
      },
    );
  };

  const handleBuscar = () => {
    setFeedback(undefined);
    buscar.mutate(undefined, {
      onSuccess: (data) => setResumen(data),
      onError: () => setFeedback({ variant: 'error', message: SYSTEM_ERROR }),
    });
  };

  /** Cierra el diálogo y deja su aviso en la página (patrón F3/F7). */
  const handleDescripcionGuardada = (aviso: { message: string }) => {
    setEditando(undefined);
    setFeedback({ variant: 'success', message: aviso.message });
  };

  /** Cierra el diálogo de unión y deja su aviso (unión o «no son el mismo»). */
  const handleUnionHecha = (aviso: { message: string }) => {
    setUnion(undefined);
    setFeedback({ variant: 'success', message: aviso.message });
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
          <BuscarEpisodios pending={buscar.isPending} onSearch={handleBuscar} />

          {resumen && <ResumenImportacion resumen={resumen} />}

          {feedback && <Notice variant={feedback.variant}>{feedback.message}</Notice>}

          <FiltroMarcas value={filtro} counts={counts} onChange={setFiltro} />

          <EpisodiosList
            items={lista.data?.items ?? []}
            loading={lista.isPending && !lista.data}
            error={lista.isError}
            hasData={Boolean(lista.data)}
            onRetry={() => void lista.refetch()}
            filtered={filtro !== 'todo'}
            busy={buscar.isPending}
            onFeedback={(variant, message) => setFeedback({ variant, message })}
            onEditDescription={abrirDescripcion}
            onReviewDuplicate={abrirUnion}
            onShowAgain={abrirMostrarDeNuevo}
          />
        </>
      )}

      {editando && (
        <DescripcionForm
          item={editando}
          onDone={handleDescripcionGuardada}
          onCancel={cerrarDescripcion}
        />
      )}

      {union && (
        <UnirDuplicadosDialog
          item={union.item}
          pareja={union.pareja}
          onDone={handleUnionHecha}
          onCancel={cerrarUnion}
        />
      )}

      {mostrarDeNuevo && (
        <ResolverMarcaDialog
          item={mostrarDeNuevo.item}
          mark={mostrarDeNuevo.mark}
          loading={resolver.isPending}
          onConfirm={confirmarMostrarDeNuevo}
          onClose={cerrarMostrarDeNuevo}
        />
      )}
    </section>
  );
}
