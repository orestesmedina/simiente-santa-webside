import { zodResolver } from '@hookform/resolvers/zod';
import { useEffect, useMemo } from 'react';
import { useForm, useWatch, type UseFormSetError } from 'react-hook-form';
import { z } from 'zod';
import { ApiError } from '../../../api/client';
import type { RoleItem } from '../../../api/roles';
import { Button } from '../../../components/Button';
import { EmptyState } from '../../../components/EmptyState';
import { Field } from '../../../components/Field';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { useCreateRole } from '../hooks/useCreateRole';
import { usePermissions } from '../hooks/usePermissions';
import { useUpdateRole } from '../hooks/useUpdateRole';
import {
  NO_PERMISSIONS_ERROR,
  ROLE_NAME_IN_USE,
  SESSION_EXPIRED_ERROR,
  SYSTEM_ERROR,
} from '../messages';
import { PENDING_PERMISSION_NOTE, isPermissionAvailable } from '../permissions';

export type RoleFormMode = 'create' | 'edit';

export interface RoleFormProps {
  mode: RoleFormMode;
  /** Rol que se edita; obligatorio en modo `edit`. */
  role?: RoleItem;
  onSuccess: (role: RoleItem, mode: RoleFormMode) => void;
  onCancel: () => void;
}

const roleSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, 'Escribe el nombre.')
    .max(80, 'El nombre no puede superar los 80 caracteres.'),
  permissions: z.array(z.string()).min(1, NO_PERMISSIONS_ERROR),
});

type RoleFormValues = z.infer<typeof roleSchema>;

interface ServerFailure {
  name?: string;
  notice?: { variant: NoticeVariant; message: string };
}

function stringDetail(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() !== '' ? value : undefined;
}

/** Traduce el fallo del servidor a los campos y avisos de ux.md §4.d/§7. */
function describeServerError(error: ApiError): ServerFailure {
  if (error.reason !== 'http') {
    return { notice: { variant: 'error', message: SYSTEM_ERROR } };
  }
  if (error.status === 401) {
    return { notice: { variant: 'error', message: SESSION_EXPIRED_ERROR } };
  }
  if (error.status === 409) {
    return { name: error.message || ROLE_NAME_IN_USE };
  }
  if (error.status === 400) {
    const details = error.details ?? {};
    const name = stringDetail(details.name);
    if (name) {
      return { name };
    }
    const permissions = stringDetail(details.permissions);
    return { notice: { variant: 'error', message: permissions ?? error.message ?? SYSTEM_ERROR } };
  }
  return { notice: { variant: 'error', message: SYSTEM_ERROR } };
}

function applyServerFailure(
  failure: ServerFailure,
  setError: UseFormSetError<RoleFormValues>,
): void {
  if (failure.name) {
    setError('name', { type: 'server', message: failure.name });
  }
}

/**
 * Ficha de rol (ux.md §3.7): nombre único (se normaliza en el servidor) y
 * casillas de permisos por módulo. Los módulos F3–F9 llevan el sello
 * "disponible más adelante" (información, no error). Exige **al menos un
 * permiso** (FR-014/US4 esc. 4): sin ninguno, el botón queda deshabilitado y se
 * explica el motivo. Las combinaciones de permisos son libres (US4 esc. 2).
 */
export function RoleForm({ mode, role, onSuccess, onCancel }: RoleFormProps) {
  const createRole = useCreateRole();
  const updateRole = useUpdateRole();
  const permissionsQuery = usePermissions();
  const isEditing = mode === 'edit';

  const defaults = useMemo<RoleFormValues>(
    () => ({ name: role?.name ?? '', permissions: role?.permissions ?? [] }),
    [role],
  );

  const {
    register,
    handleSubmit,
    control,
    setError,
    reset,
    formState: { errors },
  } = useForm<RoleFormValues>({
    resolver: zodResolver(roleSchema),
    defaultValues: defaults,
  });

  const selectedPermissions = useWatch({ control, name: 'permissions' }) ?? [];
  const noPermissions = selectedPermissions.length === 0;

  const mutation = isEditing ? updateRole : createRole;
  const serverFailure = useMemo(
    () => (mutation.error ? describeServerError(mutation.error) : null),
    [mutation.error],
  );
  const notice = serverFailure?.notice ?? null;

  useEffect(() => {
    if (!serverFailure) {
      return;
    }
    applyServerFailure(serverFailure, setError);
  }, [serverFailure, setError]);

  const onSubmit = handleSubmit((values) => {
    const input = { name: values.name, permissions: values.permissions };
    if (isEditing && role) {
      updateRole.mutate(
        { id: role.id, input },
        { onSuccess: (updated) => onSuccess(updated, 'edit') },
      );
      return;
    }
    createRole.mutate(input, {
      onSuccess: (created) => {
        reset({ name: '', permissions: [] });
        onSuccess(created, 'create');
      },
    });
  });

  return (
    <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
      {notice && <Notice variant={notice.variant}>{notice.message}</Notice>}

      <Field
        label="Nombre"
        required
        help="Se compara ignorando mayúsculas y espacios."
        error={errors.name?.message}
        {...register('name')}
      />

      <fieldset className="rounded border border-slate-200 p-4">
        <legend className="px-1 font-medium text-slate-900">Permisos por módulo</legend>

        {permissionsQuery.isPending && <EmptyState kind="loading" message="Cargando permisos…" />}

        {permissionsQuery.isError && (
          <EmptyState
            kind="error"
            message="No se pudieron cargar los permisos."
            action={<Button onClick={() => void permissionsQuery.refetch()}>Reintentar</Button>}
          />
        )}

        {permissionsQuery.isSuccess && (
          <ul className="flex flex-col gap-2">
            {(permissionsQuery.data?.items ?? []).map((permission) => (
              <li key={permission.code} className="flex min-h-11 flex-wrap items-center gap-2">
                <input
                  id={permission.code}
                  type="checkbox"
                  value={permission.code}
                  className="h-5 w-5 rounded border-slate-400"
                  {...register('permissions')}
                />
                <label htmlFor={permission.code} className="text-slate-900">
                  {permission.label}
                </label>
                {!isPermissionAvailable(permission.code) && (
                  <span className="rounded-full border border-slate-400 bg-slate-50 px-2 py-0.5 text-sm text-slate-700">
                    {PENDING_PERMISSION_NOTE}
                  </span>
                )}
              </li>
            ))}
          </ul>
        )}

        {noPermissions && (
          <p role="alert" className="mt-3 text-sm font-medium text-red-700">
            {NO_PERMISSIONS_ERROR}
          </p>
        )}
      </fieldset>

      <div className="flex flex-wrap justify-end gap-3">
        <Button type="button" variant="secondary" onClick={onCancel}>
          Cancelar
        </Button>
        <Button type="submit" loading={mutation.isPending} disabled={noPermissions}>
          {isEditing ? 'Guardar cambios' : 'Crear rol'}
        </Button>
      </div>
    </form>
  );
}
