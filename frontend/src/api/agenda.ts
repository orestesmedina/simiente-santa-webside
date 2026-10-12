import { apiFetch, buildUrl } from './client';
import type { components } from './schema';

/**
 * Cliente tipado del módulo «Eventos y actividades» (F4), generado desde
 * `backend/api/openapi.yaml` (v0.5.0, T406/T407). Ningún componente llama a
 * `fetch`: todo pasa por aquí (regla F2/T2xx) y los tipos salen **solo** del
 * contrato vivo (`schema.d.ts`).
 */

/** Idioma soportado por la sección pública (FR-007/FR-008). */
export type AgendaLang = 'es' | 'en';

export type PublicationState = components['schemas']['PublicationState'];

export type ProximamentePublica = components['schemas']['ProximamentePublica'];
export type EventoPublico = components['schemas']['EventoPublico'];
export type ActividadPublica = components['schemas']['ActividadPublica'];
export type OcurrenciaPublica = components['schemas']['OcurrenciaPublica'];
export type CategoriaPublica = components['schemas']['CategoriaPublica'];
export type ImagenPublica = components['schemas']['ImagenPublica'];

export type EventoAdmin = components['schemas']['EventoAdmin'];
export type EventoAdminList = components['schemas']['EventoAdminList'];
export type EventoInput = components['schemas']['EventoInput'];
export type EventoPatch = components['schemas']['EventoPatch'];
export type FechaAdmin = components['schemas']['FechaAdmin'];
export type FechaInput = components['schemas']['FechaInput'];

export type ActividadAdmin = components['schemas']['ActividadAdmin'];
export type ActividadAdminList = components['schemas']['ActividadAdminList'];
export type ActividadInput = components['schemas']['ActividadInput'];
export type ActividadPatch = components['schemas']['ActividadPatch'];

export type CategoriaAdmin = components['schemas']['CategoriaAdmin'];
export type CategoriaAdminList = components['schemas']['CategoriaAdminList'];
export type CategoriaInput = components['schemas']['CategoriaInput'];
export type CategoriaPatch = components['schemas']['CategoriaPatch'];

export type ImageUploadResult = components['schemas']['ImageUploadResult'];

/** Raíces del contrato (P4-16: las rutas no cambian). */
export const EVENTS_BASE = '/api/v1/admin/eventos';
export const ACTIVITIES_BASE = '/api/v1/admin/actividades';
export const CATEGORIES_BASE = '/api/v1/admin/categorias';

const JSON_HEADERS = { 'Content-Type': 'application/json' } as const;

/** Paginación del panel (`platform/paginate`) y filtro de estado (FR-013). */
export interface AgendaListParams {
  limit?: number;
  offset?: number;
  publicationState?: PublicationState;
}

/** Traduce los parámetros del listado a la forma que acepta `buildUrl`. */
function listQuery(params: AgendaListParams): Record<string, string | number | undefined> {
  return { limit: params.limit, offset: params.offset, publicationState: params.publicationState };
}

/**
 * «Lo que viene» (FR-001…FR-008): publicado **y** con fecha próxima, ya
 * resuelto por el servidor al idioma pedido (fallback `en → es` por campo) y
 * con las claves de bloque **omitidas** si están vacías (SC-012). Sin sesión y
 * `no-store`: el cambio guardado se ve en la primera carga posterior (SC-003).
 */
export async function getProximamente(lang: AgendaLang): Promise<ProximamentePublica> {
  return apiFetch<ProximamentePublica>(buildUrl('/api/v1/proximamente', { lang }));
}

/* ── Eventos (panel) ─────────────────────────────────────────────────── */

/** Listado paginado de eventos del panel (FR-010). */
export async function listEventos(params: AgendaListParams = {}): Promise<EventoAdminList> {
  return apiFetch<EventoAdminList>(buildUrl(EVENTS_BASE, listQuery(params)));
}

