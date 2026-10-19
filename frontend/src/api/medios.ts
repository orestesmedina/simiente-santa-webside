import { apiFetch, buildUrl } from './client';
import type { components } from './schema';

/**
 * Cliente tipado del módulo «Medios» (F9), generado desde
 * `backend/api/openapi.yaml` (v0.9.0, T905). Ningún componente llama a `fetch`:
 * todo pasa por aquí (regla F2) y los tipos salen **solo** del contrato vivo
 * (`schema.d.ts`, 0 `any`).
 *
 * El cliente **no filtra ni reimplementa** el fallback de idioma (`en → es`, solo
 * sobre la descripción corta), ni el orden (fecha descendente con desempate por
 * importación), ni la regla de publicables (visibilidad, absorbidas y marcas):
 * el servidor ya resolvió todo (FR-009/FR-025/FR-026). Tampoco reproduce ni
 * incrusta nada de YouTube ni de Spotify (FR-004): los enlaces son texto que el
 * visitante activa y las carátulas son copias servidas por el propio sitio.
 */

/** Idioma soportado por las superficies públicas (FR-025). */
export type MediosLang = 'es' | 'en';

export type Proveedor = components['schemas']['Proveedor'];
export type OrigenMedio = components['schemas']['OrigenMedio'];
export type Visibilidad = components['schemas']['Visibilidad'];
export type MarcaRevision = components['schemas']['MarcaRevision'];
export type MarcaDisponibilidad = components['schemas']['MarcaDisponibilidad'];
export type ResultadoDisponibilidad = components['schemas']['ResultadoDisponibilidad'];
export type EstadoFeed = components['schemas']['EstadoFeed'];

/* ── Público ─────────────────────────────────────────────────────────── */

export type MediosPublicos = components['schemas']['MediosPublicos'];
export type MedioPublico = components['schemas']['MedioPublico'];

/** Bloque «Medios» de la portada (US6/FR-027, tope 5 en el servidor). */
export type PortadaMedios = components['schemas']['PortadaMedios'];

/* ── Panel ───────────────────────────────────────────────────────────── */

export type MedioAdmin = components['schemas']['MedioAdmin'];
export type MedioAdminList = components['schemas']['MedioAdminList'];
export type FuenteAdmin = components['schemas']['FuenteAdmin'];
export type MedioPatch = components['schemas']['MedioPatch'];
export type UnirInput = components['schemas']['UnirInput'];
export type ResolverMarcaInput = components['schemas']['ResolverMarcaInput'];
export type ResumenBusqueda = components['schemas']['ResumenBusqueda'];
export type SyncRunAdmin = components['schemas']['SyncRunAdmin'];

/** Raíz de la sección pública (sin sesión y `Cache-Control: no-store`). */
export const MEDIOS_PUBLIC_BASE = '/api/v1/medios';

/** Raíz del bloque de la portada (tope 5 firme en el servidor). */
export const MEDIOS_PORTADA_BASE = '/api/v1/portada/medios';

/** Raíz del módulo de panel (el permiso es `medios`, P9-15). */
export const MEDIOS_BASE = '/api/v1/admin/medios';

/**
 * Tanda de la sección pública (default revisable #5): el contrato admite hasta
 * 50, pero la lista agrega tandas de 12 con «Cargar más» (Q11/D-3).
 */
export const MEDIOS_PAGE_LIMIT = 12;

/**
 * Tope de página del listado del **panel**: el máximo que admite el contrato
 * (`Limit`). El panel pide «todo» el catálogo con este tope (criterio heredado
 * de F7); cuando el catálogo lo supera, la lista **lo dice** en vez de truncar
 * en silencio (B3 de la revisión).
 */
export const MEDIOS_ADMIN_PAGE_LIMIT = 100;

const JSON_HEADERS = { 'Content-Type': 'application/json' } as const;

/** Paginación y filtros del listado del panel (FR-015/FR-017). */
export interface MediosListParams {
  limit?: number;
  offset?: number;
  /** Filtra por visibilidad; sin valor: todas. */
  visibility?: Visibilidad;
  /** Filtra por la marca de esa clase **pendiente** de resolver; sin valor: todas. */
  reviewMark?: MarcaRevision;
  /** Incluye las fichas absorbidas por una unión (por defecto el servidor no las lista). */
  includeMerged?: boolean;
}

/** Traduce los parámetros del listado a la forma que acepta `buildUrl`. */
function listQuery(params: MediosListParams): Record<string, string | number | undefined> {
  return {
    limit: params.limit,
    offset: params.offset,
    visibility: params.visibility,
    reviewMark: params.reviewMark,
    includeMerged: params.includeMerged === undefined ? undefined : String(params.includeMerged),
  };
}

/* ── Público ─────────────────────────────────────────────────────────── */

/**
 * Sección pública `/medios` (FR-002/FR-006): una **tanda** de fichas
 * publicables —solo visibles, no absorbidas y sin par de duplicados pendiente
 * (FR-009)— del más reciente al más antiguo, con `total` para saber cuántas
 * quedan. El servidor resuelve el idioma (fallback `en → es` solo en la
 * descripción) y **omite** las claves `description`/`coverUrl`/`videoUrl`/
 * `audioUrl` cuando no aplican (SC-010: sin líneas vacías). Con 0 publicables la
 * respuesta sigue siendo `200` con `items` vacío y `total: 0`: la decisión de
 * redirigir a la portada es del frontend (P9-16). Sin sesión y `no-store`.
 */
