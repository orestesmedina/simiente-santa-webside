import type { ScheduleItemAdmin } from '../../../api/portada';
import { usePortadaAdmin } from './usePortadaAdmin';

/** Servicios del horario desde el agregado del panel (FR-011). */
export function useServices(): { items: ScheduleItemAdmin[] } {
  const { data } = usePortadaAdmin();
  return { items: data?.schedule.items ?? [] };
}
