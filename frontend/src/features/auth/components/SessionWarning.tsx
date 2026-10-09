import { useQueryClient } from '@tanstack/react-query';
import { useEffect, useId, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getSession } from '../../../api/auth';
import { Button } from '../../../components/Button';
import { useLogout } from '../hooks/useLogout';
import { SESSION_QUERY_KEY } from '../hooks/useSession';

/** Inactividad máxima de la sesión (FR-005): 30 minutos. */
export const INACTIVITY_MS = 30 * 60 * 1000;
/** Antelación del aviso "¿Sigues ahí?" (FR-005): 2 minutos. */
export const SESSION_WARNING_MS = 2 * 60 * 1000;
/** Techo absoluto de la sesión (FR-005): 1 hora, haya actividad o no. */
export const ABSOLUTE_SESSION_MS = 60 * 60 * 1000;

const DEFAULT_TICK_MS = 1000;
const ACTIVITY_EVENTS = ['mousemove', 'mousedown', 'keydown', 'touchstart', 'scroll', 'wheel'];

export interface SessionWarningProps {
  /** Inactividad que cierra la sesión. */
  inactivityMs?: number;
  /** Antelación con la que se muestra el aviso. */
  warningMs?: number;
  /** Vida absoluta de la sesión. */
  absoluteMs?: number;
  /** Frecuencia de comprobación del temporizador. */
  tickMs?: number;
}

function formatRemaining(remainingMs: number): string {
  const minutes = Math.max(1, Math.ceil(remainingMs / 60_000));
  return `Tu sesión se cerrará en ${minutes} minuto${minutes === 1 ? '' : 's'} si no la usamos.`;
}

/**
 * Aviso global de sesión por caducar (ux.md §3.11, FR-005): una banda superior y
 * un diálogo con "Sigo aquí" (renueva la actividad) y "Salir". Al alcanzar los
 * 30 minutos de inactividad o el techo absoluto de 1 hora, redirige a `/login`
 * con su aviso. No cierra con `Esc`: la sesión solo se renueva o se abandona de
 * forma explícita.
 */
export function SessionWarning({
  inactivityMs = INACTIVITY_MS,
  warningMs = SESSION_WARNING_MS,
  absoluteMs = ABSOLUTE_SESSION_MS,
  tickMs = DEFAULT_TICK_MS,
}: SessionWarningProps) {
  const navigate = useNavigate();
  const logoutMutation = useLogout();
  const queryClient = useQueryClient();
  const [visible, setVisible] = useState(false);
  const [remainingMs, setRemainingMs] = useState(0);

  const titleId = useId();
  const descriptionId = useId();
  const dialogRef = useRef<HTMLDivElement>(null);
  const previouslyFocusedRef = useRef<HTMLElement | null>(null);
  const startedAtRef = useRef(0);
  const lastActivityRef = useRef(0);
  const expiredRef = useRef(false);

  useEffect(() => {
    const now = Date.now();
    startedAtRef.current = now;
    lastActivityRef.current = now;
    const registerActivity = () => {
      lastActivityRef.current = Date.now();
    };
    ACTIVITY_EVENTS.forEach((event) =>
      window.addEventListener(event, registerActivity, { passive: true }),
    );
    return () => {
      ACTIVITY_EVENTS.forEach((event) => window.removeEventListener(event, registerActivity));
    };
  }, []);

  useEffect(() => {
    const expire = (motivo: 'expirada' | 'techo') => {
      if (expiredRef.current) {
        return;
      }
      expiredRef.current = true;
      navigate(`/login?motivo=${motivo}`, { replace: true });
    };

    const id = window.setInterval(() => {
      const now = Date.now();
      if (now - startedAtRef.current >= absoluteMs) {
        expire('techo');
        return;
      }
      const inactiveFor = now - lastActivityRef.current;
      if (inactiveFor >= inactivityMs) {
        expire('expirada');
        return;
      }
      if (inactiveFor >= inactivityMs - warningMs) {
        const deadline = lastActivityRef.current + inactivityMs;
        setRemainingMs(deadline - now);
        setVisible(true);
      } else {
        setVisible(false);
      }
    }, tickMs);

    return () => window.clearInterval(id);
  }, [absoluteMs, inactivityMs, warningMs, tickMs, navigate]);

  useEffect(() => {
    if (visible) {
      previouslyFocusedRef.current =
        document.activeElement instanceof HTMLElement ? document.activeElement : null;
      dialogRef.current?.querySelector<HTMLButtonElement>('button')?.focus();
      return;
    }
    previouslyFocusedRef.current?.focus();
    previouslyFocusedRef.current = null;
  }, [visible]);

  const handleStay = () => {
    lastActivityRef.current = Date.now();
    setVisible(false);
    // Una petición real refresca la inactividad del servidor sin suscribir al
    // componente a la consulta de sesión (evita re-renders innecesarios).
    void queryClient
      .fetchQuery({ queryKey: SESSION_QUERY_KEY, queryFn: getSession, staleTime: 0 })
      .catch(() => undefined);
  };

  const handleLogout = () => {
    logoutMutation.mutate(undefined, {
      onSuccess: () => navigate('/login', { replace: true }),
    });
  };

  if (!visible) {
    return null;
  }

  const message = formatRemaining(remainingMs);

  return (
    <>
      <div
        role="status"
        className="fixed inset-x-0 top-0 z-40 border-b border-amber-600 bg-amber-50 px-4 py-2 text-center text-sm font-medium text-amber-900"
      >
        {message}
      </div>
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
        <div
          ref={dialogRef}
          role="alertdialog"
          aria-modal="true"
          aria-labelledby={titleId}
          aria-describedby={descriptionId}
          className="w-full max-w-md rounded bg-white p-6 shadow-xl"
        >
          <h2 id={titleId} className="text-xl font-bold text-slate-900">
            ¿Sigues ahí?
          </h2>
          <p id={descriptionId} className="mt-2 text-slate-700">
            {message}
          </p>
          <div className="mt-4 flex flex-wrap gap-3">
            <Button onClick={handleStay}>Sigo aquí</Button>
            <Button variant="secondary" loading={logoutMutation.isPending} onClick={handleLogout}>
              Salir
            </Button>
          </div>
        </div>
      </div>
    </>
  );
}