export async function getMedios(
  lang: MediosLang,
  limit: number = MEDIOS_PAGE_LIMIT,
  offset = 0,
): Promise<MediosPublicos> {
  return apiFetch<MediosPublicos>(buildUrl(MEDIOS_PUBLIC_BASE, { lang, limit, offset }));
}

/**
 * Bloque «Medios» de la portada (US6/FR-027/SC-015): hasta **5** fichas
 * publicables (tope firme del servidor, sin parámetro de límite) con el
 * **mismo** DTO de ítem que la sección y en el mismo orden. `items` se **omite**
 * cuando no hay ninguna publicable: la portada **no dibuja el bloque** (sin
 * mensaje). `no-store`, sin sesión.
 */
export async function getPortadaMedios(lang: MediosLang): Promise<PortadaMedios> {
  return apiFetch<PortadaMedios>(buildUrl(MEDIOS_PORTADA_BASE, { lang }));
}

/* ── Panel ───────────────────────────────────────────────────────────── */

/**
 * Listado del panel (FR-019): incluye las fichas **ocultas** y las marcadas, con
 * el **mismo orden** que ve el visitante, el origen de cada entrada (`origin` y
 * `createdAt`) y las `sources` de cada enlace. `lastSync` viene solo si ya se
 * ejecutó alguna búsqueda (FR-018).
 */
export async function listMedios(params: MediosListParams = {}): Promise<MedioAdminList> {
  return apiFetch<MedioAdminList>(buildUrl(MEDIOS_BASE, listQuery(params)));
}

/**
 * Tiempo de espera **propio** de la búsqueda de episodios: 120 s, por encima del
 * presupuesto del servidor (90 s). El resto de consultas usa los 5 s del cliente
 * (`REQUEST_TIMEOUT_MS`, D20), pero una primera importación real (dos feeds,
 * hasta 30 elementos y hasta 30 descargas de carátula) tarda entre 3 y 30 s:
 * con el tiempo del cliente la búsqueda se abortaba a media importación —y el
 * abort cancelaba también el trabajo del servidor—, de modo que el equipo veía
 * «error del sistema» y **nunca** el resumen de FR-018 (H1 de la revisión).
 */
export const SYNC_TIMEOUT_MS = 120_000;

/**
 * Busca una ficha del panel por identificador contra el servidor (B3).
 *
 * El contrato no tiene `GET /api/v1/admin/medios/{id}`, así que recorre las
 * páginas del listado (tope `MEDIOS_ADMIN_PAGE_LIMIT`) hasta encontrarla o
 * agotar `total`. Con esto la unión de duplicados no depende de que la pareja
 * esté en la página que el panel tiene cargada: si el par quedó fuera de las
 * primeras 100 fichas, se trae igual.
 *
 * Devuelve `undefined` cuando la ficha no existe (p. ej. ya se unió a otra).
 */
export async function findMedio(id: string): Promise<MedioAdmin | undefined> {
  let offset = 0;
  for (;;) {
    const pagina = await listMedios({ limit: MEDIOS_ADMIN_PAGE_LIMIT, offset });
    const encontrada = pagina.items.find((item) => item.id === id);
    if (encontrada) {
      return encontrada;
    }
    offset += pagina.items.length;
    if (pagina.items.length === 0 || offset >= pagina.total) {
      return undefined;
    }
  }
}

/**
 * «Buscar episodios nuevos» (FR-011/FR-018): la **única** automatización del
 * módulo. Lee los dos feeds, agrega solo lo que falte, une lo que corresponde y
 * devuelve el resumen —también cuando un feed falló o el presupuesto de tiempo se
 * agotó: eso no es error y nada se pierde (FR-016)—. Puede tardar unos segundos
 * (presupuesto de 90 s en el servidor): quien llama desactiva la interfaz
 * mientras corre y espera con `SYNC_TIMEOUT_MS`, no con el tiempo del cliente.
 */
export async function syncMedios(): Promise<ResumenBusqueda> {
  return apiFetch<ResumenBusqueda>(`${MEDIOS_BASE}/buscar-nuevos`, {
    method: 'POST',
    timeoutMs: SYNC_TIMEOUT_MS,
  });
}

/**
 * Única edición permitida (FR-019/FR-020): descripción corta en español y/o
 * inglés y visibilidad. Cualquier otra clave la rechaza el servidor con `400`.
 * `descriptionEn` exige `descriptionEs` (FR-023) y un `null`/`""` borra el texto.
 */
export async function updateMedio(id: string, input: MedioPatch): Promise<MedioAdmin> {
  return apiFetch<MedioAdmin>(`${MEDIOS_BASE}/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/**
 * Une dos entradas del mismo sermón (FR-014/US3): `{id}` se conserva y `otherId`
 * queda absorbida (sus enlaces pasan a la ficha resultante). Es **definitiva**
 * (default revisable #8): la interfaz confirma el resultado antes de llamar.
 */
export async function mergeMedio(id: string, input: UnirInput): Promise<MedioAdmin> {
  return apiFetch<MedioAdmin>(`${MEDIOS_BASE}/${encodeURIComponent(id)}/unir`, {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/**
 * Resuelve la marca de revisión indicada (FR-017/US3 esc. 6) y aplica la
 * visibilidad elegida. Los dos ejes (duplicados y disponibilidad) se resuelven
 * por separado y una marca resuelta no se reabre por la misma causa.
 */
export async function resolveMedioMark(id: string, input: ResolverMarcaInput): Promise<MedioAdmin> {
  return apiFetch<MedioAdmin>(`${MEDIOS_BASE}/${encodeURIComponent(id)}/marcas/resolver`, {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}
