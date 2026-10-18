import { describe, expect, it } from 'vitest';
import {
  ACTIVIDADES,
  ADMIN_USERS_ROLES,
  CATEGORIAS,
  DONACIONES,
  EVENTOS,
  MEDIOS,
  MINISTERIOS,
  PORTADA,
} from '../../lib/permissions';
import { isPermissionAvailable } from './permissions';

describe('isPermissionAvailable', () => {
  it('reconoce los módulos ya construidos: usuarios/roles, portada, agenda, ministerios, donaciones y medios (FR-012/FR-011/FR-021)', () => {
    expect(isPermissionAvailable(ADMIN_USERS_ROLES)).toBe(true);
    expect(isPermissionAvailable(PORTADA)).toBe(true);
    expect(isPermissionAvailable(EVENTOS)).toBe(true);
    expect(isPermissionAvailable(ACTIVIDADES)).toBe(true);
    expect(isPermissionAvailable(CATEGORIAS)).toBe(true);
    expect(isPermissionAvailable(MINISTERIOS)).toBe(true);
    // F7 activa `donaciones` (permiso ya sembrado por F2): deja de llevar el
    // sello «Disponible más adelante» (cambio acotado RG7-6, FR-011).
    expect(isPermissionAvailable(DONACIONES)).toBe(true);
    // F9 activa `medios` (permiso ya sembrado por F2, el catálogo sigue en 10):
    // también deja de llevar el sello (cambio acotado RG9-6, FR-021).
    expect(isPermissionAvailable(MEDIOS)).toBe(true);
  });

  it('mantiene el resto del catálogo (F5, F8) como «Disponible más adelante»', () => {
    expect(isPermissionAvailable('noticias')).toBe(false);
  });
});
