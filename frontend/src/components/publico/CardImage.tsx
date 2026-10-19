import { useState, type ReactNode } from 'react';
import type { ImagenPublica } from '../../api/agenda';
import { mediaUrl } from '../../api/client';

export interface CardImageProps {
  image?: ImagenPublica;
  /**
   * Qué pintar cuando **no hay** imagen o cuando la copia local falla al cargar
   * (`onError`). Por defecto nada: la tarjeta se queda sin marco de imagen y
   * conserva el resto de su contenido (comportamiento de F4/F6).
   *
   * F9 (T933) lo usa para el **marcador neutro** de un episodio sin carátula
   * (⟲ D-4): mismo marco 16:9 con la identidad de la iglesia, de modo que la
   * grilla no «salte» de forma ni quede un hueco, tampoco si la copia falla.
   */
  fallback?: ReactNode;
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
export function CardImage({ image, fallback }: CardImageProps) {
  const [failed, setFailed] = useState(false);
  const src = mediaUrl(image?.url);
  if (!src || failed) {
    return fallback ?? null;
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
