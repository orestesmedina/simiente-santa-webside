import { useState } from 'react';
import type { CategoriaAdmin, CategoriaInput } from '../../../api/agenda';
import { Button } from '../../../components/Button';
import { Dialog } from '../../../components/Dialog';
import { Field } from '../../../components/Field';
import { Notice } from '../../../components/Notice';
import { detailMessage, generalErrorMessage, isConflict } from '../errors';
import { useCreateCategory, useUpdateCategory } from '../hooks/useEventCategories';
import {
  ADD_CATEGORY,
  CATEGORY_DUPLICATE,
  CATEGORY_NAME_MAX,
  ENGLISH_HELP,
  REQUIRED_CATEGORY_NAME,
  SAVE,
  textTooLong,
} from '../messages';

export interface CategoryFormProps {
  /** Categoría a editar; sin ella, alta. */
  category?: CategoriaAdmin;
  onDone: (message: string) => void;
  onCancel: () => void;
}

/**
 * Alta/edición de una categoría del catálogo (ux.md §4.6, Q6): nombre en
 * español **obligatorio** y en inglés opcional. El catálogo no es contenido
 * publicable —son etiquetas— así que no tiene estado de publicación: el cambio
 * se ve de inmediato en los elementos que la usan (relación por referencia).
 */
export function CategoryForm({ category, onDone, onCancel }: CategoryFormProps) {
  const createCategory = useCreateCategory();
  const updateCategory = useUpdateCategory();
  const isEdit = Boolean(category);

  const [nameEs, setNameEs] = useState(category?.nameEs ?? '');
  const [nameEn, setNameEn] = useState(category?.nameEn ?? '');
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({});

  const mutation = isEdit ? updateCategory : createCategory;
  const busy = createCategory.isPending || updateCategory.isPending;

  const serverErrors: Record<string, string> = {};
  for (const field of ['nameEs', 'nameEn'] as const) {
    const message = detailMessage(mutation.error, field);
    if (message) {
      serverErrors[field] = message;
    }
  }
  const errors = { ...clientErrors, ...serverErrors };
  const duplicate = isConflict(mutation.error);
  const showGeneralError = mutation.error && Object.keys(serverErrors).length === 0 && !duplicate;

  const handleSubmit = (formEvent: React.FormEvent<HTMLFormElement>) => {
    formEvent.preventDefault();
    const nextErrors: Record<string, string> = {};
    if (nameEs.trim() === '') {
      nextErrors.nameEs = REQUIRED_CATEGORY_NAME;
    } else if (nameEs.trim().length > CATEGORY_NAME_MAX) {
      nextErrors.nameEs = textTooLong(CATEGORY_NAME_MAX);
    }
    if (nameEn.trim().length > CATEGORY_NAME_MAX) {
      nextErrors.nameEn = textTooLong(CATEGORY_NAME_MAX);
    }
    if (Object.keys(nextErrors).length > 0) {
      setClientErrors(nextErrors);
      document.getElementById('agenda-category-name-es')?.focus();
      return;
    }
    setClientErrors({});

    const payload: CategoriaInput = {
      nameEs: nameEs.trim(),
      nameEn: nameEn.trim() === '' ? null : nameEn.trim(),
    };

    if (category) {
      updateCategory.mutate(
        { id: category.id, input: payload },
        { onSuccess: () => onDone('Categoría guardada.') },
      );
      return;
    }
    createCategory.mutate(payload, { onSuccess: () => onDone('Categoría creada.') });
  };

  return (
    <Dialog title={isEdit ? 'Editar categoría' : ADD_CATEGORY} onClose={onCancel}>
      <form onSubmit={handleSubmit} noValidate className="space-y-4">
        {duplicate && <Notice variant="error">{CATEGORY_DUPLICATE}</Notice>}
        {showGeneralError && <Notice variant="error">{generalErrorMessage(mutation.error)}</Notice>}

        <Field
          id="agenda-category-name-es"
          label="Nombre (Español)"
          required
          value={nameEs}
          error={errors.nameEs}
          onChange={(inputEvent) => setNameEs(inputEvent.target.value)}
        />
        <Field
          label="Nombre (English, opcional)"
          value={nameEn}
          help={ENGLISH_HELP}
          error={errors.nameEn}
          onChange={(inputEvent) => setNameEn(inputEvent.target.value)}
        />

        <div className="flex flex-wrap justify-end gap-3">
          <Button type="button" variant="secondary" onClick={onCancel} disabled={busy}>
            Cancelar
          </Button>
          <Button type="submit" loading={busy}>
            {isEdit ? SAVE : ADD_CATEGORY}
          </Button>
        </div>
      </form>
    </Dialog>
  );
}
