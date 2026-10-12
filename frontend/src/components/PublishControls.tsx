import { useState } from 'react';
import type { PublicationState } from '../api/portada';
import { Button } from './Button';
import { ConfirmDialog } from './ConfirmDialog';
import {
  UNPUBLISH_CANCEL,
  UNPUBLISH_CONFIRM_ACTION,
  UNPUBLISH_CONFIRM_TITLE,
  unpublishConfirmDescription,
} from '../features/informacion/messages';

export interface PublishControlsProps {
  /** Nombre legible del elemento o sección, para la confirmación de retiro. */
  name: string;
  state: PublicationState;
  loading?: boolean;
  onPublish: () => void;
  onUnpublish: () => void;
}

/**
 * Publicar / retirar de la portada (ux.md §5.c, FR-013): **publicar** es una
 * acción directa, sin confirmación; **retirar** pide una confirmación que
 * explica que los datos se conservan. Se usa por sección (singletons) y por
 * elemento (listados), nunca agrupando varias secciones.
 *
 * Vive en el **inventario único** (`src/components/`, regla de F2/T242): lo usan
 * la portada de F3 y los elementos de agenda de F4 (promoción T431, sin cambios
 * de props ni de textos por defecto).
 */
export function PublishControls({
  name,
  state,
  loading = false,
  onPublish,
  onUnpublish,
}: PublishControlsProps) {
  const [confirming, setConfirming] = useState(false);

  if (state === 'published') {
    return (
      <>
        <Button variant="secondary" disabled={loading} onClick={() => setConfirming(true)}>
          Retirar de la portada
        </Button>
        {confirming && (
          <ConfirmDialog
            title={UNPUBLISH_CONFIRM_TITLE}
            description={unpublishConfirmDescription(name)}
            confirmText={UNPUBLISH_CONFIRM_ACTION}
            cancelText={UNPUBLISH_CANCEL}
            danger
            loading={loading}
            onConfirm={() => {
              setConfirming(false);
              onUnpublish();
            }}
            onClose={() => setConfirming(false)}
          />
        )}
      </>
    );
  }

  return (
    <Button loading={loading} onClick={onPublish}>
      Publicar
    </Button>
  );
}
