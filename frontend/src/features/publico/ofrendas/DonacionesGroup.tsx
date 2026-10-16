import type { MetodoDonacionPublico } from '../../../api/ofrendas';
import { SectionHeading } from '../../../components/publico/SectionHeading';
import type { PublicMessageKey } from '../i18n/messages';
import { useLanguage } from '../i18n/useLanguage';
import { Reveal } from '../motion/Reveal';
import { RevealGroup } from '../motion/RevealGroup';
import { DonacionCard } from './DonacionCard';

export interface DonacionesGroupProps {
  /** `id` base del grupo (para el `id` del encabezado y su `aria-labelledby`). */
  id: string;
  /** Kicker numerado del sistema (`01` IBAN, `02` SINPE). */
  kicker: string;
  /** Clave i18n del rótulo del grupo. */
  titleKey: PublicMessageKey;
  /** Métodos **publicados** del grupo, en el orden que entrega el servidor. */
  items: MetodoDonacionPublico[];
}

/** Grilla de fichas: 1 columna en móvil, 2 desde tableta (ux.md §4.1). */
const GRID = 'grid grid-cols-1 gap-block md:grid-cols-2';

/**
 * Grupo rotulado de la sección `/ofrendas` (FR-004, ux.md §4.1): `SectionHeading`
 * con kicker numerado y grilla de `DonacionCard` con `RevealGroup` (stagger
 * suave, revelado lateral una vez).
 *
 * **Sin fichas el grupo no se dibuja**: cero rótulo huérfano (D-5 de F4). El
 * tipo lo nombra el propio grupo, así que las fichas dentro no repiten chip.
 */
export function DonacionesGroup({ id, kicker, titleKey, items }: DonacionesGroupProps) {
  const { t } = useLanguage();

  if (items.length === 0) {
    return null;
  }

  const titleId = `${id}-title`;

  return (
    <section aria-labelledby={titleId} className="flex flex-col gap-block">
      <SectionHeading id={titleId} kicker={kicker}>
        {t(titleKey)}
      </SectionHeading>
      <RevealGroup>
        <Reveal from="left">
          <ul className={GRID}>
            {items.map((item) => (
              <DonacionCard key={item.id} item={item} />
            ))}
          </ul>
        </Reveal>
      </RevealGroup>
    </section>
  );
}
