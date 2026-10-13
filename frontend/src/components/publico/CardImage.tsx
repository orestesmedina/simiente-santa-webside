import { useState } from 'react';
import type { ImagenPublica } from '../../api/agenda';
import { mediaUrl } from '../../api/client';

export interface CardImageProps {
  image?: ImagenPublica;
}

/**
 * Imagen opcional de la tarjeta (FR-020): a sangre en el marco superior, 16:9 y
 * recorte al centro. Sin imagen la tarjeta **no muestra hueco** y, si la imagen
 * falla, un `onError` la retira sin transformar el layout (la tarjeta conserva
 * el resto de su contenido).
 *
 * Vive en el **inventario único** (`src/components/publico/`, regla de F2/T242):
 * la promovió F6 (T623) desde `features/publico/eventos/ItemCardParts.tsx` para
 * reutilizarla en las fichas y el extracto de ministerios; F4 la sigue usando
 * **sin cambios de props ni de textos**.
 */
export function CardImage({ image }: CardImageProps) {
  const [failed, setFailed] = useState(false);
  const src = mediaUrl(image?.url);
  if (!src || failed) {
    return null;
  }

  return (
    <img
      src={src}
      alt={image?.alt ?? ''}
      onError={() => setFailed(true)}
      className="aspect-video w-full object-cover"
    />
  );
}
