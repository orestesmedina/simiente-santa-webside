import { useMutation } from '@tanstack/react-query';
import { ApiError } from '../../../api/client';
import { syncMedios, type ResumenBusqueda } from '../../../api/medios';
import { useInvalidateMedios } from './useMedios';

/**
 * «Buscar episodios nuevos» (FR-011/FR-018): la mutation dispara la búsqueda
 * —que corre **dentro de la misma petición**, sin proceso de fondo—, devuelve el
 * **resumen** (también cuando un feed falló: eso no es error) y refresca el
 * catálogo, el documento público y el bloque de la portada.
 *
 * Quien la usa desactiva la interfaz mientras `isPending` (botón en «Buscando…» y
 * acciones de fila deshabilitadas, D-16) y **no** la repite hasta que termine: una
 * sola búsqueda activa a la vez.
 */
export function useBuscarEpisodios() {
  const invalidate = useInvalidateMedios();

  return useMutation<ResumenBusqueda, ApiError, void>({
    mutationFn: () => syncMedios(),
    onSuccess: invalidate,
  });
}
