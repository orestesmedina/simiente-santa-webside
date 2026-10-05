import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../api/client';
import type { AccessEventItem, AdminActionItem } from '../../api/auditoria';
import type { UserItem } from '../../api/usuarios';
import { AppProviders } from '../../app/providers';
import { server } from '../../test/server';
import { NO_PERMISSION_ERROR } from './messages';
import { AuditPage } from './pages/AuditPage';

const accesosUrl = `${API_BASE_URL}/api/v1/admin/auditoria/accesos`;
const accionesUrl = `${API_BASE_URL}/api/v1/admin/auditoria/acciones`;
const usersUrl = `${API_BASE_URL}/api/v1/admin/usuarios`;

const anaId = '11111111-1111-1111-1111-111111111111';

function makeUser(overrides: Partial<UserItem> = {}): UserItem {
  return {
    id: anaId,
    email: 'ana@ejemplo.com',
    firstName: 'Ana',
    lastName: 'Pérez',
    phone: '612345678',
    roleId: '22222222-2222-2222-2222-222222222222',
    roleName: 'Administración',
    isActive: true,
    mustChangePassword: false,
    lastLoginAt: '2026-10-03T17:42:00Z',
    lastLoginIp: '189.2.4.15',
    createdAt: '2026-10-01T10:00:00Z',
    ...overrides,
  };
}

function makeAccess(overrides: Partial<AccessEventItem> = {}): AccessEventItem {
  return {
    id: '44444444-4444-4444-4444-444444444444',
    userId: anaId,
    userEmail: 'ana@ejemplo.com',
    userName: 'Ana Pérez',
    result: 'success',
    ip: '189.2.4.15',
    createdAt: '2026-10-03T17:42:00Z',
    ...overrides,
  };
}

function makeAction(overrides: Partial<AdminActionItem> = {}): AdminActionItem {
  return {
    id: '55555555-5555-5555-5555-555555555555',
    actorId: anaId,
    actorEmail: 'ana@ejemplo.com',
    actorName: 'Ana Pérez',
    action: 'user.deactivate',
    targetKind: 'user',
    targetId: '66666666-6666-6666-6666-666666666666',
    targetLabel: 'luis@ejemplo.com',
    result: 'success',
    createdAt: '2026-10-03T17:45:00Z',
    ...overrides,
  };
}

function useAccounts() {
  server.use(
    http.get(usersUrl, () =>
      HttpResponse.json({ items: [makeUser()], total: 1, limit: 100, offset: 0 }),
    ),
  );
}

function useAccessEvents(items: AccessEventItem[], total = items.length) {
  server.use(http.get(accesosUrl, () => HttpResponse.json({ items, total, limit: 20, offset: 0 })));
}

function useAdminActions(items: AdminActionItem[], total = items.length) {
  server.use(
    http.get(accionesUrl, () => HttpResponse.json({ items, total, limit: 20, offset: 0 })),
  );
}

function renderAudit() {
  return render(
    <AppProviders>
      <AuditPage />
    </AppProviders>,
  );
}

describe('AuditPage · dos historiales', () => {
  it('los accesos muestran cuenta, resultado e IP; sin cuenta → "Intento sin cuenta asociada" sin correo (F-01)', async () => {
    useAccounts();
    useAccessEvents([
      makeAccess(),
      makeAccess({
        id: '77777777-7777-7777-7777-777777777777',
        userId: null,
        userEmail: null,
        userName: null,
        result: 'failure',
        ip: '10.0.0.9',
      }),
    ]);

    renderAudit();
    const table = await screen.findByRole('table');

    expect(within(table).getByText('Ana Pérez')).toBeInTheDocument();
    expect(within(table).getByText('Exitoso')).toBeInTheDocument();
    expect(within(table).getByText('Fallido')).toBeInTheDocument();
    expect(within(table).getByText('189.2.4.15')).toBeInTheDocument();
    expect(within(table).getByText('Intento sin cuenta asociada')).toBeInTheDocument();
    // Nunca se muestra el correo del intento no identificado (FR-003/FR-026).
    expect(within(table).queryByText(/@/)).toBeNull();
  });

  it('las acciones muestran quién, qué, sobre qué y el resultado (FR-023)', async () => {
    useAccounts();
    useAdminActions([
      makeAction(),
      makeAction({
        id: '88888888-8888-8888-8888-888888888888',
        action: 'role.delete',
        targetKind: 'role',
        targetLabel: 'Editores',
        result: 'denied',
      }),
    ]);

    renderAudit();
    await screen.findByRole('heading', { name: 'Auditoría' });
    await userEvent.click(screen.getByRole('button', { name: 'Acciones administrativas' }));

    const table = await screen.findByRole('table');
    expect(within(table).getAllByText('Ana Pérez').length).toBeGreaterThan(0);
    expect(within(table).getByText('Desactivó una cuenta')).toBeInTheDocument();
    expect(within(table).getByText('luis@ejemplo.com')).toBeInTheDocument();
    expect(within(table).getByText('Eliminó un rol')).toBeInTheDocument();
    expect(within(table).getByText('Denegada')).toBeInTheDocument();
  });
});

