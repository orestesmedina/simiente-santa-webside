import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { uploadMinistryImage, type ImageUploadResult } from '../../../api/ministerios';

/**
 * Sube la imagen de un ministerio (FR-020) al endpoint del módulo (exige el
 * permiso `ministerios`). Devuelve el `fileName` que el formulario guarda; la
 * subida no cambia por sí sola el contenido visible.
 */
export function useUploadMinistryImage() {
  return useMutation<ImageUploadResult, ApiError, File>({
    mutationFn: (file) => uploadMinistryImage(file),
  });
}
