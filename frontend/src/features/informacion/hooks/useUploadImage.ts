import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { uploadImage, type ImageUploadResult } from '../../../api/portada';

/**
 * Sube el logotipo o la imagen de portada (FR-002/FR-011). Devuelve el nombre
 * del archivo generado por el servidor para referenciarlo desde la identidad;
 * la subida **no** cambia por sí sola el contenido visible.
 */
export function useUploadImage() {
  return useMutation<ImageUploadResult, ApiError, File>({
    mutationFn: (file) => uploadImage(file),
  });
}
