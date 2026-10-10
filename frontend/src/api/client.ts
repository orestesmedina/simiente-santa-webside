import type { components } from './schema';

export type SystemStatus = components['schemas']['SystemStatus'];
export type ErrorEnvelope = components['schemas']['ErrorEnvelope'];
export type ErrorBody = components['schemas']['ErrorBody'];

/** URL base de la API. Se hornea en build time (plan R6). */
export const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

/** Un `path` ya absoluto: con esquema (`http:`, `https:`, `data:`…) o `//host`. */
const ABSOLUTE_URL_PATTERN = /^(?:[a-z][a-z0-9+.-]*:|\/\/)/i;

/**
 * Resuelve una URL de media devuelta por la API contra la base de la API.
 *
 * La API expone las imágenes de portada como rutas **relativas**
 * (`/api/v1/media/img_….jpg`). Pintarlas tal cual como `src` las resolvería
 * contra el origen de la SPA (p. ej. `http://localhost:5173`), donde el nginx
 * del frontend no proxya `/api/` y devuelve `index.html` (`200 text/html`) en
 * vez de la imagen: el bug de F3. Aquí se prefijan con `API_BASE_URL` para que
 * el navegador las pida directamente al backend.
 *
 * Las URLs ya absolutas (`http://`, `https://`, `data:`…) se dejan intactas;
 * `undefined`/`''` se propagan tal cual (el componente decide el respaldo).
 */
export function mediaUrl(path: string | undefined): string | undefined {
  if (!path || ABSOLUTE_URL_PATTERN.test(path)) {
    return path;
  }
  return `${API_BASE_URL}${path}`;
}

/** Timeout de la consulta: 5 s con `AbortController` (D20). */
export const REQUEST_TIMEOUT_MS = 5000;

/** Nombre de la cookie *double-submit* firmada que emite el backend (P10). */
export const CSRF_COOKIE_NAME = 'csrf_token';

/** Cabecera que acompaña a todo método no seguro con sesión (P10). */
export const CSRF_HEADER_NAME = 'X-CSRF-Token';

/**
 * Métodos que no modifican estado: para ellos no se exige CSRF. El resto
 * (POST, PUT, PATCH, DELETE) es "no seguro" y lleva la cabecera si hay cookie.
 */
const SAFE_METHODS = new Set(['GET', 'HEAD', 'OPTIONS', 'TRACE']);

export type ApiErrorReason = 'http' | 'network' | 'timeout' | 'invalid-response';

interface ApiErrorInit {
  status?: number;
  code?: string;
  details?: Record<string, unknown>;
  cause?: unknown;
}

/**
 * Error único del cliente HTTP. `reason` distingue el error del sobre (4xx/5xx
 * con `ErrorEnvelope`) de las situaciones sin respuesta interpretable (red,
 * timeout o respuesta no conforme), que la feature `status` mapea al estado
 * *inaccesible* de `ux.md`. Nunca pierde `code`/`message`/`details`.
 */
export class ApiError extends Error {
  readonly reason: ApiErrorReason;
  readonly status?: number;
  readonly code?: string;
  readonly details?: Record<string, unknown>;

  constructor(reason: ApiErrorReason, message: string, init: ApiErrorInit = {}) {
    super(message, { cause: init.cause });
    this.name = 'ApiError';
    this.reason = reason;
    this.status = init.status;
    this.code = init.code;
    this.details = init.details;
  }

  static fromEnvelope(status: number, body: ErrorBody): ApiError {
    return new ApiError('http', body.message, {
      status,
      code: body.code,
      details: body.details,
    });
  }

  static network(cause?: unknown): ApiError {
    return new ApiError('network', 'No hubo respuesta del servidor', { cause });
  }

  static timeout(cause?: unknown): ApiError {
    return new ApiError('timeout', 'La consulta superó el tiempo de espera', { cause });
  }

