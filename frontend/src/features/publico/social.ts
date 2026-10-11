import type { SocialNetwork } from '../../api/portada';

/**
 * Nombre visible de cada red del catálogo fijo (FR-006). Son nombres propios:
 * no se traducen; la cadena que los rodea («Ver en {red}» / «Visit us on
 * {network}») sí sale del diccionario de interfaz.
 */
export const SOCIAL_NETWORK_LABELS: Record<SocialNetwork, string> = {
  facebook: 'Facebook',
  instagram: 'Instagram',
  youtube: 'YouTube',
  tiktok: 'TikTok',
  spotify: 'Spotify',
};

/** Orden de presentación del catálogo de redes en el sitio público. */
export const SOCIAL_NETWORKS: readonly SocialNetwork[] = [
  'facebook',
  'instagram',
  'youtube',
  'tiktok',
  'spotify',
];
