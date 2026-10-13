import { UPCOMING_QUERY_KEY } from '../../publico/eventos/hooks/useUpcomingEvents';

/**
 * Claves de caché del módulo de agenda (patrón F2/F3). El panel invalida su
 * propia clave tras cada mutación **y** la de la consulta pública compartida
 * (`proximamente`), para que un cambio guardado o publicado se vea en la
 * primera carga posterior de la sección pública (SC-003/FR-014).
 */
export const AGENDA_EVENTS_QUERY_KEY = ['agenda', 'eventos'] as const;
export const AGENDA_ACTIVITIES_QUERY_KEY = ['agenda', 'actividades'] as const;
export const AGENDA_CATEGORIES_QUERY_KEY = ['agenda', 'categorias'] as const;

export { UPCOMING_QUERY_KEY };

/**
 * Tope de página del panel. El contrato admite `limit` hasta 100 (§8.1.3) y la
 * UX no pagina los listados (⟲ D-17): se pide el tope para mostrar «todos» los
 * elementos sin dejar ninguno fuera.
 */
export const AGENDA_PAGE_LIMIT = 100;
