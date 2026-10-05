import { Button } from './Button';
import { Dialog } from './Dialog';

export interface ConfirmDialogProps {
  title: string;
  /** Texto que explica la consecuencia de la acción. */
  description: string;
  confirmText: string;
  cancelText?: string;
  /** Estilo destructivo (p. ej. desactivar una cuenta o eliminar un rol). */
  danger?: boolean;
  loading?: boolean;
  onConfirm: () => void;
  onClose: () => void;
}

/**
 * Confirmación accesible construida sobre `Dialog`: botón de cancelar primero
 * en el orden de foco y confirmación explícita. Nunca borra ni desactiva nada
 * sin esta confirmación.
 */
export function ConfirmDialog({
  title,
  description,
  confirmText,
  cancelText = 'Cancelar',
  danger = false,
  loading = false,
  onConfirm,
  onClose,
}: ConfirmDialogProps) {
  return (
    <Dialog title={title} description={description} onClose={onClose}>
      <div className="flex flex-wrap justify-end gap-3">
        <Button variant="secondary" onClick={onClose} disabled={loading}>
          {cancelText}
        </Button>
        <Button variant={danger ? 'danger' : 'primary'} loading={loading} onClick={onConfirm}>
          {confirmText}
        </Button>
      </div>
    </Dialog>
  );
}
