import { BrandLogo } from '../components/BrandLogo';
import { useLanguage } from '../i18n/useLanguage';

/**
 * Marcador **neutro** para un episodio sin carátula (ux.md §4.1/§6.3, ⟲ D-4):
 * el mismo marco 16:9 de la carátula real —la grilla nunca «salta» de forma y
 * no hay hueco— con la identidad de la iglesia (logotipo de marca del
 * repositorio y nombre). **Nunca** carga nada de terceros (SC-005) y **nunca**
 * es un `<img>` con `alt` inventado: es decoración, porque el nombre del
 * episodio va justo debajo en la propia ficha.
 */
export function CaratulaNeutra() {
  const { t } = useLanguage();

  return (
    <div
      data-testid="caratula-neutra"
      aria-hidden="true"
      className="flex aspect-video w-full flex-col items-center justify-center gap-2 bg-navy-soft text-navy"
    >
      <BrandLogo alt="" size={44} />
      <span className="font-display text-lg tracking-[0.18em]">{t('brand.name')}</span>
    </div>
  );
}
