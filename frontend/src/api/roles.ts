import { apiFetch, buildUrl } from './client';
import type { PageParams } from './usuarios';
import type { components } from './schema';

export type RoleItem = components['schemas']['RoleItem'];
export type RoleList = components['schemas']['RoleList'];
export type RoleCreateInput = components['schemas']['RoleCreateInput'];
export type RoleUpdateInput = components['schemas']['RoleUpdateInput'];
export type RoleDeletedResponse = components['schemas']['RoleDeletedResponse'];
export type PermissionItem = components['schemas']['PermissionItem'];
export type PermissionList = components['schemas']['PermissionList'];

const JSON_HEADERS = { 'Content-Type': 'application/json' } as const;

/** Listado paginado de roles con sus permisos y cuentas asignadas (FR-017). */
export async function listRoles(params: PageParams = {}): Promise<RoleList> {
  return apiFetch<RoleList>(
    buildUrl('/api/v1/admin/roles', { limit: params.limit, offset: params.offset }),
  );
}

/** Detalle de un rol para el formulario de edición. */
export async function getRole(id: string): Promise<RoleItem> {
  return apiFetch<RoleItem>(`/api/v1/admin/roles/${encodeURIComponent(id)}`);
}

/** Crea un rol con al menos un permiso (FR-014). Requiere CSRF. */
export async function createRole(input: RoleCreateInput): Promise<RoleItem> {
  return apiFetch<RoleItem>('/api/v1/admin/roles', {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Edita nombre y/o permisos de un rol (FR-017/FR-018). Requiere CSRF. */
export async function updateRole(id: string, input: RoleUpdateInput): Promise<RoleItem> {
  return apiFetch<RoleItem>(`/api/v1/admin/roles/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Elimina un rol sin uso (FR-017). Requiere CSRF. */
export async function deleteRole(id: string): Promise<RoleDeletedResponse> {
  return apiFetch<RoleDeletedResponse>(`/api/v1/admin/roles/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  });
}

/** Catálogo fijo de permisos por módulo (FR-015). */
export async function listPermissions(): Promise<PermissionList> {
  return apiFetch<PermissionList>('/api/v1/admin/permisos');
}
