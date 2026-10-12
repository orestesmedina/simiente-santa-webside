import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createCategoria,
  deleteCategoria,
  listCategorias,
  updateCategoria,
  type CategoriaAdmin,
  type CategoriaAdminList,
  type CategoriaInput,
  type CategoriaPatch,
} from '../../../api/agenda';
import { ApiError } from '../../../api/client';
import { AGENDA_CATEGORIES_QUERY_KEY, UPCOMING_QUERY_KEY } from './queryKeys';

/** Invalida el catálogo y la sección pública (el nombre viaja con cada elemento). */
function useInvalidateCategories() {
  const queryClient = useQueryClient();
  return () => {
    void queryClient.invalidateQueries({ queryKey: AGENDA_CATEGORIES_QUERY_KEY });
    void queryClient.invalidateQueries({ queryKey: UPCOMING_QUERY_KEY });
  };
}

/**
 * Catálogo compartido de categorías (Q6). Alimenta el `Select` de los
 * formularios y el listado de `/panel/categorias`. Requiere el permiso
 * `categorias` (FR-012).
 */
export function useEventCategories() {
  return useQuery<CategoriaAdminList, ApiError>({
    queryKey: AGENDA_CATEGORIES_QUERY_KEY,
    queryFn: () => listCategorias(),
  });
}

/** Alta en el catálogo (Q6); nombre duplicado → `409`. */
export function useCreateCategory() {
  const invalidate = useInvalidateCategories();

  return useMutation<CategoriaAdmin, ApiError, CategoriaInput>({
    mutationFn: (input) => createCategoria(input),
    onSuccess: invalidate,
  });
}

/** Renombra o reordena una categoría (el cambio se ve de inmediato en lo publicado). */
export function useUpdateCategory() {
  const invalidate = useInvalidateCategories();

  return useMutation<CategoriaAdmin, ApiError, { id: string; input: CategoriaPatch }>({
    mutationFn: ({ id, input }) => updateCategoria(id, input),
    onSuccess: invalidate,
  });
}

/**
 * Elimina una categoría. En uso → `409 conflict` (decisión firme P4-6): el
 * error se muestra tal cual y **no** se desliga de los elementos.
 */
export function useDeleteCategory() {
  const invalidate = useInvalidateCategories();

  return useMutation<void, ApiError, string>({
    mutationFn: (id) => deleteCategoria(id),
    onSuccess: invalidate,
  });
}
