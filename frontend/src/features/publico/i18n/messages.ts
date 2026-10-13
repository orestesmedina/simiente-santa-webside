/**
 * Catálogo de claves de la interfaz del sitio público (ux.md §7.2).
 *
 * El tipo `PublicMessageKey` es la fuente única de las claves: los diccionarios
 * `es`/`en` se tipan como `Record<PublicMessageKey, string>`, de modo que si
 * falta una traducción en un idioma **no compila** (SC-006 garantizado en
 * compilación, R3-9). El contenido editorial de la iglesia no vive aquí: llega
 * ya resuelto del servidor (R3-2/analyze I2).
 */
export const publicMessageKeys = [
  'nav.label',
  'nav.menu.open',
  'nav.menu.close',
  'nav.menu.label',
  'nav.sections.who',
  'nav.sections.schedule',
  'nav.sections.whatsapp',
  'nav.sections.contact',
  'nav.sections.social',
  'nav.sections.events',
  'home.upcoming.title',
  'home.upcoming.subtitle',
  'home.upcoming.seeAll',
  'home.upcoming.empty',
  'home.upcoming.error.load',
  'home.upcoming.error.retry',
  'home.upcoming.badge.event',
  'home.upcoming.badge.activity',
  'events.page.title',
  'events.page.subtitle',
  'events.block.scheduled',
  'events.block.activities',
  'events.location.label',
  'events.location.link',
  'events.date.next',
  'events.date.more',
  'events.empty.hidden',
  'events.category.aria',
  'events.loading',
  'events.error.load',
  'events.error.retry',
  'events.error.partial',
  'footer.sitemap.label',
  'footer.sitemap.explore',
  'footer.sitemap.contact',
  'footer.sitemap.social',
  'footer.sitemap.home',
  'theme.label',
  'theme.toDark',
  'theme.toLight',
  'lang.label',
  'lang.es',
  'lang.en',
  'identity.mission',
  'identity.vision',
  'identity.lema',
  'hero.cta.schedule',
  'hero.cta.whatsapp',
  'contact.phone',
  'contact.email',
  'contact.address',
  'whatsapp.action.dm',
  'whatsapp.action.group',
  'social.action',
  'brand.name',
  'footer.welcome',
  'schedule.day.0',
  'schedule.day.1',
  'schedule.day.2',
  'schedule.day.3',
  'schedule.day.4',
  'schedule.day.5',
  'schedule.day.6',
  'schedule.col.day',
  'schedule.col.time',
  'schedule.col.service',
  'schedule.col.place',
  'schedule.time.am',
  'schedule.time.pm',
  'schedule.time.noon',
  'error.load',
  'error.retry',
  'error.partial',
  'status.loading',
  'a11y.skip',
  'a11y.logo',
  'a11y.hero',
  'a11y.events.heading',
  'a11y.home.upcoming.heading',
] as const;

export type PublicMessageKey = (typeof publicMessageKeys)[number];

/** Diccionario completo en un idioma: todas las claves, sin excepción. */
export type PublicMessages = Record<PublicMessageKey, string>;

/** Idiomas de la interfaz pública. Español es el idioma base (FR-010). */
export type Language = 'es' | 'en';

/** Clave de `localStorage` que recuerda la elección entre visitas (R3-9). */
export const LANGUAGE_STORAGE_KEY = 'ss.lang';
