import { describe, expect, it } from 'vitest';
import { ADMIN_USERS_ROLES, PORTADA } from '../../lib/permissions';
import { isPermissionAvailable } from './permissions';

describe('isPermissionAvailable', () => {
  it('reconoce los módulos ya construidos: usuarios/roles y portada (F3, FR-012)', () => {
    expect(isPermissionAvailable(ADMIN_USERS_ROLES)).toBe(true);
    expect(isPermissionAvailable(PORTADA)).toBe(true);
  });

  it('mantiene el resto del catálogo (F4–F9) como «Disponible más adelante»', () => {
    expect(isPermissionAvailable('eventos')).toBe(false);
    expect(isPermissionAvailable('noticias')).toBe(false);
  });
});
