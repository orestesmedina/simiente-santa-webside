import { zodResolver } from '@hookform/resolvers/zod';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useForm, useWatch } from 'react-hook-form';
import { Link, useNavigate } from 'react-router-dom';
import { z } from 'zod';
import { ApiError } from '../../../api/client';
import { Button } from '../../../components/Button';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { PasswordField } from '../../../components/PasswordField';
import {
  evaluatePasswordPolicy,
  isPasswordPolicyMet,
  type PasswordPolicyContext,
} from '../passwordPolicy';
import { useChangePassword } from '../hooks/useChangePassword';
import { useSessionQuery } from '../hooks/useSession';

export const PASSWORD_CHANGED_MESSAGE =
  'Contraseña cambiada. La próxima vez que entres, usa la nueva.';
export const MANDATORY_PASSWORD_MESSAGE =
  'Por seguridad, cambia esta contraseña antes de continuar.';

const SYSTEM_ERROR =
  'No se pudo completar la operación. Vuelve a intentarlo en unos minutos; si sigue, avísanos.';
const SESSION_EXPIRED_ERROR =
  'Tu sesión terminó. Entra de nuevo y vuelve a hacer la acción; no se guardó a medias.';
const CURRENT_PASSWORD_ERROR = 'Tu contraseña actual no coincide. Vuelve a escribirla.';

type PasswordFieldName = 'currentPassword' | 'newPassword';

type PasswordErrorView =
  { field: PasswordFieldName; message: string } | { variant: NoticeVariant; message: string };

/** Traduce el fallo del cambio a los mensajes de ux.md §4.e/§7 (US7). */
function describePasswordError(error: ApiError): PasswordErrorView {
  if (error.reason !== 'http') {
    return { variant: 'error', message: SYSTEM_ERROR };
  }
  if (error.status === 400) {
    if (typeof error.details?.currentPassword === 'string') {
      return { field: 'currentPassword', message: CURRENT_PASSWORD_ERROR };
    }
    if (typeof error.details?.newPassword === 'string') {
      return { field: 'newPassword', message: error.details.newPassword };
    }
    return { field: 'newPassword', message: SYSTEM_ERROR };
  }
  if (error.status === 401) {
    return { variant: 'error', message: SESSION_EXPIRED_ERROR };
  }
  return { variant: 'error', message: SYSTEM_ERROR };
}

function buildSchema(context: PasswordPolicyContext) {
  return z
    .object({
      currentPassword: z.string().min(1, 'Escribe tu contraseña actual.'),
      newPassword: z.string().min(1, 'Escribe la contraseña nueva con las reglas indicadas abajo.'),
      confirmPassword: z.string().min(1, 'Repite la contraseña nueva.'),
    })
    .superRefine((values, ctx) => {
      const unmet = evaluatePasswordPolicy(values.newPassword, context).filter(
        (check) => !check.met,
      );
      if (unmet.length > 0) {
        ctx.addIssue({
          code: 'custom',
          path: ['newPassword'],
          message: `Requisito incumplido: ${unmet[0].label}.`,
        });
      }
      if (values.newPassword !== values.confirmPassword) {
        ctx.addIssue({
          code: 'custom',
          path: ['confirmPassword'],
          message: 'Las contraseñas nuevas no coinciden.',
        });
      }
    });
}

type ChangePasswordForm = z.infer<ReturnType<typeof buildSchema>>;

/**
 * Cambio de la propia contraseña (ux.md §3.8) en dos modos: **voluntario**
 * (desde "Mi cuenta") y **obligatorio** (sin navegación y con el aviso de
 * seguridad, FR-010/US7 esc. 4). Incluye la checklist en vivo de la política
 * FR-010 y los mensajes del requisito incumplido y de contraseña actual
 * incorrecta. Al completarse, vuelve al panel.
 *
 * Nota: el contrato exige `currentPassword` también en el modo obligatorio, así
 * que el campo se pide en ambos modos (ver reporte de la tarea).
 */
