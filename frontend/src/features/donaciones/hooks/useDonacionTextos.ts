import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { ApiError } from '../../../api/client';
import { getDonacionTextos, type DonacionTextosAdmin } from '../../../api/ofrendas';
import { DONACIONES_TEXTOS_QUERY_KEY } from './queryKeys';

/**
 * Zona de textos del panel (FR-021/Q5/Q11): los dos singletons (`intro` y
 * `usage`) con sus propios estados; cada uno viene en `null` hasta el primer
 * guardado (patrón de `GET /api/v1/admin/portada` de F3). Conserva la última
 * versión **válida** ante un error (ux.md §5.a: aviso discreto sin perder el
 * contenido ya conocido).
 */
export function useDonacionTextos() {
  const query = useQuery<DonacionTextosAdmin, ApiError>({
    queryKey: DONACIONES_TEXTOS_QUERY_KEY,
    queryFn: () => getDonacionTextos(),
  });

  const [retained, setRetained] = useState<DonacionTextosAdmin>();
  if (query.data && query.data !== retained) {
    setRetained(query.data);
  }

  return { ...query, data: query.data ?? retained };
}
