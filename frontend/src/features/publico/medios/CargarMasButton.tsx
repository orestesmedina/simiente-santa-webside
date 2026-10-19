import { PublicButton } from '../../../components/publico/PublicButton';
import { useLanguage } from '../i18n/useLanguage';

export interface CargarMasButtonProps {
  /** Agrega la siguiente tanda (12 fichas) al final de la misma lista (Q11/D-3). */
  onClick: () => void;
}

/**
 * «Cargar más» de la sección `/medios` (ux.md §4.1, ⟲ D-3/FR-006): **no** es
 * paginación por páginas —la lista se agranda y el visitante no pierde su
 * lugar—, va a todo el ancho en móvil y cumple el objetivo táctil de 44 px.
 *
 * Solo se pinta **si quedan episodios** por mostrar; al agotarse desaparece sin
 * mensaje («no hay más» es ruido).
 */
export function CargarMasButton({ onClick }: CargarMasButtonProps) {
  const { t } = useLanguage();

  return (
    <div className="flex justify-center max-md:justify-stretch">
      <PublicButton variant="secondary" onClick={onClick} className="max-md:w-full">
        {t('media.loadMore')}
      </PublicButton>
    </div>
  );
}
