import type { components } from './schema';

export type SystemStatus = components['schemas']['SystemStatus'];
export type ErrorEnvelope = components['schemas']['ErrorEnvelope'];
export type ErrorBody = components['schemas']['ErrorBody'];

/** URL base de la API. Se hornea en build time (plan R6). */
export const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

/** Timeout de la consulta: 5 s con `AbortController` (D20). */
export const REQUEST_TIMEOUT_MS = 5000;

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

export async function apiFetch<T>(path: string, options: ApiFetchOptions = {}): Promise<T> {
  const { timeoutMs = REQUEST_TIMEOUT_MS, headers, ...rest } = options;
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      ...rest,
      headers: { Accept: 'application/json', ...headers },
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
