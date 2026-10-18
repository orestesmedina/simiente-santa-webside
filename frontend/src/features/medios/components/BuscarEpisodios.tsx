import { Button } from '../../../components/Button';
import {
  SEARCHING,
  SEARCHING_STATUS,
  SEARCH_ACTION,
  SEARCH_SUPPORT,
  SEARCH_ZONE_TITLE,
} from '../messages';

export interface BuscarEpisodiosProps {
  /** `true` mientras corre la búsqueda (una sola a la vez, ⟲ D-16). */
  pending: boolean;
  onSearch: () => void;
}

/**
 * Zona «Buscar episodios nuevos» del panel (ux.md §4.3/§4.4, FR-011/FR-018): la
 * **única** automatización del módulo y la única puerta de entrada de contenido
 * (F9 no tiene alta manual, FR-019).
 *
 * · **En reposo**: título de la zona, botón primario y la línea de apoyo que
 *   **dice la idempotencia** en palabras de la iglesia (⟲ D-9).
 * · **Buscando**: el botón queda **deshabilitado** en «Buscando…» y aparece un
 *   `role="status"` con el contexto («Buscando episodios nuevos en YouTube y en
 *   el podcast…»), que el lector de pantalla anuncia solo. Sin barra de progreso
 *   ni porcentajes (⟲ D-18) y sin confirmación previa (⟲ D-16: la búsqueda solo
 *   agrega, nunca borra ni pisa lo decidido).
 *
 * La zona **no depende del listado**: está disponible desde el primer render,
 * también cuando el catálogo todavía está cargando o está vacío (el estado vacío
 * remite aquí, ⟲ D-33).
 */
export function BuscarEpisodios({ pending, onSearch }: BuscarEpisodiosProps) {
  return (
    <section
      aria-labelledby="medios-buscar-title"
      className="space-y-3 rounded border border-slate-200 p-4"
    >
      <h2 id="medios-buscar-title" className="text-xl font-semibold text-slate-900">
        {SEARCH_ZONE_TITLE}
      </h2>
      <Button loading={pending} onClick={onSearch}>
        {pending ? SEARCHING : SEARCH_ACTION}
      </Button>
      <p className="text-slate-700">{SEARCH_SUPPORT}</p>
      {pending && (
        <p role="status" aria-live="polite" className="text-slate-700">
          {SEARCHING_STATUS}
        </p>
      )}
    </section>
  );
}
