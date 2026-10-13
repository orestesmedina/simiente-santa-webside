import { useEffect, useId, useRef, type ReactNode } from 'react';
import { Button } from './Button';

/** Ancho del panel del diálogo (QA Bug 3). */
export type DialogSize = 'md' | 'wide';

export interface DialogProps {
  title: string;
  description?: string;
  /**
   * Ancho máximo del panel: `md` es el de siempre (32 rem; confirmaciones) y
   * `wide` da aire a los formularios con rejillas —el editor de fechas de la
   * agenda— a partir de `sm` (42 rem), sin encoger los diálogos de confirmación.
   */
  size?: DialogSize;
  onClose: () => void;
  children: ReactNode;
}

const SIZE_CLASSES: Record<DialogSize, string> = {
  md: 'max-w-lg',
  wide: 'max-w-lg sm:max-w-2xl',
};

const FOCUSABLE_SELECTOR = [
  'a[href]',
  'button:not([disabled])',
  'textarea:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',');

/**
 * Diálogo modal accesible: `role="dialog"` con `aria-modal`, título asociado,
 * cierre con `Esc`, trampa de foco y devolución del foco al elemento que lo
 * abrió (ux.md §6). Es la base de `ConfirmDialog` y de los formularios en modal.
 *
 * El panel desplaza **solo** en vertical (`overflow-x-hidden`, QA Bug 3): el
 * `overflow-y-auto` a solas convierte el eje horizontal en desplazable y un
 * contenido ancho —el editor de fechas— sacaba una barra horizontal.
 */
export function Dialog({ title, description, size = 'md', onClose, children }: DialogProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const titleId = useId();
  const descriptionId = useId();

  useEffect(() => {
    const previouslyFocused =
      document.activeElement instanceof HTMLElement ? document.activeElement : null;

    const getFocusable = () =>
      containerRef.current
        ? Array.from(containerRef.current.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR))
        : [];

    getFocusable()[0]?.focus();

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        onClose();
        return;
      }
      if (event.key !== 'Tab') {
        return;
      }
      const focusable = getFocusable();
      if (focusable.length === 0) {
        return;
      }
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      previouslyFocused?.focus();
    };
  }, [onClose]);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          onClose();
        }
      }}
    >
      <div
        ref={containerRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={description ? descriptionId : undefined}
        className={`max-h-full w-full ${SIZE_CLASSES[size]} overflow-y-auto overflow-x-hidden rounded bg-white p-6 shadow-xl`}
      >
        <div className="flex items-start justify-between gap-4">
          <h2 id={titleId} className="text-xl font-bold text-slate-900">
            {title}
          </h2>
          <Button variant="link" onClick={onClose} aria-label="Cerrar">
            Cerrar
          </Button>
        </div>
        {description && (
          <p id={descriptionId} className="mt-2 text-slate-700">
            {description}
          </p>
        )}
        <div className="mt-4">{children}</div>
      </div>
    </div>
  );
}
