import type { PublicMessageKey } from '../i18n/messages';
import type { LanguageContextValue } from '../i18n/useLanguage';

/**
 * Anuncio completo de un enlace externo de medios (FR-005/SC-012).
 *
 * Las plantillas del catálogo (ux.md §7.2) interpolan el nombre del episodio con
 * **dos** nombres de variable distintos —`{nombre}` en español y `{name}` en
 * inglés—, así que se pasan los dos: el mismo componente sirve en ambos idiomas
 * sin reimplementar el fallback y sin que quede ninguna variable sin sustituir.
 *
 * Vive en su propio módulo (y no junto a un componente) porque lo usan
 * `MedioLinks` (T933) y `MedioCard`/`MedioItem` (T933/T935).
 */
export function announceTitle(
  t: LanguageContextValue['t'],
  key: PublicMessageKey,
  title: string,
): string {
  return t(key, { nombre: title, name: title });
}
