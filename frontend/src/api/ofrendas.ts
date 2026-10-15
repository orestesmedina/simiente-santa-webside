import { apiFetch, buildUrl } from './client';
import type { components } from './schema';

/**
 * Cliente tipado del módulo de donaciones (F7), generado desde
 * `backend/api/openapi.yaml` (v0.8.0, T705). Ningún componente llama a `fetch`:
 * todo pasa por aquí (regla F2) y los tipos salen **solo** del contrato vivo
 * (`schema.d.ts`, 0 `any`).
 *
 * El cliente **no filtra ni reimplementa** el fallback de idioma, el orden ni la
 * regla de borradores: el servidor ya resolvió todo (FR-004/FR-008/FR-012). La
 * sección pública se llama «ofrendas» y el módulo del panel «donaciones»
 * (Q8: nombres deliberadamente distintos según superficie).
 */

/** Idioma soportado por las superficies públicas (FR-007/FR-008). */
export type OfrendasLang = 'es' | 'en';

export type PublicationState = components['schemas']['PublicationState'];
export type MetodoKind = components['schemas']['MetodoKind'];
export type Moneda = components['schemas']['Moneda'];

/* ── Público ─────────────────────────────────────────────────────────── */

export type OfrendasPublicas = components['schemas']['OfrendasPublicas'];
export type MetodoDonacionPublico = components['schemas']['MetodoDonacionPublico'];
export type TextoPublico = components['schemas']['TextoPublico'];

/** Bloque «Ofrendas y diezmos» de la portada (US6/FR-022, tope 5 en el servidor). */
export type PortadaOfrendas = components['schemas']['PortadaOfrendas'];

/* ── Panel ───────────────────────────────────────────────────────────── */

export type DonacionAdmin = components['schemas']['DonacionAdmin'];
export type DonacionAdminList = components['schemas']['DonacionAdminList'];
export type DonacionInput = components['schemas']['DonacionInput'];
export type DonacionPatch = components['schemas']['DonacionPatch'];
export type DonacionTextosAdmin = components['schemas']['DonacionTextosAdmin'];
export type DonacionTextoAdmin = components['schemas']['DonacionTextoAdmin'];
export type DonacionTextoInput = components['schemas']['DonacionTextoInput'];
export type DonacionTextKey = components['schemas']['DonacionTextoAdmin']['textKey'];

/** Raíz del módulo de panel (el permiso es `donaciones`, P7-6). */
export const DONACIONES_BASE = '/api/v1/admin/donaciones';

const JSON_HEADERS = { 'Content-Type': 'application/json' } as const;

/** Paginación del panel (`platform/paginate`) y filtro de estado (FR-012). */
export interface DonacionesListParams {
  limit?: number;
  offset?: number;
  publicationState?: PublicationState;
}

/** Traduce los parámetros del listado a la forma que acepta `buildUrl`. */
function listQuery(params: DonacionesListParams): Record<string, string | number | undefined> {
  return { limit: params.limit, offset: params.offset, publicationState: params.publicationState };
}

/**
 * Sección pública `/ofrendas` (FR-002): solo los métodos **publicados** (tope
 * 200 del servidor) en el orden de la sección —`iban` primero y luego `sinpe`,
 * cada grupo por `sortOrder`, `id`— y los dos textos publicados, ya resueltos al
 * idioma pedido (fallback `en → es` por campo). Las claves `items`/`intro`/
 * `usage` se **omiten** cuando no hay nada publicado de esa clave; la decisión de
 * redirigir a la portada es del frontend (P7-21) y el servidor sigue `200`. Sin
 * sesión y `no-store`.
 */
export async function getOfrendas(lang: OfrendasLang): Promise<OfrendasPublicas> {
  return apiFetch<OfrendasPublicas>(buildUrl('/api/v1/ofrendas', { lang }));
}

/**
 * Bloque «Ofrendas y diezmos» de la portada (US6/FR-022/SC-014): hasta **5**
 * métodos publicados (tope firme del servidor, sin parámetro de límite) con el
 * **mismo** DTO de ítem que la sección. `items` omitida con 0 métodos publicados:
 * la portada **no dibuja el bloque** (sin mensaje). Los textos de la sección no
 * forman parte de este documento. `no-store`, sin sesión.
 */
export async function getPortadaOfrendas(lang: OfrendasLang): Promise<PortadaOfrendas> {
  return apiFetch<PortadaOfrendas>(buildUrl('/api/v1/portada/ofrendas', { lang }));
}

/* ── Panel ───────────────────────────────────────────────────────────── */

/** Listado paginado de métodos del panel (FR-010), en el orden del visitante. */
export async function listDonaciones(
  params: DonacionesListParams = {},
): Promise<DonacionAdminList> {
  return apiFetch<DonacionAdminList>(buildUrl(DONACIONES_BASE, listQuery(params)));
}

/** Alta de método (FR-010/FR-014): el input acepta el número tal como se escribe. */
export async function createDonacion(input: DonacionInput): Promise<DonacionAdmin> {
  return apiFetch<DonacionAdmin>(DONACIONES_BASE, {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Edición parcial: solo los campos presentes cambian; `null` borra lo opcional. */
export async function updateDonacion(id: string, input: DonacionPatch): Promise<DonacionAdmin> {
  return apiFetch<DonacionAdmin>(`${DONACIONES_BASE}/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Elimina un método (FR-010); queda registrado en la auditoría (FR-017). */
export async function deleteDonacion(id: string): Promise<void> {
  return apiFetch<void>(`${DONACIONES_BASE}/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

/**
 * Los dos textos de la sección (FR-021/Q5/Q11) con sus propios estados: son
 * singletons por clave y cada uno viene en `null` hasta el primer guardado.
 */
export async function getDonacionTextos(): Promise<DonacionTextosAdmin> {
  return apiFetch<DonacionTextosAdmin>(`${DONACIONES_BASE}/textos`);
}

/**
 * Guarda (crea o reemplaza) el texto indicado (FR-021): `textEs` obligatorio,
 * `textEn` opcional con español base por campo, estado propio. **No hay borrado**:
 * un texto se retira poniéndolo en `draft` (P7-17).
 */
export async function putDonacionTexto(
  textKey: DonacionTextKey,
  input: DonacionTextoInput,
): Promise<DonacionTextoAdmin> {
  return apiFetch<DonacionTextoAdmin>(`${DONACIONES_BASE}/textos/${encodeURIComponent(textKey)}`, {
    method: 'PUT',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}
