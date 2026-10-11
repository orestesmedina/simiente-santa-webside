import type { WhatsappChannelAdmin } from '../../../api/portada';
import { usePortadaAdmin } from './usePortadaAdmin';

/** Canales de WhatsApp desde el agregado del panel (FR-011). */
export function useWhatsappChannels(): { items: WhatsappChannelAdmin[] } {
  const { data } = usePortadaAdmin();
  return { items: data?.whatsapp.items ?? [] };
}
