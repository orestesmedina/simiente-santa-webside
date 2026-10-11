import { useState } from 'react';

/**
 * Estado local de un formulario que se reajusta cuando cambia la revisión de
 * los datos del servidor (carga inicial, guardado, publicación/retiro), sin un
 * efecto que provoque renders en cascada. El llamador pasa una `revision` (p.
 * ej. `updatedAt`) y una fábrica de valores iniciales.
 */
export function useSyncedFields<T>(
  revision: string | null,
  build: () => T,
): [T, (patch: Partial<T>) => void] {
  const [fields, setFields] = useState<T>(build);
  const [synced, setSynced] = useState(revision);

  if (revision !== synced) {
    setSynced(revision);
    setFields(build());
  }

  const update = (patch: Partial<T>) => {
    setFields((current) => ({ ...current, ...patch }));
  };

  return [fields, update];
}