export function ChangePasswordPage() {
  const navigate = useNavigate();
  const { data: session } = useSessionQuery();
  const changePassword = useChangePassword();
  const [success, setSuccess] = useState(false);
  const noticeRef = useRef<HTMLDivElement>(null);

  const mandatory = session?.mustChangePassword ?? false;
  const policyContext = useMemo<PasswordPolicyContext>(
    () => ({
      firstName: session?.firstName,
      lastName: session?.lastName,
      email: session?.email,
    }),
    [session?.firstName, session?.lastName, session?.email],
  );
  const schema = useMemo(() => buildSchema(policyContext), [policyContext]);

  const {
    register,
    handleSubmit,
    control,
    setError,
    formState: { errors },
  } = useForm<ChangePasswordForm>({
    resolver: zodResolver(schema),
    defaultValues: { currentPassword: '', newPassword: '', confirmPassword: '' },
  });

  const newPassword = useWatch({ control, name: 'newPassword' }) ?? '';
  const policyChecks = evaluatePasswordPolicy(newPassword, policyContext);
  const policyMet = isPasswordPolicyMet(policyChecks);

  const serverErrorView = useMemo(
    () => (changePassword.error ? describePasswordError(changePassword.error) : null),
    [changePassword.error],
  );
  const fieldError = serverErrorView && 'field' in serverErrorView ? serverErrorView : null;
  const noticeError = serverErrorView && 'variant' in serverErrorView ? serverErrorView : null;

  useEffect(() => {
    if (fieldError) {
      setError(fieldError.field, { type: 'server', message: fieldError.message });
    }
  }, [fieldError, setError]);

  useEffect(() => {
    if (noticeError) {
      noticeRef.current?.focus();
    }
  }, [noticeError]);

  const onSubmit = handleSubmit((values) => {
    changePassword.mutate(
      { currentPassword: values.currentPassword, newPassword: values.newPassword },
      { onSuccess: () => setSuccess(true) },
    );
  });

  if (success) {
    return (
      <main className="mx-auto w-full max-w-md px-4 py-8">
        <h1 className="text-2xl font-bold text-slate-900">Cambiar contraseña</h1>
        <Notice variant="success" className="mt-4">
          {PASSWORD_CHANGED_MESSAGE}
        </Notice>
        <div className="mt-6">
          <Button onClick={() => navigate('/panel', { replace: true })}>Ir al panel</Button>
        </div>
      </main>
    );
  }

  return (
    <main className="mx-auto w-full max-w-md px-4 py-8">
      <h1 className="text-2xl font-bold text-slate-900">Cambiar contraseña</h1>

      {mandatory && (
        <Notice variant="warning" className="mt-4">
          {MANDATORY_PASSWORD_MESSAGE}
        </Notice>
      )}

      {noticeError && (
        <div ref={noticeRef} tabIndex={-1} className="mt-4 focus:outline-none">
          <Notice variant={noticeError.variant}>{noticeError.message}</Notice>
        </div>
      )}

      <form onSubmit={onSubmit} noValidate className="mt-6 flex flex-col gap-4">
        <PasswordField
          label="Contraseña actual"
          autoComplete="current-password"
          required
          help={mandatory ? 'Escribe la contraseña con la que has entrado.' : undefined}
          error={errors.currentPassword?.message}
          {...register('currentPassword')}
        />
        <PasswordField
          label="Contraseña nueva"
          autoComplete="new-password"
          required
          policy={policyChecks}
          error={errors.newPassword?.message}
          {...register('newPassword')}
        />
        <PasswordField
          label="Confirmar contraseña nueva"
          autoComplete="new-password"
          required
          error={errors.confirmPassword?.message}
          {...register('confirmPassword')}
        />
        <Button type="submit" loading={changePassword.isPending} disabled={!policyMet}>
          {changePassword.isPending ? 'Guardando…' : 'Guardar'}
        </Button>
      </form>

      {!mandatory && (
        <p className="mt-6">
          <Link
            to="/panel"
            className="font-medium text-slate-900 underline underline-offset-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-900"
          >
            Volver al panel
          </Link>
        </p>
      )}
    </main>
  );
}
