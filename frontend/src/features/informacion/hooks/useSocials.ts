import type { SocialLinkAdmin } from '../../../api/portada';
import { usePortadaAdmin } from './usePortadaAdmin';

/** Enlaces de redes desde el agregado del panel (FR-011). */
export function useSocials(): { items: SocialLinkAdmin[] } {
  const { data } = usePortadaAdmin();
  return { items: data?.socials.items ?? [] };
}