  static invalidResponse(cause?: unknown): ApiError {
    return new ApiError('invalid-response', 'La respuesta del servidor no es válida', { cause });
  }
}

export type ApiFetchOptions = RequestInit & {
  /** Tiempo máximo de espera en milisegundos. Por defecto, 5 s (D20). */
  timeoutMs?: number;
};

/**
 * Construye una URL con sus parámetros de consulta, omitiendo los que no están
 * definidos (o están vacíos). Sin parámetros devuelve el `path` tal cual.
 */
export function buildUrl(
  path: string,
  params: Record<string, string | number | undefined | null> = {},
): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== null && value !== '') {
      search.set(key, String(value));
    }
  }
  const query = search.toString();
  return query ? `${path}?${query}` : path;
}

/**
 * Lee una cookie del documento por nombre (P10). Devuelve `null` si no existe
 * o si no hay `document` (entornos sin navegador). No accede a `localStorage`:
 * la sesión vive en cookies `HttpOnly` gestionadas por el backend.
 */
export function readCookie(name: string): string | null {
  if (typeof document === 'undefined') {
    return null;
  }
  const prefix = `${name}=`;
  for (const part of document.cookie.split(';')) {
    const trimmed = part.trim();
    if (trimmed.startsWith(prefix)) {
      return decodeURIComponent(trimmed.slice(prefix.length));
    }
  }
  return null;
}

export async function apiFetch<T>(path: string, options: ApiFetchOptions = {}): Promise<T> {
  const { timeoutMs = REQUEST_TIMEOUT_MS, headers, method, ...rest } = options;
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

  const httpMethod = (method ?? 'GET').toUpperCase();
  const requestHeaders = new Headers(headers);
  if (!requestHeaders.has('Accept')) {
    requestHeaders.set('Accept', 'application/json');
  }
  if (!SAFE_METHODS.has(httpMethod)) {
    const csrfToken = readCookie(CSRF_COOKIE_NAME);
    if (csrfToken !== null && !requestHeaders.has(CSRF_HEADER_NAME)) {
      requestHeaders.set(CSRF_HEADER_NAME, csrfToken);
    }
  }

  try {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      ...rest,
      method: httpMethod,
      credentials: 'include',
      headers: requestHeaders,
      signal: controller.signal,
    });

    const body = await readJsonBody(response);

    if (response.ok) {
      return body as T;
    }

    const errorBody = parseErrorEnvelope(body);
    if (!errorBody) {
      throw ApiError.invalidResponse();
    }

    throw ApiError.fromEnvelope(response.status, errorBody);
  } catch (cause) {
    if (cause instanceof ApiError) {
      throw cause;
    }
    if (isAbortError(cause)) {
      throw ApiError.timeout(cause);
    }
    throw ApiError.network(cause);
  } finally {
    clearTimeout(timeoutId);
  }
}

async function readJsonBody(response: Response): Promise<unknown> {
  const text = await response.text();
  if (text.trim() === '') {
    return undefined;
  }
  try {
    return JSON.parse(text);
  } catch (cause) {
    throw ApiError.invalidResponse(cause);
  }
}

function parseErrorEnvelope(body: unknown): ErrorBody | undefined {
  if (typeof body !== 'object' || body === null) {
    return undefined;
  }

  const error = (body as { error?: unknown }).error;
  if (typeof error !== 'object' || error === null) {
    return undefined;
  }

  const { code, message, details } = error as Record<string, unknown>;
  if (typeof code !== 'string' || typeof message !== 'string') {
    return undefined;
  }

  if (details === undefined) {
    return { code: code as ErrorBody['code'], message };
  }

  if (typeof details !== 'object' || details === null || Array.isArray(details)) {
    return undefined;
  }

  return { code: code as ErrorBody['code'], message, details: details as Record<string, unknown> };
}

function isAbortError(value: unknown): boolean {
  return (
    typeof value === 'object' &&
    value !== null &&
    'name' in value &&
    (value as { name?: unknown }).name === 'AbortError'
  );
}
