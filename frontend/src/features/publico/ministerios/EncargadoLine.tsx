import type { LiderPublico } from '../../../api/ministerios';
import { useLanguage } from '../i18n/useLanguage';

export interface EncargadoLineProps {
  /** Ficha **publicada** del encargado, ya resuelta por el servidor (FR-026). */
  leader: LiderPublico;
}

/**
 * Línea de la ficha del encargado de una ficha pública (FR-026/SC-017, ux.md
 * §4.1): icono 👤 + **nombre** y, **solo si existe**, el **único** medio de
 * contacto guardado (📱 celular o ✉️ correo). Con ficha publicada y sin contacto
 * se muestra solo el nombre y **no es un error** (Q8).
 *
 * **Sin lógica de visibilidad**: si el interruptor está apagado el servidor
 * **no manda** `leader` y `MinistryCard` no monta este componente (ni línea
 * vacía ni nombre huérfano). El contacto es **texto**, nunca un enlace que el
 * equipo no aprobó (⟲ D-4).
 */
export function EncargadoLine({ leader }: EncargadoLineProps) {
  const { t } = useLanguage();

  return (
    <div className="flex flex-col gap-1 text-muted">
      <p className="flex items-center gap-1">
        <span aria-hidden="true">👤</span>
        <span className="sr-only">{t('ministries.leader.label')}: </span>
        <span>{leader.name}</span>
      </p>
      {leader.phone && (
        <p className="flex items-center gap-1">
          <span aria-hidden="true">📱</span>
          <span>{leader.phone}</span>
        </p>
      )}
      {leader.email && (
        <p className="flex items-center gap-1">
          <span aria-hidden="true">✉️</span>
          <span>{leader.email}</span>
        </p>
      )}
    </div>
  );
}
