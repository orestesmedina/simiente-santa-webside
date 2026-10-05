import { useCallback, useState } from 'react';
import { ApiError } from '../../../api/client';
import type { RoleItem } from '../../../api/roles';
import { Button } from '../../../components/Button';
import { ConfirmDialog } from '../../../components/ConfirmDialog';
import { Dialog } from '../../../components/Dialog';
import { EmptyState } from '../../../components/EmptyState';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { Pagination } from '../../../components/Pagination';
import { Table, type Column } from '../../../components/Table';
import { RoleForm, type RoleFormMode } from '../components/RoleForm';
import { useDeleteRole } from '../hooks/useDeleteRole';
import { usePermissions } from '../hooks/usePermissions';
import { ROLES_PAGE_SIZE, useRoles } from '../hooks/useRoles';
import {
  CHANGES_SAVED,
  EMPTY_ROLES_MESSAGE,
  LOAD_ROLES_ERROR,
  ROLE_CREATED,
  ROLE_DELETED,
  SESSION_EXPIRED_ERROR,
  SYSTEM_ERROR,
  deleteConfirmDescription,
  roleInUseMessage,
} from '../messages';
import { PENDING_PERMISSION_NOTE, isPermissionAvailable, permissionLabel } from '../permissions';

interface RoleNotice {
  variant: NoticeVariant;
  message: string;
}

function describeDeleteError(error: ApiError, role: RoleItem): RoleNotice {
  if (error.reason === 'http' && error.status === 409) {
    return { variant: 'warning', message: error.message || roleInUseMessage(role.userCount) };
  }
  if (error.reason === 'http' && error.status === 401) {
    return { variant: 'error', message: SESSION_EXPIRED_ERROR };
  }
  return { variant: 'error', message: SYSTEM_ERROR };
}

function pluralCuentas(count: number): string {
  return count === 1 ? '1 cuenta' : `${count} cuentas`;
}

function RolePermissionSummary({ permissions }: { permissions: string[] }) {
  const permissionsQuery = usePermissions();
  const catalog = permissionsQuery.data?.items;

  if (permissions.length === 0) {
    return <span className="text-slate-600">Sin permisos</span>;
  }

  return (
    <ul className="flex flex-wrap gap-1">
      {permissions.map((code) => (
        <li
          key={code}
          className="inline-flex flex-wrap items-center gap-1 rounded-full border border-slate-300 bg-slate-50 px-2 py-0.5 text-sm text-slate-900"
        >
          {permissionLabel(code, catalog)}
          {!isPermissionAvailable(code) && (
            <span className="text-slate-700">· {PENDING_PERMISSION_NOTE}</span>
          )}
        </li>
      ))}
    </ul>
  );
}

/**
 * Listado y formulario de roles (ux.md §3.6/§3.7, FR-014…FR-018): nombre,
 * resumen de permisos con sello "disponible más adelante" en los módulos de
 * F3–F9 y número de cuentas asignadas. Editar abre la ficha; eliminar pide
 * confirmación **solo si el rol no está en uso** y, si lo está, explica que
 * primero hay que reasignar sin llegar a llamar a `DELETE` (FR-017/US6).
 */
