import { PORTADA_OFRENDAS_QUERY_KEY } from '../../publico/ofrendas/hooks/usePortadaOfrendas';
import { PUBLIC_OFRENDAS_QUERY_KEY } from '../../publico/ofrendas/hooks/usePublicOfrendas';

/**
 * Claves de caché del módulo de donaciones (patrón F2/F3/F4/F6). El panel
 * invalida sus propias claves tras cada mutación **y** las dos consultas
 * públicas (sección `ofrendas` y bloque de la portada), para que un cambio
 * guardado o publicado se vea en la primera carga posterior (SC-003/FR-013).
 */
export const DONACIONES_ADMIN_QUERY_KEY = ['donaciones', 'admin'] as const;
export const DONACIONES_TEXTOS_QUERY_KEY = ['donaciones', 'textos'] as const;

export { PORTADA_OFRENDAS_QUERY_KEY, PUBLIC_OFRENDAS_QUERY_KEY };

/**
 * Tope de página del panel. El contrato admite `limit` hasta 100 (§8.1.3) y la
 * UX no pagina el listado (ux.md §4.3): se pide el tope para mostrar «todos» los
 * métodos en el orden del visitante (`iban` primero; `sortOrder`, `id` en cada
 * grupo) sin dejar ninguno fuera.
 */
export const DONACIONES_PAGE_LIMIT = 100;