/** Alta de evento con su lista de fechas (FR-015). Nace en borrador salvo estado explícito. */
export async function createEvento(input: EventoInput): Promise<EventoAdmin> {
  return apiFetch<EventoAdmin>(EVENTS_BASE, {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Edición parcial: si `dates` viene, **reemplaza la lista completa** (R4-11). */
export async function updateEvento(id: string, input: EventoPatch): Promise<EventoAdmin> {
  return apiFetch<EventoAdmin>(`${EVENTS_BASE}/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Elimina un evento y sus fechas (FR-010). */
export async function deleteEvento(id: string): Promise<void> {
  return apiFetch<void>(`${EVENTS_BASE}/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

/** Sube la imagen de un evento (FR-020); exige el permiso `eventos`. */
export async function uploadEventoImage(file: File): Promise<ImageUploadResult> {
  return uploadModuleImage(`${EVENTS_BASE}/imagenes`, file);
}

/* ── Actividades (panel) ─────────────────────────────────────────────── */

/** Listado paginado de actividades del panel (FR-011). */
export async function listActividades(params: AgendaListParams = {}): Promise<ActividadAdminList> {
  return apiFetch<ActividadAdminList>(buildUrl(ACTIVITIES_BASE, listQuery(params)));
}

/** Alta de actividad puntual: fecha única, sin periodicidad (FR-015/Decisión 3). */
export async function createActividad(input: ActividadInput): Promise<ActividadAdmin> {
  return apiFetch<ActividadAdmin>(ACTIVITIES_BASE, {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Edición parcial de una actividad (los campos ausentes no se tocan). */
export async function updateActividad(id: string, input: ActividadPatch): Promise<ActividadAdmin> {
  return apiFetch<ActividadAdmin>(`${ACTIVITIES_BASE}/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Elimina una actividad (FR-011). */
export async function deleteActividad(id: string): Promise<void> {
  return apiFetch<void>(`${ACTIVITIES_BASE}/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

/** Sube la imagen de una actividad (FR-020); exige el permiso `actividades`. */
export async function uploadActividadImage(file: File): Promise<ImageUploadResult> {
  return uploadModuleImage(`${ACTIVITIES_BASE}/imagenes`, file);
}

/* ── Categorías (catálogo del panel) ─────────────────────────────────── */

/**
 * Catálogo completo de categorías (Q6), acotado en el servidor (≤ 100) y
 * ordenado por `sortOrder`; sin paginación (excepción documentada §8.1.3).
 */
export async function listCategorias(): Promise<CategoriaAdminList> {
  return apiFetch<CategoriaAdminList>(CATEGORIES_BASE);
}

/** Alta en el catálogo (nombre en español obligatorio, en opcional). */
export async function createCategoria(input: CategoriaInput): Promise<CategoriaAdmin> {
  return apiFetch<CategoriaAdmin>(CATEGORIES_BASE, {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Edición parcial de una categoría (renombrar se ve de inmediato en lo publicado). */
export async function updateCategoria(id: string, input: CategoriaPatch): Promise<CategoriaAdmin> {
  return apiFetch<CategoriaAdmin>(`${CATEGORIES_BASE}/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/**
 * Elimina una categoría (Q6/Q7). Si está **en uso** el servidor responde
 * `409 conflict` (FK `ON DELETE RESTRICT`, decisión firme P4-6): la interfaz lo
 * explica y **no** ofrece desligarla de los elementos.
 */
export async function deleteCategoria(id: string): Promise<void> {
  return apiFetch<void>(`${CATEGORIES_BASE}/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

/**
 * Subida de imagen común a los dos módulos (multipart con el campo `file`); el
 * servidor detecta el tipo por firma binaria y genera el nombre.
 */
async function uploadModuleImage(path: string, file: File): Promise<ImageUploadResult> {
  const body = new FormData();
  body.append('file', file);
  return apiFetch<ImageUploadResult>(path, { method: 'POST', body });
}
