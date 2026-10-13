import { apiFetch, buildUrl } from './client';
import type { components } from './schema';

/**
 * Cliente tipado del módulo «Ministerios» (F6), generado desde
 * `backend/api/openapi.yaml` (v0.7.0, T604/T605). Ningún componente llama a
 * `fetch`: todo pasa por aquí (regla F2/T2xx) y los tipos salen **solo** del
 * contrato vivo (`schema.d.ts`, 0 `any`).
 *
 * El cliente **no filtra ni reimplementa** el fallback de idioma ni la
 * privacidad del encargado: el servidor ya resolvió todo (FR-008/FR-026). Con la
 * ficha no publicada, el DTO público ni siquiera trae la clave `leader`.
 */

/** Idioma soportado por las superficies públicas (FR-007/FR-008). */
export type MinistryLang = 'es' | 'en';

export type PublicationState = components['schemas']['PublicationState'];

/* ── Público ─────────────────────────────────────────────────────────── */

export type MinisteriosPublicos = components['schemas']['MinisteriosPublicos'];
export type MinisterioPublico = components['schemas']['MinisterioPublico'];
export type LiderPublico = components['schemas']['LiderPublico'];

/** Bloque «Ministerios» de la portada (US6/FR-027, tope 5 en el servidor). */
export type PortadaMinisterios = components['schemas']['PortadaMinisterios'];
export type ItemMinisterioPortada = components['schemas']['ItemMinisterioPortada'];

/* ── Panel ───────────────────────────────────────────────────────────── */

export type MinisterioAdmin = components['schemas']['MinisterioAdmin'];
export type MinisterioAdminList = components['schemas']['MinisterioAdminList'];
export type MinisterioInput = components['schemas']['MinisterioInput'];
export type MinisterioPatch = components['schemas']['MinisterioPatch'];

export type ImageUploadResult = components['schemas']['ImageUploadResult'];

/** Raíz del módulo de panel (el permiso es `ministerios`, P6-6). */
export const MINISTRIES_BASE = '/api/v1/admin/ministerios';

const JSON_HEADERS = { 'Content-Type': 'application/json' } as const;

/** Paginación del panel (`platform/paginate`) y filtro de estado (FR-012). */
export interface MinistriesListParams {
  limit?: number;
  offset?: number;
  publicationState?: PublicationState;
}

/** Traduce los parámetros del listado a la forma que acepta `buildUrl`. */
function listQuery(params: MinistriesListParams): Record<string, string | number | undefined> {
  return { limit: params.limit, offset: params.offset, publicationState: params.publicationState };
}

/**
 * Sección pública `/ministerios` (FR-002): solo los ministerios **publicados**,
 * ya resueltos al idioma pedido (fallback `en → es` por campo) y en el orden
 * manual del servidor (`sortOrder`, `id`). `items` **se omite** cuando no hay
 * ninguno publicado; la decisión de redirigir a la portada es del frontend
 * (P6-21) y el servidor sigue respondiendo `200`. Sin sesión y `no-store`.
 */
export async function getMinisterios(lang: MinistryLang): Promise<MinisteriosPublicos> {
  return apiFetch<MinisteriosPublicos>(buildUrl('/api/v1/ministerios', { lang }));
}

/**
 * Bloque «Ministerios» de la portada (US6/FR-027/FR-028): hasta **5** ítems
 * publicados (tope firme del servidor, sin parámetro de límite) con el DTO
 * mínimo de la tarjeta compacta y **sin campos de encargado en ningún caso**.
 * `items` omitida sin publicados: la portada **no dibuja el bloque** (sin
 * mensaje). `no-store`, sin sesión.
 */
export async function getPortadaMinisterios(lang: MinistryLang): Promise<PortadaMinisterios> {
  return apiFetch<PortadaMinisterios>(buildUrl('/api/v1/portada/ministerios', { lang }));
}

/* ── Panel ───────────────────────────────────────────────────────────── */

/** Listado paginado de ministerios del panel (FR-010). */
export async function listMinisterios(
  params: MinistriesListParams = {},
): Promise<MinisterioAdminList> {
  return apiFetch<MinisterioAdminList>(buildUrl(MINISTRIES_BASE, listQuery(params)));
}

/**
 * Alta de ministerio (FR-010/FR-014): solo `nameEs` y `descriptionEs` son
 * obligatorios. `leaderVisible` viaja **explícitamente** con el valor decidido
 * por el equipo: nace en `false` («Oculto por defecto», D-13/P6-3).
 */
export async function createMinisterio(input: MinisterioInput): Promise<MinisterioAdmin> {
  return apiFetch<MinisterioAdmin>(MINISTRIES_BASE, {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Edición parcial: solo los campos presentes cambian; `null` borra lo opcional. */
export async function updateMinisterio(
  id: string,
  input: MinisterioPatch,
): Promise<MinisterioAdmin> {
  return apiFetch<MinisterioAdmin>(`${MINISTRIES_BASE}/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Elimina un ministerio (FR-010); queda registrado en la auditoría (FR-017). */
export async function deleteMinisterio(id: string): Promise<void> {
  return apiFetch<void>(`${MINISTRIES_BASE}/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

/**
 * Sube la imagen de un ministerio (FR-020): multipart con el campo `file`; el
 * servidor detecta el tipo por firma binaria y genera el nombre. Exige el
 * permiso `ministerios` y no cambia por sí sola el contenido visible.
 */
export async function uploadMinistryImage(file: File): Promise<ImageUploadResult> {
  const body = new FormData();
  body.append('file', file);
  return apiFetch<ImageUploadResult>(`${MINISTRIES_BASE}/imagenes`, { method: 'POST', body });
}
