import { apiFetch, buildUrl } from './client';
import type { components } from './schema';

export type UserItem = components['schemas']['UserItem'];
export type UserList = components['schemas']['UserList'];
export type CreateUserInput = components['schemas']['CreateUserInput'];
export type UpdateUserInput = components['schemas']['UpdateUserInput'];
export type ResetPasswordInput = components['schemas']['ResetPasswordInput'];
export type PasswordResetResponse = components['schemas']['PasswordResetResponse'];

export interface PageParams {
  limit?: number;
  offset?: number;
}

const JSON_HEADERS = { 'Content-Type': 'application/json' } as const;

/** Listado paginado de cuentas (FR-019). */
export async function listUsers(params: PageParams = {}): Promise<UserList> {
  return apiFetch<UserList>(
    buildUrl('/api/v1/admin/usuarios', { limit: params.limit, offset: params.offset }),
  );
}

/** Detalle de una cuenta para el formulario de edición. */
export async function getUser(id: string): Promise<UserItem> {
  return apiFetch<UserItem>(`/api/v1/admin/usuarios/${encodeURIComponent(id)}`);
}

/** Crea una cuenta con su contraseña inicial (FR-009/FR-010). Requiere CSRF. */
export async function createUser(input: CreateUserInput): Promise<UserItem> {
  return apiFetch<UserItem>('/api/v1/admin/usuarios', {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Edita datos, rol o estado de una cuenta (FR-011/FR-012). Requiere CSRF. */
export async function updateUser(id: string, input: UpdateUserInput): Promise<UserItem> {
  return apiFetch<UserItem>(`/api/v1/admin/usuarios/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Restablece la contraseña de una cuenta (FR-010). Requiere CSRF. */
export async function resetUserPassword(
  id: string,
  input: ResetPasswordInput,
): Promise<PasswordResetResponse> {
  return apiFetch<PasswordResetResponse>(
    `/api/v1/admin/usuarios/${encodeURIComponent(id)}/password`,
    {
      method: 'POST',
      headers: JSON_HEADERS,
      body: JSON.stringify(input),
    },
  );
}
