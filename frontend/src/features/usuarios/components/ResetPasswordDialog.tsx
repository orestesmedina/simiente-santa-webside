import { zodResolver } from '@hookform/resolvers/zod';
import { useEffect, useMemo } from 'react';
import { useForm, useWatch } from 'react-hook-form';
import { z } from 'zod';
import type { UserItem } from '../../../api/usuarios';
import { Button } from '../../../components/Button';
import { Dialog } from '../../../components/Dialog';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { PasswordField } from '../../../components/PasswordField';
import {
  evaluatePasswordPolicy,
  isPasswordPolicyMet,
  type PasswordPolicyContext,
} from '../../auth/passwordPolicy';
import { SESSION_EXPIRED_ERROR, SYSTEM_ERROR } from '../messages';
import { useResetPassword } from '../hooks/useResetPassword';

export interface ResetPasswordDialogProps {
  user: UserItem;
  onClose: () => void;
  onSuccess: () => void;
}

const passwordSchema = (context: PasswordPolicyContext) =>
  z
    .object({
      password: z.string(),
      confirmPassword: z.string(),
    })
    .superRefine((values, ctx) => {
      if (values.password.trim() === '') {
        ctx.addIssue({
          code: 'custom',
          path: ['password'],
          message: 'Escribe la contraseña con las reglas indicadas abajo.',
        });
        return;
      }
      const unmet = evaluatePasswordPolicy(values.password, context).filter((check) => !check.met);
      if (unmet.length > 0) {
        ctx.addIssue({
          code: 'custom',
          path: ['password'],
          message: `Requisito incumplido: ${unmet[0].label}.`,
        });
      }
      if (values.password !== values.confirmPassword) {
        ctx.addIssue({
          code: 'custom',
          path: ['confirmPassword'],
          message: 'Las contraseñas no coinciden.',
        });
      }
    });

type ResetPasswordValues = z.infer<ReturnType<typeof passwordSchema>>;

function stringDetail(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() !== '' ? value : undefined;
}

/**
 * Restablecimiento de la contraseña de una cuenta (FR-010, US7 esc. 5): el
 * administrador define una contraseña nueva que cumple la política; su titular
 * deberá cambiarla al entrar y todas sus sesiones abiertas se revocan. La
 * contraseña vive solo en el formulario: nunca se devuelve ni se muestra tras
 * guardar (FR-003/FR-026).
 */
export function ResetPasswordDialog({ user, onClose, onSuccess }: ResetPasswordDialogProps) {
  const resetPassword = useResetPassword();

  const policyContext = useMemo<PasswordPolicyContext>(
    () => ({ firstName: user.firstName, lastName: user.lastName, email: user.email }),
    [user.firstName, user.lastName, user.email],
  );
  const schema = useMemo(() => passwordSchema(policyContext), [policyContext]);

  const {
    register,
    handleSubmit,
    control,
    setError,
    formState: { errors },
  } = useForm<ResetPasswordValues>({
    resolver: zodResolver(schema),
    defaultValues: { password: '', confirmPassword: '' },
  });

  const password = useWatch({ control, name: 'password' }) ?? '';
  const policyChecks = evaluatePasswordPolicy(password, policyContext);
  const policyMet = isPasswordPolicyMet(policyChecks);

  useEffect(() => {
    const error = resetPassword.error;
    if (!error) {
      return;
    }
    if (error.reason === 'http' && error.status === 400) {
      const detail = stringDetail(error.details?.password);
      if (detail) {
        setError('password', { type: 'server', message: detail });
      }
    }
  }, [resetPassword.error, setError]);

  const serverNotice = useMemo<{ variant: NoticeVariant; message: string } | null>(() => {
    const error = resetPassword.error;
    if (!error) {
      return null;
    }
    if (error.reason !== 'http') {
      return { variant: 'error', message: SYSTEM_ERROR };
    }
    if (error.status === 401) {
      return { variant: 'error', message: SESSION_EXPIRED_ERROR };
    }
    if (error.status === 400 && stringDetail(error.details?.password)) {
      return null; // ya se muestra junto al campo
    }
    return { variant: 'error', message: SYSTEM_ERROR };
  }, [resetPassword.error]);

  const onSubmit = handleSubmit((values) => {
    resetPassword.mutate(
      { id: user.id, input: { password: values.password } },
      { onSuccess: () => onSuccess() },
    );
  });

  return (
    <Dialog
      title="Definir una contraseña nueva"
      description={`Para ${user.firstName} ${user.lastName} (${user.email}).`}
      onClose={onClose}
    >
      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        {serverNotice && <Notice variant={serverNotice.variant}>{serverNotice.message}</Notice>}
        <Notice variant="info">
          La persona deberá cambiar esta contraseña la primera vez que entre. Sus sesiones abiertas
          se cerrarán.
        </Notice>
        <PasswordField
          label="Contraseña nueva"
          required
          autoComplete="new-password"
          policy={policyChecks}
          error={errors.password?.message}
          {...register('password')}
        />
        <PasswordField
          label="Confirmar contraseña nueva"
          required
          autoComplete="new-password"
          error={errors.confirmPassword?.message}
          {...register('confirmPassword')}
        />
        <div className="flex flex-wrap justify-end gap-3">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancelar
          </Button>
          <Button type="submit" loading={resetPassword.isPending} disabled={!policyMet}>
            Restablecer contraseña
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
