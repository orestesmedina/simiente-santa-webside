import { useMemo } from 'react';
import type { ProximamentePublica } from '../../../api/agenda';
import type { MediosPublicos } from '../../../api/medios';
import type { PortadaMinisterios } from '../../../api/ministerios';
import type { OfrendasPublicas } from '../../../api/ofrendas';
import type { PortadaPublica } from '../../../api/portada';
import { hasUpcomingContent, useUpcomingEvents } from '../eventos/hooks/useUpcomingEvents';
import type { PublicMessageKey } from '../i18n/messages';
import {
  hasPortadaMinistryContent,
  usePortadaMinisterios,
} from '../ministerios/hooks/usePortadaMinisterios';
import { hasPublicMediosContent, usePublicMedios } from '../medios/hooks/usePublicMedios';
import { hasPublicOfrendasContent, usePublicOfrendas } from '../ofrendas/hooks/usePublicOfrendas';
import { presentSections, type PublicSectionId } from '../sections';
import { usePublicHomeData } from './usePublicHomeData';

/**
 * Identificador de un enlace del sitio público: las secciones de la portada
 * (ancla dentro de `/`) y las secciones que son una **ruta** propia (agenda,
 * ministerios, ofrendas y medios).
 */
export type SiteLinkId = PublicSectionId | 'events' | 'ministries' | 'offerings' | 'media';

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
 * Enlace único de la sección de ministerios (F6/⟲ D-1): «Ministerios» también es
 * una **ruta real** (`/ministerios`) y va **al final** de la lista, después de
 * «Eventos y actividades», respetando la secuencia de la portada (los bloques
 * llevan a la ruta al final).
 */
const MINISTRIES_LINK: SiteLink = {
  id: 'ministries',
  messageKey: 'nav.sections.ministries',
  href: '/ministerios',
};

/**
 * Enlace único de la sección de ofrendas (F7/⟲ D-1): «Ofrendas y diezmos» es una
 * **ruta real** (`/ofrendas`) y va **al final** de la lista, después de
 * «Ministerios».
 */
const OFFERINGS_LINK: SiteLink = {
  id: 'offerings',
  messageKey: 'nav.sections.offerings',
  href: '/ofrendas',
};

/**
 * Enlace único de la sección de medios (F9/P9-16, ⟲ D-1): «Medios» es una
 * **ruta real** (`/medios`) y va **al final** de la lista, después de «Ofrendas
 * y diezmos». La clave i18n es `nav.sections.media` (ux.md §7.2) y la consumen a
 * la vez `PublicNav`, `MobileMenu` y `FooterSiteMap`.
 */
const MEDIA_LINK: SiteLink = {
  id: 'media',
  messageKey: 'nav.sections.media',
  href: '/medios',
};

/**
 * Ruta de la portada: los enlaces de sección son **absolutos** a la raíz
 * (`/#ancla`, QA Bug 1). Como se pintan con `<a>` nativos desde cualquier ruta
 * —incluida `/eventos`—, un ancla relativa (`#ancla`) se resolvería contra la
 * ruta actual (`/eventos#ancla`) y el usuario no navegaría a ninguna parte.
 */
const HOME_PATH = '/';

/**
 * Deriva los enlaces visibles de las respuestas públicas (FR-012/FR-022/
 * SC-012): una sección de portada sin contenido publicado **no** es enlace, la
 * sección de agenda **solo** existe si el documento público trae algún evento o
 * actividad publicado y próximo, la sección de ministerios **solo** si el
 * documento del **bloque de la portada** trae al menos un ministerio publicado
 * (P6-9/T628), la de ofrendas **solo** si el **documento de la sección**
 * (`usePublicOfrendas`) trae métodos **o** textos publicados (P7-8/P7-13/T727) y
 * la de medios **solo** si el **documento de la sección** (`usePublicMedios`)
 * trae al menos un episodio publicable (P9-16/T934). Cero enlaces rotos o
 * huérfanos en las tres superficies, que beben de esta única función.
 *
 * Es una función pura para poder probarla sin React; el hook la envuelve sobre
 * las consultas compartidas por `queryClient`.
 */
export function buildSiteLinks(
  data: PortadaPublica | undefined,
  upcoming?: ProximamentePublica,
  portadaMinisterios?: PortadaMinisterios,
  ofrendas?: OfrendasPublicas,
  medios?: MediosPublicos,
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
  if (hasPortadaMinistryContent(portadaMinisterios)) {
    links.push(MINISTRIES_LINK);
  }
  // P7-8: la presencia de «Ofrendas y diezmos» sale del **documento de la
  // sección** (métodos **o** textos publicados), no del bloque de la portada.
  if (hasPublicOfrendasContent(ofrendas)) {
    links.push(OFFERINGS_LINK);
  }
  // P9-16/P7-8: la presencia de «Medios» sale del **documento de la sección**
  // (`usePublicMedios`, compartido por `queryClient` con la página `/medios`):
  // existe solo con `total > 0` publicables. El bloque de la portada
  // (`usePortadaMedios`) **no** es fuente de navegación.
  if (hasPublicMediosContent(medios)) {
    links.push(MEDIA_LINK);
  }
  return links;
}

/**
 * Fuente única de los enlaces de navegación del sitio público: secciones de la
 * portada con contenido publicado (en el orden de la página) más las secciones
 * con ruta propia cuando existen. La presencia de «Eventos y actividades» sale
 * del documento público de la agenda, la de «Ministerios» del documento del
 * **bloque de la portada** (`usePortadaMinisterios`, compartida por
 * `queryClient` con `MinisteriosBlock`: una sola petición paga bloque + barra +
 * menú + pie) y la de «Ofrendas y diezmos» del **documento de la sección**
 * (`usePublicOfrendas`, P7-8/T727 —compartida por `queryClient` con la página
 * `/ofrendas`—: existe con ≥1 método **o** ≥1 texto publicado, P7-13) y la de
 * «Medios» del **documento de la sección** (`usePublicMedios`, P9-16/T934
 * —compartida por `queryClient` con la página `/medios`—: existe solo con
 * `total > 0` publicables). La consumen a la vez la cabecera, el menú móvil y el
 * mapa del sitio del pie.
 *
 * **No** consume la consulta de la sección `/ministerios` (`usePublicMinistries`):
 * esa decide **su** redirección (P6-21) y son dos consultas a propósito. El
 * bloque `usePortadaOfrendas` tampoco alimenta esta lista (FR-022: solo mira
 * métodos) y el bloque `usePortadaMedios` **no** es fuente de navegación (P9-16:
 * esa sale del documento de la sección).
 */
export function useSiteLinks(): SiteLink[] {
  const { data } = usePublicHomeData();
  const { data: upcoming } = useUpcomingEvents();
  const { data: portadaMinisterios } = usePortadaMinisterios();
  const { data: ofrendas } = usePublicOfrendas();
  const { data: medios } = usePublicMedios();
  return useMemo(
    () => buildSiteLinks(data, upcoming, portadaMinisterios, ofrendas, medios),
    [data, upcoming, portadaMinisterios, ofrendas, medios],
  );
}
