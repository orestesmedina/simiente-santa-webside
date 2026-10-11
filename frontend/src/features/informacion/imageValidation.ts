import { imageInvalidMessage } from './messages';

/** Límite del contrato (`UPLOAD_MAX_BYTES`, T312): 8 MB. */
export const MAX_IMAGE_BYTES = 8 * 1024 * 1024;

/** Tipos aceptados por el contrato (firma binaria en el servidor). */
export const ACCEPTED_IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/webp'] as const;

/** Tamaño legible en MB para los mensajes («8 MB»). */
export function formatFileSize(bytes: number): string {
  return `${Math.round(bytes / (1024 * 1024))} MB`;
}

/**
 * Validación de cliente del archivo (ux.md §10: tipo y tamaño **antes** de
 * enviar). Es solo experiencia de usuario: la autoridad es la firma binaria del
 * servidor. Devuelve el mensaje de error, o `undefined` si el archivo es válido.
 */
export function validateImage(file: File): string | undefined {
  const max = formatFileSize(MAX_IMAGE_BYTES);
  if (!(ACCEPTED_IMAGE_TYPES as readonly string[]).includes(file.type)) {
    return imageInvalidMessage(max);
  }
  if (file.size > MAX_IMAGE_BYTES) {
    return imageInvalidMessage(max);
  }
  return undefined;
}
