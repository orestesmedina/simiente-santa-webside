import { useQuery } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { getPortada, type PortadaPublica } from '../../../api/portada';
import { useLanguage } from '../i18n/useLanguage';

/** Clave raíz de la consulta pública de la portada (una por idioma). */
export const PUBLIC_HOME_QUERY_KEY = ['portada', 'publica'] as const;

/**
 * Consulta pública de la portada (FR-001). El servidor entrega los contenidos
 * **ya resueltos** al idioma activo (fallback `en → es` por campo, FR-009) y
 * **solo** lo publicado (SC-002); el frontend no reimplementa el fallback ni
 * filtra borradores (analyze I2). Cambiar de idioma dispara una consulta nueva.
 */
export function usePublicHomeData() {
  const { lang } = useLanguage();

  return useQuery<PortadaPublica, ApiError>({
    queryKey: [...PUBLIC_HOME_QUERY_KEY, lang],
    queryFn: () => getPortada(lang),
  });
}