export function RolesPage() {
  const [page, setPage] = useState(1);
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<RoleItem | null>(null);
  const [pendingDelete, setPendingDelete] = useState<RoleItem | null>(null);
  const [notice, setNotice] = useState<RoleNotice | null>(null);

  const rolesQuery = useRoles({ page });
  const deleteRole = useDeleteRole();

  const closeCreate = useCallback(() => setCreateOpen(false), []);
  const closeEdit = useCallback(() => setEditing(null), []);
  const closeDelete = useCallback(() => setPendingDelete(null), []);

  const handleFormSuccess = useCallback((_role: RoleItem, mode: RoleFormMode) => {
    setCreateOpen(false);
    setEditing(null);
    setNotice({ variant: 'success', message: mode === 'create' ? ROLE_CREATED : CHANGES_SAVED });
    if (mode === 'create') {
      setPage(1);
    }
  }, []);

  const handleDelete = () => {
    if (!pendingDelete) {
      return;
    }
    const role = pendingDelete;
    deleteRole.mutate(
      { id: role.id },
      {
        onSuccess: () => {
          setNotice({ variant: 'success', message: ROLE_DELETED });
          setPendingDelete(null);
        },
        onError: (error) => {
          setNotice(describeDeleteError(error, role));
          setPendingDelete(null);
        },
      },
    );
  };

  const data = rolesQuery.data;
  const limit = data?.limit ?? ROLES_PAGE_SIZE;
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const firstShown = total === 0 ? 0 : (page - 1) * limit + 1;
  const lastShown = Math.min(page * limit, total);
  const pendingDeleteId = deleteRole.isPending ? deleteRole.variables?.id : undefined;

  const columns: Column<RoleItem>[] = [
    { key: 'name', header: 'Nombre', render: (role) => role.name },
    {
      key: 'permissions',
      header: 'Permisos',
      render: (role) => <RolePermissionSummary permissions={role.permissions} />,
    },
    {
      key: 'userCount',
      header: 'Cuentas',
      render: (role) => pluralCuentas(role.userCount),
    },
    {
      key: 'actions',
      header: 'Acciones',
      render: (role) => (
        <div className="flex flex-wrap gap-2">
          <Button size="sm" variant="secondary" onClick={() => setEditing(role)}>
            Editar
          </Button>
          <Button
            size="sm"
            variant="secondary"
            disabled={pendingDeleteId === role.id}
            onClick={() => setPendingDelete(role)}
          >
            Eliminar
          </Button>
        </div>
      ),
    },
  ];

  const hasItems = (data?.items.length ?? 0) > 0;

  return (
    <section className="space-y-4">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Roles</h1>
          {data && <p className="text-slate-700">{total === 1 ? '1 rol' : `${total} roles`}</p>}
        </div>
        <Button onClick={() => setCreateOpen(true)}>Crear rol</Button>
      </header>

      {notice && (
        <Notice variant={notice.variant} temporal>
          {notice.message}
        </Notice>
      )}

      {rolesQuery.isPending && <EmptyState kind="loading" message="Cargando roles…" />}

      {rolesQuery.isError && (
        <EmptyState
          kind="error"
          message={LOAD_ROLES_ERROR}
          action={<Button onClick={() => void rolesQuery.refetch()}>Reintentar</Button>}
        />
      )}

      {rolesQuery.isSuccess && !hasItems && (
        <EmptyState
          message={EMPTY_ROLES_MESSAGE}
          action={<Button onClick={() => setCreateOpen(true)}>Crear rol</Button>}
        />
      )}

      {rolesQuery.isSuccess && hasItems && (
        <div>
          <Table
            columns={columns}
            rows={data?.items ?? []}
            rowKey={(role) => role.id}
            caption="Listado de roles del panel"
          />
          <Pagination
            page={page}
            totalPages={totalPages}
            onChange={setPage}
            summaryText={`Mostrando ${firstShown}–${lastShown} de ${total} roles`}
          />
        </div>
      )}

      {createOpen && (
        <Dialog title="Crear rol" onClose={closeCreate}>
          <RoleForm mode="create" onSuccess={handleFormSuccess} onCancel={closeCreate} />
        </Dialog>
      )}

      {editing && (
        <Dialog title="Editar rol" onClose={closeEdit}>
          <RoleForm mode="edit" role={editing} onSuccess={handleFormSuccess} onCancel={closeEdit} />
        </Dialog>
      )}

      {pendingDelete && pendingDelete.userCount > 0 && (
        <Dialog title="No se puede eliminar" onClose={closeDelete}>
          <Notice variant="warning">{roleInUseMessage(pendingDelete.userCount)}</Notice>
          <div className="mt-4 flex justify-end">
            <Button variant="secondary" onClick={closeDelete}>
              Entendido
            </Button>
          </div>
        </Dialog>
      )}

      {pendingDelete && pendingDelete.userCount === 0 && (
        <ConfirmDialog
          title="Eliminar rol"
          description={deleteConfirmDescription(
            pendingDelete.name,
            pendingDelete.permissions.length,
          )}
          confirmText="Eliminar"
          danger
          loading={deleteRole.isPending}
          onConfirm={handleDelete}
          onClose={closeDelete}
        />
      )}
    </section>
  );
}
