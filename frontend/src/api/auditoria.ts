import { apiFetch, buildUrl } from './client';
import type { PageParams } from './usuarios';
import type { components } from './schema';

export type AccessEventItem = components['schemas']['AccessEventItem'];
export type AccessEventList = components['schemas']['AccessEventList'];
export type AdminActionItem = components['schemas']['AdminActionItem'];
export type AdminActionList = components['schemas']['AdminActionList'];

/**
 * Filtros comunes de los historiales de auditoría (FR-024): cuenta y rango de
 * fechas `[from, to)` en ISO-8601, más paginación.
 */
export interface AuditParams extends PageParams {
  userId?: string;
  from?: string;
  to?: string;
}

/**
 * Historial de intentos de inicio de sesión (FR-022). **Solo lectura**
 * (FR-025): este módulo no expone ninguna operación de escritura.
 */
export async function listAccessEvents(params: AuditParams = {}): Promise<AccessEventList> {
  return apiFetch<AccessEventList>(
    buildUrl('/api/v1/admin/auditoria/accesos', {
      userId: params.userId,
      from: params.from,
      to: params.to,
      limit: params.limit,
      offset: params.offset,
    }),
  );
}

/**
 * Historial de acciones administrativas (FR-023). **Solo lectura** (FR-025):
 * persistir, editar o borrar registros no está disponible por ninguna función.
 */
export async function listAdminActions(params: AuditParams = {}): Promise<AdminActionList> {
  return apiFetch<AdminActionList>(
    buildUrl('/api/v1/admin/auditoria/acciones', {
      userId: params.userId,
      from: params.from,
      to: params.to,
      limit: params.limit,
      offset: params.offset,
    }),
  );
}
