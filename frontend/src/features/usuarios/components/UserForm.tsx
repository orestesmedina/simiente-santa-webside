import { zodResolver } from '@hookform/resolvers/zod';
import { useEffect, useMemo } from 'react';
import { useForm, useWatch, type UseFormSetError } from 'react-hook-form';
import { z } from 'zod';
import { ApiError } from '../../../api/client';
import type { UserItem } from '../../../api/usuarios';
import { Button } from '../../../components/Button';
import { Field } from '../../../components/Field';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { PasswordField } from '../../../components/PasswordField';
import { Select } from '../../../components/Select';
import { StatusPill } from '../../../components/StatusPill';
import {
  evaluatePasswordPolicy,
  isPasswordPolicyMet,
  type PasswordPolicyContext,
} from '../../auth/passwordPolicy';
import { formatLastLogin } from '../../../lib/format';
import {
  EMAIL_IN_USE_ERROR,
  LAST_ADMIN_ERROR,
  SESSION_EXPIRED_ERROR,
  SYSTEM_ERROR,
} from '../messages';
import { useCreateUser } from '../hooks/useCreateUser';
import { useRoleOptions } from '../hooks/useRoleOptions';
import { useUpdateUser } from '../hooks/useUpdateUser';

export type UserFormMode = 'create' | 'edit';

export interface UserFormProps {
  mode: UserFormMode;
  /** Cuenta que se edita; obligatoria en modo `edit`. */
  user?: UserItem;
  onSuccess: (user: UserItem, mode: UserFormMode) => void;
  onCancel: () => void;
  /** Abre el diálogo de restablecimiento desde la ficha (solo en edición). */
  onRequestResetPassword?: (user: UserItem) => void;
}

/** Teléfono del contrato: separadores habituales, prefijo opcional y ≥ 7 dígitos. */
function isValidPhone(value: string): boolean {
  if (!/^\+?[0-9\s().-]+$/.test(value)) {
    return false;
  }
  const digits = value.replace(/\D/g, '');
  return digits.length >= 7 && value.length <= 32;
}

function buildSchema(mode: UserFormMode, context: PasswordPolicyContext) {
  return z
    .object({
      firstName: z
        .string()
        .trim()
        .min(1, 'Escribe el nombre.')
        .max(80, 'El nombre no puede superar los 80 caracteres.'),
      lastName: z
        .string()
        .trim()
        .min(1, 'Escribe los apellidos.')
        .max(120, 'Los apellidos no pueden superar los 120 caracteres.'),
      email: z
        .string()
        .trim()
        .min(1, 'Escribe el correo.')
        .email('Escribe un correo con este formato: nombre@dominio.com')
        .max(254, 'El correo no puede superar los 254 caracteres.'),
      phone: z
        .string()
        .trim()
        .min(1, 'Escribe un número de teléfono.')
        .refine(isValidPhone, 'Escribe un número de teléfono válido.'),
      roleId: z.string().min(1, 'Elige un rol de la lista.'),
      password: z.string().optional(),
    })
    .superRefine((values, ctx) => {
      if (mode !== 'create') {
        return;
      }
      const password = values.password ?? '';
      if (password.trim() === '') {
        ctx.addIssue({
          code: 'custom',
          path: ['password'],
          message: 'Escribe la contraseña con las reglas indicadas abajo.',
        });
        return;
      }
      const unmet = evaluatePasswordPolicy(password, context).filter((check) => !check.met);
      if (unmet.length > 0) {
        ctx.addIssue({
          code: 'custom',
          path: ['password'],
          message: `Requisito incumplido: ${unmet[0].label}.`,
        });
      }
    });
}

type UserFormValues = z.infer<ReturnType<typeof buildSchema>>;
type UserFormField = 'firstName' | 'lastName' | 'email' | 'phone' | 'roleId' | 'password';

interface ServerFailure {
  fields: Partial<Record<UserFormField, string>>;
  notice?: { variant: NoticeVariant; message: string };
}

function stringDetail(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() !== '' ? value : undefined;
}

