import { render, screen, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import type { SessionUser } from '../../../api/auth';
import { ADMIN_USERS_ROLES, MINISTERIOS } from '../../../lib/permissions';
import { AccesosRapidos } from './AccesosRapidos';

/**
 * Accesos rápidos del Inicio del panel (ux.md §4.2): la tarjeta «Gestión de
 * ministerios» aparece **solo** con el permiso `ministerios` (FR-011), sin
 * mezclarse con los demás módulos.
 */

function makeSession(permissions: string[]): SessionUser {
  return {
    id: '11111111-1111-1111-1111-111111111111',
    email: 'ana@ejemplo.com',
    firstName: 'Ana',
    lastName: 'Pérez',
    phone: '612345678',
    roleId: '22222222-2222-2222-2222-222222222222',
    roleName: 'Administración',
    permissions,
    mustChangePassword: false,
  };
}

function renderAccesos(permissions: string[]) {
  return render(
    <MemoryRouter initialEntries={['/panel']}>
      <AccesosRapidos session={makeSession(permissions)} />
    </MemoryRouter>,
  );
}

describe('AccesosRapidos · ministerios (FR-011)', () => {
  it('muestra la tarjeta «Gestión de ministerios» con el permiso `ministerios`', () => {
    renderAccesos([MINISTERIOS]);

    const nav = screen.getByRole('navigation', { name: 'Accesos rápidos' });
    expect(within(nav).getByRole('link', { name: 'Gestión de ministerios' })).toHaveAttribute(
      'href',
      '/panel/ministerios',
    );
    // Ni «ministerios» da acceso a los otros módulos ni al revés (Decisión 5).
    expect(within(nav).queryByRole('link', { name: 'Usuarios' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Gestión de eventos' })).toBeNull();
    expect(within(nav).queryByRole('link', { name: 'Portada e información general' })).toBeNull();
  });

  it('sin el permiso `ministerios` la tarjeta no aparece', () => {
    renderAccesos([ADMIN_USERS_ROLES]);

    const nav = screen.getByRole('navigation', { name: 'Accesos rápidos' });
    expect(within(nav).queryByRole('link', { name: 'Gestión de ministerios' })).toBeNull();
  });
});
