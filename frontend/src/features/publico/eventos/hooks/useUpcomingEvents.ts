import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { useState } from 'react';
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

  const query = useQuery<ProximamentePublica, ApiError>({
    queryKey: [...UPCOMING_QUERY_KEY, lang],
    queryFn: () => getProximamente(lang),
    placeholderData: keepPreviousData,
  });

  // Última versión **válida** vista (ux.md §4.1: «error con caché previa»). El
  // `placeholderData` solo cubre la recarga en curso; si la recarga falla, el
  // contenido ya conocido se conserva y la página solo añade un aviso discreto.
  // Ajuste de estado durante el render (patrón de React para estado derivado,
  // el mismo de `Reveal`/`IdentityForm`): sin efecto ni render en cascada.
  const [retained, setRetained] = useState<ProximamentePublica>();
  if (query.data && query.data !== retained) {
    setRetained(query.data);
  }

  return { ...query, data: query.data ?? retained };
}

/**
 * `true` si el documento público trae al menos un evento o una actividad
 * (FR-022/SC-012). Con solo borradores o vencidos el servidor omite ambos
 * bloques y la sección **no existe** en la navegación.
 */
export function hasUpcomingContent(data: ProximamentePublica | undefined): boolean {
  return (data?.events?.length ?? 0) > 0 || (data?.activities?.length ?? 0) > 0;
}