/** Traduce el fallo del servidor a los campos y avisos de ux.md §4.c/§7. */
function describeServerError(error: ApiError): ServerFailure {
  if (error.reason !== 'http') {
    return { fields: {}, notice: { variant: 'error', message: SYSTEM_ERROR } };
  }
  if (error.status === 401) {
    return { fields: {}, notice: { variant: 'error', message: SESSION_EXPIRED_ERROR } };
  }
  if (error.status === 409) {
    if (error.details?.reason === 'admin_required') {
      return { fields: {}, notice: { variant: 'warning', message: LAST_ADMIN_ERROR } };
    }
    return { fields: { email: error.message || EMAIL_IN_USE_ERROR } };
  }
  if (error.status === 400) {
    const details = error.details ?? {};
    const fields: ServerFailure['fields'] = {};
    const phone = stringDetail(details.phone);
    const roleId = stringDetail(details.roleId);
    const email = stringDetail(details.email);
    const firstName = stringDetail(details.firstName);
    const lastName = stringDetail(details.lastName);
    const password = stringDetail(details.password);
    if (phone) fields.phone = phone;
    if (roleId) fields.roleId = roleId;
    if (email) fields.email = email;
    if (firstName) fields.firstName = firstName;
    if (lastName) fields.lastName = lastName;
    if (password) fields.password = password;
    if (Object.keys(fields).length > 0) {
      return { fields };
    }
    return { fields: {}, notice: { variant: 'error', message: SYSTEM_ERROR } };
  }
  return { fields: {}, notice: { variant: 'error', message: SYSTEM_ERROR } };
}

function applyServerFailure(
  failure: ServerFailure,
  setError: UseFormSetError<UserFormValues>,
): void {
  for (const [field, message] of Object.entries(failure.fields)) {
    if (message) {
      setError(field as UserFormField, { type: 'server', message });
    }
  }
}

/**
 * Ficha de cuenta: detalle y edición (ux.md §3.5). Al crear pide la contraseña
 * inicial con la checklist en vivo de la política FR-010; al editar nunca
 * muestra ni reenvía la contraseña (FR-003/FR-026) y ofrece el
 * restablecimiento como acción aparte. Los errores del servidor se mapean a su
 * campo (correo duplicado, `details.phone`, `details.roleId`) o a un aviso para
 * las reglas que no dependen de un campo (último administrador, FR-008).
 */
