import { useMemo } from 'react';
import type { ProximamentePublica } from '../../../api/agenda';
import type { PortadaPublica } from '../../../api/portada';
import { hasUpcomingContent, useUpcomingEvents } from '../eventos/hooks/useUpcomingEvents';
import type { PublicMessageKey } from '../i18n/messages';
import { presentSections, type PublicSectionId } from '../sections';
import { usePublicHomeData } from './usePublicHomeData';

/**
 * Identificador de un enlace del sitio público: las secciones de la portada
 * (ancla dentro de `/`) y la sección de agenda, que es una **ruta** propia.
 */
export type SiteLinkId = PublicSectionId | 'events';

/**
 * Enlace navegable del sitio público. Es la pieza común que alimenta **tanto**
 * la cabecera (`PublicNav`/`MobileMenu`) **como** el mapa del sitio del pie
 * (`FooterSiteMap`), de modo que nunca se desincronizan ni en nombres ni en
 * borradores (FR-012/FR-027).
 */
export interface SiteLink {
  /** Identificador estable de la sección (clave React y agrupación). */
  id: SiteLinkId;
  /** `id`/ancla del elemento de sección; ausente en los enlaces de ruta. */
  anchor?: string;
  /** Clave i18n del nombre visible (la misma en cabecera y pie). */
  messageKey: PublicMessageKey;
  /**
   * Destino del enlace: `/#ancla` (ancla **absoluta** a la portada, válida
   * desde cualquier ruta) o una ruta propia (`/eventos`).
   */
  href: string;
}

/**
 * Enlace único de la sección de agenda (⟲ D-1): «Eventos y actividades» es una
 * **ruta real** (`/eventos`), no un ancla de la portada, y va al final de la
 * lista (después de las señales de la portada).
 */
const EVENTS_LINK: SiteLink = {
  id: 'events',
  messageKey: 'nav.sections.events',
  href: '/eventos',
};

/**
 * Ruta de la portada: los enlaces de sección son **absolutos** a la raíz
 * (`/#ancla`, QA Bug 1). Como se pintan con `<a>` nativos desde cualquier ruta
 * —incluida `/eventos`—, un ancla relativa (`#ancla`) se resolvería contra la
 * ruta actual (`/eventos#ancla`) y el usuario no navegaría a ninguna parte.
 */
const HOME_PATH = '/';

/**
 * Deriva los enlaces visibles de las dos respuestas públicas (FR-012/FR-022/
 * SC-012): una sección de portada sin contenido publicado **no** es enlace, y
 * la sección de agenda **solo** existe si el documento público trae algún
 * evento o actividad publicado y próximo (con solo borradores o vencidos el
 * servidor omite ambos bloques y el ítem desaparece). Cero enlaces rotos o
 * huérfanos en las tres superficies, que beben de esta única función.
 *
 * Es una función pura para poder probarla sin React; el hook la envuelve sobre
 * las consultas compartidas por `queryClient` (⟲ D-23).
 */
export function buildSiteLinks(
  data: PortadaPublica | undefined,
  upcoming?: ProximamentePublica,
): SiteLink[] {
  const links: SiteLink[] = presentSections(data).map((section) => ({
    id: section.id,
    anchor: section.anchor,
    messageKey: section.messageKey,
    href: `${HOME_PATH}#${section.anchor}`,
  }));

  if (hasUpcomingContent(upcoming)) {
    links.push(EVENTS_LINK);
  }
  return links;
}

/**
 * Fuente única de los enlaces de navegación del sitio público: secciones de la
 * portada con contenido publicado (en el orden de la página) más la sección de
 * agenda cuando existe (⟲ D-23: la presencia se deduce de la misma respuesta
 * pública que pinta `/eventos`, sin pedir el módulo dos veces). La consumen a la
 * vez la cabecera, el menú móvil y el mapa del sitio del pie.
 */
export function useSiteLinks(): SiteLink[] {
  const { data } = usePublicHomeData();
  const { data: upcoming } = useUpcomingEvents();
  return useMemo(() => buildSiteLinks(data, upcoming), [data, upcoming]);
}