describe('AuditPage · filtros y paginación', () => {
  it('envía cuenta y rango de fechas a la API y la paginación conserva los filtros (FR-024)', async () => {
    const urls: string[] = [];
    server.use(
      http.get(usersUrl, () =>
        HttpResponse.json({ items: [makeUser()], total: 1, limit: 100, offset: 0 }),
      ),
      http.get(accesosUrl, ({ request }) => {
        urls.push(request.url);
        const offset = Number(new URL(request.url).searchParams.get('offset') ?? '0');
        return HttpResponse.json({ items: [makeAccess()], total: 45, limit: 20, offset });
      }),
    );

    renderAudit();
    await screen.findByRole('table');

    await userEvent.selectOptions(screen.getByLabelText('Cuenta'), anaId);
    fireEvent.change(screen.getByLabelText('Desde'), { target: { value: '2026-10-01' } });
    fireEvent.change(screen.getByLabelText('Hasta'), { target: { value: '2026-10-05' } });
    await userEvent.click(screen.getByRole('button', { name: 'Filtrar' }));

    await waitFor(() => {
      const filtered = urls.find(
        (url) =>
          url.includes(`userId=${anaId}`) &&
          url.includes('from=2026-10-01T00%3A00%3A00Z') &&
          url.includes('to=2026-10-06T00%3A00%3A00Z'),
      );
      expect(filtered).toBeDefined();
    });

    await userEvent.click(screen.getByRole('button', { name: 'Siguiente' }));

    await waitFor(() => {
      expect(urls.some((url) => url.includes('offset=20') && url.includes(`userId=${anaId}`))).toBe(
        true,
      );
    });
  });

  it('aplicar filtros reinicia a la página 1', async () => {
    const urls: string[] = [];
    server.use(
      http.get(usersUrl, () =>
        HttpResponse.json({ items: [makeUser()], total: 1, limit: 100, offset: 0 }),
      ),
      http.get(accesosUrl, ({ request }) => {
        urls.push(request.url);
        const offset = Number(new URL(request.url).searchParams.get('offset') ?? '0');
        return HttpResponse.json({ items: [makeAccess()], total: 45, limit: 20, offset });
      }),
    );

    renderAudit();
    await screen.findByRole('table');
    await userEvent.click(screen.getByRole('button', { name: 'Siguiente' }));
    await waitFor(() => expect(urls.some((url) => url.includes('offset=20'))).toBe(true));

    await userEvent.selectOptions(screen.getByLabelText('Cuenta'), anaId);
    await userEvent.click(screen.getByRole('button', { name: 'Filtrar' }));

    await waitFor(() => {
      expect(urls.some((url) => url.includes(`userId=${anaId}`) && url.includes('offset=0'))).toBe(
        true,
      );
    });
  });

  it('cada pestaña conserva sus filtros al alternar', async () => {
    useAccounts();
    useAccessEvents([makeAccess()]);
    useAdminActions([makeAction()]);

    renderAudit();
    await screen.findByRole('table');
    await userEvent.selectOptions(screen.getByLabelText('Cuenta'), anaId);
    await userEvent.click(screen.getByRole('button', { name: 'Filtrar' }));

    await userEvent.click(screen.getByRole('button', { name: 'Acciones administrativas' }));
    const table = await screen.findByRole('table');
    expect(within(table).getByText('Desactivó una cuenta')).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Historial de accesos' }));

    expect(await screen.findByLabelText('Cuenta')).toHaveValue(anaId);
  });
});

describe('AuditPage · estados', () => {
  it('sin registros muestra el estado vacío, con el texto propio si hay filtros', async () => {
    useAccounts();
    useAccessEvents([], 0);

    renderAudit();

    expect(
      await screen.findByText(
        'Todavía no hay registros de actividad. Aparecerán cuando alguien entre al panel o se gestione una cuenta o un rol.',
      ),
    ).toBeInTheDocument();

    await userEvent.selectOptions(screen.getByLabelText('Cuenta'), anaId);
    await userEvent.click(screen.getByRole('button', { name: 'Filtrar' }));

    expect(
      await screen.findByText(/Todavía no hay registros que coincidan con esos filtros/),
    ).toBeInTheDocument();
  });

  it('con 403 muestra un mensaje claro de sin permiso (FR-024)', async () => {
    useAccounts();
    server.use(
      http.get(accesosUrl, () =>
        HttpResponse.json(
          { error: { code: 'forbidden', message: 'Sin permiso' } },
          { status: 403 },
        ),
      ),
    );

    renderAudit();

    expect(await screen.findByText(NO_PERMISSION_ERROR)).toBeInTheDocument();
  });

  it('no ofrece ningún control de edición ni borrado (FR-025/SC-013)', async () => {
    useAccounts();
    useAccessEvents([makeAccess()]);

    renderAudit();
    const table = await screen.findByRole('table');

    expect(within(table).queryAllByRole('button')).toHaveLength(0);
    expect(screen.queryByRole('button', { name: /editar/i })).toBeNull();
    expect(screen.queryByRole('button', { name: /eliminar|borrar/i })).toBeNull();
  });
});
