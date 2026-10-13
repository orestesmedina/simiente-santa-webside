import { PublicButton } from '../../../components/publico/PublicButton';
import { useLanguage } from '../i18n/useLanguage';

export interface MapLinkProps {
  /** Enlace de mapa del elemento (solo se valida el formato; no se verifica el destino). */
  href: string;
  /** Título del elemento, para el nombre accesible completo (⟲ D-20). */
  title: string;
}

/**
 * Enlace «Ver ubicación» (ux.md §4.1, FR-024): abre el mapa en pestaña nueva
 * con `rel="noopener noreferrer"` y lleva **icono + texto** (nunca un icono
 * solo) más el nombre accesible completo «Ver ubicación de "{título}"» /
 * «See location for "{title}"» (⟲ D-20), de modo que un lector de pantalla lo
 * entienda sin depender del contexto del DOM.
 */
export function MapLink({ href, title }: MapLinkProps) {
  const { t, lang } = useLanguage();
  const accessibleName =
    lang === 'en' ? `See location for "${title}"` : `Ver ubicación de "${title}"`;

  return (
    <PublicButton variant="secondary" href={href} external ariaLabel={accessibleName}>
      <span aria-hidden="true">📍</span>
      {t('events.location.link')}
    </PublicButton>
  );
}
