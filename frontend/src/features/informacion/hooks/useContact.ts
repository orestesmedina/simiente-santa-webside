import type { ContactAdmin } from '../../../api/portada';
import { usePortadaAdmin } from './usePortadaAdmin';

/** Datos de contacto desde el agregado del panel (FR-011). */
export function useContact(): { contact: ContactAdmin | null } {
  const { data } = usePortadaAdmin();
  return { contact: data?.contact ?? null };
}
