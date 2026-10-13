import { describe, expect, it } from 'vitest';
import {
  ACTIVIDADES,
  ADMIN_USERS_ROLES,
  CATEGORIAS,
  EVENTOS,
  MINISTERIOS,
  PORTADA,
} from '../../lib/permissions';
import { isPermissionAvailable } from './permissions';

describe('isPermissionAvailable', () => {
  it('reconoce los módulos ya construidos: usuarios/roles, portada, agenda y ministerios (FR-012/FR-011)', () => {
    expect(isPermissionAvailable(ADMIN_USERS_ROLES)).toBe(true);
    expect(isPermissionAvailable(PORTADA)).toBe(true);
    expect(isPermissionAvailable(EVENTOS)).toBe(true);
    expect(isPermissionAvailable(ACTIVIDADES)).toBe(true);
    expect(isPermissionAvailable(CATEGORIAS)).toBe(true);
    expect(isPermissionAvailable(MINISTERIOS)).toBe(true);
  });

  it('mantiene el resto del catálogo (F5–F9) como «Disponible más adelante»', () => {
    expect(isPermissionAvailable('noticias')).toBe(false);
    expect(isPermissionAvailable('donaciones')).toBe(false);
  });
});
