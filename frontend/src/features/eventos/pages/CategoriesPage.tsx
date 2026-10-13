import { useState } from 'react';
import type { CategoriaAdmin } from '../../../api/agenda';
import { Button } from '../../../components/Button';
import { EmptyState } from '../../../components/EmptyState';
import { Notice } from '../../../components/Notice';
import { Table, type Column } from '../../../components/Table';
import { isForbidden } from '../errors';
import { useEventCategories } from '../hooks/useEventCategories';
import {
  ADD_CATEGORY,
  CATEGORIES_HELP,
  CATEGORIES_LOAD_ERROR,
  CATEGORIES_TITLE,
  DELETED,
  DELETE_ACTION,
  EDIT_ACTION,
  NO_ENGLISH_LABEL,
  NO_PERMISSION_ERROR,
  RETRY,
} from '../messages';
import { CategoryForm } from '../components/CategoryForm';
import { DeleteConfirm } from '../components/DeleteConfirm';

/**
 * Catálogo de categorías del panel (`/panel/categorias`, ux.md §4.6, Q6/Q7):
 * listado completo (≤ 100, ordenado por `sortOrder`) con el nombre en español y
 * su versión en inglés cuando existe, alta/edición en `Dialog` y borrado con
 * confirmación. Se autoriza con el permiso propio **`categorias`** (⟲ D-6): ni
 * «eventos» ni «actividades» dan acceso a este catálogo (FR-012).
 */
export function CategoriesPage() {
  const categoriesQuery = useEventCategories();

  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<CategoriaAdmin | null>(null);
  const [deleting, setDeleting] = useState<CategoriaAdmin | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);

  const columns: Column<CategoriaAdmin>[] = [
    {
      key: 'name',
      header: 'Nombre',
      render: (category) => (
        <span className="flex flex-col">
          <span>{category.nameEs}</span>
          <span className="text-sm text-slate-600">{category.nameEn ?? NO_ENGLISH_LABEL}</span>
        </span>
      ),
    },
    { key: 'order', header: 'Orden', render: (category) => String(category.sortOrder) },
    {
      key: 'actions',
      header: 'Acciones',
      render: (category) => (
        <div className="flex flex-wrap items-center gap-2">
          <Button size="sm" variant="secondary" onClick={() => setEditing(category)}>
            {EDIT_ACTION}
          </Button>
          <Button size="sm" variant="danger" onClick={() => setDeleting(category)}>
            {DELETE_ACTION}
          </Button>
        </div>
      ),
    },
  ];

  return (
    <section className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">{CATEGORIES_TITLE}</h1>
        <p className="mt-2 text-slate-700">{CATEGORIES_HELP}</p>
      </div>

      {categoriesQuery.isError && isForbidden(categoriesQuery.error) && (
        <Notice variant="error">{NO_PERMISSION_ERROR}</Notice>
      )}

      {categoriesQuery.isPending && <EmptyState kind="loading" message="Cargando categorías…" />}

      {categoriesQuery.isError && !isForbidden(categoriesQuery.error) && (
        <EmptyState
          kind="error"
          message={CATEGORIES_LOAD_ERROR}
          action={<Button onClick={() => void categoriesQuery.refetch()}>{RETRY}</Button>}
        />
      )}

      {categoriesQuery.isSuccess && (
        <div className="space-y-3">
          {feedback && <Notice variant="success">{feedback}</Notice>}

          {categoriesQuery.data.items.length === 0 ? (
            <EmptyState
              message="Todavía no hay categorías. Agrega la primera para etiquetar los elementos."
              action={<Button onClick={() => setCreating(true)}>{ADD_CATEGORY}</Button>}
            />
          ) : (
            <>
              <div className="flex justify-end">
                <Button onClick={() => setCreating(true)}>{ADD_CATEGORY}</Button>
              </div>
              <Table
                columns={columns}
                rows={categoriesQuery.data.items}
                rowKey={(category) => category.id}
                caption="Catálogo de categorías de agenda"
              />
            </>
          )}
        </div>
      )}

      {creating && (
        <CategoryForm
          onDone={(message) => {
            setCreating(false);
            setFeedback(message);
          }}
          onCancel={() => setCreating(false)}
        />
      )}
      {editing && (
        <CategoryForm
          category={editing}
          onDone={(message) => {
            setEditing(null);
            setFeedback(message);
          }}
          onCancel={() => setEditing(null)}
        />
      )}
      {deleting && (
        <DeleteConfirm
          category={deleting}
          onDeleted={() => {
            setDeleting(null);
            setFeedback(DELETED);
          }}
          onClose={() => setDeleting(null)}
        />
      )}
    </section>
  );
}
