import { describe, expect, it } from 'vitest';
import {
  ACTIVIDADES,
  ADMIN_USERS_ROLES,
  CATEGORIAS,
  EVENTOS,
  MINISTERIOS,
  MODULE_PERMISSIONS,
  PORTADA,
  hasAnyPermission,
  hasPermission,
} from './permissions';

describe('hasPermission', () => {
  it('devuelve true cuando la sesión tiene el permiso', () => {
    expect(hasPermission({ permissions: [ADMIN_USERS_ROLES] }, ADMIN_USERS_ROLES)).toBe(true);
  });

  it('devuelve true cuando la sesión tiene el permiso portada (F3)', () => {
    expect(hasPermission({ permissions: [PORTADA] }, PORTADA)).toBe(true);
    expect(hasPermission({ permissions: [PORTADA] }, ADMIN_USERS_ROLES)).toBe(false);
  });

  it('reconoce los tres permisos de F4 (FR-012)', () => {
    expect(hasPermission({ permissions: [EVENTOS] }, EVENTOS)).toBe(true);
    expect(hasPermission({ permissions: [ACTIVIDADES] }, ACTIVIDADES)).toBe(true);
    expect(hasPermission({ permissions: [CATEGORIAS] }, CATEGORIAS)).toBe(true);
  });

  it('los tres módulos de F4 son independientes entre sí (FR-012)', () => {
    const soloEventos = { permissions: [EVENTOS] };
    expect(hasPermission(soloEventos, ACTIVIDADES)).toBe(false);
    expect(hasPermission(soloEventos, CATEGORIAS)).toBe(false);

    const soloCategorias = { permissions: [CATEGORIAS] };
    expect(hasPermission(soloCategorias, EVENTOS)).toBe(false);
    expect(hasPermission(soloCategorias, ACTIVIDADES)).toBe(false);
  });

  it('devuelve false cuando la sesión no tiene el permiso', () => {
    expect(hasPermission({ permissions: ['noticias'] }, ADMIN_USERS_ROLES)).toBe(false);
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
    expect(hasAnyPermission({ permissions: ['noticias'] }, ['roles', 'noticias'])).toBe(true);
  });

  it('devuelve false si no coincide ninguno', () => {
    expect(hasAnyPermission({ permissions: ['noticias'] }, ['roles', 'donaciones'])).toBe(false);
  });
});

describe('MODULE_PERMISSIONS', () => {
  it('reúne los permisos de módulo construidos (F2, F3, F4 y F6)', () => {
    expect(MODULE_PERMISSIONS).toEqual([
      ADMIN_USERS_ROLES,
      PORTADA,
      EVENTOS,
      ACTIVIDADES,
      CATEGORIAS,
      MINISTERIOS,
    ]);
  });
});
