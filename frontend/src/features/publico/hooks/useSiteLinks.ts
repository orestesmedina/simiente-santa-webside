import { useMemo } from 'react';
import type { PortadaPublica } from '../../../api/portada';
import type { PublicMessageKey } from '../i18n/messages';
import { presentSections, type PublicSectionId } from '../sections';
import { usePublicHomeData } from './usePublicHomeData';

/**
 * Enlace navegable del sitio público. Es la pieza común que alimenta **tanto**
 * la cabecera (`PublicNav`/`MobileMenu`) **como** el mapa del sitio del pie
 * (`FooterSiteMap`), de modo que nunca se desincronizan ni en nombres ni en
 * borradores (FR-012/FR-027).
 */
export interface SiteLink {
  /** Identificador estable de la sección (clave React y agrupación). */
  id: PublicSectionId;
  /** `id`/ancla del elemento de sección en la página. */
  anchor: string;
  /** Clave i18n del nombre visible (la misma en cabecera y pie). */
  messageKey: PublicMessageKey;
  /** Destino del enlace; hoy siempre un ancla de la portada. */
  href: string;
}

/**
 * Deriva los enlaces visibles de la respuesta publicada (FR-012/SC-012): una
 * sección sin contenido publicado **no** es enlace (0 enlaces rotos u
 * huérfanos). Es una función pura para poder probarla sin React; el hook la
 * envuelve sobre la consulta de la portada.
 *
 * Preparado para F4–F9: cuando una sección futura tenga ruta propia, su entrada
 * del modelo llevará un `href` de ruta ya implementada; hoy no se inventa ningún
 * enlace (todo `href` es un ancla existente).
 */
export function buildSiteLinks(data: PortadaPublica | undefined): SiteLink[] {
  return presentSections(data).map((section) => ({
    id: section.id,
    anchor: section.anchor,
    messageKey: section.messageKey,
    href: `#${section.anchor}`,
  }));
}

/**
 * Fuente única de los enlaces de navegación del sitio público: secciones con
 * contenido publicado, en el orden de la página y con las claves i18n de la
 * cabecera. La consumen la cabecera y el pie (FR-012/FR-027).
 */
export function useSiteLinks(): SiteLink[] {
  const { data } = usePublicHomeData();
  return useMemo(() => buildSiteLinks(data), [data]);
}
