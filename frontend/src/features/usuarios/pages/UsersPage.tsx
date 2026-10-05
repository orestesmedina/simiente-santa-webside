import { useCallback, useState } from 'react';
import { ApiError } from '../../../api/client';
import type { UserItem } from '../../../api/usuarios';
import { Button } from '../../../components/Button';
import { ConfirmDialog } from '../../../components/ConfirmDialog';
import { Dialog } from '../../../components/Dialog';
import { EmptyState } from '../../../components/EmptyState';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { Pagination } from '../../../components/Pagination';
import { StatusPill } from '../../../components/StatusPill';
import { Table, type Column } from '../../../components/Table';
import { formatLastLogin } from '../../../lib/format';
import { ResetPasswordDialog } from '../components/ResetPasswordDialog';
import { UserForm, type UserFormMode } from '../components/UserForm';
import { useSetUserActive } from '../hooks/useSetUserActive';
import { USERS_PAGE_SIZE, useUsers } from '../hooks/useUsers';
import {
  ACTIVATE_CONFIRM_DESCRIPTION,
  CAMBIOS_GUARDADOS,
  CUENTA_ACTIVADA,
  CUENTA_CREADA,
  CUENTA_DESACTIVADA,
  DEACTIVATE_CONFIRM_DESCRIPTION,
  EMPTY_USERS_MESSAGE,
  LAST_ADMIN_ERROR,
  LOAD_USERS_ERROR,
  PASSWORD_RESET,
  SESSION_EXPIRED_ERROR,
  SYSTEM_ERROR,
} from '../messages';

interface UserNotice {
  variant: NoticeVariant;
  message: string;
}

interface PendingToggle {
  user: UserItem;
  nextActive: boolean;
}

/** Traduce el fallo de una acción de fila a los avisos de ux.md §4.b/§7. */
function describeActionError(error: ApiError): UserNotice {
  if (
    error.reason === 'http' &&
    error.status === 409 &&
    error.details?.reason === 'admin_required'
  ) {
    return { variant: 'warning', message: LAST_ADMIN_ERROR };
  }
  if (error.reason === 'http' && error.status === 401) {
    return { variant: 'error', message: SESSION_EXPIRED_ERROR };
  }
  return { variant: 'error', message: SYSTEM_ERROR };
}

/**
 * Listado de cuentas (ux.md §3.4, FR-019/FR-021): estado, correo, rol y último
 * acceso por fila; tarjetas en móvil y tabla desde tableta con `Table`;
 * paginación `limit`/`offset` sin buscador de texto libre (F-04, fuera del
 * MVP). La ficha de crear/editar, la activación/desactivación confirmada y el
 * restablecimiento de contraseña viven en los diálogos de T248.
 */
