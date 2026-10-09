import type { AboutAdmin } from '../../../api/portada';
import { usePortadaAdmin } from './usePortadaAdmin';

/** Texto de «quiénes somos» desde el agregado del panel (FR-011). */
export function useWhoWeAre(): { about: AboutAdmin | null } {
  const { data } = usePortadaAdmin();
  return { about: data?.about ?? null };
}
