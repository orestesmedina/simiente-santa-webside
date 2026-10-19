import { MEDIOS_ADMIN_PAGE_LIMIT } from '../../../api/medios';
import { PORTADA_MEDIOS_QUERY_KEY } from '../../publico/medios/hooks/usePortadaMedios';
import { PUBLIC_MEDIOS_QUERY_KEY } from '../../publico/medios/hooks/usePublicMedios';

/**
 * Claves de caché del módulo «Medios» (patrón F2/F3/F4/F6/F7). El panel invalida
 * sus propias claves tras cada mutación **y** las dos consultas públicas (sección
 * `/medios` y bloque de la portada), para que un cambio guardado se vea en la
 * primera carga posterior (SC-003/FR-015).
 *
 * La clave del panel vive **bajo** la raíz pública (`['medios', 'admin', …]`):
 * invalidar la sección pública alcanza también al panel, que en todas las
 * mutaciones se invalida igualmente, así que no hay trabajo de más.
 */
export const MEDIOS_ADMIN_QUERY_KEY = ['medios', 'admin'] as const;

export { PORTADA_MEDIOS_QUERY_KEY, PUBLIC_MEDIOS_QUERY_KEY };

/**
 * Tope de página del listado del panel (`MEDIOS_ADMIN_PAGE_LIMIT`: el máximo que
 * admite el contrato). La UX no pagina el listado (ux.md §4.3): se pide el tope
 * para mostrar «todo» el catálogo —visibles, ocultos y marcados— en el orden del
 * visitante (mismo criterio que el panel de F7). Si el catálogo supera el tope,
 * la lista **avisa** de que hay más fichas en vez de truncar en silencio (B3).
 */
export const MEDIOS_PAGE_LIMIT = MEDIOS_ADMIN_PAGE_LIMIT;
