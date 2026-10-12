import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { ApiError } from '../../../../api/client';
import { getProximamente, type ProximamentePublica } from '../../../../api/agenda';
import { useLanguage } from '../../i18n/useLanguage';

/**
 * Clave raíz de la consulta pública de agenda (una entrada por idioma). Es la
 * **única** petición de la sección: alimenta a la vez la página `/eventos`
 * (D-23), el ítem de navegación y el mapa del sitio del pie.
 */
export const UPCOMING_QUERY_KEY = ['proximamente'] as const;

/**
 * «Lo que viene» (FR-001/FR-002): el servidor entrega **solo** lo publicado y
 * con fecha próxima, ya resuelto al idioma activo (fallback `en → es` por
 * campo) y con los bloques vacíos **omitidos** (SC-012). El cliente no filtra
 * ni reimplementa el fallback.
 *
 * Al cambiar de idioma se conserva la respuesta anterior
 * (`placeholderData: keepPreviousData`, patrón de F3/F10): los bloques no se
 * desmontan ni parpadean mientras llega el nuevo idioma.
 */
export function useUpcomingEvents() {
  const { lang } = useLanguage();

  return useQuery<ProximamentePublica, ApiError>({
    queryKey: [...UPCOMING_QUERY_KEY, lang],
    queryFn: () => getProximamente(lang),
    placeholderData: keepPreviousData,
  });
}

/**
 * `true` si el documento público trae al menos un evento o una actividad
 * (FR-022/SC-012). Con solo borradores o vencidos el servidor omite ambos
 * bloques y la sección **no existe** en la navegación.
 */
export function hasUpcomingContent(data: ProximamentePublica | undefined): boolean {
  return (data?.events?.length ?? 0) > 0 || (data?.activities?.length ?? 0) > 0;
}
