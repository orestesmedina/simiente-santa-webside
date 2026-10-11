import type { PortadaPublica } from '../../api/portada';
import type { PublicMessageKey } from './i18n/messages';

/**
 * Secciones navegables de la portada (ux.md §2.1/§4.1). El orden es el de la
 * página: quiénes somos · horario · WhatsApp · contacto · redes. La identidad
 * (hero) no es un ancla: encabeza la página.
 *
 * SC-012: una sección **solo existe** si tiene al menos un elemento publicado;
 * `presentSections` decide qué anclas de la cabecera y qué secciones del cuerpo
 * se pintan, de modo que nunca quede un título sin contenido.
 */
export type PublicSectionId = 'who' | 'schedule' | 'whatsapp' | 'contact' | 'social';

export interface PublicSection {
  id: PublicSectionId;
  /** `id` del elemento HTML de la sección, destino del enlace de ancla. */
  anchor: string;
  /** Clave i18n del título de la sección. */
  messageKey: PublicMessageKey;
}

export const PUBLIC_SECTIONS: readonly PublicSection[] = [
  { id: 'who', anchor: 'quienes-somos', messageKey: 'nav.sections.who' },
  { id: 'schedule', anchor: 'horario', messageKey: 'nav.sections.schedule' },
  { id: 'whatsapp', anchor: 'whatsapp', messageKey: 'nav.sections.whatsapp' },
  { id: 'contact', anchor: 'contacto', messageKey: 'nav.sections.contact' },
  { id: 'social', anchor: 'redes', messageKey: 'nav.sections.social' },
];

/**
 * `true` si la sección tiene contenido **publicado** en la respuesta ya
 * resuelta por el servidor (FR-013/SC-012). Las secciones sin publicados no
 * aparecen ni en el cuerpo ni en la navegación.
 */
export function isSectionPresent(id: PublicSectionId, data: PortadaPublica): boolean {
  switch (id) {
    case 'who':
      return Boolean(data.about);
    case 'schedule':
      return (data.schedule?.length ?? 0) > 0;
    case 'whatsapp':
      return (data.whatsapp?.length ?? 0) > 0;
    case 'contact':
      return Boolean(data.contact);
    case 'social':
      return (data.socials?.length ?? 0) > 0;
  }
}

/** Secciones presentes en la respuesta, en el orden de la página. */
export function presentSections(data: PortadaPublica | undefined): PublicSection[] {
  if (!data) {
    return [];
  }
  return PUBLIC_SECTIONS.filter((section) => isSectionPresent(section.id, data));
}
