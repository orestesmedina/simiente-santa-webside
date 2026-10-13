import { useMutation } from '@tanstack/react-query';
import {
  uploadActividadImage,
  uploadEventoImage,
  type ImageUploadResult,
} from '../../../api/agenda';
import { ApiError } from '../../../api/client';

/** Módulo al que pertenece la imagen: cada uno tiene su endpoint y su permiso. */
export type AgendaImageModule = 'eventos' | 'actividades';

/**
 * Sube la imagen de un elemento de agenda (FR-020) al endpoint de **su**
 * módulo: una cuenta con solo `eventos` no sube al área de actividades ni al
 * revés (el servidor lo verifica). Devuelve el `fileName` que el formulario
 * guarda; la subida no cambia por sí sola el contenido visible.
 */
export function useUploadAgendaImage(module: AgendaImageModule) {
  return useMutation<ImageUploadResult, ApiError, File>({
    mutationFn: (file) =>
      module === 'eventos' ? uploadEventoImage(file) : uploadActividadImage(file),
  });
}