export function UserForm({
  mode,
  user,
  onSuccess,
  onCancel,
  onRequestResetPassword,
}: UserFormProps) {
  const createUser = useCreateUser();
  const updateUser = useUpdateUser();
  const roleOptions = useRoleOptions();
  const isEditing = mode === 'edit';

  const defaults = useMemo<UserFormValues>(
    () => ({
      firstName: user?.firstName ?? '',
      lastName: user?.lastName ?? '',
      email: user?.email ?? '',
      phone: user?.phone ?? '',
      roleId: user?.roleId ?? '',
      password: '',
    }),
    [user],
  );

  const schema = useMemo(
    () =>
      buildSchema(mode, {
        firstName: isEditing ? user?.firstName : undefined,
        lastName: isEditing ? user?.lastName : undefined,
        email: isEditing ? user?.email : undefined,
      }),
    [mode, isEditing, user?.firstName, user?.lastName, user?.email],
  );

  const {
    register,
    handleSubmit,
    control,
    setError,
    reset,
    formState: { errors },
  } = useForm<UserFormValues>({
    resolver: zodResolver(schema),
    defaultValues: defaults,
  });

  const password = useWatch({ control, name: 'password' }) ?? '';
  const watchedFirstName = useWatch({ control, name: 'firstName' }) ?? '';
  const watchedLastName = useWatch({ control, name: 'lastName' }) ?? '';
  const watchedEmail = useWatch({ control, name: 'email' }) ?? '';
  const policyContext = useMemo<PasswordPolicyContext>(
    () => ({
      firstName: watchedFirstName || user?.firstName,
      lastName: watchedLastName || user?.lastName,
      email: watchedEmail || user?.email,
    }),
    [watchedFirstName, watchedLastName, watchedEmail, user],
  );
  const policyChecks = evaluatePasswordPolicy(password, policyContext);
  const policyMet = isPasswordPolicyMet(policyChecks);

  const mutation = isEditing ? updateUser : createUser;
  const serverFailure = useMemo(
    () => (mutation.error ? describeServerError(mutation.error) : null),
    [mutation.error],
  );
  const notice = serverFailure?.notice ?? null;

  // Aplica los errores del servidor a su campo cada vez que cambia el fallo.
  useEffect(() => {
    if (!serverFailure) {
      return;
    }
    applyServerFailure(serverFailure, setError);
  }, [serverFailure, setError]);

  const onSubmit = handleSubmit((values) => {
    if (isEditing && user) {
      updateUser.mutate(
        {
          id: user.id,
          input: {
            firstName: values.firstName,
            lastName: values.lastName,
            email: values.email,
            phone: values.phone,
            roleId: values.roleId,
          },
        },
        { onSuccess: (updated) => onSuccess(updated, 'edit') },
      );
      return;
    }
    createUser.mutate(
      {
        firstName: values.firstName,
        lastName: values.lastName,
        email: values.email,
        phone: values.phone,
        roleId: values.roleId,
        password: values.password ?? '',
      },
      {
        onSuccess: (created) => {
          // Ninguna credencial sobrevive al envío (FR-003/FR-026).
          reset({ ...defaults, password: '' });
          onSuccess(created, 'create');
        },
      },
    );
  });

  const roleSelectOptions = [
    {
      value: '',
      label: roleOptions.isLoading ? 'Cargando roles…' : 'Elige un rol',
      disabled: true,
    },
    ...(roleOptions.data ?? []).map((role) => ({ value: role.id, label: role.name })),
  ];

  return (
    <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
      {notice && <Notice variant={notice.variant}>{notice.message}</Notice>}

      <Field
        label="Nombre"
        required
        autoComplete="given-name"
        error={errors.firstName?.message}
        {...register('firstName')}
      />
      <Field
        label="Apellidos"
        required
        autoComplete="family-name"
        error={errors.lastName?.message}
        {...register('lastName')}
      />
      <Field
        label="Correo"
        type="email"
        required
        autoComplete="email"
        help="Se compara ignorando mayúsculas y espacios."
        error={errors.email?.message}
        {...register('email')}
      />
      <Field
        label="Teléfono"
        required
        autoComplete="tel"
        help="Con código del país si procede, p. ej. 612 345 678 o +34 612 345 678."
        error={errors.phone?.message}
        {...register('phone')}
      />
      <Select
        label="Rol"
        required
        options={roleSelectOptions}
        help={
          roleOptions.isError
            ? 'No se pudieron cargar los roles. Vuelve a abrir la ficha.'
            : undefined
        }
        error={errors.roleId?.message}
        {...register('roleId')}
      />

      {!isEditing && (
        <>
          <PasswordField
            label="Contraseña inicial"
            required
            autoComplete="new-password"
            policy={policyChecks}
            help="La persona deberá cambiar esta contraseña la primera vez que entre."
            error={errors.password?.message}
            {...register('password')}
          />
        </>
      )}

      {isEditing && user && (
        <section aria-label="Detalle de la cuenta" className="rounded border border-slate-200 p-4">
          <dl className="grid grid-cols-1 gap-2 sm:grid-cols-2">
            <div>
              <dt className="text-sm font-medium text-slate-700">Estado</dt>
              <dd className="mt-1">
                <StatusPill value={user.isActive ? 'active' : 'inactive'} />
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-slate-700">Último acceso</dt>
              <dd className="mt-1 text-slate-900">
                {formatLastLogin(user.lastLoginAt, user.lastLoginIp)}
              </dd>
            </div>
          </dl>
          <p className="mt-3 text-sm text-slate-600">
            La contraseña nunca se muestra ni se guarda aquí.
          </p>
          {onRequestResetPassword && (
            <div className="mt-3">
              <Button
                type="button"
                variant="secondary"
                onClick={() => onRequestResetPassword(user)}
              >
                Definir una contraseña nueva
              </Button>
            </div>
          )}
        </section>
      )}

      <div className="flex flex-wrap justify-end gap-3">
        <Button type="button" variant="secondary" onClick={onCancel}>
          Cancelar
        </Button>
        <Button type="submit" loading={mutation.isPending} disabled={!isEditing && !policyMet}>
          {isEditing ? 'Guardar cambios' : 'Crear cuenta'}
        </Button>
      </div>
    </form>
  );
}
