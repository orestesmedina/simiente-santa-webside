import type { IdentityAdmin } from '../../../api/portada';
import { usePortadaAdmin } from './usePortadaAdmin';

/**
 * Identidad de la iglesia desde el agregado del panel (FR-011): `null` hasta el
 * primer guardado. Comparte la consulta con `InformationPage` (misma clave), así
 * que no genera una petición extra.
 */
export function useIdentity(): { identity: IdentityAdmin | null } {
  const { data } = usePortadaAdmin();
  return { identity: data?.identity ?? null };
}
