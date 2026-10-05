import { zodResolver } from '@hookform/resolvers/zod';
import { useEffect, useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { z } from 'zod';
import { ApiError } from '../../../api/client';
import { Button } from '../../../components/Button';
import { Field } from '../../../components/Field';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { PasswordField } from '../../../components/PasswordField';
import { useLogin } from '../hooks/useLogin';

const loginSchema = z.object({
  email: z
    .string()
    .trim()
    .min(1, 'Escribe tu correo.')
    .email('Este correo no tiene el formato correcto.'),
  password: z.string().min(1, 'Escribe tu contraseña.'),
});

type LoginForm = z.infer<typeof loginSchema>;

const GENERIC_CREDENTIALS_ERROR = 'Correo o contraseña incorrectos.';
const ACCESS_DISABLED_ERROR =
  'Ese acceso está desactivado. Pide a un administrador de la iglesia que lo reactive para poder entrar.';
const SYSTEM_ERROR = 'No se pudo conectar con el sistema. Vuelve a intentarlo en unos minutos.';
const DEFAULT_RETRY_AFTER_SECONDS = 900;

interface LoginNotice {
  variant: NoticeVariant;
  message: string;
}

/** Aviso pedido por la URL (`motivo`) al volver al acceso sin sesión (FR-005). */
function describeLoginMotivo(motivo: string | null): LoginNotice | null {
  switch (motivo) {
    case 'expirada':
      return { variant: 'warning', message: 'Tu sesión terminó. Vuelve a entrar para continuar.' };
    case 'techo':
      return {
        variant: 'warning',
        message:
          'Tu sesión terminó (cada sesión dura como máximo 1 hora). Vuelve a entrar para continuar.',
      };
    case 'cuenta-desactivada':
      return {
        variant: 'warning',
        message:
          'Tu cuenta se desactivó. Para entrar de nuevo, pide a un administrador que la reactive.',
      };
    default:
      return null;
  }
}

/** Traduce el fallo del login a los mensajes de ux.md §4.a/§7 (FR-003/SC-008). */
function describeLoginError(error: ApiError): LoginNotice {
  if (error.reason !== 'http') {
    return { variant: 'error', message: SYSTEM_ERROR };
  }
  switch (error.status) {
    case 401:
      return { variant: 'error', message: GENERIC_CREDENTIALS_ERROR };
    case 403:
      if (error.details?.reason === 'access_disabled') {
        return { variant: 'warning', message: ACCESS_DISABLED_ERROR };
      }
      return { variant: 'error', message: SYSTEM_ERROR };
    default:
      return { variant: 'error', message: SYSTEM_ERROR };
  }
}

function retryAfterSeconds(error: ApiError): number {
  const value = error.details?.retryAfterSeconds;
  return typeof value === 'number' && Number.isFinite(value) && value > 0
    ? value
    : DEFAULT_RETRY_AFTER_SECONDS;
}

function formatBlockedMessage(blockedUntil: number, now: number): string {
  const remainingMinutes = Math.max(1, Math.ceil((blockedUntil - now) / 60_000));
  const time = new Date(blockedUntil).toLocaleTimeString('es-ES', {
    hour: '2-digit',
    minute: '2-digit',
  });
  return `Por seguridad, se han superado los intentos permitidos. Podrás volver a intentarlo a partir de las ${time} (quedan ${remainingMinutes} minutos).`;
}

/**
 * Pantalla de acceso (ux.md §3.1): correo y contraseña con mostrar/ocultar,
 * `autocomplete` correcto y sin "regístrate" ni "olvidé mi contraseña" (fuera de
 * alcance del MVP). Tras entrar, vuelve al `destino` o a `/panel` (SC-002).
 */
export function LoginPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const loginMutation = useLogin();
  const noticeRef = useRef<HTMLDivElement>(null);
  const [blockedUntil, setBlockedUntil] = useState<number | null>(null);
  const [now, setNow] = useState(() => Date.now());

  const motivoNotice = describeLoginMotivo(searchParams.get('motivo'));
  const serverNotice =
    loginMutation.error && blockedUntil === null ? describeLoginError(loginMutation.error) : null;
  const activeNotice = blockedUntil !== null ? null : serverNotice;

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  });

  // Cuenta atrás visible del bloqueo; al expirar, el formulario vuelve solo.
  useEffect(() => {
    if (blockedUntil === null) {
      return;
    }
    const deadline = blockedUntil;
    const id = window.setInterval(() => {
      const current = Date.now();
      setNow(current);
      if (current >= deadline) {
        window.clearInterval(id);
      }
    }, 1000);
    return () => window.clearInterval(id);
  }, [blockedUntil]);

  // Foco en el aviso cuando aparece un error del servidor (ux.md §6).
  useEffect(() => {
    if (activeNotice) {
      noticeRef.current?.focus();
    }
  }, [activeNotice]);

  const bloqueado = blockedUntil !== null && now < blockedUntil;

  const onSubmit = handleSubmit((values) => {
    setBlockedUntil(null);
    loginMutation.mutate(values, {
      onSuccess: () => {
        const destino = searchParams.get('destino');
        const target =
          destino && destino.startsWith('/') && !destino.startsWith('//') ? destino : '/panel';
        navigate(target, { replace: true });
      },
      onError: (error) => {
        if (error.reason === 'http' && error.status === 429) {
          const current = Date.now();
          setBlockedUntil(current + retryAfterSeconds(error) * 1000);
          setNow(current);
        }
      },
    });
  });

  return (
    <main className="mx-auto w-full max-w-md px-4 py-8">
      <h1 className="text-2xl font-bold text-slate-900">Entrar al panel</h1>

      {motivoNotice && (
        <Notice variant={motivoNotice.variant} className="mt-4">
          {motivoNotice.message}
        </Notice>
      )}

      {activeNotice && (
        <div ref={noticeRef} tabIndex={-1} className="mt-4 focus:outline-none">
          <Notice variant={activeNotice.variant}>{activeNotice.message}</Notice>
        </div>
      )}

      {bloqueado && blockedUntil !== null && (
        <Notice variant="warning" className="mt-4">
          {formatBlockedMessage(blockedUntil, now)}
        </Notice>
      )}

      <form onSubmit={onSubmit} noValidate className="mt-6 flex flex-col gap-4">
        <Field
          label="Correo"
          type="email"
          autoComplete="email"
          required
          disabled={bloqueado}
          error={errors.email?.message}
          {...register('email')}
        />
        <PasswordField
          label="Contraseña"
          autoComplete="current-password"
          required
          disabled={bloqueado}
          error={errors.password?.message}
          {...register('password')}
        />
        <Button type="submit" loading={loginMutation.isPending} disabled={bloqueado}>
          {loginMutation.isPending ? 'Comprobando…' : 'Entrar'}
        </Button>
      </form>
    </main>
  );
}
