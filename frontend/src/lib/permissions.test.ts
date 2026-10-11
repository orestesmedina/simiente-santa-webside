import { describe, expect, it } from 'vitest';
import { ADMIN_USERS_ROLES, PORTADA, hasAnyPermission, hasPermission } from './permissions';

describe('hasPermission', () => {
  it('devuelve true cuando la sesión tiene el permiso', () => {
    expect(hasPermission({ permissions: [ADMIN_USERS_ROLES] }, ADMIN_USERS_ROLES)).toBe(true);
  });

  it('devuelve true cuando la sesión tiene el permiso portada (F3)', () => {
    expect(hasPermission({ permissions: [PORTADA] }, PORTADA)).toBe(true);
    expect(hasPermission({ permissions: [PORTADA] }, ADMIN_USERS_ROLES)).toBe(false);
  });

  it('devuelve false cuando la sesión no tiene el permiso', () => {
    expect(hasPermission({ permissions: ['eventos'] }, ADMIN_USERS_ROLES)).toBe(false);
  });

  it('devuelve false con cuenta sin permisos de módulo (Edge Case válido)', () => {
    expect(hasPermission({ permissions: [] }, ADMIN_USERS_ROLES)).toBe(false);
  });

  it('devuelve false sin sesión o con sesión nula', () => {
    expect(hasPermission(null, ADMIN_USERS_ROLES)).toBe(false);
    expect(hasPermission(undefined, ADMIN_USERS_ROLES)).toBe(false);
  });
});

describe('hasAnyPermission', () => {
  it('devuelve true si coincide alguno', () => {
    expect(hasAnyPermission({ permissions: ['eventos'] }, ['roles', 'eventos'])).toBe(true);
  });

  it('devuelve false si no coincide ninguno', () => {
    expect(hasAnyPermission({ permissions: ['eventos'] }, ['roles', 'donaciones'])).toBe(false);
  });
});
