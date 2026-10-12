import type { CategoriaAdmin } from '../../../api/agenda';
import { Button } from '../../../components/Button';
import { ConfirmDialog } from '../../../components/ConfirmDialog';
import { Dialog } from '../../../components/Dialog';
import { Notice } from '../../../components/Notice';
import { isConflict } from '../errors';
import { useDeleteCategory } from '../hooks/useEventCategories';
import {
  CATEGORY_IN_USE_TITLE,
  CATEGORY_IN_USE_HINT,
  DELETE_CANCEL,
  DELETE_CONFIRM_ACTION,
  DELETE_CONFIRM_TITLE,
  deleteConfirmDescription,
  categoryInUseMessage,
} from '../messages';

export interface DeleteConfirmProps {
  category: CategoriaAdmin;
  onDeleted: () => void;
  onClose: () => void;
}

/** Recuento del mensaje del servidor («…en uso por 4 elementos») si lo trae. */
function usageCount(message: string | undefined): number | undefined {
  const match = /por (\d+)/.exec(message ?? '');
  return match ? Number(match[1]) : undefined;
}

/**
 * Confirmación del borrado de una categoría (ux.md §4.6/⟲ D-16, decisión firme
 * P4-6): si el servidor responde **`409 conflict`** porque está en uso, el
 * diálogo **lo explica con la cuenta** («Está en uso por N elementos: quítala
 * primero de los elementos») y su **única acción es cerrar**: la interfaz nunca
 * ofrece desligarla de los elementos ni continuar.
 */
export function DeleteConfirm({ category, onDeleted, onClose }: DeleteConfirmProps) {
  const deleteCategory = useDeleteCategory();
  const conflict = isConflict(deleteCategory.error);

  if (conflict) {
    const count = usageCount(deleteCategory.error?.message);
    const explanation =
      count === undefined
        ? (deleteCategory.error?.message ?? CATEGORY_IN_USE_HINT)
        : categoryInUseMessage(count);

    return (
      <Dialog title={CATEGORY_IN_USE_TITLE} onClose={onClose}>
        <Notice variant="error">
          <p>{explanation}</p>
          {count !== undefined && <p className="mt-1">{CATEGORY_IN_USE_HINT}</p>}
        </Notice>
        <div className="mt-4 flex justify-end">
          <Button onClick={onClose}>{DELETE_CANCEL}</Button>
        </div>
      </Dialog>
    );
  }

  return (
    <ConfirmDialog
      title={DELETE_CONFIRM_TITLE}
      description={deleteConfirmDescription(category.nameEs)}
      confirmText={DELETE_CONFIRM_ACTION}
      cancelText={DELETE_CANCEL}
      danger
      loading={deleteCategory.isPending}
      onConfirm={() =>
        deleteCategory.mutate(category.id, {
          onSuccess: onDeleted,
        })
      }
      onClose={onClose}
    />
  );
}