export function UsersPage() {
  const [page, setPage] = useState(1);
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<UserItem | null>(null);
  const [resetting, setResetting] = useState<UserItem | null>(null);
  const [confirming, setConfirming] = useState<PendingToggle | null>(null);
  const [notice, setNotice] = useState<UserNotice | null>(null);

  const usersQuery = useUsers({ page });
  const setUserActive = useSetUserActive();

  const closeCreate = useCallback(() => setCreateOpen(false), []);
  const closeEdit = useCallback(() => setEditing(null), []);
  const closeReset = useCallback(() => setResetting(null), []);
  const closeConfirm = useCallback(() => setConfirming(null), []);

  const handleFormSuccess = useCallback((_user: UserItem, mode: UserFormMode) => {
    setCreateOpen(false);
    setEditing(null);
    setNotice({
      variant: 'success',
      message: mode === 'create' ? CUENTA_CREADA : CAMBIOS_GUARDADOS,
    });
    if (mode === 'create') {
      setPage(1);
    }
  }, []);

  const handleResetSuccess = useCallback(() => {
    setResetting(null);
    setNotice({ variant: 'success', message: PASSWORD_RESET });
  }, []);

  const handleRequestReset = useCallback((user: UserItem) => {
    setEditing(null);
    setResetting(user);
  }, []);

  const handleConfirmToggle = () => {
    if (!confirming) {
      return;
    }
    const { user, nextActive } = confirming;
    setUserActive.mutate(
      { user, isActive: nextActive },
      {
        onSuccess: () => {
          setNotice({
            variant: 'success',
            message: nextActive ? CUENTA_ACTIVADA : CUENTA_DESACTIVADA,
          });
          setConfirming(null);
        },
        onError: (error) => {
          setNotice(describeActionError(error));
          setConfirming(null);
        },
      },
    );
  };

  const data = usersQuery.data;
  const limit = data?.limit ?? USERS_PAGE_SIZE;
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const firstShown = total === 0 ? 0 : (page - 1) * limit + 1;
  const lastShown = Math.min(page * limit, total);
  const pendingUserId = setUserActive.isPending ? setUserActive.variables?.user.id : undefined;

  const columns: Column<UserItem>[] = [
    {
      key: 'name',
      header: 'Nombre',
      render: (user) => `${user.firstName} ${user.lastName}`,
    },
    { key: 'email', header: 'Correo', render: (user) => user.email },
    { key: 'role', header: 'Rol', render: (user) => user.roleName },
    {
      key: 'status',
      header: 'Estado',
      render: (user) => <StatusPill value={user.isActive ? 'active' : 'inactive'} />,
    },
    {
      key: 'lastLogin',
      header: 'Último acceso',
      render: (user) => formatLastLogin(user.lastLoginAt, user.lastLoginIp),
    },
    {
      key: 'actions',
      header: 'Acciones',
      render: (user) => (
        <div className="flex flex-wrap gap-2">
          <Button size="sm" variant="secondary" onClick={() => setEditing(user)}>
            Editar
          </Button>
          <Button
            size="sm"
            variant="secondary"
            disabled={pendingUserId === user.id}
            onClick={() => setConfirming({ user, nextActive: !user.isActive })}
          >
            {user.isActive ? 'Desactivar' : 'Activar'}
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
          <h1 className="text-2xl font-bold text-slate-900">Usuarios</h1>
          {data && (
            <p className="text-slate-700">{total === 1 ? '1 usuario' : `${total} usuarios`}</p>
          )}
        </div>
        <Button onClick={() => setCreateOpen(true)}>Crear usuario</Button>
      </header>

      {notice && (
        <Notice variant={notice.variant} temporal>
          {notice.message}
        </Notice>
      )}

      {usersQuery.isPending && <EmptyState kind="loading" message="Cargando usuarios…" />}

      {usersQuery.isError && (
        <EmptyState
          kind="error"
          message={LOAD_USERS_ERROR}
          action={<Button onClick={() => void usersQuery.refetch()}>Reintentar</Button>}
        />
      )}

      {usersQuery.isSuccess && !hasItems && (
        <EmptyState
          message={EMPTY_USERS_MESSAGE}
          action={<Button onClick={() => setCreateOpen(true)}>Crear usuario</Button>}
        />
      )}

      {usersQuery.isSuccess && hasItems && (
        <div>
          <Table
            columns={columns}
            rows={data?.items ?? []}
            rowKey={(user) => user.id}
            caption="Listado de cuentas del panel"
          />
          <Pagination
            page={page}
            totalPages={totalPages}
            onChange={setPage}
            summaryText={`Mostrando ${firstShown}–${lastShown} de ${total} usuarios`}
          />
        </div>
      )}

      {createOpen && (
        <Dialog title="Crear usuario" onClose={closeCreate}>
          <UserForm mode="create" onSuccess={handleFormSuccess} onCancel={closeCreate} />
        </Dialog>
      )}

      {editing && (
        <Dialog title="Editar cuenta" onClose={closeEdit}>
          <UserForm
            mode="edit"
            user={editing}
            onSuccess={handleFormSuccess}
            onCancel={closeEdit}
            onRequestResetPassword={handleRequestReset}
          />
        </Dialog>
      )}

      {resetting && (
        <ResetPasswordDialog user={resetting} onClose={closeReset} onSuccess={handleResetSuccess} />
      )}

      {confirming && (
        <ConfirmDialog
          title={confirming.nextActive ? 'Activar cuenta' : 'Desactivar cuenta'}
          description={
            confirming.nextActive ? ACTIVATE_CONFIRM_DESCRIPTION : DEACTIVATE_CONFIRM_DESCRIPTION
          }
          confirmText={confirming.nextActive ? 'Activar' : 'Desactivar'}
          danger={!confirming.nextActive}
          loading={setUserActive.isPending}
          onConfirm={handleConfirmToggle}
          onClose={closeConfirm}
        />
      )}
    </section>
  );
}
