import type { MedioPublico } from '../../../api/medios';
import { useLanguage } from '../i18n/useLanguage';
import { Reveal } from '../motion/Reveal';
import { RevealGroup } from '../motion/RevealGroup';
import { MedioCard } from './MedioCard';

/** Una sola lista: 1 columna en móvil y 2 desde tableta (ux.md §4.1, ⟲ D-13). */
const GRID = 'grid grid-cols-1 gap-block md:grid-cols-2';

export interface MediosGridProps {
  /** Fichas publicables en el orden del servidor (fecha desc.; el cliente no reordena). */
  items: MedioPublico[];
}

/**
 * Lista de fichas de la sección `/medios` (ux.md §4.1): `<ul>` semántica con el
 * nombre accesible de la lista, grilla de 1 columna en móvil y 2 en
 * tableta/escritorio, y entrada con `RevealGroup` (stagger suave acotado que se
 * anula con `prefers-reduced-motion`, F10).
 *
 * El `overflow-x: clip` del revelado lo aporta `RevealGroup` (F10 T1018): sin él
 * la grilla desbordaría el ancho del documento y la página ganaría
 * desplazamiento horizontal (SC-017).
 */
export function MediosGrid({ items }: MediosGridProps) {
  const { t } = useLanguage();

  return (
    <RevealGroup>
      <Reveal from="left">
        <ul aria-label={t('a11y.media.card.heading')} className={GRID}>
          {items.map((item) => (
            <MedioCard key={item.id} item={item} />
          ))}
        </ul>
      </Reveal>
    </RevealGroup>
  );
}
