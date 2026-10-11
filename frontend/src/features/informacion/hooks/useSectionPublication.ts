import { useMemo, useState } from 'react';
import type { UseMutationResult } from '@tanstack/react-query';
import type { ApiError } from '../../../api/client';
import type { PublicationState } from '../../../api/portada';
import type { NoticeVariant } from '../../../components/Notice';
import { detailMessage } from '../errors';
import { PUBLISHED, SAVED, SAVED_PUBLISHED, UNPUBLISHED } from '../messages';

export interface SectionFeedback {
  variant: NoticeVariant;
  message: string;
}

export interface SectionPublication<Input> {
  /** Errores por campo que devolvió el servidor (`details` de `400 invalid`). */
  serverErrors: Record<string, string>;
  feedback: SectionFeedback | null;
  draftNotice: boolean;
  busy: boolean;
  save: (buildInput: (state: PublicationState) => Input, onSaved?: () => void) => void;
  publish: (buildInput: (state: PublicationState) => Input) => void;
  unpublish: (buildInput: (state: PublicationState) => Input) => void;
}

/**
 * Lógica compartida de guardado y publicación de un singleton del panel
 * (FR-013/FR-014/FR-015/FR-017): guarda con el estado actual, publica/retira
 * por sección con `PublishControls` y traduce los `details` del servidor a
 * errores de campo. La usan «quiénes somos» y contacto; cada formulario aporta
 * su `buildInput` (que lee sus propios valores).
 */
export function useSectionPublication<Admin extends { publicationState: PublicationState }, Input>({
  admin,
  mutation,
  serverFieldNames,
}: {
  admin: Admin | null;
  mutation: UseMutationResult<Admin, ApiError, Input>;
  serverFieldNames: readonly string[];
}): SectionPublication<Input> {
  const [feedback, setFeedback] = useState<SectionFeedback | null>(null);
  const [draftNotice, setDraftNotice] = useState(false);

  const serverErrors = useMemo(() => {
    const mapped: Record<string, string> = {};
    if (mutation.error) {
      for (const field of serverFieldNames) {
        const message = detailMessage(mutation.error, field);
        if (message) {
          mapped[field] = message;
        }
      }
    }
    return mapped;
  }, [mutation.error, serverFieldNames]);

  const save = (buildInput: (state: PublicationState) => Input, onSaved?: () => void) => {
    setFeedback(null);
    setDraftNotice(false);
    const wasPublished = admin?.publicationState === 'published';
    const isCreate = admin === null;
    mutation.mutate(buildInput(admin?.publicationState ?? 'draft'), {
      onSuccess: () => {
        setFeedback({ variant: 'success', message: wasPublished ? SAVED_PUBLISHED : SAVED });
        setDraftNotice(isCreate);
        onSaved?.();
      },
    });
  };

  const publish = (buildInput: (state: PublicationState) => Input) => {
    if (!admin) {
      return;
    }
    setFeedback(null);
    setDraftNotice(false);
    mutation.mutate(buildInput('published'), {
      onSuccess: () => setFeedback({ variant: 'success', message: PUBLISHED }),
    });
  };

  const unpublish = (buildInput: (state: PublicationState) => Input) => {
    if (!admin) {
      return;
    }
    setFeedback(null);
    setDraftNotice(false);
    mutation.mutate(buildInput('draft'), {
      onSuccess: () => setFeedback({ variant: 'success', message: UNPUBLISHED }),
    });
  };

  return {
    serverErrors,
    feedback,
    draftNotice,
    busy: mutation.isPending,
    save,
    publish,
    unpublish,
  };
}
