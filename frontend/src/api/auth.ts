import { apiFetch } from './client';
import type { components } from './schema';

export type SessionUser = components['schemas']['SessionUser'];
export type LoginInput = components['schemas']['LoginInput'];
export type ChangePasswordInput = components['schemas']['ChangePasswordInput'];
export type LogoutResponse = components['schemas']['LogoutResponse'];
export type PasswordChangedResponse = components['schemas']['PasswordChangedResponse'];

const JSON_HEADERS = { 'Content-Type': 'application/json' } as const;

/**
 * Inicia sesión con correo y contraseña. El backend emite las cookies de sesión
 * (`HttpOnly`) y de CSRF; el cliente nunca guarda tokens. Lanza `ApiError` con
 * su `code` y `details` (p. ej. `access_disabled` o `rate_limited`).
 */
export async function login(input: LoginInput): Promise<SessionUser> {
  return apiFetch<SessionUser>('/api/v1/auth/login', {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Cierra la sesión actual (FR-004). Requiere CSRF. */
export async function logout(): Promise<LogoutResponse> {
  return apiFetch<LogoutResponse>('/api/v1/auth/logout', { method: 'POST' });
}

/** Identidad de la sesión iniciada con sus permisos por módulo (FR-015). */
export async function getSession(): Promise<SessionUser> {
  return apiFetch<SessionUser>('/api/v1/auth/session');
}

/** Cambia la contraseña de la propia cuenta (FR-020). Requiere CSRF. */
export async function changeMyPassword(
  input: ChangePasswordInput,
): Promise<PasswordChangedResponse> {
  return apiFetch<PasswordChangedResponse>('/api/v1/auth/password', {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}
