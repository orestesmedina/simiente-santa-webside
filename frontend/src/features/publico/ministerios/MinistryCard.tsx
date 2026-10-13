import type { MinisterioPublico } from '../../../api/ministerios';
import { Card } from '../../../components/publico/Card';
import { CardImage } from '../../../components/publico/CardImage';
import { EncargadoLine } from './EncargadoLine';
import { WhatsappJoinButton } from './WhatsappJoinButton';

export interface MinistryCardProps {
  /** Ministerio **publicado**, ya resuelto a `lang` por el servidor (FR-008). */
  ministry: MinisterioPublico;
}

/**
 * Ficha pública completa de un ministerio (FR-002/FR-024/FR-026, ux.md §4.1):
 * `Card solid` del sistema de F10 con imagen 16:9 opcional, nombre (h3),
 * descripción **completa** (sin «Ver más»), líneas de horario y lugar —**texto
 * libre** tal como lo escribió el equipo: sin `DateBadge`, sin `<time>` y sin
 * calcular nada (FR-024/SC-015)—, la línea del encargado solo si el servidor
 * publicó su ficha, y el botón de unirse solo si hay enlace.
 *
 * **Cada elemento existe solo si tiene dato** (SC-017): sin imagen no hay hueco
 * ni reserva, sin horario o lugar no hay esa línea y sin enlace no hay botón.
 * La tarjeta **no es un enlace en bloque**: solo el botón es interactivo.
 */
export function MinistryCard({ ministry }: MinistryCardProps) {
  return (
    <Card as="li" padded={false} className="flex flex-col overflow-hidden">
      <CardImage image={ministry.image} />

      <div className="flex flex-1 flex-col gap-3 p-fluid">
        <h3 className="font-sans text-[1.25rem] font-semibold leading-snug text-ink">
          {ministry.name}
        </h3>

        {/* Descripción completa, con los saltos de línea del equipo preservados. */}
        <p className="whitespace-pre-line text-lg leading-relaxed text-ink">
          {ministry.description}
        </p>

        {ministry.schedule && (
          <p className="flex items-start gap-1 text-muted">
            <span aria-hidden="true">🕒</span>
            <span className="[text-wrap:balance]">{ministry.schedule}</span>
          </p>
        )}

        {ministry.place && (
          <p className="flex items-start gap-1 text-muted">
            <span aria-hidden="true">📍</span>
            <span className="[text-wrap:balance]">{ministry.place}</span>
          </p>
        )}

        {ministry.leader && <EncargadoLine leader={ministry.leader} />}

        {ministry.whatsappUrl && (
          <WhatsappJoinButton url={ministry.whatsappUrl} name={ministry.name} />
        )}
      </div>
    </Card>
  );
}
