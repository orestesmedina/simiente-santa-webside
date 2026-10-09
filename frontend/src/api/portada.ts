import { apiFetch, buildUrl } from './client';
import type { components } from './schema';

/** Idioma soportado por la portada pública (FR-008/FR-009). */
export type PortadaLang = 'es' | 'en';

export type PortadaPublica = components['schemas']['PortadaPublica'];
export type PortadaAdmin = components['schemas']['PortadaAdmin'];
export type PublicationState = components['schemas']['PublicationState'];

export type IdentityAdmin = components['schemas']['IdentityAdmin'];
export type IdentityInput = components['schemas']['IdentityInput'];
export type AboutAdmin = components['schemas']['AboutAdmin'];
export type AboutInput = components['schemas']['AboutInput'];
export type ContactAdmin = components['schemas']['ContactAdmin'];
export type ContactInput = components['schemas']['ContactInput'];

export type ScheduleItemAdmin = components['schemas']['ScheduleItemAdmin'];
export type ScheduleItemInput = components['schemas']['ScheduleItemInput'];
export type ScheduleItemPatch = components['schemas']['ScheduleItemPatch'];

export type WhatsappChannelAdmin = components['schemas']['WhatsappChannelAdmin'];
export type WhatsappChannelInput = components['schemas']['WhatsappChannelInput'];
export type WhatsappChannelPatch = components['schemas']['WhatsappChannelPatch'];

export type SocialLinkAdmin = components['schemas']['SocialLinkAdmin'];
export type SocialLinkInput = components['schemas']['SocialLinkInput'];
export type SocialLinkPatch = components['schemas']['SocialLinkPatch'];
export type SocialNetwork = components['schemas']['SocialNetwork'];

export type ImageUploadResult = components['schemas']['ImageUploadResult'];

/** Raíz del módulo de panel (P3-11). */
const ADMIN_BASE = '/api/v1/admin/portada';

const JSON_HEADERS = { 'Content-Type': 'application/json' } as const;

/**
 * Portada pública ya resuelta por el servidor al idioma pedido (FR-001/FR-009):
 * cada campo traducible usa el inglés si existe y, si no, el español. Nunca
 * incluye borradores (SC-002) y omite las secciones vacías (SC-012). Es de sola
 * lectura y no requiere sesión.
 */
export async function getPortada(lang: PortadaLang): Promise<PortadaPublica> {
  return apiFetch<PortadaPublica>(buildUrl('/api/v1/portada', { lang }));
}

/**
 * Estado completo del módulo para el panel (FR-011): ambos idiomas y
 * `publicationState` por elemento, incluidos los borradores. Requiere el
 * permiso `portada`.
 */
export async function getPortadaAdmin(): Promise<PortadaAdmin> {
  return apiFetch<PortadaAdmin>(ADMIN_BASE);
}

/** Guarda (crea o reemplaza) la identidad de la iglesia (FR-002). */
export async function updateIdentity(input: IdentityInput): Promise<IdentityAdmin> {
  return apiFetch<IdentityAdmin>(`${ADMIN_BASE}/identidad`, {
    method: 'PUT',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Guarda (crea o reemplaza) el texto de «quiénes somos» (FR-003). */
export async function updateAbout(input: AboutInput): Promise<AboutAdmin> {
  return apiFetch<AboutAdmin>(`${ADMIN_BASE}/quienes-somos`, {
    method: 'PUT',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Guarda (crea o reemplaza) los datos de contacto (FR-007). */
export async function updateContact(input: ContactInput): Promise<ContactAdmin> {
  return apiFetch<ContactAdmin>(`${ADMIN_BASE}/contacto`, {
    method: 'PUT',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Crea un servicio del horario (FR-004). Nace en borrador salvo estado explícito. */
export async function createService(input: ScheduleItemInput): Promise<ScheduleItemAdmin> {
  return apiFetch<ScheduleItemAdmin>(`${ADMIN_BASE}/horario`, {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Edita un servicio del horario y/o lo publica o retira (FR-004/FR-013). */
export async function updateService(
  id: string,
  input: ScheduleItemPatch,
): Promise<ScheduleItemAdmin> {
  return apiFetch<ScheduleItemAdmin>(`${ADMIN_BASE}/horario/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Elimina un servicio del horario (FR-004). */
export async function deleteService(id: string): Promise<void> {
  return apiFetch<void>(`${ADMIN_BASE}/horario/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

/** Crea un canal de WhatsApp (FR-005). */
export async function createWhatsappChannel(
  input: WhatsappChannelInput,
): Promise<WhatsappChannelAdmin> {
  return apiFetch<WhatsappChannelAdmin>(`${ADMIN_BASE}/whatsapp`, {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Edita un canal de WhatsApp y/o lo publica o retira (FR-005/FR-013). */
export async function updateWhatsappChannel(
  id: string,
  input: WhatsappChannelPatch,
): Promise<WhatsappChannelAdmin> {
  return apiFetch<WhatsappChannelAdmin>(`${ADMIN_BASE}/whatsapp/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Elimina un canal de WhatsApp (FR-005). */
export async function deleteWhatsappChannel(id: string): Promise<void> {
  return apiFetch<void>(`${ADMIN_BASE}/whatsapp/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

/** Crea el enlace de una red social (FR-006). Un segundo enlace de la misma red → 409. */
export async function createSocialLink(input: SocialLinkInput): Promise<SocialLinkAdmin> {
  return apiFetch<SocialLinkAdmin>(`${ADMIN_BASE}/redes`, {
    method: 'POST',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Edita el enlace de una red social y/o lo publica o retira (FR-006/FR-013). */
export async function updateSocialLink(
  id: string,
  input: SocialLinkPatch,
): Promise<SocialLinkAdmin> {
  return apiFetch<SocialLinkAdmin>(`${ADMIN_BASE}/redes/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    headers: JSON_HEADERS,
    body: JSON.stringify(input),
  });
}

/** Elimina el enlace de una red social (FR-006). */
export async function deleteSocialLink(id: string): Promise<void> {
  return apiFetch<void>(`${ADMIN_BASE}/redes/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

/**
 * Sube el logotipo o la imagen de portada (FR-002/FR-011). Envía
 * `multipart/form-data` con el campo `file`; el servidor genera el nombre y
 * detecta el tipo por firma binaria. No cambia contenido visible por sí sola.
 */
export async function uploadImage(file: File): Promise<ImageUploadResult> {
  const body = new FormData();
  body.append('file', file);
  return apiFetch<ImageUploadResult>(`${ADMIN_BASE}/imagenes`, { method: 'POST', body });
}
